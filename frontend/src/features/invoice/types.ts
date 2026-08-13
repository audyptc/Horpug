export type InvoiceStatus = 'unpaid' | 'paid' | 'overdue' | 'cancelled'

export type InvoiceItemType = 'rent' | 'electricity' | 'water' | 'other'

export type ApiInvoiceItem = {
  id: string
  invoice_id: string
  item_type: InvoiceItemType
  description: string
  reference_id?: string
  amount: number
  created_at: string
}

export type ApiInvoice = {
  id: string
  contract_id: string
  tenant_id?: string
  tenant_name?: string
  room_id?: string
  room_number?: string
  dormitory_id?: string
  dormitory_name?: string
  period_year: number
  period_month: number
  issue_date: string
  due_date: string
  total_amount: number
  status: InvoiceStatus
  paid_at?: string
  note: string
  items?: ApiInvoiceItem[]
  created_at: string
  updated_at: string
}
