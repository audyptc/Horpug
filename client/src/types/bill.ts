export type BillStatus = 'unpaid' | 'paid' | 'overdue'

export interface ApiBill {
  id: string
  contract_id: string
  billing_month: string
  rent_amount: number
  electric_amount: number
  water_amount: number
  other_amount: number
  other_note: string
  total_amount: number
  status: BillStatus
  due_date: string | null
  paid_at: string | null
  note: string
  created_at: string
  updated_at: string
  tenant_first_name: string
  tenant_last_name: string
  room_number: string
}

export interface CreateBillPayload {
  contract_id: string
  billing_month: string
  rent_amount: number
  electric_amount: number
  water_amount: number
  other_amount: number
  other_note: string
  due_date: string | null
  note: string
}

export interface UpdateBillPayload {
  rent_amount: number
  electric_amount: number
  water_amount: number
  other_amount: number
  other_note: string
  status: BillStatus
  due_date: string | null
  note: string
}
