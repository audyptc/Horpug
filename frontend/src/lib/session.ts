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
}

export type Session = {
  accessToken: string
  accessTokenExpiresAt: string
  refreshToken: string
  user: SessionUser
}

const STORAGE_KEY = 'horpug_session'

function readFromStorage(): Session | null {
  try {
    const raw = localStorage.getItem(STORAGE_KEY)
    return raw ? (JSON.parse(raw) as Session) : null
  } catch {
    return null
  }
}

let session: Session | null = readFromStorage()
const listeners = new Set<() => void>()

export function getSession(): Session | null {
  return session
}

export function setSession(next: Session | null) {
  session = next
  if (next) {
    localStorage.setItem(STORAGE_KEY, JSON.stringify(next))
  } else {
    localStorage.removeItem(STORAGE_KEY)
  }
  listeners.forEach((listener) => listener())
}

export function subscribe(listener: () => void) {
  listeners.add(listener)
  return () => listeners.delete(listener)
}
