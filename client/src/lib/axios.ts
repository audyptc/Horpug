import axios from 'axios'

const BASE_URL = (import.meta.env.VITE_API_URL ?? 'http://localhost:3003') + '/api/v1'
const ACCESS_TOKEN_EXPIRES_AT_KEY = 'access_token_expires_at'
const SESSION_TIMEOUT_HOURS = Number(import.meta.env.VITE_SESSION_TIMEOUT_HOURS ?? '6')
const SESSION_TIMEOUT_MS = Number.isFinite(SESSION_TIMEOUT_HOURS) && SESSION_TIMEOUT_HOURS > 0
  ? SESSION_TIMEOUT_HOURS * 60 * 60 * 1000
  : 6 * 60 * 60 * 1000

function setAccessTokenExpiresAt(expiresInSeconds: number | undefined) {
  if (!Number.isFinite(expiresInSeconds) || !expiresInSeconds || expiresInSeconds <= 0) {
    return
  }

  const apiExpiresAt = Date.now() + expiresInSeconds * 1000
  const clientTimeoutAt = Date.now() + SESSION_TIMEOUT_MS
  localStorage.setItem(ACCESS_TOKEN_EXPIRES_AT_KEY, String(Math.min(apiExpiresAt, clientTimeoutAt)))
}

const api = axios.create({
  baseURL: BASE_URL,
  timeout: 10000,
  headers: { 'Content-Type': 'application/json' },
})

api.interceptors.request.use((config) => {
  const token = localStorage.getItem('access_token')
  if (token) config.headers.Authorization = `Bearer ${token}`
  return config
})

let isRefreshing = false
let failedQueue: Array<{ resolve: (v: string) => void; reject: (e: unknown) => void }> = []

function processQueue(error: unknown, token: string | null) {
  failedQueue.forEach((p) => (error ? p.reject(error) : p.resolve(token!)))
  failedQueue = []
}

api.interceptors.response.use(
  (res) => res,
  async (err) => {
    const original = err.config
    if (err.response?.status !== 401 || original._retry) {
      return Promise.reject(err)
    }

    const refreshToken = localStorage.getItem('refresh_token')
    if (!refreshToken) {
      localStorage.removeItem('access_token')
      localStorage.removeItem(ACCESS_TOKEN_EXPIRES_AT_KEY)
      window.location.href = '/login'
      return Promise.reject(err)
    }

    if (isRefreshing) {
      return new Promise((resolve, reject) => {
        failedQueue.push({ resolve, reject })
      }).then((token) => {
        original.headers.Authorization = `Bearer ${token}`
        return api(original)
      })
    }

    original._retry = true
    isRefreshing = true

    try {
      const { data } = await axios.post(`${BASE_URL}/auth/refresh`, {
        refresh_token: refreshToken,
      })
      const newToken = data.data.access_token
      localStorage.setItem('access_token', newToken)
      localStorage.setItem('refresh_token', data.data.refresh_token)
      setAccessTokenExpiresAt(Number(data.data.expires_in))
      api.defaults.headers.common.Authorization = `Bearer ${newToken}`
      processQueue(null, newToken)
      original.headers.Authorization = `Bearer ${newToken}`
      return api(original)
    } catch (refreshErr) {
      processQueue(refreshErr, null)
      localStorage.removeItem('access_token')
      localStorage.removeItem('refresh_token')
      localStorage.removeItem(ACCESS_TOKEN_EXPIRES_AT_KEY)
      window.location.href = '/login'
      return Promise.reject(refreshErr)
    } finally {
      isRefreshing = false
    }
  }
)

export default api
