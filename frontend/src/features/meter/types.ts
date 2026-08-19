export type BillingMethod = 'metered' | 'flat'

export type ApiMeter = {
  id: string
  room_id: string
  room_number?: string
  dormitory_id?: string
  dormitory_name?: string
  billing_method: BillingMethod
  reading_date: string
  previous_unit: number
  current_unit: number
  unit_used: number
  price_per_unit: number
  flat_amount?: number
  total_amount: number
  is_billed: boolean
  note: string
  created_at: string
  updated_at: string
}
