import { useEffect, useState, type FormEvent } from 'react'
import axios from 'axios'
import { Plus, X } from 'lucide-react'
import { api, extractErrorMessage } from '@/shared/api/client'
import { useLanguage } from '@/shared/i18n/language'
import { Button } from '@/shared/components/ui/button'
import { ConfirmDialog } from '@/shared/components/confirm-dialog'
import { DatePickerField } from '@/shared/components/date-picker-field'
import {
  Sheet,
  SheetContent,
  SheetDescription,
  SheetFooter,
  SheetHeader,
  SheetTitle,
} from '@/shared/components/ui/sheet'
import type { ApiContract } from '@/features/contract/types'
import { toApiDate } from '@/features/invoice/utils'
import type { ApiMoveOut, ApiMoveOutPreview } from './types'
import { formatBaht } from './utils'

const PREVIEW_DEBOUNCE_MS = 400

type Deduction = { key: number; description: string; amount: string }

function todayInputValue(): string {
  const now = new Date()
  const month = String(now.getMonth() + 1).padStart(2, '0')
  const day = String(now.getDate()).padStart(2, '0')
  return `${now.getFullYear()}-${month}-${day}`
}

type MoveOutSheetProps = {
  open: boolean
  onOpenChange: (open: boolean) => void
  contract: ApiContract
  // Called once the move-out is confirmed.
  onMovedOut: (moveOut: ApiMoveOut) => void
}

// Settles a tenant's move-out. The server works out the deposit reckoning
// (unpaid invoices, rent for the days stayed, unbilled meter readings) and
// staff add deductions, from the dormitory's presets or typed in; the preview
// refreshes as they edit. The parent remounts this (via key) for each
// contract, so every opening starts fresh.
export function MoveOutSheet({ open, onOpenChange, contract, onMovedOut }: MoveOutSheetProps) {
  const { t } = useLanguage()

  const [moveOutDate, setMoveOutDate] = useState(todayInputValue)
  const [note, setNote] = useState('')
  const [deductions, setDeductions] = useState<Deduction[]>([])
  const [nextKey, setNextKey] = useState(1)

  // Stored with the request it answers, so a stale preview reads as loading.
  const [preview, setPreview] = useState<{ key: string; data: ApiMoveOutPreview } | null>(null)
  const [previewError, setPreviewError] = useState<{ key: string; message: string } | null>(null)
  const [confirmOpen, setConfirmOpen] = useState(false)
  const [saving, setSaving] = useState(false)
  const [error, setError] = useState<string | null>(null)

  // Only complete lines go to the server; half-typed ones would be rejected.
  const validDeductions = deductions
    .map((d) => ({ description: d.description.trim(), amount: Number(d.amount) }))
    .filter((d) => d.description && Number.isFinite(d.amount) && d.amount > 0)
  const hasIncompleteDeduction = validDeductions.length !== deductions.length

  const requestBody = moveOutDate
    ? { move_out_date: toApiDate(moveOutDate), note: note.trim(), deductions: validDeductions }
    : null
  const requestKey = requestBody ? JSON.stringify({ d: requestBody.move_out_date, x: requestBody.deductions }) : ''

  useEffect(() => {
    if (!open || !requestKey || !requestBody) return

    const controller = new AbortController()
    const timer = window.setTimeout(() => {
      api
        .post<ApiMoveOutPreview>(`/contracts/${contract.id}/move-out/preview`, requestBody, {
          signal: controller.signal,
        })
        .then(({ data }) => setPreview({ key: requestKey, data }))
        .catch((err) => {
          if (axios.isCancel(err)) return
          setPreviewError({ key: requestKey, message: extractErrorMessage(err, t('moveOutPreviewError')) })
        })
    }, PREVIEW_DEBOUNCE_MS)

    return () => {
      window.clearTimeout(timer)
      controller.abort()
    }
    // eslint-disable-next-line react-hooks/exhaustive-deps
  }, [open, requestKey, contract.id])

  const current = preview && preview.key === requestKey ? preview.data : null
  const currentError = previewError && previewError.key === requestKey ? previewError.message : null
  const previewLoading = requestKey !== '' && !current && !currentError
  // Presets outlive a date change, so keep offering the last ones seen.
  const presets = (current ?? preview?.data)?.presets ?? []
  const autoItems = (current?.items ?? []).filter((item) => item.item_type !== 'other')

  function addDeduction(description = '', amount = '') {
    setDeductions((prev) => [...prev, { key: nextKey, description, amount }])
    setNextKey((value) => value + 1)
  }

  function updateDeduction(key: number, patch: Partial<Deduction>) {
    setDeductions((prev) => prev.map((d) => (d.key === key ? { ...d, ...patch } : d)))
  }

  function removeDeduction(key: number) {
    setDeductions((prev) => prev.filter((d) => d.key !== key))
  }

  function handleSubmit(event: FormEvent<HTMLFormElement>) {
    event.preventDefault()
    if (!current || hasIncompleteDeduction) {
      setError(t('moveOutIncompleteError'))
      return
    }
    setError(null)
    setConfirmOpen(true)
  }

  async function handleConfirm() {
    if (!requestBody) return
    setSaving(true)
    setError(null)
    try {
      const { data } = await api.post<ApiMoveOut>(`/contracts/${contract.id}/move-out`, requestBody)
      setConfirmOpen(false)
      onMovedOut(data)
    } catch (err) {
      setError(extractErrorMessage(err, t('moveOutConfirmError')))
    } finally {
      setSaving(false)
    }
  }

  return (
    <Sheet open={open} onOpenChange={onOpenChange}>
      <SheetContent>
        <form className="flex h-full flex-col gap-4" onSubmit={handleSubmit}>
          <SheetHeader>
            <SheetTitle>{t('moveOutTitle')}</SheetTitle>
            <SheetDescription>
              {contract.tenant_name ?? '—'} · {t('moveOutRoom')} {contract.room_number ?? '—'}
              {contract.dormitory_name ? ` (${contract.dormitory_name})` : ''}
            </SheetDescription>
          </SheetHeader>

          <div className="flex flex-1 flex-col gap-5 overflow-y-auto pr-1">
            <label className="flex flex-col gap-1.5 text-sm font-medium">
              {t('moveOutDateLabel')}
              <DatePickerField value={moveOutDate} onChange={setMoveOutDate} placeholder={t('moveOutDateLabel')} />
            </label>
            <p className="rounded-md bg-muted px-3 py-2 text-xs text-muted-foreground">{t('moveOutMeterHint')}</p>

            <section className="flex flex-col gap-2">
              <h3 className="text-xs font-semibold uppercase tracking-wide text-muted-foreground">
                {t('moveOutAutoItemsLabel')}
              </h3>
              {previewLoading ? (
                <p className="metric-detail">{t('loading')}</p>
              ) : currentError ? (
                <p className="resource-error">{currentError}</p>
              ) : autoItems.length === 0 ? (
                <p className="text-sm text-muted-foreground">{t('moveOutNoAutoItems')}</p>
              ) : (
                <ul className="divide-y divide-border rounded-md border border-input text-sm">
                  {autoItems.map((item, index) => (
                    <li key={index} className="flex items-center justify-between gap-2 px-3 py-2">
                      <span>{item.description}</span>
                      <span className={item.amount < 0 ? 'text-green-700' : 'text-muted-foreground'}>
                        {formatBaht(item.amount)}
                      </span>
                    </li>
                  ))}
                </ul>
              )}
            </section>

            <section className="flex flex-col gap-2">
              <div className="flex items-center justify-between gap-2">
                <h3 className="text-xs font-semibold uppercase tracking-wide text-muted-foreground">
                  {t('moveOutDeductionsLabel')}
                </h3>
                <Button type="button" size="sm" variant="outline" onClick={() => addDeduction()}>
                  <Plus />
                  {t('moveOutAddDeduction')}
                </Button>
              </div>
              {presets.length > 0 && (
                <div className="flex flex-wrap gap-1.5">
                  {presets.map((preset) => (
                    <Button
                      key={preset.id}
                      type="button"
                      size="sm"
                      variant="secondary"
                      onClick={() => addDeduction(preset.name, String(preset.amount))}
                    >
                      <Plus />
                      {preset.name} {formatBaht(preset.amount)}
                    </Button>
                  ))}
                </div>
              )}
              {deductions.map((d) => (
                <div key={d.key} className="flex items-center gap-1.5">
                  <input
                    type="text"
                    className="h-9 min-w-0 flex-1 rounded-md border border-input bg-transparent px-2.5 text-sm"
                    placeholder={t('moveOutDeductionDescription')}
                    value={d.description}
                    onChange={(event) => updateDeduction(d.key, { description: event.target.value })}
                  />
                  <input
                    type="number"
                    min="0"
                    step="0.01"
                    className="h-9 w-28 shrink-0 rounded-md border border-input bg-transparent px-2.5 text-right text-sm"
                    placeholder={t('moveOutDeductionAmount')}
                    value={d.amount}
                    onChange={(event) => updateDeduction(d.key, { amount: event.target.value })}
                  />
                  <Button
                    type="button"
                    size="icon"
                    variant="ghost"
                    className="h-8 w-8 shrink-0 text-muted-foreground"
                    title={t('moveOutRemoveDeduction')}
                    aria-label={t('moveOutRemoveDeduction')}
                    onClick={() => removeDeduction(d.key)}
                  >
                    <X className="size-4" />
                  </Button>
                </div>
              ))}
            </section>

            <label className="flex flex-col gap-1.5 text-sm font-medium">
              {t('moveOutNoteLabel')}
              <textarea
                className="min-h-16 rounded-md border border-input bg-transparent px-3 py-2 text-sm font-normal"
                value={note}
                onChange={(event) => setNote(event.target.value)}
              />
            </label>

            {current && (
              <section className="rounded-md border border-input text-sm">
                <div className="flex justify-between gap-2 px-3 py-2">
                  <span>{t('moveOutDeposit')}</span>
                  <span>{formatBaht(current.deposit)}</span>
                </div>
                <div className="flex justify-between gap-2 border-t border-border px-3 py-2">
                  <span>{t('moveOutTotalDeductions')}</span>
                  <span>{formatBaht(current.total_deductions)}</span>
                </div>
                {current.amount_due > 0 ? (
                  <div className="flex justify-between gap-2 border-t border-border px-3 py-2 font-semibold text-red-600">
                    <span>{t('moveOutAmountDue')}</span>
                    <span>{formatBaht(current.amount_due)}</span>
                  </div>
                ) : (
                  <div className="flex justify-between gap-2 border-t border-border px-3 py-2 font-semibold">
                    <span>{t('moveOutRefund')}</span>
                    <span>{formatBaht(current.refund_amount)}</span>
                  </div>
                )}
              </section>
            )}
          </div>

          {error && <p className="resource-error">{error}</p>}

          <SheetFooter>
            <Button type="button" variant="outline" onClick={() => onOpenChange(false)}>
              {t('cancel')}
            </Button>
            <Button type="submit" disabled={saving || !current || hasIncompleteDeduction}>
              {t('moveOutSubmit')}
            </Button>
          </SheetFooter>
        </form>

        <ConfirmDialog
          open={confirmOpen}
          onOpenChange={(value) => !saving && setConfirmOpen(value)}
          title={t('moveOutConfirmTitle')}
          description={
            current
              ? (current.amount_due > 0 ? t('moveOutConfirmDue') : t('moveOutConfirmRefund'))
                  .replace('{tenant}', current.tenant_name)
                  .replace('{room}', current.room_number)
                  .replace('{amount}', formatBaht(current.amount_due > 0 ? current.amount_due : current.refund_amount))
              : ''
          }
          confirmLabel={t('moveOutSubmit')}
          cancelLabel={t('cancel')}
          loading={saving}
          onConfirm={handleConfirm}
        />
      </SheetContent>
    </Sheet>
  )
}
