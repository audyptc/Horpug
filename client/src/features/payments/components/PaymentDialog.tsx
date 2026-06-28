import { useState, useEffect, useMemo } from 'react'
import { useTranslation } from 'react-i18next'
import { Check, ChevronDown, Search, Plus, X } from 'lucide-react'
import { Button } from '@/components/ui/button'
import { Input } from '@/components/ui/input'
import { Label } from '@/components/ui/label'
import {
  Dialog,
  DialogContent,
  DialogHeader,
  DialogTitle,
  DialogDescription,
  DialogFooter,
} from '@/components/ui/dialog'
import {
  Select,
  SelectContent,
  SelectItem,
  SelectTrigger,
  SelectValue,
} from '@/components/ui/select'
import { DatePicker } from '@/components/ui/date-picker'
import { Popover, PopoverContent, PopoverTrigger } from '@/components/ui/popover'
import { ScrollArea } from '@/components/ui/scroll-area'
import { billService } from '@/features/billing/billService'
import type { ApiPayment, ApiBill } from '@/types'
import { cn } from '@/lib/utils'

type SplitMethod = 'cash' | 'transfer' | 'qr'

export type PaymentFormSplit = {
  method: SplitMethod
  amount: string
}

export type PaymentForm = {
  bill_id: string
  payment_date: string
  note: string
  splits: PaymentFormSplit[]
}

type Props = {
  open: boolean
  onOpenChange: (open: boolean) => void
  editingItem: ApiPayment | null
  form: PaymentForm
  onFormChange: (form: PaymentForm) => void
  onSave: () => void
  saving: boolean
  error?: string
}

const ALL_METHODS: SplitMethod[] = ['cash', 'transfer', 'qr']

function formatBaht(n: number) {
  return n.toLocaleString('th-TH', { minimumFractionDigits: 2, maximumFractionDigits: 2 })
}

export function PaymentDialog({
  open,
  onOpenChange,
  editingItem,
  form,
  onFormChange,
  onSave,
  saving,
  error,
}: Props) {
  const { t } = useTranslation()

  const [bills, setBills] = useState<ApiBill[]>([])
  const [billsLoading, setBillsLoading] = useState(false)
  const [billSearch, setBillSearch] = useState('')
  const [pickerOpen, setPickerOpen] = useState(false)

  const totalAmount = form.splits.reduce((sum, s) => sum + (Number(s.amount) || 0), 0)

  const isSaveDisabled =
    saving ||
    !form.bill_id ||
    form.splits.length === 0 ||
    form.splits.some((s) => !s.amount || Number(s.amount) <= 0) ||
    !form.payment_date

  function set(patch: Partial<PaymentForm>) {
    onFormChange({ ...form, ...patch })
  }

  function updateSplit(idx: number, patch: Partial<PaymentFormSplit>) {
    const splits = form.splits.map((s, i) => (i === idx ? { ...s, ...patch } : s))
    set({ splits })
  }

  function removeSplit(idx: number) {
    set({ splits: form.splits.filter((_, i) => i !== idx) })
  }

  function addSplit() {
    const used = form.splits.map((s) => s.method)
    const next = ALL_METHODS.find((m) => !used.includes(m))
    if (next) set({ splits: [...form.splits, { method: next, amount: '' }] })
  }

  useEffect(() => {
    if (!open || editingItem) return
    setBillsLoading(true)
    billService
      .list(1, 200)
      .then((r) => setBills(r.data.filter((b) => b.status !== 'paid')))
      .catch(() => {})
      .finally(() => setBillsLoading(false))
  }, [open, editingItem])

  const selectedBill = bills.find((b) => b.id === form.bill_id) ?? null

  const filteredBills = useMemo(() => {
    const q = billSearch.toLowerCase()
    if (!q) return bills
    return bills.filter((b) => {
      const name = `${b.tenant_first_name} ${b.tenant_last_name}`.toLowerCase()
      return name.includes(q) || b.room_number.toLowerCase().includes(q)
    })
  }, [bills, billSearch])

  function selectBill(bill: ApiBill) {
    set({
      bill_id: bill.id,
      splits: [{ method: 'cash', amount: String(bill.total_amount) }],
    })
    setPickerOpen(false)
    setBillSearch('')
  }

  const billInfo = editingItem
    ? {
        tenantName: `${editingItem.tenant_first_name} ${editingItem.tenant_last_name}`,
        roomNumber: editingItem.room_number,
        billingMonth: editingItem.billing_month?.slice(0, 7) ?? '-',
        total: editingItem.bill_total,
      }
    : selectedBill
      ? {
          tenantName: `${selectedBill.tenant_first_name} ${selectedBill.tenant_last_name}`,
          roomNumber: selectedBill.room_number,
          billingMonth: selectedBill.billing_month?.slice(0, 7) ?? '-',
          total: selectedBill.total_amount,
        }
      : null

  return (
    <Dialog open={open} onOpenChange={onOpenChange}>
      <DialogContent className="max-w-md">
        <DialogHeader>
          <DialogTitle>
            {editingItem ? t('payments.editPayment') : t('payments.createPayment')}
          </DialogTitle>
          <DialogDescription>
            {editingItem ? t('payments.editDesc') : t('payments.createDesc')}
          </DialogDescription>
        </DialogHeader>

        <div className="grid gap-4 py-4">
          {!editingItem && (
            <div className="space-y-1.5">
              <Label>{t('payments.selectBill')} *</Label>
              <Popover open={pickerOpen} onOpenChange={setPickerOpen}>
                <PopoverTrigger asChild>
                  <button
                    type="button"
                    className={cn(
                      'flex w-full items-center justify-between rounded-md border border-input bg-background px-3 py-2 text-sm',
                      'ring-offset-background focus-visible:outline-none focus-visible:ring-2 focus-visible:ring-ring focus-visible:ring-offset-2',
                      'hover:bg-accent/40 transition-colors',
                      !selectedBill && 'text-muted-foreground',
                    )}
                  >
                    <span className="truncate">
                      {selectedBill
                        ? `${selectedBill.tenant_first_name} ${selectedBill.tenant_last_name} · ห้อง ${selectedBill.room_number} · ${selectedBill.billing_month?.slice(0, 7)}`
                        : t('payments.selectBillPlaceholder')}
                    </span>
                    <ChevronDown className="w-4 h-4 shrink-0 ml-2 text-muted-foreground" />
                  </button>
                </PopoverTrigger>
                <PopoverContent className="w-[--radix-popover-trigger-width] p-0" align="start">
                  <div className="p-2 border-b">
                    <div className="relative">
                      <Search className="absolute left-2.5 top-1/2 -translate-y-1/2 w-3.5 h-3.5 text-muted-foreground" />
                      <Input
                        placeholder={t('payments.searchBillPlaceholder')}
                        className="pl-8 h-8 text-sm"
                        value={billSearch}
                        onChange={(e) => setBillSearch(e.target.value)}
                        autoFocus
                      />
                    </div>
                  </div>
                  <ScrollArea className="max-h-56">
                    {billsLoading ? (
                      <div className="p-4 text-center text-sm text-muted-foreground">
                        {t('common.loading')}
                      </div>
                    ) : filteredBills.length === 0 ? (
                      <div className="p-4 text-center text-sm text-muted-foreground">
                        {t('payments.noPendingBills')}
                      </div>
                    ) : (
                      <div className="py-1">
                        {filteredBills.map((bill) => (
                          <button
                            key={bill.id}
                            type="button"
                            onClick={() => selectBill(bill)}
                            className={cn(
                              'w-full flex items-center gap-2 px-3 py-2 text-left text-sm hover:bg-accent transition-colors',
                              bill.id === form.bill_id && 'bg-accent',
                            )}
                          >
                            <Check
                              className={cn(
                                'w-3.5 h-3.5 shrink-0 text-primary',
                                bill.id !== form.bill_id && 'invisible',
                              )}
                            />
                            <div className="flex-1 min-w-0">
                              <div className="flex items-center justify-between gap-2">
                                <span className="font-medium truncate">
                                  {bill.tenant_first_name} {bill.tenant_last_name}
                                </span>
                                <span className="text-xs font-semibold text-emerald-600 shrink-0">
                                  ฿{formatBaht(bill.total_amount)}
                                </span>
                              </div>
                              <div className="flex items-center justify-between gap-2 text-xs text-muted-foreground">
                                <span>ห้อง {bill.room_number}</span>
                                <span>{bill.billing_month?.slice(0, 7)}</span>
                              </div>
                            </div>
                          </button>
                        ))}
                      </div>
                    )}
                  </ScrollArea>
                </PopoverContent>
              </Popover>
            </div>
          )}

          {billInfo && (
            <div className="rounded-md border bg-muted/30 p-3 grid grid-cols-2 gap-y-1.5 text-sm">
              <span className="text-muted-foreground">{t('payments.colTenant')}</span>
              <span className="text-right font-medium">{billInfo.tenantName}</span>
              <span className="text-muted-foreground">{t('payments.colRoom')}</span>
              <span className="text-right">{billInfo.roomNumber}</span>
              <span className="text-muted-foreground">{t('payments.colMonth')}</span>
              <span className="text-right">{billInfo.billingMonth}</span>
              <span className="text-muted-foreground border-t pt-1.5">{t('payments.colBillTotal')}</span>
              <span className="text-right font-semibold text-emerald-600 border-t pt-1.5">
                ฿{formatBaht(billInfo.total)}
              </span>
            </div>
          )}

          <div className="space-y-1.5">
            <Label>{t('payments.method')} *</Label>
            <div className="space-y-2">
              {form.splits.map((split, idx) => {
                const usedMethods = form.splits.filter((_, i) => i !== idx).map((s) => s.method)
                return (
                  <div key={idx} className="flex gap-2 items-center">
                    <Select
                      value={split.method}
                      onValueChange={(v) => updateSplit(idx, { method: v as SplitMethod })}
                    >
                      <SelectTrigger className="flex-1">
                        <SelectValue />
                      </SelectTrigger>
                      <SelectContent>
                        {ALL_METHODS.map((m) => (
                          <SelectItem key={m} value={m} disabled={usedMethods.includes(m)}>
                            {t(`payments.methods.${m}`)}
                          </SelectItem>
                        ))}
                      </SelectContent>
                    </Select>
                    <Input
                      className="w-28 text-right"
                      type="number"
                      min="0"
                      step="0.01"
                      placeholder="0.00"
                      value={split.amount}
                      onChange={(e) => updateSplit(idx, { amount: e.target.value })}
                    />
                    {form.splits.length > 1 && (
                      <Button
                        type="button"
                        variant="ghost"
                        size="icon"
                        className="h-9 w-9 shrink-0 text-muted-foreground hover:text-destructive"
                        onClick={() => removeSplit(idx)}
                      >
                        <X className="w-4 h-4" />
                      </Button>
                    )}
                  </div>
                )
              })}
            </div>

            {form.splits.length < ALL_METHODS.length && (
              <Button
                type="button"
                variant="outline"
                size="sm"
                className="gap-1.5 h-8 mt-1"
                onClick={addSplit}
              >
                <Plus className="w-3.5 h-3.5" />
                {t('payments.addSplit')}
              </Button>
            )}

            {form.splits.length > 1 && (
              <div className="flex justify-between text-sm pt-1">
                <span className="text-muted-foreground">{t('payments.splitTotal')}</span>
                <span className="font-semibold text-emerald-600">฿{formatBaht(totalAmount)}</span>
              </div>
            )}
          </div>

          <div className="space-y-1.5">
            <Label>{t('payments.paymentDate')} *</Label>
            <DatePicker value={form.payment_date} onChange={(v) => set({ payment_date: v })} />
          </div>

          <div className="space-y-1.5">
            <Label>{t('payments.note')}</Label>
            <textarea
              rows={2}
              placeholder={t('payments.notePlaceholder')}
              value={form.note}
              onChange={(e: React.ChangeEvent<HTMLTextAreaElement>) => set({ note: e.target.value })}
              className="flex min-h-16 w-full rounded-md border border-input bg-background px-3 py-2 text-sm ring-offset-background placeholder:text-muted-foreground focus-visible:outline-none focus-visible:ring-2 focus-visible:ring-ring focus-visible:ring-offset-2 resize-none"
            />
          </div>
        </div>

        {error && (
          <p className="text-sm text-destructive -mt-2">{error}</p>
        )}

        <DialogFooter>
          <Button variant="outline" onClick={() => onOpenChange(false)}>
            {t('common.cancel')}
          </Button>
          <Button onClick={onSave} disabled={isSaveDisabled}>
            {saving
              ? t('common.loading')
              : editingItem
                ? t('payments.saveChanges')
                : t('payments.createPayment')}
          </Button>
        </DialogFooter>
      </DialogContent>
    </Dialog>
  )
}
