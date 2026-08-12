import { createContext, useContext, useSyncExternalStore, type ReactNode } from 'react'
import { api, extractErrorMessage } from './api'
import { getSession, setSession, subscribe, type Session } from './session'

type AuthContextValue = {
  session: Session | null
  isAuthenticated: boolean
  login: (username: string, password: string) => Promise<void>
  logout: () => Promise<void>
}

const AuthContext = createContext<AuthContextValue | null>(null)

function getServerSnapshot(): Session | null {
  return null
}

export function AuthProvider({ children }: { children: ReactNode }) {
  const session = useSyncExternalStore(subscribe, getSession, getServerSnapshot)

  const login = async (username: string, password: string) => {
    try {
      const { data } = await api.post('/auth/login', { username, password })
      setSession({
        accessToken: data.access_token,
        accessTokenExpiresAt: data.access_token_expires_at,
        refreshToken: data.refresh_token,
        user: data.user,
      })
    } catch (error) {
      throw new Error(extractErrorMessage(error, 'Login failed'))
    }
  }

  const logout = async () => {
    const current = getSession()
    setSession(null)
    if (current?.refreshToken) {
      try {
        await api.post('/auth/logout', { refresh_token: current.refreshToken })
      } catch {
        // Local session is already cleared; the server-side token will simply expire.
      }
    }
  }

  return (
    <AuthContext.Provider value={{ session, isAuthenticated: !!session, login, logout }}>
      {children}
    </AuthContext.Provider>
  )
}

export function useAuth() {
  const context = useContext(AuthContext)
  if (!context) throw new Error('useAuth must be used within AuthProvider')
  return context
}
