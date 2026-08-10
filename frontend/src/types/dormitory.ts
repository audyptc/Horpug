export interface ApiDormitory {
  id: string
  name: string
  address: string
  is_active: boolean
  created_at: string
  updated_at: string
}

export interface CreateDormitoryPayload {
  name: string
  address: string
}

export interface UpdateDormitoryPayload {
  name?: string
  address?: string
  is_active?: boolean
}

export interface AssignDormitoriesPayload {
  dormitory_ids: string[]
}
