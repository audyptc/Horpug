import { useEffect, useState } from 'react'
import { Building2, History, KeyRound, ShieldCheck, Users2, type LucideIcon } from 'lucide-react'
import { api } from './api'
import { useAuth } from './auth'
import type { TranslationKey } from './language'

export type ApiMenu = {
  id: string
  name: string
  path: string
  description: string
  is_active: boolean
}

export type MenuMeta = {
  icon: LucideIcon
  labelKey: TranslationKey
  descriptionKey: TranslationKey
}

// The backend seeds a fixed, known set of menu paths (see
// backend/internal/features/menu/repository/postgres/seed.go). It has no
// concept of locale, so the icon and bilingual label for each one live here.
export const menuMeta: Record<string, MenuMeta> = {
  '/users': { icon: Users2, labelKey: 'menuUsers', descriptionKey: 'menuUsersDescription' },
  '/roles': { icon: ShieldCheck, labelKey: 'menuRoles', descriptionKey: 'menuRolesDescription' },
  '/permissions': {
    icon: KeyRound,
    labelKey: 'menuPermissions',
    descriptionKey: 'menuPermissionsDescription',
  },
  '/dormitories': {
    icon: Building2,
    labelKey: 'menuDormitories',
    descriptionKey: 'menuDormitoriesDescription',
  },
  '/activity-logs': {
    icon: History,
    labelKey: 'menuActivityLogs',
    descriptionKey: 'menuActivityLogsDescription',
  },
}

export function useMenus() {
  const { isAuthenticated } = useAuth()
  const [menus, setMenus] = useState<ApiMenu[]>([])
  const [loading, setLoading] = useState(true)
  const [error, setError] = useState<string | null>(null)

  useEffect(() => {
    if (!isAuthenticated) {
      setMenus([])
      setLoading(false)
      return
    }

    let cancelled = false
    setLoading(true)
    setError(null)

    api
      .get<ApiMenu[]>('/menus')
      .then(({ data }) => {
        if (!cancelled) setMenus(data.filter((menu) => menu.is_active && menu.path in menuMeta))
      })
      .catch(() => {
        if (!cancelled) setError('failed to load menus')
      })
      .finally(() => {
        if (!cancelled) setLoading(false)
      })

    return () => {
      cancelled = true
    }
  }, [isAuthenticated])

  return { menus, loading, error }
}
