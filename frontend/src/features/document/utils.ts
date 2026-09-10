import type { DocumentCategory } from './types'

export const DOCUMENT_PAGE_SIZE_OPTIONS = [10, 20, 50, 100] as const

export const DOCUMENT_CATEGORIES: DocumentCategory[] = ['contract', 'id_card', 'receipt', 'other']

// Mirrors the sort fields the documents endpoint whitelists; anything else is
// rejected there with a 400.
export type DocumentSortKey = 'name' | 'category' | 'dormitory_name' | 'tenant_name' | 'room_number' | 'uploaded_date'
export type DocumentSortDirection = 'asc' | 'desc'
export type DocumentCategoryFilter = 'all' | DocumentCategory

// Mirrors the columns the endpoint accepts as f[<column>]=value; anything
// else is rejected there with a 400.
export const DOCUMENT_TEXT_FILTER_KEYS = ['name', 'dormitory_name', 'tenant_name', 'room_number'] as const

export type DocumentTextFilterKey = (typeof DOCUMENT_TEXT_FILTER_KEYS)[number]
export type DocumentColumnFilters = Partial<Record<DocumentTextFilterKey, string>>

export function isDocumentTextFilterKey(key: string): key is DocumentTextFilterKey {
  return (DOCUMENT_TEXT_FILTER_KEYS as readonly string[]).includes(key)
}

export function toDateInputValue(value?: string): string {
  return value ? value.slice(0, 10) : ''
}

export function toApiDate(value: string): string {
  return `${value}T00:00:00Z`
}
