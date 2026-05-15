package models

import "time"

// Category constants for index classification
const (
	CategoryStock  = "stock"
	CategoryCrypto = "crypto"
	CategoryVN     = "vn"
)

// Index represents a market index snapshot.
type Index struct {
	Symbol        string    `json:"symbol"         db:"symbol"`
	Name          string    `json:"name"           db:"name"`
	Category      string    `json:"category"       db:"category"`
	Price         float64   `json:"price"          db:"price"`
	Change        float64   `json:"change"         db:"change"`
	ChangePercent float64   `json:"change_percent" db:"change_percent"`
	Volume        float64   `json:"volume"         db:"volume"`
	MarketCap     float64   `json:"market_cap"     db:"market_cap"`
	High24h       float64   `json:"high_24h"       db:"high_24h"`
	Low24h        float64   `json:"low_24h"        db:"low_24h"`
	UpdatedAt     time.Time `json:"updated_at"     db:"updated_at"`
}

// Candle represents a single OHLCV candlestick.
type Candle struct {
	Symbol string    `json:"symbol" db:"symbol"`
	Time   time.Time `json:"time"   db:"time"`
	Open   float64   `json:"open"   db:"open"`
	High   float64   `json:"high"   db:"high"`
	Low    float64   `json:"low"    db:"low"`
	Close  float64   `json:"close"  db:"close"`
	Volume float64   `json:"volume" db:"volume"`
}

// PriceUpdate is the real-time WebSocket broadcast payload.
type PriceUpdate struct {
	Symbol        string    `json:"symbol"`
	Price         float64   `json:"price"`
	Change        float64   `json:"change"`
	ChangePercent float64   `json:"change_percent"`
	Volume        float64   `json:"volume"`
	Timestamp     time.Time `json:"timestamp"`
}

// CandleQuery holds query parameters for OHLCV requests.
type CandleQuery struct {
	Interval string `query:"interval"` // 1m 5m 15m 1h 4h 1d 1w
	From     int64  `query:"from"`     // unix timestamp
	To       int64  `query:"to"`       // unix timestamp
	Limit    int    `query:"limit"`    // max rows (default 300)
}

// IndexStats holds aggregated statistics for an index.
type IndexStats struct {
	Symbol        string  `json:"symbol"`
	ATH           float64 `json:"ath"`
	ATL           float64 `json:"atl"`
	Change7d      float64 `json:"change_7d"`
	Change30d     float64 `json:"change_30d"`
	Change1y      float64 `json:"change_1y"`
	AvgVolume30d  float64 `json:"avg_volume_30d"`
	Volatility30d float64 `json:"volatility_30d"`
}

// WatchlistEntry is used internally for the list of tracked indices.
type WatchlistEntry struct {
	Symbol   string `json:"symbol"`
	Name     string `json:"name"`
	Category string `json:"category"`
}

// DefaultIndices is the seeded list of tracked market indices.
var DefaultIndices = []WatchlistEntry{
	// US Stocks
	{Symbol: "SPX", Name: "S&P 500", Category: CategoryStock},
	{Symbol: "NDX", Name: "Nasdaq 100", Category: CategoryStock},
	{Symbol: "DJI", Name: "Dow Jones Industrial", Category: CategoryStock},
	// VN Market
	{Symbol: "VNINDEX", Name: "VN-Index", Category: CategoryVN},
	{Symbol: "VN30", Name: "VN30 Index", Category: CategoryVN},
	{Symbol: "HNX30", Name: "HNX30 Index", Category: CategoryVN},
	// Crypto
	{Symbol: "BTC", Name: "Bitcoin", Category: CategoryCrypto},
	{Symbol: "ETH", Name: "Ethereum", Category: CategoryCrypto},
	{Symbol: "COIN50", Name: "Crypto Top 50", Category: CategoryCrypto},
}
