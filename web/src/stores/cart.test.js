import { beforeEach, describe, expect, it } from 'vitest'
import { createPinia, setActivePinia } from 'pinia'
import { useCartStore } from './cart'

const soup = { id: 1, name: 'Lentil Soup', price: 6.5, chef_id: 2 }
const pilaf = { id: 2, name: 'Plot Pilaf', price: 7.25, chef_id: 3 }
const chefA = { business_name: 'Live Kitchen' }

describe('cart store', () => {
  beforeEach(() => {
    localStorage.clear()
    setActivePinia(createPinia())
  })

  it('adds items and increments quantity on repeat adds', () => {
    const cart = useCartStore()
    cart.add(soup, chefA)
    cart.add(soup, chefA)

    expect(cart.lines).toHaveLength(1)
    expect(cart.lines[0].quantity).toBe(2)
    expect(cart.count).toBe(2)
  })

  it('computes the total and groups lines per chef (multi-chef cart)', () => {
    const cart = useCartStore()
    cart.add(soup, chefA)
    cart.add(pilaf) // no chef object -> fallback label

    expect(cart.total).toBeCloseTo(6.5 + 7.25)
    expect(cart.byChef).toHaveLength(2)
    const groups = Object.fromEntries(cart.byChef.map((g) => [g.chefId, g]))
    expect(groups[2].chefName).toBe('Live Kitchen')
    expect(groups[3].chefName).toBe('Chef #3')
  })

  it('clamps quantities to a minimum of 1', () => {
    const cart = useCartStore()
    cart.add(soup, chefA)
    cart.setQuantity(1, 0)
    expect(cart.lines[0].quantity).toBe(1)
    cart.setQuantity(1, 5)
    expect(cart.lines[0].quantity).toBe(5)
  })

  it('removes lines and clears', () => {
    const cart = useCartStore()
    cart.add(soup, chefA)
    cart.add(pilaf)
    cart.remove(1)
    expect(cart.lines.map((l) => l.menuItemId)).toEqual([2])
    cart.clear()
    expect(cart.count).toBe(0)
  })

  it('persists to localStorage so a reload keeps the cart', () => {
    const cart = useCartStore()
    cart.add(soup, chefA)

    // A brand-new pinia (≈ page reload) rehydrates from localStorage.
    setActivePinia(createPinia())
    const reloaded = useCartStore()
    expect(reloaded.lines).toHaveLength(1)
    expect(reloaded.lines[0].name).toBe('Lentil Soup')
  })
})
