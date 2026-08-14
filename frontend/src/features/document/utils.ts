import type { DocumentCategory } from './types'

export const DOCUMENT_PAGE_SIZE_OPTIONS = [10, 20, 50, 100] as const

export const DOCUMENT_CATEGORIES: DocumentCategory[] = ['contract', 'id_card', 'receipt', 'other']

export function toDateInputValue(value?: string): string {
  return value ? value.slice(0, 10) : ''
}

export function toApiDate(value: string): string {
  return `${value}T00:00:00Z`
}
