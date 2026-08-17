import { useEffect, useMemo, useState, type FormEvent } from 'react'
import { api, extractErrorMessage, type ApiPage } from '@/shared/api/client'
import { usePagination } from '@/shared/hooks/use-pagination'
import { useLanguage } from '@/shared/i18n/language'
import { ConfirmDialog } from '@/shared/components/confirm-dialog'
import type { ApiTenant } from '@/features/tenant/types'
import type { ApiRoom } from '@/features/room/types'
import { ParcelListCard } from './components/ParcelListCard'
import { ParcelFormSheet } from './components/ParcelFormSheet'
import type { ApiParcel, ParcelStatus } from './types'
import { PARCEL_PAGE_SIZE_OPTIONS, toApiDate, toDateInputValue } from './utils'

export default function ParcelPage() {
  const { t } = useLanguage()

  const [parcels, setParcels] = useState<ApiParcel[] | null>(null)
  const [tenants, setTenants] = useState<ApiTenant[]>([])
  const [rooms, setRooms] = useState<ApiRoom[]>([])
  const [loadError, setLoadError] = useState<string | null>(null)

  const [query, setQuery] = useState('')

  const [formOpen, setFormOpen] = useState(false)
  const [formParcelId, setFormParcelId] = useState<string | null>(null)
  const [formTenantId, setFormTenantId] = useState('')
  const [formTenantDisplayName, setFormTenantDisplayName] = useState('')
  const [formRoomId, setFormRoomId] = useState('')
  const [formCourier, setFormCourier] = useState('')
  const [formTrackingNumber, setFormTrackingNumber] = useState('')
  const [formStatus, setFormStatus] = useState<ParcelStatus>('pending')
  const [formReceivedDate, setFormReceivedDate] = useState('')
  const [formNote, setFormNote] = useState('')
  const [formSaving, setFormSaving] = useState(false)
  const [formError, setFormError] = useState<string | null>(null)

  const [deletingParcelId, setDeletingParcelId] = useState<string | null>(null)
  const [deleteError, setDeleteError] = useState<string | null>(null)
  const [confirmDeleteParcel, setConfirmDeleteParcel] = useState<ApiParcel | null>(null)

  useEffect(() => {
    let cancelled = false

    Promise.all([
      api.get<ApiPage<ApiParcel[]>>('/parcels', { params: { per_page: 100 } }),
      api.get<ApiTenant[]>('/tenants/active', { params: { limit: 100 } }),
      api.get<ApiRoom[]>('/rooms/active', { params: { limit: 100 } }),
    ])
      .then(([parcelsRes, tenantsRes, roomsRes]) => {
        if (cancelled) return
        setParcels(parcelsRes.data.data)
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

  const filteredParcels = useMemo(() => {
    const q = query.trim().toLocaleLowerCase()
    if (!q) return parcels ?? []

    return (parcels ?? []).filter((parcel) => {
      return (
        (parcel.tenant_name ?? '').toLocaleLowerCase().includes(q) ||
        (parcel.room_number ?? '').toLocaleLowerCase().includes(q) ||
        (parcel.dormitory_name ?? '').toLocaleLowerCase().includes(q) ||
        parcel.courier.toLocaleLowerCase().includes(q) ||
        parcel.tracking_number.toLocaleLowerCase().includes(q)
      )
    })
  }, [query, parcels])

  const {
    page: currentPage,
    pageSize,
    setPageSize,
    totalPages,
    rangeStart,
    rangeEnd,
    paginatedItems: paginatedParcels,
    resetPage,
    prevPage,
    nextPage,
  } = usePagination(filteredParcels, PARCEL_PAGE_SIZE_OPTIONS[0])

  const isLoading = !loadError && parcels === null

  function openCreateForm() {
    setFormParcelId(null)
    setFormTenantId('')
    setFormTenantDisplayName('')
    setFormRoomId('')
    setFormCourier('')
    setFormTrackingNumber('')
    setFormStatus('pending')
    setFormReceivedDate(toDateInputValue(new Date().toISOString()))
    setFormNote('')
    setFormError(null)
    setFormOpen(true)
  }

  function openEditForm(parcel: ApiParcel) {
    setFormParcelId(parcel.id)
    setFormTenantId(parcel.tenant_id)
    setFormTenantDisplayName(parcel.tenant_name ?? '')
    setFormRoomId(parcel.room_id ?? '')
    setFormCourier(parcel.courier)
    setFormTrackingNumber(parcel.tracking_number)
    setFormStatus(parcel.status)
    setFormReceivedDate(toDateInputValue(parcel.received_date))
    setFormNote(parcel.note)
    setFormError(null)
    setFormOpen(true)
  }

  async function handleFormSubmit(event: FormEvent<HTMLFormElement>) {
    event.preventDefault()

    const isEdit = formParcelId !== null

    if (!isEdit && !formTenantId) {
      setFormError(t('parcelTenantRequired'))
      return
    }
    if (!formReceivedDate) {
      setFormError(t('parcelReceivedDateRequired'))
      return
    }

    setFormSaving(true)
    setFormError(null)

    try {
      if (!isEdit) {
        const payload = {
          tenant_id: formTenantId,
          room_id: formRoomId || null,
          courier: formCourier.trim(),
          tracking_number: formTrackingNumber.trim(),
          status: formStatus,
          received_date: toApiDate(formReceivedDate),
          note: formNote.trim(),
        }
        const { data } = await api.post<ApiParcel>('/parcels', payload)
        setParcels((prev) => [data, ...(prev ?? [])])
      } else {
        const payload = {
          room_id: formRoomId || null,
          courier: formCourier.trim(),
          tracking_number: formTrackingNumber.trim(),
          status: formStatus,
          received_date: toApiDate(formReceivedDate),
          note: formNote.trim(),
        }
        const { data } = await api.put<ApiParcel>(`/parcels/${formParcelId}`, payload)
        setParcels((prev) => prev?.map((item) => (item.id === data.id ? data : item)) ?? prev)
      }
      setFormOpen(false)
    } catch (err) {
      const fallback = isEdit ? t('parcelUpdateError') : t('parcelCreateError')
      setFormError(extractErrorMessage(err, fallback))
    } finally {
      setFormSaving(false)
    }
  }

  async function handleDeleteParcel() {
    if (!confirmDeleteParcel) return
    const parcel = confirmDeleteParcel

    setDeletingParcelId(parcel.id)
    setDeleteError(null)

    try {
      await api.delete(`/parcels/${parcel.id}`)
      setParcels((prev) => prev?.filter((item) => item.id !== parcel.id) ?? prev)
      setConfirmDeleteParcel(null)
    } catch (err) {
      setDeleteError(extractErrorMessage(err, t('parcelDeleteError')))
    } finally {
      setDeletingParcelId(null)
    }
  }

  return (
    <main className="content">
      <section className="welcome">
        <h1>{t('menuParcels')}</h1>
        <p>{t('menuParcelsDescription')}</p>
      </section>

      <ParcelListCard
        isLoading={isLoading}
        loadError={loadError}
        deleteError={deleteError}
        parcels={parcels}
        query={query}
        onQueryChange={(value) => {
          setQuery(value)
          resetPage()
        }}
        filteredParcels={filteredParcels}
        paginatedParcels={paginatedParcels}
        currentPage={currentPage}
        totalPages={totalPages}
        rangeStart={rangeStart}
        rangeEnd={rangeEnd}
        pageSize={pageSize}
        onPageSizeChange={setPageSize}
        onPrevPage={prevPage}
        onNextPage={nextPage}
        deletingParcelId={deletingParcelId}
        onCreateParcel={openCreateForm}
        onEditParcel={openEditForm}
        onDeleteParcel={setConfirmDeleteParcel}
      />

      <ConfirmDialog
        open={confirmDeleteParcel !== null}
        onOpenChange={(open) => !open && setConfirmDeleteParcel(null)}
        title={t('confirmDeleteTitle')}
        description={t('parcelDeleteConfirm')}
        confirmLabel={t('parcelDelete')}
        cancelLabel={t('cancel')}
        loading={deletingParcelId === confirmDeleteParcel?.id}
        error={deleteError}
        onConfirm={handleDeleteParcel}
      />

      <ParcelFormSheet
        open={formOpen}
        onOpenChange={setFormOpen}
        isEdit={formParcelId !== null}
        tenantId={formTenantId}
        onTenantIdChange={setFormTenantId}
        tenants={tenants}
        tenantDisplayName={formTenantDisplayName}
        roomId={formRoomId}
        onRoomIdChange={setFormRoomId}
        rooms={rooms}
        courier={formCourier}
        onCourierChange={setFormCourier}
        trackingNumber={formTrackingNumber}
        onTrackingNumberChange={setFormTrackingNumber}
        status={formStatus}
        onStatusChange={setFormStatus}
        receivedDate={formReceivedDate}
        onReceivedDateChange={setFormReceivedDate}
        note={formNote}
        onNoteChange={setFormNote}
        saving={formSaving}
        error={formError}
        onSubmit={handleFormSubmit}
      />
    </main>
  )
}
