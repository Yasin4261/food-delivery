// Shared helpers for the order status lifecycle (mirrors the backend §4).

const CLASSES = {
  pending: 'bg-gray-100 text-gray-700',
  confirmed: 'bg-blue-100 text-blue-700',
  preparing: 'bg-amber-100 text-amber-700',
  ready: 'bg-indigo-100 text-indigo-700',
  delivering: 'bg-purple-100 text-purple-700',
  delivered: 'bg-green-100 text-green-700',
  cancelled: 'bg-red-100 text-red-700',
}

export function statusClass(status) {
  return CLASSES[status] || 'bg-gray-100 text-gray-700'
}

// The action a chef can take to move an order forward from its current status,
// plus "decline" while still cancellable. Maps to POST /chef/orders/:id/status.
const NEXT_ACTION = {
  pending: { action: 'confirm', label: 'Accept' },
  confirmed: { action: 'preparing', label: 'Start preparing' },
  preparing: { action: 'ready', label: 'Mark ready' },
  ready: { action: 'delivering', label: 'Out for delivery' },
  delivering: { action: 'delivered', label: 'Mark delivered' },
}

export function nextAction(status) {
  return NEXT_ACTION[status] || null
}

export function canDecline(status) {
  return status === 'pending' || status === 'confirmed'
}
