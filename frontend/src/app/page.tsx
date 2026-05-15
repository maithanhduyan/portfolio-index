'use client'

import { useEffect, useState } from 'react'
import { useQuery } from '@tanstack/react-query'
import { QueryClient, QueryClientProvider } from '@tanstack/react-query'
import { indexApi } from '@/lib/api'
import { useWSStore } from '@/lib/websocket'
import IndexCard from '@/components/IndexCard'
import PriceChart from '@/components/PriceChart'
import MarketOverview from '@/components/MarketOverview'
import type { Index } from '@/types'

const queryClient = new QueryClient({
  defaultOptions: { queries: { staleTime: 5_000, retry: 2 } },
})

function Dashboard() {
  const [selected, setSelected] = useState<string>('BTC')
  const connect = useWSStore(s => s.connect)
  const disconnect = useWSStore(s => s.disconnect)
  const connected = useWSStore(s => s.connected)

  useEffect(() => {
    connect()
    return () => disconnect()
  }, [connect, disconnect])

  const { data: indices = [], isLoading } = useQuery({
    queryKey: ['indices'],
    queryFn: indexApi.getAll,
    refetchInterval: 10_000,
  })

  const categories = [
    { key: 'crypto', label: 'Crypto' },
    { key: 'stock',  label: 'US Stocks' },
    { key: 'vn',     label: 'VN Market' },
  ] as const

  return (
    <div className="min-h-screen bg-bg-primary">
      {/* Header */}
      <header className="border-b border-bg-border px-6 py-4 flex items-center justify-between">
        <div className="flex items-center gap-3">
          <span className="text-xl font-bold text-text-primary tracking-tight">
            📈 Portfolio Index
          </span>
          <span className="text-xs text-text-muted font-mono">v1.0</span>
        </div>
        <div className="flex items-center gap-2">
          <span className={`w-2 h-2 rounded-full ${connected ? 'bg-accent-green' : 'bg-accent-red'}`} />
          <span className="text-xs text-text-secondary font-mono">
            {connected ? 'LIVE' : 'OFFLINE'}
          </span>
        </div>
      </header>

      <main className="px-6 py-6 space-y-8 max-w-[1600px] mx-auto">
        {/* Index Cards by Category */}
        {categories.map(({ key, label }) => {
          const group = indices.filter((i: Index) => i.category === key)
          if (group.length === 0 && !isLoading) return null
          return (
            <section key={key}>
              <h2 className="text-sm font-semibold text-text-secondary uppercase tracking-widest mb-3">
                {label}
              </h2>
              <div className="grid grid-cols-2 sm:grid-cols-3 lg:grid-cols-4 xl:grid-cols-5 gap-3">
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
                      />
                    ))}
              </div>
            </section>
          )
        })}

        {/* Chart + Stats Row */}
        <div className="grid grid-cols-1 xl:grid-cols-3 gap-6">
          <div className="xl:col-span-2">
            <PriceChart symbol={selected} />
          </div>
          <div>
            <MarketOverview symbol={selected} />
          </div>
        </div>
      </main>
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
