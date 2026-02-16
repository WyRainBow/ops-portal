export type Role = 'admin' | 'member' | 'user' | 'unknown'

const TOKEN_KEY = 'ops_portal_token'

export function getToken(): string | null {
  if (typeof window === 'undefined') return null
  return window.localStorage.getItem(TOKEN_KEY)
}

export function setToken(token: string) {
  if (typeof window === 'undefined') return
  window.localStorage.setItem(TOKEN_KEY, token)
}

export function clearToken() {
  if (typeof window === 'undefined') return
  window.localStorage.removeItem(TOKEN_KEY)
}

// JWT: base64url header.payload.signature
export function parseJwtRole(token: string | null): Role {
  if (!token) return 'unknown'
  try {
    const payload = token.split('.')[1]
    if (!payload) return 'unknown'
    const json = JSON.parse(atob(payload.replace(/-/g, '+').replace(/_/g, '/')))
    const role = String(json?.role || 'unknown')
    if (role === 'admin' || role === 'member' || role === 'user') return role
    return 'unknown'
  } catch {
    return 'unknown'
  }
}

