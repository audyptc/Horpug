import api from '@/lib/axios'
import type {
  ApiPaginatedResponse,
  ApiResponse,
  ApiMeterReading,
  CreateMeterReadingPayload,
  UpdateMeterReadingPayload,
} from '@/types'

export const meterReadingService = {
  async list(page = 1, perPage = 20): Promise<ApiPaginatedResponse<ApiMeterReading>> {
    const offset = (page - 1) * perPage
    const { data } = await api.get<ApiPaginatedResponse<ApiMeterReading>>(
      `/meters/?limit=${perPage}&offset=${offset}`
    )
    return data
  },

  async getById(id: string): Promise<ApiMeterReading> {
    const { data } = await api.get<ApiResponse<ApiMeterReading>>(`/meters/${id}`)
    return data.data
  },

  async create(payload: CreateMeterReadingPayload): Promise<ApiMeterReading> {
    const { data } = await api.post<ApiResponse<ApiMeterReading>>('/meters/', payload)
    return data.data
  },

  async update(id: string, payload: UpdateMeterReadingPayload): Promise<ApiMeterReading> {
    const { data } = await api.put<ApiResponse<ApiMeterReading>>(`/meters/${id}`, payload)
    return data.data
  },

  async delete(id: string): Promise<void> {
    await api.delete(`/meters/${id}`)
  },
}
