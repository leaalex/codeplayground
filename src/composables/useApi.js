import { useAuth } from './useAuth'

const API_BASE = '/api'

export async function api(path, options = {}) {
  const { token } = useAuth()
  const headers = {
    'Content-Type': 'application/json',
    ...options.headers,
  }
  if (token.value) {
    headers['Authorization'] = `Bearer ${token.value}`
  }
  const res = await fetch(API_BASE + path, {
    ...options,
    headers,
  })
  if (res.status === 401) {
    const { setToken } = useAuth()
    setToken(null)
    const redirect = `${window.location.pathname}${window.location.search}${window.location.hash}`
    const loginUrl =
      redirect && redirect !== '/login'
        ? `/login?redirect=${encodeURIComponent(redirect)}`
        : '/login'
    window.location.href = loginUrl
    throw new Error('Unauthorized')
  }
  if (!res.ok) {
    const err = await res.json().catch(() => ({ error: res.statusText }))
    throw new Error(err.error || 'Request failed')
  }
  if (res.status === 204) {
    return null
  }
  return res.json()
}
