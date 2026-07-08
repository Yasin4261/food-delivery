import { defineStore } from 'pinia'
import { api, page } from '@/api/client'

// Favorited chef ids for the logged-in customer. Loaded lazily once per
// session; toggle talks to the idempotent POST/DELETE endpoints.
export const useFavoritesStore = defineStore('favorites', {
  state: () => ({
    ids: [],
    loaded: false,
  }),
  getters: {
    has: (s) => (chefId) => s.ids.includes(chefId),
    count: (s) => s.ids.length,
  },
  actions: {
    async load(force = false) {
      if (this.loaded && !force) return
      const chefs = page(await api.get('/favorites?limit=100')).items
      this.ids = chefs.map((c) => c.id)
      this.loaded = true
    },
    async toggle(chefId) {
      if (this.has(chefId)) {
        await api.del(`/favorites/${chefId}`)
        this.ids = this.ids.filter((id) => id !== chefId)
      } else {
        await api.post(`/favorites/${chefId}`)
        this.ids.push(chefId)
      }
    },
    reset() {
      this.ids = []
      this.loaded = false
    },
  },
})
