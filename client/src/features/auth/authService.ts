import api from '@/lib/axios'
import type { ApiResponse, LoginResponse } from '@/types'

export const authService = {
  async login(email: string, password: string): Promise<LoginResponse> {
    const { data } = await api.post<ApiResponse<LoginResponse>>('/auth/login', { email, password })
    return data.data
  },

  async logout(): Promise<void> {
    await api.post('/auth/logout')
  },
}
