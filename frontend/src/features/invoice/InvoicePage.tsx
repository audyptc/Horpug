import { useEffect, useMemo, useState, type FormEvent } from 'react'
import { api, extractErrorMessage, type ApiPage } from '@/shared/api/client'
import { usePagination } from '@/shared/hooks/use-pagination'
import { useLanguage } from '@/shared/i18n/language'
import { ConfirmDialog } from '@/shared/components/confirm-dialog'
import type { ApiContract } from '@/features/contract/types'
import type { ApiMeter } from '@/features/meter/types'
import type { ApiWaterMeter } from '@/features/watermeter/types'
import { InvoiceListCard } from './components/InvoiceListCard'
import { InvoiceFormSheet } from './components/InvoiceFormSheet'
import type { ApiInvoice, InvoiceStatus } from './types'
import {
  INVOICE_PAGE_SIZE_OPTIONS,
  parsePeriodInputValue,
  sumMeterAmountForPeriod,
  toApiDate,
  toDateInputValue,
  toPeriodInputValue,
} from './utils'

export default function InvoicePage() {
  const { t } = useLanguage()

  const [invoices, setInvoices] = useState<ApiInvoice[] | null>(null)
  const [contracts, setContracts] = useState<ApiContract[]>([])
  const [loadError, setLoadError] = useState<string | null>(null)

  const [query, setQuery] = useState('')

  const [formOpen, setFormOpen] = useState(false)
  const [formInvoiceId, setFormInvoiceId] = useState<string | null>(null)
  const [formInvoiceDetail, setFormInvoiceDetail] = useState<ApiInvoice | null>(null)
  const [formDetailLoading, setFormDetailLoading] = useState(false)
  const [formContractId, setFormContractId] = useState('')
  const [formPeriod, setFormPeriod] = useState('')
  const [formIssueDate, setFormIssueDate] = useState('')
  const [formDueDate, setFormDueDate] = useState('')
  const [formStatus, setFormStatus] = useState<InvoiceStatus>('unpaid')
  const [formNote, setFormNote] = useState('')
  const [formSaving, setFormSaving] = useState(false)
  const [formError, setFormError] = useState<string | null>(null)
  const [formPendingItems, setFormPendingItems] = useState<{ description: string; amount: number }[]>([])
  // Keyed by "roomId|YYYY-MM" so switching back and forth between contracts/periods
  // within one sheet session doesn't refetch. Reset whenever the create sheet reopens.
  const [utilityPreviewCache, setUtilityPreviewCache] = useState<
    Record<string, { electricity: number | null; water: number | null }>
  >({})

  const [deletingInvoiceId, setDeletingInvoiceId] = useState<string | null>(null)
  const [deleteError, setDeleteError] = useState<string | null>(null)
  const [confirmDeleteInvoice, setConfirmDeleteInvoice] = useState<ApiInvoice | null>(null)

  const [itemSaving, setItemSaving] = useState(false)
  const [itemError, setItemError] = useState<string | null>(null)
  const [removingItemId, setRemovingItemId] = useState<string | null>(null)

  const [sendingLineInvoiceId, setSendingLineInvoiceId] = useState<string | null>(null)

  useEffect(() => {
    let cancelled = false

    Promise.all([
      api.get<ApiPage<ApiInvoice[]>>('/invoices', { params: { per_page: 100 } }),
      api.get<ApiPage<ApiContract[]>>('/contracts', { params: { status: 'active', per_page: 100 } }),
    ])
      .then(([invoicesRes, contractsRes]) => {
        if (cancelled) return
        setInvoices(invoicesRes.data.data)
        setContracts(contractsRes.data.data)
      })
      .catch((err) => {
        if (!cancelled) setLoadError(extractErrorMessage(err, t('resourceLoadError')))
      })

    return () => {
      cancelled = true
    }
    // eslint-disable-next-line react-hooks/exhaustive-deps
  }, [])

  const filteredInvoices = useMemo(() => {
    const q = query.trim().toLocaleLowerCase()
    if (!q) return invoices ?? []

    return (invoices ?? []).filter((invoice) => {
      return (
        (invoice.tenant_name ?? '').toLocaleLowerCase().includes(q) ||
        (invoice.room_number ?? '').toLocaleLowerCase().includes(q) ||
        (invoice.dormitory_name ?? '').toLocaleLowerCase().includes(q)
      )
    })
  }, [query, invoices])

  const {
    page: currentPage,
    pageSize,
    setPageSize,
    totalPages,
    rangeStart,
    rangeEnd,
    paginatedItems: paginatedInvoices,
    resetPage,
    firstPage,
    prevPage,
    nextPage,
    lastPage,
  } = usePagination(filteredInvoices, INVOICE_PAGE_SIZE_OPTIONS[0])

  const isLoading = !loadError && invoices === null

  const selectedFormContract = contracts.find((item) => item.id === formContractId)
  const selectedFormPeriod = parsePeriodInputValue(formPeriod)
  const utilityPreviewKey =
    selectedFormContract && selectedFormPeriod
      ? `${selectedFormContract.room_id}|${toPeriodInputValue(selectedFormPeriod.year, selectedFormPeriod.month)}`
      : null
  const utilityPreview = utilityPreviewKey ? utilityPreviewCache[utilityPreviewKey] : undefined
  const formUtilityLoading = utilityPreviewKey !== null && utilityPreview === undefined

  // Preview the electricity/water charges the backend will auto-pull onto the
  // invoice (see Repository.Create), so the create form shows the same
  // rent + electricity + water breakdown before the user submits.
  useEffect(() => {
    if (formInvoiceId !== null || !formOpen || !utilityPreviewKey || utilityPreviewKey in utilityPreviewCache) return

    const contract = selectedFormContract
    const period = selectedFormPeriod
    if (!contract || !period) return

    let cancelled = false

    Promise.all([
      api.get<ApiPage<ApiMeter[]>>('/meters', { params: { room_id: contract.room_id, per_page: 100 } }),
      api.get<ApiPage<ApiWaterMeter[]>>('/water-meters', { params: { room_id: contract.room_id, per_page: 100 } }),
    ])
      .then(([electricityRes, waterRes]) => {
        if (cancelled) return
        setUtilityPreviewCache((prev) => ({
          ...prev,
          [utilityPreviewKey]: {
            electricity: sumMeterAmountForPeriod(electricityRes.data.data, period.year, period.month),
            water: sumMeterAmountForPeriod(waterRes.data.data, period.year, period.month),
          },
        }))
      })
      .catch(() => {
        if (!cancelled) {
          setUtilityPreviewCache((prev) => ({ ...prev, [utilityPreviewKey]: { electricity: null, water: null } }))
        }
      })

    return () => {
      cancelled = true
    }
    // eslint-disable-next-line react-hooks/exhaustive-deps
  }, [utilityPreviewKey, formInvoiceId, formOpen])

  function openCreateForm() {
    setFormInvoiceId(null)
    setFormInvoiceDetail(null)
    setFormDetailLoading(false)
    setFormContractId('')
    setFormPeriod('')
    setFormIssueDate('')
    setFormDueDate('')
    setFormStatus('unpaid')
    setFormNote('')
    setFormError(null)
    setItemError(null)
    setFormPendingItems([])
    setUtilityPreviewCache({})
    setFormOpen(true)
  }

  function openEditForm(invoice: ApiInvoice) {
    setFormInvoiceId(invoice.id)
    setFormInvoiceDetail(null)
    setFormDueDate(toDateInputValue(invoice.due_date))
    setFormStatus(invoice.status)
    setFormNote(invoice.note)
    setFormError(null)
    setItemError(null)
    setFormOpen(true)

    setFormDetailLoading(true)
    api
      .get<ApiInvoice>(`/invoices/${invoice.id}`)
      .then(({ data }) => {
        setFormInvoiceDetail(data)
        setFormDueDate(toDateInputValue(data.due_date))
        setFormStatus(data.status)
        setFormNote(data.note)
      })
      .catch((err) => {
        setFormError(extractErrorMessage(err, t('resourceLoadError')))
      })
      .finally(() => {
        setFormDetailLoading(false)
      })
  }

  async function handleFormSubmit(event: FormEvent<HTMLFormElement>) {
    event.preventDefault()

    const isEdit = formInvoiceId !== null

    if (!isEdit) {
      if (!formContractId) {
        setFormError(t('invoiceContractRequired'))
        return
      }
      const period = parsePeriodInputValue(formPeriod)
      if (!period) {
        setFormError(t('invoicePeriodRequired'))
        return
      }
      if (!formIssueDate) {
        setFormError(t('invoiceIssueDateRequired'))
        return
      }
      if (!formDueDate) {
        setFormError(t('invoiceDueDateRequired'))
        return
      }

      setFormSaving(true)
      setFormError(null)

      try {
        const payload = {
          contract_id: formContractId,
          period_year: period.year,
          period_month: period.month,
          issue_date: toApiDate(formIssueDate),
          due_date: toApiDate(formDueDate),
          note: formNote.trim(),
        }
        const { data: created } = await api.post<ApiInvoice>('/invoices', payload)
        setInvoices((prev) => [...(prev ?? []), created])

        let finalInvoice = created
        for (const item of formPendingItems) {
          try {
            const { data: updated } = await api.post<ApiInvoice>(`/invoices/${created.id}/items`, item)
            finalInvoice = updated
            setInvoices((prev) => prev?.map((inv) => (inv.id === updated.id ? updated : inv)) ?? prev)
          } catch (err) {
            openEditForm(finalInvoice)
            setFormError(extractErrorMessage(err, t('invoiceItemAddError')))
            return
          }
        }

        setFormOpen(false)
      } catch (err) {
        setFormError(extractErrorMessage(err, t('invoiceCreateError')))
      } finally {
        setFormSaving(false)
      }
      return
    }

    if (!formDueDate) {
      setFormError(t('invoiceDueDateRequired'))
      return
    }

    setFormSaving(true)
    setFormError(null)

    try {
      const payload = {
        status: formStatus,
        due_date: toApiDate(formDueDate),
        note: formNote.trim(),
      }
      const { data } = await api.put<ApiInvoice>(`/invoices/${formInvoiceId}`, payload)
      setInvoices((prev) => prev?.map((item) => (item.id === data.id ? data : item)) ?? prev)
      setFormOpen(false)
    } catch (err) {
      setFormError(extractErrorMessage(err, t('invoiceUpdateError')))
    } finally {
      setFormSaving(false)
    }
  }

  async function handleDeleteInvoice() {
    if (!confirmDeleteInvoice) return
    const invoice = confirmDeleteInvoice

    setDeletingInvoiceId(invoice.id)
    setDeleteError(null)

    try {
      await api.delete(`/invoices/${invoice.id}`)
      setInvoices((prev) => prev?.filter((item) => item.id !== invoice.id) ?? prev)
      setConfirmDeleteInvoice(null)
    } catch (err) {
      setDeleteError(extractErrorMessage(err, t('invoiceDeleteError')))
    } finally {
      setDeletingInvoiceId(null)
    }
  }

  async function handleSendLineInvoice(invoice: ApiInvoice) {
    setSendingLineInvoiceId(invoice.id)

    try {
      await api.post(`/invoices/${invoice.id}/send-line`)
      window.alert(t('invoiceSendLineSuccess'))
    } catch (err) {
      window.alert(extractErrorMessage(err, t('invoiceSendLineError')))
    } finally {
      setSendingLineInvoiceId(null)
    }
  }

  function applyInvoiceUpdate(data: ApiInvoice) {
    setFormInvoiceDetail(data)
    setInvoices((prev) => prev?.map((item) => (item.id === data.id ? data : item)) ?? prev)
  }

  async function handleAddItem(description: string, amount: number) {
    if (!formInvoiceId) return

    setItemSaving(true)
    setItemError(null)

    try {
      const { data } = await api.post<ApiInvoice>(`/invoices/${formInvoiceId}/items`, { description, amount })
      applyInvoiceUpdate(data)
    } catch (err) {
      setItemError(extractErrorMessage(err, t('invoiceItemAddError')))
    } finally {
      setItemSaving(false)
    }
  }

  function addPendingItem(description: string, amount: number) {
    setFormPendingItems((prev) => [...prev, { description, amount }])
  }

  function removePendingItem(index: number) {
    setFormPendingItems((prev) => prev.filter((_, i) => i !== index))
  }

  async function handleRemoveItem(itemId: string) {
    if (!formInvoiceId) return

    setRemovingItemId(itemId)
    setItemError(null)

    try {
      const { data } = await api.delete<ApiInvoice>(`/invoices/${formInvoiceId}/items/${itemId}`)
      applyInvoiceUpdate(data)
    } catch (err) {
      setItemError(extractErrorMessage(err, t('invoiceItemRemoveError')))
    } finally {
      setRemovingItemId(null)
    }
  }

  return (
    <main className="content">
      <section className="welcome">
        <h1>{t('menuInvoices')}</h1>
        <p>{t('menuInvoicesDescription')}</p>
      </section>

      <InvoiceListCard
        isLoading={isLoading}
        loadError={loadError}
        deleteError={deleteError}
        invoices={invoices}
        query={query}
        onQueryChange={(value) => {
          setQuery(value)
          resetPage()
        }}
        filteredInvoices={filteredInvoices}
        paginatedInvoices={paginatedInvoices}
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
        deletingInvoiceId={deletingInvoiceId}
        onCreateInvoice={openCreateForm}
        onEditInvoice={openEditForm}
        onDeleteInvoice={setConfirmDeleteInvoice}
        sendingLineInvoiceId={sendingLineInvoiceId}
        onSendLineInvoice={handleSendLineInvoice}
      />

      <ConfirmDialog
        open={confirmDeleteInvoice !== null}
        onOpenChange={(open) => !open && setConfirmDeleteInvoice(null)}
        title={t('confirmDeleteTitle')}
        description={t('invoiceDeleteConfirm')}
        confirmLabel={t('invoiceDelete')}
        cancelLabel={t('cancel')}
        loading={deletingInvoiceId === confirmDeleteInvoice?.id}
        error={deleteError}
        onConfirm={handleDeleteInvoice}
      />

      <InvoiceFormSheet
        open={formOpen}
        onOpenChange={setFormOpen}
        isEdit={formInvoiceId !== null}
        contracts={contracts}
        contractId={formContractId}
        onContractIdChange={setFormContractId}
        period={formPeriod}
        onPeriodChange={setFormPeriod}
        electricityAmount={utilityPreview?.electricity ?? null}
        waterAmount={utilityPreview?.water ?? null}
        utilityLoading={formUtilityLoading}
        issueDate={formIssueDate}
        onIssueDateChange={setFormIssueDate}
        dueDate={formDueDate}
        onDueDateChange={setFormDueDate}
        status={formStatus}
        onStatusChange={setFormStatus}
        note={formNote}
        onNoteChange={setFormNote}
        invoice={formInvoiceDetail}
        detailLoading={formDetailLoading}
        saving={formSaving}
        error={formError}
        onSubmit={handleFormSubmit}
        itemSaving={itemSaving}
        itemError={itemError}
        removingItemId={removingItemId}
        onAddItem={handleAddItem}
        onRemoveItem={handleRemoveItem}
        pendingItems={formPendingItems}
        onAddPendingItem={addPendingItem}
        onRemovePendingItem={removePendingItem}
      />
    </main>
  )
}
