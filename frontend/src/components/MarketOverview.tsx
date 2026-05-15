'use client'

import { useQuery } from '@tanstack/react-query'
import { indexApi } from '@/lib/api'
import { clsx } from 'clsx'
import numeral from 'numeral'

interface Props {
  symbol: string
}

function Stat({ label, value, color }: { label: string; value: string; color?: string }) {
  return (
    <div className="flex items-center justify-between py-1.5 border-b border-bg-border last:border-0">
      <span className="text-xs text-text-secondary">{label}</span>
      <span className={clsx('text-xs font-mono font-medium', color ?? 'text-text-primary')}>
        {value}
      </span>
    </div>
  )
}

function SignalBadge({ signal }: { signal: 'BUY' | 'SELL' | 'NEUTRAL' }) {
  const map = {
    BUY:     'text-accent-green bg-accent-green/10 border-accent-green/30',
    SELL:    'text-accent-red   bg-accent-red/10   border-accent-red/30',
    NEUTRAL: 'text-accent-yellow bg-accent-yellow/10 border-accent-yellow/30',
  }
  return (
    <span className={clsx('text-xs font-bold px-2 py-0.5 rounded border', map[signal])}>
      {signal}
    </span>
  )
}

export default function MarketOverview({ symbol }: Props) {
  const { data: stats } = useQuery({
    queryKey: ['stats', symbol],
    queryFn: () => indexApi.getStats(symbol),
    staleTime: 60_000,
  })

  const { data: ai } = useQuery({
    queryKey: ['ai', symbol],
    queryFn: () => indexApi.getAIAnalysis(symbol),
    staleTime: 120_000,
    retry: false,
  })

  return (
    <div className="space-y-4">
      {/* Stats Panel */}
      <div className="card">
        <h3 className="text-xs font-semibold text-text-secondary uppercase tracking-widest mb-3">
          Thống kê {symbol}
        </h3>
        {stats ? (
          <>
            <Stat label="ATH"        value={numeral(stats.ath).format('0,0.00')} />
            <Stat label="ATL"        value={numeral(stats.atl).format('0,0.00')} />
            <Stat
              label="7 ngày"
              value={`${stats.change_7d >= 0 ? '+' : ''}${stats.change_7d.toFixed(2)}%`}
              color={stats.change_7d >= 0 ? 'text-accent-green' : 'text-accent-red'}
            />
            <Stat
              label="30 ngày"
              value={`${stats.change_30d >= 0 ? '+' : ''}${stats.change_30d.toFixed(2)}%`}
              color={stats.change_30d >= 0 ? 'text-accent-green' : 'text-accent-red'}
            />
            <Stat
              label="1 năm"
              value={`${stats.change_1y >= 0 ? '+' : ''}${stats.change_1y.toFixed(2)}%`}
              color={stats.change_1y >= 0 ? 'text-accent-green' : 'text-accent-red'}
            />
            <Stat label="Biến động 30d" value={`${stats.volatility_30d.toFixed(2)}%`} />
          </>
        ) : (
          <div className="space-y-2">
            {Array.from({ length: 6 }).map((_, i) => (
              <div key={i} className="h-4 bg-bg-border rounded animate-pulse" />
            ))}
          </div>
        )}
      </div>

      {/* AI Analysis Panel */}
      <div className="card">
        <h3 className="text-xs font-semibold text-text-secondary uppercase tracking-widest mb-3">
          Phân tích AI
        </h3>
        {ai ? (
          <div className="space-y-3">
            <div className="flex items-center justify-between">
              <SignalBadge signal={ai.signal} />
              <span className="text-xs text-text-secondary font-mono">
                Tin cậy: <span className="text-text-primary">{ai.confidence}%</span>
              </span>
            </div>
            <Stat label="Xu hướng"    value={ai.trend} />
            <Stat label="RSI"         value={ai.rsi.toFixed(1)} />
            <Stat label="MACD"        value={ai.macd.value.toFixed(4)} />
            <Stat label="Hỗ trợ"      value={numeral(ai.support).format('0,0.00')} />
            <Stat label="Kháng cự"    value={numeral(ai.resistance).format('0,0.00')} />
            <p className="text-[11px] text-text-secondary leading-relaxed mt-2 border-t border-bg-border pt-2">
              {ai.summary}
            </p>
          </div>
        ) : (
          <div className="space-y-2">
            {Array.from({ length: 5 }).map((_, i) => (
              <div key={i} className="h-4 bg-bg-border rounded animate-pulse" />
            ))}
          </div>
        )}
      </div>
    </div>
  )
}
