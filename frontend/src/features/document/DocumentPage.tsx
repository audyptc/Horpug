import { useEffect, useState } from 'react'
import axios from 'axios'
import { api, extractErrorMessage, type ApiPage } from '@/shared/api/client'
import { useLanguage } from '@/shared/i18n/language'
import { ConfirmDialog } from '@/shared/components/confirm-dialog'
import { DocumentListCard } from './components/DocumentListCard'
import { DocumentFormSheet } from './components/DocumentFormSheet'
import { DocumentPreviewSheet } from './components/DocumentPreviewSheet'
import type { ApiDocument } from './types'
import { downloadDocumentFile } from './files'
import { useDocumentForm } from './useDocumentForm'
import {
  DOCUMENT_PAGE_SIZE_OPTIONS,
  type DocumentCategoryFilter,
  type DocumentColumnFilters,
  type DocumentSortDirection,
  type DocumentSortKey,
  type DocumentTextFilterKey,
} from './utils'

const SEARCH_DEBOUNCE_MS = 300

export default function DocumentPage() {
  const { t } = useLanguage()

  const [documents, setDocuments] = useState<ApiDocument[] | null>(null)
  const [total, setTotal] = useState(0)
  const [totalPages, setTotalPages] = useState(1)
  const [loadError, setLoadError] = useState<string | null>(null)

  const [query, setQuery] = useState('')
  const [debouncedQuery, setDebouncedQuery] = useState('')
  const [categoryFilter, setCategoryFilter] = useState<DocumentCategoryFilter>('all')
  // Applied on submit from each column's menu, so no debounce is needed here.
  const [columnFilters, setColumnFilters] = useState<DocumentColumnFilters>({})
  const [sortKey, setSortKey] = useState<DocumentSortKey>('uploaded_date')
  const [sortDirection, setSortDirection] = useState<DocumentSortDirection>('desc')

  const [page, setPage] = useState(1)
  const [pageSize, setPageSizeState] = useState<number>(DOCUMENT_PAGE_SIZE_OPTIONS[0])
  // Bumped by mutations so the list refetches; the server owns the ordering
  // and page boundaries now, so patching rows locally would misplace them.
  const [refreshToken, setRefreshToken] = useState(0)

  const [deletingDocumentId, setDeletingDocumentId] = useState<string | null>(null)
  const [deleteError, setDeleteError] = useState<string | null>(null)
  const [confirmDeleteDocument, setConfirmDeleteDocument] = useState<ApiDocument | null>(null)
  const [previewDocument, setPreviewDocument] = useState<ApiDocument | null>(null)
  const [downloadError, setDownloadError] = useState<string | null>(null)

  const form = useDocumentForm({ onSaved: refresh })

  useEffect(() => {
    const timer = window.setTimeout(() => setDebouncedQuery(query), SEARCH_DEBOUNCE_MS)
    return () => window.clearTimeout(timer)
  }, [query])

  useEffect(() => {
    const controller = new AbortController()

    const columnParams: Record<string, string> = {}
    for (const [key, value] of Object.entries(columnFilters)) {
      if (value) columnParams[`f[${key}]`] = value
    }

    api
      .get<ApiPage<ApiDocument[]>>('/documents', {
        signal: controller.signal,
        params: {
          page,
          per_page: pageSize,
          q: debouncedQuery.trim() || undefined,
          category: categoryFilter === 'all' ? undefined : categoryFilter,
          sort: sortKey,
          order: sortDirection,
          ...columnParams,
        },
      })
      .then(({ data }) => {
        setDocuments(data.data)
        setTotal(data.meta.total)
        setTotalPages(data.meta.total_pages)
        setLoadError(null)
      })
      .catch((err) => {
        // A cancelled request is this effect being superseded by a newer one,
        // not a failure — letting it through would overwrite the newer result.
        if (axios.isCancel(err)) return
        setLoadError(extractErrorMessage(err, t('resourceLoadError')))
      })

    return () => controller.abort()
    // eslint-disable-next-line react-hooks/exhaustive-deps
  }, [page, pageSize, debouncedQuery, categoryFilter, columnFilters, sortKey, sortDirection, refreshToken])

  const isLoading = !loadError && documents === null
  const hasFilters = query !== '' || categoryFilter !== 'all' || Object.values(columnFilters).some(Boolean)
  const rangeStart = total === 0 ? 0 : (page - 1) * pageSize + 1
  const rangeEnd = Math.min(page * pageSize, total)

  function refresh() {
    setRefreshToken((value) => value + 1)
  }

  function setPageSize(size: number) {
    setPageSizeState(size)
    setPage(1)
  }

  function handleSort(key: DocumentSortKey) {
    if (key === sortKey) {
      setSortDirection(sortDirection === 'asc' ? 'desc' : 'asc')
    } else {
      setSortKey(key)
      setSortDirection('asc')
    }
    setPage(1)
  }

  async function handleDeleteDocument() {
    if (!confirmDeleteDocument) return
    const document = confirmDeleteDocument

    setDeletingDocumentId(document.id)
    setDeleteError(null)

    try {
      await api.delete(`/documents/${document.id}`)
      // Deleting the last row of the final page would otherwise strand the
      // view on a page the server no longer has.
      setPage((value) => (documents?.length === 1 ? Math.max(1, value - 1) : value))
      refresh()
      setConfirmDeleteDocument(null)
    } catch (err) {
      setDeleteError(extractErrorMessage(err, t('documentDeleteError')))
    } finally {
      setDeletingDocumentId(null)
    }
  }

  async function handleDownloadDocument(document: ApiDocument) {
    setDownloadError(null)
    try {
      await downloadDocumentFile(document)
    } catch {
      setDownloadError(t('documentDownloadError'))
    }
  }

  return (
    <main className="content">
      <section className="welcome">
        <h1>{t('menuDocuments')}</h1>
        <p>{t('menuDocumentsDescription')}</p>
      </section>

      <DocumentListCard
        isLoading={isLoading}
        loadError={loadError}
        deleteError={deleteError ?? downloadError}
        query={query}
        onQueryChange={(value) => {
          setQuery(value)
          setPage(1)
        }}
        categoryFilter={categoryFilter}
        onCategoryFilterChange={(value) => {
          setCategoryFilter(value)
          setPage(1)
        }}
        columnFilters={columnFilters}
        onColumnFilterChange={(key: DocumentTextFilterKey, value: string) => {
          setColumnFilters((prev) => ({ ...prev, [key]: value }))
          setPage(1)
        }}
        hasFilters={hasFilters}
        sortKey={sortKey}
        sortDirection={sortDirection}
        onSort={handleSort}
        documents={documents ?? []}
        total={total}
        currentPage={page}
        totalPages={totalPages}
        rangeStart={rangeStart}
        rangeEnd={rangeEnd}
        pageSize={pageSize}
        onPageSizeChange={setPageSize}
        onFirstPage={() => setPage(1)}
        onPrevPage={() => setPage((value) => Math.max(1, value - 1))}
        onNextPage={() => setPage((value) => Math.min(totalPages, value + 1))}
        onLastPage={() => setPage(totalPages)}
        deletingDocumentId={deletingDocumentId}
        onCreateDocument={() => form.openCreate()}
        onPreviewDocument={setPreviewDocument}
        onDownloadDocument={handleDownloadDocument}
        onEditDocument={form.openEdit}
        onDeleteDocument={setConfirmDeleteDocument}
      />

      <ConfirmDialog
        open={confirmDeleteDocument !== null}
        onOpenChange={(open) => !open && setConfirmDeleteDocument(null)}
        title={t('confirmDeleteTitle')}
        description={t('documentDeleteConfirm')}
        confirmLabel={t('documentDelete')}
        cancelLabel={t('cancel')}
        loading={deletingDocumentId === confirmDeleteDocument?.id}
        error={deleteError}
        onConfirm={handleDeleteDocument}
      />

      <DocumentPreviewSheet
        document={previewDocument}
        onOpenChange={(open) => !open && setPreviewDocument(null)}
      />

      <DocumentFormSheet {...form.sheetProps} />
    </main>
  )
}
