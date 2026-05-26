import api from '@/lib/axios'
import type {
  ApiPaginatedResponse,
  ApiResponse,
  ApiRole,
  CreateRolePayload,
  UpdateRolePayload,
} from '@/types/api'

export const roleService = {
  async list(page = 1, perPage = 20): Promise<ApiPaginatedResponse<ApiRole>> {
    const offset = (page - 1) * perPage
    const { data } = await api.get<ApiPaginatedResponse<ApiRole>>(
      `/roles/?limit=${perPage}&offset=${offset}`
    )
    return data
  },

  async listActive(): Promise<ApiRole[]> {
    const { data } = await api.get<ApiResponse<ApiRole[]>>('/roles/active')
    return data.data
  },

  async getById(id: string): Promise<ApiRole> {
    const { data } = await api.get<ApiResponse<ApiRole>>(`/roles/${id}`)
    return data.data
  },

  async create(payload: CreateRolePayload): Promise<ApiRole> {
    const { data } = await api.post<ApiResponse<ApiRole>>('/roles/', payload)
    return data.data
  },

  async update(id: string, payload: UpdateRolePayload): Promise<ApiRole> {
    const { data } = await api.put<ApiResponse<ApiRole>>(`/roles/${id}`, payload)
    return data.data
  },

  async delete(id: string): Promise<void> {
    await api.delete(`/roles/${id}`)
  },
}
