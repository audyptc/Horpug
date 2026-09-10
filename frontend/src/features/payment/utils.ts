import type { PaymentMethod } from './types'

export const PAYMENT_PAGE_SIZE_OPTIONS = [10, 20, 50, 100] as const

export const PAYMENT_METHODS: PaymentMethod[] = ['cash', 'transfer', 'credit_card', 'other']

// Mirrors the sort fields the payments endpoint whitelists; anything else is
// rejected there with a 400.
export type PaymentSortKey = 'tenant_name' | 'room_number' | 'amount' | 'payment_date'
export type PaymentSortDirection = 'asc' | 'desc'
export type PaymentMethodFilter = 'all' | PaymentMethod

// Mirrors the columns the endpoint accepts as f[<column>]=value; anything else
// is rejected there with a 400.
export const PAYMENT_TEXT_FILTER_KEYS = ['tenant_name', 'room_number', 'reference_no'] as const

export type PaymentTextFilterKey = (typeof PAYMENT_TEXT_FILTER_KEYS)[number]
export type PaymentColumnFilters = Partial<Record<PaymentTextFilterKey, string>>

export function isPaymentTextFilterKey(key: string): key is PaymentTextFilterKey {
  return (PAYMENT_TEXT_FILTER_KEYS as readonly string[]).includes(key)
}

// One row of the payment-items form: amount stays a string while edited so
// the input can be empty/partial without fighting number coercion.
export type PaymentItemFormRow = {
  key: number
  paymentMethod: PaymentMethod
  amount: string
  referenceNo: string
}

let nextItemRowKey = 1

export function createPaymentItemRow(): PaymentItemFormRow {
  return { key: nextItemRowKey++, paymentMethod: 'cash', amount: '', referenceNo: '' }
}

export function toDateInputValue(value?: string): string {
  return value ? value.slice(0, 10) : ''
}

export function toApiDate(value: string): string {
  return `${value}T00:00:00Z`
}
