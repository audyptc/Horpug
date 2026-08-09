import api from '@/lib/axios'
import type {
  ApiPaginatedResponse,
  ApiResponse,
  ApiTenant,
  CreateTenantPayload,
  UpdateTenantPayload,
} from '@/types'

export const tenantService = {
  async list(page = 1, perPage = 20): Promise<ApiPaginatedResponse<ApiTenant>> {
    const offset = (page - 1) * perPage
    const { data } = await api.get<ApiPaginatedResponse<ApiTenant>>(
      `/tenants/?limit=${perPage}&offset=${offset}`
    )
    return data
  },

  async getById(id: string): Promise<ApiTenant> {
    const { data } = await api.get<ApiResponse<ApiTenant>>(`/tenants/${id}`)
    return data.data
  },

  async create(payload: CreateTenantPayload): Promise<ApiTenant> {
    const { data } = await api.post<ApiResponse<ApiTenant>>('/tenants/', payload)
    return data.data
  },

  async update(id: string, payload: UpdateTenantPayload): Promise<ApiTenant> {
    const { data } = await api.put<ApiResponse<ApiTenant>>(`/tenants/${id}`, payload)
    return data.data
  },

  async delete(id: string): Promise<void> {
    await api.delete(`/tenants/${id}`)
  },
}
