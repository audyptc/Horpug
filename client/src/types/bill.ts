export type BillStatus = 'unpaid' | 'paid' | 'overdue'

export interface BillOtherItem {
  id: string
  bill_id: string
  label: string
  amount: number
  sort_order: number
}

export interface BillOtherItemInput {
  label: string
  amount: number
}

export interface ApiBill {
  id: string
  contract_id: string
  billing_month: string
  rent_amount: number
  electric_amount: number
  water_amount: number
  other_amount: number
  other_items: BillOtherItem[]
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
  other_items: BillOtherItemInput[]
  due_date: string | null
  note: string
}

export interface UpdateBillPayload {
  rent_amount: number
  electric_amount: number
  water_amount: number
  other_items?: BillOtherItemInput[]
  status: BillStatus
  due_date: string | null
  note: string
}
