import { createContext, useContext, useState, useCallback, useMemo, useEffect, type ReactNode } from 'react'
import { authService } from '@/features/auth/authService'
import { decodeJwt } from '@/lib/utils'
import { SESSION_TIMEOUT_MS, IDLE_TIMEOUT_MS } from '@/lib/authConstants'

interface AuthState {
  accessToken: string | null
  isAuthenticated: boolean
}

interface AuthContextValue extends AuthState {
  userId: string | null
  login: (email: string, password: string) => Promise<void>
  logout: () => Promise<void>
}

const AuthContext = createContext<AuthContextValue | null>(null)

const ACCESS_TOKEN_KEY = 'access_token'
const ACCESS_TOKEN_EXPIRES_AT_KEY = 'access_token_expires_at'
const LAST_ACTIVITY_AT_KEY = 'last_activity_at'

function getStoredToken() {
  return localStorage.getItem(ACCESS_TOKEN_KEY)
}

function clearStoredAuth() {
  localStorage.removeItem(ACCESS_TOKEN_KEY)
  localStorage.removeItem(ACCESS_TOKEN_EXPIRES_AT_KEY)
  localStorage.removeItem(LAST_ACTIVITY_AT_KEY)
}

function touchLastActivity() {
  localStorage.setItem(LAST_ACTIVITY_AT_KEY, String(Date.now()))
}

function getLastActivityAtMs(): number | null {
  const raw = localStorage.getItem(LAST_ACTIVITY_AT_KEY)
  if (!raw) return null
  const parsed = Number(raw)
  return Number.isFinite(parsed) && parsed > 0 ? parsed : null
}

function isIdleExpired(): boolean {
  const lastActivityAt = getLastActivityAtMs()
  if (!lastActivityAt) return true  // Fix 3: ถ้าไม่มีค่าให้ถือว่า expired
  return Date.now() - lastActivityAt >= IDLE_TIMEOUT_MS
}

function setStoredExpiresAt(expiresInSeconds: number) {
  if (!Number.isFinite(expiresInSeconds) || expiresInSeconds <= 0) {
    return
  }
  const apiExpiresAt = Date.now() + expiresInSeconds * 1000
  const clientTimeoutAt = Date.now() + SESSION_TIMEOUT_MS
  localStorage.setItem(ACCESS_TOKEN_EXPIRES_AT_KEY, String(Math.min(apiExpiresAt, clientTimeoutAt)))
}

function getTokenExpFromJwtMs(token: string | null): number | null {
  if (!token) return null
  const payload = decodeJwt(token)
  const exp = payload?.exp
  if (typeof exp !== 'number') return null
  return exp * 1000
}

function getStoredExpiresAtMs(): number | null {
  const raw = localStorage.getItem(ACCESS_TOKEN_EXPIRES_AT_KEY)
  if (!raw) return null
  const parsed = Number(raw)
  return Number.isFinite(parsed) && parsed > 0 ? parsed : null
}

function isExpired(token: string | null): boolean {
  if (!token) return true
  const storedExpiry = getStoredExpiresAtMs()
  const jwtExpiry = getTokenExpFromJwtMs(token)
  const effectiveExpiry = storedExpiry ?? jwtExpiry
  if (!effectiveExpiry) return false
  return Date.now() >= effectiveExpiry
}

function getUserIdFromToken(token: string | null): string | null {
  if (!token) return null
  const payload = decodeJwt(token)
  return (payload?.user_id as string) ?? null
}

export function AuthProvider({ children }: { children: ReactNode }) {
  const initialToken = getStoredToken()
  const validInitialToken = initialToken && !isExpired(initialToken) && !isIdleExpired() ? initialToken : null
  if (initialToken && !validInitialToken) {
    clearStoredAuth()
  }
  if (validInitialToken && !getLastActivityAtMs()) {
    touchLastActivity()
  }

  const [state, setState] = useState<AuthState>({
    accessToken: validInitialToken,
    isAuthenticated: !!validInitialToken,
  })

  const userId = useMemo(() => getUserIdFromToken(state.accessToken), [state.accessToken])

  const login = useCallback(async (email: string, password: string) => {
    const resp = await authService.login(email, password)
    // Fix 1: refresh_token ถูก set เป็น HttpOnly cookie โดย server แล้ว ไม่ต้อง store ใน localStorage
    localStorage.setItem(ACCESS_TOKEN_KEY, resp.access_token)
    setStoredExpiresAt(resp.expires_in)
    touchLastActivity()
    setState({ accessToken: resp.access_token, isAuthenticated: true })
  }, [])

  const logout = useCallback(async () => {
    // Fix 1: server อ่าน refresh_token จาก cookie โดยตรง ไม่ต้องส่งใน body
    try {
      await authService.logout()
    } catch {
      // ignore
    }
    clearStoredAuth()
    setState({ accessToken: null, isAuthenticated: false })
  }, [])

  useEffect(() => {
    if (!state.isAuthenticated) return

    const activityEvents: Array<keyof WindowEventMap> = ['mousedown', 'keydown', 'scroll', 'touchstart', 'mousemove']
    let lastRecordedAt = 0

    const onUserActivity = () => {
      const now = Date.now()
      if (now - lastRecordedAt < 15000) {
        return
      }
      lastRecordedAt = now
      touchLastActivity()
    }

    activityEvents.forEach((event) => window.addEventListener(event, onUserActivity, { passive: true }))
    touchLastActivity()

    const checkSessionExpiry = () => {
      const token = localStorage.getItem(ACCESS_TOKEN_KEY)
      if (!token || isExpired(token) || isIdleExpired()) {
        clearStoredAuth()
        setState({ accessToken: null, isAuthenticated: false })
        if (window.location.pathname !== '/login') {
          window.location.href = '/login'
        }
      }
    }

    checkSessionExpiry()
    const intervalId = window.setInterval(checkSessionExpiry, 30000)
    return () => {
      window.clearInterval(intervalId)
      activityEvents.forEach((event) => window.removeEventListener(event, onUserActivity))
    }
  }, [state.isAuthenticated])

  return (
    <AuthContext.Provider value={{ ...state, userId, login, logout }}>
      {children}
    </AuthContext.Provider>
  )
}

export function useAuth() {
  const ctx = useContext(AuthContext)
  if (!ctx) throw new Error('useAuth must be used inside AuthProvider')
  return ctx
}
