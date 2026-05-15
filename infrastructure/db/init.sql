-- ============================================================
-- Portfolio Index — TimescaleDB Schema
-- ============================================================

-- Enable TimescaleDB extension
CREATE EXTENSION IF NOT EXISTS timescaledb CASCADE;

-- ─── Index Snapshots (latest price per symbol) ───────────────
CREATE TABLE IF NOT EXISTS index_snapshots (
    symbol          TEXT        PRIMARY KEY,
    name            TEXT        NOT NULL,
    category        TEXT        NOT NULL CHECK (category IN ('stock','crypto','vn')),
    price           DOUBLE PRECISION NOT NULL DEFAULT 0,
    change          DOUBLE PRECISION NOT NULL DEFAULT 0,
    change_percent  DOUBLE PRECISION NOT NULL DEFAULT 0,
    volume          DOUBLE PRECISION NOT NULL DEFAULT 0,
    market_cap      DOUBLE PRECISION NOT NULL DEFAULT 0,
    high_24h        DOUBLE PRECISION NOT NULL DEFAULT 0,
    low_24h         DOUBLE PRECISION NOT NULL DEFAULT 0,
    updated_at      TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

-- ─── OHLCV Candles (time-series) ─────────────────────────────
CREATE TABLE IF NOT EXISTS candles (
    symbol  TEXT             NOT NULL,
    time    TIMESTAMPTZ      NOT NULL,
    open    DOUBLE PRECISION NOT NULL,
    high    DOUBLE PRECISION NOT NULL,
    low     DOUBLE PRECISION NOT NULL,
    close   DOUBLE PRECISION NOT NULL,
    volume  DOUBLE PRECISION NOT NULL DEFAULT 0,
    PRIMARY KEY (symbol, time)
);

-- Convert to hypertable (TimescaleDB time-partitioning)
SELECT create_hypertable('candles', 'time',
    chunk_time_interval => INTERVAL '7 days',
    if_not_exists => TRUE
);

-- Indexes for common queries
CREATE INDEX IF NOT EXISTS idx_candles_symbol_time
    ON candles (symbol, time DESC);

-- ─── Continuous Aggregates (pre-computed OHLCV) ───────────────
-- 1-hour aggregate
CREATE MATERIALIZED VIEW IF NOT EXISTS candles_1h
WITH (timescaledb.continuous) AS
SELECT
    symbol,
    time_bucket('1 hour', time)  AS bucket,
    first(open, time)            AS open,
    max(high)                    AS high,
    min(low)                     AS low,
    last(close, time)            AS close,
    sum(volume)                  AS volume
FROM candles
GROUP BY symbol, bucket
WITH NO DATA;

-- 1-day aggregate
CREATE MATERIALIZED VIEW IF NOT EXISTS candles_1d
WITH (timescaledb.continuous) AS
SELECT
    symbol,
    time_bucket('1 day', time)   AS bucket,
    first(open, time)            AS open,
    max(high)                    AS high,
    min(low)                     AS low,
    last(close, time)            AS close,
    sum(volume)                  AS volume
FROM candles
GROUP BY symbol, bucket
WITH NO DATA;

-- Auto-refresh policies
SELECT add_continuous_aggregate_policy('candles_1h',
    start_offset => INTERVAL '3 hours',
    end_offset   => INTERVAL '1 hour',
    schedule_interval => INTERVAL '30 minutes',
    if_not_exists => TRUE
);

SELECT add_continuous_aggregate_policy('candles_1d',
    start_offset => INTERVAL '3 days',
    end_offset   => INTERVAL '1 day',
    schedule_interval => INTERVAL '1 hour',
    if_not_exists => TRUE
);

-- ─── Data Retention Policy (keep 2 years of raw candles) ─────
SELECT add_retention_policy('candles',
    INTERVAL '2 years',
    if_not_exists => TRUE
);

-- ─── Seed: Default Indices ───────────────────────────────────
INSERT INTO index_snapshots (symbol, name, category) VALUES
    ('BTC',     'Bitcoin',              'crypto'),
    ('ETH',     'Ethereum',             'crypto'),
    ('COIN50',  'Crypto Top 50',        'crypto'),
    ('SPX',     'S&P 500',              'stock'),
    ('NDX',     'Nasdaq 100',           'stock'),
    ('DJI',     'Dow Jones Industrial', 'stock'),
    ('VNINDEX', 'VN-Index',             'vn'),
    ('VN30',    'VN30 Index',           'vn'),
    ('HNX30',   'HNX30 Index',          'vn')
ON CONFLICT (symbol) DO NOTHING;
