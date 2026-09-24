import { useEffect, useState } from 'react'
import { useParams } from 'react-router-dom'
import axios from 'axios'
import { Printer } from 'lucide-react'
import { api, extractErrorMessage } from '@/shared/api/client'
import { useLanguage, type TranslationKey } from '@/shared/i18n/language'
import { Button } from '@/shared/components/ui/button'
import { formatPeriod } from '@/features/invoice/utils'
import type { ApiReceipt } from './types'

const paymentMethodLabelKeys: Record<string, TranslationKey> = {
  cash: 'paymentMethodCash',
  transfer: 'paymentMethodTransfer',
  credit_card: 'paymentMethodCreditCard',
  other: 'paymentMethodOther',
  deposit: 'paymentMethodDeposit',
}

function formatMoney(value: number): string {
  return value.toLocaleString('th-TH', { minimumFractionDigits: 2, maximumFractionDigits: 2 })
}

// Printable receipt for one payment, opened in its own tab from the payment
// list. Every payment is a numbered receipt; a voided one still prints, stamped
// as voided with its reason, so a reprint can't pass for a valid receipt.
export default function ReceiptPrintPage() {
  const { id } = useParams<{ id: string }>()
  const { t, language } = useLanguage()
  const [receipt, setReceipt] = useState<ApiReceipt | null>(null)
  const [error, setError] = useState<string | null>(null)

  useEffect(() => {
    if (!id) return
    const controller = new AbortController()
    api
      .get<ApiReceipt>(`/payments/${id}/receipt`, { signal: controller.signal })
      .then(({ data }) => setReceipt(data))
      .catch((err) => {
        if (axios.isCancel(err)) return
        setError(extractErrorMessage(err, t('resourceLoadError')))
      })
    return () => controller.abort()
    // eslint-disable-next-line react-hooks/exhaustive-deps
  }, [id])

  // Browsers use the tab title as the default PDF file name.
  const receiptNo = receipt?.payment.receipt_no
  useEffect(() => {
    if (!receiptNo) return
    const previous = document.title
    document.title = `${t('receiptPrintTitle')} ${receiptNo}`
    return () => {
      document.title = previous
    }
  }, [receiptNo, t])

  if (error) {
    return <p className="p-6 text-sm text-red-600">{error}</p>
  }
  if (!receipt) {
    return <p className="p-6 text-sm text-gray-500">{t('loading')}</p>
  }

  const { payment, dormitory, invoice } = receipt
  const isVoided = payment.status === 'voided'
  const dateLocale = language === 'th' ? 'th-TH' : 'en-US'
  const formatDate = (value: string) =>
    new Date(value).toLocaleDateString(dateLocale, { year: 'numeric', month: 'long', day: 'numeric', timeZone: 'UTC' })

  return (
    <div className="min-h-screen bg-gray-100 py-6 text-gray-900 print:bg-white print:py-0">
      <div className="mx-auto mb-4 flex max-w-[210mm] justify-end px-4 print:hidden">
        <Button onClick={() => window.print()}>
          <Printer className="size-4" />
          {t('invoicePrintAction')}
        </Button>
      </div>

      <article className="relative mx-auto flex max-w-[210mm] flex-col gap-6 overflow-hidden bg-white p-8 shadow-sm print:max-w-none print:p-0 print:shadow-none">
        {isVoided && (
          <div
            aria-hidden="true"
            className="pointer-events-none absolute inset-0 flex items-center justify-center text-7xl font-bold tracking-widest text-red-600/20 select-none"
            style={{ transform: 'rotate(-20deg)' }}
          >
            {t('receiptPrintVoided')}
          </div>
        )}

        <header className="flex flex-col gap-4 border-b border-gray-300 pb-4 sm:flex-row sm:items-start sm:justify-between">
          <div className="flex flex-col gap-1">
            <h1 className="text-xl font-semibold">{dormitory.name}</h1>
            {dormitory.address && <p className="text-sm text-gray-600">{dormitory.address}</p>}
            {dormitory.phone && (
              <p className="text-sm text-gray-600">
                {t('invoicePrintPhone')} {dormitory.phone}
              </p>
            )}
          </div>
          <div className="flex flex-col gap-1 sm:text-right">
            <h2 className="text-lg font-semibold">{t('receiptPrintTitle')}</h2>
            <p className="text-sm">
              {t('receiptPrintNo')} <span className="font-semibold">{payment.receipt_no || '—'}</span>
            </p>
            <p className="text-sm text-gray-600">
              {t('receiptPrintDate')} {formatDate(payment.payment_date)}
            </p>
          </div>
        </header>

        {isVoided && (
          <p className="rounded-md border border-red-300 bg-red-50 px-3 py-2 text-sm text-red-700">
            {t('receiptPrintVoided')}
            {payment.voided_at && ` · ${formatDate(payment.voided_at)}`}
            {payment.void_reason && ` · ${t('receiptPrintVoidReason')}: ${payment.void_reason}`}
          </p>
        )}

        <section className="grid grid-cols-2 gap-2 text-sm">
          <div>
            <span className="text-gray-600">{t('invoiceFormTenantLabel')}: </span>
            <span className="font-medium">{payment.tenant_name || '—'}</span>
          </div>
          <div className="text-right">
            <span className="text-gray-600">{t('invoiceFormRoomLabel')}: </span>
            <span className="font-medium">{payment.room_number || '—'}</span>
          </div>
          <div className="col-span-2">
            <span className="text-gray-600">{t('receiptPrintFor')} </span>
            <span className="font-medium">{formatPeriod(invoice.period_year, invoice.period_month)}</span>
          </div>
        </section>

        <table className="w-full border-collapse text-sm">
          <thead>
            <tr className="border-b border-gray-300 text-left">
              <th className="py-2 font-semibold">{t('receiptPrintMethodColumn')}</th>
              <th className="py-2 font-semibold">{t('receiptPrintReferenceColumn')}</th>
              <th className="py-2 text-right font-semibold">{t('invoicePrintAmountColumn')}</th>
            </tr>
          </thead>
          <tbody>
            {payment.items.map((item) => (
              <tr key={item.id} className="border-b border-gray-200">
                <td className="py-2">{t(paymentMethodLabelKeys[item.payment_method] ?? 'paymentMethodOther')}</td>
                <td className="py-2 text-gray-600">{item.reference_no || '—'}</td>
                <td className="py-2 text-right">{formatMoney(item.amount)}</td>
              </tr>
            ))}
          </tbody>
          <tfoot>
            <tr>
              <td colSpan={2} className="pt-3 text-base font-semibold">
                {t('receiptPrintTotal')}
              </td>
              <td className="pt-3 text-right text-base font-semibold">{formatMoney(payment.total_amount)}</td>
            </tr>
          </tfoot>
        </table>

        <section className="ml-auto flex w-full max-w-xs flex-col gap-1 text-sm text-gray-600">
          <div className="flex justify-between">
            <span>{t('receiptPrintInvoiceTotal')}</span>
            <span>{formatMoney(invoice.total_amount)}</span>
          </div>
          <div className="flex justify-between">
            <span>{t('receiptPrintPaidToDate')}</span>
            <span>{formatMoney(invoice.paid_amount)}</span>
          </div>
          <div className="flex justify-between font-medium text-gray-900">
            <span>{t('receiptPrintRemaining')}</span>
            <span>{formatMoney(invoice.outstanding)}</span>
          </div>
        </section>

        {payment.note && <p className="text-sm text-gray-600">{payment.note}</p>}

        <footer className="mt-8 flex justify-end text-sm">
          <div className="flex w-56 flex-col items-center gap-1">
            <div className="h-10 w-full border-b border-gray-400" />
            <span className="text-gray-600">{t('invoicePrintReceivedBy')}</span>
          </div>
        </footer>
      </article>
    </div>
  )
}
