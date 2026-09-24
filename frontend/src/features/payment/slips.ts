export type SlipStatus = 'pending' | 'approved' | 'rejected' | 'cancelled'

// A transfer slip a tenant sent from the LINE pages.
export type ApiSlip = {
  id: string
  invoice_id: string
  invoice_no: string
  tenant_id: string
  tenant_name: string
  room_number: string
  dormitory_id: string
  dormitory_name: string
  period_year: number
  period_month: number
  invoice_total: number
  outstanding: number
  amount: number
  transfer_date: string
  note: string
  file_mime: string
  file_size: number
  status: SlipStatus
  reject_reason?: string
  payment_id?: string
  receipt_no?: string
  reviewed_at?: string
  created_at: string
}
