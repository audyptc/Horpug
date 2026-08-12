import { menuMeta, type ApiMenu } from '@/features/menu/menus'
import type { TranslationKey } from '@/shared/i18n/language'
import type { ApiRole } from './types'

export const ACTION_ORDER = ['create', 'read', 'update', 'delete']

export const ROLE_PAGE_SIZE_OPTIONS = [10, 20, 50, 100] as const

export const actionLabelKeys: Record<string, TranslationKey> = {
  create: 'permissionActionCreate',
  read: 'permissionActionRead',
  update: 'permissionActionUpdate',
  delete: 'permissionActionDelete',
}

export function menuLabel(menu: ApiMenu, t: (key: TranslationKey) => string): string {
  const meta = menuMeta[menu.path]
  return meta ? t(meta.labelKey) : menu.name
}

export function buildRoleMatrix(role: ApiRole | null): Record<string, Set<string>> {
  const next: Record<string, Set<string>> = {}
  for (const item of role?.menu_permissions ?? []) {
    if (!next[item.menu_id]) next[item.menu_id] = new Set()
    next[item.menu_id].add(item.permission_id)
  }
  return next
}

export function areMatricesEqual(
  left: Record<string, Set<string>>,
  right: Record<string, Set<string>>
): boolean {
  const menuIds = new Set([...Object.keys(left), ...Object.keys(right)])
  for (const menuId of menuIds) {
    const leftIds = left[menuId] ?? new Set<string>()
    const rightIds = right[menuId] ?? new Set<string>()
    if (leftIds.size !== rightIds.size) return false
    for (const permissionId of leftIds) {
      if (!rightIds.has(permissionId)) return false
    }
  }
  return true
}
