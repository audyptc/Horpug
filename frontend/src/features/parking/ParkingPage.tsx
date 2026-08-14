import { useEffect, useMemo, useState, type FormEvent } from 'react'
import { api, extractErrorMessage, type ApiPage } from '@/shared/api/client'
import { useLanguage } from '@/shared/i18n/language'
import { ConfirmDialog } from '@/shared/components/confirm-dialog'
import type { ApiTenant } from '@/features/tenant/types'
import type { ApiRoom } from '@/features/room/types'
import { ParkingListCard } from './components/ParkingListCard'
import { ParkingFormSheet } from './components/ParkingFormSheet'
import type { ApiParking, VehicleType } from './types'
import { PARKING_PAGE_SIZE_OPTIONS } from './utils'

export default function ParkingPage() {
  const { t } = useLanguage()

  const [parkings, setParkings] = useState<ApiParking[] | null>(null)
  const [tenants, setTenants] = useState<ApiTenant[]>([])
  const [rooms, setRooms] = useState<ApiRoom[]>([])
  const [loadError, setLoadError] = useState<string | null>(null)

  const [query, setQuery] = useState('')
  const [page, setPage] = useState(1)
  const [pageSize, setPageSize] = useState<number>(PARKING_PAGE_SIZE_OPTIONS[0])

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
    let cancelled = false

    Promise.all([
      api.get<ApiPage<ApiParking[]>>('/parking', { params: { per_page: 100 } }),
      api.get<ApiTenant[]>('/tenants/active', { params: { limit: 100 } }),
      api.get<ApiRoom[]>('/rooms/active', { params: { limit: 100 } }),
    ])
      .then(([parkingsRes, tenantsRes, roomsRes]) => {
        if (cancelled) return
        setParkings(parkingsRes.data.data)
        setTenants(tenantsRes.data)
        setRooms(roomsRes.data)
      })
      .catch((err) => {
        if (!cancelled) setLoadError(extractErrorMessage(err, t('resourceLoadError')))
      })

    return () => {
      cancelled = true
    }
    // eslint-disable-next-line react-hooks/exhaustive-deps
  }, [])

  const filteredParkings = useMemo(() => {
    const q = query.trim().toLocaleLowerCase()
    if (!q) return parkings ?? []

    return (parkings ?? []).filter((parking) => {
      return (
        (parking.tenant_name ?? '').toLocaleLowerCase().includes(q) ||
        (parking.room_number ?? '').toLocaleLowerCase().includes(q) ||
        (parking.dormitory_name ?? '').toLocaleLowerCase().includes(q) ||
        parking.license_plate.toLocaleLowerCase().includes(q) ||
        parking.parking_spot.toLocaleLowerCase().includes(q)
      )
    })
  }, [query, parkings])

  const totalPages = Math.max(1, Math.ceil(filteredParkings.length / pageSize))
  const currentPage = Math.min(page, totalPages)
  const rangeStart = filteredParkings.length === 0 ? 0 : (currentPage - 1) * pageSize + 1
  const rangeEnd = Math.min(currentPage * pageSize, filteredParkings.length)
  const paginatedParkings = filteredParkings.slice((currentPage - 1) * pageSize, currentPage * pageSize)

  const isLoading = !loadError && parkings === null

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
        const { data } = await api.post<ApiParking>('/parking', payload)
        setParkings((prev) => [data, ...(prev ?? [])])
      } else {
        const payload = {
          room_id: formRoomId || null,
          vehicle_type: formVehicleType,
          license_plate: formLicensePlate.trim(),
          parking_spot: formParkingSpot.trim(),
        }
        const { data } = await api.put<ApiParking>(`/parking/${formParkingId}`, payload)
        setParkings((prev) => prev?.map((item) => (item.id === data.id ? data : item)) ?? prev)
      }
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
      setParkings((prev) => prev?.filter((item) => item.id !== parking.id) ?? prev)
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
        parkings={parkings}
        query={query}
        onQueryChange={(value) => {
          setQuery(value)
          setPage(1)
        }}
        filteredParkings={filteredParkings}
        paginatedParkings={paginatedParkings}
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
