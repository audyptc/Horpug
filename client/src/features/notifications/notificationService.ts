import api from '@/lib/axios'
import type { ApiResponse } from '@/types/api'

export interface NotificationItem {
  id: string
  type: 'overdue_bill'
  tenant_name: string
  room_number: string
  total_amount: number
  days_overdue: number
  bill_id: string
  created_at: string
}

export const notificationService = {
  async list(): Promise<NotificationItem[]> {
    const { data } = await api.get<ApiResponse<NotificationItem[]>>('/notifications')
    return data.data
  },
}
