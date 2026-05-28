import api from '@/lib/axios'
import type { ApiResponse } from '@/types/api'

export interface SearchTenant {
  id: string
  first_name: string
  last_name: string
  phone: string
}

export interface SearchRoom {
  id: string
  room_number: string
  floor: number
  type: string
  status: string
  rent_price: number
}

export interface SearchBill {
  id: string
  billing_month: string
  total_amount: number
  status: string
  tenant_first_name: string
  tenant_last_name: string
  room_number: string
}

export interface SearchContract {
  id: string
  start_date: string
  end_date: string | null
  status: string
  tenant_first_name: string
  tenant_last_name: string
  room_number: string
}

export interface SearchResults {
  tenants: SearchTenant[]
  rooms: SearchRoom[]
  bills: SearchBill[]
  contracts: SearchContract[]
}

export const searchService = {
  async search(q: string): Promise<SearchResults> {
    const { data } = await api.get<ApiResponse<SearchResults>>(
      `/search?q=${encodeURIComponent(q)}`
    )
    return data.data
  },
}
