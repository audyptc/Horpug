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
  promptpay_id: string
  description: string
  is_active: boolean
  managers?: ApiDormitoryManager[]
  created_at: string
  updated_at: string
}

export type ApiDormitoryDeletionCheck = {
  can_delete: boolean
  room_count: number
}

export type ApiUser = {
  id: string
  username: string
  email: string
  is_active: boolean
}

// A user picked as a dormitory manager, as held in the form. It carries the
// display fields so the chosen managers can be shown without being in any
// loaded page of users.
export type FormManager = Pick<ApiUser, 'id' | 'username' | 'email'>
