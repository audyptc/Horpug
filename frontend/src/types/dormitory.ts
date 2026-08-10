export interface ApiDormitory {
  id: string
  name: string
  address: string
  is_active: boolean
  created_at: string
  updated_at: string
}

export interface ApiDormitoryRoleAssignment {
  dormitory_id: string
  dormitory_name: string
  role_id: string
  role_name: string
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

export interface AssignDormitoryRoleItem {
  dormitory_id: string
  role_id: string
}

export interface AssignDormitoriesPayload {
  items: AssignDormitoryRoleItem[]
}
