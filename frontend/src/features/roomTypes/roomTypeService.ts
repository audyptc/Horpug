import api from '@/lib/axios'
import type { ApiResponse, ApiRoomType } from '@/types'

export const roomTypeService = {
  async list(): Promise<ApiRoomType[]> {
    const { data } = await api.get<ApiResponse<ApiRoomType[]>>('/room-types/')
    return data.data
  },

  async create(payload: { id: string; name: string; sort_order: number }): Promise<ApiRoomType> {
    const { data } = await api.post<ApiResponse<ApiRoomType>>('/room-types/', payload)
    return data.data
  },

  async update(id: string, payload: { name: string; sort_order: number }): Promise<ApiRoomType> {
    const { data } = await api.put<ApiResponse<ApiRoomType>>(`/room-types/${id}`, payload)
    return data.data
  },

  async delete(id: string): Promise<void> {
    await api.delete(`/room-types/${id}`)
  },
}
