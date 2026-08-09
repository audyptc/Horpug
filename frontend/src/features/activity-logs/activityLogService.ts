import api from '@/lib/axios'
import type { ApiActivityLog, ApiPaginatedResponse } from '@/types'

interface ActivityLogParams {
  page: number
  perPage: number
  entityType?: string
  from?: string
  to?: string
}

export const activityLogService = {
  list({ page, perPage, entityType, from, to }: ActivityLogParams) {
    return api
      .get<ApiPaginatedResponse<ApiActivityLog>>('/activity-logs', {
        params: {
          page,
          per_page: perPage,
          entity_type: entityType || undefined,
          from: from || undefined,
          to: to || undefined,
        },
      })
      .then((r) => r.data)
  },
}
