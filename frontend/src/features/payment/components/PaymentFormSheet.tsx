import { useMemo, type FormEvent } from 'react'
import { Plus, Trash2 } from 'lucide-react'
import { useLanguage, type TranslationKey } from '@/shared/i18n/language'
import { Button } from '@/shared/components/ui/button'
import { DatePickerField } from '@/shared/components/date-picker-field'
import { Combobox } from '@/shared/components/ui/combobox'
import {
  Sheet,
  SheetContent,
  SheetDescription,
  SheetFooter,
  SheetHeader,
  SheetTitle,
} from '@/shared/components/ui/sheet'
import type { ApiInvoice } from '@/features/invoice/types'
import { InvoiceSearchSelect } from '@/features/invoice/components/InvoiceSearchSelect'
import type { PaymentMethod } from '../types'
import { PAYMENT_METHODS, sumPaymentItems, type PaymentItemFormRow } from '../utils'

const paymentMethodLabelKeys: Record<PaymentMethod, TranslationKey> = {
  cash: 'paymentMethodCash',
  transfer: 'paymentMethodTransfer',
  credit_card: 'paymentMethodCreditCard',
  other: 'paymentMethodOther',
  deposit: 'paymentMethodDeposit',
}

type PaymentFormSheetProps = {
  open: boolean
  onOpenChange: (open: boolean) => void
  // Editing keeps the payment on its invoice: the invoice picker is locked,
  // since moving a payment elsewhere is a delete and a fresh record.
  isEditing: boolean
  invoiceLabel: string
  onSelectInvoice: (invoice: ApiInvoice) => void
  paymentDate: string
  onPaymentDateChange: (value: string) => void
  items: PaymentItemFormRow[]
  onAddItem: () => void
  onRemoveItem: (key: number) => void
  onItemChange: (key: number, patch: Partial<PaymentItemFormRow>) => void
  note: string
  onNoteChange: (value: string) => void
  saving: boolean
  error: string | null
  onSubmit: (event: FormEvent<HTMLFormElement>) => void
}

export function PaymentFormSheet({
  open,
  onOpenChange,
  isEditing,
  invoiceLabel,
  onSelectInvoice,
  paymentDate,
  onPaymentDateChange,
  items,
  onAddItem,
  onRemoveItem,
  onItemChange,
  note,
  onNoteChange,
  saving,
  error,
  onSubmit,
}: PaymentFormSheetProps) {
  const { t } = useLanguage()

  const total = useMemo(() => sumPaymentItems(items), [items])

  return (
    <Sheet open={open} onOpenChange={onOpenChange}>
      <SheetContent>
        <form className="flex h-full flex-col gap-4" onSubmit={onSubmit}>
          <SheetHeader>
            <SheetTitle>{t(isEditing ? 'paymentFormEditTitle' : 'paymentFormCreateTitle')}</SheetTitle>
            <SheetDescription>
              {t(isEditing ? 'paymentFormEditDescription' : 'paymentFormCreateDescription')}
            </SheetDescription>
          </SheetHeader>

          <div className="flex flex-1 flex-col gap-4 overflow-y-auto overflow-x-hidden pr-1">
            <label className="flex flex-col gap-1.5 text-sm font-medium">
              {t('paymentFormInvoiceLabel')}
              <InvoiceSearchSelect
                selectedLabel={invoiceLabel}
                onSelectInvoice={onSelectInvoice}
                placeholder={t('paymentFormInvoicePlaceholder')}
                searchPlaceholder={t('pickerSearchInvoicePlaceholder')}
                noResultsLabel={t('pickerNoInvoices')}
                disabled={isEditing}
              />
            </label>

            {isEditing && (
              <p className="rounded-md bg-muted px-3 py-2 text-xs font-normal text-muted-foreground">
                {t('paymentReceiptLockedHint')}
              </p>
            )}

            <label className="flex flex-col gap-1.5 text-sm font-medium">
              {t('paymentFormDateLabel')}
              {isEditing ? (
                <span className="h-10 rounded-md border border-input bg-muted/40 px-3 py-2 text-sm font-normal">
                  {paymentDate}
                </span>
              ) : (
                <DatePickerField
                  value={paymentDate}
                  onChange={onPaymentDateChange}
                  placeholder={t('paymentFormDateLabel')}
                />
              )}
            </label>

            <div className="flex flex-col gap-2">
              <div className="flex items-center justify-between gap-2">
                <span className="text-sm font-medium">{t('paymentFormItemsLabel')}</span>
                {!isEditing && (
                  <Button type="button" size="sm" variant="outline" onClick={onAddItem}>
                    <Plus />
                    {t('paymentAddItem')}
                  </Button>
                )}
              </div>

              <div className="flex flex-col gap-3">
                {items.map((item, index) => (
                  <div key={item.key} className="flex flex-col gap-3 rounded-lg border border-border bg-muted/30 p-3">
                    <div className="flex items-center justify-between gap-2">
                      <span className="text-xs font-semibold uppercase tracking-wide text-muted-foreground">
                        {t('paymentItemLabel')} #{index + 1}
                      </span>
                      <Button
                        type="button"
                        size="icon"
                        variant="ghost"
                        className="h-7 w-7 shrink-0 text-destructive hover:bg-destructive/10 hover:text-destructive"
                        title={t('paymentRemoveItem')}
                        aria-label={t('paymentRemoveItem')}
                        onClick={() => onRemoveItem(item.key)}
                        disabled={isEditing || items.length <= 1}
                      >
                        <Trash2 className="h-4 w-4" />
                      </Button>
                    </div>

                    <label className="flex flex-col gap-1.5 text-sm font-medium">
                      {t('paymentFormMethodLabel')}
                      {isEditing ? (
                        <span className="h-10 rounded-md border border-input bg-muted/40 px-3 py-2 text-sm font-normal">
                          {t(paymentMethodLabelKeys[item.paymentMethod])}
                        </span>
                      ) : (
                        <Combobox
                          options={PAYMENT_METHODS.map((value) => ({
                            value,
                            label: t(paymentMethodLabelKeys[value]),
                          }))}
                          value={item.paymentMethod}
                          onChange={(value) => onItemChange(item.key, { paymentMethod: value as PaymentMethod })}
                          placeholder={t('paymentFormMethodLabel')}
                          searchPlaceholder={t('paymentFormMethodSearchPlaceholder')}
                          emptyText={t('paymentFormMethodNoResults')}
                        />
                      )}
                    </label>

                    <label className="flex flex-col gap-1.5 text-sm font-medium">
                      {t('paymentFormAmountLabel')}
                      <input
                        type="number"
                        min="0"
                        step="0.01"
                        className="h-10 w-full min-w-0 rounded-md border border-input bg-background px-3 text-right text-sm"
                        value={item.amount}
                        readOnly={isEditing}
                        onChange={(event) => onItemChange(item.key, { amount: event.target.value })}
                      />
                    </label>

                    <label className="flex flex-col gap-1.5 text-sm font-medium">
                      {t('paymentFormReferenceLabel')}
                      <input
                        type="text"
                        className="h-10 w-full min-w-0 rounded-md border border-input bg-background px-3 text-sm"
                        value={item.referenceNo}
                        onChange={(event) => onItemChange(item.key, { referenceNo: event.target.value })}
                      />
                    </label>
                  </div>
                ))}
              </div>

              <div className="flex items-center justify-between rounded-md bg-muted px-3 py-2 text-sm font-semibold">
                <span>{t('paymentFormTotalLabel')}</span>
                <span>{total.toLocaleString()}</span>
              </div>
            </div>

            <label className="flex flex-col gap-1.5 text-sm font-medium">
              {t('paymentFormNoteLabel')}
              <textarea
                className="min-h-20 w-full min-w-0 rounded-md border border-input bg-transparent px-3 py-2 text-sm"
                value={note}
                onChange={(event) => onNoteChange(event.target.value)}
              />
            </label>
          </div>

          {error && <p className="resource-error">{error}</p>}

          <SheetFooter>
            <Button type="button" variant="outline" onClick={() => onOpenChange(false)}>
              {t('paymentFormCancel')}
            </Button>
            <Button type="submit" disabled={saving}>
              {saving ? t('paymentSaving') : t('paymentFormSave')}
            </Button>
          </SheetFooter>
        </form>
      </SheetContent>
    </Sheet>
  )
}
