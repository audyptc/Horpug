export type ContractStatus = 'active' | 'expired' | 'terminated'

export interface ApiContract {
  id: string
  tenant_id: string
  room_id: string
  start_date: string
  end_date: string | null
  rent_price: number
  deposit: number
  num_occupants: number
  status: ContractStatus
  note: string
  created_at: string
  updated_at: string
  created_by: string
  updated_by: string
  updated_by_name: string
  tenant_first_name: string
  tenant_last_name: string
  room_number: string
}

export interface CreateContractPayload {
  tenant_id: string
  room_id: string
  start_date: string
  end_date: string | null
  rent_price: number
  deposit: number
  num_occupants: number
  note: string
}

export interface UpdateContractPayload {
  end_date?: string | null
  rent_price?: number
  deposit?: number
  num_occupants?: number
  status?: ContractStatus
  note?: string
}
