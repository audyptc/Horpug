import { useState, useEffect } from 'react'
import { notificationService, type NotificationItem } from './notificationService'

export function useNotifications() {
  const [items, setItems] = useState<NotificationItem[]>([])

  useEffect(() => {
    notificationService.list().then(setItems).catch(() => {})
  }, [])

  return items
}
