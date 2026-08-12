import {
  createContext,
  useContext,
  useEffect,
  useState,
  useSyncExternalStore,
  type ReactNode,
} from 'react'
import { api, extractErrorMessage, refreshAccessToken } from '@/shared/api/client'
import { getSession, setSession, subscribe, type Session } from './session'

type AuthContextValue = {
  session: Session | null
  isAuthenticated: boolean
  isLoading: boolean
  login: (username: string, password: string) => Promise<void>
  logout: () => Promise<void>
}

const AuthContext = createContext<AuthContextValue | null>(null)

function getServerSnapshot(): Session | null {
  return null
}

export function AuthProvider({ children }: { children: ReactNode }) {
  const session = useSyncExternalStore(subscribe, getSession, getServerSnapshot)
  const [isLoading, setIsLoading] = useState(true)

  // The access token lives only in memory (see session.ts), so a hard
  // refresh starts with none. Trade the httpOnly refresh cookie — which
  // survives the refresh — for a new one before rendering protected routes.
  useEffect(() => {
    refreshAccessToken().finally(() => setIsLoading(false))
  }, [])

  const login = async (username: string, password: string) => {
    try {
      const { data } = await api.post('/auth/login', { username, password })
      setSession({
        accessToken: data.access_token,
        accessTokenExpiresAt: data.access_token_expires_at,
        user: data.user,
      })
    } catch (error) {
      throw new Error(extractErrorMessage(error, 'Login failed'))
    }
  }

  const logout = async () => {
    setSession(null)
    try {
      await api.post('/auth/logout')
    } catch {
      // Local session is already cleared; the server-side token will simply expire.
    }
  }

  return (
    <AuthContext.Provider value={{ session, isAuthenticated: !!session, isLoading, login, logout }}>
      {children}
    </AuthContext.Provider>
  )
}

export function useAuth() {
  const context = useContext(AuthContext)
  if (!context) throw new Error('useAuth must be used within AuthProvider')
  return context
}
