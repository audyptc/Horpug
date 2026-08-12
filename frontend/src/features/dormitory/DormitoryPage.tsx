import { useEffect, useMemo, useState, type FormEvent } from 'react'
import { api, extractErrorMessage, type ApiPage } from '@/shared/api/client'
import { useLanguage } from '@/shared/i18n/language'
import { DormitoryListCard } from './components/DormitoryListCard'
import { DormitoryFormSheet } from './components/DormitoryFormSheet'
import type { ApiDormitory, ApiUser } from './types'
import { DORMITORY_PAGE_SIZE_OPTIONS } from './utils'

export default function DormitoryPage() {
  const { t } = useLanguage()

  const [dormitories, setDormitories] = useState<ApiDormitory[] | null>(null)
  const [users, setUsers] = useState<ApiUser[]>([])
  const [loadError, setLoadError] = useState<string | null>(null)

  const [query, setQuery] = useState('')
  const [page, setPage] = useState(1)
  const [pageSize, setPageSize] = useState<number>(DORMITORY_PAGE_SIZE_OPTIONS[0])

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
  const [deleteError, setDeleteError] = useState<string | null>(null)

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

  const totalPages = Math.max(1, Math.ceil(filteredDormitories.length / pageSize))
  const currentPage = Math.min(page, totalPages)
  const rangeStart = filteredDormitories.length === 0 ? 0 : (currentPage - 1) * pageSize + 1
  const rangeEnd = Math.min(currentPage * pageSize, filteredDormitories.length)
  const paginatedDormitories = filteredDormitories.slice(
    (currentPage - 1) * pageSize,
    currentPage * pageSize
  )

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

  async function handleDeleteDormitory(dormitory: ApiDormitory) {
    if (!window.confirm(t('dormitoryDeleteConfirm'))) return

    setDeletingDormitoryId(dormitory.id)
    setDeleteError(null)

    try {
      await api.delete(`/dormitories/${dormitory.id}`)
      setDormitories((prev) => prev?.filter((item) => item.id !== dormitory.id) ?? prev)
    } catch (err) {
      setDeleteError(extractErrorMessage(err, t('dormitoryDeleteError')))
    } finally {
      setDeletingDormitoryId(null)
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
          setPage(1)
        }}
        filteredDormitories={filteredDormitories}
        paginatedDormitories={paginatedDormitories}
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
        deletingDormitoryId={deletingDormitoryId}
        onCreateDormitory={openCreateForm}
        onEditDormitory={openEditForm}
        onDeleteDormitory={handleDeleteDormitory}
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
