import api from '@/lib/axios'
import type {
  ApiPaginatedResponse,
  ApiResponse,
  ApiRoom,
  CreateRoomPayload,
  UpdateRoomPayload,
} from '@/types/api'

export const roomService = {
  async list(page = 1, perPage = 20): Promise<ApiPaginatedResponse<ApiRoom>> {
    const offset = (page - 1) * perPage
    const { data } = await api.get<ApiPaginatedResponse<ApiRoom>>(
      `/rooms/?limit=${perPage}&offset=${offset}`
    )
    return data
  },

  async getById(id: string): Promise<ApiRoom> {
    const { data } = await api.get<ApiResponse<ApiRoom>>(`/rooms/${id}`)
    return data.data
  },

  async create(payload: CreateRoomPayload): Promise<ApiRoom> {
    const { data } = await api.post<ApiResponse<ApiRoom>>('/rooms/', payload)
    return data.data
  },

  async update(id: string, payload: UpdateRoomPayload): Promise<ApiRoom> {
    const { data } = await api.put<ApiResponse<ApiRoom>>(`/rooms/${id}`, payload)
    return data.data
  },

  async delete(id: string): Promise<void> {
    await api.delete(`/rooms/${id}`)
  },
}
