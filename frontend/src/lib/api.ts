import axios from 'axios'
import type { Index, Candle, IndexStats, AIAnalysis, ChartInterval, Portfolio, WatchlistItem, Note, AuthResponse } from '@/types'

const api = axios.create({
  baseURL: process.env.NEXT_PUBLIC_API_URL ?? '/api/v1',
  timeout: 10_000,
})

// Attach JWT token from localStorage on every request
api.interceptors.request.use((config) => {
  if (typeof window !== 'undefined') {
    try {
      const raw = localStorage.getItem('pi-auth')
      if (raw) {
        const { state } = JSON.parse(raw)
        if (state?.token) config.headers.Authorization = `Bearer ${state.token}`
      }
    } catch { /* ignore */ }
  }
  return config
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

export const authApi = {
  register: (email: string, name: string, password: string): Promise<AuthResponse> =>
    api.post('/auth/register', { email, name, password }).then(r => r.data),

  login: (email: string, password: string): Promise<AuthResponse> =>
    api.post('/auth/login', { email, password }).then(r => r.data),
}

export const userApi = {
  getWatchlist: (): Promise<WatchlistItem[]> =>
    api.get('/user/watchlist').then(r => r.data),

  addWatchlist: (symbol: string) =>
    api.post(`/user/watchlist/${symbol}`).then(r => r.data),

  removeWatchlist: (symbol: string) =>
    api.delete(`/user/watchlist/${symbol}`).then(r => r.data),

  getNotes: (symbol: string): Promise<Note[]> =>
    api.get(`/user/notes/${symbol}`).then(r => r.data),

  createNote: (symbol: string, content: string): Promise<Note> =>
    api.post(`/user/notes/${symbol}`, { content }).then(r => r.data),

  updateNote: (id: string, content: string): Promise<Note> =>
    api.put(`/user/notes/${id}`, { content }).then(r => r.data),

  deleteNote: (id: string) =>
    api.delete(`/user/notes/${id}`).then(r => r.data),

  getPortfolios: (): Promise<Portfolio[]> =>
    api.get('/user/portfolios').then(r => r.data),

  createPortfolio: (name: string, description: string, currency: string): Promise<Portfolio> =>
    api.post('/user/portfolios', { name, description, currency }).then(r => r.data),

  deletePortfolio: (id: string) =>
    api.delete(`/user/portfolios/${id}`).then(r => r.data),

  addHolding: (portfolioId: string, data: { symbol: string; quantity: number; avg_price: number; buy_date?: string; note?: string }) =>
    api.post(`/user/portfolios/${portfolioId}/holdings`, data).then(r => r.data),

  removeHolding: (portfolioId: string, holdingId: string) =>
    api.delete(`/user/portfolios/${portfolioId}/holdings/${holdingId}`).then(r => r.data),
}

