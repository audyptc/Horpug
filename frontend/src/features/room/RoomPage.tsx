import { useEffect, useMemo, useState, type FormEvent } from 'react'
import { api, extractErrorMessage, type ApiPage } from '@/shared/api/client'
import { usePagination } from '@/shared/hooks/use-pagination'
import { useLanguage } from '@/shared/i18n/language'
import { ConfirmDialog } from '@/shared/components/confirm-dialog'
import { InformationDialog } from '@/shared/components/information-dialog'
import type { ApiDormitory } from '@/features/dormitory/types'
import type { ApiRoomType } from '@/features/roomtype/types'
import { RoomListCard } from './components/RoomListCard'
import { RoomFormSheet } from './components/RoomFormSheet'
import type { ApiRoom, ApiRoomDeletionCheck, RoomStatus } from './types'
import { ROOM_PAGE_SIZE_OPTIONS } from './utils'

export default function RoomPage() {
  const { t } = useLanguage()

  const [rooms, setRooms] = useState<ApiRoom[] | null>(null)
  const [dormitories, setDormitories] = useState<ApiDormitory[]>([])
  const [loadError, setLoadError] = useState<string | null>(null)

  const [query, setQuery] = useState('')

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
    let cancelled = false

    Promise.all([
      api.get<ApiPage<ApiRoom[]>>('/rooms', { params: { per_page: 100 } }),
      api.get<ApiDormitory[]>('/dormitories/active', { params: { limit: 100 } }),
    ])
      .then(([roomsRes, dormitoriesRes]) => {
        if (cancelled) return
        setRooms(roomsRes.data.data)
        setDormitories(dormitoriesRes.data)
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

  const filteredRooms = useMemo(() => {
    const q = query.trim().toLocaleLowerCase()
    if (!q) return rooms ?? []

    return (rooms ?? []).filter((room) => {
      return (
        room.room_number.toLocaleLowerCase().includes(q) ||
        (room.dormitory_name ?? '').toLocaleLowerCase().includes(q) ||
        (room.room_type_name ?? '').toLocaleLowerCase().includes(q)
      )
    })
  }, [query, rooms])

  const {
    page: currentPage,
    pageSize,
    setPageSize,
    totalPages,
    rangeStart,
    rangeEnd,
    paginatedItems: paginatedRooms,
    resetPage,
    prevPage,
    nextPage,
  } = usePagination(filteredRooms, ROOM_PAGE_SIZE_OPTIONS[0])

  const isLoading = !loadError && rooms === null

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
        const { data } = await api.post<ApiRoom>('/rooms', payload)
        setRooms((prev) => [...(prev ?? []), data])
      } else {
        const payload = {
          room_type_id: formRoomTypeId,
          room_number: roomNumber,
          floor,
          status: formStatus,
          is_active: formIsActive,
        }
        const { data } = await api.put<ApiRoom>(`/rooms/${formRoomId}`, payload)
        setRooms((prev) => prev?.map((item) => (item.id === data.id ? data : item)) ?? prev)
      }
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
      setRooms((prev) => prev?.filter((item) => item.id !== room.id) ?? prev)
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
        rooms={rooms}
        query={query}
        onQueryChange={(value) => {
          setQuery(value)
          resetPage()
        }}
        filteredRooms={filteredRooms}
        paginatedRooms={paginatedRooms}
        currentPage={currentPage}
        totalPages={totalPages}
        rangeStart={rangeStart}
        rangeEnd={rangeEnd}
        pageSize={pageSize}
        onPageSizeChange={setPageSize}
        onPrevPage={prevPage}
        onNextPage={nextPage}
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
