import { useEffect, useState, type FormEvent } from 'react'
import axios from 'axios'
import { api, extractErrorMessage, type ApiPage } from '@/shared/api/client'
import { useLanguage } from '@/shared/i18n/language'
import { ConfirmDialog } from '@/shared/components/confirm-dialog'
import { InformationDialog } from '@/shared/components/information-dialog'
import type { ApiDormitory } from '@/features/dormitory/types'
import type { ApiRoomType } from '@/features/roomtype/types'
import { RoomListCard } from './components/RoomListCard'
import { RoomFormSheet } from './components/RoomFormSheet'
import type { ApiRoom, ApiRoomDeletionCheck, RoomStatus } from './types'
import {
  ROOM_PAGE_SIZE_OPTIONS,
  type RoomActiveFilter,
  type RoomColumnFilters,
  type RoomSortDirection,
  type RoomSortKey,
  type RoomStatusFilter,
  type RoomTextFilterKey,
} from './utils'

const SEARCH_DEBOUNCE_MS = 300

export default function RoomPage() {
  const { t } = useLanguage()

  const [rooms, setRooms] = useState<ApiRoom[] | null>(null)
  const [dormitories, setDormitories] = useState<ApiDormitory[]>([])
  const [total, setTotal] = useState(0)
  const [totalPages, setTotalPages] = useState(1)
  const [loadError, setLoadError] = useState<string | null>(null)

  const [query, setQuery] = useState('')
  const [debouncedQuery, setDebouncedQuery] = useState('')
  const [statusFilter, setStatusFilter] = useState<RoomStatusFilter>('all')
  const [activeFilter, setActiveFilter] = useState<RoomActiveFilter>('all')
  // Applied on submit from each column's menu, so no debounce is needed here.
  const [columnFilters, setColumnFilters] = useState<RoomColumnFilters>({})
  const [sortKey, setSortKey] = useState<RoomSortKey>('room_number')
  const [sortDirection, setSortDirection] = useState<RoomSortDirection>('asc')

  const [page, setPage] = useState(1)
  const [pageSize, setPageSizeState] = useState<number>(ROOM_PAGE_SIZE_OPTIONS[0])
  // Bumped by mutations so the list refetches; the server owns the ordering
  // and page boundaries now, so patching rows locally would misplace them.
  const [refreshToken, setRefreshToken] = useState(0)

  const [formOpen, setFormOpen] = useState(false)
  const [formRoomId, setFormRoomId] = useState<string | null>(null)
  const [formDormitoryId, setFormDormitoryId] = useState('')
  const [formRoomTypeId, setFormRoomTypeId] = useState('')
  const [formRoomTypes, setFormRoomTypes] = useState<ApiRoomType[]>([])
  const [formRoomNumber, setFormRoomNumber] = useState('')
  const [formFloor, setFormFloor] = useState('')
  const [formStatus, setFormStatus] = useState<RoomStatus>('available')
  const [formIsActive, setFormIsActive] = useState(true)
  const [formSaving, setFormSaving] = useState(false)
  const [formError, setFormError] = useState<string | null>(null)

  const [deletingRoomId, setDeletingRoomId] = useState<string | null>(null)
  const [checkingRoomId, setCheckingRoomId] = useState<string | null>(null)
  const [deleteError, setDeleteError] = useState<string | null>(null)
  const [confirmDeleteRoom, setConfirmDeleteRoom] = useState<ApiRoom | null>(null)
  const [blockedDeletionContractCount, setBlockedDeletionContractCount] = useState<number | null>(null)

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
    if (!formOpen || !formDormitoryId) return

    let cancelled = false

    api
      .get<ApiRoomType[]>('/room-types/active', { params: { dormitory_id: formDormitoryId, limit: 100 } })
      .then(({ data }) => {
        if (!cancelled) setFormRoomTypes(data)
      })
      .catch(() => {
        if (!cancelled) setFormRoomTypes([])
      })

    return () => {
      cancelled = true
    }
  }, [formOpen, formDormitoryId])

  useEffect(() => {
    const controller = new AbortController()

    const columnParams: Record<string, string> = {}
    for (const [key, value] of Object.entries(columnFilters)) {
      if (value) columnParams[`f[${key}]`] = value
    }

    api
      .get<ApiPage<ApiRoom[]>>('/rooms', {
        signal: controller.signal,
        params: {
          page,
          per_page: pageSize,
          q: debouncedQuery.trim() || undefined,
          is_active: activeFilter === 'all' ? undefined : activeFilter === 'active',
          status: statusFilter === 'all' ? undefined : statusFilter,
          sort: sortKey,
          order: sortDirection,
          ...columnParams,
        },
      })
      .then(({ data }) => {
        setRooms(data.data)
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
  }, [page, pageSize, debouncedQuery, statusFilter, activeFilter, columnFilters, sortKey, sortDirection, refreshToken])

  const isLoading = !loadError && rooms === null
  const hasFilters =
    query !== '' ||
    statusFilter !== 'all' ||
    activeFilter !== 'all' ||
    Object.values(columnFilters).some(Boolean)
  const rangeStart = total === 0 ? 0 : (page - 1) * pageSize + 1
  const rangeEnd = Math.min(page * pageSize, total)

  function refresh() {
    setRefreshToken((value) => value + 1)
  }

  function setPageSize(size: number) {
    setPageSizeState(size)
    setPage(1)
  }

  function handleSort(key: RoomSortKey) {
    if (key === sortKey) {
      setSortDirection(sortDirection === 'asc' ? 'desc' : 'asc')
    } else {
      setSortKey(key)
      setSortDirection('asc')
    }
    setPage(1)
  }

  function openCreateForm() {
    setFormRoomId(null)
    setFormDormitoryId('')
    setFormRoomTypeId('')
    setFormRoomTypes([])
    setFormRoomNumber('')
    setFormFloor('')
    setFormStatus('available')
    setFormIsActive(true)
    setFormError(null)
    setFormOpen(true)
  }

  function openEditForm(room: ApiRoom) {
    setFormRoomId(room.id)
    setFormDormitoryId(room.dormitory_id)
    setFormRoomTypeId(room.room_type_id)
    setFormRoomNumber(room.room_number)
    setFormFloor(String(room.floor))
    setFormStatus(room.status)
    setFormIsActive(room.is_active)
    setFormError(null)
    setFormOpen(true)
  }

  async function handleFormSubmit(event: FormEvent<HTMLFormElement>) {
    event.preventDefault()

    const roomNumber = formRoomNumber.trim()
    if (!roomNumber) {
      setFormError(t('roomNumberRequired'))
      return
    }
    if (formRoomId === null && !formDormitoryId) {
      setFormError(t('roomDormitoryRequired'))
      return
    }
    if (!formRoomTypeId) {
      setFormError(t('roomRoomTypeRequired'))
      return
    }
    const floor = formFloor.trim() === '' ? 0 : Number(formFloor)
    if (!Number.isInteger(floor)) {
      setFormError(t('roomFloorInvalid'))
      return
    }

    setFormSaving(true)
    setFormError(null)

    try {
      if (formRoomId === null) {
        const payload = {
          dormitory_id: formDormitoryId,
          room_type_id: formRoomTypeId,
          room_number: roomNumber,
          floor,
          status: formStatus,
          is_active: formIsActive,
        }
        await api.post<ApiRoom>('/rooms', payload)
      } else {
        const payload = {
          room_type_id: formRoomTypeId,
          room_number: roomNumber,
          floor,
          status: formStatus,
          is_active: formIsActive,
        }
        await api.put<ApiRoom>(`/rooms/${formRoomId}`, payload)
      }
      refresh()
      setFormOpen(false)
    } catch (err) {
      const fallback = formRoomId === null ? t('roomCreateError') : t('roomUpdateError')
      setFormError(extractErrorMessage(err, fallback))
    } finally {
      setFormSaving(false)
    }
  }

  async function handleDeleteRoom() {
    if (!confirmDeleteRoom) return
    const room = confirmDeleteRoom

    setDeletingRoomId(room.id)
    setDeleteError(null)

    try {
      await api.delete(`/rooms/${room.id}`)
      // Deleting the last row of the final page would otherwise strand the
      // view on a page the server no longer has.
      setPage((value) => (rooms?.length === 1 ? Math.max(1, value - 1) : value))
      refresh()
      setConfirmDeleteRoom(null)
    } catch (err) {
      setDeleteError(extractErrorMessage(err, t('roomDeleteError')))
    } finally {
      setDeletingRoomId(null)
    }
  }

  async function handleRequestDeleteRoom(room: ApiRoom) {
    setCheckingRoomId(room.id)
    setDeleteError(null)

    try {
      const { data } = await api.get<ApiRoomDeletionCheck>(`/rooms/${room.id}/deletion-check`)
      if (data.can_delete) {
        setConfirmDeleteRoom(room)
      } else {
        setBlockedDeletionContractCount(data.contract_count)
      }
    } catch (err) {
      setDeleteError(extractErrorMessage(err, t('roomDeleteError')))
    } finally {
      setCheckingRoomId(null)
    }
  }

  return (
    <main className="content">
      <section className="welcome">
        <h1>{t('menuRooms')}</h1>
        <p>{t('menuRoomsDescription')}</p>
      </section>

      <RoomListCard
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
        activeFilter={activeFilter}
        onActiveFilterChange={(value) => {
          setActiveFilter(value)
          setPage(1)
        }}
        columnFilters={columnFilters}
        onColumnFilterChange={(key: RoomTextFilterKey, value: string) => {
          setColumnFilters((prev) => ({ ...prev, [key]: value }))
          setPage(1)
        }}
        hasFilters={hasFilters}
        sortKey={sortKey}
        sortDirection={sortDirection}
        onSort={handleSort}
        rooms={rooms ?? []}
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
        deletingRoomId={checkingRoomId ?? deletingRoomId}
        onCreateRoom={openCreateForm}
        onEditRoom={openEditForm}
        onDeleteRoom={handleRequestDeleteRoom}
      />

      <ConfirmDialog
        open={confirmDeleteRoom !== null}
        onOpenChange={(open) => !open && setConfirmDeleteRoom(null)}
        title={t('confirmDeleteTitle')}
        description={t('roomDeleteConfirm')}
        confirmLabel={t('roomDelete')}
        cancelLabel={t('cancel')}
        loading={deletingRoomId === confirmDeleteRoom?.id}
        error={deleteError}
        onConfirm={handleDeleteRoom}
      />

      <InformationDialog
        open={blockedDeletionContractCount !== null}
        onOpenChange={(open) => !open && setBlockedDeletionContractCount(null)}
        title={t('roomDeleteBlockedTitle')}
        description={t('roomDeleteBlockedDescription').replace('{count}', String(blockedDeletionContractCount ?? 0))}
        actionLabel={t('acknowledge')}
      />

      <RoomFormSheet
        open={formOpen}
        onOpenChange={setFormOpen}
        isEdit={formRoomId !== null}
        dormitoryId={formDormitoryId}
        onDormitoryIdChange={(dormitoryId) => {
          setFormDormitoryId(dormitoryId)
          setFormRoomTypeId('')
          if (!dormitoryId) setFormRoomTypes([])
        }}
        dormitories={dormitories}
        roomTypeId={formRoomTypeId}
        onRoomTypeIdChange={setFormRoomTypeId}
        roomTypes={formRoomTypes}
        roomNumber={formRoomNumber}
        onRoomNumberChange={setFormRoomNumber}
        floor={formFloor}
        onFloorChange={setFormFloor}
        status={formStatus}
        onStatusChange={setFormStatus}
        isActive={formIsActive}
        onIsActiveChange={setFormIsActive}
        saving={formSaving}
        error={formError}
        onSubmit={handleFormSubmit}
      />
    </main>
  )
}
