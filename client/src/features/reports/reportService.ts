import api from '@/lib/axios'
import type { ApiResponse, ApiIncomeReport, ApiExpenseReport, ApiOccupancyReport } from '@/types/api'

export const reportService = {
  async income(from: string, to: string): Promise<ApiIncomeReport> {
    const { data } = await api.get<ApiResponse<ApiIncomeReport>>(
      `/reports/income?from=${from}&to=${to}`
    )
    return data.data
  },

  async expenses(from: string, to: string): Promise<ApiExpenseReport> {
    const { data } = await api.get<ApiResponse<ApiExpenseReport>>(
      `/reports/expenses?from=${from}&to=${to}`
    )
    return data.data
  },

  async occupancy(): Promise<ApiOccupancyReport> {
    const { data } = await api.get<ApiResponse<ApiOccupancyReport>>('/reports/occupancy')
    return data.data
  },
}
