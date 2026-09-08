import { useEffect, useState, type FormEvent } from 'react'
import axios from 'axios'
import { api, extractErrorCode, extractErrorMessage, type ApiPage } from '@/shared/api/client'
import { useLanguage } from '@/shared/i18n/language'
import { ConfirmDialog } from '@/shared/components/confirm-dialog'
import { InformationDialog } from '@/shared/components/information-dialog'
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
  type InvoiceColumnFilters,
  type InvoiceSortDirection,
  type InvoiceSortKey,
  type InvoiceStatusFilter,
  type InvoiceTextFilterKey,
} from './utils'

const SEARCH_DEBOUNCE_MS = 300

export default function InvoicePage() {
  const { t } = useLanguage()

  const [invoices, setInvoices] = useState<ApiInvoice[] | null>(null)
  const [contracts, setContracts] = useState<ApiContract[]>([])
  const [total, setTotal] = useState(0)
  const [totalPages, setTotalPages] = useState(1)
  const [loadError, setLoadError] = useState<string | null>(null)

  const [query, setQuery] = useState('')
  const [debouncedQuery, setDebouncedQuery] = useState('')
  const [statusFilter, setStatusFilter] = useState<InvoiceStatusFilter>('all')
  // Applied on submit from each column's menu, so no debounce is needed here.
  const [columnFilters, setColumnFilters] = useState<InvoiceColumnFilters>({})
  const [sortKey, setSortKey] = useState<InvoiceSortKey>('period')
  const [sortDirection, setSortDirection] = useState<InvoiceSortDirection>('desc')

  const [page, setPage] = useState(1)
  const [pageSize, setPageSizeState] = useState<number>(INVOICE_PAGE_SIZE_OPTIONS[0])
  // Bumped by mutations so the list refetches; the server owns the ordering
  // and page boundaries now, so patching rows locally would misplace them.
  const [refreshToken, setRefreshToken] = useState(0)

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
  const [lineSendResult, setLineSendResult] = useState<{ title: string; description: string } | null>(null)

  useEffect(() => {
    const timer = window.setTimeout(() => setDebouncedQuery(query), SEARCH_DEBOUNCE_MS)
    return () => window.clearTimeout(timer)
  }, [query])

  // The contract selector in the create form isn't affected by the list's
  // filters, so it's loaded once rather than on every refetch.
  useEffect(() => {
    let cancelled = false

    api
      .get<ApiPage<ApiContract[]>>('/contracts', { params: { status: 'active', per_page: 100 } })
      .then(({ data }) => {
        if (!cancelled) setContracts(data.data)
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
      .get<ApiPage<ApiInvoice[]>>('/invoices', {
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
        setInvoices(data.data)
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

  const isLoading = !loadError && invoices === null
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

  function handleSort(key: InvoiceSortKey) {
    if (key === sortKey) {
      setSortDirection(sortDirection === 'asc' ? 'desc' : 'asc')
    } else {
      setSortKey(key)
      setSortDirection('asc')
    }
    setPage(1)
  }

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

        let finalInvoice = created
        for (const item of formPendingItems) {
          try {
            const { data: updated } = await api.post<ApiInvoice>(`/invoices/${created.id}/items`, item)
            finalInvoice = updated
          } catch (err) {
            openEditForm(finalInvoice)
            setFormError(extractErrorMessage(err, t('invoiceItemAddError')))
            return
          }
        }

        refresh()
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
      await api.put<ApiInvoice>(`/invoices/${formInvoiceId}`, payload)
      refresh()
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
      // Deleting the last row of the final page would otherwise strand the
      // view on a page the server no longer has.
      setPage((value) => (invoices?.length === 1 ? Math.max(1, value - 1) : value))
      refresh()
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
      setLineSendResult({
        title: t('invoiceSendLineSuccessTitle'),
        description: t('invoiceSendLineSuccessDescription')
          .replace('{room}', invoice.room_number ?? '-')
          .replace('{tenant}', invoice.tenant_name ?? '-'),
      })
    } catch (err) {
      const code = extractErrorCode(err)
      const description =
        code === 'tenant_line_not_linked'
          ? t('invoiceSendLineNotLinkedDescription')
          : code === 'tenant_line_unreachable'
            ? t('invoiceSendLineUnreachableDescription')
            : extractErrorMessage(err, t('invoiceSendLineError'))

      setLineSendResult({ title: t('invoiceSendLineErrorTitle'), description })
    } finally {
      setSendingLineInvoiceId(null)
    }
  }

  // Tenants who only have a personal LINE ID (haven't gone through the
  // OA/LIFF linking flow) can't be pushed to automatically — the LINE
  // Messaging API only reaches linked users. So we open a manual chat with
  // them and copy the same formatted invoice text to the clipboard, ready
  // to paste in.
  async function handleOpenLineChat(invoice: ApiInvoice) {
    if (!invoice.tenant_line_id) return

    window.open(
      `https://line.me/ti/p/~${encodeURIComponent(invoice.tenant_line_id)}`,
      '_blank',
      'noopener,noreferrer',
    )

    try {
      const { data } = await api.get<{ message: string }>(`/invoices/${invoice.id}/line-message`)
      await navigator.clipboard.writeText(data.message)
      setLineSendResult({
        title: t('invoiceLineChatCopiedTitle'),
        description: t('invoiceLineChatCopiedDescription'),
      })
    } catch {
      setLineSendResult({
        title: t('invoiceLineChatManualTitle'),
        description: t('invoiceLineChatManualDescription'),
      })
    }
  }

  function applyInvoiceUpdate(data: ApiInvoice) {
    setFormInvoiceDetail(data)
    refresh()
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
        onColumnFilterChange={(key: InvoiceTextFilterKey, value: string) => {
          setColumnFilters((prev) => ({ ...prev, [key]: value }))
          setPage(1)
        }}
        hasFilters={hasFilters}
        sortKey={sortKey}
        sortDirection={sortDirection}
        onSort={handleSort}
        invoices={invoices ?? []}
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
        deletingInvoiceId={deletingInvoiceId}
        onCreateInvoice={openCreateForm}
        onEditInvoice={openEditForm}
        onDeleteInvoice={setConfirmDeleteInvoice}
        sendingLineInvoiceId={sendingLineInvoiceId}
        onSendLineInvoice={handleSendLineInvoice}
        onOpenLineChat={handleOpenLineChat}
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

      <InformationDialog
        open={lineSendResult !== null}
        onOpenChange={(open) => !open && setLineSendResult(null)}
        title={lineSendResult?.title ?? ''}
        description={lineSendResult?.description ?? ''}
        actionLabel={t('acknowledge')}
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
