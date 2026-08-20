import { useEffect, useMemo, useState, type FormEvent } from 'react'
import { api, extractErrorMessage, type ApiPage } from '@/shared/api/client'
import { usePagination } from '@/shared/hooks/use-pagination'
import { useLanguage } from '@/shared/i18n/language'
import { ConfirmDialog } from '@/shared/components/confirm-dialog'
import { InformationDialog } from '@/shared/components/information-dialog'
import type { ApiRoom } from '@/features/room/types'
import { WaterMeterListCard } from './components/WaterMeterListCard'
import { WaterMeterFormSheet } from './components/WaterMeterFormSheet'
import type { ApiWaterMeter, BillingMethod } from './types'
import { WATER_METER_PAGE_SIZE_OPTIONS, toApiDate, toDateInputValue } from './utils'

export default function WaterMeterPage() {
  const { t } = useLanguage()

  const [meters, setMeters] = useState<ApiWaterMeter[] | null>(null)
  const [rooms, setRooms] = useState<ApiRoom[]>([])
  const [loadError, setLoadError] = useState<string | null>(null)

  const [query, setQuery] = useState('')

  const [formOpen, setFormOpen] = useState(false)
  const [formMeterId, setFormMeterId] = useState<string | null>(null)
  const [formRoomId, setFormRoomId] = useState('')
  const [formRoomDisplayLabel, setFormRoomDisplayLabel] = useState('')
  const [formBillingMethod, setFormBillingMethod] = useState<BillingMethod>('metered')
  const [formReadingDate, setFormReadingDate] = useState('')
  const [formPreviousUnit, setFormPreviousUnit] = useState('0')
  const [formCurrentUnit, setFormCurrentUnit] = useState('0')
  const [formPricePerUnit, setFormPricePerUnit] = useState('0')
  const [formFlatAmount, setFormFlatAmount] = useState('')
  const [formNote, setFormNote] = useState('')
  const [formSaving, setFormSaving] = useState(false)
  const [formError, setFormError] = useState<string | null>(null)

  const [deletingMeterId, setDeletingMeterId] = useState<string | null>(null)
  const [deleteError, setDeleteError] = useState<string | null>(null)
  const [confirmDeleteMeter, setConfirmDeleteMeter] = useState<ApiWaterMeter | null>(null)
  const [deleteBlocked, setDeleteBlocked] = useState(false)

  useEffect(() => {
    let cancelled = false

    Promise.all([
      api.get<ApiPage<ApiWaterMeter[]>>('/water-meters', { params: { per_page: 100 } }),
      api.get<ApiRoom[]>('/rooms/active', { params: { limit: 100 } }),
    ])
      .then(([metersRes, roomsRes]) => {
        if (cancelled) return
        setMeters(metersRes.data.data)
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

  const filteredMeters = useMemo(() => {
    const q = query.trim().toLocaleLowerCase()
    if (!q) return meters ?? []

    return (meters ?? []).filter((meter) => {
      return (
        (meter.room_number ?? '').toLocaleLowerCase().includes(q) ||
        (meter.dormitory_name ?? '').toLocaleLowerCase().includes(q)
      )
    })
  }, [query, meters])

  const {
    page: currentPage,
    pageSize,
    setPageSize,
    totalPages,
    rangeStart,
    rangeEnd,
    paginatedItems: paginatedMeters,
    resetPage,
    firstPage,
    prevPage,
    nextPage,
    lastPage,
  } = usePagination(filteredMeters, WATER_METER_PAGE_SIZE_OPTIONS[0])

  const isLoading = !loadError && meters === null

  function openCreateForm() {
    setFormMeterId(null)
    setFormRoomId('')
    setFormRoomDisplayLabel('')
    setFormBillingMethod('metered')
    setFormReadingDate('')
    setFormPreviousUnit('0')
    setFormCurrentUnit('0')
    setFormPricePerUnit('0')
    setFormFlatAmount('')
    setFormNote('')
    setFormError(null)
    setFormOpen(true)
  }

  function openEditForm(meter: ApiWaterMeter) {
    setFormMeterId(meter.id)
    setFormRoomId(meter.room_id)
    setFormRoomDisplayLabel([meter.room_number, meter.dormitory_name].filter(Boolean).join(' - '))
    setFormBillingMethod(meter.billing_method)
    setFormReadingDate(toDateInputValue(meter.reading_date))
    setFormPreviousUnit(String(meter.previous_unit))
    setFormCurrentUnit(String(meter.current_unit))
    setFormPricePerUnit(String(meter.price_per_unit))
    setFormFlatAmount(meter.flat_amount !== undefined ? String(meter.flat_amount) : '')
    setFormNote(meter.note)
    setFormError(null)
    setFormOpen(true)
  }

  async function handleFormSubmit(event: FormEvent<HTMLFormElement>) {
    event.preventDefault()

    const isEdit = formMeterId !== null

    if (!isEdit && !formRoomId) {
      setFormError(t('waterMeterRoomRequired'))
      return
    }
    if (!formReadingDate) {
      setFormError(t('waterMeterReadingDateRequired'))
      return
    }

    const previousUnit = Number(formPreviousUnit)
    const currentUnit = Number(formCurrentUnit)
    const pricePerUnit = Number(formPricePerUnit)

    if (formBillingMethod === 'metered') {
      if (
        !Number.isFinite(previousUnit) ||
        !Number.isFinite(currentUnit) ||
        previousUnit < 0 ||
        currentUnit < 0 ||
        currentUnit < previousUnit
      ) {
        setFormError(t('waterMeterUnitsInvalid'))
        return
      }
      if (!Number.isFinite(pricePerUnit) || pricePerUnit < 0) {
        setFormError(t('waterMeterPriceInvalid'))
        return
      }
    }

    const flatAmount = Number(formFlatAmount)
    if (formBillingMethod === 'flat' && (formFlatAmount.trim() === '' || !Number.isFinite(flatAmount) || flatAmount < 0)) {
      setFormError(t('waterMeterFlatAmountRequired'))
      return
    }

    setFormSaving(true)
    setFormError(null)

    try {
      if (!isEdit) {
        const payload = {
          room_id: formRoomId,
          billing_method: formBillingMethod,
          reading_date: toApiDate(formReadingDate),
          previous_unit: previousUnit,
          current_unit: currentUnit,
          price_per_unit: pricePerUnit,
          flat_amount: formBillingMethod === 'flat' ? flatAmount : null,
          note: formNote.trim(),
        }
        const { data } = await api.post<ApiWaterMeter>('/water-meters', payload)
        setMeters((prev) => [...(prev ?? []), data])
      } else {
        const payload = {
          billing_method: formBillingMethod,
          reading_date: toApiDate(formReadingDate),
          previous_unit: previousUnit,
          current_unit: currentUnit,
          price_per_unit: pricePerUnit,
          flat_amount: formBillingMethod === 'flat' ? flatAmount : null,
          note: formNote.trim(),
        }
        const { data } = await api.put<ApiWaterMeter>(`/water-meters/${formMeterId}`, payload)
        setMeters((prev) => prev?.map((item) => (item.id === data.id ? data : item)) ?? prev)
      }
      setFormOpen(false)
    } catch (err) {
      const fallback = isEdit ? t('waterMeterUpdateError') : t('waterMeterCreateError')
      setFormError(extractErrorMessage(err, fallback))
    } finally {
      setFormSaving(false)
    }
  }

  function handleRequestDeleteMeter(meter: ApiWaterMeter) {
    if (meter.is_billed) {
      setDeleteBlocked(true)
      return
    }
    setConfirmDeleteMeter(meter)
  }

  async function handleDeleteMeter() {
    if (!confirmDeleteMeter) return
    const meter = confirmDeleteMeter

    setDeletingMeterId(meter.id)
    setDeleteError(null)

    try {
      await api.delete(`/water-meters/${meter.id}`)
      setMeters((prev) => prev?.filter((item) => item.id !== meter.id) ?? prev)
      setConfirmDeleteMeter(null)
    } catch (err) {
      setDeleteError(extractErrorMessage(err, t('waterMeterDeleteError')))
    } finally {
      setDeletingMeterId(null)
    }
  }

  return (
    <main className="content">
      <section className="welcome">
        <h1>{t('menuWaterMeters')}</h1>
        <p>{t('menuWaterMetersDescription')}</p>
      </section>

      <WaterMeterListCard
        isLoading={isLoading}
        loadError={loadError}
        deleteError={deleteError}
        meters={meters}
        query={query}
        onQueryChange={(value) => {
          setQuery(value)
          resetPage()
        }}
        filteredMeters={filteredMeters}
        paginatedMeters={paginatedMeters}
        currentPage={currentPage}
        totalPages={totalPages}
        rangeStart={rangeStart}
        rangeEnd={rangeEnd}
        pageSize={pageSize}
        onPageSizeChange={setPageSize}
        onFirstPage={firstPage}
        onPrevPage={prevPage}
        onNextPage={nextPage}
        onLastPage={lastPage}
        deletingMeterId={deletingMeterId}
        onCreateMeter={openCreateForm}
        onEditMeter={openEditForm}
        onDeleteMeter={handleRequestDeleteMeter}
      />

      <ConfirmDialog
        open={confirmDeleteMeter !== null}
        onOpenChange={(open) => !open && setConfirmDeleteMeter(null)}
        title={t('confirmDeleteTitle')}
        description={t('waterMeterDeleteConfirm')}
        confirmLabel={t('waterMeterDelete')}
        cancelLabel={t('cancel')}
        loading={deletingMeterId === confirmDeleteMeter?.id}
        error={deleteError}
        onConfirm={handleDeleteMeter}
      />

      <InformationDialog
        open={deleteBlocked}
        onOpenChange={(open) => !open && setDeleteBlocked(false)}
        title={t('waterMeterDeleteBlockedTitle')}
        description={t('waterMeterDeleteBlockedDescription')}
        actionLabel={t('acknowledge')}
      />

      <WaterMeterFormSheet
        open={formOpen}
        onOpenChange={setFormOpen}
        isEdit={formMeterId !== null}
        roomId={formRoomId}
        onRoomIdChange={setFormRoomId}
        rooms={rooms}
        roomDisplayLabel={formRoomDisplayLabel}
        billingMethod={formBillingMethod}
        onBillingMethodChange={setFormBillingMethod}
        readingDate={formReadingDate}
        onReadingDateChange={setFormReadingDate}
        previousUnit={formPreviousUnit}
        onPreviousUnitChange={setFormPreviousUnit}
        currentUnit={formCurrentUnit}
        onCurrentUnitChange={setFormCurrentUnit}
        pricePerUnit={formPricePerUnit}
        onPricePerUnitChange={setFormPricePerUnit}
        flatAmount={formFlatAmount}
        onFlatAmountChange={setFormFlatAmount}
        note={formNote}
        onNoteChange={setFormNote}
        saving={formSaving}
        error={formError}
        onSubmit={handleFormSubmit}
      />
    </main>
  )
}
