import type { ApiDocument, DocumentCategory } from './types'

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

// Mirrors what the server accepts for upload (see the backend's
// allowedFileTypes); anything else is rejected there.
export const DOCUMENT_UPLOAD_EXTENSIONS = [
  '.pdf',
  '.jpg',
  '.jpeg',
  '.png',
  '.gif',
  '.webp',
  '.doc',
  '.docx',
  '.xls',
  '.xlsx',
  '.ppt',
  '.pptx',
] as const

export const DOCUMENT_UPLOAD_ACCEPT = DOCUMENT_UPLOAD_EXTENSIONS.join(',')
export const DOCUMENT_MAX_UPLOAD_BYTES = 8 * 1024 * 1024

export function isAllowedUploadName(fileName: string): boolean {
  const lower = fileName.toLowerCase()
  return DOCUMENT_UPLOAD_EXTENSIONS.some((extension) => lower.endsWith(extension))
}

export function formatFileSize(bytes: number): string {
  if (bytes < 1024) return `${bytes} B`
  if (bytes < 1024 * 1024) return `${(bytes / 1024).toFixed(0)} KB`
  return `${(bytes / (1024 * 1024)).toFixed(1)} MB`
}

export type DocumentFileKind = 'image' | 'pdf' | 'google' | 'other'

// Kind of a document's file: uploaded ones are known from the stored MIME
// type, linked ones are guessed from the URL.
export function getDocumentKind(document: Pick<ApiDocument, 'has_file' | 'file_mime' | 'file_url'>): DocumentFileKind {
  if (document.has_file) {
    if (document.file_mime?.startsWith('image/')) return 'image'
    if (document.file_mime === 'application/pdf') return 'pdf'
    return 'other'
  }
  return getDocumentFileKind(document.file_url)
}

const IMAGE_EXTENSIONS = ['jpg', 'jpeg', 'png', 'gif', 'webp', 'bmp', 'svg']

function parseHttpUrl(value: string): URL | null {
  try {
    const url = new URL(value.trim())
    return url.protocol === 'http:' || url.protocol === 'https:' ? url : null
  } catch {
    return null
  }
}

// The link is free text entered by a user, so only http(s) ones are ever put
// in an href/src — anything else (e.g. a javascript: URL) is treated as
// unopenable.
export function isOpenableUrl(value: string): boolean {
  return parseHttpUrl(value) !== null
}

// Google Drive/Docs share links open a full viewer page that refuses to be
// framed; swapping the trailing action for /preview gives the embeddable one.
function toGooglePreviewUrl(url: URL): string | null {
  const match = url.pathname.match(/^\/(file|document|spreadsheets|presentation)\/d\/([^/]+)/)
  if (!match) return null

  const host = url.hostname === 'drive.google.com' ? 'drive.google.com' : 'docs.google.com'
  return `https://${host}/${match[1]}/d/${match[2]}/preview`
}

export function getDocumentFileKind(fileUrl: string): DocumentFileKind {
  const url = parseHttpUrl(fileUrl)
  if (!url) return 'other'

  if (
    (url.hostname === 'drive.google.com' || url.hostname === 'docs.google.com') &&
    toGooglePreviewUrl(url)
  ) {
    return 'google'
  }

  const extension = url.pathname.split('.').pop()?.toLowerCase() ?? ''
  if (extension === 'pdf') return 'pdf'
  if (IMAGE_EXTENSIONS.includes(extension)) return 'image'
  return 'other'
}

// The URL to put in the preview frame, or null when the file can't be
// previewed inline (only opened or downloaded).
export function getDocumentPreviewUrl(fileUrl: string): string | null {
  const url = parseHttpUrl(fileUrl)
  if (!url) return null

  switch (getDocumentFileKind(fileUrl)) {
    case 'image':
    case 'pdf':
      return url.toString()
    case 'google':
      return toGooglePreviewUrl(url)
    default:
      return null
  }
}

export function toDateInputValue(value?: string): string {
  return value ? value.slice(0, 10) : ''
}

export function toApiDate(value: string): string {
  return `${value}T00:00:00Z`
}
