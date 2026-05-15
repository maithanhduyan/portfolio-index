'use client'

import { useState } from 'react'
import { useQuery, useMutation, useQueryClient } from '@tanstack/react-query'
import { userApi } from '@/lib/api'
import { clsx } from 'clsx'
import numeral from 'numeral'
import type { Portfolio, Holding } from '@/types'

function fmt(n: number) { return numeral(n).format('0,0.00') }
function fmtPct(n: number) {
  return (n >= 0 ? '+' : '') + numeral(n).format('0.00') + '%'
}

function PnLBadge({ value, percent }: { value: number; percent: number }) {
  const pos = value >= 0
  return (
    <span className={clsx('text-xs font-mono font-semibold', pos ? 'text-accent-green' : 'text-accent-red')}>
      {pos ? '▲' : '▼'} {fmt(Math.abs(value))} ({fmtPct(percent)})
    </span>
  )
}

interface AddHoldingFormProps {
  portfolioId: string
  onDone: () => void
}

function AddHoldingForm({ portfolioId, onDone }: AddHoldingFormProps) {
  const qc = useQueryClient()
  const [symbol, setSymbol] = useState('')
  const [qty, setQty] = useState('')
  const [price, setPrice] = useState('')
  const [buyDate, setBuyDate] = useState('')
  const [note, setNote] = useState('')

  const mutation = useMutation({
    mutationFn: () => userApi.addHolding(portfolioId, {
      symbol: symbol.toUpperCase(),
      quantity: parseFloat(qty),
      avg_price: parseFloat(price),
      buy_date: buyDate || undefined,
      note,
    }),
    onSuccess: () => {
      qc.invalidateQueries({ queryKey: ['portfolios'] })
      onDone()
    },
  })

  return (
    <form
      onSubmit={e => { e.preventDefault(); mutation.mutate() }}
      className="mt-3 bg-bg-primary rounded-xl p-4 space-y-3 border border-bg-border"
    >
      <p className="text-xs font-semibold text-text-secondary uppercase tracking-wider">Thêm vị thế</p>
      <div className="grid grid-cols-2 gap-2">
        <div>
          <label className="text-[10px] text-text-muted">Symbol</label>
          <input value={symbol} onChange={e => setSymbol(e.target.value)}
            placeholder="BTC" required
            className="input-field mt-0.5 text-sm" />
        </div>
        <div>
          <label className="text-[10px] text-text-muted">Số lượng</label>
          <input type="number" value={qty} onChange={e => setQty(e.target.value)}
            placeholder="0.5" required min="0" step="any"
            className="input-field mt-0.5 text-sm" />
        </div>
        <div>
          <label className="text-[10px] text-text-muted">Giá mua TB</label>
          <input type="number" value={price} onChange={e => setPrice(e.target.value)}
            placeholder="50000" required min="0" step="any"
            className="input-field mt-0.5 text-sm" />
        </div>
        <div>
          <label className="text-[10px] text-text-muted">Ngày mua</label>
          <input type="date" value={buyDate} onChange={e => setBuyDate(e.target.value)}
            className="input-field mt-0.5 text-sm" />
        </div>
      </div>
      <div>
        <label className="text-[10px] text-text-muted">Ghi chú</label>
        <input value={note} onChange={e => setNote(e.target.value)}
          placeholder="Ghi chú tùy chọn..."
          className="input-field mt-0.5 text-sm" />
      </div>
      <div className="flex gap-2">
        <button type="submit" disabled={mutation.isPending}
          className="flex-1 bg-accent-blue hover:bg-accent-blue/80 text-white text-sm font-semibold py-2 rounded-lg transition-all disabled:opacity-50">
          {mutation.isPending ? 'Đang lưu...' : 'Lưu'}
        </button>
        <button type="button" onClick={onDone}
          className="px-4 text-sm text-text-secondary hover:text-text-primary rounded-lg border border-bg-border transition-colors">
          Hủy
        </button>
      </div>
    </form>
  )
}

function HoldingRow({ holding, portfolioId }: { holding: Holding; portfolioId: string }) {
  const qc = useQueryClient()
  const remove = useMutation({
    mutationFn: () => userApi.removeHolding(portfolioId, holding.id),
    onSuccess: () => qc.invalidateQueries({ queryKey: ['portfolios'] }),
  })
  const pos = holding.pnl >= 0

  return (
    <div className="flex items-center gap-2 py-2.5 border-b border-bg-border/50 last:border-0 group">
      <div className="flex-1 min-w-0">
        <div className="flex items-baseline gap-2">
          <span className="text-sm font-bold text-text-primary font-mono">{holding.symbol}</span>
          <span className="text-[10px] text-text-muted">×{holding.quantity}</span>
        </div>
        <div className="text-[10px] text-text-muted">
          Giá vốn: {fmt(holding.avg_price)} · Chi phí: {fmt(holding.cost)}
        </div>
      </div>
      <div className="text-right shrink-0">
        <div className="text-sm font-mono font-semibold text-text-primary">{fmt(holding.current_value)}</div>
        <span className={clsx('text-[10px] font-mono', pos ? 'text-accent-green' : 'text-accent-red')}>
          {pos ? '+' : ''}{fmt(holding.pnl)} ({fmtPct(holding.pnl_percent)})
        </span>
      </div>
      <button
        onClick={() => remove.mutate()}
        disabled={remove.isPending}
        className="opacity-0 group-hover:opacity-100 p-1 rounded hover:bg-accent-red/10 text-text-muted hover:text-accent-red transition-all text-xs"
      >✕</button>
    </div>
  )
}

function PortfolioCard({ portfolio }: { portfolio: Portfolio }) {
  const [showAdd, setShowAdd] = useState(false)
  const qc = useQueryClient()
  const pos = portfolio.pnl >= 0

  const deletePort = useMutation({
    mutationFn: () => userApi.deletePortfolio(portfolio.id),
    onSuccess: () => qc.invalidateQueries({ queryKey: ['portfolios'] }),
  })

  return (
    <div className="card mb-4">
      <div className="flex items-start justify-between mb-3">
        <div>
          <h3 className="font-semibold text-text-primary">{portfolio.name}</h3>
          {portfolio.description && (
            <p className="text-xs text-text-muted mt-0.5">{portfolio.description}</p>
          )}
        </div>
        <div className="flex items-center gap-2">
          <span className="text-xs text-text-muted font-mono">{portfolio.currency}</span>
          <button
            onClick={() => deletePort.mutate()}
            className="p-1 rounded hover:bg-accent-red/10 text-text-muted hover:text-accent-red transition-all text-xs"
            title="Xóa danh mục"
          >🗑</button>
        </div>
      </div>

      {/* Portfolio summary */}
      <div className="grid grid-cols-3 gap-2 mb-3 bg-bg-primary rounded-xl p-3">
        <div className="text-center">
          <p className="text-[10px] text-text-muted">Tổng giá trị</p>
          <p className="text-sm font-mono font-bold text-text-primary">{fmt(portfolio.total_value)}</p>
        </div>
        <div className="text-center border-x border-bg-border">
          <p className="text-[10px] text-text-muted">Chi phí</p>
          <p className="text-sm font-mono font-bold text-text-primary">{fmt(portfolio.total_cost)}</p>
        </div>
        <div className="text-center">
          <p className="text-[10px] text-text-muted">P&L</p>
          <PnLBadge value={portfolio.pnl} percent={portfolio.pnl_percent} />
        </div>
      </div>

      {/* Holdings list */}
      {portfolio.holdings?.length > 0 ? (
        <div className="mb-2">
          {portfolio.holdings.map(h => (
            <HoldingRow key={h.id} holding={h} portfolioId={portfolio.id} />
          ))}
        </div>
      ) : (
        <p className="text-xs text-text-muted text-center py-3">Chưa có vị thế nào</p>
      )}

      {showAdd
        ? <AddHoldingForm portfolioId={portfolio.id} onDone={() => setShowAdd(false)} />
        : (
          <button
            onClick={() => setShowAdd(true)}
            className="w-full mt-2 py-2 text-xs text-accent-blue hover:bg-accent-blue/10 border border-dashed border-accent-blue/30 rounded-xl transition-all"
          >
            + Thêm vị thế
          </button>
        )}
    </div>
  )
}

export default function PortfolioPanel() {
  const { data: portfolios = [], isLoading } = useQuery({
    queryKey: ['portfolios'],
    queryFn: userApi.getPortfolios,
  })
  const [showCreate, setShowCreate] = useState(false)
  const [newName, setNewName] = useState('')
  const [newDesc, setNewDesc] = useState('')
  const [newCurrency, setNewCurrency] = useState('USD')
  const qc = useQueryClient()

  const createPort = useMutation({
    mutationFn: () => userApi.createPortfolio(newName, newDesc, newCurrency),
    onSuccess: () => {
      qc.invalidateQueries({ queryKey: ['portfolios'] })
      setNewName(''); setNewDesc(''); setShowCreate(false)
    },
  })

  if (isLoading) {
    return (
      <div className="space-y-3 p-4">
        {[1, 2].map(i => <div key={i} className="card animate-pulse h-40" />)}
      </div>
    )
  }

  return (
    <div className="p-4 overflow-y-auto h-full">
      <div className="flex items-center justify-between mb-4">
        <h2 className="text-sm font-semibold text-text-secondary uppercase tracking-widest">
          Danh mục của tôi
        </h2>
        <button
          onClick={() => setShowCreate(v => !v)}
          className="text-xs bg-accent-blue hover:bg-accent-blue/80 text-white px-3 py-1.5 rounded-lg transition-all"
        >
          + Tạo mới
        </button>
      </div>

      {showCreate && (
        <form
          onSubmit={e => { e.preventDefault(); createPort.mutate() }}
          className="card mb-4 space-y-3 border-accent-blue/30 animate-fade-in"
        >
          <p className="text-xs font-semibold text-accent-blue">Danh mục mới</p>
          <input value={newName} onChange={e => setNewName(e.target.value)}
            placeholder="Tên danh mục (e.g. Portfolio cá nhân)" required
            className="input-field text-sm" />
          <input value={newDesc} onChange={e => setNewDesc(e.target.value)}
            placeholder="Mô tả ngắn (tùy chọn)"
            className="input-field text-sm" />
          <div className="flex gap-2 items-center">
            <label className="text-xs text-text-muted">Tiền tệ:</label>
            <select value={newCurrency} onChange={e => setNewCurrency(e.target.value)}
              className="input-field text-sm flex-1">
              <option value="USD">USD</option>
              <option value="VND">VND</option>
              <option value="BTC">BTC</option>
            </select>
          </div>
          <div className="flex gap-2">
            <button type="submit" disabled={createPort.isPending}
              className="flex-1 bg-accent-blue hover:bg-accent-blue/80 text-white text-sm font-semibold py-2 rounded-lg transition-all">
              Tạo
            </button>
            <button type="button" onClick={() => setShowCreate(false)}
              className="px-4 text-sm text-text-secondary border border-bg-border rounded-lg hover:text-text-primary">
              Hủy
            </button>
          </div>
        </form>
      )}

      {portfolios.length === 0 && !showCreate && (
        <div className="text-center py-12 text-text-muted">
          <p className="text-3xl mb-3">💼</p>
          <p className="text-sm">Chưa có danh mục nào</p>
          <p className="text-xs mt-1">Tạo danh mục để theo dõi P&L của bạn</p>
        </div>
      )}

      {portfolios.map(p => <PortfolioCard key={p.id} portfolio={p} />)}
    </div>
  )
}
