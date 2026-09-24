export type MoveOutItemType = 'invoice' | 'rent' | 'rent_credit' | 'electricity' | 'water' | 'other'

export type ApiMoveOutItem = {
  item_type: MoveOutItemType
  description: string
  amount: number
  reference_id?: string
}

export type ApiDeductionPreset = {
  id: string
  name: string
  amount: number
}

export type ApiSettlement = {
  contract_id: string
  tenant_name: string
  room_id: string
  room_number: string
  dormitory_id: string
  dormitory_name: string
  start_date: string
  move_out_date: string
  rent_price: number
  deposit: number
  days_stayed: number
  days_in_month: number
  items: ApiMoveOutItem[]
  total_deductions: number
  refund_amount: number
  amount_due: number
}

export type ApiMoveOutPreview = ApiSettlement & {
  presets: ApiDeductionPreset[]
}

export type ApiMoveOut = ApiSettlement & {
  id: string
  dormitory_address: string
  dormitory_phone: string
  note: string
  deposit_payments: { invoice_id: string; payment_id: string; receipt_no: string; amount: number }[]
  created_at: string
}
