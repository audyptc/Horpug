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

export type MeterType = 'electric' | 'water'

export interface ApiMeterReading {
  id: string
  room_id: string
  meter_type: MeterType
  reading_date: string
  previous_reading: number
  current_reading: number
  unit_price: number
  unit_used: number
  total_amount: number
  note: string
  created_at: string
  updated_at: string
  room_number: string
}

export interface CreateMeterReadingPayload {
  room_id: string
  meter_type: MeterType
  reading_date: string
  previous_reading: number
  current_reading: number
  unit_price: number
  note: string
}

export interface UpdateMeterReadingPayload {
  reading_date?: string
  previous_reading?: number
  current_reading?: number
  unit_price?: number
  note?: string
}

export type BillStatus = 'unpaid' | 'paid' | 'overdue'

export interface ApiBill {
  id: string
  contract_id: string
  billing_month: string
  rent_amount: number
  electric_amount: number
  water_amount: number
  other_amount: number
  other_note: string
  total_amount: number
  status: BillStatus
  due_date: string | null
  paid_at: string | null
  note: string
  created_at: string
  updated_at: string
  tenant_first_name: string
  tenant_last_name: string
  room_number: string
}

export interface MonthlyRevenue {
  month: string
  revenue: number
}

export interface MonthlyCount {
  month: string
  count: number
}

export interface BillStatusCount {
  unpaid: number
  paid: number
  overdue: number
}

export interface ApiAnalyticsSummary {
  monthly_revenue: MonthlyRevenue[]
  monthly_tenants: MonthlyCount[]
  bill_status: BillStatusCount
  total_revenue: number
  avg_monthly: number
}

export interface ApiDashboardSummary {
  total_rooms: number
  occupied_rooms: number
  available_rooms: number
  maintenance_rooms: number
  occupancy_rate: number
  total_tenants: number
  active_contracts: number
  expiring_contracts: number
  unpaid_bills: number
  overdue_bills: number
  unpaid_amount: number
  overdue_amount: number
  revenue_this_month: number
}

export interface CreateBillPayload {
  contract_id: string
  billing_month: string
  rent_amount: number
  electric_amount: number
  water_amount: number
  other_amount: number
  other_note: string
  due_date: string | null
  note: string
}

export interface UpdateBillPayload {
  rent_amount: number
  electric_amount: number
  water_amount: number
  other_amount: number
  other_note: string
  status: BillStatus
  due_date: string | null
  note: string
}

export type ExpenseCategory = 'repair' | 'utilities' | 'supplies' | 'salary' | 'other'

export interface ApiExpense {
  id: string
  expense_date: string
  category: ExpenseCategory
  description: string
  amount: number
  note: string
  created_at: string
  updated_at: string
}

export interface CreateExpensePayload {
  expense_date: string
  category: ExpenseCategory
  description: string
  amount: number
  note: string
}

export interface UpdateExpensePayload {
  expense_date: string
  category: ExpenseCategory
  description: string
  amount: number
  note: string
}

export type MaintenanceStatus = 'open' | 'in_progress' | 'done' | 'cancelled'
export type MaintenancePriority = 'low' | 'normal' | 'high' | 'urgent'

export interface ApiMaintenanceRequest {
  id: string
  room_id: string
  title: string
  description: string
  status: MaintenanceStatus
  priority: MaintenancePriority
  reported_date: string
  resolved_date: string | null
  note: string
  created_at: string
  updated_at: string
  room_number: string
}

export interface CreateMaintenanceRequestPayload {
  room_id: string
  title: string
  description: string
  status: MaintenanceStatus
  priority: MaintenancePriority
  reported_date: string
  resolved_date: string | null
  note: string
}

export interface UpdateMaintenanceRequestPayload {
  room_id: string
  title: string
  description: string
  status: MaintenanceStatus
  priority: MaintenancePriority
  reported_date: string
  resolved_date: string | null
  note: string
}

export type PaymentMethod = 'cash' | 'transfer' | 'qr'

export interface ApiPayment {
  id: string
  bill_id: string
  amount: number
  method: PaymentMethod
  payment_date: string
  note: string
  created_at: string
  updated_at: string
  room_number: string
  tenant_first_name: string
  tenant_last_name: string
  billing_month: string
  bill_total: number
}

export interface CreatePaymentPayload {
  bill_id: string
  amount: number
  method: PaymentMethod
  payment_date: string
  note: string
}

export interface UpdatePaymentPayload {
  amount: number
  method: PaymentMethod
  payment_date: string
  note: string
}

export type AnnouncementType = 'general' | 'maintenance' | 'payment' | 'emergency'

export interface ApiAnnouncement {
  id: string
  title: string
  content: string
  type: AnnouncementType
  is_pinned: boolean
  published_at: string
  expired_at: string | null
  created_at: string
  updated_at: string
}

export interface CreateAnnouncementPayload {
  title: string
  content: string
  type: AnnouncementType
  is_pinned: boolean
  published_at: string
  expired_at: string | null
}

export interface UpdateAnnouncementPayload {
  title: string
  content: string
  type: AnnouncementType
  is_pinned: boolean
  published_at: string
  expired_at: string | null
}
