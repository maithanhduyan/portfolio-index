'use client'

import { useState } from 'react'
import { useMutation, useQueryClient } from '@tanstack/react-query'
import { authApi } from '@/lib/api'
import { useAuthStore } from '@/lib/auth'

interface Props {
  onClose: () => void
}

export default function AuthModal({ onClose }: Props) {
  const [mode, setMode] = useState<'login' | 'register'>('login')
  const [email, setEmail] = useState('')
  const [name, setName] = useState('')
  const [password, setPassword] = useState('')
  const [error, setError] = useState('')
  const login = useAuthStore(s => s.login)
  const qc = useQueryClient()

  const mutation = useMutation({
    mutationFn: async () => {
      if (mode === 'login') return authApi.login(email, password)
      return authApi.register(email, name, password)
    },
    onSuccess: (data) => {
      login(data.token, data.user)
      qc.invalidateQueries({ queryKey: ['watchlist'] })
      qc.invalidateQueries({ queryKey: ['portfolios'] })
      onClose()
    },
    onError: (err: any) => {
      setError(err?.response?.data?.error ?? 'Đã có lỗi xảy ra')
    },
  })

  const submit = (e: React.FormEvent) => {
    e.preventDefault()
    setError('')
    mutation.mutate()
  }

  return (
    <div className="fixed inset-0 z-50 flex items-center justify-center p-4">
      <div className="absolute inset-0 bg-black/70 backdrop-blur-sm" onClick={onClose} />
      <div className="relative w-full max-w-md bg-bg-card border border-bg-border rounded-2xl p-8 shadow-2xl animate-fade-in">
        {/* Logo */}
        <div className="text-center mb-6">
          <span className="text-2xl font-bold text-text-primary">📈 Portfolio Index</span>
          <p className="text-xs text-text-muted mt-1">Theo dõi & quản lý danh mục đầu tư</p>
        </div>

        {/* Tab switcher */}
        <div className="flex rounded-xl bg-bg-primary p-1 mb-6">
          {(['login', 'register'] as const).map((m) => (
            <button
              key={m}
              onClick={() => { setMode(m); setError('') }}
              className={`flex-1 py-2 text-sm font-medium rounded-lg transition-all ${
                mode === m
                  ? 'bg-accent-blue text-white shadow'
                  : 'text-text-secondary hover:text-text-primary'
              }`}
            >
              {m === 'login' ? 'Đăng nhập' : 'Đăng ký'}
            </button>
          ))}
        </div>

        <form onSubmit={submit} className="space-y-4">
          {mode === 'register' && (
            <div>
              <label className="block text-xs text-text-secondary mb-1">Tên hiển thị</label>
              <input
                type="text"
                value={name}
                onChange={e => setName(e.target.value)}
                placeholder="Nguyễn Văn A"
                className="w-full bg-bg-primary border border-bg-border rounded-lg px-4 py-2.5 text-sm text-text-primary placeholder-text-muted focus:outline-none focus:border-accent-blue transition-colors"
                required
              />
            </div>
          )}
          <div>
            <label className="block text-xs text-text-secondary mb-1">Email</label>
            <input
              type="email"
              value={email}
              onChange={e => setEmail(e.target.value)}
              placeholder="email@example.com"
              className="w-full bg-bg-primary border border-bg-border rounded-lg px-4 py-2.5 text-sm text-text-primary placeholder-text-muted focus:outline-none focus:border-accent-blue transition-colors"
              required
            />
          </div>
          <div>
            <label className="block text-xs text-text-secondary mb-1">Mật khẩu</label>
            <input
              type="password"
              value={password}
              onChange={e => setPassword(e.target.value)}
              placeholder="Tối thiểu 8 ký tự"
              className="w-full bg-bg-primary border border-bg-border rounded-lg px-4 py-2.5 text-sm text-text-primary placeholder-text-muted focus:outline-none focus:border-accent-blue transition-colors"
              required
              minLength={8}
            />
          </div>

          {error && (
            <p className="text-accent-red text-xs bg-accent-red/10 border border-accent-red/20 rounded-lg px-3 py-2">
              {error}
            </p>
          )}

          <button
            type="submit"
            disabled={mutation.isPending}
            className="w-full bg-accent-blue hover:bg-accent-blue/80 text-white font-semibold py-2.5 rounded-lg transition-all disabled:opacity-50 disabled:cursor-not-allowed"
          >
            {mutation.isPending ? 'Đang xử lý...' : mode === 'login' ? 'Đăng nhập' : 'Tạo tài khoản'}
          </button>
        </form>
      </div>
    </div>
  )
}
