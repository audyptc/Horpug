import { useEffect, useState } from 'react'
import { useParams } from 'react-router-dom'
import axios from 'axios'
import { Printer } from 'lucide-react'
import { api, extractErrorMessage } from '@/shared/api/client'
import { useLanguage } from '@/shared/i18n/language'
import { Button } from '@/shared/components/ui/button'
import { ReceiptDocument } from './ReceiptDocument'
import type { ApiReceipt } from './types'

// Printable receipt for one payment, opened in its own tab from the payment
// list. A voided one still prints, stamped as voided (see ReceiptDocument).
export default function ReceiptPrintPage() {
  const { id } = useParams<{ id: string }>()
  const { t } = useLanguage()
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

  return (
    <div className="min-h-screen bg-gray-100 py-6 text-gray-900 print:bg-white print:py-0">
      <div className="mx-auto mb-4 flex max-w-[210mm] justify-end px-4 print:hidden">
        <Button onClick={() => window.print()}>
          <Printer className="size-4" />
          {t('invoicePrintAction')}
        </Button>
      </div>

      <ReceiptDocument receipt={receipt} />
    </div>
  )
}
