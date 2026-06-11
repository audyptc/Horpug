export type RoomType = string
export type RoomStatus = 'available' | 'occupied' | 'maintenance'

export interface ApiRoomType {
  id: string
  name: string
  sort_order: number
  created_at: string
  updated_at: string
}

export interface ApiRoom {
  id: string
  room_number: string
  floor: number
  type: RoomType
  status: RoomStatus
  rent_price: number
  description: string
  created_at: string
  updated_at: string
  updated_by: string
  updated_by_name: string
}

export interface CreateRoomPayload {
  room_number: string
  floor: number
  type: RoomType
  status: RoomStatus
  rent_price: number
  description: string
}

export interface UpdateRoomPayload {
  room_number?: string
  floor?: number
  type?: RoomType
  status?: RoomStatus
  rent_price?: number
  description?: string
}
