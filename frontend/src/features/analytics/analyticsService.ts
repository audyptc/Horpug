import api from '@/lib/axios'
import type { ApiResponse, ApiAnalyticsSummary } from '@/types'

export const analyticsService = {
  async summary(months = 12): Promise<ApiAnalyticsSummary> {
    const { data } = await api.get<ApiResponse<ApiAnalyticsSummary>>(
      `/analytics/summary?months=${months}`
    )
    return data.data
  },
}
