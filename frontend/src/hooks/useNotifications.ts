import { useState, useEffect } from 'react'
import { usePermission } from '@/hooks/usePermission'
import { notificationService, type NotificationItem } from '@/features/notifications/notificationService'

export function useNotifications() {
  const { canRead } = usePermission('/notifications')
  const [items, setItems] = useState<NotificationItem[]>([])

  useEffect(() => {
    if (!canRead) {
      setItems([])
      return
    }

    notificationService.list().then(setItems).catch(() => {})
  }, [canRead])

  return items
}
