import api from '@/lib/axios'
import type {
  ApiPaginatedResponse,
  ApiResponse,
  ApiBill,
  CreateBillPayload,
  UpdateBillPayload,
} from '@/types/api'

export const billService = {
  async list(page = 1, perPage = 20): Promise<ApiPaginatedResponse<ApiBill>> {
    const offset = (page - 1) * perPage
    const { data } = await api.get<ApiPaginatedResponse<ApiBill>>(
      `/bills/?limit=${perPage}&offset=${offset}`
    )
    return data
  },

  async getById(id: string): Promise<ApiBill> {
    const { data } = await api.get<ApiResponse<ApiBill>>(`/bills/${id}`)
    return data.data
  },

  async create(payload: CreateBillPayload): Promise<ApiBill> {
    const { data } = await api.post<ApiResponse<ApiBill>>('/bills/', payload)
    return data.data
  },

  async update(id: string, payload: UpdateBillPayload): Promise<ApiBill> {
    const { data } = await api.put<ApiResponse<ApiBill>>(`/bills/${id}`, payload)
    return data.data
  },

  async delete(id: string): Promise<void> {
    await api.delete(`/bills/${id}`)
  },
}
