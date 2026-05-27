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
