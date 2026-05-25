export interface ApiResponse<T> {
  success: boolean
  data: T
}

export interface ApiPaginatedResponse<T> {
  success: boolean
  data: T[]
  meta: PaginationMeta
}

export interface PaginationMeta {
  page: number
  per_page: number
  total: number
  total_pages: number
}

export interface ApiRole {
  id: string
  name: string
  description: string
  is_active: boolean
  created_at: string
  updated_at: string
}

export interface ApiUser {
  id: string
  full_name: string
  email: string
  is_active: boolean
  created_at: string
  updated_at: string
  role?: ApiRole
}

export interface LoginResponse {
  access_token: string
  refresh_token: string
  expires_in: number
}

export interface CreateUserPayload {
  full_name: string
  email: string
  password: string
}

export interface UpdateUserPayload {
  full_name?: string
  password?: string
  is_active?: boolean
}

export interface AssignRolePayload {
  role_id: string
}
