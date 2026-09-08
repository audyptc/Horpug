import type { BillingMethod } from './types'

export const METER_PAGE_SIZE_OPTIONS = [10, 20, 50, 100] as const

export const METER_BILLING_METHODS: BillingMethod[] = ['metered', 'flat']

// Mirrors the sort fields the meters endpoint whitelists; anything else is
// rejected there with a 400.
export type MeterSortKey =
  | 'room_number'
  | 'dormitory_name'
  | 'reading_date'
  | 'billing_method'
  | 'previous_unit'
  | 'current_unit'
  | 'unit_used'
  | 'price_per_unit'
  | 'total_amount'
  | 'is_billed'

export type MeterSortDirection = 'asc' | 'desc'
export type MeterBillingMethodFilter = 'all' | BillingMethod
export type MeterBilledFilter = 'all' | 'billed' | 'unbilled'

// Mirrors the columns the endpoint accepts as f[<column>]=value; anything else
// is rejected there with a 400.
export const METER_TEXT_FILTER_KEYS = ['room_number', 'dormitory_name'] as const

export type MeterTextFilterKey = (typeof METER_TEXT_FILTER_KEYS)[number]
export type MeterColumnFilters = Partial<Record<MeterTextFilterKey, string>>

export function isTextFilterKey(key: string): key is MeterTextFilterKey {
  return (METER_TEXT_FILTER_KEYS as readonly string[]).includes(key)
}

export function toDateInputValue(value?: string): string {
  return value ? value.slice(0, 10) : ''
}

export function toApiDate(value: string): string {
  return `${value}T00:00:00Z`
}
