import type { ParcelStatus } from './types'

export const PARCEL_PAGE_SIZE_OPTIONS = [10, 20, 50, 100] as const

export const PARCEL_STATUSES: ParcelStatus[] = ['pending', 'picked_up', 'returned']

// Mirrors the sort fields the parcels endpoint whitelists; anything else is
// rejected there with a 400.
export type ParcelSortKey = 'tenant_name' | 'room_number' | 'courier' | 'tracking_number' | 'status' | 'received_date'
export type ParcelSortDirection = 'asc' | 'desc'
export type ParcelStatusFilter = 'all' | ParcelStatus

// Mirrors the columns the endpoint accepts as f[<column>]=value; anything
// else is rejected there with a 400.
export const PARCEL_TEXT_FILTER_KEYS = ['tenant_name', 'room_number', 'courier', 'tracking_number'] as const

export type ParcelTextFilterKey = (typeof PARCEL_TEXT_FILTER_KEYS)[number]
export type ParcelColumnFilters = Partial<Record<ParcelTextFilterKey, string>>

export function isParcelTextFilterKey(key: string): key is ParcelTextFilterKey {
  return (PARCEL_TEXT_FILTER_KEYS as readonly string[]).includes(key)
}

export function toDateInputValue(value?: string): string {
  return value ? value.slice(0, 10) : ''
}

export function toApiDate(value: string): string {
  return `${value}T00:00:00Z`
}
