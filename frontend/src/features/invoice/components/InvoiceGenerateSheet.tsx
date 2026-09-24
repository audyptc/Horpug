import { useEffect, useState, type FormEvent } from 'react'
import axios from 'axios'
import { api, extractErrorMessage } from '@/shared/api/client'
import { useLanguage } from '@/shared/i18n/language'
import { Badge } from '@/shared/components/ui/badge'
import { Button } from '@/shared/components/ui/button'
import { DatePickerField } from '@/shared/components/date-picker-field'
import {
  Sheet,
  SheetContent,
  SheetDescription,
  SheetFooter,
  SheetHeader,
  SheetTitle,
} from '@/shared/components/ui/sheet'
import type { ApiDormitory } from '@/features/dormitory/types'
import { DormitorySearchSelect } from '@/features/dormitory/components/DormitorySearchSelect'
import type { ApiGenerationCandidate, ApiGenerationResult } from '../types'
import { parsePeriodInputValue, toApiDate, toPeriodInputValue } from '../utils'
import { FormSection, MonthPickerField } from './InvoiceFormSheet'

function todayInputValue(): string {
  const now = new Date()
  const month = String(now.getMonth() + 1).padStart(2, '0')
  const day = String(now.getDate()).padStart(2, '0')
  return `${now.getFullYear()}-${month}-${day}`
}

// The parent remounts this (via key) each time it opens, so every run starts
// from fresh state.
type InvoiceGenerateSheetProps = {
  open: boolean
  onOpenChange: (open: boolean) => void
  // Called after a run that created at least one invoice, so the list refetches.
  onGenerated: () => void
}

// Bills every active contract of one dormitory for a period in one go. The
// preview lists each contract with what's missing (meter readings, a passed
// end date) so staff can record readings first or untick rooms before billing.
export function InvoiceGenerateSheet({ open, onOpenChange, onGenerated }: InvoiceGenerateSheetProps) {
  const { t } = useLanguage()

  const [dormitory, setDormitory] = useState<ApiDormitory | null>(null)
  const [period, setPeriod] = useState(() => {
    const now = new Date()
    return toPeriodInputValue(now.getFullYear(), now.getMonth() + 1)
  })
  const [issueDate, setIssueDate] = useState(todayInputValue)
  const [dueDate, setDueDate] = useState('')
  const [note, setNote] = useState('')

  // The preview is stored with the dormitory/period it was fetched for, so a
  // stale one (or none yet) reads as loading without resetting state inside
  // the fetch effect.
  const [preview, setPreview] = useState<{ key: string; candidates: ApiGenerationCandidate[] } | null>(null)
  const [selected, setSelected] = useState<Set<string>>(new Set())

  const [saving, setSaving] = useState(false)
  const [error, setError] = useState<string | null>(null)
  const [result, setResult] = useState<ApiGenerationResult | null>(null)

  const parsedPeriod = parsePeriodInputValue(period)
  const dormitoryId = dormitory?.id ?? ''
  const periodYear = parsedPeriod?.year
  const periodMonth = parsedPeriod?.month
  const previewKey = dormitoryId && periodYear && periodMonth ? `${dormitoryId}|${periodYear}|${periodMonth}` : ''
  const candidates = previewKey && preview?.key === previewKey ? preview.candidates : null
  const previewLoading = previewKey !== '' && candidates === null && error === null

  useEffect(() => {
    if (!previewKey) return

    const controller = new AbortController()
    api
      .get<ApiGenerationCandidate[]>('/invoices/generate/preview', {
        signal: controller.signal,
        params: { dormitory_id: dormitoryId, period_year: periodYear, period_month: periodMonth },
      })
      .then(({ data }) => {
        setPreview({ key: previewKey, candidates: data })
        setSelected(new Set(data.filter((c) => !c.already_invoiced).map((c) => c.contract_id)))
      })
      .catch((err) => {
        if (axios.isCancel(err)) return
        setError(extractErrorMessage(err, t('resourceLoadError')))
      })

    return () => controller.abort()
    // eslint-disable-next-line react-hooks/exhaustive-deps
  }, [previewKey])

  const billable = (candidates ?? []).filter((c) => !c.already_invoiced)
  const allSelected = billable.length > 0 && billable.every((c) => selected.has(c.contract_id))
  const missingReadings = billable.filter(
    (c) => selected.has(c.contract_id) && (!c.has_electricity || !c.has_water)
  ).length

  function toggle(contractId: string) {
    setSelected((prev) => {
      const next = new Set(prev)
      if (next.has(contractId)) next.delete(contractId)
      else next.add(contractId)
      return next
    })
  }

  function toggleAll() {
    setSelected(allSelected ? new Set() : new Set(billable.map((c) => c.contract_id)))
  }

  async function handleSubmit(event: FormEvent<HTMLFormElement>) {
    event.preventDefault()
    if (!dormitory || !parsedPeriod || !issueDate || !dueDate) {
      setError(t('invoiceGenerateRequiredError'))
      return
    }
    if (selected.size === 0) {
      setError(t('invoiceGenerateNothingSelected'))
      return
    }

    setSaving(true)
    setError(null)
    try {
      const { data } = await api.post<ApiGenerationResult>('/invoices/generate', {
        dormitory_id: dormitory.id,
        period_year: parsedPeriod.year,
        period_month: parsedPeriod.month,
        issue_date: toApiDate(issueDate),
        due_date: toApiDate(dueDate),
        note: note.trim(),
        contract_ids: Array.from(selected),
      })
      setResult(data)
      if (data.created.length > 0) onGenerated()
    } catch (err) {
      setError(extractErrorMessage(err, t('invoiceGenerateError')))
    } finally {
      setSaving(false)
    }
  }

  return (
    <Sheet open={open} onOpenChange={onOpenChange}>
      <SheetContent>
        <form className="flex h-full flex-col gap-4" onSubmit={handleSubmit}>
          <SheetHeader>
            <SheetTitle>{t('invoiceGenerateTitle')}</SheetTitle>
            <SheetDescription>{t('invoiceGenerateDescription')}</SheetDescription>
          </SheetHeader>

          <div className="flex flex-1 flex-col gap-6 overflow-y-auto pr-1">
            {result ? (
              <FormSection title={t('invoiceGenerateResultTitle')}>
                <div className="rounded-md border border-input">
                  <ul className="divide-y divide-border text-sm font-normal">
                    <li className="flex items-center justify-between gap-2 px-3 py-2">
                      <span>{t('invoiceGenerateResultCreated')}</span>
                      <span className="font-medium">{result.created.length}</span>
                    </li>
                    <li className="flex items-center justify-between gap-2 px-3 py-2">
                      <span>{t('invoiceGenerateResultSkipped')}</span>
                      <span className="font-medium">{result.skipped}</span>
                    </li>
                    <li className="flex items-center justify-between gap-2 px-3 py-2">
                      <span>{t('invoiceGenerateResultFailed')}</span>
                      <span className="font-medium">{result.failed.length}</span>
                    </li>
                  </ul>
                </div>
                {result.created.length > 0 && (
                  <div className="rounded-md border border-input">
                    <ul className="divide-y divide-border text-sm font-normal">
                      {result.created.map((item) => (
                        <li key={item.invoice_id} className="flex items-center justify-between gap-2 px-3 py-2">
                          <span>
                            {item.room_number} · {item.tenant_name}
                          </span>
                          <span className="text-muted-foreground">{item.total_amount.toLocaleString()}</span>
                        </li>
                      ))}
                    </ul>
                  </div>
                )}
                {result.failed.length > 0 && (
                  <ul className="flex flex-col gap-1 text-sm">
                    {result.failed.map((item) => (
                      <li key={item.contract_id} className="resource-error">
                        {item.room_number} · {item.tenant_name}: {item.error}
                      </li>
                    ))}
                  </ul>
                )}
              </FormSection>
            ) : (
              <>
                <FormSection title={t('invoiceFormSectionBilling')}>
                  <label className="flex flex-col gap-1.5 text-sm font-medium">
                    {t('invoiceGenerateDormitoryLabel')}
                    <DormitorySearchSelect
                      selectedLabel={dormitory?.name ?? ''}
                      onSelectDormitory={(value) => {
                        setDormitory(value)
                        setError(null)
                      }}
                      placeholder={t('invoiceGenerateDormitoryPlaceholder')}
                      searchPlaceholder={t('invoiceGenerateDormitoryPlaceholder')}
                      noResultsLabel={t('invoiceGenerateDormitoryNoResults')}
                    />
                  </label>

                  <label className="flex flex-col gap-1.5 text-sm font-medium">
                    {t('invoiceFormPeriodLabel')}
                    <MonthPickerField
                      value={period}
                      onChange={(value) => {
                        setPeriod(value)
                        setError(null)
                      }}
                      placeholder={t('invoiceFormPeriodLabel')}
                    />
                  </label>

                  <div className="grid grid-cols-1 gap-3 sm:grid-cols-2">
                    <label className="flex flex-col gap-1.5 text-sm font-medium">
                      {t('invoiceFormIssueDateLabel')}
                      <DatePickerField
                        value={issueDate}
                        onChange={setIssueDate}
                        placeholder={t('invoiceFormIssueDateLabel')}
                      />
                    </label>
                    <label className="flex flex-col gap-1.5 text-sm font-medium">
                      {t('invoiceFormDueDateLabel')}
                      <DatePickerField value={dueDate} onChange={setDueDate} placeholder={t('invoiceFormDueDateLabel')} />
                    </label>
                  </div>
                </FormSection>

                {dormitory && parsedPeriod && (
                  <FormSection title={t('invoiceGenerateContractsLabel')}>
                    {previewLoading ? (
                      <p className="metric-detail">{t('loading')}</p>
                    ) : candidates && candidates.length === 0 ? (
                      <p className="text-sm font-normal text-muted-foreground">{t('invoiceGenerateNoContracts')}</p>
                    ) : candidates ? (
                      <>
                        {billable.length > 0 && (
                          <label className="flex items-center gap-2 text-sm font-medium">
                            <input type="checkbox" checked={allSelected} onChange={toggleAll} />
                            {t('invoiceGenerateSelectAll')} ({selected.size}/{billable.length})
                          </label>
                        )}
                        <div className="rounded-md border border-input">
                          <ul className="divide-y divide-border text-sm font-normal">
                            {candidates.map((c) => (
                              <li key={c.contract_id} className="flex items-start gap-2 px-3 py-2">
                                <input
                                  type="checkbox"
                                  className="mt-1"
                                  checked={selected.has(c.contract_id)}
                                  disabled={c.already_invoiced}
                                  onChange={() => toggle(c.contract_id)}
                                  aria-label={`${c.room_number} ${c.tenant_name}`}
                                />
                                <div className="flex min-w-0 flex-1 flex-col gap-1">
                                  <div className="flex items-center justify-between gap-2">
                                    <span className="truncate">
                                      <span className="font-medium">{c.room_number}</span> · {c.tenant_name}
                                    </span>
                                    <span className="shrink-0 text-muted-foreground">
                                      {c.rent_price.toLocaleString()}
                                    </span>
                                  </div>
                                  <div className="flex flex-wrap gap-1">
                                    {c.already_invoiced && (
                                      <Badge variant="secondary">{t('invoiceGenerateBadgeInvoiced')}</Badge>
                                    )}
                                    {!c.already_invoiced && !c.has_electricity && (
                                      <Badge variant="warning">{t('invoiceGenerateBadgeNoElectricity')}</Badge>
                                    )}
                                    {!c.already_invoiced && !c.has_water && (
                                      <Badge variant="warning">{t('invoiceGenerateBadgeNoWater')}</Badge>
                                    )}
                                    {c.ended_before_period && (
                                      <Badge variant="warning">{t('invoiceGenerateBadgeEnded')}</Badge>
                                    )}
                                  </div>
                                </div>
                              </li>
                            ))}
                          </ul>
                        </div>
                        {missingReadings > 0 && (
                          <p className="text-xs font-normal text-muted-foreground">
                            {t('invoiceGenerateMissingReadingsHint').replace('{count}', String(missingReadings))}
                          </p>
                        )}
                      </>
                    ) : null}
                  </FormSection>
                )}

                <FormSection title={t('invoiceFormNoteLabel')}>
                  <textarea
                    className="min-h-20 rounded-md border border-input bg-transparent px-3 py-2 text-sm"
                    value={note}
                    onChange={(event) => setNote(event.target.value)}
                  />
                </FormSection>
              </>
            )}
          </div>

          {error && <p className="resource-error">{error}</p>}

          <SheetFooter>
            {result ? (
              <Button type="button" onClick={() => onOpenChange(false)}>
                {t('invoiceGenerateClose')}
              </Button>
            ) : (
              <>
                <Button type="button" variant="outline" onClick={() => onOpenChange(false)}>
                  {t('invoiceFormCancel')}
                </Button>
                <Button type="submit" disabled={saving || previewLoading || selected.size === 0}>
                  {saving ? t('invoiceGenerating') : t('invoiceGenerateSubmit').replace('{count}', String(selected.size))}
                </Button>
              </>
            )}
          </SheetFooter>
        </form>
      </SheetContent>
    </Sheet>
  )
}
