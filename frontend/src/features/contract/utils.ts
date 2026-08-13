import type { ContractStatus } from './types'

export const CONTRACT_PAGE_SIZE_OPTIONS = [10, 20, 50, 100] as const

export const CONTRACT_STATUSES: ContractStatus[] = ['active', 'expired', 'terminated']

export function toDateInputValue(value?: string): string {
  return value ? value.slice(0, 10) : ''
}

export function toApiDate(value: string): string {
  return `${value}T00:00:00Z`
}
