import type { ExpenseCategory } from './types'

export const EXPENSE_PAGE_SIZE_OPTIONS = [10, 20, 50, 100] as const

export const EXPENSE_CATEGORIES: ExpenseCategory[] = ['maintenance', 'utility', 'salary', 'supplies', 'other']

export function toDateInputValue(value?: string): string {
  return value ? value.slice(0, 10) : ''
}

export function toApiDate(value: string): string {
  return `${value}T00:00:00Z`
}
