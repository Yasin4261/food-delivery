import { defineStore } from 'pinia'
import { api } from '@/api/client'

// Poll cadence for the navbar badges. Ticks are skipped while the tab is
// hidden so background tabs don't hammer the API.
export const POLL_MS = 15000

export const useNotificationsStore = defineStore('notifications', {
  state: () => ({
    activeOrders: 0,
    pendingChefOrders: 0,
    timer: null,
  }),
  actions: {
    async refresh() {
      try {
        const s = await api.get('/notifications/summary')
        this.activeOrders = s.active_orders ?? 0
        this.pendingChefOrders = s.pending_chef_orders ?? 0
      } catch {
        // Polling is best-effort; badges just go stale on errors.
      }
    },
    start() {
      if (this.timer) return
      this.refresh()
      this.timer = setInterval(() => {
        if (!document.hidden) this.refresh()
      }, POLL_MS)
    },
    stop() {
      if (this.timer) clearInterval(this.timer)
      this.timer = null
      this.activeOrders = 0
      this.pendingChefOrders = 0
    },
  },
})
