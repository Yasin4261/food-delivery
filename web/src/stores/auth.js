import { defineStore } from 'pinia'
import { api } from '@/api/client'
import { useFavoritesStore } from '@/stores/favorites'
import { useNotificationsStore } from '@/stores/notifications'

export const useAuthStore = defineStore('auth', {
  state: () => ({
    token: localStorage.getItem('token') || '',
    user: JSON.parse(localStorage.getItem('user') || 'null'),
  }),
  getters: {
    isAuthenticated: (s) => !!s.token,
    role: (s) => s.user?.role || null,
    isChef: (s) => s.user?.role === 'chef',
    isAdmin: (s) => s.user?.role === 'admin',
  },
  actions: {
    persist() {
      if (this.token) localStorage.setItem('token', this.token)
      else localStorage.removeItem('token')
      if (this.user) localStorage.setItem('user', JSON.stringify(this.user))
      else localStorage.removeItem('user')
    },
    apply(result) {
      this.token = result.token
      this.user = result.user
      this.persist()
    },
    async login(email, password) {
      this.apply(await api.post('/auth/login', { email, password }))
    },
    async register(input) {
      this.apply(await api.post('/auth/register', input))
    },
    // refresh re-reads the current account (e.g. after email verification flips
    // is_verified) and updates the cached user without touching the token.
    async refresh() {
      if (!this.token) return
      this.user = await api.get('/auth/me')
      this.persist()
    },
    async logout() {
      try {
        if (this.token) await api.post('/auth/logout')
      } catch {
        // best-effort; clear locally regardless
      }
      this.clear()
    },
    clear() {
      this.token = ''
      this.user = null
      this.persist()
      // Per-user caches must not leak into the next session.
      useFavoritesStore().reset()
      useNotificationsStore().stop()
    },
  },
})
