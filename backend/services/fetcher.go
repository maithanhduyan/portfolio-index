package services

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"time"

	"github.com/gorilla/websocket"

	"portfolio-index/config"
	"portfolio-index/models"
	"portfolio-index/repository"
)

const priceUpdateChannel = "price_updates"

// Fetcher pulls market data from multiple sources and stores it.
type Fetcher struct {
	db     *repository.TimescaleDB
	cache  *RedisCache
	cfg    *config.Config
	client *http.Client
}

func NewFetcher(db *repository.TimescaleDB, cache *RedisCache, cfg *config.Config) *Fetcher {
	return &Fetcher{
		db:    db,
		cache: cache,
		cfg:   cfg,
		client: &http.Client{
			Timeout: 10 * time.Second,
		},
	}
}

// Start launches all data-fetching goroutines.
func (f *Fetcher) Start() {
	log.Println("[fetcher] starting data sources...")

	// Crypto: connect to Binance WebSocket (free, no API key needed)
	go f.streamBinance()

	// VN market: poll TCBS every 10 seconds
	go f.pollVNMarket()

	// US indices: poll Alpha Vantage every minute (free tier: 25 req/day)
	go f.pollUSIndices()
}

// ─── Crypto (Binance WebSocket) ───────────────────────────────

type binanceTicker struct {
	Symbol    string `json:"s"`
	Price     string `json:"c"`
	Change    string `json:"P"` // percent change
	Volume    string `json:"q"` // quote volume
	High      string `json:"h"`
	Low       string `json:"l"`
	EventTime int64  `json:"E"`
}

func (f *Fetcher) streamBinance() {
	symbols := []string{"btcusdt", "ethusdt"}
	streams := ""
	for i, s := range symbols {
		if i > 0 {
			streams += "/"
		}
		streams += s + "@ticker"
	}
	url := fmt.Sprintf("%s/stream?streams=%s", f.cfg.BinanceWSUrl, streams)

	for {
		if err := f.connectBinance(url); err != nil {
			log.Printf("[binance] disconnected: %v — reconnecting in 5s", err)
		}
		time.Sleep(5 * time.Second)
	}
}

func (f *Fetcher) connectBinance(url string) error {
	conn, _, err := websocket.DefaultDialer.Dial(url, nil)
	if err != nil {
		return err
	}
	defer conn.Close()
	log.Println("[binance] connected")

	type streamMsg struct {
		Stream string        `json:"stream"`
		Data   binanceTicker `json:"data"`
	}

	for {
		_, msg, err := conn.ReadMessage()
		if err != nil {
			return err
		}

		var sm streamMsg
		if err := json.Unmarshal(msg, &sm); err != nil {
			continue
		}

		ticker := sm.Data
		symbol := mapBinanceSymbol(ticker.Symbol)
		if symbol == "" {
			continue
		}

		price := parseFloat(ticker.Price)
		changePct := parseFloat(ticker.Change)
		high := parseFloat(ticker.High)
		low := parseFloat(ticker.Low)
		vol := parseFloat(ticker.Volume)

		idx := &models.Index{
			Symbol:        symbol,
			Name:          symbolName(symbol),
			Category:      models.CategoryCrypto,
			Price:         price,
			Change:        price * changePct / 100,
			ChangePercent: changePct,
			Volume:        vol,
			High24h:       high,
			Low24h:        low,
			UpdatedAt:     time.UnixMilli(ticker.EventTime),
		}

		ctx := context.Background()
		if err := f.db.UpsertIndex(ctx, idx); err != nil {
			log.Printf("[binance] db upsert error: %v", err)
		}

		update := models.PriceUpdate{
			Symbol:        idx.Symbol,
			Price:         idx.Price,
			Change:        idx.Change,
			ChangePercent: idx.ChangePercent,
			Volume:        idx.Volume,
			Timestamp:     idx.UpdatedAt,
		}
		_ = f.cache.Publish(ctx, priceUpdateChannel, update)
		_ = f.cache.InvalidateIndex(ctx, symbol)
	}
}

// ─── VN Market (TCBS) ────────────────────────────────────────

func (f *Fetcher) pollVNMarket() {
	ticker := time.NewTicker(10 * time.Second)
	defer ticker.Stop()

	for range ticker.C {
		ctx, cancel := context.WithTimeout(context.Background(), 8*time.Second)
		f.fetchVNIndex(ctx, "VNINDEX")
		f.fetchVNIndex(ctx, "VN30")
		cancel()
	}
}

func (f *Fetcher) fetchVNIndex(ctx context.Context, symbol string) {
	// TCBS public endpoint — no auth required
	url := fmt.Sprintf("%s/stock-insight/v1/index/time-series?ticker=%s&type=stock&count=1&resolution=1D",
		f.cfg.TCBSBaseUrl, symbol)

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		return
	}
	req.Header.Set("Accept", "application/json")

	resp, err := f.client.Do(req)
	if err != nil {
		log.Printf("[tcbs] %s fetch error: %v", symbol, err)
		return
	}
	defer resp.Body.Close()

	var result struct {
		Data []struct {
			Close  float64 `json:"close"`
			Open   float64 `json:"open"`
			High   float64 `json:"high"`
			Low    float64 `json:"low"`
			Volume float64 `json:"volume"`
			Time   int64   `json:"tradingDate"`
		} `json:"data"`
	}

	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil || len(result.Data) == 0 {
		return
	}

	d := result.Data[0]
	changePct := 0.0
	if d.Open > 0 {
		changePct = (d.Close - d.Open) / d.Open * 100
	}

	idx := &models.Index{
		Symbol:        symbol,
		Name:          symbolName(symbol),
		Category:      models.CategoryVN,
		Price:         d.Close,
		Change:        d.Close - d.Open,
		ChangePercent: changePct,
		Volume:        d.Volume,
		High24h:       d.High,
		Low24h:        d.Low,
		UpdatedAt:     time.Now(),
	}

	if err := f.db.UpsertIndex(ctx, idx); err != nil {
		log.Printf("[tcbs] %s db upsert error: %v", symbol, err)
		return
	}

	update := models.PriceUpdate{
		Symbol:        idx.Symbol,
		Price:         idx.Price,
		Change:        idx.Change,
		ChangePercent: idx.ChangePercent,
		Volume:        idx.Volume,
		Timestamp:     idx.UpdatedAt,
	}
	_ = f.cache.Publish(ctx, priceUpdateChannel, update)
	_ = f.cache.InvalidateIndex(ctx, symbol)
}

// ─── US Indices (Alpha Vantage) ───────────────────────────────

func (f *Fetcher) pollUSIndices() {
	if f.cfg.AlphaVantageKey == "" {
		log.Println("[alpha-vantage] no API key set, skipping US index polling")
		return
	}

	ticker := time.NewTicker(60 * time.Second)
	defer ticker.Stop()

	symbols := map[string]string{
		"SPX": "^GSPC",
		"NDX": "^NDX",
	}

	for range ticker.C {
		for local, remote := range symbols {
			ctx, cancel := context.WithTimeout(context.Background(), 8*time.Second)
			f.fetchAlphaVantage(ctx, local, remote)
			cancel()
			// Respect free-tier rate limit
			time.Sleep(15 * time.Second)
		}
	}
}

func (f *Fetcher) fetchAlphaVantage(ctx context.Context, symbol, avSymbol string) {
	url := fmt.Sprintf(
		"https://www.alphavantage.co/query?function=GLOBAL_QUOTE&symbol=%s&apikey=%s",
		avSymbol, f.cfg.AlphaVantageKey,
	)

	req, _ := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	resp, err := f.client.Do(req)
	if err != nil {
		log.Printf("[alpha-vantage] %s error: %v", symbol, err)
		return
	}
	defer resp.Body.Close()

	var result struct {
		Quote struct {
			Price         string `json:"05. price"`
			Change        string `json:"09. change"`
			ChangePercent string `json:"10. change percent"`
			Volume        string `json:"06. volume"`
			High          string `json:"03. high"`
			Low           string `json:"04. low"`
		} `json:"Global Quote"`
	}

	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return
	}

	q := result.Quote
	changePctStr := q.ChangePercent
	if len(changePctStr) > 0 && changePctStr[len(changePctStr)-1] == '%' {
		changePctStr = changePctStr[:len(changePctStr)-1]
	}

	idx := &models.Index{
		Symbol:        symbol,
		Name:          symbolName(symbol),
		Category:      models.CategoryStock,
		Price:         parseFloat(q.Price),
		Change:        parseFloat(q.Change),
		ChangePercent: parseFloat(changePctStr),
		Volume:        parseFloat(q.Volume),
		High24h:       parseFloat(q.High),
		Low24h:        parseFloat(q.Low),
		UpdatedAt:     time.Now(),
	}

	if err := f.db.UpsertIndex(ctx, idx); err != nil {
		log.Printf("[alpha-vantage] %s db upsert error: %v", symbol, err)
		return
	}

	update := models.PriceUpdate{
		Symbol:        idx.Symbol,
		Price:         idx.Price,
		Change:        idx.Change,
		ChangePercent: idx.ChangePercent,
		Volume:        idx.Volume,
		Timestamp:     idx.UpdatedAt,
	}
	_ = f.cache.Publish(ctx, priceUpdateChannel, update)
	_ = f.cache.InvalidateIndex(ctx, symbol)
}

// ─── Helpers ─────────────────────────────────────────────────

func mapBinanceSymbol(s string) string {
	m := map[string]string{
		"BTCUSDT": "BTC",
		"ETHUSDT": "ETH",
	}
	return m[s]
}

func symbolName(s string) string {
	m := map[string]string{
		"BTC":     "Bitcoin",
		"ETH":     "Ethereum",
		"COIN50":  "Crypto Top 50",
		"SPX":     "S&P 500",
		"NDX":     "Nasdaq 100",
		"VNINDEX": "VN-Index",
		"VN30":    "VN30 Index",
		"HNX30":   "HNX30 Index",
	}
	if name, ok := m[s]; ok {
		return name
	}
	return s
}

func parseFloat(s string) float64 {
	var f float64
	fmt.Sscanf(s, "%f", &f)
	return f
}
