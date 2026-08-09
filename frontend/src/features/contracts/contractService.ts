import api from '@/lib/axios'
import type {
  ApiPaginatedResponse,
  ApiResponse,
  ApiContract,
  CreateContractPayload,
  UpdateContractPayload,
} from '@/types'

export const contractService = {
  async list(page = 1, perPage = 20): Promise<ApiPaginatedResponse<ApiContract>> {
    const offset = (page - 1) * perPage
    const { data } = await api.get<ApiPaginatedResponse<ApiContract>>(
      `/contracts/?limit=${perPage}&offset=${offset}`
    )
    return data
  },

  async getById(id: string): Promise<ApiContract> {
    const { data } = await api.get<ApiResponse<ApiContract>>(`/contracts/${id}`)
    return data.data
  },

  async create(payload: CreateContractPayload): Promise<ApiContract> {
    const { data } = await api.post<ApiResponse<ApiContract>>('/contracts/', payload)
    return data.data
  },

  async update(id: string, payload: UpdateContractPayload): Promise<ApiContract> {
    const { data } = await api.put<ApiResponse<ApiContract>>(`/contracts/${id}`, payload)
    return data.data
  },

  async delete(id: string): Promise<void> {
    await api.delete(`/contracts/${id}`)
  },
}
