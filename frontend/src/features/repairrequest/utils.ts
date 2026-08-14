import type { RepairCategory, RepairStatus } from './types'

export const REPAIR_REQUEST_PAGE_SIZE_OPTIONS = [10, 20, 50, 100] as const

export const REPAIR_CATEGORIES: RepairCategory[] = ['electrical', 'plumbing', 'furniture', 'aircon', 'other']
export const REPAIR_STATUSES: RepairStatus[] = ['pending', 'in_progress', 'completed', 'cancelled']

export function toDateInputValue(value?: string): string {
  return value ? value.slice(0, 10) : ''
}

export function toApiDate(value: string): string {
  return `${value}T00:00:00Z`
}
