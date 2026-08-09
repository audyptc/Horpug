import api from '@/lib/axios'
import type {
  ApiPaginatedResponse,
  ApiResponse,
  ApiRole,
  ApiMenu,
  ApiPermission,
  CreateRolePayload,
  UpdateRolePayload,
  AssignPermissionsPayload,
} from '@/types'

export const roleService = {
  async list(page = 1, perPage = 20): Promise<ApiPaginatedResponse<ApiRole>> {
    const offset = (page - 1) * perPage
    const { data } = await api.get<ApiPaginatedResponse<ApiRole>>(
      `/roles/?limit=${perPage}&offset=${offset}`
    )
    return data
  },

  async listActive(): Promise<ApiRole[]> {
    const { data } = await api.get<ApiResponse<ApiRole[]>>('/roles/active')
    return data.data
  },

  async getById(id: string): Promise<ApiRole> {
    const { data } = await api.get<ApiResponse<ApiRole>>(`/roles/${id}`)
    return data.data
  },

  async create(payload: CreateRolePayload): Promise<ApiRole> {
    const { data } = await api.post<ApiResponse<ApiRole>>('/roles/', payload)
    return data.data
  },

  async update(id: string, payload: UpdateRolePayload): Promise<ApiRole> {
    const { data } = await api.put<ApiResponse<ApiRole>>(`/roles/${id}`, payload)
    return data.data
  },

  async delete(id: string): Promise<void> {
    await api.delete(`/roles/${id}`)
  },

  async listMenus(): Promise<ApiMenu[]> {
    const { data } = await api.get<ApiPaginatedResponse<ApiMenu>>('/menus/?limit=100&offset=0')
    return data.data
  },

  async listPermissions(): Promise<ApiPermission[]> {
    const { data } = await api.get<ApiPaginatedResponse<ApiPermission>>('/permissions/?limit=100&offset=0')
    return data.data
  },

  async assignPermissions(id: string, payload: AssignPermissionsPayload): Promise<void> {
    await api.put(`/roles/${id}/permissions`, payload)
  },
}
