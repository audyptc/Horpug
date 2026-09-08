import { useEffect, useState, type FormEvent } from 'react'
import axios from 'axios'
import { api, extractErrorMessage, type ApiPage } from '@/shared/api/client'
import { useLanguage } from '@/shared/i18n/language'
import { ConfirmDialog } from '@/shared/components/confirm-dialog'
import { InformationDialog } from '@/shared/components/information-dialog'
import { DormitoryListCard } from './components/DormitoryListCard'
import { DormitoryFormSheet } from './components/DormitoryFormSheet'
import type { ApiDormitory, ApiDormitoryDeletionCheck, ApiUser } from './types'
import {
  DORMITORY_PAGE_SIZE_OPTIONS,
  type DormitoryColumnFilters,
  type DormitorySortDirection,
  type DormitorySortKey,
  type DormitoryStatusFilter,
  type DormitoryTextFilterKey,
} from './utils'

const SEARCH_DEBOUNCE_MS = 300

export default function DormitoryPage() {
  const { t } = useLanguage()

  const [dormitories, setDormitories] = useState<ApiDormitory[] | null>(null)
  const [users, setUsers] = useState<ApiUser[]>([])
  const [total, setTotal] = useState(0)
  const [totalPages, setTotalPages] = useState(1)
  const [loadError, setLoadError] = useState<string | null>(null)

  const [query, setQuery] = useState('')
  const [debouncedQuery, setDebouncedQuery] = useState('')
  const [statusFilter, setStatusFilter] = useState<DormitoryStatusFilter>('all')
  // Applied on submit from each column's menu, so no debounce is needed here.
  const [columnFilters, setColumnFilters] = useState<DormitoryColumnFilters>({})
  const [sortKey, setSortKey] = useState<DormitorySortKey>('name')
  const [sortDirection, setSortDirection] = useState<DormitorySortDirection>('asc')

  const [page, setPage] = useState(1)
  const [pageSize, setPageSizeState] = useState<number>(DORMITORY_PAGE_SIZE_OPTIONS[0])
  // Bumped by mutations so the list refetches; the server owns the ordering
  // and page boundaries now, so patching rows locally would misplace them.
  const [refreshToken, setRefreshToken] = useState(0)

  const [formOpen, setFormOpen] = useState(false)
  const [formDormitoryId, setFormDormitoryId] = useState<string | null>(null)
  const [formName, setFormName] = useState('')
  const [formAddress, setFormAddress] = useState('')
  const [formPhone, setFormPhone] = useState('')
  const [formDescription, setFormDescription] = useState('')
  const [formIsActive, setFormIsActive] = useState(true)
  const [formManagerIds, setFormManagerIds] = useState<string[]>([])
  const [formSaving, setFormSaving] = useState(false)
  const [formError, setFormError] = useState<string | null>(null)

  const [deletingDormitoryId, setDeletingDormitoryId] = useState<string | null>(null)
  const [checkingDormitoryId, setCheckingDormitoryId] = useState<string | null>(null)
  const [deleteError, setDeleteError] = useState<string | null>(null)
  const [confirmDeleteDormitory, setConfirmDeleteDormitory] = useState<ApiDormitory | null>(null)
  const [blockedDeletionRoomCount, setBlockedDeletionRoomCount] = useState<number | null>(null)

  useEffect(() => {
    const timer = window.setTimeout(() => setDebouncedQuery(query), SEARCH_DEBOUNCE_MS)
    return () => window.clearTimeout(timer)
  }, [query])

  // The manager selector in the form isn't affected by the list's filters, so
  // it's loaded once rather than on every refetch.
  useEffect(() => {
    let cancelled = false

    api
      .get<ApiPage<ApiUser[]>>('/users', { params: { per_page: 100 } })
      .then(({ data }) => {
        if (!cancelled) setUsers(data.data.filter((user) => user.is_active))
      })
      .catch((err) => {
        if (!cancelled) setLoadError(extractErrorMessage(err, t('resourceLoadError')))
      })

    return () => {
      cancelled = true
    }
    // eslint-disable-next-line react-hooks/exhaustive-deps
  }, [])

  useEffect(() => {
    const controller = new AbortController()

    const columnParams: Record<string, string> = {}
    for (const [key, value] of Object.entries(columnFilters)) {
      if (value) columnParams[`f[${key}]`] = value
    }

    api
      .get<ApiPage<ApiDormitory[]>>('/dormitories', {
        signal: controller.signal,
        params: {
          page,
          per_page: pageSize,
          q: debouncedQuery.trim() || undefined,
          is_active: statusFilter === 'all' ? undefined : statusFilter === 'active',
          sort: sortKey,
          order: sortDirection,
          ...columnParams,
        },
      })
      .then(({ data }) => {
        setDormitories(data.data)
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
  }, [page, pageSize, debouncedQuery, statusFilter, columnFilters, sortKey, sortDirection, refreshToken])

  const isLoading = !loadError && dormitories === null
  const hasFilters =
    query !== '' || statusFilter !== 'all' || Object.values(columnFilters).some(Boolean)
  const rangeStart = total === 0 ? 0 : (page - 1) * pageSize + 1
  const rangeEnd = Math.min(page * pageSize, total)

  function refresh() {
    setRefreshToken((value) => value + 1)
  }

  function setPageSize(size: number) {
    setPageSizeState(size)
    setPage(1)
  }

  function handleSort(key: DormitorySortKey) {
    if (key === sortKey) {
      setSortDirection(sortDirection === 'asc' ? 'desc' : 'asc')
    } else {
      setSortKey(key)
      setSortDirection('asc')
    }
    setPage(1)
  }

  function openCreateForm() {
    setFormDormitoryId(null)
    setFormName('')
    setFormAddress('')
    setFormPhone('')
    setFormDescription('')
    setFormIsActive(true)
    setFormManagerIds([])
    setFormError(null)
    setFormOpen(true)
  }

  function openEditForm(dormitory: ApiDormitory) {
    setFormDormitoryId(dormitory.id)
    setFormName(dormitory.name)
    setFormAddress(dormitory.address)
    setFormPhone(dormitory.phone)
    setFormDescription(dormitory.description)
    setFormIsActive(dormitory.is_active)
    setFormManagerIds((dormitory.managers ?? []).map((manager) => manager.user_id))
    setFormError(null)
    setFormOpen(true)
  }

  async function handleFormSubmit(event: FormEvent<HTMLFormElement>) {
    event.preventDefault()

    const name = formName.trim()
    if (!name) {
      setFormError(t('dormitoryNameRequired'))
      return
    }

    setFormSaving(true)
    setFormError(null)

    const payload = {
      name,
      address: formAddress,
      phone: formPhone,
      description: formDescription,
      is_active: formIsActive,
      manager_ids: formManagerIds,
    }

    try {
      if (formDormitoryId === null) {
        await api.post<ApiDormitory>('/dormitories', payload)
      } else {
        await api.put<ApiDormitory>(`/dormitories/${formDormitoryId}`, payload)
      }
      refresh()
      setFormOpen(false)
    } catch (err) {
      const fallback = formDormitoryId === null ? t('dormitoryCreateError') : t('dormitoryUpdateError')
      setFormError(extractErrorMessage(err, fallback))
    } finally {
      setFormSaving(false)
    }
  }

  async function handleDeleteDormitory() {
    if (!confirmDeleteDormitory) return
    const dormitory = confirmDeleteDormitory

    setDeletingDormitoryId(dormitory.id)
    setDeleteError(null)

    try {
      await api.delete(`/dormitories/${dormitory.id}`)
      // Deleting the last row of the final page would otherwise strand the
      // view on a page the server no longer has.
      setPage((value) => (dormitories?.length === 1 ? Math.max(1, value - 1) : value))
      refresh()
      setConfirmDeleteDormitory(null)
    } catch (err) {
      const message = extractErrorMessage(err, t('dormitoryDeleteError'))
      setDeleteError(
        message === 'dormitory has rooms and cannot be deleted'
          ? t('dormitoryHasRoomsError')
          : message
      )
    } finally {
      setDeletingDormitoryId(null)
    }
  }

  async function handleRequestDeleteDormitory(dormitory: ApiDormitory) {
    setCheckingDormitoryId(dormitory.id)
    setDeleteError(null)

    try {
      const { data } = await api.get<ApiDormitoryDeletionCheck>(`/dormitories/${dormitory.id}/deletion-check`)
      if (data.can_delete) {
        setConfirmDeleteDormitory(dormitory)
      } else {
        setBlockedDeletionRoomCount(data.room_count)
      }
    } catch (err) {
      setDeleteError(extractErrorMessage(err, t('dormitoryDeleteError')))
    } finally {
      setCheckingDormitoryId(null)
    }
  }

  return (
    <main className="content">
      <section className="welcome">
        <h1>{t('menuDormitories')}</h1>
        <p>{t('menuDormitoriesDescription')}</p>
      </section>

      <DormitoryListCard
        isLoading={isLoading}
        loadError={loadError}
        deleteError={deleteError}
        query={query}
        onQueryChange={(value) => {
          setQuery(value)
          setPage(1)
        }}
        statusFilter={statusFilter}
        onStatusFilterChange={(value) => {
          setStatusFilter(value)
          setPage(1)
        }}
        columnFilters={columnFilters}
        onColumnFilterChange={(key: DormitoryTextFilterKey, value: string) => {
          setColumnFilters((prev) => ({ ...prev, [key]: value }))
          setPage(1)
        }}
        hasFilters={hasFilters}
        sortKey={sortKey}
        sortDirection={sortDirection}
        onSort={handleSort}
        dormitories={dormitories ?? []}
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
        deletingDormitoryId={checkingDormitoryId ?? deletingDormitoryId}
        onCreateDormitory={openCreateForm}
        onEditDormitory={openEditForm}
        onDeleteDormitory={handleRequestDeleteDormitory}
      />

      <ConfirmDialog
        open={confirmDeleteDormitory !== null}
        onOpenChange={(open) => !open && setConfirmDeleteDormitory(null)}
        title={t('confirmDeleteTitle')}
        description={t('dormitoryDeleteConfirm')}
        confirmLabel={t('dormitoryDelete')}
        cancelLabel={t('cancel')}
        loading={deletingDormitoryId === confirmDeleteDormitory?.id}
        error={deleteError}
        onConfirm={handleDeleteDormitory}
      />

      <InformationDialog
        open={blockedDeletionRoomCount !== null}
        onOpenChange={(open) => !open && setBlockedDeletionRoomCount(null)}
        title={t('dormitoryDeleteBlockedTitle')}
        description={t('dormitoryDeleteBlockedDescription').replace('{count}', String(blockedDeletionRoomCount ?? 0))}
        actionLabel={t('acknowledge')}
      />

      <DormitoryFormSheet
        open={formOpen}
        onOpenChange={setFormOpen}
        isEdit={formDormitoryId !== null}
        name={formName}
        onNameChange={setFormName}
        address={formAddress}
        onAddressChange={setFormAddress}
        phone={formPhone}
        onPhoneChange={setFormPhone}
        description={formDescription}
        onDescriptionChange={setFormDescription}
        isActive={formIsActive}
        onIsActiveChange={setFormIsActive}
        users={users}
        managerIds={formManagerIds}
        onManagerIdsChange={setFormManagerIds}
        saving={formSaving}
        error={formError}
        onSubmit={handleFormSubmit}
      />
    </main>
  )
}
