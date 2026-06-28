export type PaymentMethod = 'cash' | 'transfer' | 'qr' | 'mixed'

export interface PaymentSplit {
  id: string
  payment_id: string
  method: Exclude<PaymentMethod, 'mixed'>
  amount: number
}

export interface ApiPayment {
  id: string
  bill_id: string
  amount: number
  method: PaymentMethod
  payment_date: string
  note: string
  splits: PaymentSplit[]
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
  payment_date: string
  note: string
  splits: { method: Exclude<PaymentMethod, 'mixed'>; amount: number }[]
}

export interface UpdatePaymentPayload {
  payment_date: string
  note: string
  splits: { method: Exclude<PaymentMethod, 'mixed'>; amount: number }[]
}
