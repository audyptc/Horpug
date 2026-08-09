import api from '@/lib/axios'
import type {
  ApiPaginatedResponse,
  ApiParcel,
  CreateParcelPayload,
  UpdateParcelPayload,
} from '@/types'

export const parcelService = {
  list(page: number, perPage: number) {
    return api
      .get<ApiPaginatedResponse<ApiParcel>>('/parcels', { params: { page, per_page: perPage } })
      .then((r) => r.data)
  },

  getById(id: string) {
    return api.get<{ data: ApiParcel }>(`/parcels/${id}`).then((r) => r.data.data)
  },

  create(payload: CreateParcelPayload) {
    return api.post<{ data: ApiParcel }>('/parcels', payload).then((r) => r.data.data)
  },

  update(id: string, payload: UpdateParcelPayload) {
    return api.put<{ data: ApiParcel }>(`/parcels/${id}`, payload).then((r) => r.data.data)
  },

  delete(id: string) {
    return api.delete(`/parcels/${id}`)
  },
}
