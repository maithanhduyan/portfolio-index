'use client'

import { useMutation, useQueryClient } from '@tanstack/react-query'
import { useWSStore } from '@/lib/websocket'
import { useAuthStore } from '@/lib/auth'
import { userApi } from '@/lib/api'
import type { Index, WatchlistItem } from '@/types'
import numeral from 'numeral'
import { clsx } from 'clsx'

interface Props {
  index: Index
  selected?: boolean
  onClick?: () => void
  watchlist?: WatchlistItem[]
}

function formatPrice(price: number): string {
  if (price >= 1000) return numeral(price).format('0,0.00')
  if (price >= 1)    return numeral(price).format('0,0.000')
  return numeral(price).format('0.00000000')
}

export default function IndexCard({ index, selected, onClick, watchlist = [] }: Props) {
  const livePrice = useWSStore(s => s.prices[index.symbol])
  const isAuth = useAuthStore(s => s.isAuthenticated())
  const qc = useQueryClient()

  const price     = livePrice?.price          ?? index.price
  const changePct = livePrice?.change_percent  ?? index.change_percent
  const isPositive = changePct >= 0
  const isWatched  = watchlist.some(w => w.symbol === index.symbol)

  const toggleWatch = useMutation({
    mutationFn: () => isWatched
      ? userApi.removeWatchlist(index.symbol)
      : userApi.addWatchlist(index.symbol),
    onSuccess: () => qc.invalidateQueries({ queryKey: ['watchlist'] }),
  })

  return (
    <button
      onClick={onClick}
      className={clsx(
        'card text-left w-full transition-all duration-150 hover:border-accent-blue/60 cursor-pointer relative',
        selected && 'border-accent-blue ring-1 ring-accent-blue/40',
      )}
    >
      {/* Watchlist star */}
      {isAuth && (
        <button
          onClick={e => { e.stopPropagation(); toggleWatch.mutate() }}
          className={clsx(
            'absolute top-2 right-2 text-sm transition-all hover:scale-125',
            isWatched ? 'text-accent-yellow' : 'text-text-muted opacity-0 group-hover:opacity-100',
          )}
          title={isWatched ? 'Bỏ theo dõi' : 'Theo dõi'}
        >
          {isWatched ? '★' : '☆'}
        </button>
      )}

      {/* Symbol + Change */}
      <div className="flex items-start justify-between mb-2">
        <div>
          <span className="text-xs font-mono font-bold text-text-primary">
            {index.symbol}
          </span>
          <p className="text-[10px] text-text-muted truncate max-w-[100px]">
            {index.name}
          </p>
        </div>
        <span
          className={clsx(
            'text-[10px] font-medium px-1.5 py-0.5 rounded',
            isAuth ? 'mr-5' : '',
            isPositive
              ? 'text-accent-green bg-accent-green/10'
              : 'text-accent-red bg-accent-red/10',
          )}
        >
          {isPositive ? '+' : ''}{changePct.toFixed(2)}%
        </span>
      </div>

      {/* Price */}
      <p className="font-mono text-sm font-semibold text-text-primary tabular-nums">
        {formatPrice(price)}
      </p>

      {/* 24h Range */}
      <div className="mt-2 flex items-center gap-1 text-[10px] text-text-muted font-mono">
        <span className="text-accent-red">{formatPrice(index.low_24h)}</span>
        <div className="flex-1 h-0.5 bg-bg-border rounded-full overflow-hidden">
          <div
            className="h-full bg-gradient-to-r from-accent-red to-accent-green"
            style={{
              width: `${
                index.high_24h > index.low_24h
                  ? ((price - index.low_24h) / (index.high_24h - index.low_24h)) * 100
                  : 50
              }%`,
            }}
          />
        </div>
        <span className="text-accent-green">{formatPrice(index.high_24h)}</span>
      </div>
    </button>
  )
}
