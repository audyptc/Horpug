import type { RepairCategory, RepairStatus } from './types'

export const REPAIR_REQUEST_PAGE_SIZE_OPTIONS = [10, 20, 50, 100] as const

export const REPAIR_CATEGORIES: RepairCategory[] = ['electrical', 'plumbing', 'furniture', 'aircon', 'other']
export const REPAIR_STATUSES: RepairStatus[] = ['pending', 'in_progress', 'completed', 'cancelled']

// Mirrors the sort fields the repair-requests endpoint whitelists; anything
// else is rejected there with a 400.
export type RepairSortKey = 'room_number' | 'tenant_name' | 'category' | 'status' | 'reported_date' | 'description'
export type RepairSortDirection = 'asc' | 'desc'
export type RepairCategoryFilter = 'all' | RepairCategory
export type RepairStatusFilter = 'all' | RepairStatus

// Mirrors the columns the endpoint accepts as f[<column>]=value; anything
// else is rejected there with a 400.
export const REPAIR_TEXT_FILTER_KEYS = ['room_number', 'tenant_name', 'description'] as const

export type RepairTextFilterKey = (typeof REPAIR_TEXT_FILTER_KEYS)[number]
export type RepairColumnFilters = Partial<Record<RepairTextFilterKey, string>>

export function isRepairTextFilterKey(key: string): key is RepairTextFilterKey {
  return (REPAIR_TEXT_FILTER_KEYS as readonly string[]).includes(key)
}

export function toDateInputValue(value?: string): string {
  return value ? value.slice(0, 10) : ''
}

export function toApiDate(value: string): string {
  return `${value}T00:00:00Z`
}
