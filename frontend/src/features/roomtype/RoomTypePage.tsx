import { useEffect, useMemo, useState, type FormEvent } from 'react'
import { api, extractErrorMessage, type ApiPage } from '@/shared/api/client'
import { useLanguage } from '@/shared/i18n/language'
import { ConfirmDialog } from '@/shared/components/confirm-dialog'
import type { ApiDormitory } from '@/features/dormitory/types'
import { RoomTypeListCard } from './components/RoomTypeListCard'
import { RoomTypeFormSheet } from './components/RoomTypeFormSheet'
import type { ApiRoomType } from './types'
import { ROOM_TYPE_PAGE_SIZE_OPTIONS } from './utils'

export default function RoomTypePage() {
  const { t } = useLanguage()

  const [roomTypes, setRoomTypes] = useState<ApiRoomType[] | null>(null)
  const [dormitories, setDormitories] = useState<ApiDormitory[]>([])
  const [loadError, setLoadError] = useState<string | null>(null)

  const [query, setQuery] = useState('')
  const [page, setPage] = useState(1)
  const [pageSize, setPageSize] = useState<number>(ROOM_TYPE_PAGE_SIZE_OPTIONS[0])

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
  const [deleteError, setDeleteError] = useState<string | null>(null)
  const [confirmDeleteRoomType, setConfirmDeleteRoomType] = useState<ApiRoomType | null>(null)

  useEffect(() => {
    let cancelled = false

    Promise.all([
      api.get<ApiPage<ApiRoomType[]>>('/room-types', { params: { per_page: 100 } }),
      api.get<ApiDormitory[]>('/dormitories/active', { params: { limit: 100 } }),
    ])
      .then(([roomTypesRes, dormitoriesRes]) => {
        if (cancelled) return
        setRoomTypes(roomTypesRes.data.data)
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

  const filteredRoomTypes = useMemo(() => {
    const q = query.trim().toLocaleLowerCase()
    if (!q) return roomTypes ?? []

    return (roomTypes ?? []).filter((roomType) => {
      return (
        roomType.name.toLocaleLowerCase().includes(q) ||
        (roomType.dormitory_name ?? '').toLocaleLowerCase().includes(q)
      )
    })
  }, [query, roomTypes])

  const totalPages = Math.max(1, Math.ceil(filteredRoomTypes.length / pageSize))
  const currentPage = Math.min(page, totalPages)
  const rangeStart = filteredRoomTypes.length === 0 ? 0 : (currentPage - 1) * pageSize + 1
  const rangeEnd = Math.min(currentPage * pageSize, filteredRoomTypes.length)
  const paginatedRoomTypes = filteredRoomTypes.slice(
    (currentPage - 1) * pageSize,
    currentPage * pageSize
  )

  const isLoading = !loadError && roomTypes === null

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
        const { data } = await api.post<ApiRoomType>('/room-types', payload)
        setRoomTypes((prev) => [...(prev ?? []), data])
      } else {
        const payload = {
          name,
          description: formDescription,
          price,
          is_active: formIsActive,
        }
        const { data } = await api.put<ApiRoomType>(`/room-types/${formRoomTypeId}`, payload)
        setRoomTypes((prev) => prev?.map((item) => (item.id === data.id ? data : item)) ?? prev)
      }
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
      setRoomTypes((prev) => prev?.filter((item) => item.id !== roomType.id) ?? prev)
      setConfirmDeleteRoomType(null)
    } catch (err) {
      setDeleteError(extractErrorMessage(err, t('roomTypeDeleteError')))
    } finally {
      setDeletingRoomTypeId(null)
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
        roomTypes={roomTypes}
        query={query}
        onQueryChange={(value) => {
          setQuery(value)
          setPage(1)
        }}
        filteredRoomTypes={filteredRoomTypes}
        paginatedRoomTypes={paginatedRoomTypes}
        currentPage={currentPage}
        totalPages={totalPages}
        rangeStart={rangeStart}
        rangeEnd={rangeEnd}
        pageSize={pageSize}
        onPageSizeChange={(size) => {
          setPageSize(size)
          setPage(1)
        }}
        onPrevPage={() => setPage((p) => Math.max(1, p - 1))}
        onNextPage={() => setPage((p) => Math.min(totalPages, p + 1))}
        deletingRoomTypeId={deletingRoomTypeId}
        onCreateRoomType={openCreateForm}
        onEditRoomType={openEditForm}
        onDeleteRoomType={setConfirmDeleteRoomType}
      />

      <ConfirmDialog
        open={confirmDeleteRoomType !== null}
        onOpenChange={(open) => !open && setConfirmDeleteRoomType(null)}
        title={t('confirmDeleteTitle')}
        description={t('roomTypeDeleteConfirm')}
        confirmLabel={t('roomTypeDelete')}
        cancelLabel={t('cancel')}
        loading={deletingRoomTypeId === confirmDeleteRoomType?.id}
        onConfirm={handleDeleteRoomType}
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
