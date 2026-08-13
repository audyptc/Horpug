import { useEffect, useMemo, useState, type FormEvent } from 'react'
import { api, extractErrorMessage, type ApiPage } from '@/shared/api/client'
import { useLanguage } from '@/shared/i18n/language'
import { ConfirmDialog } from '@/shared/components/confirm-dialog'
import type { ApiInvoice } from '@/features/invoice/types'
import { PaymentListCard } from './components/PaymentListCard'
import { PaymentFormSheet } from './components/PaymentFormSheet'
import type { ApiPayment } from './types'
import {
  PAYMENT_PAGE_SIZE_OPTIONS,
  createPaymentItemRow,
  toApiDate,
  toDateInputValue,
  type PaymentItemFormRow,
} from './utils'

export default function PaymentPage() {
  const { t } = useLanguage()

  const [payments, setPayments] = useState<ApiPayment[] | null>(null)
  const [invoices, setInvoices] = useState<ApiInvoice[]>([])
  const [loadError, setLoadError] = useState<string | null>(null)

  const [query, setQuery] = useState('')
  const [page, setPage] = useState(1)
  const [pageSize, setPageSize] = useState<number>(PAYMENT_PAGE_SIZE_OPTIONS[0])

  const [formOpen, setFormOpen] = useState(false)
  const [formInvoiceId, setFormInvoiceId] = useState('')
  const [formPaymentDate, setFormPaymentDate] = useState('')
  const [formItems, setFormItems] = useState<PaymentItemFormRow[]>([])
  const [formNote, setFormNote] = useState('')
  const [formSaving, setFormSaving] = useState(false)
  const [formError, setFormError] = useState<string | null>(null)

  const [deletingPaymentId, setDeletingPaymentId] = useState<string | null>(null)
  const [deleteError, setDeleteError] = useState<string | null>(null)
  const [confirmDeletePayment, setConfirmDeletePayment] = useState<ApiPayment | null>(null)

  useEffect(() => {
    let cancelled = false

    Promise.all([
      api.get<ApiPage<ApiPayment[]>>('/payments', { params: { per_page: 100 } }),
      api.get<ApiPage<ApiInvoice[]>>('/invoices', { params: { per_page: 100 } }),
    ])
      .then(([paymentsRes, invoicesRes]) => {
        if (cancelled) return
        setPayments(paymentsRes.data.data)
        setInvoices(invoicesRes.data.data.filter((invoice) => invoice.status !== 'cancelled'))
      })
      .catch((err) => {
        if (!cancelled) setLoadError(extractErrorMessage(err, t('resourceLoadError')))
      })

    return () => {
      cancelled = true
    }
    // eslint-disable-next-line react-hooks/exhaustive-deps
  }, [])

  const filteredPayments = useMemo(() => {
    const q = query.trim().toLocaleLowerCase()
    if (!q) return payments ?? []

    return (payments ?? []).filter((payment) => {
      return (
        (payment.tenant_name ?? '').toLocaleLowerCase().includes(q) ||
        (payment.room_number ?? '').toLocaleLowerCase().includes(q) ||
        (payment.dormitory_name ?? '').toLocaleLowerCase().includes(q) ||
        payment.items.some((item) => item.reference_no.toLocaleLowerCase().includes(q))
      )
    })
  }, [query, payments])

  const totalPages = Math.max(1, Math.ceil(filteredPayments.length / pageSize))
  const currentPage = Math.min(page, totalPages)
  const rangeStart = filteredPayments.length === 0 ? 0 : (currentPage - 1) * pageSize + 1
  const rangeEnd = Math.min(currentPage * pageSize, filteredPayments.length)
  const paginatedPayments = filteredPayments.slice((currentPage - 1) * pageSize, currentPage * pageSize)

  const isLoading = !loadError && payments === null

  function openCreateForm() {
    setFormInvoiceId('')
    setFormPaymentDate(toDateInputValue(new Date().toISOString()))
    setFormItems([createPaymentItemRow()])
    setFormNote('')
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

    try {
      const payload = {
        invoice_id: formInvoiceId,
        payment_date: toApiDate(formPaymentDate),
        note: formNote.trim(),
        items: formItems.map((item) => ({
          payment_method: item.paymentMethod,
          amount: Number(item.amount),
          reference_no: item.referenceNo.trim(),
        })),
      }
      const { data } = await api.post<ApiPayment>('/payments', payload)
      setPayments((prev) => [data, ...(prev ?? [])])
      setFormOpen(false)
    } catch (err) {
      setFormError(extractErrorMessage(err, t('paymentCreateError')))
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
      await api.delete(`/payments/${payment.id}`)
      setPayments((prev) => prev?.filter((item) => item.id !== payment.id) ?? prev)
      setConfirmDeletePayment(null)
    } catch (err) {
      setDeleteError(extractErrorMessage(err, t('paymentDeleteError')))
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

      <PaymentListCard
        isLoading={isLoading}
        loadError={loadError}
        deleteError={deleteError}
        payments={payments}
        query={query}
        onQueryChange={(value) => {
          setQuery(value)
          setPage(1)
        }}
        filteredPayments={filteredPayments}
        paginatedPayments={paginatedPayments}
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
        deletingPaymentId={deletingPaymentId}
        onCreatePayment={openCreateForm}
        onDeletePayment={setConfirmDeletePayment}
      />

      <ConfirmDialog
        open={confirmDeletePayment !== null}
        onOpenChange={(open) => !open && setConfirmDeletePayment(null)}
        title={t('confirmDeleteTitle')}
        description={t('paymentDeleteConfirm')}
        confirmLabel={t('paymentDelete')}
        cancelLabel={t('cancel')}
        loading={deletingPaymentId === confirmDeletePayment?.id}
        onConfirm={handleDeletePayment}
      />

      <PaymentFormSheet
        open={formOpen}
        onOpenChange={setFormOpen}
        invoices={invoices}
        invoiceId={formInvoiceId}
        onInvoiceIdChange={setFormInvoiceId}
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
