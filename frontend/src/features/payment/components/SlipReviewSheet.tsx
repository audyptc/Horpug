import { useEffect, useState } from 'react'
import axios from 'axios'
import { api, extractErrorCode, extractErrorMessage } from '@/shared/api/client'
import { useLanguage } from '@/shared/i18n/language'
import { Button } from '@/shared/components/ui/button'
import { DatePickerField } from '@/shared/components/date-picker-field'
import {
  Sheet,
  SheetContent,
  SheetDescription,
  SheetHeader,
  SheetTitle,
} from '@/shared/components/ui/sheet'
import { formatPeriod, toApiDate, toDateInputValue } from '@/features/invoice/utils'
import type { ApiSlip } from '../slips'

function money(value: number): string {
  return value.toLocaleString('th-TH', { minimumFractionDigits: 2, maximumFractionDigits: 2 })
}

// The slip image is behind the staff API (bearer token), so it can't be an
// <img src>; it's fetched as a blob and shown from an object URL.
function SlipImage({ slipId, alt }: { slipId: string; alt: string }) {
  const [url, setUrl] = useState<string | null>(null)
  const [failed, setFailed] = useState(false)

  useEffect(() => {
    const controller = new AbortController()
    let objectUrl: string | null = null
    api
      .get<Blob>(`/payment-slips/${slipId}/image`, { responseType: 'blob', signal: controller.signal })
      .then(({ data }) => {
        objectUrl = URL.createObjectURL(data)
        setUrl(objectUrl)
      })
      .catch((err) => {
        if (!axios.isCancel(err)) setFailed(true)
      })
    return () => {
      controller.abort()
      if (objectUrl) URL.revokeObjectURL(objectUrl)
    }
  }, [slipId])

  if (failed) return <p className="resource-error">—</p>
  if (!url) return <div className="h-48 w-full animate-pulse rounded-md bg-muted" />
  return (
    <a href={url} target="_blank" rel="noreferrer" className="block">
      <img src={url} alt={alt} className="max-h-80 w-full rounded-md border border-border object-contain" />
    </a>
  )
}

type SlipCardProps = {
  slip: ApiSlip
  onDone: () => void
}

function SlipCard({ slip, onDone }: SlipCardProps) {
  const { t } = useLanguage()
  const [amount, setAmount] = useState(String(slip.amount))
  const [paymentDate, setPaymentDate] = useState(toDateInputValue(slip.transfer_date))
  const [referenceNo, setReferenceNo] = useState('')
  const [rejecting, setRejecting] = useState(false)
  const [reason, setReason] = useState('')
  const [busy, setBusy] = useState(false)
  const [error, setError] = useState<string | null>(null)

  async function approve() {
    const value = Number(amount)
    if (!Number.isFinite(value) || value <= 0 || !paymentDate) {
      setError(t('slipApproveInvalid'))
      return
    }
    setBusy(true)
    setError(null)
    try {
      await api.post(`/payment-slips/${slip.id}/approve`, {
        amount: value,
        payment_date: toApiDate(paymentDate),
        reference_no: referenceNo.trim(),
      })
      onDone()
    } catch (err) {
      const code = extractErrorCode(err)
      setError(
        code === 'payment_exceeds_invoice'
          ? t('paymentExceedsInvoice')
          : code === 'slip_not_pending'
            ? t('slipAlreadyReviewed')
            : extractErrorMessage(err, t('slipReviewError'))
      )
    } finally {
      setBusy(false)
    }
  }

  async function reject() {
    if (!reason.trim()) {
      setError(t('slipRejectReasonRequired'))
      return
    }
    setBusy(true)
    setError(null)
    try {
      await api.post(`/payment-slips/${slip.id}/reject`, { reason: reason.trim() })
      onDone()
    } catch (err) {
      setError(
        extractErrorCode(err) === 'slip_not_pending'
          ? t('slipAlreadyReviewed')
          : extractErrorMessage(err, t('slipReviewError'))
      )
    } finally {
      setBusy(false)
    }
  }

  const mismatch = Math.abs(slip.amount - slip.outstanding) > 0.005

  return (
    <li className="flex flex-col gap-3 rounded-lg border border-border p-3">
      <div className="flex flex-wrap items-start justify-between gap-2 text-sm">
        <div>
          <p className="font-semibold">
            {slip.tenant_name} · {t('moveOutRoom')} {slip.room_number}
          </p>
          <p className="text-xs text-muted-foreground">
            {slip.dormitory_name} · {t('portalPeriod')} {formatPeriod(slip.period_year, slip.period_month)} ·{' '}
            {t('slipSentAt')} {new Date(slip.created_at).toLocaleString('th-TH')}
          </p>
        </div>
        <div className="text-right">
          <p className="font-semibold tabular-nums">{money(slip.amount)}</p>
          <p className={`text-xs ${mismatch ? 'text-amber-700 dark:text-amber-400' : 'text-muted-foreground'}`}>
            {t('slipOutstanding')} {money(slip.outstanding)}
            {mismatch ? ` · ${t('slipAmountDiffers')}` : ''}
          </p>
        </div>
      </div>

      <SlipImage slipId={slip.id} alt={`${t('slipImageAlt')} ${slip.tenant_name}`} />

      <p className="text-xs text-muted-foreground">
        {t('slipTransferDate')} {toDateInputValue(slip.transfer_date)}
        {slip.note ? ` · ${slip.note}` : ''}
      </p>

      {rejecting ? (
        <div className="flex flex-col gap-2">
          <label className="flex flex-col gap-1 text-sm font-medium" htmlFor={`reason-${slip.id}`}>
            {t('slipRejectReason')}
            <input
              id={`reason-${slip.id}`}
              className="h-9 rounded-md border border-input bg-transparent px-2.5 text-sm font-normal"
              placeholder={t('slipRejectReasonPlaceholder')}
              value={reason}
              onChange={(event) => setReason(event.target.value)}
            />
          </label>
          <div className="flex justify-end gap-2">
            <Button type="button" variant="outline" size="sm" onClick={() => setRejecting(false)} disabled={busy}>
              {t('cancel')}
            </Button>
            <Button type="button" variant="destructive" size="sm" onClick={reject} disabled={busy}>
              {t('slipReject')}
            </Button>
          </div>
        </div>
      ) : (
        <div className="flex flex-col gap-2">
          <div className="grid grid-cols-1 gap-2 sm:grid-cols-3">
            <label className="flex flex-col gap-1 text-xs font-medium" htmlFor={`amount-${slip.id}`}>
              {t('paymentFormAmountLabel')}
              <input
                id={`amount-${slip.id}`}
                type="number"
                min="0"
                step="0.01"
                className="h-9 rounded-md border border-input bg-transparent px-2.5 text-right text-sm font-normal"
                value={amount}
                onChange={(event) => setAmount(event.target.value)}
              />
            </label>
            <label className="flex flex-col gap-1 text-xs font-medium">
              {t('paymentFormDateLabel')}
              <DatePickerField value={paymentDate} onChange={setPaymentDate} placeholder={t('paymentFormDateLabel')} />
            </label>
            <label className="flex flex-col gap-1 text-xs font-medium" htmlFor={`ref-${slip.id}`}>
              {t('paymentFormReferenceLabel')}
              <input
                id={`ref-${slip.id}`}
                className="h-9 rounded-md border border-input bg-transparent px-2.5 text-sm font-normal"
                value={referenceNo}
                onChange={(event) => setReferenceNo(event.target.value)}
              />
            </label>
          </div>
          <div className="flex justify-end gap-2">
            <Button type="button" variant="outline" size="sm" onClick={() => setRejecting(true)} disabled={busy}>
              {t('slipReject')}
            </Button>
            <Button type="button" size="sm" onClick={approve} disabled={busy}>
              {t('slipApprove')}
            </Button>
          </div>
        </div>
      )}

      {error && <p className="resource-error">{error}</p>}
    </li>
  )
}

type SlipReviewSheetProps = {
  open: boolean
  onOpenChange: (open: boolean) => void
  // Called after a slip is approved or rejected, so the payment list and
  // the pending count refresh.
  onReviewed: () => void
}

// Queue of transfer slips tenants sent from LINE, oldest first. Approving
// records a transfer payment (with a receipt number) for the slip's invoice;
// rejecting sends the tenant the reason. Either way the tenant is told on LINE.
export function SlipReviewSheet({ open, onOpenChange, onReviewed }: SlipReviewSheetProps) {
  const { t } = useLanguage()
  const [slips, setSlips] = useState<{ version: number; data: ApiSlip[] } | null>(null)
  const [error, setError] = useState<string | null>(null)
  const [version, setVersion] = useState(0)

  useEffect(() => {
    if (!open) return
    const controller = new AbortController()
    api
      .get<ApiSlip[]>('/payment-slips', { params: { status: 'pending' }, signal: controller.signal })
      .then(({ data }) => setSlips({ version, data }))
      .catch((err) => {
        if (!axios.isCancel(err)) setError(extractErrorMessage(err, t('resourceLoadError')))
      })
    return () => controller.abort()
    // eslint-disable-next-line react-hooks/exhaustive-deps
  }, [open, version])

  const current = slips && slips.version === version ? slips.data : null

  return (
    <Sheet open={open} onOpenChange={onOpenChange}>
      <SheetContent className="sm:max-w-xl">
        <SheetHeader>
          <SheetTitle>{t('slipReviewTitle')}</SheetTitle>
          <SheetDescription>{t('slipReviewDescription')}</SheetDescription>
        </SheetHeader>
        <div className="flex flex-1 flex-col gap-3 overflow-y-auto pr-1">
          {error ? (
            <p className="resource-error">{error}</p>
          ) : !current ? (
            <p className="metric-detail">{t('loading')}</p>
          ) : current.length === 0 ? (
            <p className="metric-detail">{t('slipNonePending')}</p>
          ) : (
            <ul className="flex flex-col gap-3">
              {current.map((slip) => (
                <SlipCard
                  key={slip.id}
                  slip={slip}
                  onDone={() => {
                    setVersion((v) => v + 1)
                    onReviewed()
                  }}
                />
              ))}
            </ul>
          )}
        </div>
      </SheetContent>
    </Sheet>
  )
}
