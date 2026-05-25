import api from '@/lib/axios'
import type {
  ApiResponse,
  ApiPaginatedResponse,
  ApiUser,
  CreateUserPayload,
  UpdateUserPayload,
  AssignRolePayload,
} from '@/types/api'

export const userService = {
  async list(page = 1, perPage = 20): Promise<ApiPaginatedResponse<ApiUser>> {
    const offset = (page - 1) * perPage
    const { data } = await api.get<ApiPaginatedResponse<ApiUser>>(
      `/users/?limit=${perPage}&offset=${offset}`
    )
    return data
  },

  async getById(id: string): Promise<ApiUser> {
    const { data } = await api.get<ApiResponse<ApiUser>>(`/users/${id}`)
    return data.data
  },

  async create(payload: CreateUserPayload): Promise<ApiUser> {
    const { data } = await api.post<ApiResponse<ApiUser>>('/users/', payload)
    return data.data
  },

  async update(id: string, payload: UpdateUserPayload): Promise<ApiUser> {
    const { data } = await api.put<ApiResponse<ApiUser>>(`/users/${id}`, payload)
    return data.data
  },

  async delete(id: string): Promise<void> {
    await api.delete(`/users/${id}`)
  },

  async assignRole(id: string, payload: AssignRolePayload): Promise<void> {
    await api.put(`/users/${id}/role`, payload)
  },
}
