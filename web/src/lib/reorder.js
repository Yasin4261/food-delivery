// Repopulate the cart from a past order (#96). Uses CURRENT dish state — a
// dish whose price changed comes back at the new price; one that's inactive,
// unavailable or out of stock is dropped and reported so the customer knows.
//
// No new API: it reads each participating chef's live menu-items and matches
// the past order's lines by menu_item_id, then adds to the cart store.
export async function reorderIntoCart(order, cart, api) {
  const items = order.items ?? []
  const chefIds = [...new Set(items.map((i) => i.chef_id))]

  // Chef display names from the order's sub-orders (fallback to a stub).
  const chefName = {}
  for (const s of order.sub_orders ?? []) chefName[s.chef_id] = s.chef_name

  // Current dishes per chef, keyed by dish id.
  const liveByChef = {}
  await Promise.all(
    chefIds.map(async (id) => {
      try {
        const res = await api.get(`/chefs/${id}/menu-items?limit=100`)
        const list = res?.data ?? res ?? []
        liveByChef[id] = new Map(list.map((d) => [d.id, d]))
      } catch {
        liveByChef[id] = new Map()
      }
    }),
  )

  let added = 0
  const dropped = []
  for (const line of items) {
    const dish = liveByChef[line.chef_id]?.get(line.menu_item_id)
    const inStock = dish && (dish.is_unlimited || (dish.available_quantity ?? 0) >= line.quantity)
    if (!dish || !dish.is_active || !dish.is_available || !inStock) {
      dropped.push(line.item_name)
      continue
    }
    cart.add(dish, { business_name: chefName[line.chef_id] })
    if (line.quantity > 1) cart.setQuantity(dish.id, line.quantity)
    added++
  }
  return { added, dropped }
}
