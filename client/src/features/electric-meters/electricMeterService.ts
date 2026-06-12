import api from '@/lib/axios'
import type {
  ApiPaginatedResponse,
  ApiResponse,
  ApiElectricMeter,
  CreateElectricMeterPayload,
  UpdateElectricMeterPayload,
} from '@/types'

export const electricMeterService = {
  async list(page = 1, perPage = 20): Promise<ApiPaginatedResponse<ApiElectricMeter>> {
    const offset = (page - 1) * perPage
    const { data } = await api.get<ApiPaginatedResponse<ApiElectricMeter>>(
      `/electric-meters/?limit=${perPage}&offset=${offset}`
    )
    return data
  },

  async getById(id: string): Promise<ApiElectricMeter> {
    const { data } = await api.get<ApiResponse<ApiElectricMeter>>(`/electric-meters/${id}`)
    return data.data
  },

  async create(payload: CreateElectricMeterPayload): Promise<ApiElectricMeter> {
    const { data } = await api.post<ApiResponse<ApiElectricMeter>>('/electric-meters/', payload)
    return data.data
  },

  async update(id: string, payload: UpdateElectricMeterPayload): Promise<ApiElectricMeter> {
    const { data } = await api.put<ApiResponse<ApiElectricMeter>>(`/electric-meters/${id}`, payload)
    return data.data
  },

  async delete(id: string): Promise<void> {
    await api.delete(`/electric-meters/${id}`)
  },

  async getLatestByRoomId(roomId: string): Promise<ApiElectricMeter | null> {
    try {
      const { data } = await api.get<ApiResponse<ApiElectricMeter>>(`/electric-meters/latest?room_id=${roomId}`)
      return data.data
    } catch {
      return null
    }
  },
}
