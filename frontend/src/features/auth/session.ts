export type Role = {
  id: string
  name: string
  description: string
  is_active: boolean
}

export type SessionUser = {
  id: string
  username: string
  email: string
  role_id: string
  role?: Role
  is_active: boolean
  // The seeded admin: its password comes from the ADMIN_PASSWORD setting and
  // can't be changed from the app.
  is_protected?: boolean
}

export type Session = {
  accessToken: string
  accessTokenExpiresAt: string
  user: SessionUser
}

// Deliberately in-memory only: the refresh token lives in an httpOnly cookie
// the browser manages, and the access token is short-lived, so nothing
// session-related needs to (or should) sit in localStorage where any
// injected script could read it. A hard refresh loses this and re-derives it
// via a silent /auth/refresh call — see lib/auth.tsx.
let session: Session | null = null
const listeners = new Set<() => void>()

export function getSession(): Session | null {
  return session
}

export function setSession(next: Session | null) {
  session = next
  listeners.forEach((listener) => listener())
}

export function subscribe(listener: () => void) {
  listeners.add(listener)
  return () => listeners.delete(listener)
}
