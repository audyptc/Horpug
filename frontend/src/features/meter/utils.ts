import type { BillingMethod } from './types'

export const METER_PAGE_SIZE_OPTIONS = [10, 20, 50, 100] as const

export const METER_BILLING_METHODS: BillingMethod[] = ['metered', 'flat']

export function toDateInputValue(value?: string): string {
  return value ? value.slice(0, 10) : ''
}

export function toApiDate(value: string): string {
  return `${value}T00:00:00Z`
}
