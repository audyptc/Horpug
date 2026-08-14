export type VehicleType = 'car' | 'motorcycle' | 'other'

export type ApiParking = {
  id: string
  tenant_id: string
  tenant_name?: string
  room_id?: string
  room_number?: string
  dormitory_id?: string
  dormitory_name?: string
  vehicle_type: VehicleType
  license_plate: string
  parking_spot: string
  created_by?: string
  updated_by?: string
  created_at: string
  updated_at: string
}
