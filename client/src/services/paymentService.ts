import api from '@/lib/axios'
import type { ApiPaginatedResponse, ApiPayment, CreatePaymentPayload, UpdatePaymentPayload } from '@/types/api'

export const paymentService = {
  list(page: number, perPage: number) {
    return api.get<ApiPaginatedResponse<ApiPayment>>('/payments', {
      params: { page, per_page: perPage },
    }).then((r) => r.data)
  },

  getById(id: string) {
    return api.get<{ data: ApiPayment }>(`/payments/${id}`).then((r) => r.data.data)
  },

  create(payload: CreatePaymentPayload) {
    return api.post<{ data: ApiPayment }>('/payments', payload).then((r) => r.data.data)
  },

  update(id: string, payload: UpdatePaymentPayload) {
    return api.put<{ data: ApiPayment }>(`/payments/${id}`, payload).then((r) => r.data.data)
  },

  delete(id: string) {
    return api.delete(`/payments/${id}`)
  },
}
