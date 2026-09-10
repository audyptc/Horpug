import { useEffect, useState, type FormEvent } from 'react'
import axios from 'axios'
import { api, extractErrorMessage, type ApiPage } from '@/shared/api/client'
import { useLanguage } from '@/shared/i18n/language'
import { ConfirmDialog } from '@/shared/components/confirm-dialog'
import type { ApiRoom } from '@/features/room/types'
import type { ApiTenant } from '@/features/tenant/types'
import { RepairRequestListCard } from './components/RepairRequestListCard'
import { RepairRequestFormSheet } from './components/RepairRequestFormSheet'
import type { ApiRepairRequest, RepairCategory, RepairStatus } from './types'
import {
  REPAIR_REQUEST_PAGE_SIZE_OPTIONS,
  toApiDate,
  toDateInputValue,
  type RepairCategoryFilter,
  type RepairColumnFilters,
  type RepairSortDirection,
  type RepairSortKey,
  type RepairStatusFilter,
  type RepairTextFilterKey,
} from './utils'

const SEARCH_DEBOUNCE_MS = 300

export default function RepairRequestPage() {
  const { t } = useLanguage()

  const [repairRequests, setRepairRequests] = useState<ApiRepairRequest[] | null>(null)
  const [total, setTotal] = useState(0)
  const [totalPages, setTotalPages] = useState(1)
  const [rooms, setRooms] = useState<ApiRoom[]>([])
  const [tenants, setTenants] = useState<ApiTenant[]>([])
  const [loadError, setLoadError] = useState<string | null>(null)

  const [query, setQuery] = useState('')
  const [debouncedQuery, setDebouncedQuery] = useState('')
  const [categoryFilter, setCategoryFilter] = useState<RepairCategoryFilter>('all')
  const [statusFilter, setStatusFilter] = useState<RepairStatusFilter>('all')
  // Applied on submit from each column's menu, so no debounce is needed here.
  const [columnFilters, setColumnFilters] = useState<RepairColumnFilters>({})
  const [sortKey, setSortKey] = useState<RepairSortKey>('reported_date')
  const [sortDirection, setSortDirection] = useState<RepairSortDirection>('desc')

  const [page, setPage] = useState(1)
  const [pageSize, setPageSizeState] = useState<number>(REPAIR_REQUEST_PAGE_SIZE_OPTIONS[0])
  // Bumped by mutations so the list refetches; the server owns the ordering
  // and page boundaries now, so patching rows locally would misplace them.
  const [refreshToken, setRefreshToken] = useState(0)

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
    const timer = window.setTimeout(() => setDebouncedQuery(query), SEARCH_DEBOUNCE_MS)
    return () => window.clearTimeout(timer)
  }, [query])

  useEffect(() => {
    let cancelled = false

    Promise.all([
      api.get<ApiRoom[]>('/rooms/active', { params: { limit: 100 } }),
      api.get<ApiTenant[]>('/tenants/active', { params: { limit: 100 } }),
    ])
      .then(([roomsRes, tenantsRes]) => {
        if (cancelled) return
        setRooms(roomsRes.data)
        setTenants(tenantsRes.data)
      })
      .catch(() => {
        // Ignore — the create form just shows no room/tenant choices.
      })

    return () => {
      cancelled = true
    }
  }, [])

  useEffect(() => {
    const controller = new AbortController()

    const columnParams: Record<string, string> = {}
    for (const [key, value] of Object.entries(columnFilters)) {
      if (value) columnParams[`f[${key}]`] = value
    }

    api
      .get<ApiPage<ApiRepairRequest[]>>('/repair-requests', {
        signal: controller.signal,
        params: {
          page,
          per_page: pageSize,
          q: debouncedQuery.trim() || undefined,
          category: categoryFilter === 'all' ? undefined : categoryFilter,
          status: statusFilter === 'all' ? undefined : statusFilter,
          sort: sortKey,
          order: sortDirection,
          ...columnParams,
        },
      })
      .then(({ data }) => {
        setRepairRequests(data.data)
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
  }, [page, pageSize, debouncedQuery, categoryFilter, statusFilter, columnFilters, sortKey, sortDirection, refreshToken])

  const isLoading = !loadError && repairRequests === null
  const hasFilters =
    query !== '' || categoryFilter !== 'all' || statusFilter !== 'all' || Object.values(columnFilters).some(Boolean)
  const rangeStart = total === 0 ? 0 : (page - 1) * pageSize + 1
  const rangeEnd = Math.min(page * pageSize, total)

  function refresh() {
    setRefreshToken((value) => value + 1)
  }

  function setPageSize(size: number) {
    setPageSizeState(size)
    setPage(1)
  }

  function handleSort(key: RepairSortKey) {
    if (key === sortKey) {
      setSortDirection(sortDirection === 'asc' ? 'desc' : 'asc')
    } else {
      setSortKey(key)
      setSortDirection('asc')
    }
    setPage(1)
  }

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
        await api.post<ApiRepairRequest>('/repair-requests', payload)
      } else {
        const payload = {
          tenant_id: formTenantId || null,
          category: formCategory,
          description: formDescription.trim(),
          status: formStatus,
          reported_date: toApiDate(formReportedDate),
        }
        await api.put<ApiRepairRequest>(`/repair-requests/${formRepairRequestId}`, payload)
      }
      refresh()
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
      // Deleting the last row of the final page would otherwise strand the
      // view on a page the server no longer has.
      setPage((value) => (repairRequests?.length === 1 ? Math.max(1, value - 1) : value))
      refresh()
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
        statusFilter={statusFilter}
        onStatusFilterChange={(value) => {
          setStatusFilter(value)
          setPage(1)
        }}
        columnFilters={columnFilters}
        onColumnFilterChange={(key: RepairTextFilterKey, value: string) => {
          setColumnFilters((prev) => ({ ...prev, [key]: value }))
          setPage(1)
        }}
        hasFilters={hasFilters}
        sortKey={sortKey}
        sortDirection={sortDirection}
        onSort={handleSort}
        repairRequests={repairRequests ?? []}
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
