import { describe, it, expect, vi } from 'vitest'
import { reorderIntoCart } from './reorder'

// A minimal cart stub recording add/setQuantity calls.
function fakeCart() {
  const lines = []
  return {
    lines,
    add(dish, chef) {
      lines.push({ id: dish.id, name: dish.name, price: dish.price, chef: chef?.business_name, quantity: 1 })
    },
    setQuantity(id, q) {
      lines.find((l) => l.id === id).quantity = q
    },
  }
}

// api.get returns the live menu-items list for a chef.
function fakeApi(byChef) {
  return {
    get: vi.fn(async (path) => {
      const id = Number(path.match(/\/chefs\/(\d+)\/menu-items/)[1])
      return { data: byChef[id] ?? [] }
    }),
  }
}

const order = {
  items: [
    { menu_item_id: 1, item_name: 'Soup', chef_id: 10, quantity: 2 },
    { menu_item_id: 2, item_name: 'Kebab', chef_id: 10, quantity: 1 },
    { menu_item_id: 3, item_name: 'Baklava', chef_id: 20, quantity: 1 },
  ],
  sub_orders: [
    { chef_id: 10, chef_name: 'Kitchen A' },
    { chef_id: 20, chef_name: 'Kitchen B' },
  ],
}

describe('reorderIntoCart', () => {
  it('adds available dishes at current price and quantity', async () => {
    const cart = fakeCart()
    const api = fakeApi({
      10: [
        { id: 1, name: 'Soup', price: 7, chef_id: 10, is_active: true, is_available: true, is_unlimited: true },
        { id: 2, name: 'Kebab', price: 15, chef_id: 10, is_active: true, is_available: true, available_quantity: 5 },
      ],
      20: [{ id: 3, name: 'Baklava', price: 9, chef_id: 20, is_active: true, is_available: true, is_unlimited: true }],
    })

    const { added, dropped } = await reorderIntoCart(order, cart, api)
    expect(added).toBe(3)
    expect(dropped).toEqual([])
    const soup = cart.lines.find((l) => l.id === 1)
    expect(soup.quantity).toBe(2) // original quantity restored
    expect(soup.price).toBe(7) // current price
    expect(soup.chef).toBe('Kitchen A')
  })

  it('drops dishes that are gone, inactive, unavailable or out of stock', async () => {
    const cart = fakeCart()
    const api = fakeApi({
      10: [
        // Soup is now inactive; Kebab only has 0 in stock (< requested 1... here qty 1, stock 0).
        { id: 1, name: 'Soup', price: 7, chef_id: 10, is_active: false, is_available: true, is_unlimited: true },
        { id: 2, name: 'Kebab', price: 15, chef_id: 10, is_active: true, is_available: true, available_quantity: 0 },
      ],
      20: [], // Baklava's chef has no dishes anymore
    })

    const { added, dropped } = await reorderIntoCart(order, cart, api)
    expect(added).toBe(0)
    expect(dropped.sort()).toEqual(['Baklava', 'Kebab', 'Soup'])
  })

  it('drops when requested quantity exceeds remaining stock', async () => {
    const cart = fakeCart()
    const api = fakeApi({
      10: [
        { id: 1, name: 'Soup', price: 7, chef_id: 10, is_active: true, is_available: true, available_quantity: 1 }, // want 2
        { id: 2, name: 'Kebab', price: 15, chef_id: 10, is_active: true, is_available: true, is_unlimited: true },
      ],
      20: [{ id: 3, name: 'Baklava', price: 9, chef_id: 20, is_active: true, is_available: true, is_unlimited: true }],
    })

    const { added, dropped } = await reorderIntoCart(order, cart, api)
    expect(added).toBe(2) // Kebab + Baklava
    expect(dropped).toEqual(['Soup'])
  })
})
