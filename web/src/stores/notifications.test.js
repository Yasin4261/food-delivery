import { afterEach, beforeEach, describe, expect, it, vi } from 'vitest'
import { createPinia, setActivePinia } from 'pinia'

vi.mock('@/api/client', () => ({
  api: { get: vi.fn(), post: vi.fn(), put: vi.fn(), patch: vi.fn(), del: vi.fn() },
  page: (e) => ({ items: e?.data ?? [], total: e?.total ?? 0, limit: e?.limit ?? 0, offset: e?.offset ?? 0 }),
  ApiError: class ApiError extends Error {},
  setUnauthorizedHandler: vi.fn(),
}))

import { api } from '@/api/client'
import { useNotificationsStore, POLL_MS } from './notifications'

describe('notifications store', () => {
  beforeEach(() => {
    setActivePinia(createPinia())
    vi.clearAllMocks()
    vi.useFakeTimers()
  })
  afterEach(() => {
    vi.useRealTimers()
  })

  it('refresh pulls the summary counts', async () => {
    api.get.mockResolvedValue({ active_orders: 2, pending_chef_orders: 3 })
    const n = useNotificationsStore()

    await n.refresh()

    expect(api.get).toHaveBeenCalledWith('/notifications/summary')
    expect(n.activeOrders).toBe(2)
    expect(n.pendingChefOrders).toBe(3)
  })

  it('refresh is best-effort: API errors leave the counts alone', async () => {
    const n = useNotificationsStore()
    n.activeOrders = 1
    api.get.mockRejectedValue(new Error('down'))

    await n.refresh()

    expect(n.activeOrders).toBe(1)
  })

  it('start polls on an interval and dedupes; stop clears and resets', async () => {
    api.get.mockResolvedValue({ active_orders: 1, pending_chef_orders: 0 })
    const n = useNotificationsStore()

    n.start()
    n.start() // second start must not add a second timer
    expect(api.get).toHaveBeenCalledTimes(1) // immediate refresh

    await vi.advanceTimersByTimeAsync(POLL_MS)
    expect(api.get).toHaveBeenCalledTimes(2)
    await vi.advanceTimersByTimeAsync(POLL_MS)
    expect(api.get).toHaveBeenCalledTimes(3)

    n.stop()
    expect(n.activeOrders).toBe(0)
    await vi.advanceTimersByTimeAsync(POLL_MS * 2)
    expect(api.get).toHaveBeenCalledTimes(3) // no more polling after stop
  })
})
