import { useEffect, useMemo, useState, type FormEvent } from 'react'
import { api, extractErrorMessage, type ApiPage } from '@/shared/api/client'
import { usePagination } from '@/shared/hooks/use-pagination'
import { useLanguage } from '@/shared/i18n/language'
import { ConfirmDialog } from '@/shared/components/confirm-dialog'
import type { ApiDormitory } from '@/features/dormitory/types'
import type { ApiTenant } from '@/features/tenant/types'
import type { ApiRoom } from '@/features/room/types'
import { DocumentListCard } from './components/DocumentListCard'
import { DocumentFormSheet } from './components/DocumentFormSheet'
import type { ApiDocument, DocumentCategory } from './types'
import { DOCUMENT_PAGE_SIZE_OPTIONS, toApiDate, toDateInputValue } from './utils'

export default function DocumentPage() {
  const { t } = useLanguage()

  const [documents, setDocuments] = useState<ApiDocument[] | null>(null)
  const [dormitories, setDormitories] = useState<ApiDormitory[]>([])
  const [tenants, setTenants] = useState<ApiTenant[]>([])
  const [rooms, setRooms] = useState<ApiRoom[]>([])
  const [loadError, setLoadError] = useState<string | null>(null)

  const [query, setQuery] = useState('')

  const [formOpen, setFormOpen] = useState(false)
  const [formDocumentId, setFormDocumentId] = useState<string | null>(null)
  const [formDormitoryId, setFormDormitoryId] = useState('')
  const [formTenantId, setFormTenantId] = useState('')
  const [formRoomId, setFormRoomId] = useState('')
  const [formName, setFormName] = useState('')
  const [formCategory, setFormCategory] = useState<DocumentCategory>('other')
  const [formFileUrl, setFormFileUrl] = useState('')
  const [formUploadedDate, setFormUploadedDate] = useState('')
  const [formNote, setFormNote] = useState('')
  const [formSaving, setFormSaving] = useState(false)
  const [formError, setFormError] = useState<string | null>(null)

  const [deletingDocumentId, setDeletingDocumentId] = useState<string | null>(null)
  const [deleteError, setDeleteError] = useState<string | null>(null)
  const [confirmDeleteDocument, setConfirmDeleteDocument] = useState<ApiDocument | null>(null)

  useEffect(() => {
    let cancelled = false

    Promise.all([
      api.get<ApiPage<ApiDocument[]>>('/documents', { params: { per_page: 100 } }),
      api.get<ApiDormitory[]>('/dormitories/active', { params: { limit: 100 } }),
      api.get<ApiTenant[]>('/tenants/active', { params: { limit: 100 } }),
      api.get<ApiRoom[]>('/rooms/active', { params: { limit: 100 } }),
    ])
      .then(([documentsRes, dormitoriesRes, tenantsRes, roomsRes]) => {
        if (cancelled) return
        setDocuments(documentsRes.data.data)
        setDormitories(dormitoriesRes.data)
        setTenants(tenantsRes.data)
        setRooms(roomsRes.data)
      })
      .catch((err) => {
        if (!cancelled) setLoadError(extractErrorMessage(err, t('resourceLoadError')))
      })

    return () => {
      cancelled = true
    }
    // eslint-disable-next-line react-hooks/exhaustive-deps
  }, [])

  const filteredDocuments = useMemo(() => {
    const q = query.trim().toLocaleLowerCase()
    if (!q) return documents ?? []

    return (documents ?? []).filter((document) => {
      return (
        document.name.toLocaleLowerCase().includes(q) ||
        (document.dormitory_name ?? '').toLocaleLowerCase().includes(q) ||
        (document.tenant_name ?? '').toLocaleLowerCase().includes(q) ||
        (document.room_number ?? '').toLocaleLowerCase().includes(q)
      )
    })
  }, [query, documents])

  const {
    page: currentPage,
    pageSize,
    setPageSize,
    totalPages,
    rangeStart,
    rangeEnd,
    paginatedItems: paginatedDocuments,
    resetPage,
    prevPage,
    nextPage,
  } = usePagination(filteredDocuments, DOCUMENT_PAGE_SIZE_OPTIONS[0])

  const isLoading = !loadError && documents === null

  function openCreateForm() {
    setFormDocumentId(null)
    setFormDormitoryId('')
    setFormTenantId('')
    setFormRoomId('')
    setFormName('')
    setFormCategory('other')
    setFormFileUrl('')
    setFormUploadedDate(toDateInputValue(new Date().toISOString()))
    setFormNote('')
    setFormError(null)
    setFormOpen(true)
  }

  function openEditForm(document: ApiDocument) {
    setFormDocumentId(document.id)
    setFormDormitoryId(document.dormitory_id)
    setFormTenantId(document.tenant_id ?? '')
    setFormRoomId(document.room_id ?? '')
    setFormName(document.name)
    setFormCategory(document.category)
    setFormFileUrl(document.file_url)
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
    if (!formFileUrl.trim()) {
      setFormError(t('documentFileUrlRequired'))
      return
    }

    setFormSaving(true)
    setFormError(null)

    try {
      if (!isEdit) {
        const payload = {
          dormitory_id: formDormitoryId,
          tenant_id: formTenantId || null,
          room_id: formRoomId || null,
          name: formName.trim(),
          category: formCategory,
          file_url: formFileUrl.trim(),
          uploaded_date: formUploadedDate ? toApiDate(formUploadedDate) : undefined,
          note: formNote.trim(),
        }
        const { data } = await api.post<ApiDocument>('/documents', payload)
        setDocuments((prev) => [data, ...(prev ?? [])])
      } else {
        const payload = {
          tenant_id: formTenantId || null,
          room_id: formRoomId || null,
          name: formName.trim(),
          category: formCategory,
          file_url: formFileUrl.trim(),
          uploaded_date: formUploadedDate ? toApiDate(formUploadedDate) : undefined,
          note: formNote.trim(),
        }
        const { data } = await api.put<ApiDocument>(`/documents/${formDocumentId}`, payload)
        setDocuments((prev) => prev?.map((item) => (item.id === data.id ? data : item)) ?? prev)
      }
      setFormOpen(false)
    } catch (err) {
      const fallback = isEdit ? t('documentUpdateError') : t('documentCreateError')
      setFormError(extractErrorMessage(err, fallback))
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
      setDocuments((prev) => prev?.filter((item) => item.id !== document.id) ?? prev)
      setConfirmDeleteDocument(null)
    } catch (err) {
      setDeleteError(extractErrorMessage(err, t('documentDeleteError')))
    } finally {
      setDeletingDocumentId(null)
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
        deleteError={deleteError}
        documents={documents}
        query={query}
        onQueryChange={(value) => {
          setQuery(value)
          resetPage()
        }}
        filteredDocuments={filteredDocuments}
        paginatedDocuments={paginatedDocuments}
        currentPage={currentPage}
        totalPages={totalPages}
        rangeStart={rangeStart}
        rangeEnd={rangeEnd}
        pageSize={pageSize}
        onPageSizeChange={setPageSize}
        onPrevPage={prevPage}
        onNextPage={nextPage}
        deletingDocumentId={deletingDocumentId}
        onCreateDocument={openCreateForm}
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

      <DocumentFormSheet
        open={formOpen}
        onOpenChange={setFormOpen}
        isEdit={formDocumentId !== null}
        dormitoryId={formDormitoryId}
        onDormitoryIdChange={setFormDormitoryId}
        dormitories={dormitories}
        tenantId={formTenantId}
        onTenantIdChange={setFormTenantId}
        tenants={tenants}
        roomId={formRoomId}
        onRoomIdChange={setFormRoomId}
        rooms={rooms}
        name={formName}
        onNameChange={setFormName}
        category={formCategory}
        onCategoryChange={setFormCategory}
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
