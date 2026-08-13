import { useEffect, useMemo, useState, type FormEvent } from 'react'
import { api, extractErrorMessage, type ApiPage } from '@/shared/api/client'
import { useLanguage } from '@/shared/i18n/language'
import { ConfirmDialog } from '@/shared/components/confirm-dialog'
import type { ApiTenant } from '@/features/tenant/types'
import { ContractListCard } from './components/ContractListCard'
import { ContractFormSheet } from './components/ContractFormSheet'
import type { ApiContract, ContractStatus } from './types'
import { CONTRACT_PAGE_SIZE_OPTIONS, toApiDate, toDateInputValue } from './utils'

export default function ContractPage() {
  const { t } = useLanguage()

  const [contracts, setContracts] = useState<ApiContract[] | null>(null)
  const [tenants, setTenants] = useState<ApiTenant[]>([])
  const [loadError, setLoadError] = useState<string | null>(null)

  const [query, setQuery] = useState('')
  const [page, setPage] = useState(1)
  const [pageSize, setPageSize] = useState<number>(CONTRACT_PAGE_SIZE_OPTIONS[0])

  const [formOpen, setFormOpen] = useState(false)
  const [formContractId, setFormContractId] = useState<string | null>(null)
  const [formTenantId, setFormTenantId] = useState('')
  const [formTenantDisplayName, setFormTenantDisplayName] = useState('')
  const [formRoomId, setFormRoomId] = useState('')
  const [formRoomDisplayLabel, setFormRoomDisplayLabel] = useState('')
  const [formStartDate, setFormStartDate] = useState('')
  const [formEndDate, setFormEndDate] = useState('')
  const [formRentPrice, setFormRentPrice] = useState('')
  const [formDeposit, setFormDeposit] = useState('')
  const [formNumOccupants, setFormNumOccupants] = useState('1')
  const [formStatus, setFormStatus] = useState<ContractStatus>('active')
  const [formNote, setFormNote] = useState('')
  const [formSaving, setFormSaving] = useState(false)
  const [formError, setFormError] = useState<string | null>(null)

  const [deletingContractId, setDeletingContractId] = useState<string | null>(null)
  const [deleteError, setDeleteError] = useState<string | null>(null)
  const [confirmDeleteContract, setConfirmDeleteContract] = useState<ApiContract | null>(null)

  useEffect(() => {
    let cancelled = false

    Promise.all([
      api.get<ApiPage<ApiContract[]>>('/contracts', { params: { per_page: 100 } }),
      api.get<ApiTenant[]>('/tenants/active', { params: { limit: 100 } }),
    ])
      .then(([contractsRes, tenantsRes]) => {
        if (cancelled) return
        setContracts(contractsRes.data.data)
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

  const filteredContracts = useMemo(() => {
    const q = query.trim().toLocaleLowerCase()
    if (!q) return contracts ?? []

    return (contracts ?? []).filter((contract) => {
      return (
        (contract.tenant_name ?? '').toLocaleLowerCase().includes(q) ||
        (contract.room_number ?? '').toLocaleLowerCase().includes(q) ||
        (contract.dormitory_name ?? '').toLocaleLowerCase().includes(q)
      )
    })
  }, [query, contracts])

  const totalPages = Math.max(1, Math.ceil(filteredContracts.length / pageSize))
  const currentPage = Math.min(page, totalPages)
  const rangeStart = filteredContracts.length === 0 ? 0 : (currentPage - 1) * pageSize + 1
  const rangeEnd = Math.min(currentPage * pageSize, filteredContracts.length)
  const paginatedContracts = filteredContracts.slice((currentPage - 1) * pageSize, currentPage * pageSize)

  const isLoading = !loadError && contracts === null

  function openCreateForm() {
    setFormContractId(null)
    setFormTenantId('')
    setFormTenantDisplayName('')
    setFormRoomId('')
    setFormRoomDisplayLabel('')
    setFormStartDate('')
    setFormEndDate('')
    setFormRentPrice('')
    setFormDeposit('')
    setFormNumOccupants('1')
    setFormStatus('active')
    setFormNote('')
    setFormError(null)
    setFormOpen(true)
  }

  function openEditForm(contract: ApiContract) {
    setFormContractId(contract.id)
    setFormTenantId(contract.tenant_id)
    setFormTenantDisplayName(contract.tenant_name ?? '')
    setFormRoomId(contract.room_id)
    setFormRoomDisplayLabel(
      [contract.room_number, contract.dormitory_name].filter(Boolean).join(' - ')
    )
    setFormStartDate(toDateInputValue(contract.start_date))
    setFormEndDate(toDateInputValue(contract.end_date))
    setFormRentPrice(String(contract.rent_price))
    setFormDeposit(String(contract.deposit))
    setFormNumOccupants(String(contract.num_occupants))
    setFormStatus(contract.status)
    setFormNote(contract.note)
    setFormError(null)
    setFormOpen(true)
  }

  async function handleFormSubmit(event: FormEvent<HTMLFormElement>) {
    event.preventDefault()

    const isEdit = formContractId !== null

    if (!isEdit && !formTenantId) {
      setFormError(t('contractTenantRequired'))
      return
    }
    if (!isEdit && !formRoomId) {
      setFormError(t('contractRoomRequired'))
      return
    }
    if (!isEdit && !formStartDate) {
      setFormError(t('contractStartDateRequired'))
      return
    }

    const rentPrice = Number(formRentPrice)
    const deposit = Number(formDeposit)
    const numOccupants = Number(formNumOccupants)

    if (!Number.isFinite(rentPrice) || rentPrice < 0) {
      setFormError(t('contractRentPriceInvalid'))
      return
    }
    if (!Number.isFinite(deposit) || deposit < 0) {
      setFormError(t('contractDepositInvalid'))
      return
    }
    if (!Number.isInteger(numOccupants) || numOccupants < 1) {
      setFormError(t('contractNumOccupantsInvalid'))
      return
    }

    setFormSaving(true)
    setFormError(null)

    try {
      if (!isEdit) {
        const payload = {
          tenant_id: formTenantId,
          room_id: formRoomId,
          start_date: toApiDate(formStartDate),
          end_date: formEndDate ? toApiDate(formEndDate) : null,
          rent_price: rentPrice,
          deposit,
          num_occupants: numOccupants,
          note: formNote.trim(),
        }
        const { data } = await api.post<ApiContract>('/contracts', payload)
        setContracts((prev) => [...(prev ?? []), data])
      } else {
        const payload = {
          end_date: formEndDate ? toApiDate(formEndDate) : null,
          rent_price: rentPrice,
          deposit,
          num_occupants: numOccupants,
          status: formStatus,
          note: formNote.trim(),
        }
        const { data } = await api.put<ApiContract>(`/contracts/${formContractId}`, payload)
        setContracts((prev) => prev?.map((item) => (item.id === data.id ? data : item)) ?? prev)
      }
      setFormOpen(false)
    } catch (err) {
      const fallback = isEdit ? t('contractUpdateError') : t('contractCreateError')
      setFormError(extractErrorMessage(err, fallback))
    } finally {
      setFormSaving(false)
    }
  }

  async function handleDeleteContract() {
    if (!confirmDeleteContract) return
    const contract = confirmDeleteContract

    setDeletingContractId(contract.id)
    setDeleteError(null)

    try {
      await api.delete(`/contracts/${contract.id}`)
      setContracts((prev) => prev?.filter((item) => item.id !== contract.id) ?? prev)
      setConfirmDeleteContract(null)
    } catch (err) {
      setDeleteError(extractErrorMessage(err, t('contractDeleteError')))
    } finally {
      setDeletingContractId(null)
    }
  }

  return (
    <main className="content">
      <section className="welcome">
        <h1>{t('menuContracts')}</h1>
        <p>{t('menuContractsDescription')}</p>
      </section>

      <ContractListCard
        isLoading={isLoading}
        loadError={loadError}
        deleteError={deleteError}
        contracts={contracts}
        query={query}
        onQueryChange={(value) => {
          setQuery(value)
          setPage(1)
        }}
        filteredContracts={filteredContracts}
        paginatedContracts={paginatedContracts}
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
        deletingContractId={deletingContractId}
        onCreateContract={openCreateForm}
        onEditContract={openEditForm}
        onDeleteContract={setConfirmDeleteContract}
      />

      <ConfirmDialog
        open={confirmDeleteContract !== null}
        onOpenChange={(open) => !open && setConfirmDeleteContract(null)}
        title={t('confirmDeleteTitle')}
        description={t('contractDeleteConfirm')}
        confirmLabel={t('contractDelete')}
        cancelLabel={t('cancel')}
        loading={deletingContractId === confirmDeleteContract?.id}
        onConfirm={handleDeleteContract}
      />

      <ContractFormSheet
        open={formOpen}
        onOpenChange={setFormOpen}
        isEdit={formContractId !== null}
        tenantId={formTenantId}
        onTenantIdChange={setFormTenantId}
        tenants={tenants}
        tenantDisplayName={formTenantDisplayName}
        onSelectRoom={(room) => {
          setFormRoomId(room.id)
          setFormRoomDisplayLabel([room.room_number, room.dormitory_name].filter(Boolean).join(' - '))
        }}
        onClearRoomSelection={() => {
          setFormRoomId('')
          setFormRoomDisplayLabel('')
        }}
        roomDisplayLabel={formRoomDisplayLabel}
        startDate={formStartDate}
        onStartDateChange={setFormStartDate}
        endDate={formEndDate}
        onEndDateChange={setFormEndDate}
        rentPrice={formRentPrice}
        onRentPriceChange={setFormRentPrice}
        deposit={formDeposit}
        onDepositChange={setFormDeposit}
        numOccupants={formNumOccupants}
        onNumOccupantsChange={setFormNumOccupants}
        status={formStatus}
        onStatusChange={setFormStatus}
        note={formNote}
        onNoteChange={setFormNote}
        saving={formSaving}
        error={formError}
        onSubmit={handleFormSubmit}
      />
    </main>
  )
}
