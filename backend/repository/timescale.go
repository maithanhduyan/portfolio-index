package repository

import (
	"context"
	"fmt"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"

	"portfolio-index/models"
)

// TimescaleDB wraps a pgxpool connection pool.
type TimescaleDB struct {
	pool *pgxpool.Pool
}

func NewTimescaleDB(dsn string) (*TimescaleDB, error) {
	cfg, err := pgxpool.ParseConfig(dsn)
	if err != nil {
		return nil, fmt.Errorf("parse db config: %w", err)
	}
	// Tuned for high concurrency
	cfg.MaxConns = 50
	cfg.MinConns = 5
	cfg.MaxConnLifetime = 30 * time.Minute
	cfg.MaxConnIdleTime = 5 * time.Minute

	pool, err := pgxpool.NewWithConfig(context.Background(), cfg)
	if err != nil {
		return nil, fmt.Errorf("create pool: %w", err)
	}

	if err := pool.Ping(context.Background()); err != nil {
		return nil, fmt.Errorf("db ping: %w", err)
	}

	return &TimescaleDB{pool: pool}, nil
}

func (db *TimescaleDB) Close() {
	db.pool.Close()
}

// ─── Index Queries ────────────────────────────────────────────

func (db *TimescaleDB) GetAllIndices(ctx context.Context) ([]models.Index, error) {
	rows, err := db.pool.Query(ctx, `
		SELECT symbol, name, category, price, change, change_percent,
		       volume, market_cap, high_24h, low_24h, updated_at
		FROM index_snapshots
		ORDER BY category, symbol
	`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var result []models.Index
	for rows.Next() {
		var idx models.Index
		if err := rows.Scan(
			&idx.Symbol, &idx.Name, &idx.Category,
			&idx.Price, &idx.Change, &idx.ChangePercent,
			&idx.Volume, &idx.MarketCap, &idx.High24h, &idx.Low24h,
			&idx.UpdatedAt,
		); err != nil {
			return nil, err
		}
		result = append(result, idx)
	}
	return result, rows.Err()
}

func (db *TimescaleDB) GetIndex(ctx context.Context, symbol string) (*models.Index, error) {
	var idx models.Index
	err := db.pool.QueryRow(ctx, `
		SELECT symbol, name, category, price, change, change_percent,
		       volume, market_cap, high_24h, low_24h, updated_at
		FROM index_snapshots
		WHERE symbol = $1
	`, symbol).Scan(
		&idx.Symbol, &idx.Name, &idx.Category,
		&idx.Price, &idx.Change, &idx.ChangePercent,
		&idx.Volume, &idx.MarketCap, &idx.High24h, &idx.Low24h,
		&idx.UpdatedAt,
	)
	if err != nil {
		return nil, err
	}
	return &idx, nil
}

func (db *TimescaleDB) UpsertIndex(ctx context.Context, idx *models.Index) error {
	_, err := db.pool.Exec(ctx, `
		INSERT INTO index_snapshots
			(symbol, name, category, price, change, change_percent, volume, market_cap, high_24h, low_24h, updated_at)
		VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9,$10,$11)
		ON CONFLICT (symbol) DO UPDATE SET
			price          = EXCLUDED.price,
			change         = EXCLUDED.change,
			change_percent = EXCLUDED.change_percent,
			volume         = EXCLUDED.volume,
			market_cap     = EXCLUDED.market_cap,
			high_24h       = EXCLUDED.high_24h,
			low_24h        = EXCLUDED.low_24h,
			updated_at     = EXCLUDED.updated_at
	`, idx.Symbol, idx.Name, idx.Category,
		idx.Price, idx.Change, idx.ChangePercent,
		idx.Volume, idx.MarketCap, idx.High24h, idx.Low24h,
		idx.UpdatedAt,
	)
	return err
}

// ─── OHLCV Candle Queries ─────────────────────────────────────

func (db *TimescaleDB) GetCandles(ctx context.Context, symbol, interval string, from, to time.Time, limit int) ([]models.Candle, error) {
	// Map interval to TimescaleDB time_bucket width
	bucketMap := map[string]string{
		"1m": "1 minute", "5m": "5 minutes", "15m": "15 minutes",
		"1h": "1 hour", "4h": "4 hours",
		"1d": "1 day", "1w": "1 week",
	}
	bucket, ok := bucketMap[interval]
	if !ok {
		bucket = "1 hour"
	}

	if limit <= 0 || limit > 1000 {
		limit = 300
	}

	rows, err := db.pool.Query(ctx, `
		SELECT
			$1::text              AS symbol,
			time_bucket($2, time) AS bucket,
			first(open, time)     AS open,
			max(high)             AS high,
			min(low)              AS low,
			last(close, time)     AS close,
			sum(volume)           AS volume
		FROM candles
		WHERE symbol = $1
		  AND time BETWEEN $3 AND $4
		GROUP BY bucket
		ORDER BY bucket DESC
		LIMIT $5
	`, symbol, bucket, from, to, limit)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	result := make([]models.Candle, 0) // always return [] not null
	for rows.Next() {
		var c models.Candle
		if err := rows.Scan(&c.Symbol, &c.Time, &c.Open, &c.High, &c.Low, &c.Close, &c.Volume); err != nil {
			return nil, err
		}
		result = append(result, c)
	}
	return result, rows.Err()
}

func (db *TimescaleDB) InsertCandle(ctx context.Context, c *models.Candle) error {
	_, err := db.pool.Exec(ctx, `
		INSERT INTO candles (symbol, time, open, high, low, close, volume)
		VALUES ($1,$2,$3,$4,$5,$6,$7)
		ON CONFLICT (symbol, time) DO UPDATE SET
			high   = GREATEST(candles.high, EXCLUDED.high),
			low    = LEAST(candles.low, EXCLUDED.low),
			close  = EXCLUDED.close,
			volume = candles.volume + EXCLUDED.volume
	`, c.Symbol, c.Time, c.Open, c.High, c.Low, c.Close, c.Volume)
	return err
}

// ─── Stats Queries ────────────────────────────────────────────

func (db *TimescaleDB) GetStats(ctx context.Context, symbol string) (*models.IndexStats, error) {
	var s models.IndexStats
	s.Symbol = symbol

	err := db.pool.QueryRow(ctx, `
		WITH stats AS (
			SELECT
				COALESCE(max(high), 0)   AS ath,
				COALESCE(min(low),  0)   AS atl,
				COALESCE(avg(volume) FILTER (WHERE time >= NOW() - INTERVAL '30 days'), 0) AS avg_vol_30d,
				COALESCE(stddev(close) FILTER (WHERE time >= NOW() - INTERVAL '30 days'), 0) AS vol_30d
			FROM candles
			WHERE symbol = $1
		),
		prices AS (
			SELECT
				COALESCE(last(close, time) FILTER (WHERE time >= NOW() - INTERVAL '7 days'),  0) AS p7,
				COALESCE(last(close, time) FILTER (WHERE time >= NOW() - INTERVAL '30 days'), 0) AS p30,
				COALESCE(last(close, time) FILTER (WHERE time >= NOW() - INTERVAL '1 year'),  0) AS p1y,
				COALESCE(last(close, time), 0) AS current
			FROM candles
			WHERE symbol = $1
		)
		SELECT
			s.ath, s.atl,
			CASE WHEN p.p7  > 0 THEN (p.current - p.p7)  / p.p7  * 100 ELSE 0 END AS change_7d,
			CASE WHEN p.p30 > 0 THEN (p.current - p.p30) / p.p30 * 100 ELSE 0 END AS change_30d,
			CASE WHEN p.p1y > 0 THEN (p.current - p.p1y) / p.p1y * 100 ELSE 0 END AS change_1y,
			s.avg_vol_30d,
			s.vol_30d
		FROM stats s, prices p
	`, symbol).Scan(
		&s.ATH, &s.ATL,
		&s.Change7d, &s.Change30d, &s.Change1y,
		&s.AvgVolume30d, &s.Volatility30d,
	)
	if err != nil {
		// Return zeroed stats on empty table — not a fatal error
		return &s, nil
	}
	return &s, nil
}
