import { useEffect, useState, type FormEvent } from 'react'
import axios from 'axios'
import { api, extractErrorMessage, type ApiPage } from '@/shared/api/client'
import { useLanguage } from '@/shared/i18n/language'
import { ConfirmDialog } from '@/shared/components/confirm-dialog'
import { InformationDialog } from '@/shared/components/information-dialog'
import type { ApiDormitory } from '@/features/dormitory/types'
import { RoomTypeListCard } from './components/RoomTypeListCard'
import { RoomTypeFormSheet } from './components/RoomTypeFormSheet'
import type { ApiRoomType, ApiRoomTypeDeletionCheck } from './types'
import {
  ROOM_TYPE_PAGE_SIZE_OPTIONS,
  type RoomTypeColumnFilters,
  type RoomTypeSortDirection,
  type RoomTypeSortKey,
  type RoomTypeStatusFilter,
  type RoomTypeTextFilterKey,
} from './utils'

const SEARCH_DEBOUNCE_MS = 300

export default function RoomTypePage() {
  const { t } = useLanguage()

  const [roomTypes, setRoomTypes] = useState<ApiRoomType[] | null>(null)
  const [dormitories, setDormitories] = useState<ApiDormitory[]>([])
  const [total, setTotal] = useState(0)
  const [totalPages, setTotalPages] = useState(1)
  const [loadError, setLoadError] = useState<string | null>(null)

  const [query, setQuery] = useState('')
  const [debouncedQuery, setDebouncedQuery] = useState('')
  const [statusFilter, setStatusFilter] = useState<RoomTypeStatusFilter>('all')
  // Applied on submit from each column's menu, so no debounce is needed here.
  const [columnFilters, setColumnFilters] = useState<RoomTypeColumnFilters>({})
  const [sortKey, setSortKey] = useState<RoomTypeSortKey>('name')
  const [sortDirection, setSortDirection] = useState<RoomTypeSortDirection>('asc')

  const [page, setPage] = useState(1)
  const [pageSize, setPageSizeState] = useState<number>(ROOM_TYPE_PAGE_SIZE_OPTIONS[0])
  // Bumped by mutations so the list refetches; the server owns the ordering
  // and page boundaries now, so patching rows locally would misplace them.
  const [refreshToken, setRefreshToken] = useState(0)

  const [formOpen, setFormOpen] = useState(false)
  const [formRoomTypeId, setFormRoomTypeId] = useState<string | null>(null)
  const [formDormitoryId, setFormDormitoryId] = useState('')
  const [formName, setFormName] = useState('')
  const [formDescription, setFormDescription] = useState('')
  const [formPrice, setFormPrice] = useState('')
  const [formIsActive, setFormIsActive] = useState(true)
  const [formSaving, setFormSaving] = useState(false)
  const [formError, setFormError] = useState<string | null>(null)

  const [deletingRoomTypeId, setDeletingRoomTypeId] = useState<string | null>(null)
  const [checkingRoomTypeId, setCheckingRoomTypeId] = useState<string | null>(null)
  const [deleteError, setDeleteError] = useState<string | null>(null)
  const [confirmDeleteRoomType, setConfirmDeleteRoomType] = useState<ApiRoomType | null>(null)
  const [blockedDeletionRoomCount, setBlockedDeletionRoomCount] = useState<number | null>(null)

  useEffect(() => {
    const timer = window.setTimeout(() => setDebouncedQuery(query), SEARCH_DEBOUNCE_MS)
    return () => window.clearTimeout(timer)
  }, [query])

  // The dormitory selector in the form isn't affected by the list's filters,
  // so it's loaded once rather than on every refetch.
  useEffect(() => {
    let cancelled = false

    api
      .get<ApiDormitory[]>('/dormitories/active', { params: { limit: 100 } })
      .then(({ data }) => {
        if (!cancelled) setDormitories(data)
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
      .get<ApiPage<ApiRoomType[]>>('/room-types', {
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
        setRoomTypes(data.data)
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

  const isLoading = !loadError && roomTypes === null
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

  function handleSort(key: RoomTypeSortKey) {
    if (key === sortKey) {
      setSortDirection(sortDirection === 'asc' ? 'desc' : 'asc')
    } else {
      setSortKey(key)
      setSortDirection('asc')
    }
    setPage(1)
  }

  function openCreateForm() {
    setFormRoomTypeId(null)
    setFormDormitoryId('')
    setFormName('')
    setFormDescription('')
    setFormPrice('')
    setFormIsActive(true)
    setFormError(null)
    setFormOpen(true)
  }

  function openEditForm(roomType: ApiRoomType) {
    setFormRoomTypeId(roomType.id)
    setFormDormitoryId(roomType.dormitory_id)
    setFormName(roomType.name)
    setFormDescription(roomType.description)
    setFormPrice(String(roomType.price))
    setFormIsActive(roomType.is_active)
    setFormError(null)
    setFormOpen(true)
  }

  async function handleFormSubmit(event: FormEvent<HTMLFormElement>) {
    event.preventDefault()

    const name = formName.trim()
    if (!name) {
      setFormError(t('roomTypeNameRequired'))
      return
    }
    if (formRoomTypeId === null && !formDormitoryId) {
      setFormError(t('roomTypeDormitoryRequired'))
      return
    }
    const price = formPrice.trim() === '' ? 0 : Number(formPrice)
    if (Number.isNaN(price) || price < 0) {
      setFormError(t('roomTypePriceInvalid'))
      return
    }

    setFormSaving(true)
    setFormError(null)

    try {
      if (formRoomTypeId === null) {
        const payload = {
          dormitory_id: formDormitoryId,
          name,
          description: formDescription,
          price,
          is_active: formIsActive,
        }
        await api.post<ApiRoomType>('/room-types', payload)
      } else {
        const payload = {
          name,
          description: formDescription,
          price,
          is_active: formIsActive,
        }
        await api.put<ApiRoomType>(`/room-types/${formRoomTypeId}`, payload)
      }
      refresh()
      setFormOpen(false)
    } catch (err) {
      const fallback = formRoomTypeId === null ? t('roomTypeCreateError') : t('roomTypeUpdateError')
      setFormError(extractErrorMessage(err, fallback))
    } finally {
      setFormSaving(false)
    }
  }

  async function handleDeleteRoomType() {
    if (!confirmDeleteRoomType) return
    const roomType = confirmDeleteRoomType

    setDeletingRoomTypeId(roomType.id)
    setDeleteError(null)

    try {
      await api.delete(`/room-types/${roomType.id}`)
      // Deleting the last row of the final page would otherwise strand the
      // view on a page the server no longer has.
      setPage((value) => (roomTypes?.length === 1 ? Math.max(1, value - 1) : value))
      refresh()
      setConfirmDeleteRoomType(null)
    } catch (err) {
      setDeleteError(extractErrorMessage(err, t('roomTypeDeleteError')))
    } finally {
      setDeletingRoomTypeId(null)
    }
  }

  async function handleRequestDeleteRoomType(roomType: ApiRoomType) {
    setCheckingRoomTypeId(roomType.id)
    setDeleteError(null)

    try {
      const { data } = await api.get<ApiRoomTypeDeletionCheck>(`/room-types/${roomType.id}/deletion-check`)
      if (data.can_delete) {
        setConfirmDeleteRoomType(roomType)
      } else {
        setBlockedDeletionRoomCount(data.room_count)
      }
    } catch (err) {
      setDeleteError(extractErrorMessage(err, t('roomTypeDeleteError')))
    } finally {
      setCheckingRoomTypeId(null)
    }
  }

  return (
    <main className="content">
      <section className="welcome">
        <h1>{t('menuRoomTypes')}</h1>
        <p>{t('menuRoomTypesDescription')}</p>
      </section>

      <RoomTypeListCard
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
        onColumnFilterChange={(key: RoomTypeTextFilterKey, value: string) => {
          setColumnFilters((prev) => ({ ...prev, [key]: value }))
          setPage(1)
        }}
        hasFilters={hasFilters}
        sortKey={sortKey}
        sortDirection={sortDirection}
        onSort={handleSort}
        roomTypes={roomTypes ?? []}
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
        deletingRoomTypeId={checkingRoomTypeId ?? deletingRoomTypeId}
        onCreateRoomType={openCreateForm}
        onEditRoomType={openEditForm}
        onDeleteRoomType={handleRequestDeleteRoomType}
      />

      <ConfirmDialog
        open={confirmDeleteRoomType !== null}
        onOpenChange={(open) => !open && setConfirmDeleteRoomType(null)}
        title={t('confirmDeleteTitle')}
        description={t('roomTypeDeleteConfirm')}
        confirmLabel={t('roomTypeDelete')}
        cancelLabel={t('cancel')}
        loading={deletingRoomTypeId === confirmDeleteRoomType?.id}
        error={deleteError}
        onConfirm={handleDeleteRoomType}
      />

      <InformationDialog
        open={blockedDeletionRoomCount !== null}
        onOpenChange={(open) => !open && setBlockedDeletionRoomCount(null)}
        title={t('roomTypeDeleteBlockedTitle')}
        description={t('roomTypeDeleteBlockedDescription').replace('{count}', String(blockedDeletionRoomCount ?? 0))}
        actionLabel={t('acknowledge')}
      />

      <RoomTypeFormSheet
        open={formOpen}
        onOpenChange={setFormOpen}
        isEdit={formRoomTypeId !== null}
        dormitoryId={formDormitoryId}
        onDormitoryIdChange={setFormDormitoryId}
        dormitories={dormitories}
        name={formName}
        onNameChange={setFormName}
        description={formDescription}
        onDescriptionChange={setFormDescription}
        price={formPrice}
        onPriceChange={setFormPrice}
        isActive={formIsActive}
        onIsActiveChange={setFormIsActive}
        saving={formSaving}
        error={formError}
        onSubmit={handleFormSubmit}
      />
    </main>
  )
}
