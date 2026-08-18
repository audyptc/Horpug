export type ApiUserRole = {
  id: string
  name: string
}

export type ApiUser = {
  id: string
  username: string
  email: string
  role_id: string
  role?: ApiUserRole
  is_active: boolean
  is_protected: boolean
  created_at: string
  updated_at: string
}

export type ApiUserDeletionCheck = {
  can_delete: boolean
  record_count: number
  is_protected: boolean
}
