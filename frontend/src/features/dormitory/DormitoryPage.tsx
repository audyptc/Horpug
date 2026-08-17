import { useEffect, useMemo, useState, type FormEvent } from 'react'
import { api, extractErrorMessage, type ApiPage } from '@/shared/api/client'
import { usePagination } from '@/shared/hooks/use-pagination'
import { useLanguage } from '@/shared/i18n/language'
import { ConfirmDialog } from '@/shared/components/confirm-dialog'
import { InformationDialog } from '@/shared/components/information-dialog'
import { DormitoryListCard } from './components/DormitoryListCard'
import { DormitoryFormSheet } from './components/DormitoryFormSheet'
import type { ApiDormitory, ApiDormitoryDeletionCheck, ApiUser } from './types'
import { DORMITORY_PAGE_SIZE_OPTIONS } from './utils'

export default function DormitoryPage() {
  const { t } = useLanguage()

  const [dormitories, setDormitories] = useState<ApiDormitory[] | null>(null)
  const [users, setUsers] = useState<ApiUser[]>([])
  const [loadError, setLoadError] = useState<string | null>(null)

  const [query, setQuery] = useState('')

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
    let cancelled = false

    Promise.all([
      api.get<ApiPage<ApiDormitory[]>>('/dormitories', { params: { per_page: 100 } }),
      api.get<ApiPage<ApiUser[]>>('/users', { params: { per_page: 100 } }),
    ])
      .then(([dormitoriesRes, usersRes]) => {
        if (cancelled) return
        setDormitories(dormitoriesRes.data.data)
        setUsers(usersRes.data.data.filter((user) => user.is_active))
      })
      .catch((err) => {
        if (!cancelled) setLoadError(extractErrorMessage(err, t('resourceLoadError')))
      })

    return () => {
      cancelled = true
    }
    // eslint-disable-next-line react-hooks/exhaustive-deps
  }, [])

  const filteredDormitories = useMemo(() => {
    const q = query.trim().toLocaleLowerCase()
    if (!q) return dormitories ?? []

    return (dormitories ?? []).filter((dormitory) => {
      return (
        dormitory.name.toLocaleLowerCase().includes(q) ||
        dormitory.address.toLocaleLowerCase().includes(q) ||
        dormitory.phone.toLocaleLowerCase().includes(q)
      )
    })
  }, [query, dormitories])

  const {
    page: currentPage,
    pageSize,
    setPageSize,
    totalPages,
    rangeStart,
    rangeEnd,
    paginatedItems: paginatedDormitories,
    resetPage,
    prevPage,
    nextPage,
  } = usePagination(filteredDormitories, DORMITORY_PAGE_SIZE_OPTIONS[0])

  const isLoading = !loadError && dormitories === null

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
        const { data } = await api.post<ApiDormitory>('/dormitories', payload)
        setDormitories((prev) => [...(prev ?? []), data])
      } else {
        const { data } = await api.put<ApiDormitory>(`/dormitories/${formDormitoryId}`, payload)
        setDormitories((prev) => prev?.map((item) => (item.id === data.id ? data : item)) ?? prev)
      }
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
      setDormitories((prev) => prev?.filter((item) => item.id !== dormitory.id) ?? prev)
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
        dormitories={dormitories}
        query={query}
        onQueryChange={(value) => {
          setQuery(value)
          resetPage()
        }}
        filteredDormitories={filteredDormitories}
        paginatedDormitories={paginatedDormitories}
        currentPage={currentPage}
        totalPages={totalPages}
        rangeStart={rangeStart}
        rangeEnd={rangeEnd}
        pageSize={pageSize}
        onPageSizeChange={setPageSize}
        onPrevPage={prevPage}
        onNextPage={nextPage}
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
