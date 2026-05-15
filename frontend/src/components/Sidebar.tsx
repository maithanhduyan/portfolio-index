'use client'

import { clsx } from 'clsx'

export type NavTab = 'dashboard' | 'portfolio' | 'watchlist'

interface Props {
  active: NavTab
  onChange: (tab: NavTab) => void
  watchlistCount: number
  portfolioCount: number
}

const navItems: { key: NavTab; icon: string; label: string }[] = [
  { key: 'dashboard', icon: '📊', label: 'Thị trường' },
  { key: 'portfolio', icon: '💼', label: 'Danh mục' },
  { key: 'watchlist', icon: '⭐', label: 'Theo dõi' },
]

export default function Sidebar({ active, onChange, watchlistCount, portfolioCount }: Props) {
  const badges: Record<NavTab, number> = {
    dashboard: 0,
    portfolio: portfolioCount,
    watchlist: watchlistCount,
  }

  return (
    <aside className="w-16 lg:w-56 shrink-0 border-r border-bg-border bg-bg-secondary flex flex-col">
      <nav className="flex-1 pt-4 space-y-1 px-2">
        {navItems.map(({ key, icon, label }) => (
          <button
            key={key}
            onClick={() => onChange(key)}
            className={clsx(
              'w-full flex items-center gap-3 px-3 py-2.5 rounded-xl text-left transition-all group',
              active === key
                ? 'bg-accent-blue/15 text-accent-blue border border-accent-blue/20'
                : 'text-text-secondary hover:text-text-primary hover:bg-bg-card',
            )}
          >
            <span className="text-lg shrink-0">{icon}</span>
            <span className="hidden lg:block text-sm font-medium flex-1">{label}</span>
            {badges[key] > 0 && (
              <span className={clsx(
                'hidden lg:flex items-center justify-center min-w-[20px] h-5 px-1 rounded-full text-[10px] font-bold',
                active === key ? 'bg-accent-blue text-white' : 'bg-bg-border text-text-secondary',
              )}>
                {badges[key]}
              </span>
            )}
          </button>
        ))}
      </nav>

      {/* Bottom branding */}
      <div className="hidden lg:block p-4 border-t border-bg-border">
        <p className="text-[10px] text-text-muted font-mono text-center">Portfolio Index v1.0</p>
      </div>
    </aside>
  )
}
