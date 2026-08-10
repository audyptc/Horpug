import api from '@/lib/axios'
import type { ApiDormitory, ApiResponse } from '@/types'

export const dormitoryService = {
  async mine(): Promise<ApiDormitory[]> {
    const { data } = await api.get<ApiResponse<ApiDormitory[]>>('/dormitories/mine')
    return data.data
  },
}
