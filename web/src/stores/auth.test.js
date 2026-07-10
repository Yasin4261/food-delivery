import { beforeEach, describe, expect, it, vi } from 'vitest'
import { createPinia, setActivePinia } from 'pinia'

vi.mock('@/api/client', () => ({
  api: { get: vi.fn(), post: vi.fn(), put: vi.fn(), patch: vi.fn(), del: vi.fn() },
  page: (e) => ({ items: e?.data ?? [], total: e?.total ?? 0, limit: e?.limit ?? 0, offset: e?.offset ?? 0 }),
  ApiError: class ApiError extends Error {},
  setUnauthorizedHandler: vi.fn(),
}))

import { api } from '@/api/client'
import { useAuthStore } from './auth'
import { useFavoritesStore } from './favorites'

describe('auth store', () => {
  beforeEach(() => {
    localStorage.clear()
    setActivePinia(createPinia())
    vi.clearAllMocks()
  })

  it('login stores the session and persists it', async () => {
    api.post.mockResolvedValue({ token: 'jwt-1', user: { id: 7, email: 'a@b.c', role: 'chef' } })

    const auth = useAuthStore()
    await auth.login('a@b.c', 'secret')

    expect(auth.isAuthenticated).toBe(true)
    expect(auth.isChef).toBe(true)
    expect(localStorage.getItem('token')).toBe('jwt-1')
    expect(JSON.parse(localStorage.getItem('user')).id).toBe(7)
  })

  it('logout clears the session even when the API call fails, and resets per-user caches', async () => {
    api.post.mockResolvedValueOnce({ token: 'jwt-1', user: { id: 7, role: 'customer' } })
    const auth = useAuthStore()
    await auth.login('a@b.c', 'secret')

    const favorites = useFavoritesStore()
    favorites.ids = [1, 2]
    favorites.loaded = true

    api.post.mockRejectedValueOnce(new Error('network down'))
    await auth.logout()

    expect(auth.isAuthenticated).toBe(false)
    expect(localStorage.getItem('token')).toBeNull()
    expect(favorites.ids).toEqual([])
    expect(favorites.loaded).toBe(false)
  })

  it('rehydrates from localStorage on store creation (page reload)', () => {
    localStorage.setItem('token', 'jwt-2')
    localStorage.setItem('user', JSON.stringify({ id: 3, role: 'customer' }))

    setActivePinia(createPinia())
    const auth = useAuthStore()
    expect(auth.isAuthenticated).toBe(true)
    expect(auth.role).toBe('customer')
  })
})
