import api from '@/lib/axios'
import type {
  ApiPaginatedResponse,
  ApiResponse,
  ApiMaintenanceRequest,
  CreateMaintenanceRequestPayload,
  UpdateMaintenanceRequestPayload,
} from '@/types/api'

export const maintenanceService = {
  async list(page = 1, perPage = 20): Promise<ApiPaginatedResponse<ApiMaintenanceRequest>> {
    const offset = (page - 1) * perPage
    const { data } = await api.get<ApiPaginatedResponse<ApiMaintenanceRequest>>(
      `/maintenance/?limit=${perPage}&offset=${offset}`
    )
    return data
  },

  async getById(id: string): Promise<ApiMaintenanceRequest> {
    const { data } = await api.get<ApiResponse<ApiMaintenanceRequest>>(`/maintenance/${id}`)
    return data.data
  },

  async create(payload: CreateMaintenanceRequestPayload): Promise<ApiMaintenanceRequest> {
    const { data } = await api.post<ApiResponse<ApiMaintenanceRequest>>('/maintenance/', payload)
    return data.data
  },

  async update(id: string, payload: UpdateMaintenanceRequestPayload): Promise<ApiMaintenanceRequest> {
    const { data } = await api.put<ApiResponse<ApiMaintenanceRequest>>(`/maintenance/${id}`, payload)
    return data.data
  },

  async delete(id: string): Promise<void> {
    await api.delete(`/maintenance/${id}`)
  },
}
