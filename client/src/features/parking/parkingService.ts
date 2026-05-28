import api from '@/lib/axios'
import type {
  ApiPaginatedResponse,
  ApiParkingSlot,
  CreateParkingSlotPayload,
  UpdateParkingSlotPayload,
} from '@/types'

export const parkingService = {
  list(page: number, perPage: number) {
    return api
      .get<ApiPaginatedResponse<ApiParkingSlot>>('/parking', { params: { page, per_page: perPage } })
      .then((r) => r.data)
  },

  getById(id: string) {
    return api.get<{ data: ApiParkingSlot }>(`/parking/${id}`).then((r) => r.data.data)
  },

  create(payload: CreateParkingSlotPayload) {
    return api.post<{ data: ApiParkingSlot }>('/parking', payload).then((r) => r.data.data)
  },

  update(id: string, payload: UpdateParkingSlotPayload) {
    return api.put<{ data: ApiParkingSlot }>(`/parking/${id}`, payload).then((r) => r.data.data)
  },

  delete(id: string) {
    return api.delete(`/parking/${id}`)
  },
}
