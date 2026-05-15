import { create } from 'zustand'
import type { PriceUpdate } from '@/types'

const WS_URL = process.env.NEXT_PUBLIC_WS_URL ?? 'ws://localhost/ws'

interface WSStore {
  prices: Record<string, PriceUpdate>
  connected: boolean
  connect: () => void
  disconnect: () => void
}

let socket: WebSocket | null = null
let reconnectTimer: ReturnType<typeof setTimeout> | null = null

export const useWSStore = create<WSStore>((set, get) => ({
  prices: {},
  connected: false,

  connect() {
    if (socket?.readyState === WebSocket.OPEN) return

    socket = new WebSocket(`${WS_URL}/prices`)

    socket.onopen = () => {
      set({ connected: true })
      if (reconnectTimer) {
        clearTimeout(reconnectTimer)
        reconnectTimer = null
      }
    }

    socket.onmessage = (event: MessageEvent<string>) => {
      try {
        const update = JSON.parse(event.data) as PriceUpdate
        set(state => ({
          prices: { ...state.prices, [update.symbol]: update },
        }))
      } catch {
        // ignore malformed messages
      }
    }

    socket.onerror = () => {
      socket?.close()
    }

    socket.onclose = () => {
      set({ connected: false })
      // Exponential back-off reconnect
      reconnectTimer = setTimeout(() => {
        if (get().connected === false) get().connect()
      }, 3_000)
    }
  },

  disconnect() {
    if (reconnectTimer) clearTimeout(reconnectTimer)
    socket?.close()
    socket = null
    set({ connected: false })
  },
}))
