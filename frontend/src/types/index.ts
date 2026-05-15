// ─── Market Index Types ───────────────────────────────────────

export interface Index {
  symbol: string
  name: string
  category: 'stock' | 'crypto' | 'vn'
  price: number
  change: number
  change_percent: number
  volume: number
  market_cap: number
  high_24h: number
  low_24h: number
  updated_at: string
}

export interface Candle {
  symbol: string
  time: string
  open: number
  high: number
  low: number
  close: number
  volume: number
}

export interface IndexStats {
  symbol: string
  ath: number
  atl: number
  change_7d: number
  change_30d: number
  change_1y: number
  avg_volume_30d: number
  volatility_30d: number
}

// ─── AI Analysis Types ────────────────────────────────────────

export interface AIAnalysis {
  symbol: string
  signal: 'BUY' | 'SELL' | 'NEUTRAL'
  confidence: number        // 0-100
  trend: 'UPTREND' | 'DOWNTREND' | 'SIDEWAYS'
  support: number
  resistance: number
  rsi: number
  macd: { value: number; signal: number; histogram: number }
  forecast_7d: number[]     // predicted close prices
  summary: string
  generated_at: string
}

// ─── WebSocket Types ──────────────────────────────────────────

export interface PriceUpdate {
  symbol: string
  price: number
  change: number
  change_percent: number
  volume: number
  timestamp: string
}

// ─── UI State Types ───────────────────────────────────────────

export type ChartInterval = '1m' | '5m' | '15m' | '1h' | '4h' | '1d' | '1w'

export interface ChartState {
  symbol: string
  interval: ChartInterval
}
