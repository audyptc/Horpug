import { useEffect, useMemo, useState, type FormEvent } from 'react'
import { api, extractErrorMessage, type ApiPage } from '@/shared/api/client'
import { usePagination } from '@/shared/hooks/use-pagination'
import { useLanguage } from '@/shared/i18n/language'
import { ConfirmDialog } from '@/shared/components/confirm-dialog'
import type { ApiRoom } from '@/features/room/types'
import type { ApiTenant } from '@/features/tenant/types'
import { RepairRequestListCard } from './components/RepairRequestListCard'
import { RepairRequestFormSheet } from './components/RepairRequestFormSheet'
import type { ApiRepairRequest, RepairCategory, RepairStatus } from './types'
import { REPAIR_REQUEST_PAGE_SIZE_OPTIONS, toApiDate, toDateInputValue } from './utils'

export default function RepairRequestPage() {
  const { t } = useLanguage()

  const [repairRequests, setRepairRequests] = useState<ApiRepairRequest[] | null>(null)
  const [rooms, setRooms] = useState<ApiRoom[]>([])
  const [tenants, setTenants] = useState<ApiTenant[]>([])
  const [loadError, setLoadError] = useState<string | null>(null)

  const [query, setQuery] = useState('')

  const [formOpen, setFormOpen] = useState(false)
  const [formRepairRequestId, setFormRepairRequestId] = useState<string | null>(null)
  const [formRoomId, setFormRoomId] = useState('')
  const [formRoomDisplayLabel, setFormRoomDisplayLabel] = useState('')
  const [formTenantId, setFormTenantId] = useState('')
  const [formCategory, setFormCategory] = useState<RepairCategory>('other')
  const [formDescription, setFormDescription] = useState('')
  const [formStatus, setFormStatus] = useState<RepairStatus>('pending')
  const [formReportedDate, setFormReportedDate] = useState('')
  const [formSaving, setFormSaving] = useState(false)
  const [formError, setFormError] = useState<string | null>(null)

  const [deletingRepairRequestId, setDeletingRepairRequestId] = useState<string | null>(null)
  const [deleteError, setDeleteError] = useState<string | null>(null)
  const [confirmDeleteRepairRequest, setConfirmDeleteRepairRequest] = useState<ApiRepairRequest | null>(null)

  useEffect(() => {
    let cancelled = false

    Promise.all([
      api.get<ApiPage<ApiRepairRequest[]>>('/repair-requests', { params: { per_page: 100 } }),
      api.get<ApiRoom[]>('/rooms/active', { params: { limit: 100 } }),
      api.get<ApiTenant[]>('/tenants/active', { params: { limit: 100 } }),
    ])
      .then(([repairRequestsRes, roomsRes, tenantsRes]) => {
        if (cancelled) return
        setRepairRequests(repairRequestsRes.data.data)
        setRooms(roomsRes.data)
        setTenants(tenantsRes.data)
      })
      .catch((err) => {
        if (!cancelled) setLoadError(extractErrorMessage(err, t('resourceLoadError')))
      })

    return () => {
      cancelled = true
    }
    // eslint-disable-next-line react-hooks/exhaustive-deps
  }, [])

  const filteredRepairRequests = useMemo(() => {
    const q = query.trim().toLocaleLowerCase()
    if (!q) return repairRequests ?? []

    return (repairRequests ?? []).filter((repairRequest) => {
      return (
        (repairRequest.room_number ?? '').toLocaleLowerCase().includes(q) ||
        (repairRequest.dormitory_name ?? '').toLocaleLowerCase().includes(q) ||
        (repairRequest.tenant_name ?? '').toLocaleLowerCase().includes(q) ||
        repairRequest.description.toLocaleLowerCase().includes(q)
      )
    })
  }, [query, repairRequests])

  const {
    page: currentPage,
    pageSize,
    setPageSize,
    totalPages,
    rangeStart,
    rangeEnd,
    paginatedItems: paginatedRepairRequests,
    resetPage,
    prevPage,
    nextPage,
  } = usePagination(filteredRepairRequests, REPAIR_REQUEST_PAGE_SIZE_OPTIONS[0])

  const isLoading = !loadError && repairRequests === null

  function openCreateForm() {
    setFormRepairRequestId(null)
    setFormRoomId('')
    setFormRoomDisplayLabel('')
    setFormTenantId('')
    setFormCategory('other')
    setFormDescription('')
    setFormStatus('pending')
    setFormReportedDate(toDateInputValue(new Date().toISOString()))
    setFormError(null)
    setFormOpen(true)
  }

  function openEditForm(repairRequest: ApiRepairRequest) {
    setFormRepairRequestId(repairRequest.id)
    setFormRoomId(repairRequest.room_id)
    setFormRoomDisplayLabel(
      [repairRequest.room_number, repairRequest.dormitory_name].filter(Boolean).join(' - ')
    )
    setFormTenantId(repairRequest.tenant_id ?? '')
    setFormCategory(repairRequest.category)
    setFormDescription(repairRequest.description)
    setFormStatus(repairRequest.status)
    setFormReportedDate(toDateInputValue(repairRequest.reported_date))
    setFormError(null)
    setFormOpen(true)
  }

  async function handleFormSubmit(event: FormEvent<HTMLFormElement>) {
    event.preventDefault()

    const isEdit = formRepairRequestId !== null

    if (!isEdit && !formRoomId) {
      setFormError(t('repairRoomRequired'))
      return
    }
    if (!formDescription.trim()) {
      setFormError(t('repairDescriptionRequired'))
      return
    }
    if (!formReportedDate) {
      setFormError(t('repairReportedDateRequired'))
      return
    }

    setFormSaving(true)
    setFormError(null)

    try {
      if (!isEdit) {
        const payload = {
          room_id: formRoomId,
          tenant_id: formTenantId || null,
          category: formCategory,
          description: formDescription.trim(),
          status: formStatus,
          reported_date: toApiDate(formReportedDate),
        }
        const { data } = await api.post<ApiRepairRequest>('/repair-requests', payload)
        setRepairRequests((prev) => [data, ...(prev ?? [])])
      } else {
        const payload = {
          tenant_id: formTenantId || null,
          category: formCategory,
          description: formDescription.trim(),
          status: formStatus,
          reported_date: toApiDate(formReportedDate),
        }
        const { data } = await api.put<ApiRepairRequest>(`/repair-requests/${formRepairRequestId}`, payload)
        setRepairRequests((prev) => prev?.map((item) => (item.id === data.id ? data : item)) ?? prev)
      }
      setFormOpen(false)
    } catch (err) {
      const fallback = isEdit ? t('repairUpdateError') : t('repairCreateError')
      setFormError(extractErrorMessage(err, fallback))
    } finally {
      setFormSaving(false)
    }
  }

  async function handleDeleteRepairRequest() {
    if (!confirmDeleteRepairRequest) return
    const repairRequest = confirmDeleteRepairRequest

    setDeletingRepairRequestId(repairRequest.id)
    setDeleteError(null)

    try {
      await api.delete(`/repair-requests/${repairRequest.id}`)
      setRepairRequests((prev) => prev?.filter((item) => item.id !== repairRequest.id) ?? prev)
      setConfirmDeleteRepairRequest(null)
    } catch (err) {
      setDeleteError(extractErrorMessage(err, t('repairDeleteError')))
    } finally {
      setDeletingRepairRequestId(null)
    }
  }

  return (
    <main className="content">
      <section className="welcome">
        <h1>{t('menuRepairRequests')}</h1>
        <p>{t('menuRepairRequestsDescription')}</p>
      </section>

      <RepairRequestListCard
        isLoading={isLoading}
        loadError={loadError}
        deleteError={deleteError}
        repairRequests={repairRequests}
        query={query}
        onQueryChange={(value) => {
          setQuery(value)
          resetPage()
        }}
        filteredRepairRequests={filteredRepairRequests}
        paginatedRepairRequests={paginatedRepairRequests}
        currentPage={currentPage}
        totalPages={totalPages}
        rangeStart={rangeStart}
        rangeEnd={rangeEnd}
        pageSize={pageSize}
        onPageSizeChange={setPageSize}
        onPrevPage={prevPage}
        onNextPage={nextPage}
        deletingRepairRequestId={deletingRepairRequestId}
        onCreateRepairRequest={openCreateForm}
        onEditRepairRequest={openEditForm}
        onDeleteRepairRequest={setConfirmDeleteRepairRequest}
      />

      <ConfirmDialog
        open={confirmDeleteRepairRequest !== null}
        onOpenChange={(open) => !open && setConfirmDeleteRepairRequest(null)}
        title={t('confirmDeleteTitle')}
        description={t('repairDeleteConfirm')}
        confirmLabel={t('repairDelete')}
        cancelLabel={t('cancel')}
        loading={deletingRepairRequestId === confirmDeleteRepairRequest?.id}
        error={deleteError}
        onConfirm={handleDeleteRepairRequest}
      />

      <RepairRequestFormSheet
        open={formOpen}
        onOpenChange={setFormOpen}
        isEdit={formRepairRequestId !== null}
        roomId={formRoomId}
        onRoomIdChange={setFormRoomId}
        rooms={rooms}
        roomDisplayLabel={formRoomDisplayLabel}
        tenantId={formTenantId}
        onTenantIdChange={setFormTenantId}
        tenants={tenants}
        category={formCategory}
        onCategoryChange={setFormCategory}
        description={formDescription}
        onDescriptionChange={setFormDescription}
        status={formStatus}
        onStatusChange={setFormStatus}
        reportedDate={formReportedDate}
        onReportedDateChange={setFormReportedDate}
        saving={formSaving}
        error={formError}
        onSubmit={handleFormSubmit}
      />
    </main>
  )
}
