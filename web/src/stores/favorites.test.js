import { beforeEach, describe, expect, it, vi } from 'vitest'
import { createPinia, setActivePinia } from 'pinia'

vi.mock('@/api/client', () => ({
  api: { get: vi.fn(), post: vi.fn(), put: vi.fn(), patch: vi.fn(), del: vi.fn() },
  page: (e) => ({ items: e?.data ?? [], total: e?.total ?? 0, limit: e?.limit ?? 0, offset: e?.offset ?? 0 }),
  ApiError: class ApiError extends Error {},
  setUnauthorizedHandler: vi.fn(),
}))

import { api } from '@/api/client'
import { useFavoritesStore } from './favorites'

describe('favorites store', () => {
  beforeEach(() => {
    setActivePinia(createPinia())
    vi.clearAllMocks()
  })

  it('loads favorited ids once and caches', async () => {
    api.get.mockResolvedValue({ data: [{ id: 4 }, { id: 9 }], total: 2 })
    const fav = useFavoritesStore()

    await fav.load()
    await fav.load() // cached — no second request

    expect(fav.ids).toEqual([4, 9])
    expect(fav.has(9)).toBe(true)
    expect(api.get).toHaveBeenCalledTimes(1)
  })

  it('toggle favorites and unfavorites via the idempotent endpoints', async () => {
    api.post.mockResolvedValue(null)
    api.del.mockResolvedValue(null)
    const fav = useFavoritesStore()

    await fav.toggle(5)
    expect(api.post).toHaveBeenCalledWith('/favorites/5')
    expect(fav.has(5)).toBe(true)

    await fav.toggle(5)
    expect(api.del).toHaveBeenCalledWith('/favorites/5')
    expect(fav.has(5)).toBe(false)
  })

  it('reset empties the cache for the next user', async () => {
    const fav = useFavoritesStore()
    fav.ids = [1]
    fav.loaded = true
    fav.reset()
    expect(fav.ids).toEqual([])
    expect(fav.loaded).toBe(false)
  })
})
