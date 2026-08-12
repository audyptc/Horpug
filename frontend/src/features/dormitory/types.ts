export type ApiDormitoryManager = {
  user_id: string
  username: string
  email: string
  created_at: string
}

export type ApiDormitory = {
  id: string
  name: string
  address: string
  phone: string
  description: string
  is_active: boolean
  managers?: ApiDormitoryManager[]
  created_at: string
  updated_at: string
}

export type ApiUser = {
  id: string
  username: string
  email: string
  is_active: boolean
}
