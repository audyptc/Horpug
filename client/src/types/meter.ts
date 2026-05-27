export type MeterType = 'electric' | 'water'

export interface ApiMeterReading {
  id: string
  room_id: string
  meter_type: MeterType
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

export interface CreateMeterReadingPayload {
  room_id: string
  meter_type: MeterType
  reading_date: string
  previous_reading: number
  current_reading: number
  unit_price: number
  note: string
}

export interface UpdateMeterReadingPayload {
  reading_date?: string
  previous_reading?: number
  current_reading?: number
  unit_price?: number
  note?: string
}
