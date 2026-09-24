import { useEffect, useState, type FormEvent } from 'react'
import axios from 'axios'
import { api, extractErrorCode, extractErrorMessage, type ApiPage } from '@/shared/api/client'
import { useLanguage } from '@/shared/i18n/language'
import { ConfirmDialog } from '@/shared/components/confirm-dialog'
import type { ApiInvoice } from '@/features/invoice/types'
import { formatInvoiceLabel } from '@/features/invoice/utils'
import { PaymentListCard } from './components/PaymentListCard'
import { PaymentFormSheet } from './components/PaymentFormSheet'
import { SlipReviewSheet } from './components/SlipReviewSheet'
import type { ApiSlip } from './slips'
import type { ApiPayment } from './types'
import {
  PAYMENT_PAGE_SIZE_OPTIONS,
  createPaymentItemRow,
  toApiDate,
  toDateInputValue,
  type PaymentColumnFilters,
  type PaymentItemFormRow,
  type PaymentMethodFilter,
  type PaymentSortDirection,
  type PaymentSortKey,
  type PaymentTextFilterKey,
} from './utils'

const SEARCH_DEBOUNCE_MS = 300

export default function PaymentPage() {
  const { t } = useLanguage()

  const [payments, setPayments] = useState<ApiPayment[] | null>(null)
  const [total, setTotal] = useState(0)
  const [totalPages, setTotalPages] = useState(1)
  const [loadError, setLoadError] = useState<string | null>(null)

  const [query, setQuery] = useState('')
  const [debouncedQuery, setDebouncedQuery] = useState('')
  const [methodFilter, setMethodFilter] = useState<PaymentMethodFilter>('all')
  // Applied on submit from each column's menu, so no debounce is needed here.
  const [columnFilters, setColumnFilters] = useState<PaymentColumnFilters>({})
  const [sortKey, setSortKey] = useState<PaymentSortKey>('payment_date')
  const [sortDirection, setSortDirection] = useState<PaymentSortDirection>('desc')

  const [page, setPage] = useState(1)
  const [pageSize, setPageSizeState] = useState<number>(PAYMENT_PAGE_SIZE_OPTIONS[0])
  // Bumped by mutations so the list refetches; the server owns the ordering
  // and page boundaries now, so patching rows locally would misplace them.
  const [refreshToken, setRefreshToken] = useState(0)
  const [slipSheetOpen, setSlipSheetOpen] = useState(false)
  const [pendingSlipCount, setPendingSlipCount] = useState(0)

  // Pending slips drive the badge on the review button; refetched with the
  // payment list, since approving a slip adds a payment.
  useEffect(() => {
    const controller = new AbortController()
    api
      .get<ApiSlip[]>('/payment-slips', { params: { status: 'pending' }, signal: controller.signal })
      .then(({ data }) => setPendingSlipCount(data.length))
      .catch(() => {
        // The badge is a hint; the list itself reports errors when opened.
      })
    return () => controller.abort()
  }, [refreshToken])

  const [formOpen, setFormOpen] = useState(false)
  // Set while editing an existing payment; null means the form is recording a new one.
  const [editingPaymentId, setEditingPaymentId] = useState<string | null>(null)
  const [formInvoiceId, setFormInvoiceId] = useState('')
  const [formInvoiceLabel, setFormInvoiceLabel] = useState('')
  const [formPaymentDate, setFormPaymentDate] = useState('')
  const [formItems, setFormItems] = useState<PaymentItemFormRow[]>([])
  const [formNote, setFormNote] = useState('')
  const [formSaving, setFormSaving] = useState(false)
  const [formError, setFormError] = useState<string | null>(null)

  const [deletingPaymentId, setDeletingPaymentId] = useState<string | null>(null)
  const [deleteError, setDeleteError] = useState<string | null>(null)
  const [confirmDeletePayment, setConfirmDeletePayment] = useState<ApiPayment | null>(null)
  const [voidReason, setVoidReason] = useState('')

  useEffect(() => {
    const timer = window.setTimeout(() => setDebouncedQuery(query), SEARCH_DEBOUNCE_MS)
    return () => window.clearTimeout(timer)
  }, [query])

  useEffect(() => {
    const controller = new AbortController()

    const columnParams: Record<string, string> = {}
    for (const [key, value] of Object.entries(columnFilters)) {
      if (value) columnParams[`f[${key}]`] = value
    }

    api
      .get<ApiPage<ApiPayment[]>>('/payments', {
        signal: controller.signal,
        params: {
          page,
          per_page: pageSize,
          q: debouncedQuery.trim() || undefined,
          payment_method: methodFilter === 'all' ? undefined : methodFilter,
          sort: sortKey,
          order: sortDirection,
          ...columnParams,
        },
      })
      .then(({ data }) => {
        setPayments(data.data)
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
  }, [page, pageSize, debouncedQuery, methodFilter, columnFilters, sortKey, sortDirection, refreshToken])

  const isLoading = !loadError && payments === null
  const hasFilters = query !== '' || methodFilter !== 'all' || Object.values(columnFilters).some(Boolean)
  const rangeStart = total === 0 ? 0 : (page - 1) * pageSize + 1
  const rangeEnd = Math.min(page * pageSize, total)

  function refresh() {
    setRefreshToken((value) => value + 1)
  }

  function setPageSize(size: number) {
    setPageSizeState(size)
    setPage(1)
  }

  function handleSort(key: PaymentSortKey) {
    if (key === sortKey) {
      setSortDirection(sortDirection === 'asc' ? 'desc' : 'asc')
    } else {
      setSortKey(key)
      setSortDirection('asc')
    }
    setPage(1)
  }

  function openCreateForm() {
    setEditingPaymentId(null)
    setFormInvoiceId('')
    setFormInvoiceLabel('')
    setFormPaymentDate(toDateInputValue(new Date().toISOString()))
    setFormItems([createPaymentItemRow()])
    setFormNote('')
    setFormError(null)
    setFormOpen(true)
  }

  function openEditForm(payment: ApiPayment) {
    const room = `${payment.room_number ?? ''}${payment.dormitory_name ? ` (${payment.dormitory_name})` : ''}`

    setEditingPaymentId(payment.id)
    setFormInvoiceId(payment.invoice_id)
    setFormInvoiceLabel(`${payment.tenant_name ?? ''} · ${room}`)
    setFormPaymentDate(toDateInputValue(payment.payment_date))
    setFormItems(
      payment.items.map((item) => ({
        ...createPaymentItemRow(),
        paymentMethod: item.payment_method,
        amount: String(item.amount),
        referenceNo: item.reference_no,
      }))
    )
    setFormNote(payment.note)
    setFormError(null)
    setFormOpen(true)
  }

  function handleAddItem() {
    setFormItems((prev) => [...prev, createPaymentItemRow()])
  }

  function handleRemoveItem(key: number) {
    setFormItems((prev) => (prev.length <= 1 ? prev : prev.filter((item) => item.key !== key)))
  }

  function handleItemChange(key: number, patch: Partial<PaymentItemFormRow>) {
    setFormItems((prev) => prev.map((item) => (item.key === key ? { ...item, ...patch } : item)))
  }

  async function handleFormSubmit(event: FormEvent<HTMLFormElement>) {
    event.preventDefault()

    if (!formInvoiceId) {
      setFormError(t('paymentInvoiceRequired'))
      return
    }
    if (!formPaymentDate) {
      setFormError(t('paymentDateRequired'))
      return
    }
    if (formItems.length === 0) {
      setFormError(t('paymentItemsRequired'))
      return
    }
    for (const item of formItems) {
      const amount = Number(item.amount)
      if (!item.amount || Number.isNaN(amount) || amount <= 0) {
        setFormError(t('paymentAmountInvalid'))
        return
      }
    }

    setFormSaving(true)
    setFormError(null)

    const isEditing = editingPaymentId !== null

    try {
      const payload = {
        payment_date: toApiDate(formPaymentDate),
        note: formNote.trim(),
        items: formItems.map((item) => ({
          payment_method: item.paymentMethod,
          amount: Number(item.amount),
          reference_no: item.referenceNo.trim(),
        })),
      }
      if (isEditing) {
        await api.put<ApiPayment>(`/payments/${editingPaymentId}`, payload)
      } else {
        await api.post<ApiPayment>('/payments', { ...payload, invoice_id: formInvoiceId })
      }
      refresh()
      setFormOpen(false)
    } catch (err) {
      const code = extractErrorCode(err)
      setFormError(
        code === 'payment_exceeds_invoice'
          ? t('paymentExceedsInvoice')
          : code === 'receipt_issued'
            ? t('paymentReceiptIssuedError')
            : extractErrorMessage(err, t(isEditing ? 'paymentUpdateError' : 'paymentCreateError')),
      )
    } finally {
      setFormSaving(false)
    }
  }

  async function handleDeletePayment() {
    if (!confirmDeletePayment) return
    const payment = confirmDeletePayment

    setDeletingPaymentId(payment.id)
    setDeleteError(null)

    try {
      // Payments carry receipt numbers, so they are voided (kept, marked
      // cancelled) rather than deleted; the row stays in the list.
      await api.post(`/payments/${payment.id}/void`, { reason: voidReason.trim() })
      refresh()
      setConfirmDeletePayment(null)
    } catch (err) {
      setDeleteError(extractErrorMessage(err, t('paymentVoidError')))
    } finally {
      setDeletingPaymentId(null)
    }
  }

  return (
    <main className="content">
      <section className="welcome">
        <h1>{t('menuPayments')}</h1>
        <p>{t('menuPaymentsDescription')}</p>
      </section>

      <SlipReviewSheet open={slipSheetOpen} onOpenChange={setSlipSheetOpen} onReviewed={refresh} />

      <PaymentListCard
        pendingSlipCount={pendingSlipCount}
        onReviewSlips={() => setSlipSheetOpen(true)}
        isLoading={isLoading}
        loadError={loadError}
        deleteError={deleteError}
        query={query}
        onQueryChange={(value) => {
          setQuery(value)
          setPage(1)
        }}
        methodFilter={methodFilter}
        onMethodFilterChange={(value) => {
          setMethodFilter(value)
          setPage(1)
        }}
        columnFilters={columnFilters}
        onColumnFilterChange={(key: PaymentTextFilterKey, value: string) => {
          setColumnFilters((prev) => ({ ...prev, [key]: value }))
          setPage(1)
        }}
        hasFilters={hasFilters}
        sortKey={sortKey}
        sortDirection={sortDirection}
        onSort={handleSort}
        payments={payments ?? []}
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
        deletingPaymentId={deletingPaymentId}
        onCreatePayment={openCreateForm}
        onEditPayment={openEditForm}
        onVoidPayment={(payment) => {
          setVoidReason('')
          setConfirmDeletePayment(payment)
        }}
      />

      <ConfirmDialog
        open={confirmDeletePayment !== null}
        onOpenChange={(open) => !open && setConfirmDeletePayment(null)}
        title={t('paymentVoidTitle')}
        description={
          confirmDeletePayment
            ? t('paymentVoidConfirm')
                .replace('{receipt}', confirmDeletePayment.receipt_no || '-')
                .replace('{tenant}', confirmDeletePayment.tenant_name || '-')
                .replace('{room}', confirmDeletePayment.room_number || '-')
                .replace('{amount}', confirmDeletePayment.total_amount.toLocaleString())
            : ''
        }
        confirmLabel={t('paymentVoid')}
        cancelLabel={t('cancel')}
        loading={deletingPaymentId === confirmDeletePayment?.id}
        error={deleteError}
        onConfirm={handleDeletePayment}
      >
        <label className="flex flex-col gap-1.5 text-sm font-medium">
          {t('paymentVoidReasonLabel')}
          <textarea
            className="min-h-16 w-full min-w-0 rounded-md border border-input bg-transparent px-3 py-2 text-sm font-normal"
            value={voidReason}
            onChange={(event) => setVoidReason(event.target.value)}
          />
        </label>
      </ConfirmDialog>

      <PaymentFormSheet
        open={formOpen}
        onOpenChange={setFormOpen}
        isEditing={editingPaymentId !== null}
        invoiceLabel={formInvoiceLabel}
        onSelectInvoice={(invoice: ApiInvoice) => {
          setFormInvoiceId(invoice.id)
          setFormInvoiceLabel(formatInvoiceLabel(invoice))
        }}
        paymentDate={formPaymentDate}
        onPaymentDateChange={setFormPaymentDate}
        items={formItems}
        onAddItem={handleAddItem}
        onRemoveItem={handleRemoveItem}
        onItemChange={handleItemChange}
        note={formNote}
        onNoteChange={setFormNote}
        saving={formSaving}
        error={formError}
        onSubmit={handleFormSubmit}
      />
    </main>
  )
}
