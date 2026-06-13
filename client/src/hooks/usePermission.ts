import { useAuth } from '@/features/auth/AuthContext'

export interface PermissionFlags {
  canRead: boolean
  canCreate: boolean
  canUpdate: boolean
  canDelete: boolean
  canExport: boolean
}

// /rooms → "rooms", /settings/users → "settings/users", / → ""
function pathToPermKey(path: string): string {
  return path.split(/[?#]/)[0].trim().replace(/^\/+|\/+$/g, '')
}

export function usePermission(path: string): PermissionFlags {
  const { isAdmin, rawPermissions } = useAuth()

  if (isAdmin) {
    return { canRead: true, canCreate: true, canUpdate: true, canDelete: true, canExport: true }
  }

  const key = pathToPermKey(path)
  const pre = key ? `${key}.` : '.'

  return {
    canRead:   rawPermissions.has(`${pre}read`),
    canCreate: rawPermissions.has(`${pre}create`),
    canUpdate: rawPermissions.has(`${pre}update`),
    canDelete: rawPermissions.has(`${pre}delete`),
    canExport: rawPermissions.has(`${pre}export`),
  }
}
