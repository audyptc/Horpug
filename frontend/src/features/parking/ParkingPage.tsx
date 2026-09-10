import { useEffect, useState, type FormEvent } from 'react'
import axios from 'axios'
import { api, extractErrorMessage, type ApiPage } from '@/shared/api/client'
import { useLanguage } from '@/shared/i18n/language'
import { ConfirmDialog } from '@/shared/components/confirm-dialog'
import type { ApiTenant } from '@/features/tenant/types'
import type { ApiRoom } from '@/features/room/types'
import { ParkingListCard } from './components/ParkingListCard'
import { ParkingFormSheet } from './components/ParkingFormSheet'
import type { ApiParking, VehicleType } from './types'
import {
  PARKING_PAGE_SIZE_OPTIONS,
  type ParkingColumnFilters,
  type ParkingSortDirection,
  type ParkingSortKey,
  type ParkingTextFilterKey,
  type VehicleTypeFilter,
} from './utils'

const SEARCH_DEBOUNCE_MS = 300

export default function ParkingPage() {
  const { t } = useLanguage()

  const [parkings, setParkings] = useState<ApiParking[] | null>(null)
  const [total, setTotal] = useState(0)
  const [totalPages, setTotalPages] = useState(1)
  const [tenants, setTenants] = useState<ApiTenant[]>([])
  const [rooms, setRooms] = useState<ApiRoom[]>([])
  const [loadError, setLoadError] = useState<string | null>(null)

  const [query, setQuery] = useState('')
  const [debouncedQuery, setDebouncedQuery] = useState('')
  const [vehicleTypeFilter, setVehicleTypeFilter] = useState<VehicleTypeFilter>('all')
  // Applied on submit from each column's menu, so no debounce is needed here.
  const [columnFilters, setColumnFilters] = useState<ParkingColumnFilters>({})
  const [sortKey, setSortKey] = useState<ParkingSortKey>('tenant_name')
  const [sortDirection, setSortDirection] = useState<ParkingSortDirection>('asc')

  const [page, setPage] = useState(1)
  const [pageSize, setPageSizeState] = useState<number>(PARKING_PAGE_SIZE_OPTIONS[0])
  // Bumped by mutations so the list refetches; the server owns the ordering
  // and page boundaries now, so patching rows locally would misplace them.
  const [refreshToken, setRefreshToken] = useState(0)

  const [formOpen, setFormOpen] = useState(false)
  const [formParkingId, setFormParkingId] = useState<string | null>(null)
  const [formTenantId, setFormTenantId] = useState('')
  const [formTenantDisplayName, setFormTenantDisplayName] = useState('')
  const [formRoomId, setFormRoomId] = useState('')
  const [formVehicleType, setFormVehicleType] = useState<VehicleType>('motorcycle')
  const [formLicensePlate, setFormLicensePlate] = useState('')
  const [formParkingSpot, setFormParkingSpot] = useState('')
  const [formSaving, setFormSaving] = useState(false)
  const [formError, setFormError] = useState<string | null>(null)

  const [deletingParkingId, setDeletingParkingId] = useState<string | null>(null)
  const [deleteError, setDeleteError] = useState<string | null>(null)
  const [confirmDeleteParking, setConfirmDeleteParking] = useState<ApiParking | null>(null)

  useEffect(() => {
    const timer = window.setTimeout(() => setDebouncedQuery(query), SEARCH_DEBOUNCE_MS)
    return () => window.clearTimeout(timer)
  }, [query])

  useEffect(() => {
    let cancelled = false

    Promise.all([
      api.get<ApiTenant[]>('/tenants/active', { params: { limit: 100 } }),
      api.get<ApiRoom[]>('/rooms/active', { params: { limit: 100 } }),
    ])
      .then(([tenantsRes, roomsRes]) => {
        if (cancelled) return
        setTenants(tenantsRes.data)
        setRooms(roomsRes.data)
      })
      .catch(() => {
        // Ignore — the create form just shows no tenant/room choices.
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
      .get<ApiPage<ApiParking[]>>('/parking', {
        signal: controller.signal,
        params: {
          page,
          per_page: pageSize,
          q: debouncedQuery.trim() || undefined,
          vehicle_type: vehicleTypeFilter === 'all' ? undefined : vehicleTypeFilter,
          sort: sortKey,
          order: sortDirection,
          ...columnParams,
        },
      })
      .then(({ data }) => {
        setParkings(data.data)
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
  }, [page, pageSize, debouncedQuery, vehicleTypeFilter, columnFilters, sortKey, sortDirection, refreshToken])

  const isLoading = !loadError && parkings === null
  const hasFilters = query !== '' || vehicleTypeFilter !== 'all' || Object.values(columnFilters).some(Boolean)
  const rangeStart = total === 0 ? 0 : (page - 1) * pageSize + 1
  const rangeEnd = Math.min(page * pageSize, total)

  function refresh() {
    setRefreshToken((value) => value + 1)
  }

  function setPageSize(size: number) {
    setPageSizeState(size)
    setPage(1)
  }

  function handleSort(key: ParkingSortKey) {
    if (key === sortKey) {
      setSortDirection(sortDirection === 'asc' ? 'desc' : 'asc')
    } else {
      setSortKey(key)
      setSortDirection('asc')
    }
    setPage(1)
  }

  function openCreateForm() {
    setFormParkingId(null)
    setFormTenantId('')
    setFormTenantDisplayName('')
    setFormRoomId('')
    setFormVehicleType('motorcycle')
    setFormLicensePlate('')
    setFormParkingSpot('')
    setFormError(null)
    setFormOpen(true)
  }

  function openEditForm(parking: ApiParking) {
    setFormParkingId(parking.id)
    setFormTenantId(parking.tenant_id)
    setFormTenantDisplayName(parking.tenant_name ?? '')
    setFormRoomId(parking.room_id ?? '')
    setFormVehicleType(parking.vehicle_type)
    setFormLicensePlate(parking.license_plate)
    setFormParkingSpot(parking.parking_spot)
    setFormError(null)
    setFormOpen(true)
  }

  async function handleFormSubmit(event: FormEvent<HTMLFormElement>) {
    event.preventDefault()

    const isEdit = formParkingId !== null

    if (!isEdit && !formTenantId) {
      setFormError(t('parkingTenantRequired'))
      return
    }
    if (!formLicensePlate.trim()) {
      setFormError(t('parkingLicensePlateRequired'))
      return
    }

    setFormSaving(true)
    setFormError(null)

    try {
      if (!isEdit) {
        const payload = {
          tenant_id: formTenantId,
          room_id: formRoomId || null,
          vehicle_type: formVehicleType,
          license_plate: formLicensePlate.trim(),
          parking_spot: formParkingSpot.trim(),
        }
        await api.post<ApiParking>('/parking', payload)
      } else {
        const payload = {
          room_id: formRoomId || null,
          vehicle_type: formVehicleType,
          license_plate: formLicensePlate.trim(),
          parking_spot: formParkingSpot.trim(),
        }
        await api.put<ApiParking>(`/parking/${formParkingId}`, payload)
      }
      refresh()
      setFormOpen(false)
    } catch (err) {
      const fallback = isEdit ? t('parkingUpdateError') : t('parkingCreateError')
      setFormError(extractErrorMessage(err, fallback))
    } finally {
      setFormSaving(false)
    }
  }

  async function handleDeleteParking() {
    if (!confirmDeleteParking) return
    const parking = confirmDeleteParking

    setDeletingParkingId(parking.id)
    setDeleteError(null)

    try {
      await api.delete(`/parking/${parking.id}`)
      // Deleting the last row of the final page would otherwise strand the
      // view on a page the server no longer has.
      setPage((value) => (parkings?.length === 1 ? Math.max(1, value - 1) : value))
      refresh()
      setConfirmDeleteParking(null)
    } catch (err) {
      setDeleteError(extractErrorMessage(err, t('parkingDeleteError')))
    } finally {
      setDeletingParkingId(null)
    }
  }

  return (
    <main className="content">
      <section className="welcome">
        <h1>{t('menuParking')}</h1>
        <p>{t('menuParkingDescription')}</p>
      </section>

      <ParkingListCard
        isLoading={isLoading}
        loadError={loadError}
        deleteError={deleteError}
        query={query}
        onQueryChange={(value) => {
          setQuery(value)
          setPage(1)
        }}
        vehicleTypeFilter={vehicleTypeFilter}
        onVehicleTypeFilterChange={(value) => {
          setVehicleTypeFilter(value)
          setPage(1)
        }}
        columnFilters={columnFilters}
        onColumnFilterChange={(key: ParkingTextFilterKey, value: string) => {
          setColumnFilters((prev) => ({ ...prev, [key]: value }))
          setPage(1)
        }}
        hasFilters={hasFilters}
        sortKey={sortKey}
        sortDirection={sortDirection}
        onSort={handleSort}
        parkings={parkings ?? []}
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
        deletingParkingId={deletingParkingId}
        onCreateParking={openCreateForm}
        onEditParking={openEditForm}
        onDeleteParking={setConfirmDeleteParking}
      />

      <ConfirmDialog
        open={confirmDeleteParking !== null}
        onOpenChange={(open) => !open && setConfirmDeleteParking(null)}
        title={t('confirmDeleteTitle')}
        description={t('parkingDeleteConfirm')}
        confirmLabel={t('parkingDelete')}
        cancelLabel={t('cancel')}
        loading={deletingParkingId === confirmDeleteParking?.id}
        error={deleteError}
        onConfirm={handleDeleteParking}
      />

      <ParkingFormSheet
        open={formOpen}
        onOpenChange={setFormOpen}
        isEdit={formParkingId !== null}
        tenantId={formTenantId}
        onTenantIdChange={setFormTenantId}
        tenants={tenants}
        tenantDisplayName={formTenantDisplayName}
        roomId={formRoomId}
        onRoomIdChange={setFormRoomId}
        rooms={rooms}
        vehicleType={formVehicleType}
        onVehicleTypeChange={setFormVehicleType}
        licensePlate={formLicensePlate}
        onLicensePlateChange={setFormLicensePlate}
        parkingSpot={formParkingSpot}
        onParkingSpotChange={setFormParkingSpot}
        saving={formSaving}
        error={formError}
        onSubmit={handleFormSubmit}
      />
    </main>
  )
}
