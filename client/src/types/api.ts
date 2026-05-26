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

export interface ApiTenant {
  id: string
  first_name: string
  last_name: string
  phone: string
  id_card: string
  email: string
  emergency_contact: string
  note: string
  created_at: string
  updated_at: string
}

export interface CreateTenantPayload {
  first_name: string
  last_name: string
  phone: string
  id_card: string
  email: string
  emergency_contact: string
  note: string
}

export interface UpdateTenantPayload {
  first_name?: string
  last_name?: string
  phone?: string
  id_card?: string
  email?: string
  emergency_contact?: string
  note?: string
}

export type ContractStatus = 'active' | 'expired' | 'terminated'

export interface ApiContract {
  id: string
  tenant_id: string
  room_id: string
  start_date: string
  end_date: string | null
  rent_price: number
  deposit: number
  status: ContractStatus
  note: string
  created_at: string
  updated_at: string
  tenant_first_name: string
  tenant_last_name: string
  room_number: string
}

export interface CreateContractPayload {
  tenant_id: string
  room_id: string
  start_date: string
  end_date: string | null
  rent_price: number
  deposit: number
  note: string
}

export interface UpdateContractPayload {
  end_date?: string | null
  rent_price?: number
  deposit?: number
  status?: ContractStatus
  note?: string
}

export type RoomType = 'standard' | 'deluxe' | 'suite'
export type RoomStatus = 'available' | 'occupied' | 'maintenance'

export interface ApiRoom {
  id: string
  room_number: string
  floor: number
  type: RoomType
  status: RoomStatus
  rent_price: number
  description: string
  created_at: string
  updated_at: string
}

export interface CreateRoomPayload {
  room_number: string
  floor: number
  type: RoomType
  status: RoomStatus
  rent_price: number
  description: string
}

export interface UpdateRoomPayload {
  room_number?: string
  floor?: number
  type?: RoomType
  status?: RoomStatus
  rent_price?: number
  description?: string
}
