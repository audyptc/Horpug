import { useEffect, useState } from 'react'
import { Building2, History, KeyRound, ShieldCheck, Users2, type LucideIcon } from 'lucide-react'
import { api, type ApiPage } from '@/shared/api/client'
import { useAuth } from '@/features/auth/AuthProvider'
import type { TranslationKey } from '@/shared/i18n/language'

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
  const { isAuthenticated, session } = useAuth()
  const userId = isAuthenticated ? session?.user.id ?? null : null
  const [menuState, setMenuState] = useState<{
    userId: string | null
    menus: ApiMenu[]
    error: string | null
  }>({
    userId: null,
    menus: [],
    error: null,
  })

  useEffect(() => {
    if (!userId) return

    let cancelled = false

    api
      .get<ApiPage<ApiMenu[]>>('/menus', { params: { per_page: 100 } })
      .then(({ data }) => {
        if (!cancelled) {
          setMenuState({
            userId,
            menus: data.data.filter((menu) => menu.is_active && menu.path in menuMeta),
            error: null,
          })
        }
      })
      .catch(() => {
        if (!cancelled) {
          setMenuState({
            userId,
            menus: [],
            error: 'failed to load menus',
          })
        }
      })

    return () => {
      cancelled = true
    }
  }, [userId])

  const menus = menuState.userId === userId ? menuState.menus : []
  const error = menuState.userId === userId ? menuState.error : null
  const loading = Boolean(userId) && menuState.userId !== userId

  return { menus, loading, error }
}
