import axios from 'axios'
import type { Index, Candle, IndexStats, AIAnalysis, ChartInterval } from '@/types'

const api = axios.create({
  baseURL: process.env.NEXT_PUBLIC_API_URL ?? '/api/v1',
  timeout: 10_000,
})

export const indexApi = {
  getAll: (): Promise<Index[]> =>
    api.get('/indices').then(r => r.data),

  getOne: (symbol: string): Promise<Index> =>
    api.get(`/indices/${symbol}`).then(r => r.data),

  getCandles: (
    symbol: string,
    interval: ChartInterval = '1h',
    limit = 300,
  ): Promise<Candle[]> =>
    api
      .get(`/indices/${symbol}/candles`, { params: { interval, limit } })
      .then(r => r.data),

  getStats: (symbol: string): Promise<IndexStats> =>
    api.get(`/indices/${symbol}/stats`).then(r => r.data),

  getAIAnalysis: (symbol: string): Promise<AIAnalysis> =>
    api.get(`/indices/${symbol}/ai-analysis`).then(r => r.data),
}
