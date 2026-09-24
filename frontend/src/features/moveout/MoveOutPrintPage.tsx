import { useEffect, useState } from 'react'
import { useParams } from 'react-router-dom'
import axios from 'axios'
import { Printer } from 'lucide-react'
import { api, extractErrorMessage } from '@/shared/api/client'
import { useLanguage } from '@/shared/i18n/language'
import { Button } from '@/shared/components/ui/button'
import type { ApiMoveOut } from './types'
import { formatBaht } from './utils'

// Printable deposit settlement for a confirmed move-out, for the tenant and
// staff to sign. Opened in its own tab from the contract list.
export default function MoveOutPrintPage() {
  const { id } = useParams<{ id: string }>()
  const { t, language } = useLanguage()
  const [moveOut, setMoveOut] = useState<ApiMoveOut | null>(null)
  const [error, setError] = useState<string | null>(null)

  useEffect(() => {
    if (!id) return
    const controller = new AbortController()
    api
      .get<ApiMoveOut>(`/move-outs/${id}`, { signal: controller.signal })
      .then(({ data }) => setMoveOut(data))
      .catch((err) => {
        if (axios.isCancel(err)) return
        setError(extractErrorMessage(err, t('resourceLoadError')))
      })
    return () => controller.abort()
    // eslint-disable-next-line react-hooks/exhaustive-deps
  }, [id])

  // Browsers use the tab title as the default PDF file name.
  const title = moveOut ? `${t('moveOutPrintTitle')} ${moveOut.room_number}` : ''
  useEffect(() => {
    if (!title) return
    const previous = document.title
    document.title = title
    return () => {
      document.title = previous
    }
  }, [title])

  if (error) {
    return <p className="p-6 text-sm text-red-600">{error}</p>
  }
  if (!moveOut) {
    return <p className="p-6 text-sm text-gray-500">{t('loading')}</p>
  }

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

      <article className="mx-auto flex max-w-[210mm] flex-col gap-6 bg-white p-8 shadow-sm print:max-w-none print:p-0 print:shadow-none">
        <header className="flex flex-col gap-4 border-b border-gray-300 pb-4 sm:flex-row sm:items-start sm:justify-between">
          <div className="flex flex-col gap-1">
            <h1 className="text-xl font-semibold">{moveOut.dormitory_name}</h1>
            {moveOut.dormitory_address && <p className="text-sm text-gray-600">{moveOut.dormitory_address}</p>}
            {moveOut.dormitory_phone && (
              <p className="text-sm text-gray-600">
                {t('invoicePrintPhone')} {moveOut.dormitory_phone}
              </p>
            )}
          </div>
          <div className="flex flex-col gap-1 sm:text-right">
            <h2 className="text-lg font-semibold">{t('moveOutPrintTitle')}</h2>
            <p className="text-sm text-gray-600">
              {t('moveOutDateLabel')}: {formatDate(moveOut.move_out_date)}
            </p>
          </div>
        </header>

        <section className="grid grid-cols-2 gap-2 text-sm">
          <div>
            <span className="text-gray-600">{t('invoiceFormTenantLabel')}: </span>
            <span className="font-medium">{moveOut.tenant_name}</span>
          </div>
          <div className="text-right">
            <span className="text-gray-600">{t('invoiceFormRoomLabel')}: </span>
            <span className="font-medium">{moveOut.room_number}</span>
          </div>
          <div>
            <span className="text-gray-600">{t('moveOutPrintStart')}: </span>
            {formatDate(moveOut.start_date)}
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
            <tr className="border-b border-gray-200">
              <td className="py-2 font-medium">{t('moveOutDeposit')}</td>
              <td className="py-2 text-right font-medium">{formatBaht(moveOut.deposit)}</td>
            </tr>
            {moveOut.items.map((item, index) => (
              <tr key={index} className="border-b border-gray-200">
                <td className="py-2 pl-4">{item.amount < 0 ? `+ ${item.description}` : `− ${item.description}`}</td>
                <td className="py-2 text-right">{formatBaht(-item.amount)}</td>
              </tr>
            ))}
          </tbody>
          <tfoot>
            <tr>
              <td className="pt-3 text-gray-600">{t('moveOutTotalDeductions')}</td>
              <td className="pt-3 text-right text-gray-600">{formatBaht(moveOut.total_deductions)}</td>
            </tr>
            {moveOut.amount_due > 0 ? (
              <tr>
                <td className="pt-1 text-base font-semibold">{t('moveOutAmountDue')}</td>
                <td className="pt-1 text-right text-base font-semibold">{formatBaht(moveOut.amount_due)}</td>
              </tr>
            ) : (
              <tr>
                <td className="pt-1 text-base font-semibold">{t('moveOutRefund')}</td>
                <td className="pt-1 text-right text-base font-semibold">{formatBaht(moveOut.refund_amount)}</td>
              </tr>
            )}
          </tfoot>
        </table>

        {moveOut.deposit_payments.length > 0 && (
          <p className="text-xs text-gray-600">
            {t('moveOutPrintReceipts')}{' '}
            {moveOut.deposit_payments.map((p) => `${p.receipt_no} (${formatBaht(p.amount)})`).join(', ')}
          </p>
        )}

        {moveOut.note && <p className="text-sm text-gray-600">{moveOut.note}</p>}

        <footer className="mt-10 grid grid-cols-2 gap-10 text-sm">
          {[t('moveOutPrintTenantSign'), t('moveOutPrintStaffSign')].map((label) => (
            <div key={label} className="flex flex-col items-center gap-1">
              <div className="h-10 w-full border-b border-gray-400" />
              <span className="text-gray-600">{label}</span>
            </div>
          ))}
        </footer>
      </article>
    </div>
  )
}
