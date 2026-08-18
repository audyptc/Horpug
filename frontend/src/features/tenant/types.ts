export type ApiTenant = {
  id: string
  first_name: string
  last_name: string
  phone: string
  line_id: string
  id_card: string
  email: string
  emergency_contact: string
  note: string
  is_active: boolean
  created_at: string
  updated_at: string
}

export type ApiTenantDeletionCheck = {
  can_delete: boolean
  contract_count: number
}
