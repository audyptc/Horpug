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

export type ApiInvoiceDocument = {
  invoice: ApiInvoice
  dormitory: {
    id: string
    name: string
    address: string
    phone: string
    promptpay_id: string
  }
  payments: {
    id: string
    receipt_no: string
    payment_date: string
    total_amount: number
    note: string
    items: { payment_method: string; amount: number; reference_no: string }[]
  }[]
  paid_amount: number
  outstanding: number
  promptpay_payload?: string
}

export type ApiGenerationCandidate = {
  contract_id: string
  tenant_name: string
  room_id: string
  room_number: string
  rent_price: number
  end_date?: string
  already_invoiced: boolean
  has_electricity: boolean
  has_water: boolean
  ended_before_period: boolean
}

export type ApiGenerationResult = {
  created: {
    invoice_id: string
    contract_id: string
    tenant_name: string
    room_number: string
    total_amount: number
  }[]
  skipped: number
  failed: { contract_id: string; tenant_name: string; room_number: string; error: string }[]
}

export type ApiInvoice = {
  id: string
  contract_id: string
  tenant_id?: string
  tenant_name?: string
  tenant_line_id?: string
  tenant_line_user_id?: string
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
