import api from '@/lib/axios'
import type {
  ApiPaginatedResponse,
  ApiResponse,
  ApiWaterMeter,
  CreateWaterMeterPayload,
  UpdateWaterMeterPayload,
} from '@/types'

export const waterMeterService = {
  async list(page = 1, perPage = 20): Promise<ApiPaginatedResponse<ApiWaterMeter>> {
    const offset = (page - 1) * perPage
    const { data } = await api.get<ApiPaginatedResponse<ApiWaterMeter>>(
      `/water-meters/?limit=${perPage}&offset=${offset}`
    )
    return data
  },

  async getById(id: string): Promise<ApiWaterMeter> {
    const { data } = await api.get<ApiResponse<ApiWaterMeter>>(`/water-meters/${id}`)
    return data.data
  },

  async create(payload: CreateWaterMeterPayload): Promise<ApiWaterMeter> {
    const { data } = await api.post<ApiResponse<ApiWaterMeter>>('/water-meters/', payload)
    return data.data
  },

  async update(id: string, payload: UpdateWaterMeterPayload): Promise<ApiWaterMeter> {
    const { data } = await api.put<ApiResponse<ApiWaterMeter>>(`/water-meters/${id}`, payload)
    return data.data
  },

  async delete(id: string): Promise<void> {
    await api.delete(`/water-meters/${id}`)
  },

  async getLatestByRoomId(roomId: string, billingMonth?: string): Promise<ApiWaterMeter | null> {
    try {
      const params = new URLSearchParams({ room_id: roomId })
      if (billingMonth) params.set('billing_month', `${billingMonth}-01`)
      const { data } = await api.get<ApiResponse<ApiWaterMeter>>(`/water-meters/latest?${params}`)
      return data.data
    } catch {
      return null
    }
  },
}
