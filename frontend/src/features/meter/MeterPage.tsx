import { useEffect, useState, type FormEvent } from 'react'
import axios from 'axios'
import { api, extractErrorMessage, type ApiPage } from '@/shared/api/client'
import { useLanguage } from '@/shared/i18n/language'
import { ConfirmDialog } from '@/shared/components/confirm-dialog'
import { InformationDialog } from '@/shared/components/information-dialog'
import type { ApiRoom } from '@/features/room/types'
import { MeterListCard } from './components/MeterListCard'
import { MeterFormSheet } from './components/MeterFormSheet'
import type { ApiMeter, BillingMethod } from './types'
import {
  METER_PAGE_SIZE_OPTIONS,
  toApiDate,
  toDateInputValue,
  type MeterBilledFilter,
  type MeterBillingMethodFilter,
  type MeterColumnFilters,
  type MeterSortDirection,
  type MeterSortKey,
  type MeterTextFilterKey,
} from './utils'

const SEARCH_DEBOUNCE_MS = 300

export default function MeterPage() {
  const { t } = useLanguage()

  const [meters, setMeters] = useState<ApiMeter[] | null>(null)
  const [rooms, setRooms] = useState<ApiRoom[]>([])
  const [total, setTotal] = useState(0)
  const [totalPages, setTotalPages] = useState(1)
  const [loadError, setLoadError] = useState<string | null>(null)

  const [query, setQuery] = useState('')
  const [debouncedQuery, setDebouncedQuery] = useState('')
  const [billingMethodFilter, setBillingMethodFilter] = useState<MeterBillingMethodFilter>('all')
  const [billedFilter, setBilledFilter] = useState<MeterBilledFilter>('all')
  // Applied on submit from each column's menu, so no debounce is needed here.
  const [columnFilters, setColumnFilters] = useState<MeterColumnFilters>({})
  // The list reads as a running log, so it opens on the newest readings.
  const [sortKey, setSortKey] = useState<MeterSortKey>('reading_date')
  const [sortDirection, setSortDirection] = useState<MeterSortDirection>('desc')

  const [page, setPage] = useState(1)
  const [pageSize, setPageSizeState] = useState<number>(METER_PAGE_SIZE_OPTIONS[0])
  // Bumped by mutations so the list refetches; the server owns the ordering
  // and page boundaries now, so patching rows locally would misplace them.
  const [refreshToken, setRefreshToken] = useState(0)

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
  const [confirmDeleteMeter, setConfirmDeleteMeter] = useState<ApiMeter | null>(null)
  const [deleteBlocked, setDeleteBlocked] = useState(false)

  useEffect(() => {
    const timer = window.setTimeout(() => setDebouncedQuery(query), SEARCH_DEBOUNCE_MS)
    return () => window.clearTimeout(timer)
  }, [query])

  // The room selector in the form isn't affected by the list's filters, so
  // it's loaded once rather than on every refetch.
  useEffect(() => {
    let cancelled = false

    api
      .get<ApiRoom[]>('/rooms/active', { params: { limit: 100 } })
      .then(({ data }) => {
        if (!cancelled) setRooms(data)
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
      .get<ApiPage<ApiMeter[]>>('/meters', {
        signal: controller.signal,
        params: {
          page,
          per_page: pageSize,
          q: debouncedQuery.trim() || undefined,
          billing_method: billingMethodFilter === 'all' ? undefined : billingMethodFilter,
          is_billed: billedFilter === 'all' ? undefined : billedFilter === 'billed',
          sort: sortKey,
          order: sortDirection,
          ...columnParams,
        },
      })
      .then(({ data }) => {
        setMeters(data.data)
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
  }, [
    page,
    pageSize,
    debouncedQuery,
    billingMethodFilter,
    billedFilter,
    columnFilters,
    sortKey,
    sortDirection,
    refreshToken,
  ])

  const isLoading = !loadError && meters === null
  const hasFilters =
    query !== '' ||
    billingMethodFilter !== 'all' ||
    billedFilter !== 'all' ||
    Object.values(columnFilters).some(Boolean)
  const rangeStart = total === 0 ? 0 : (page - 1) * pageSize + 1
  const rangeEnd = Math.min(page * pageSize, total)

  function refresh() {
    setRefreshToken((value) => value + 1)
  }

  function setPageSize(size: number) {
    setPageSizeState(size)
    setPage(1)
  }

  function handleSort(key: MeterSortKey) {
    if (key === sortKey) {
      setSortDirection(sortDirection === 'asc' ? 'desc' : 'asc')
    } else {
      setSortKey(key)
      setSortDirection('asc')
    }
    setPage(1)
  }

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

  function openEditForm(meter: ApiMeter) {
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
      setFormError(t('meterRoomRequired'))
      return
    }
    if (!formReadingDate) {
      setFormError(t('meterReadingDateRequired'))
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
        setFormError(t('meterUnitsInvalid'))
        return
      }
      if (!Number.isFinite(pricePerUnit) || pricePerUnit < 0) {
        setFormError(t('meterPriceInvalid'))
        return
      }
    }

    const flatAmount = Number(formFlatAmount)
    if (formBillingMethod === 'flat' && (formFlatAmount.trim() === '' || !Number.isFinite(flatAmount) || flatAmount < 0)) {
      setFormError(t('meterFlatAmountRequired'))
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
        await api.post<ApiMeter>('/meters', payload)
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
        await api.put<ApiMeter>(`/meters/${formMeterId}`, payload)
      }
      refresh()
      setFormOpen(false)
    } catch (err) {
      const fallback = isEdit ? t('meterUpdateError') : t('meterCreateError')
      setFormError(extractErrorMessage(err, fallback))
    } finally {
      setFormSaving(false)
    }
  }

  function handleRequestDeleteMeter(meter: ApiMeter) {
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
      await api.delete(`/meters/${meter.id}`)
      // Deleting the last row of the final page would otherwise strand the
      // view on a page the server no longer has.
      setPage((value) => (meters?.length === 1 ? Math.max(1, value - 1) : value))
      refresh()
      setConfirmDeleteMeter(null)
    } catch (err) {
      setDeleteError(extractErrorMessage(err, t('meterDeleteError')))
    } finally {
      setDeletingMeterId(null)
    }
  }

  return (
    <main className="content">
      <section className="welcome">
        <h1>{t('menuMeters')}</h1>
        <p>{t('menuMetersDescription')}</p>
      </section>

      <MeterListCard
        isLoading={isLoading}
        loadError={loadError}
        deleteError={deleteError}
        query={query}
        onQueryChange={(value) => {
          setQuery(value)
          setPage(1)
        }}
        billingMethodFilter={billingMethodFilter}
        onBillingMethodFilterChange={(value) => {
          setBillingMethodFilter(value)
          setPage(1)
        }}
        billedFilter={billedFilter}
        onBilledFilterChange={(value) => {
          setBilledFilter(value)
          setPage(1)
        }}
        columnFilters={columnFilters}
        onColumnFilterChange={(key: MeterTextFilterKey, value: string) => {
          setColumnFilters((prev) => ({ ...prev, [key]: value }))
          setPage(1)
        }}
        hasFilters={hasFilters}
        sortKey={sortKey}
        sortDirection={sortDirection}
        onSort={handleSort}
        meters={meters ?? []}
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
        deletingMeterId={deletingMeterId}
        onCreateMeter={openCreateForm}
        onEditMeter={openEditForm}
        onDeleteMeter={handleRequestDeleteMeter}
      />

      <ConfirmDialog
        open={confirmDeleteMeter !== null}
        onOpenChange={(open) => !open && setConfirmDeleteMeter(null)}
        title={t('confirmDeleteTitle')}
        description={t('meterDeleteConfirm')}
        confirmLabel={t('meterDelete')}
        cancelLabel={t('cancel')}
        loading={deletingMeterId === confirmDeleteMeter?.id}
        error={deleteError}
        onConfirm={handleDeleteMeter}
      />

      <InformationDialog
        open={deleteBlocked}
        onOpenChange={(open) => !open && setDeleteBlocked(false)}
        title={t('meterDeleteBlockedTitle')}
        description={t('meterDeleteBlockedDescription')}
        actionLabel={t('acknowledge')}
      />

      <MeterFormSheet
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
