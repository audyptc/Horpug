import api from '@/lib/axios'
import type { ApiPaginatedResponse, ApiResponse, ApiRole } from '@/types/api'

export const roleService = {
  async list(): Promise<ApiRole[]> {
    const { data } = await api.get<ApiPaginatedResponse<ApiRole>>('/roles/')
    return data.data
  },

  async listActive(): Promise<ApiRole[]> {
    const { data } = await api.get<ApiResponse<ApiRole[]>>('/roles/active')
    return data.data
  },
}
