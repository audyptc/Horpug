import { useEffect, useState } from 'react'
import { useParams } from 'react-router-dom'
import axios from 'axios'
import { QRCodeSVG } from 'qrcode.react'
import { Printer } from 'lucide-react'
import { api, extractErrorMessage } from '@/shared/api/client'
import { useLanguage, type TranslationKey } from '@/shared/i18n/language'
import { Button } from '@/shared/components/ui/button'
import type { ApiInvoiceDocument } from './types'
import { formatPeriod } from './utils'

const itemTypeLabelKeys: Record<string, TranslationKey> = {
  rent: 'invoiceItemTypeRent',
  electricity: 'invoiceItemTypeElectricity',
  water: 'invoiceItemTypeWater',
  other: 'invoiceItemTypeOther',
}

const paymentMethodLabelKeys: Record<string, TranslationKey> = {
  cash: 'paymentMethodCash',
  transfer: 'paymentMethodTransfer',
  credit_card: 'paymentMethodCreditCard',
  other: 'paymentMethodOther',
}

function formatMoney(value: number): string {
  return value.toLocaleString('th-TH', { minimumFractionDigits: 2, maximumFractionDigits: 2 })
}

// Printable invoice / receipt, opened in its own tab from the invoice list.
// A fully paid invoice prints as a receipt listing its payments; otherwise it
// prints as an invoice with a PromptPay QR for the amount still owed (when
// the dormitory has a PromptPay ID). The browser's print dialog also saves
// it as a PDF. Rendered as white paper regardless of the app theme.
export default function InvoicePrintPage() {
  const { id } = useParams<{ id: string }>()
  const { t, language } = useLanguage()
  const [doc, setDoc] = useState<ApiInvoiceDocument | null>(null)
  const [error, setError] = useState<string | null>(null)

  useEffect(() => {
    if (!id) return
    const controller = new AbortController()
    api
      .get<ApiInvoiceDocument>(`/invoices/${id}/document`, { signal: controller.signal })
      .then(({ data }) => setDoc(data))
      .catch((err) => {
        if (axios.isCancel(err)) return
        setError(extractErrorMessage(err, t('resourceLoadError')))
      })
    return () => controller.abort()
    // eslint-disable-next-line react-hooks/exhaustive-deps
  }, [id])

  const dateLocale = language === 'th' ? 'th-TH' : 'en-US'
  const formatDate = (value: string) =>
    new Date(value).toLocaleDateString(dateLocale, { year: 'numeric', month: 'long', day: 'numeric', timeZone: 'UTC' })

  // Receipts are per payment now (see ReceiptPrintPage); this is always the
  // invoice, marked paid once settled.
  const isPaid = doc?.invoice.status === 'paid'
  const title = t('invoicePrintInvoiceTitle')
  useDocumentTitle(
    doc
      ? `${title} ${doc.invoice.room_number ?? ''} ${formatPeriod(doc.invoice.period_year, doc.invoice.period_month)}`
      : ''
  )

  if (error) {
    return <p className="p-6 text-sm text-red-600">{error}</p>
  }
  if (!doc) {
    return <p className="p-6 text-sm text-gray-500">{t('loading')}</p>
  }

  const { invoice, dormitory } = doc
  const isCancelled = invoice.status === 'cancelled'
  const tenantName = invoice.tenant_name ?? '—'

  return (
    <div className="min-h-screen bg-gray-100 py-6 text-gray-900 print:bg-white print:py-0">
      <div className="mx-auto mb-4 flex max-w-[210mm] justify-end px-4 print:hidden">
        <Button onClick={() => window.print()}>
          <Printer className="size-4" />
          {t('invoicePrintAction')}
        </Button>
      </div>

      <article className="mx-auto flex max-w-[210mm] flex-col gap-6 bg-white p-8 shadow-sm print:max-w-none print:p-0 print:shadow-none">
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
            <h2 className="text-lg font-semibold">{title}</h2>
            {isCancelled && <p className="text-sm font-semibold text-red-600">{t('invoiceStatusCancelled')}</p>}
            {isPaid && <p className="text-sm font-semibold text-green-700">{t('invoiceStatusPaid')}</p>}
            <p className="text-sm text-gray-600">
              {t('invoiceFormPeriodLabel')}: {formatPeriod(invoice.period_year, invoice.period_month)}
            </p>
            <p className="text-sm text-gray-600">
              {t('invoiceFormIssueDateLabel')}: {formatDate(invoice.issue_date)}
            </p>
            {!isPaid && (
              <p className="text-sm text-gray-600">
                {t('invoiceFormDueDateLabel')}: {formatDate(invoice.due_date)}
              </p>
            )}
          </div>
        </header>

        <section className="grid grid-cols-2 gap-2 text-sm">
          <div>
            <span className="text-gray-600">{t('invoiceFormTenantLabel')}: </span>
            <span className="font-medium">{tenantName}</span>
          </div>
          <div className="text-right">
            <span className="text-gray-600">{t('invoiceFormRoomLabel')}: </span>
            <span className="font-medium">{invoice.room_number ?? '—'}</span>
          </div>
        </section>

        <table className="w-full border-collapse text-sm">
          <thead>
            <tr className="border-b border-gray-300 text-left">
              <th className="py-2 font-semibold">{t('invoicePrintItemColumn')}</th>
              <th className="py-2 text-right font-semibold">{t('invoicePrintAmountColumn')}</th>
            </tr>
          </thead>
          <tbody>
            {(invoice.items ?? []).map((item) => (
              <tr key={item.id} className="border-b border-gray-200">
                <td className="py-2">
                  {t(itemTypeLabelKeys[item.item_type] ?? 'invoiceItemTypeOther')}
                  {item.description && item.item_type === 'other' ? ` · ${item.description}` : ''}
                </td>
                <td className="py-2 text-right">{formatMoney(item.amount)}</td>
              </tr>
            ))}
          </tbody>
          <tfoot>
            <tr>
              <td className="pt-3 font-semibold">{t('invoiceFormTotalAmountLabel')}</td>
              <td className="pt-3 text-right font-semibold">{formatMoney(invoice.total_amount)}</td>
            </tr>
            {doc.paid_amount > 0 && (
              <tr>
                <td className="pt-1 text-gray-600">{t('invoicePrintPaidLabel')}</td>
                <td className="pt-1 text-right text-gray-600">{formatMoney(doc.paid_amount)}</td>
              </tr>
            )}
            {!isPaid && !isCancelled && (
              <tr>
                <td className="pt-1 text-base font-semibold">{t('invoicePrintOutstandingLabel')}</td>
                <td className="pt-1 text-right text-base font-semibold">{formatMoney(doc.outstanding)}</td>
              </tr>
            )}
          </tfoot>
        </table>

        {doc.payments.length > 0 && (
          <section className="flex flex-col gap-2 text-sm">
            <h3 className="font-semibold">{t('invoicePrintPaymentsTitle')}</h3>
            <ul className="flex flex-col gap-1">
              {doc.payments.map((payment) => (
                <li key={payment.id} className="flex justify-between gap-4 border-b border-gray-200 py-1">
                  <span>
                    {payment.receipt_no && `${payment.receipt_no} · `}
                    {formatDate(payment.payment_date)}
                    {payment.items.length > 0 &&
                      ` · ${payment.items
                        .map((item) => t(paymentMethodLabelKeys[item.payment_method] ?? 'paymentMethodOther'))
                        .join(', ')}`}
                  </span>
                  <span>{formatMoney(payment.total_amount)}</span>
                </li>
              ))}
            </ul>
          </section>
        )}

        {doc.promptpay_payload && (
          <section className="flex flex-col items-center gap-2 self-center rounded-md border border-gray-300 p-4 text-center">
            <p className="text-sm font-semibold">{t('invoicePrintPromptPayTitle')}</p>
            <QRCodeSVG value={doc.promptpay_payload} size={176} marginSize={2} />
            <p className="text-sm">
              {t('invoicePrintPromptPayAmount')} <span className="font-semibold">{formatMoney(doc.outstanding)}</span>{' '}
              {t('invoicePrintBaht')}
            </p>
            <p className="text-xs text-gray-600">
              PromptPay {dormitory.promptpay_id}
            </p>
          </section>
        )}

        {invoice.note && <p className="text-sm text-gray-600">{invoice.note}</p>}

      </article>
    </div>
  )
}

// Sets the tab title, which browsers also use as the default PDF file name.
function useDocumentTitle(title: string) {
  useEffect(() => {
    if (!title) return
    const previous = document.title
    document.title = title
    return () => {
      document.title = previous
    }
  }, [title])
}
