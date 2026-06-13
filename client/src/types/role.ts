export interface ApiPermission {
  id: string
  name: string
  description: string
  created_at: string
  updated_at: string
}

export interface ApiRoleMenuPermission {
  menu_id: string
  menu_name: string
  permissions: ApiPermission[]
}

export interface ApiRole {
  id: string
  name: string
  description: string
  is_active: boolean
  created_at: string
  updated_at: string
  menu_permissions?: ApiRoleMenuPermission[]
}

export interface CreateRolePayload {
  name: string
  description: string
  is_active: boolean
}

export interface UpdateRolePayload {
  name?: string
  description?: string
  is_active?: boolean
}

export interface AssignPermissionsItem {
  menu_id: string
  permission_ids: string[]
}

export interface AssignPermissionsPayload {
  items: AssignPermissionsItem[]
}
