import axios from 'axios'

// A separate client from the staff one (shared/api/client): tenants have no
// staff session, and the staff client's 401 handling (refresh, then back to
// the login page) must never run here. The token lives only in memory; a
// reload signs in again silently through LINE.
export const tenantApi = axios.create({ baseURL: '/api/v1' })

let token: string | null = null

export function setTenantToken(next: string | null) {
  token = next
}

tenantApi.interceptors.request.use((config) => {
  if (token) {
    config.headers.set('Authorization', `Bearer ${token}`)
  }
  return config
})
