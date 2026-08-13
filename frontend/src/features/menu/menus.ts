import { useEffect, useState } from 'react'
import {
  BedDouble,
  Building2,
  Contact,
  CreditCard,
  DoorOpen,
  Droplets,
  FileText,
  Gauge,
  History,
  Receipt,
  ShieldCheck,
  Users2,
  type LucideIcon,
} from 'lucide-react'
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
  group?: 'rooms' | 'finance' | 'reports' | 'access' | 'settings'
}

// The backend seeds a fixed, known set of menu paths (see
// backend/internal/features/menu/repository/postgres/seed.go). It has no
// concept of locale, so the icon and bilingual label for each one live here.
export const menuMeta: Record<string, MenuMeta> = {
  '/dormitories': {
    icon: Building2,
    labelKey: 'menuDormitories',
    descriptionKey: 'menuDormitoriesDescription',
    group: 'settings',
  },
  '/room-types': {
    icon: BedDouble,
    labelKey: 'menuRoomTypes',
    descriptionKey: 'menuRoomTypesDescription',
    group: 'settings',
  },
  '/rooms': {
    icon: DoorOpen,
    labelKey: 'menuRooms',
    descriptionKey: 'menuRoomsDescription',
    group: 'settings',
  },
  '/tenants': {
    icon: Contact,
    labelKey: 'menuTenants',
    descriptionKey: 'menuTenantsDescription',
    group: 'rooms',
  },
  '/contracts': {
    icon: FileText,
    labelKey: 'menuContracts',
    descriptionKey: 'menuContractsDescription',
    group: 'rooms',
  },
  '/meters': {
    icon: Gauge,
    labelKey: 'menuMeters',
    descriptionKey: 'menuMetersDescription',
    group: 'finance',
  },
  '/water-meters': {
    icon: Droplets,
    labelKey: 'menuWaterMeters',
    descriptionKey: 'menuWaterMetersDescription',
    group: 'finance',
  },
  '/invoices': {
    icon: Receipt,
    labelKey: 'menuInvoices',
    descriptionKey: 'menuInvoicesDescription',
    group: 'finance',
  },
  '/payments': {
    icon: CreditCard,
    labelKey: 'menuPayments',
    descriptionKey: 'menuPaymentsDescription',
    group: 'finance',
  },
  '/roles': {
    icon: ShieldCheck,
    labelKey: 'menuRoles',
    descriptionKey: 'menuRolesDescription',
    group: 'access',
  },
  '/users': {
    icon: Users2,
    labelKey: 'menuUsers',
    descriptionKey: 'menuUsersDescription',
    group: 'access',
  },
  '/activity-logs': {
    icon: History,
    labelKey: 'menuActivityLogs',
    descriptionKey: 'menuActivityLogsDescription',
    group: 'reports',
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
