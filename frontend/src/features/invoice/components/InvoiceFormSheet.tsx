import { useState, type FormEvent, type ReactNode } from 'react'
import { CalendarIcon, ChevronLeft, ChevronRight, X } from 'lucide-react'
import { useLanguage, type TranslationKey } from '@/shared/i18n/language'
import { cn } from '@/shared/lib/utils'
import { Button } from '@/shared/components/ui/button'
import { Calendar } from '@/shared/components/ui/calendar'
import { Combobox } from '@/shared/components/ui/combobox'
import { Popover, PopoverContent, PopoverTrigger } from '@/shared/components/ui/popover'
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
import { INVOICE_STATUSES, formatPeriod, parsePeriodInputValue, toPeriodInputValue } from '../utils'

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

function parseDateInput(value: string): Date | undefined {
  if (!value) return undefined
  const [year, month, day] = value.split('-').map(Number)
  return new Date(year, month - 1, day)
}

function formatDateInput(date: Date | undefined): string {
  if (!date) return ''
  const year = date.getFullYear()
  const month = String(date.getMonth() + 1).padStart(2, '0')
  const day = String(date.getDate()).padStart(2, '0')
  return `${year}-${month}-${day}`
}

type DatePickerFieldProps = {
  value: string
  onChange: (value: string) => void
  placeholder: string
}

function DatePickerField({ value, onChange, placeholder }: DatePickerFieldProps) {
  const { t, language } = useLanguage()
  const [open, setOpen] = useState(false)
  const date = parseDateInput(value)
  const dateLocale = language === 'th' ? 'th-TH' : 'en-US'

  return (
    <div className="flex min-w-0 items-center gap-1.5">
      <Popover open={open} onOpenChange={setOpen}>
        <PopoverTrigger asChild>
          <Button
            type="button"
            variant="outline"
            className={cn(
              'h-10 min-w-0 flex-1 justify-start gap-2 px-3 text-sm font-normal',
              !date && 'text-muted-foreground'
            )}
          >
            <CalendarIcon className="size-4 shrink-0" />
            <span className="truncate">{date ? date.toLocaleDateString(dateLocale, { dateStyle: 'medium' }) : placeholder}</span>
          </Button>
        </PopoverTrigger>
        <PopoverContent className="w-auto p-0" align="start">
          <Calendar
            mode="single"
            defaultMonth={date ?? new Date()}
            selected={date}
            onSelect={(next) => {
              onChange(formatDateInput(next))
              setOpen(false)
            }}
          />
        </PopoverContent>
      </Popover>

      {value && (
        <Button
          type="button"
          size="icon"
          variant="ghost"
          className="h-10 w-10 shrink-0 text-muted-foreground"
          title={t('invoiceFormDateClear')}
          aria-label={t('invoiceFormDateClear')}
          onClick={() => onChange('')}
        >
          <X className="size-4" />
        </Button>
      )}
    </div>
  )
}

type MonthPickerFieldProps = {
  value: string
  onChange: (value: string) => void
  placeholder: string
}

function MonthPickerField({ value, onChange, placeholder }: MonthPickerFieldProps) {
  const { t, language } = useLanguage()
  const [open, setOpen] = useState(false)
  const parsed = parsePeriodInputValue(value)
  const [viewYear, setViewYear] = useState(parsed?.year ?? new Date().getFullYear())
  const dateLocale = language === 'th' ? 'th-TH' : 'en-US'

  function handleOpenChange(next: boolean) {
    if (next) setViewYear(parsed?.year ?? new Date().getFullYear())
    setOpen(next)
  }

  const monthLabels = Array.from({ length: 12 }, (_, index) =>
    new Date(2000, index, 1).toLocaleDateString(dateLocale, { month: 'short' })
  )

  return (
    <div className="flex min-w-0 items-center gap-1.5">
      <Popover open={open} onOpenChange={handleOpenChange}>
        <PopoverTrigger asChild>
          <Button
            type="button"
            variant="outline"
            className={cn(
              'h-10 min-w-0 flex-1 justify-start gap-2 px-3 text-sm font-normal',
              !parsed && 'text-muted-foreground'
            )}
          >
            <CalendarIcon className="size-4 shrink-0" />
            <span className="truncate">
              {parsed
                ? new Date(parsed.year, parsed.month - 1, 1).toLocaleDateString(dateLocale, {
                    year: 'numeric',
                    month: 'long',
                  })
                : placeholder}
            </span>
          </Button>
        </PopoverTrigger>
        <PopoverContent className="w-56 p-3" align="start">
          <div className="flex items-center justify-between pb-2">
            <Button
              type="button"
              variant="ghost"
              size="icon"
              className="size-7"
              onClick={() => setViewYear((year) => year - 1)}
            >
              <ChevronLeft className="size-4" />
            </Button>
            <span className="text-sm font-medium">{viewYear}</span>
            <Button
              type="button"
              variant="ghost"
              size="icon"
              className="size-7"
              onClick={() => setViewYear((year) => year + 1)}
            >
              <ChevronRight className="size-4" />
            </Button>
          </div>
          <div className="grid grid-cols-3 gap-1.5">
            {monthLabels.map((label, index) => {
              const month = index + 1
              const isSelected = parsed?.year === viewYear && parsed.month === month
              return (
                <Button
                  key={label}
                  type="button"
                  variant={isSelected ? 'default' : 'outline'}
                  size="sm"
                  className="h-8"
                  onClick={() => {
                    onChange(toPeriodInputValue(viewYear, month))
                    setOpen(false)
                  }}
                >
                  {label}
                </Button>
              )
            })}
          </div>
        </PopoverContent>
      </Popover>

      {value && (
        <Button
          type="button"
          size="icon"
          variant="ghost"
          className="h-10 w-10 shrink-0 text-muted-foreground"
          title={t('invoiceFormDateClear')}
          aria-label={t('invoiceFormDateClear')}
          onClick={() => onChange('')}
        >
          <X className="size-4" />
        </Button>
      )}
    </div>
  )
}

type FormSectionProps = {
  title?: string
  children: ReactNode
}

function FormSection({ title, children }: FormSectionProps) {
  return (
    <div className="flex flex-col gap-3">
      {title && <h3 className="text-xs font-semibold uppercase tracking-wide text-muted-foreground">{title}</h3>}
      {children}
    </div>
  )
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
  electricityAmount: number | null
  waterAmount: number | null
  utilityLoading: boolean
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
  itemSaving: boolean
  itemError: string | null
  removingItemId: string | null
  onAddItem: (description: string, amount: number) => void
  onRemoveItem: (itemId: string) => void
  pendingItems: { description: string; amount: number }[]
  onAddPendingItem: (description: string, amount: number) => void
  onRemovePendingItem: (index: number) => void
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
  electricityAmount,
  waterAmount,
  utilityLoading,
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
  itemSaving,
  itemError,
  removingItemId,
  onAddItem,
  onRemoveItem,
  pendingItems,
  onAddPendingItem,
  onRemovePendingItem,
}: InvoiceFormSheetProps) {
  const { t, language } = useLanguage()
  const selectedContract = contracts.find((contract) => contract.id === contractId)
  const selectedPeriod = parsePeriodInputValue(period)
  const estimatedTotal = selectedContract
    ? selectedContract.rent_price + (electricityAmount ?? 0) + (waterAmount ?? 0)
    : 0
  const pendingItemsTotal = pendingItems.reduce((sum, item) => sum + item.amount, 0)
  const grandTotal = estimatedTotal + pendingItemsTotal
  const [newItemDescription, setNewItemDescription] = useState('')
  const [newItemAmount, setNewItemAmount] = useState('')
  const [newPendingDescription, setNewPendingDescription] = useState('')
  const [newPendingAmount, setNewPendingAmount] = useState('')

  const canEditItems = !!invoice && invoice.status !== 'paid' && invoice.status !== 'cancelled'

  function handleAddItemClick() {
    const amount = Number(newItemAmount)
    if (!newItemDescription.trim() || !Number.isFinite(amount) || amount <= 0) return
    onAddItem(newItemDescription.trim(), amount)
    setNewItemDescription('')
    setNewItemAmount('')
  }

  function handleAddPendingItemClick() {
    const amount = Number(newPendingAmount)
    if (!newPendingDescription.trim() || !Number.isFinite(amount) || amount <= 0) return
    onAddPendingItem(newPendingDescription.trim(), amount)
    setNewPendingDescription('')
    setNewPendingAmount('')
  }

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

          <div className="flex flex-1 flex-col gap-6 overflow-y-auto pr-1">
            {isEdit ? (
              detailLoading || !invoice ? (
                <p className="metric-detail">{t('loading')}</p>
              ) : (
                <>
                  <FormSection title={t('invoiceFormSectionBilling')}>
                    <div className="rounded-md border border-input">
                      <ul className="divide-y divide-border text-sm font-normal">
                        <li className="flex items-center justify-between gap-2 px-3 py-2">
                          <span className="font-medium">{t('invoiceFormTenantLabel')}</span>
                          <span className="text-muted-foreground">{invoice.tenant_name || '—'}</span>
                        </li>
                        <li className="flex items-center justify-between gap-2 px-3 py-2">
                          <span className="font-medium">{t('invoiceFormRoomLabel')}</span>
                          <span className="text-muted-foreground">
                            {invoice.room_number || '—'}
                            {invoice.dormitory_name ? ` (${invoice.dormitory_name})` : ''}
                          </span>
                        </li>
                        <li className="flex items-center justify-between gap-2 px-3 py-2">
                          <span className="font-medium">{t('invoiceFormPeriodLabel')}</span>
                          <span className="text-muted-foreground">
                            {formatPeriod(invoice.period_year, invoice.period_month)}
                          </span>
                        </li>
                      </ul>
                    </div>
                  </FormSection>

                  <FormSection title={t('invoiceFormItemsLabel')}>
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
                              className="flex items-center justify-between gap-2 px-3 py-2 text-sm font-normal"
                            >
                              <span>
                                {t(invoiceItemTypeLabelKeys[item.item_type] ?? 'invoiceItemTypeOther')}
                                {item.description ? ` · ${item.description}` : ''}
                              </span>
                              <span className="flex items-center gap-2">
                                <span className="text-muted-foreground">{item.amount.toLocaleString()}</span>
                                {canEditItems && item.item_type === 'other' && (
                                  <Button
                                    type="button"
                                    size="icon"
                                    variant="ghost"
                                    className="h-7 w-7 shrink-0 text-muted-foreground"
                                    title={t('invoiceFormItemRemove')}
                                    aria-label={t('invoiceFormItemRemove')}
                                    onClick={() => onRemoveItem(item.id)}
                                    disabled={removingItemId === item.id}
                                  >
                                    <X className="size-3.5" />
                                  </Button>
                                )}
                              </span>
                            </li>
                          ))}
                        </ul>
                      )}
                    </div>

                    {canEditItems ? (
                      <div className="flex flex-wrap items-center gap-1.5">
                        <input
                          type="text"
                          className="h-9 min-w-0 flex-1 rounded-md border border-input bg-transparent px-2.5 text-sm"
                          placeholder={t('invoiceFormItemDescriptionPlaceholder')}
                          value={newItemDescription}
                          onChange={(event) => setNewItemDescription(event.target.value)}
                        />
                        <input
                          type="number"
                          step="0.01"
                          min="0"
                          className="h-9 w-28 shrink-0 rounded-md border border-input bg-transparent px-2.5 text-right text-sm"
                          placeholder={t('invoiceFormItemAmountPlaceholder')}
                          value={newItemAmount}
                          onChange={(event) => setNewItemAmount(event.target.value)}
                        />
                        <Button
                          type="button"
                          variant="outline"
                          size="sm"
                          onClick={handleAddItemClick}
                          disabled={itemSaving || !newItemDescription.trim() || !newItemAmount}
                        >
                          {t('invoiceFormAddItem')}
                        </Button>
                      </div>
                    ) : (
                      <p className="text-xs font-normal text-muted-foreground">{t('invoiceFormItemsLocked')}</p>
                    )}
                    {itemError && <p className="resource-error">{itemError}</p>}
                  </FormSection>

                  <FormSection title={t('invoiceFormSectionSummary')}>
                    <div className="rounded-md border border-input">
                      <ul className="divide-y divide-border text-sm font-normal">
                        <li className="flex items-center justify-between gap-2 px-3 py-2 font-medium">
                          <span>{t('invoiceFormTotalAmountLabel')}</span>
                          <span>{invoice.total_amount.toLocaleString()}</span>
                        </li>
                        {invoice.paid_at && (
                          <li className="flex items-center justify-between gap-2 px-3 py-2">
                            <span>{t('invoiceFormPaidAtLabel')}</span>
                            <span className="text-muted-foreground">
                              {new Date(invoice.paid_at).toLocaleString(language === 'th' ? 'th-TH' : 'en-US')}
                            </span>
                          </li>
                        )}
                      </ul>
                    </div>

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
                  </FormSection>

                  <FormSection title={t('invoiceFormSectionDates')}>
                    <label className="flex flex-col gap-1.5 text-sm font-medium">
                      {t('invoiceFormDueDateLabel')}
                      <DatePickerField
                        value={dueDate}
                        onChange={onDueDateChange}
                        placeholder={t('invoiceFormDueDateLabel')}
                      />
                    </label>
                  </FormSection>

                  <FormSection title={t('invoiceFormNoteLabel')}>
                    <textarea
                      className="min-h-20 rounded-md border border-input bg-transparent px-3 py-2 text-sm"
                      value={note}
                      onChange={(event) => onNoteChange(event.target.value)}
                    />
                  </FormSection>
                </>
              )
            ) : (
              <>
                <FormSection title={t('invoiceFormSectionBilling')}>
                  <label className="flex flex-col gap-1.5 text-sm font-medium">
                    {t('invoiceFormContractLabel')}
                    {contracts.length === 0 ? (
                      <p className="text-xs font-normal text-muted-foreground">{t('invoiceFormNoContracts')}</p>
                    ) : (
                      <Combobox
                        options={contracts.map((contract) => ({
                          value: contract.id,
                          label: `${contract.tenant_name} · ${contract.room_number}${
                            contract.dormitory_name ? ` (${contract.dormitory_name})` : ''
                          }`,
                        }))}
                        value={contractId}
                        onChange={onContractIdChange}
                        placeholder={t('invoiceFormContractPlaceholder')}
                        searchPlaceholder={t('invoiceFormContractSearchPlaceholder')}
                        emptyText={t('invoiceFormContractNoResults')}
                      />
                    )}
                  </label>

                  <label className="flex flex-col gap-1.5 text-sm font-medium">
                    {t('invoiceFormPeriodLabel')}
                    <MonthPickerField
                      value={period}
                      onChange={onPeriodChange}
                      placeholder={t('invoiceFormPeriodLabel')}
                    />
                  </label>

                  <div className="grid grid-cols-2 gap-3">
                    <label className="flex flex-col gap-1.5 text-sm font-medium">
                      {t('invoiceFormIssueDateLabel')}
                      <DatePickerField
                        value={issueDate}
                        onChange={onIssueDateChange}
                        placeholder={t('invoiceFormIssueDateLabel')}
                      />
                    </label>

                    <label className="flex flex-col gap-1.5 text-sm font-medium">
                      {t('invoiceFormDueDateLabel')}
                      <DatePickerField
                        value={dueDate}
                        onChange={onDueDateChange}
                        placeholder={t('invoiceFormDueDateLabel')}
                      />
                    </label>
                  </div>
                </FormSection>

                {selectedContract && (
                  <FormSection title={t('invoiceFormChargesPreviewLabel')}>
                    <div className="rounded-md border border-input">
                      <ul className="divide-y divide-border text-sm font-normal">
                        <li className="flex items-center justify-between gap-2 px-3 py-2">
                          <span>{t('invoiceFormRoomLabel')}</span>
                          <span className="text-muted-foreground">
                            {selectedContract.room_number || '—'}
                            {selectedContract.dormitory_name ? ` (${selectedContract.dormitory_name})` : ''}
                          </span>
                        </li>
                        <li className="flex items-center justify-between gap-2 px-3 py-2">
                          <span>{t('invoiceItemTypeRent')}</span>
                          <span className="text-muted-foreground">
                            {selectedContract.rent_price.toLocaleString()}
                          </span>
                        </li>
                        {!selectedPeriod ? (
                          <li className="px-3 py-2 text-xs text-muted-foreground">
                            {t('invoiceFormSelectPeriodHint')}
                          </li>
                        ) : utilityLoading ? (
                          <li className="px-3 py-2 text-sm text-muted-foreground">{t('loading')}</li>
                        ) : (
                          <>
                            <li className="flex items-center justify-between gap-2 px-3 py-2">
                              <span>{t('invoiceItemTypeElectricity')}</span>
                              <span className="text-muted-foreground">
                                {electricityAmount !== null
                                  ? electricityAmount.toLocaleString()
                                  : t('invoiceFormUtilityNoReading')}
                              </span>
                            </li>
                            <li className="flex items-center justify-between gap-2 px-3 py-2">
                              <span>{t('invoiceItemTypeWater')}</span>
                              <span className="text-muted-foreground">
                                {waterAmount !== null
                                  ? waterAmount.toLocaleString()
                                  : t('invoiceFormUtilityNoReading')}
                              </span>
                            </li>
                            <li className="flex items-center justify-between gap-2 px-3 py-2 font-medium">
                              <span>{t('invoiceFormEstimatedTotalLabel')}</span>
                              <span>{estimatedTotal.toLocaleString()}</span>
                            </li>
                          </>
                        )}
                      </ul>
                    </div>
                  </FormSection>
                )}

                <FormSection title={t('invoiceFormItemsLabel')}>
                  {pendingItems.length > 0 && (
                    <div className="rounded-md border border-input">
                      <ul className="divide-y divide-border">
                        {pendingItems.map((item, index) => (
                          <li
                            key={`${item.description}-${index}`}
                            className="flex items-center justify-between gap-2 px-3 py-2 text-sm font-normal"
                          >
                            <span>{item.description}</span>
                            <span className="flex items-center gap-2">
                              <span className="text-muted-foreground">{item.amount.toLocaleString()}</span>
                              <Button
                                type="button"
                                size="icon"
                                variant="ghost"
                                className="h-7 w-7 shrink-0 text-muted-foreground"
                                title={t('invoiceFormItemRemove')}
                                aria-label={t('invoiceFormItemRemove')}
                                onClick={() => onRemovePendingItem(index)}
                              >
                                <X className="size-3.5" />
                              </Button>
                            </span>
                          </li>
                        ))}
                      </ul>
                    </div>
                  )}
                  <div className="flex flex-wrap items-center gap-1.5">
                    <input
                      type="text"
                      className="h-9 min-w-0 flex-1 rounded-md border border-input bg-transparent px-2.5 text-sm"
                      placeholder={t('invoiceFormItemDescriptionPlaceholder')}
                      value={newPendingDescription}
                      onChange={(event) => setNewPendingDescription(event.target.value)}
                    />
                    <input
                      type="number"
                      step="0.01"
                      min="0"
                      className="h-9 w-28 shrink-0 rounded-md border border-input bg-transparent px-2.5 text-right text-sm"
                      placeholder={t('invoiceFormItemAmountPlaceholder')}
                      value={newPendingAmount}
                      onChange={(event) => setNewPendingAmount(event.target.value)}
                    />
                    <Button
                      type="button"
                      variant="outline"
                      size="sm"
                      onClick={handleAddPendingItemClick}
                      disabled={!newPendingDescription.trim() || !newPendingAmount}
                    >
                      {t('invoiceFormAddItem')}
                    </Button>
                  </div>
                </FormSection>

                <FormSection>
                  <div className="flex items-center justify-between gap-2 rounded-md border border-input px-3 py-2 text-sm font-semibold">
                    <span>{t('invoiceFormGrandTotalLabel')}</span>
                    <span>{grandTotal.toLocaleString()}</span>
                  </div>
                </FormSection>

                <FormSection title={t('invoiceFormNoteLabel')}>
                  <textarea
                    className="min-h-20 rounded-md border border-input bg-transparent px-3 py-2 text-sm"
                    value={note}
                    onChange={(event) => onNoteChange(event.target.value)}
                  />
                </FormSection>
              </>
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
