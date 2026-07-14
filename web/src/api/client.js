// Thin fetch wrapper around the food-delivery API. In dev, requests go to
// /api/v2 and Vite proxies them to the Go backend. The bearer token is read
// from localStorage so the client stays decoupled from the Pinia store.

const BASE = (import.meta.env.VITE_API_BASE || '') + '/api/v2'

export class ApiError extends Error {
  constructor(status, message) {
    super(message)
    this.status = status
  }
}

// onUnauthorized is invoked on a 401 so the app can log the user out.
let onUnauthorized = () => {}
export function setUnauthorizedHandler(fn) {
  onUnauthorized = fn
}

function token() {
  return localStorage.getItem('token') || ''
}

async function request(method, path, body) {
  const headers = { 'Content-Type': 'application/json' }
  const t = token()
  if (t) headers.Authorization = `Bearer ${t}`

  const res = await fetch(BASE + path, {
    method,
    headers,
    body: body !== undefined ? JSON.stringify(body) : undefined,
  })

  if (res.status === 401) onUnauthorized()

  // 204 No Content (e.g. favorites, delete).
  if (res.status === 204) return null

  const data = await res.json().catch(() => ({}))
  if (!res.ok) {
    throw new ApiError(res.status, data.error || `request failed (${res.status})`)
  }
  return data
}

// upload POSTs a single file as multipart form data (field "image") — the
// browser sets the multipart Content-Type with its boundary itself.
async function upload(path, file) {
  const form = new FormData()
  form.append('image', file)
  const headers = {}
  const t = token()
  if (t) headers.Authorization = `Bearer ${t}`

  const res = await fetch(BASE + path, { method: 'POST', headers, body: form })
  if (res.status === 401) onUnauthorized()
  const data = await res.json().catch(() => ({}))
  if (!res.ok) {
    throw new ApiError(res.status, data.error || `upload failed (${res.status})`)
  }
  return data
}

export const api = {
  get: (path) => request('GET', path),
  post: (path, body) => request('POST', path, body),
  put: (path, body) => request('PUT', path, body),
  patch: (path, body) => request('PATCH', path, body),
  del: (path) => request('DELETE', path),
  upload,
}

// page unwraps the standard list envelope { data, limit, offset, total }.
export function page(envelope) {
  return {
    items: envelope?.data ?? [],
    total: envelope?.total ?? 0,
    limit: envelope?.limit ?? 0,
    offset: envelope?.offset ?? 0,
  }
}
