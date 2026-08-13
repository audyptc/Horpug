import type { PaymentMethod } from './types'

export const PAYMENT_PAGE_SIZE_OPTIONS = [10, 20, 50, 100] as const

export const PAYMENT_METHODS: PaymentMethod[] = ['cash', 'transfer', 'credit_card', 'other']

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
