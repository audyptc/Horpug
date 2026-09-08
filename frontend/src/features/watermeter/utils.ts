import type { BillingMethod } from './types'

export const WATER_METER_PAGE_SIZE_OPTIONS = [10, 20, 50, 100] as const

export const WATER_METER_BILLING_METHODS: BillingMethod[] = ['metered', 'flat']

// Mirrors the sort fields the water-meters endpoint whitelists; anything else
// is rejected there with a 400.
export type WaterMeterSortKey =
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

export type WaterMeterSortDirection = 'asc' | 'desc'
export type WaterMeterBillingMethodFilter = 'all' | BillingMethod
export type WaterMeterBilledFilter = 'all' | 'billed' | 'unbilled'

// Mirrors the columns the endpoint accepts as f[<column>]=value; anything else
// is rejected there with a 400.
export const WATER_METER_TEXT_FILTER_KEYS = ['room_number', 'dormitory_name'] as const

export type WaterMeterTextFilterKey = (typeof WATER_METER_TEXT_FILTER_KEYS)[number]
export type WaterMeterColumnFilters = Partial<Record<WaterMeterTextFilterKey, string>>

export function isTextFilterKey(key: string): key is WaterMeterTextFilterKey {
  return (WATER_METER_TEXT_FILTER_KEYS as readonly string[]).includes(key)
}

export function toDateInputValue(value?: string): string {
  return value ? value.slice(0, 10) : ''
}

export function toApiDate(value: string): string {
  return `${value}T00:00:00Z`
}
