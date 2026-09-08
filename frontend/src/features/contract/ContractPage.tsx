import { useEffect, useState, type FormEvent } from 'react'
import axios from 'axios'
import { api, extractErrorMessage, type ApiPage } from '@/shared/api/client'
import { useLanguage } from '@/shared/i18n/language'
import { ConfirmDialog } from '@/shared/components/confirm-dialog'
import type { ApiTenant } from '@/features/tenant/types'
import { ContractListCard } from './components/ContractListCard'
import { ContractFormSheet } from './components/ContractFormSheet'
import type { ApiContract, ContractStatus } from './types'
import {
  CONTRACT_PAGE_SIZE_OPTIONS,
  toApiDate,
  toDateInputValue,
  type ContractColumnFilters,
  type ContractSortDirection,
  type ContractSortKey,
  type ContractStatusFilter,
  type ContractTextFilterKey,
} from './utils'

const SEARCH_DEBOUNCE_MS = 300

export default function ContractPage() {
  const { t } = useLanguage()

  const [contracts, setContracts] = useState<ApiContract[] | null>(null)
  const [tenants, setTenants] = useState<ApiTenant[]>([])
  const [total, setTotal] = useState(0)
  const [totalPages, setTotalPages] = useState(1)
  const [loadError, setLoadError] = useState<string | null>(null)

  const [query, setQuery] = useState('')
  const [debouncedQuery, setDebouncedQuery] = useState('')
  const [statusFilter, setStatusFilter] = useState<ContractStatusFilter>('all')
  // Applied on submit from each column's menu, so no debounce is needed here.
  const [columnFilters, setColumnFilters] = useState<ContractColumnFilters>({})
  const [sortKey, setSortKey] = useState<ContractSortKey>('tenant_name')
  const [sortDirection, setSortDirection] = useState<ContractSortDirection>('asc')

  const [page, setPage] = useState(1)
  const [pageSize, setPageSizeState] = useState<number>(CONTRACT_PAGE_SIZE_OPTIONS[0])
  // Bumped by mutations so the list refetches; the server owns the ordering
  // and page boundaries now, so patching rows locally would misplace them.
  const [refreshToken, setRefreshToken] = useState(0)

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
    const timer = window.setTimeout(() => setDebouncedQuery(query), SEARCH_DEBOUNCE_MS)
    return () => window.clearTimeout(timer)
  }, [query])

  // The tenant selector in the form isn't affected by the list's filters, so
  // it's loaded once rather than on every refetch.
  useEffect(() => {
    let cancelled = false

    api
      .get<ApiTenant[]>('/tenants/active', { params: { limit: 100 } })
      .then(({ data }) => {
        if (!cancelled) setTenants(data)
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
      .get<ApiPage<ApiContract[]>>('/contracts', {
        signal: controller.signal,
        params: {
          page,
          per_page: pageSize,
          q: debouncedQuery.trim() || undefined,
          status: statusFilter === 'all' ? undefined : statusFilter,
          sort: sortKey,
          order: sortDirection,
          ...columnParams,
        },
      })
      .then(({ data }) => {
        setContracts(data.data)
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

  const isLoading = !loadError && contracts === null
  const hasFilters = query !== '' || statusFilter !== 'all' || Object.values(columnFilters).some(Boolean)
  const rangeStart = total === 0 ? 0 : (page - 1) * pageSize + 1
  const rangeEnd = Math.min(page * pageSize, total)

  function refresh() {
    setRefreshToken((value) => value + 1)
  }

  function setPageSize(size: number) {
    setPageSizeState(size)
    setPage(1)
  }

  function handleSort(key: ContractSortKey) {
    if (key === sortKey) {
      setSortDirection(sortDirection === 'asc' ? 'desc' : 'asc')
    } else {
      setSortKey(key)
      setSortDirection('asc')
    }
    setPage(1)
  }

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
    if (!formStartDate) {
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
        await api.post<ApiContract>('/contracts', payload)
      } else {
        const payload = {
          start_date: toApiDate(formStartDate),
          end_date: formEndDate ? toApiDate(formEndDate) : null,
          rent_price: rentPrice,
          deposit,
          num_occupants: numOccupants,
          status: formStatus,
          note: formNote.trim(),
        }
        await api.put<ApiContract>(`/contracts/${formContractId}`, payload)
      }
      refresh()
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
      // Deleting the last row of the final page would otherwise strand the
      // view on a page the server no longer has.
      setPage((value) => (contracts?.length === 1 ? Math.max(1, value - 1) : value))
      refresh()
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
        onColumnFilterChange={(key: ContractTextFilterKey, value: string) => {
          setColumnFilters((prev) => ({ ...prev, [key]: value }))
          setPage(1)
        }}
        hasFilters={hasFilters}
        sortKey={sortKey}
        sortDirection={sortDirection}
        onSort={handleSort}
        contracts={contracts ?? []}
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
        error={deleteError}
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
          setFormRentPrice(String(room.room_type_price))
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
