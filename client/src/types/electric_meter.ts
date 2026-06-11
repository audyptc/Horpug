export interface ApiElectricMeter {
  id: string
  room_id: string
  reading_date: string
  previous_reading: number
  current_reading: number
  unit_price: number
  unit_used: number
  total_amount: number
  note: string
  created_at: string
  updated_at: string
  room_number: string
}

export interface CreateElectricMeterPayload {
  room_id: string
  reading_date: string
  previous_reading: number
  current_reading: number
  unit_price: number
  note: string
}

export interface UpdateElectricMeterPayload {
  reading_date?: string
  previous_reading?: number
  current_reading?: number
  unit_price?: number
  note?: string
}
