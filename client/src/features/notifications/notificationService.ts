import api from '@/lib/axios'
import type { ApiResponse } from '@/types/api'

export interface NotificationItem {
  id: string
  type: 'overdue_bill' | 'expiring_contract' | 'open_maintenance'
  tenant_name?: string
  room_number: string
  total_amount?: number
  days_overdue?: number
  days_remaining?: number
  title?: string
  priority?: string
  created_at: string
}

export const notificationService = {
  async list(): Promise<NotificationItem[]> {
    const { data } = await api.get<ApiResponse<NotificationItem[]>>('/notifications')
    return data.data
  },
}
