import axios, { type InternalAxiosRequestConfig } from 'axios'
import { getSession, setSession } from './session'

// Relative base URL: proxied to the backend by Vite in dev (see vite.config.ts)
// and by nginx in production (see ../../nginx/nginx.conf), so the app never
// needs to know the backend's real origin.
export const api = axios.create({ baseURL: '/api/v1' })

api.interceptors.request.use((config) => {
  const session = getSession()
  if (session?.accessToken) {
    config.headers.set('Authorization', `Bearer ${session.accessToken}`)
  }
  return config
})

type RetryableConfig = InternalAxiosRequestConfig & { _retry?: boolean }

let refreshPromise: Promise<string | null> | null = null

function refreshAccessToken(): Promise<string | null> {
  const session = getSession()
  if (!session?.refreshToken) return Promise.resolve(null)

  if (!refreshPromise) {
    refreshPromise = axios
      .post<{
        access_token: string
        access_token_expires_at: string
        refresh_token: string
        refresh_token_expires_at: string
        user: typeof session.user
      }>('/api/v1/auth/refresh', { refresh_token: session.refreshToken })
      .then(({ data }) => {
        setSession({
          accessToken: data.access_token,
          accessTokenExpiresAt: data.access_token_expires_at,
          refreshToken: data.refresh_token,
          user: data.user,
        })
        return data.access_token
      })
      .catch(() => {
        setSession(null)
        return null
      })
      .finally(() => {
        refreshPromise = null
      })
  }

  return refreshPromise
}

api.interceptors.response.use(
  (response) => response,
  async (error) => {
    if (!axios.isAxiosError(error) || !error.config) {
      return Promise.reject(error)
    }

    const originalRequest = error.config as RetryableConfig
    const isAuthEndpoint = originalRequest.url?.includes('/auth/')

    if (error.response?.status === 401 && !originalRequest._retry && !isAuthEndpoint) {
      originalRequest._retry = true
      const newToken = await refreshAccessToken()
      if (newToken) {
        originalRequest.headers.set('Authorization', `Bearer ${newToken}`)
        return api(originalRequest)
      }
    }

    return Promise.reject(error)
  }
)

export function extractErrorMessage(error: unknown, fallback: string): string {
  if (axios.isAxiosError(error)) {
    const data = error.response?.data as { error?: string } | undefined
    if (data?.error) return data.error
  }
  return fallback
}
