export type VehicleType = 'car' | 'motorcycle'
export type ParkingStatus = 'available' | 'occupied'

export interface ApiParkingSlot {
  id: string
  slot_number: string
  vehicle_type: VehicleType
  status: ParkingStatus
  tenant_id: string | null
  license_plate: string
  monthly_fee: number
  start_date: string | null
  note: string
  created_at: string
  updated_at: string
  tenant_first_name: string
  tenant_last_name: string
}

export interface CreateParkingSlotPayload {
  slot_number: string
  vehicle_type: VehicleType
  status: ParkingStatus
  tenant_id: string | null
  license_plate: string
  monthly_fee: number
  start_date: string | null
  note: string
}

export interface UpdateParkingSlotPayload {
  slot_number: string
  vehicle_type: VehicleType
  status: ParkingStatus
  tenant_id: string | null
  license_plate: string
  monthly_fee: number
  start_date: string | null
  note: string
}
