'use client'

import { useWSStore } from '@/lib/websocket'
import type { Index } from '@/types'
import numeral from 'numeral'
import { clsx } from 'clsx'

interface Props {
  index: Index
  selected?: boolean
  onClick?: () => void
}

function formatPrice(price: number): string {
  if (price >= 1000) return numeral(price).format('0,0.00')
  if (price >= 1)    return numeral(price).format('0,0.000')
  return numeral(price).format('0.00000000')
}

export default function IndexCard({ index, selected, onClick }: Props) {
  const livePrice = useWSStore(s => s.prices[index.symbol])

  const price         = livePrice?.price          ?? index.price
  const changePct     = livePrice?.change_percent  ?? index.change_percent
  const isPositive    = changePct >= 0

  return (
    <button
      onClick={onClick}
      className={clsx(
        'card text-left w-full transition-all duration-150 hover:border-accent-blue/60 cursor-pointer',
        selected && 'border-accent-blue ring-1 ring-accent-blue/40',
      )}
    >
      {/* Symbol + Category */}
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
