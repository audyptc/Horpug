export type RepairCategory = 'electrical' | 'plumbing' | 'furniture' | 'aircon' | 'other'
export type RepairStatus = 'pending' | 'in_progress' | 'completed' | 'cancelled'

export type ApiRepairRequest = {
  id: string
  room_id: string
  room_number?: string
  dormitory_id?: string
  dormitory_name?: string
  tenant_id?: string
  tenant_name?: string
  category: RepairCategory
  description: string
  status: RepairStatus
  reported_date: string
  created_by?: string
  updated_by?: string
  created_at: string
  updated_at: string
}
