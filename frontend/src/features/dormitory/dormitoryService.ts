import api from '@/lib/axios'
import type {
  ApiDormitory,
  ApiResponse,
  ApiPaginatedResponse,
  CreateDormitoryPayload,
  UpdateDormitoryPayload,
  AssignDormitoriesPayload,
} from '@/types'

export const dormitoryService = {
  async mine(): Promise<ApiDormitory[]> {
    const { data } = await api.get<ApiResponse<ApiDormitory[]>>('/dormitories/mine')
    return data.data
  },

  async list(): Promise<ApiDormitory[]> {
    const { data } = await api.get<ApiPaginatedResponse<ApiDormitory>>('/dormitories/?limit=100&offset=0')
    return data.data
  },

  async getById(id: string): Promise<ApiDormitory> {
    const { data } = await api.get<ApiResponse<ApiDormitory>>(`/dormitories/${id}`)
    return data.data
  },

  async create(payload: CreateDormitoryPayload): Promise<ApiDormitory> {
    const { data } = await api.post<ApiResponse<ApiDormitory>>('/dormitories/', payload)
    return data.data
  },

  async update(id: string, payload: UpdateDormitoryPayload): Promise<ApiDormitory> {
    const { data } = await api.put<ApiResponse<ApiDormitory>>(`/dormitories/${id}`, payload)
    return data.data
  },

  async delete(id: string): Promise<void> {
    await api.delete(`/dormitories/${id}`)
  },

  async getForUser(userId: string): Promise<ApiDormitory[]> {
    const { data } = await api.get<ApiResponse<ApiDormitory[]>>(`/dormitories/users/${userId}`)
    return data.data
  },

  async assignToUser(userId: string, payload: AssignDormitoriesPayload): Promise<void> {
    await api.put(`/dormitories/users/${userId}`, payload)
  },
}
