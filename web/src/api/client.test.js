import { afterEach, beforeEach, describe, expect, it, vi } from 'vitest'
import { api, page, ApiError, setUnauthorizedHandler } from './client'

// jsonResponse builds a minimal fetch Response stand-in.
function jsonResponse(status, body) {
  return {
    ok: status >= 200 && status < 300,
    status,
    json: () => Promise.resolve(body),
  }
}

describe('api client', () => {
  beforeEach(() => {
    localStorage.clear()
    vi.stubGlobal('fetch', vi.fn())
  })
  afterEach(() => {
    vi.unstubAllGlobals()
    setUnauthorizedHandler(() => {})
  })

  it('sends the bearer token from localStorage', async () => {
    localStorage.setItem('token', 'tok-123')
    fetch.mockResolvedValue(jsonResponse(200, { ok: true }))

    await api.get('/chefs')

    const [url, opts] = fetch.mock.calls[0]
    expect(url).toBe('/api/v2/chefs')
    expect(opts.headers.Authorization).toBe('Bearer tok-123')
  })

  it('omits the Authorization header when logged out', async () => {
    fetch.mockResolvedValue(jsonResponse(200, {}))

    await api.get('/chefs')

    expect(fetch.mock.calls[0][1].headers.Authorization).toBeUndefined()
  })

  it('serialises POST bodies as JSON', async () => {
    fetch.mockResolvedValue(jsonResponse(201, { id: 1 }))

    const out = await api.post('/orders', { payment_method: 'cash' })

    const [, opts] = fetch.mock.calls[0]
    expect(opts.method).toBe('POST')
    expect(JSON.parse(opts.body)).toEqual({ payment_method: 'cash' })
    expect(out).toEqual({ id: 1 })
  })

  it('throws ApiError with the server message on failure', async () => {
    fetch.mockResolvedValue(jsonResponse(422, { error: 'item is out of stock' }))

    await expect(api.post('/orders', {})).rejects.toMatchObject({
      status: 422,
      message: 'item is out of stock',
    })
    await expect(api.post('/orders', {})).rejects.toBeInstanceOf(ApiError)
  })

  it('returns null for 204 No Content', async () => {
    fetch.mockResolvedValue({ ok: true, status: 204, json: () => Promise.reject(new Error('no body')) })

    expect(await api.del('/favorites/1')).toBeNull()
  })

  it('invokes the unauthorized handler on 401', async () => {
    const onUnauthorized = vi.fn()
    setUnauthorizedHandler(onUnauthorized)
    fetch.mockResolvedValue(jsonResponse(401, { error: 'token has been revoked' }))

    await expect(api.get('/auth/me')).rejects.toMatchObject({ status: 401 })
    expect(onUnauthorized).toHaveBeenCalledOnce()
  })

  it('page() unwraps the list envelope and tolerates blanks', () => {
    expect(page({ data: [1, 2], limit: 20, offset: 0, total: 9 })).toEqual({
      items: [1, 2],
      limit: 20,
      offset: 0,
      total: 9,
    })
    expect(page(undefined)).toEqual({ items: [], limit: 0, offset: 0, total: 0 })
  })
})
