export type PaymentMethod = 'cash' | 'transfer' | 'qr'

export interface ApiPayment {
  id: string
  bill_id: string
  amount: number
  method: PaymentMethod
  payment_date: string
  note: string
  created_at: string
  updated_at: string
  room_number: string
  tenant_first_name: string
  tenant_last_name: string
  billing_month: string
  bill_total: number
}

export interface CreatePaymentPayload {
  bill_id: string
  amount: number
  method: PaymentMethod
  payment_date: string
  note: string
}

export interface UpdatePaymentPayload {
  amount: number
  method: PaymentMethod
  payment_date: string
  note: string
}
