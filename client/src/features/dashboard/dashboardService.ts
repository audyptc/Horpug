import api from '@/lib/axios'
import type { ApiResponse, ApiDashboardSummary } from '@/types'

export const dashboardService = {
  async summary(): Promise<ApiDashboardSummary> {
    const { data } = await api.get<ApiResponse<ApiDashboardSummary>>('/dashboard/summary')
    return data.data
  },
}
