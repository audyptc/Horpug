export type RoomStatus = 'available' | 'occupied' | 'maintenance'

export type ApiRoom = {
  id: string
  dormitory_id: string
  dormitory_name?: string
  room_type_id: string
  room_type_name?: string
  room_number: string
  floor: number
  status: RoomStatus
  is_active: boolean
  created_at: string
  updated_at: string
}

export type ApiRoomDeletionCheck = {
  can_delete: boolean
  contract_count: number
}
