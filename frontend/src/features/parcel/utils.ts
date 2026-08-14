import type { ParcelStatus } from './types'

export const PARCEL_PAGE_SIZE_OPTIONS = [10, 20, 50, 100] as const

export const PARCEL_STATUSES: ParcelStatus[] = ['pending', 'picked_up', 'returned']

export function toDateInputValue(value?: string): string {
  return value ? value.slice(0, 10) : ''
}

export function toApiDate(value: string): string {
  return `${value}T00:00:00Z`
}
