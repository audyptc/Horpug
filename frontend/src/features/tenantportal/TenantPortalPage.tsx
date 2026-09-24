import { useCallback, useEffect, useState, type FormEvent } from 'react'
import { QRCodeSVG } from 'qrcode.react'
import { Building2, ChevronLeft, FileText, Megaphone, Pin, Printer, Receipt, Wrench } from 'lucide-react'
import { extractErrorCode, extractErrorMessage } from '@/shared/api/client'
import { useLanguage, type TranslationKey } from '@/shared/i18n/language'
import { Badge } from '@/shared/components/ui/badge'
import { Button } from '@/shared/components/ui/button'
import type { ApiInvoiceDocument } from '@/features/invoice/types'
import type { ApiSlip, SlipStatus } from '@/features/payment/slips'
import { ReceiptDocument } from '@/features/payment/ReceiptDocument'
import type { ApiReceipt } from '@/features/payment/types'
import { formatPeriod } from '@/features/invoice/utils'
import { tenantApi, setTenantToken } from './api'
import type {
  PortalAnnouncement,
  PortalInvoice,
  PortalProfile,
  PortalRepair,
  PortalSession,
  RepairCategory,
  RepairStatus,
} from './types'

type Tab = 'invoices' | 'repairs' | 'announcements'

const invoiceStatusKeys: Record<PortalInvoice['status'], TranslationKey> = {
  unpaid: 'invoiceStatusUnpaid',
  paid: 'invoiceStatusPaid',
  overdue: 'invoiceStatusOverdue',
  cancelled: 'invoiceStatusCancelled',
}
const invoiceStatusVariant: Record<PortalInvoice['status'], 'secondary' | 'success' | 'destructive' | 'outline'> = {
  unpaid: 'secondary',
  paid: 'success',
  overdue: 'destructive',
  cancelled: 'outline',
}
const repairCategoryKeys: Record<RepairCategory, TranslationKey> = {
  electrical: 'repairCategoryElectrical',
  plumbing: 'repairCategoryPlumbing',
  furniture: 'repairCategoryFurniture',
  aircon: 'repairCategoryAircon',
  other: 'repairCategoryOther',
}
const repairStatusKeys: Record<RepairStatus, TranslationKey> = {
  pending: 'repairStatusPending',
  in_progress: 'repairStatusInProgress',
  completed: 'repairStatusCompleted',
  cancelled: 'repairStatusCancelled',
}
const REPAIR_CATEGORIES = Object.keys(repairCategoryKeys) as RepairCategory[]
const slipStatusKeys: Record<SlipStatus, TranslationKey> = {
  pending: 'slipStatusPending',
  approved: 'slipStatusApproved',
  rejected: 'slipStatusRejected',
  cancelled: 'slipStatusCancelled',
}

function todayInput(): string {
  const now = new Date()
  return `${now.getFullYear()}-${String(now.getMonth() + 1).padStart(2, '0')}-${String(now.getDate()).padStart(2, '0')}`
}

function money(value: number): string {
  return value.toLocaleString('th-TH', { minimumFractionDigits: 2, maximumFractionDigits: 2 })
}

type Status = { kind: 'loading' } | { kind: 'error'; message: string } | { kind: 'ready'; profile: PortalProfile }

// Tenant self-service pages, opened from the dormitory's LINE OA (the LIFF
// app, see LiffEntryPage). LINE vouches for who the tenant is; the backend
// only lets a linked tenant in and only shows them their own rooms, bills,
// repairs and their dormitory's announcements.
export default function TenantPortalPage() {
  const { t, language } = useLanguage()
  const [status, setStatus] = useState<Status>({ kind: 'loading' })
  const [tab, setTab] = useState<Tab>('invoices')

  useEffect(() => {
    let cancelled = false
    async function signIn() {
      const liffId = import.meta.env.VITE_LIFF_ID as string | undefined
      if (!liffId) {
        setStatus({ kind: 'error', message: t('lineLinkNotConfigured') })
        return
      }
      try {
        const liff = (await import('@line/liff')).default
        await liff.init({ liffId })
        if (!liff.isLoggedIn()) {
          liff.login({ redirectUri: window.location.href })
          return
        }
        const idToken = liff.getIDToken()
        if (!idToken) throw new Error('missing id token')
        const { data } = await tenantApi.post<PortalSession>('/public/tenant-portal/session', { id_token: idToken })
        setTenantToken(data.access_token)
        if (!cancelled) setStatus({ kind: 'ready', profile: data.profile })
      } catch (err) {
        if (cancelled) return
        setStatus({
          kind: 'error',
          message:
            extractErrorCode(err) === 'tenant_not_linked'
              ? t('portalNotLinked')
              : extractErrorMessage(err, t('portalSignInError')),
        })
      }
    }
    signIn()
    return () => {
      cancelled = true
    }
    // eslint-disable-next-line react-hooks/exhaustive-deps
  }, [])

  const dateLocale = language === 'th' ? 'th-TH' : 'en-US'
  const formatDate = useCallback(
    (value: string) =>
      new Date(value).toLocaleDateString(dateLocale, { year: 'numeric', month: 'short', day: 'numeric', timeZone: 'UTC' }),
    [dateLocale]
  )

  return (
    <div className="min-h-screen bg-muted/40 pb-20 text-foreground print:bg-white print:pb-0">
      <header className="sticky top-0 z-10 border-b border-border bg-background px-4 py-3 print:hidden">
        <div className="mx-auto flex max-w-md items-center gap-2">
          <span className="brand-mark" aria-hidden="true">
            <Building2 size={18} strokeWidth={2.4} />
          </span>
          <div className="min-w-0">
            <p className="truncate text-sm font-semibold">
              {status.kind === 'ready' ? `${status.profile.first_name} ${status.profile.last_name}` : 'Horpug'}
            </p>
            {status.kind === 'ready' && status.profile.rooms.length > 0 && (
              <p className="truncate text-xs text-muted-foreground">
                {status.profile.rooms.map((r) => `${t('moveOutRoom')} ${r.room_number} · ${r.dormitory_name}`).join(', ')}
              </p>
            )}
          </div>
        </div>
      </header>

      <main className="mx-auto flex max-w-md flex-col gap-3 px-4 py-4 print:max-w-none print:p-0">
        {status.kind === 'loading' && <p className="metric-detail">{t('loading')}</p>}
        {status.kind === 'error' && <p className="resource-error">{status.message}</p>}
        {status.kind === 'ready' && tab === 'invoices' && <InvoicesTab formatDate={formatDate} />}
        {status.kind === 'ready' && tab === 'repairs' && <RepairsTab profile={status.profile} formatDate={formatDate} />}
        {status.kind === 'ready' && tab === 'announcements' && <AnnouncementsTab formatDate={formatDate} />}
      </main>

      {status.kind === 'ready' && (
        <nav className="fixed inset-x-0 bottom-0 border-t border-border bg-background print:hidden" aria-label={t('portalNav')}>
          <div className="mx-auto grid max-w-md grid-cols-3">
            {(
              [
                ['invoices', Receipt, 'portalTabInvoices'],
                ['repairs', Wrench, 'portalTabRepairs'],
                ['announcements', Megaphone, 'portalTabAnnouncements'],
              ] as const
            ).map(([key, Icon, labelKey]) => (
              <button
                key={key}
                type="button"
                onClick={() => setTab(key)}
                aria-current={tab === key ? 'page' : undefined}
                className={`flex flex-col items-center gap-0.5 py-2 text-xs ${
                  tab === key ? 'font-semibold text-primary' : 'text-muted-foreground'
                }`}
              >
                <Icon size={20} />
                {t(labelKey)}
              </button>
            ))}
          </div>
        </nav>
      )}
    </div>
  )
}

function useLoad<T>(path: string | null) {
  const { t } = useLanguage()
  const [state, setState] = useState<{ path: string; data?: T; error?: string } | null>(null)
  const [version, setVersion] = useState(0)

  useEffect(() => {
    if (!path) return
    const controller = new AbortController()
    tenantApi
      .get<T>(path, { signal: controller.signal })
      .then(({ data }) => setState({ path, data }))
      .catch((err) => {
        if (controller.signal.aborted) return
        setState({ path, error: extractErrorMessage(err, t('resourceLoadError')) })
      })
    return () => controller.abort()
    // eslint-disable-next-line react-hooks/exhaustive-deps
  }, [path, version])

  const current = state && state.path === path ? state : null
  return { data: current?.data, error: current?.error, loading: !!path && !current, reload: () => setVersion((v) => v + 1) }
}

function InvoicesTab({ formatDate }: { formatDate: (value: string) => string }) {
  const { t } = useLanguage()
  const [openId, setOpenId] = useState<string | null>(null)
  const list = useLoad<PortalInvoice[]>('/tenant/invoices')

  if (openId) return <InvoiceDetail id={openId} onBack={() => setOpenId(null)} formatDate={formatDate} />
  if (list.loading) return <p className="metric-detail">{t('loading')}</p>
  if (list.error) return <p className="resource-error">{list.error}</p>
  if (!list.data || list.data.length === 0) return <p className="metric-detail">{t('portalNoInvoices')}</p>

  return (
    <ul className="flex flex-col gap-2">
      {list.data.map((inv) => (
        <li key={inv.id}>
          <button
            type="button"
            onClick={() => setOpenId(inv.id)}
            className="flex w-full items-center justify-between gap-3 rounded-lg border border-border bg-background p-3 text-left"
          >
            <div className="flex min-w-0 flex-col gap-1">
              <span className="font-medium">
                {t('portalPeriod')} {formatPeriod(inv.period_year, inv.period_month)}
              </span>
              <span className="text-xs text-muted-foreground">
                {inv.invoice_no && `${inv.invoice_no} · `}
                {t('moveOutRoom')} {inv.room_number} · {t('invoiceFormDueDateLabel')} {formatDate(inv.due_date)}
              </span>
            </div>
            <div className="flex shrink-0 flex-col items-end gap-1">
              <span className="font-semibold tabular-nums">{money(inv.outstanding > 0 ? inv.outstanding : inv.total_amount)}</span>
              <Badge variant={invoiceStatusVariant[inv.status]}>{t(invoiceStatusKeys[inv.status])}</Badge>
            </div>
          </button>
        </li>
      ))}
    </ul>
  )
}

function InvoiceDetail({ id, onBack, formatDate }: { id: string; onBack: () => void; formatDate: (v: string) => string }) {
  const { t } = useLanguage()
  const doc = useLoad<ApiInvoiceDocument>(`/tenant/invoices/${id}`)
  const [receiptId, setReceiptId] = useState<string | null>(null)
  const payable = !!doc.data && doc.data.outstanding > 0 && ['unpaid', 'overdue'].includes(doc.data.invoice.status)

  if (receiptId) return <PortalReceipt id={receiptId} onBack={() => setReceiptId(null)} />

  return (
    <section className="flex flex-col gap-3">
      <Button variant="ghost" className="self-start px-0" onClick={onBack}>
        <ChevronLeft />
        {t('portalBack')}
      </Button>
      {doc.loading && <p className="metric-detail">{t('loading')}</p>}
      {doc.error && <p className="resource-error">{doc.error}</p>}
      {doc.data && (
        <div className="flex flex-col gap-3 rounded-lg border border-border bg-background p-4">
          <div>
            <p className="font-semibold">
              {t('portalPeriod')} {formatPeriod(doc.data.invoice.period_year, doc.data.invoice.period_month)}
            </p>
            <p className="text-xs text-muted-foreground">
              {doc.data.invoice.invoice_no && `${doc.data.invoice.invoice_no} · `}
              {doc.data.dormitory.name} · {t('moveOutRoom')} {doc.data.invoice.room_number} ·{' '}
              {t('invoiceFormDueDateLabel')} {formatDate(doc.data.invoice.due_date)}
            </p>
          </div>
          <ul className="divide-y divide-border text-sm">
            {(doc.data.invoice.items ?? []).map((item) => (
              <li key={item.id} className="flex justify-between gap-2 py-1.5">
                <span>{item.description}</span>
                <span className="tabular-nums">{money(item.amount)}</span>
              </li>
            ))}
            <li className="flex justify-between gap-2 py-1.5 font-semibold">
              <span>{t('invoiceFormTotalAmountLabel')}</span>
              <span className="tabular-nums">{money(doc.data.invoice.total_amount)}</span>
            </li>
          </ul>
          {doc.data.payments.length > 0 && (
            <div className="text-sm">
              <p className="font-medium">{t('invoicePrintPaymentsTitle')}</p>
              <ul className="text-xs text-muted-foreground">
                {doc.data.payments.map((p) => (
                  <li key={p.id}>
                    <button
                      type="button"
                      onClick={() => setReceiptId(p.id)}
                      className="flex w-full items-center justify-between gap-2 py-1.5 text-left"
                    >
                      <span className="flex items-center gap-1">
                        <FileText size={14} className="text-primary" />
                        <span className="text-primary underline">{p.receipt_no}</span> · {formatDate(p.payment_date)}
                      </span>
                      <span className="tabular-nums">{money(p.total_amount)}</span>
                    </button>
                  </li>
                ))}
              </ul>
            </div>
          )}
          {doc.data.outstanding > 0 && doc.data.invoice.status !== 'paid' && (
            <p className="flex justify-between text-base font-semibold">
              <span>{t('invoicePrintOutstandingLabel')}</span>
              <span className="tabular-nums">{money(doc.data.outstanding)}</span>
            </p>
          )}
          {doc.data.promptpay_payload && (
            <div className="flex flex-col items-center gap-2 rounded-md border border-border bg-white p-4 text-center text-gray-900">
              <p className="text-sm font-semibold">{t('invoicePrintPromptPayTitle')}</p>
              <QRCodeSVG value={doc.data.promptpay_payload} size={200} marginSize={2} />
              <p className="text-sm">
                {t('invoicePrintPromptPayAmount')} <span className="font-semibold">{money(doc.data.outstanding)}</span>{' '}
                {t('invoicePrintBaht')}
              </p>
              <p className="text-xs text-gray-600">{t('portalQrHint')}</p>
            </div>
          )}
        </div>
      )}
      {doc.data && <SlipSection invoiceId={id} outstanding={doc.data.outstanding} payable={payable} formatDate={formatDate} />}
    </section>
  )
}

// PortalReceipt shows the receipt for one of the tenant's payments. Printing
// hides the portal around it; LINE's in-app browser may not print, so the
// hint suggests a screenshot instead.
function PortalReceipt({ id, onBack }: { id: string; onBack: () => void }) {
  const { t } = useLanguage()
  const receipt = useLoad<ApiReceipt>(`/tenant/payments/${id}/receipt`)

  return (
    <section className="flex flex-col gap-3">
      <div className="flex items-center justify-between gap-2 print:hidden">
        <Button variant="ghost" className="px-0" onClick={onBack}>
          <ChevronLeft />
          {t('portalBack')}
        </Button>
        {receipt.data && (
          <Button variant="outline" onClick={() => window.print()}>
            <Printer className="size-4" />
            {t('invoicePrintAction')}
          </Button>
        )}
      </div>
      {receipt.loading && <p className="metric-detail">{t('loading')}</p>}
      {receipt.error && <p className="resource-error">{receipt.error}</p>}
      {receipt.data && (
        <>
          <div className="overflow-x-auto rounded-lg text-gray-900 print:overflow-visible">
            <ReceiptDocument receipt={receipt.data} />
          </div>
          <p className="text-xs text-muted-foreground print:hidden">{t('portalReceiptHint')}</p>
        </>
      )}
    </section>
  )
}

// SlipSection lets the tenant send the transfer slip for this bill and follow
// what happened to slips they sent. Staff approve or reject them; approval
// records the payment and the tenant gets a LINE message either way.
function SlipSection({
  invoiceId,
  outstanding,
  payable,
  formatDate,
}: {
  invoiceId: string
  outstanding: number
  payable: boolean
  formatDate: (v: string) => string
}) {
  const { t } = useLanguage()
  const slips = useLoad<ApiSlip[]>(`/tenant/invoices/${invoiceId}/slips`)
  const [file, setFile] = useState<File | null>(null)
  const [amount, setAmount] = useState(outstanding > 0 ? String(outstanding) : '')
  const [transferDate, setTransferDate] = useState(todayInput)
  const [note, setNote] = useState('')
  const [sending, setSending] = useState(false)
  const [error, setError] = useState<string | null>(null)
  const [sent, setSent] = useState(false)
  const [formKey, setFormKey] = useState(0)

  const hasPending = (slips.data ?? []).some((s) => s.status === 'pending')

  async function submit(event: FormEvent<HTMLFormElement>) {
    event.preventDefault()
    const value = Number(amount)
    if (!file) {
      setError(t('slipChooseImage'))
      return
    }
    if (!Number.isFinite(value) || value <= 0 || value > outstanding + 0.005) {
      setError(t('slipAmountInvalid'))
      return
    }
    const form = new FormData()
    form.append('file', file)
    form.append('amount', String(value))
    form.append('transfer_date', transferDate)
    form.append('note', note.trim())
    setSending(true)
    setError(null)
    try {
      await tenantApi.post(`/tenant/invoices/${invoiceId}/slips`, form)
      setSent(true)
      setFile(null)
      setNote('')
      setFormKey((k) => k + 1)
      slips.reload()
    } catch (err) {
      const code = extractErrorCode(err)
      setError(
        code === 'unsupported_file_type'
          ? t('slipUnsupportedImage')
          : code === 'file_too_large'
            ? t('slipImageTooLarge')
            : code === 'invalid_amount'
              ? t('slipAmountInvalid')
              : extractErrorMessage(err, t('slipSendError'))
      )
    } finally {
      setSending(false)
    }
  }

  async function cancel(slipId: string) {
    try {
      await tenantApi.post(`/tenant/slips/${slipId}/cancel`)
      slips.reload()
    } catch (err) {
      setError(extractErrorMessage(err, t('slipSendError')))
    }
  }

  return (
    <div className="flex flex-col gap-3">
      {payable && (
        <form key={formKey} className="flex flex-col gap-3 rounded-lg border border-border bg-background p-4" onSubmit={submit}>
          <div>
            <p className="font-semibold">{t('slipSendTitle')}</p>
            <p className="text-xs text-muted-foreground">{t('slipSendHint')}</p>
          </div>
          {hasPending && <p className="rounded-md bg-muted px-3 py-2 text-xs">{t('slipPendingNotice')}</p>}
          <label className="flex flex-col gap-1 text-sm font-medium" htmlFor="slip-file">
            {t('slipImageLabel')}
            <input
              id="slip-file"
              type="file"
              accept="image/jpeg,image/png,image/webp"
              className="text-sm font-normal"
              onChange={(event) => setFile(event.target.files?.[0] ?? null)}
            />
          </label>
          <div className="grid grid-cols-2 gap-2">
            <label className="flex flex-col gap-1 text-sm font-medium" htmlFor="slip-amount">
              {t('slipAmountLabel')}
              <input
                id="slip-amount"
                type="number"
                inputMode="decimal"
                min="0"
                step="0.01"
                className="h-10 rounded-md border border-input bg-transparent px-3 text-right text-sm font-normal"
                value={amount}
                onChange={(event) => setAmount(event.target.value)}
              />
            </label>
            <label className="flex flex-col gap-1 text-sm font-medium" htmlFor="slip-date">
              {t('slipTransferDate')}
              <input
                id="slip-date"
                type="date"
                max={todayInput()}
                className="h-10 rounded-md border border-input bg-transparent px-3 text-sm font-normal"
                value={transferDate}
                onChange={(event) => setTransferDate(event.target.value)}
              />
            </label>
          </div>
          <label className="flex flex-col gap-1 text-sm font-medium" htmlFor="slip-note">
            {t('slipNoteLabel')}
            <input
              id="slip-note"
              maxLength={255}
              className="h-10 rounded-md border border-input bg-transparent px-3 text-sm font-normal"
              value={note}
              onChange={(event) => setNote(event.target.value)}
            />
          </label>
          {error && <p className="resource-error">{error}</p>}
          {sent && !error && <p className="text-sm text-primary">{t('slipSent')}</p>}
          <Button type="submit" disabled={sending}>
            {sending ? t('portalSending') : t('slipSendButton')}
          </Button>
        </form>
      )}

      {slips.data && slips.data.length > 0 && (
        <div className="rounded-lg border border-border bg-background p-4 text-sm">
          <p className="mb-2 font-medium">{t('slipHistoryTitle')}</p>
          <ul className="flex flex-col divide-y divide-border">
            {slips.data.map((s) => (
              <li key={s.id} className="flex flex-col gap-1 py-2">
                <div className="flex items-center justify-between gap-2">
                  <span className="tabular-nums">
                    {money(s.amount)} · {formatDate(s.transfer_date)}
                  </span>
                  <Badge
                    variant={
                      s.status === 'approved' ? 'success' : s.status === 'rejected' ? 'destructive' : s.status === 'pending' ? 'secondary' : 'outline'
                    }
                  >
                    {t(slipStatusKeys[s.status])}
                  </Badge>
                </div>
                {s.status === 'approved' && s.receipt_no && (
                  <span className="text-xs text-muted-foreground">
                    {t('slipReceipt')} {s.receipt_no}
                  </span>
                )}
                {s.status === 'rejected' && s.reject_reason && (
                  <span className="text-xs text-muted-foreground">
                    {t('slipRejectReason')}: {s.reject_reason}
                  </span>
                )}
                {s.status === 'pending' && (
                  <Button size="sm" variant="ghost" className="self-end" onClick={() => cancel(s.id)}>
                    {t('slipCancel')}
                  </Button>
                )}
              </li>
            ))}
          </ul>
        </div>
      )}
    </div>
  )
}

function RepairsTab({ profile, formatDate }: { profile: PortalProfile; formatDate: (v: string) => string }) {
  const { t } = useLanguage()
  const list = useLoad<PortalRepair[]>('/tenant/repair-requests')
  const [roomId, setRoomId] = useState(profile.rooms[0]?.room_id ?? '')
  const [category, setCategory] = useState<RepairCategory>('other')
  const [description, setDescription] = useState('')
  const [saving, setSaving] = useState(false)
  const [error, setError] = useState<string | null>(null)
  const [cancellingId, setCancellingId] = useState<string | null>(null)

  async function submit(event: FormEvent<HTMLFormElement>) {
    event.preventDefault()
    if (!description.trim()) {
      setError(t('portalRepairRequired'))
      return
    }
    setSaving(true)
    setError(null)
    try {
      await tenantApi.post('/tenant/repair-requests', { room_id: roomId, category, description: description.trim() })
      setDescription('')
      setCategory('other')
      list.reload()
    } catch (err) {
      setError(extractErrorMessage(err, t('portalRepairError')))
    } finally {
      setSaving(false)
    }
  }

  async function cancel(id: string) {
    setCancellingId(id)
    try {
      await tenantApi.post(`/tenant/repair-requests/${id}/cancel`)
      list.reload()
    } catch (err) {
      setError(extractErrorMessage(err, t('portalRepairError')))
    } finally {
      setCancellingId(null)
    }
  }

  return (
    <section className="flex flex-col gap-4">
      {profile.rooms.length === 0 ? (
        <p className="metric-detail">{t('portalNoRoom')}</p>
      ) : (
        <form className="flex flex-col gap-3 rounded-lg border border-border bg-background p-4" onSubmit={submit}>
          <p className="font-semibold">{t('portalRepairNew')}</p>
          {profile.rooms.length > 1 && (
            <label className="flex flex-col gap-1 text-sm font-medium">
              {t('moveOutRoom')}
              <select
                className="h-10 rounded-md border border-input bg-transparent px-3 text-sm"
                value={roomId}
                onChange={(event) => setRoomId(event.target.value)}
              >
                {profile.rooms.map((room) => (
                  <option key={room.room_id} value={room.room_id}>
                    {room.room_number} · {room.dormitory_name}
                  </option>
                ))}
              </select>
            </label>
          )}
          <label className="flex flex-col gap-1 text-sm font-medium">
            {t('repairCategoryColumn')}
            <select
              className="h-10 rounded-md border border-input bg-transparent px-3 text-sm"
              value={category}
              onChange={(event) => setCategory(event.target.value as RepairCategory)}
            >
              {REPAIR_CATEGORIES.map((value) => (
                <option key={value} value={value}>
                  {t(repairCategoryKeys[value])}
                </option>
              ))}
            </select>
          </label>
          <label className="flex flex-col gap-1 text-sm font-medium">
            {t('portalRepairDescription')}
            <textarea
              className="min-h-20 rounded-md border border-input bg-transparent px-3 py-2 text-sm font-normal"
              maxLength={255}
              value={description}
              onChange={(event) => setDescription(event.target.value)}
            />
          </label>
          {error && <p className="resource-error">{error}</p>}
          <Button type="submit" disabled={saving}>
            {saving ? t('portalSending') : t('portalRepairSubmit')}
          </Button>
        </form>
      )}

      {list.loading && <p className="metric-detail">{t('loading')}</p>}
      {list.error && <p className="resource-error">{list.error}</p>}
      {list.data && list.data.length > 0 && (
        <ul className="flex flex-col gap-2">
          {list.data.map((r) => (
            <li key={r.id} className="flex flex-col gap-1 rounded-lg border border-border bg-background p-3">
              <div className="flex items-center justify-between gap-2">
                <span className="text-sm font-medium">
                  {t(repairCategoryKeys[r.category])} · {t('moveOutRoom')} {r.room_number}
                </span>
                <Badge variant={r.status === 'completed' ? 'success' : r.status === 'cancelled' ? 'outline' : 'secondary'}>
                  {t(repairStatusKeys[r.status])}
                </Badge>
              </div>
              <p className="text-sm text-muted-foreground">{r.description}</p>
              <div className="flex items-center justify-between gap-2 text-xs text-muted-foreground">
                <span>{formatDate(r.reported_date)}</span>
                {r.status === 'pending' && (
                  <Button size="sm" variant="ghost" disabled={cancellingId === r.id} onClick={() => cancel(r.id)}>
                    {t('portalRepairCancel')}
                  </Button>
                )}
              </div>
            </li>
          ))}
        </ul>
      )}
    </section>
  )
}

function AnnouncementsTab({ formatDate }: { formatDate: (v: string) => string }) {
  const { t } = useLanguage()
  const list = useLoad<PortalAnnouncement[]>('/tenant/announcements')

  if (list.loading) return <p className="metric-detail">{t('loading')}</p>
  if (list.error) return <p className="resource-error">{list.error}</p>
  if (!list.data || list.data.length === 0) return <p className="metric-detail">{t('portalNoAnnouncements')}</p>

  return (
    <ul className="flex flex-col gap-2">
      {list.data.map((a) => (
        <li key={a.id} className="flex flex-col gap-1 rounded-lg border border-border bg-background p-3">
          <p className="flex items-center gap-1.5 font-medium">
            {a.is_pinned && <Pin size={14} aria-label={t('portalPinned')} />}
            {a.title}
          </p>
          <p className="text-xs text-muted-foreground">
            {a.dormitory_name} · {formatDate(a.published_date)}
          </p>
          <p className="whitespace-pre-line text-sm">{a.content}</p>
        </li>
      ))}
    </ul>
  )
}
