export type WaterBillingType = 'meter' | 'flat'

export interface ApiWaterMeter {
  id: string
  room_id: string
  billing_type: WaterBillingType
  reading_date: string
  previous_reading: number | null
  current_reading: number | null
  unit_price: number | null
  flat_amount: number | null
  unit_used: number | null
  total_amount: number
  note: string
  created_at: string
  updated_at: string
  room_number: string
}

export interface CreateWaterMeterPayload {
  room_id: string
  billing_type: WaterBillingType
  reading_date: string
  previous_reading?: number
  current_reading?: number
  unit_price?: number
  flat_amount?: number
  note: string
}

export interface UpdateWaterMeterPayload {
  billing_type?: WaterBillingType
  reading_date?: string
  previous_reading?: number
  current_reading?: number
  unit_price?: number
  flat_amount?: number
  note?: string
}
