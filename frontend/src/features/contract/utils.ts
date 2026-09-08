import type { ContractStatus } from './types'

export const CONTRACT_PAGE_SIZE_OPTIONS = [10, 20, 50, 100] as const

export const CONTRACT_STATUSES: ContractStatus[] = ['active', 'expired', 'terminated']

// Mirrors the sort fields the contracts endpoint whitelists; anything else is
// rejected there with a 400.
export type ContractSortKey =
  | 'tenant_name'
  | 'room_number'
  | 'dormitory_name'
  | 'start_date'
  | 'end_date'
  | 'rent_price'
  | 'deposit'
  | 'status'

export type ContractSortDirection = 'asc' | 'desc'
export type ContractStatusFilter = 'all' | ContractStatus

// Mirrors the columns the endpoint accepts as f[<column>]=value; anything else
// is rejected there with a 400.
export const CONTRACT_TEXT_FILTER_KEYS = ['tenant_name', 'room_number', 'dormitory_name'] as const

export type ContractTextFilterKey = (typeof CONTRACT_TEXT_FILTER_KEYS)[number]
export type ContractColumnFilters = Partial<Record<ContractTextFilterKey, string>>

export function isTextFilterKey(key: string): key is ContractTextFilterKey {
  return (CONTRACT_TEXT_FILTER_KEYS as readonly string[]).includes(key)
}

export function toDateInputValue(value?: string): string {
  return value ? value.slice(0, 10) : ''
}

export function toApiDate(value: string): string {
  return `${value}T00:00:00Z`
}
