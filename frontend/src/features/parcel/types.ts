export type ParcelStatus = 'pending' | 'picked_up' | 'returned'

export type ApiParcel = {
  id: string
  tenant_id: string
  tenant_name?: string
  room_id?: string
  room_number?: string
  dormitory_id?: string
  dormitory_name?: string
  courier: string
  tracking_number: string
  status: ParcelStatus
  received_date: string
  note: string
  created_by?: string
  updated_by?: string
  created_at: string
  updated_at: string
}
