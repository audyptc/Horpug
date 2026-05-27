import { createContext, useContext, useState, useCallback, useMemo, type ReactNode } from 'react'
import { authService } from '@/services/authService'
import { decodeJwt } from '@/lib/utils'

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

function getStoredToken() {
  return localStorage.getItem('access_token')
}

function getUserIdFromToken(token: string | null): string | null {
  if (!token) return null
  const payload = decodeJwt(token)
  return (payload?.user_id as string) ?? null
}

export function AuthProvider({ children }: { children: ReactNode }) {
  const [state, setState] = useState<AuthState>({
    accessToken: getStoredToken(),
    isAuthenticated: !!getStoredToken(),
  })

  const userId = useMemo(() => getUserIdFromToken(state.accessToken), [state.accessToken])

  const login = useCallback(async (email: string, password: string) => {
    const resp = await authService.login(email, password)
    localStorage.setItem('access_token', resp.access_token)
    localStorage.setItem('refresh_token', resp.refresh_token)
    setState({ accessToken: resp.access_token, isAuthenticated: true })
  }, [])

  const logout = useCallback(async () => {
    const refreshToken = localStorage.getItem('refresh_token')
    if (refreshToken) {
      try {
        await authService.logout(refreshToken)
      } catch {
        // ignore
      }
    }
    localStorage.removeItem('access_token')
    localStorage.removeItem('refresh_token')
    setState({ accessToken: null, isAuthenticated: false })
  }, [])

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
