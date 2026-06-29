import { defineStore } from 'pinia'

// The cart supports items from several chefs at once (multi-chef order). Each
// line records the dish, its chef, and the quantity. Lines are keyed by dish id.
export const useCartStore = defineStore('cart', {
  state: () => ({
    lines: JSON.parse(localStorage.getItem('cart') || '[]'),
  }),
  getters: {
    count: (s) => s.lines.reduce((n, l) => n + l.quantity, 0),
    total: (s) => s.lines.reduce((sum, l) => sum + l.price * l.quantity, 0),
    // Group lines by chef for display.
    byChef: (s) => {
      const groups = {}
      for (const l of s.lines) {
        groups[l.chefId] ??= { chefId: l.chefId, chefName: l.chefName, lines: [] }
        groups[l.chefId].lines.push(l)
      }
      return Object.values(groups)
    },
  },
  actions: {
    persist() {
      localStorage.setItem('cart', JSON.stringify(this.lines))
    },
    add(item, chef) {
      const existing = this.lines.find((l) => l.menuItemId === item.id)
      if (existing) existing.quantity += 1
      else
        this.lines.push({
          menuItemId: item.id,
          name: item.name,
          price: item.price,
          chefId: item.chef_id,
          chefName: chef?.business_name || `Chef #${item.chef_id}`,
          quantity: 1,
        })
      this.persist()
    },
    setQuantity(menuItemId, quantity) {
      const line = this.lines.find((l) => l.menuItemId === menuItemId)
      if (!line) return
      line.quantity = Math.max(1, quantity)
      this.persist()
    },
    remove(menuItemId) {
      this.lines = this.lines.filter((l) => l.menuItemId !== menuItemId)
      this.persist()
    },
    clear() {
      this.lines = []
      this.persist()
    },
  },
})
