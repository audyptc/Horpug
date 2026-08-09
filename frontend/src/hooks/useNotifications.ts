import { useState, useEffect } from 'react'
import { notificationService, type NotificationItem } from '@/features/notifications/notificationService'

export function useNotifications() {
  const [items, setItems] = useState<NotificationItem[]>([])

  useEffect(() => {
    notificationService.list().then(setItems).catch(() => {})
  }, [])

  return items
}
