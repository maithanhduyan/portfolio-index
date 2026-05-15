'use client'

import { useEffect, useRef, useState } from 'react'
import { useQuery } from '@tanstack/react-query'
import {
  createChart,
  ColorType,
  CrosshairMode,
  type IChartApi,
  type ISeriesApi,
  type CandlestickData,
} from 'lightweight-charts'
import { indexApi } from '@/lib/api'
import type { ChartInterval } from '@/types'
import { clsx } from 'clsx'

const INTERVALS: ChartInterval[] = ['1m', '5m', '15m', '1h', '4h', '1d', '1w']

interface Props {
  symbol: string
}

export default function PriceChart({ symbol }: Props) {
  const chartContainerRef = useRef<HTMLDivElement>(null)
  const chartRef          = useRef<IChartApi | null>(null)
  const seriesRef         = useRef<ISeriesApi<'Candlestick'> | null>(null)
  const [interval, setInterval] = useState<ChartInterval>('1h')

  const { data: candles = [] } = useQuery({
    queryKey: ['candles', symbol, interval],
    queryFn: () => indexApi.getCandles(symbol, interval, 300),
    staleTime: 30_000,
  })

  // ─── Init chart ─────────────────────────────────────────────
  useEffect(() => {
    if (!chartContainerRef.current) return

    const chart = createChart(chartContainerRef.current, {
      layout: {
        background: { type: ColorType.Solid, color: '#1c2128' },
        textColor:  '#8b949e',
      },
      grid: {
        vertLines:   { color: '#30363d' },
        horzLines:   { color: '#30363d' },
      },
      crosshair: { mode: CrosshairMode.Normal },
      rightPriceScale: { borderColor: '#30363d' },
      timeScale:       { borderColor: '#30363d', timeVisible: true },
      width:  chartContainerRef.current.clientWidth,
      height: 420,
    })

    const series = chart.addCandlestickSeries({
      upColor:     '#3fb950',
      downColor:   '#f85149',
      borderUpColor:   '#3fb950',
      borderDownColor: '#f85149',
      wickUpColor:     '#3fb950',
      wickDownColor:   '#f85149',
    })

    chartRef.current  = chart
    seriesRef.current = series

    // Responsive resize
    const observer = new ResizeObserver(entries => {
      const { width } = entries[0].contentRect
      chart.resize(width, 420)
    })
    observer.observe(chartContainerRef.current)

    return () => {
      observer.disconnect()
      chart.remove()
    }
  }, [])

  // ─── Update data ─────────────────────────────────────────────
  useEffect(() => {
    if (!seriesRef.current || !candles?.length) return

    const data: CandlestickData[] = candles
      .map(c => ({
        time:  Math.floor(new Date(c.time).getTime() / 1000) as CandlestickData['time'],
        open:  c.open,
        high:  c.high,
        low:   c.low,
        close: c.close,
      }))
      .sort((a, b) => (a.time as number) - (b.time as number))

    seriesRef.current.setData(data)
    chartRef.current?.timeScale().fitContent()
  }, [candles])

  return (
    <div className="card space-y-3">
      {/* Toolbar */}
      <div className="flex items-center justify-between">
        <span className="font-mono font-bold text-text-primary">{symbol} / USDT</span>
        <div className="flex gap-1">
          {INTERVALS.map(i => (
            <button
              key={i}
              onClick={() => setInterval(i)}
              className={clsx(
                'px-2 py-1 text-[11px] font-mono rounded transition-colors',
                interval === i
                  ? 'bg-accent-blue text-bg-primary font-bold'
                  : 'text-text-secondary hover:text-text-primary hover:bg-bg-border',
              )}
            >
              {i}
            </button>
          ))}
        </div>
      </div>

      {/* Chart */}
      <div ref={chartContainerRef} className="w-full rounded overflow-hidden" />
    </div>
  )
}
