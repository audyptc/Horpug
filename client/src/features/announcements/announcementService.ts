import api from '@/lib/axios'
import type { ApiAnnouncement, ApiPaginatedResponse, CreateAnnouncementPayload, UpdateAnnouncementPayload } from '@/types/api'

export const announcementService = {
  list(page: number, perPage: number) {
    return api.get<ApiPaginatedResponse<ApiAnnouncement>>('/announcements', {
      params: { page, per_page: perPage },
    }).then((r) => r.data)
  },

  getById(id: string) {
    return api.get<{ data: ApiAnnouncement }>(`/announcements/${id}`).then((r) => r.data.data)
  },

  create(payload: CreateAnnouncementPayload) {
    return api.post<{ data: ApiAnnouncement }>('/announcements', payload).then((r) => r.data.data)
  },

  update(id: string, payload: UpdateAnnouncementPayload) {
    return api.put<{ data: ApiAnnouncement }>(`/announcements/${id}`, payload).then((r) => r.data.data)
  },

  delete(id: string) {
    return api.delete(`/announcements/${id}`)
  },
}
