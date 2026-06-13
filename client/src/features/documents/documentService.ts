import api from '@/lib/axios'
import type {
  ApiPaginatedResponse,
  ApiDocument,
  CreateDocumentPayload,
  UpdateDocumentPayload,
} from '@/types'

export const documentService = {
  list(page: number, perPage: number) {
    return api
      .get<ApiPaginatedResponse<ApiDocument>>('/documents', { params: { page, per_page: perPage } })
      .then((r) => r.data)
  },

  getById(id: string) {
    return api.get<{ data: ApiDocument }>(`/documents/${id}`).then((r) => r.data.data)
  },

  create(payload: CreateDocumentPayload) {
    return api.post<{ data: ApiDocument }>('/documents', payload).then((r) => r.data.data)
  },

  update(id: string, payload: UpdateDocumentPayload) {
    return api.put<{ data: ApiDocument }>(`/documents/${id}`, payload).then((r) => r.data.data)
  },

  delete(id: string) {
    return api.delete(`/documents/${id}`)
  },

  upload(file: File) {
    const form = new FormData()
    form.append('file', file)
    return api
      .post<{ data: { url: string } }>('/documents/upload', form, {
        headers: { 'Content-Type': 'multipart/form-data' },
      })
      .then((r) => r.data.data)
  },
}
