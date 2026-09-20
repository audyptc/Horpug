import { useEffect, useState, type FormEvent } from 'react'
import axios from 'axios'
import { api, extractErrorCode, extractErrorMessage, type ApiPage } from '@/shared/api/client'
import { useLanguage } from '@/shared/i18n/language'
import { ConfirmDialog } from '@/shared/components/confirm-dialog'
import type { ApiTenant } from '@/features/tenant/types'
import type { ApiRoom } from '@/features/room/types'
import { formatRoomLabel } from '@/features/room/utils'
import { DocumentListCard } from './components/DocumentListCard'
import { DocumentFormSheet } from './components/DocumentFormSheet'
import { DocumentPreviewSheet } from './components/DocumentPreviewSheet'
import type { ApiDocument, DocumentCategory } from './types'
import { downloadDocumentFile } from './files'
import {
  DOCUMENT_MAX_UPLOAD_BYTES,
  DOCUMENT_PAGE_SIZE_OPTIONS,
  isAllowedUploadName,
  toApiDate,
  toDateInputValue,
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

  const [formOpen, setFormOpen] = useState(false)
  const [formDocumentId, setFormDocumentId] = useState<string | null>(null)
  const [formDormitoryId, setFormDormitoryId] = useState('')
  // The picked dormitory is searched server-side and so isn't necessarily in
  // any loaded list; its name is kept here for the selector's label.
  const [formDormitoryName, setFormDormitoryName] = useState('')
  const [formTenantId, setFormTenantId] = useState('')
  const [formTenantDisplayName, setFormTenantDisplayName] = useState('')
  const [formRoomId, setFormRoomId] = useState('')
  const [formRoomDisplayLabel, setFormRoomDisplayLabel] = useState('')
  const [formName, setFormName] = useState('')
  const [formCategory, setFormCategory] = useState<DocumentCategory>('other')
  const [formFileUrl, setFormFileUrl] = useState('')
  const [formFile, setFormFile] = useState<File | null>(null)
  // Name of the file already stored for the document being edited, if it has
  // one; it stays unless a new file or link is supplied.
  const [formStoredFileName, setFormStoredFileName] = useState<string | null>(null)
  const [formUploadedDate, setFormUploadedDate] = useState('')
  const [formNote, setFormNote] = useState('')
  const [formSaving, setFormSaving] = useState(false)
  const [formError, setFormError] = useState<string | null>(null)

  const [deletingDocumentId, setDeletingDocumentId] = useState<string | null>(null)
  const [deleteError, setDeleteError] = useState<string | null>(null)
  const [confirmDeleteDocument, setConfirmDeleteDocument] = useState<ApiDocument | null>(null)
  const [previewDocument, setPreviewDocument] = useState<ApiDocument | null>(null)
  const [downloadError, setDownloadError] = useState<string | null>(null)

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

  function openCreateForm() {
    setFormDocumentId(null)
    setFormDormitoryId('')
    setFormDormitoryName('')
    setFormTenantId('')
    setFormTenantDisplayName('')
    setFormRoomId('')
    setFormRoomDisplayLabel('')
    setFormName('')
    setFormCategory('other')
    setFormFileUrl('')
    setFormFile(null)
    setFormStoredFileName(null)
    setFormUploadedDate(toDateInputValue(new Date().toISOString()))
    setFormNote('')
    setFormError(null)
    setFormOpen(true)
  }

  function openEditForm(document: ApiDocument) {
    setFormDocumentId(document.id)
    setFormDormitoryId(document.dormitory_id)
    setFormDormitoryName(document.dormitory_name ?? '')
    setFormTenantId(document.tenant_id ?? '')
    setFormTenantDisplayName(document.tenant_name ?? '')
    setFormRoomId(document.room_id ?? '')
    setFormRoomDisplayLabel(formatRoomLabel(document))
    setFormName(document.name)
    setFormCategory(document.category)
    setFormFileUrl(document.file_url)
    setFormFile(null)
    setFormStoredFileName(document.has_file ? document.file_name || document.name : null)
    setFormUploadedDate(toDateInputValue(document.uploaded_date))
    setFormNote(document.note)
    setFormError(null)
    setFormOpen(true)
  }

  async function handleFormSubmit(event: FormEvent<HTMLFormElement>) {
    event.preventDefault()

    const isEdit = formDocumentId !== null

    if (!isEdit && !formDormitoryId) {
      setFormError(t('documentDormitoryRequired'))
      return
    }
    if (!formName.trim()) {
      setFormError(t('documentNameRequired'))
      return
    }
    if (!formFile && !formFileUrl.trim() && !formStoredFileName) {
      setFormError(t('documentFileUrlRequired'))
      return
    }
    if (formFile) {
      if (formFile.size > DOCUMENT_MAX_UPLOAD_BYTES) {
        setFormError(t('documentFileTooLarge'))
        return
      }
      if (formFile.size === 0) {
        setFormError(t('documentFileEmpty'))
        return
      }
      if (!isAllowedUploadName(formFile.name)) {
        setFormError(t('documentFileUnsupported'))
        return
      }
    }

    setFormSaving(true)
    setFormError(null)

    try {
      // A picked file goes up as multipart (the file takes precedence over any
      // link); otherwise it's the plain JSON body. An edit that keeps the
      // stored file and has no link leaves file_url out, so the file isn't
      // replaced by an empty link.
      const fileUrl = formFileUrl.trim()
      const fields = {
        tenant_id: formTenantId || null,
        room_id: formRoomId || null,
        name: formName.trim(),
        category: formCategory,
        uploaded_date: formUploadedDate ? toApiDate(formUploadedDate) : undefined,
        note: formNote.trim(),
      }

      let body: Record<string, unknown> | FormData
      if (formFile) {
        const form = new FormData()
        if (!isEdit) form.append('dormitory_id', formDormitoryId)
        for (const [key, value] of Object.entries(fields)) {
          if (value) form.append(key, value)
        }
        // Empty note is meaningful (it clears it), so it's always sent.
        form.set('note', fields.note)
        form.append('file', formFile)
        body = form
      } else {
        body = {
          ...(isEdit ? {} : { dormitory_id: formDormitoryId }),
          ...fields,
          ...(fileUrl || !formStoredFileName ? { file_url: fileUrl } : {}),
        }
      }

      if (!isEdit) {
        await api.post<ApiDocument>('/documents', body)
      } else {
        await api.put<ApiDocument>(`/documents/${formDocumentId}`, body)
      }
      refresh()
      setFormOpen(false)
    } catch (err) {
      const fallback = isEdit ? t('documentUpdateError') : t('documentCreateError')
      const codeMessages: Record<string, string> = {
        file_too_large: t('documentFileTooLarge'),
        unsupported_file_type: t('documentFileUnsupported'),
        empty_file: t('documentFileEmpty'),
      }
      setFormError(codeMessages[extractErrorCode(err) ?? ''] ?? extractErrorMessage(err, fallback))
    } finally {
      setFormSaving(false)
    }
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
        onCreateDocument={openCreateForm}
        onPreviewDocument={setPreviewDocument}
        onDownloadDocument={handleDownloadDocument}
        onEditDocument={openEditForm}
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

      <DocumentFormSheet
        open={formOpen}
        onOpenChange={setFormOpen}
        isEdit={formDocumentId !== null}
        dormitoryName={formDormitoryName}
        onDormitorySelect={(dormitory) => {
          setFormDormitoryId(dormitory.id)
          setFormDormitoryName(dormitory.name)
        }}
        tenantDisplayName={formTenantDisplayName}
        onSelectTenant={(tenant: ApiTenant) => {
          setFormTenantId(tenant.id)
          setFormTenantDisplayName(`${tenant.first_name} ${tenant.last_name}`)
        }}
        onClearTenant={() => {
          setFormTenantId('')
          setFormTenantDisplayName('')
        }}
        roomDisplayLabel={formRoomDisplayLabel}
        onSelectRoom={(room: ApiRoom) => {
          setFormRoomId(room.id)
          setFormRoomDisplayLabel(formatRoomLabel(room))
        }}
        onClearRoom={() => {
          setFormRoomId('')
          setFormRoomDisplayLabel('')
        }}
        name={formName}
        onNameChange={setFormName}
        category={formCategory}
        onCategoryChange={setFormCategory}
        file={formFile}
        onFileChange={setFormFile}
        storedFileName={formStoredFileName}
        fileUrl={formFileUrl}
        onFileUrlChange={setFormFileUrl}
        uploadedDate={formUploadedDate}
        onUploadedDateChange={setFormUploadedDate}
        note={formNote}
        onNoteChange={setFormNote}
        saving={formSaving}
        error={formError}
        onSubmit={handleFormSubmit}
      />
    </main>
  )
}
