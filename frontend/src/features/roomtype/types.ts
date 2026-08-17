export type ApiRoomType = {
  id: string
  dormitory_id: string
  dormitory_name?: string
  name: string
  description: string
  price: number
  is_active: boolean
  created_at: string
  updated_at: string
}

export type ApiRoomTypeDeletionCheck = {
  can_delete: boolean
  room_count: number
}
