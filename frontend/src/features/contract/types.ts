export type ContractStatus = 'active' | 'expired' | 'terminated'

export type ApiContract = {
  id: string
  tenant_id: string
  tenant_name?: string
  room_id: string
  room_number?: string
  dormitory_id?: string
  dormitory_name?: string
  start_date: string
  end_date?: string
  rent_price: number
  deposit: number
  num_occupants: number
  status: ContractStatus
  note: string
  created_at: string
  updated_at: string
}
