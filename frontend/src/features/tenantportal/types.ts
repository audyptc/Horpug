export type PortalRoom = {
  contract_id: string
  room_id: string
  room_number: string
  dormitory_id: string
  dormitory_name: string
  dormitory_phone: string
  rent_price: number
  start_date: string
}

export type PortalProfile = {
  tenant_id: string
  first_name: string
  last_name: string
  rooms: PortalRoom[]
}

export type PortalSession = {
  access_token: string
  expires_at: string
  profile: PortalProfile
}

export type PortalInvoice = {
  id: string
  room_number: string
  dormitory_name: string
  period_year: number
  period_month: number
  due_date: string
  total_amount: number
  outstanding: number
  status: 'unpaid' | 'paid' | 'overdue' | 'cancelled'
}

export type RepairCategory = 'electrical' | 'plumbing' | 'furniture' | 'aircon' | 'other'
export type RepairStatus = 'pending' | 'in_progress' | 'completed' | 'cancelled'

export type PortalRepair = {
  id: string
  room_number: string
  category: RepairCategory
  description: string
  status: RepairStatus
  reported_date: string
  updated_at: string
}

export type PortalAnnouncement = {
  id: string
  dormitory_name: string
  title: string
  content: string
  category: string
  is_pinned: boolean
  published_date: string
}
