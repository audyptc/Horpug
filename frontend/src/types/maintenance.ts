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
