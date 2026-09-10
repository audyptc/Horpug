import type { ExpenseCategory } from './types'

export const EXPENSE_PAGE_SIZE_OPTIONS = [10, 20, 50, 100] as const

export const EXPENSE_CATEGORIES: ExpenseCategory[] = ['maintenance', 'utility', 'salary', 'supplies', 'other']

// Mirrors the sort fields the expenses endpoint whitelists; anything else is
// rejected there with a 400.
export type ExpenseSortKey = 'dormitory_name' | 'category' | 'expense_date' | 'amount' | 'description'
export type ExpenseSortDirection = 'asc' | 'desc'
export type ExpenseCategoryFilter = 'all' | ExpenseCategory

// Mirrors the columns the endpoint accepts as f[<column>]=value; anything else
// is rejected there with a 400.
export const EXPENSE_TEXT_FILTER_KEYS = ['dormitory_name', 'description'] as const

export type ExpenseTextFilterKey = (typeof EXPENSE_TEXT_FILTER_KEYS)[number]
export type ExpenseColumnFilters = Partial<Record<ExpenseTextFilterKey, string>>

export function isExpenseTextFilterKey(key: string): key is ExpenseTextFilterKey {
  return (EXPENSE_TEXT_FILTER_KEYS as readonly string[]).includes(key)
}

export function toDateInputValue(value?: string): string {
  return value ? value.slice(0, 10) : ''
}

export function toApiDate(value: string): string {
  return `${value}T00:00:00Z`
}
