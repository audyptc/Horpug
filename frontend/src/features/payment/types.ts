export type PaymentMethod = 'cash' | 'transfer' | 'credit_card' | 'other'

export type ApiPaymentItem = {
  id: string
  payment_id: string
  payment_method: PaymentMethod
  amount: number
  reference_no: string
  created_at: string
}

export type ApiPayment = {
  id: string
  invoice_id: string
  tenant_id?: string
  tenant_name?: string
  room_id?: string
  room_number?: string
  dormitory_id?: string
  dormitory_name?: string
  total_amount: number
  payment_date: string
  note: string
  items: ApiPaymentItem[]
  created_by?: string
  created_at: string
}
