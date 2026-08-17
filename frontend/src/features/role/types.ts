export type ApiPermission = {
  id: string
  name: string
  description: string
}

export type ApiRoleMenuPermission = {
  menu_id: string
  permission_id: string
}

export type ApiRole = {
  id: string
  name: string
  description: string
  is_active: boolean
  is_protected: boolean
  menu_permissions?: ApiRoleMenuPermission[]
}

export type ApiRoleDeletionCheck = {
  can_delete: boolean
  user_count: number
  is_protected: boolean
}
