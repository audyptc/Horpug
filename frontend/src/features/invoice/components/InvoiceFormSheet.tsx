import type { FormEvent } from 'react'
import { useLanguage, type TranslationKey } from '@/shared/i18n/language'
import { Button } from '@/shared/components/ui/button'
import {
  Sheet,
  SheetContent,
  SheetDescription,
  SheetFooter,
  SheetHeader,
  SheetTitle,
} from '@/shared/components/ui/sheet'
import type { ApiContract } from '@/features/contract/types'
import type { ApiInvoice, InvoiceStatus } from '../types'
import { INVOICE_STATUSES, formatPeriod } from '../utils'

const invoiceStatusLabelKeys: Record<InvoiceStatus, TranslationKey> = {
  unpaid: 'invoiceStatusUnpaid',
  paid: 'invoiceStatusPaid',
  overdue: 'invoiceStatusOverdue',
  cancelled: 'invoiceStatusCancelled',
}

const invoiceItemTypeLabelKeys: Record<string, TranslationKey> = {
  rent: 'invoiceItemTypeRent',
  electricity: 'invoiceItemTypeElectricity',
  water: 'invoiceItemTypeWater',
  other: 'invoiceItemTypeOther',
}

type InvoiceFormSheetProps = {
  open: boolean
  onOpenChange: (open: boolean) => void
  isEdit: boolean
  contracts: ApiContract[]
  contractId: string
  onContractIdChange: (contractId: string) => void
  period: string
  onPeriodChange: (value: string) => void
  issueDate: string
  onIssueDateChange: (value: string) => void
  dueDate: string
  onDueDateChange: (value: string) => void
  status: InvoiceStatus
  onStatusChange: (status: InvoiceStatus) => void
  note: string
  onNoteChange: (value: string) => void
  invoice: ApiInvoice | null
  detailLoading: boolean
  saving: boolean
  error: string | null
  onSubmit: (event: FormEvent<HTMLFormElement>) => void
}

export function InvoiceFormSheet({
  open,
  onOpenChange,
  isEdit,
  contracts,
  contractId,
  onContractIdChange,
  period,
  onPeriodChange,
  issueDate,
  onIssueDateChange,
  dueDate,
  onDueDateChange,
  status,
  onStatusChange,
  note,
  onNoteChange,
  invoice,
  detailLoading,
  saving,
  error,
  onSubmit,
}: InvoiceFormSheetProps) {
  const { t, language } = useLanguage()

  return (
    <Sheet open={open} onOpenChange={onOpenChange}>
      <SheetContent>
        <form className="flex h-full flex-col gap-4" onSubmit={onSubmit}>
          <SheetHeader>
            <SheetTitle>{isEdit ? t('invoiceFormEditTitle') : t('invoiceFormCreateTitle')}</SheetTitle>
            <SheetDescription>
              {isEdit ? t('invoiceFormEditDescription') : t('invoiceFormCreateDescription')}
            </SheetDescription>
          </SheetHeader>

          <div className="flex flex-1 flex-col gap-4 overflow-y-auto pr-1">
            {isEdit ? (
              detailLoading || !invoice ? (
                <p className="metric-detail">{t('loading')}</p>
              ) : (
                <>
                  <label className="flex flex-col gap-1.5 text-sm font-medium">
                    {t('invoiceFormTenantLabel')}
                    <p className="text-sm font-normal text-muted-foreground">{invoice.tenant_name || '—'}</p>
                  </label>
                  <label className="flex flex-col gap-1.5 text-sm font-medium">
                    {t('invoiceFormRoomLabel')}
                    <p className="text-sm font-normal text-muted-foreground">
                      {invoice.room_number || '—'}
                      {invoice.dormitory_name ? ` (${invoice.dormitory_name})` : ''}
                    </p>
                  </label>
                  <label className="flex flex-col gap-1.5 text-sm font-medium">
                    {t('invoiceFormPeriodLabel')}
                    <p className="text-sm font-normal text-muted-foreground">
                      {formatPeriod(invoice.period_year, invoice.period_month)}
                    </p>
                  </label>

                  <div className="flex flex-col gap-1.5 text-sm font-medium">
                    {t('invoiceFormItemsLabel')}
                    <div className="rounded-md border border-input">
                      {(invoice.items ?? []).length === 0 ? (
                        <p className="px-3 py-2 text-sm font-normal text-muted-foreground">
                          {t('invoiceFormNoItems')}
                        </p>
                      ) : (
                        <ul className="divide-y divide-border">
                          {(invoice.items ?? []).map((item) => (
                            <li
                              key={item.id}
                              className="flex items-center justify-between px-3 py-2 text-sm font-normal"
                            >
                              <span>
                                {t(invoiceItemTypeLabelKeys[item.item_type] ?? 'invoiceItemTypeOther')}
                                {item.description ? ` · ${item.description}` : ''}
                              </span>
                              <span className="text-muted-foreground">{item.amount.toLocaleString()}</span>
                            </li>
                          ))}
                        </ul>
                      )}
                    </div>
                  </div>

                  <label className="flex flex-col gap-1.5 text-sm font-medium">
                    {t('invoiceFormTotalAmountLabel')}
                    <p className="text-sm font-normal text-muted-foreground">
                      {invoice.total_amount.toLocaleString()}
                    </p>
                  </label>

                  {invoice.paid_at && (
                    <label className="flex flex-col gap-1.5 text-sm font-medium">
                      {t('invoiceFormPaidAtLabel')}
                      <p className="text-sm font-normal text-muted-foreground">
                        {new Date(invoice.paid_at).toLocaleString(language === 'th' ? 'th-TH' : 'en-US')}
                      </p>
                    </label>
                  )}

                  <label className="flex flex-col gap-1.5 text-sm font-medium">
                    {t('invoiceFormStatusLabel')}
                    <select
                      className="h-10 rounded-md border border-input bg-transparent px-3 text-sm"
                      value={status}
                      onChange={(event) => onStatusChange(event.target.value as InvoiceStatus)}
                    >
                      {INVOICE_STATUSES.map((value) => (
                        <option key={value} value={value}>
                          {t(invoiceStatusLabelKeys[value])}
                        </option>
                      ))}
                    </select>
                  </label>
                </>
              )
            ) : (
              <>
                <label className="flex flex-col gap-1.5 text-sm font-medium">
                  {t('invoiceFormContractLabel')}
                  {contracts.length === 0 ? (
                    <p className="text-xs font-normal text-muted-foreground">{t('invoiceFormNoContracts')}</p>
                  ) : (
                    <select
                      className="h-10 rounded-md border border-input bg-transparent px-3 text-sm"
                      value={contractId}
                      onChange={(event) => onContractIdChange(event.target.value)}
                    >
                      <option value="">{t('invoiceFormContractPlaceholder')}</option>
                      {contracts.map((contract) => (
                        <option key={contract.id} value={contract.id}>
                          {contract.tenant_name} · {contract.room_number}
                          {contract.dormitory_name ? ` (${contract.dormitory_name})` : ''}
                        </option>
                      ))}
                    </select>
                  )}
                </label>

                <label className="flex flex-col gap-1.5 text-sm font-medium">
                  {t('invoiceFormPeriodLabel')}
                  <input
                    type="month"
                    className="h-10 rounded-md border border-input bg-transparent px-3 text-sm"
                    value={period}
                    onChange={(event) => onPeriodChange(event.target.value)}
                  />
                </label>

                <label className="flex flex-col gap-1.5 text-sm font-medium">
                  {t('invoiceFormIssueDateLabel')}
                  <input
                    type="date"
                    className="h-10 rounded-md border border-input bg-transparent px-3 text-sm"
                    value={issueDate}
                    onChange={(event) => onIssueDateChange(event.target.value)}
                  />
                </label>
              </>
            )}

            {(!isEdit || (!detailLoading && invoice)) && (
              <label className="flex flex-col gap-1.5 text-sm font-medium">
                {t('invoiceFormDueDateLabel')}
                <input
                  type="date"
                  className="h-10 rounded-md border border-input bg-transparent px-3 text-sm"
                  value={dueDate}
                  onChange={(event) => onDueDateChange(event.target.value)}
                />
              </label>
            )}

            {(!isEdit || (!detailLoading && invoice)) && (
              <label className="flex flex-col gap-1.5 text-sm font-medium">
                {t('invoiceFormNoteLabel')}
                <textarea
                  className="min-h-20 rounded-md border border-input bg-transparent px-3 py-2 text-sm"
                  value={note}
                  onChange={(event) => onNoteChange(event.target.value)}
                />
              </label>
            )}
          </div>

          {error && <p className="resource-error">{error}</p>}

          <SheetFooter>
            <Button type="button" variant="outline" onClick={() => onOpenChange(false)}>
              {t('invoiceFormCancel')}
            </Button>
            <Button type="submit" disabled={saving || (isEdit && (detailLoading || !invoice))}>
              {saving ? t('invoiceSaving') : t('invoiceFormSave')}
            </Button>
          </SheetFooter>
        </form>
      </SheetContent>
    </Sheet>
  )
}
