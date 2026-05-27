export type RoomType = 'standard' | 'deluxe' | 'suite'
export type RoomStatus = 'available' | 'occupied' | 'maintenance'

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
