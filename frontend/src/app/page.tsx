'use client'

import { useEffect, useState } from 'react'
import { useQuery } from '@tanstack/react-query'
import { QueryClient, QueryClientProvider } from '@tanstack/react-query'
import { indexApi, userApi } from '@/lib/api'
import { useWSStore } from '@/lib/websocket'
import { useAuthStore } from '@/lib/auth'
import IndexCard from '@/components/IndexCard'
import PriceChart from '@/components/PriceChart'
import MarketOverview from '@/components/MarketOverview'
import Sidebar from '@/components/Sidebar'
import PortfolioPanel from '@/components/PortfolioPanel'
import AuthModal from '@/components/AuthModal'
import type { Index } from '@/types'
import type { NavTab } from '@/components/Sidebar'

const queryClient = new QueryClient({
  defaultOptions: { queries: { staleTime: 5_000, retry: 2 } },
})

function Dashboard() {
  const [selected, setSelected] = useState<string>('BTC')
  const [navTab, setNavTab] = useState<NavTab>('dashboard')
  const [showAuth, setShowAuth] = useState(false)

  const connect = useWSStore(s => s.connect)
  const disconnect = useWSStore(s => s.disconnect)
  const connected = useWSStore(s => s.connected)
  const { user, token, logout, isAuthenticated } = useAuthStore()

  useEffect(() => {
    connect()
    return () => disconnect()
  }, [connect, disconnect])

  const { data: indices = [], isLoading } = useQuery({
    queryKey: ['indices'],
    queryFn: indexApi.getAll,
    refetchInterval: 10_000,
  })

  const { data: watchlist = [] } = useQuery({
    queryKey: ['watchlist'],
    queryFn: userApi.getWatchlist,
    enabled: isAuthenticated(),
  })

  const { data: portfolios = [] } = useQuery({
    queryKey: ['portfolios'],
    queryFn: userApi.getPortfolios,
    enabled: isAuthenticated(),
  })

  const categories = [
    { key: 'crypto', label: '₿ Crypto' },
    { key: 'stock',  label: '🏛 US Stocks' },
    { key: 'vn',     label: '🇻🇳 VN Market' },
  ] as const

  // Watchlist view: filter indices to only watched symbols
  const displayedIndices = navTab === 'watchlist'
    ? indices.filter((i: Index) => watchlist.some(w => w.symbol === i.symbol))
    : indices

  return (
    <div className="flex flex-col min-h-screen bg-bg-primary">
      {/* ─── Header ─────────────────────────────────────────── */}
      <header className="border-b border-bg-border px-4 py-3 flex items-center justify-between z-10 bg-bg-secondary shrink-0">
        <div className="flex items-center gap-3">
          <span className="text-lg font-bold text-text-primary tracking-tight">📈 Portfolio Index</span>
          <span className="hidden sm:block text-xs text-text-muted font-mono bg-bg-border px-2 py-0.5 rounded">v1.0</span>
        </div>
        <div className="flex items-center gap-3">
          <div className="flex items-center gap-1.5">
            <span className={`w-2 h-2 rounded-full ${connected ? 'bg-accent-green animate-pulse' : 'bg-accent-red'}`} />
            <span className="text-xs text-text-secondary font-mono hidden sm:block">
              {connected ? 'LIVE' : 'OFFLINE'}
            </span>
          </div>
          {isAuthenticated() && user ? (
            <div className="flex items-center gap-2">
              <span className="hidden sm:block text-xs text-text-secondary">
                👤 {user.name}
              </span>
              <button
                onClick={logout}
                className="text-xs text-text-muted hover:text-accent-red px-2 py-1 rounded-lg border border-bg-border hover:border-accent-red/40 transition-all"
              >
                Đăng xuất
              </button>
            </div>
          ) : (
            <button
              onClick={() => setShowAuth(true)}
              className="text-xs bg-accent-blue hover:bg-accent-blue/80 text-white font-semibold px-4 py-1.5 rounded-lg transition-all"
            >
              Đăng nhập
            </button>
          )}
        </div>
      </header>

      {/* ─── Body ───────────────────────────────────────────── */}
      <div className="flex flex-1 min-h-0">
        <Sidebar
          active={navTab}
          onChange={setNavTab}
          watchlistCount={watchlist.length}
          portfolioCount={portfolios.length}
        />

        {/* ─── Main content ──────────────────────────────── */}
        <main className="flex-1 overflow-y-auto">
          {/* Portfolio tab */}
          {navTab === 'portfolio' && (
            <div className="h-full">
              {isAuthenticated() ? (
                <PortfolioPanel />
              ) : (
                <div className="flex flex-col items-center justify-center h-full text-center p-8">
                  <p className="text-4xl mb-4">💼</p>
                  <p className="text-text-secondary mb-4">Đăng nhập để quản lý danh mục đầu tư của bạn</p>
                  <button
                    onClick={() => setShowAuth(true)}
                    className="bg-accent-blue hover:bg-accent-blue/80 text-white font-semibold px-6 py-2.5 rounded-xl transition-all"
                  >
                    Đăng nhập / Đăng ký
                  </button>
                </div>
              )}
            </div>
          )}

          {/* Dashboard & Watchlist tabs */}
          {(navTab === 'dashboard' || navTab === 'watchlist') && (
            <div className="px-4 py-5 space-y-6 max-w-[1400px] mx-auto">
              {/* Watchlist empty state */}
              {navTab === 'watchlist' && !isAuthenticated() && (
                <div className="flex flex-col items-center py-16 text-center">
                  <p className="text-4xl mb-3">⭐</p>
                  <p className="text-text-secondary mb-4">Đăng nhập để lưu danh sách theo dõi</p>
                  <button onClick={() => setShowAuth(true)}
                    className="bg-accent-blue text-white px-6 py-2.5 rounded-xl text-sm font-semibold transition-all hover:bg-accent-blue/80">
                    Đăng nhập
                  </button>
                </div>
              )}

              {navTab === 'watchlist' && isAuthenticated() && watchlist.length === 0 && (
                <div className="flex flex-col items-center py-16 text-center">
                  <p className="text-4xl mb-3">⭐</p>
                  <p className="text-text-secondary">Chưa có chỉ số nào trong danh sách theo dõi</p>
                  <p className="text-xs text-text-muted mt-1">Bấm ★ trên các thẻ chỉ số để thêm vào</p>
                  <button onClick={() => setNavTab('dashboard')} className="mt-4 text-sm text-accent-blue hover:underline">
                    Quay về thị trường
                  </button>
                </div>
              )}

              {/* Index cards grouped by category */}
              {categories.map(({ key, label }) => {
                const group = displayedIndices.filter((i: Index) => i.category === key)
                if (group.length === 0) return null
                return (
                  <section key={key}>
                    <h2 className="text-xs font-semibold text-text-secondary uppercase tracking-widest mb-3">
                      {label}
                    </h2>
                    <div className="grid grid-cols-2 sm:grid-cols-3 lg:grid-cols-4 xl:grid-cols-5 gap-3 group">
                      {isLoading
                        ? Array.from({ length: 3 }).map((_, i) => (
                            <div key={i} className="card animate-pulse h-28 bg-bg-card" />
                          ))
                        : group.map((idx: Index) => (
                            <IndexCard
                              key={idx.symbol}
                              index={idx}
                              selected={selected === idx.symbol}
                              onClick={() => setSelected(idx.symbol)}
                              watchlist={watchlist}
                            />
                          ))}
                    </div>
                  </section>
                )
              })}

              {/* Chart + Overview */}
              {(navTab === 'dashboard' || (navTab === 'watchlist' && displayedIndices.length > 0)) && (
                <div className="grid grid-cols-1 xl:grid-cols-3 gap-5">
                  <div className="xl:col-span-2">
                    <PriceChart symbol={selected} />
                  </div>
                  <div>
                    <MarketOverview symbol={selected} />
                  </div>
                </div>
              )}
            </div>
          )}
        </main>
      </div>

      {/* Auth modal */}
      {showAuth && <AuthModal onClose={() => setShowAuth(false)} />}
    </div>
  )
}

export default function Home() {
  return (
    <QueryClientProvider client={queryClient}>
      <Dashboard />
    </QueryClientProvider>
  )
}
