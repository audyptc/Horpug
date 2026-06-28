import { useMemo } from 'react'
import { useTranslation } from 'react-i18next'
import { Loader2, Plus, Trash2, Car, Bike } from 'lucide-react'
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
import { MonthPicker } from '@/components/ui/month-picker'
import type { ApiBill, ApiContract, ApiParkingSlot, BillStatus } from '@/types'
import { cn } from '@/lib/utils'

export type OtherItem = { label: string; amount: string }

export type BillFormState = {
  contract_id: string
  billing_month: string
  rent_amount: string
  electric_amount: string
  water_amount: string
  parking_amount: string
  other_items: OtherItem[]
  due_date: string
  note: string
  status: BillStatus
}

interface BillDialogProps {
  open: boolean
  onOpenChange: (open: boolean) => void
  editingBill: ApiBill | null
  form: BillFormState
  onFormChange: (patch: Partial<BillFormState>) => void
  onContractChange: (contractId: string) => void
  onBillingMonthChange: (v: string) => void
  autoFilling: boolean
  contracts: ApiContract[]
  parkingSlots: ApiParkingSlot[]
  selectedSlotIds: string[]
  onSlotToggle: (id: string) => void
  onSave: () => void
  saving: boolean
}

export function BillDialog({
  open,
  onOpenChange,
  editingBill,
  form,
  onFormChange,
  onContractChange,
  onBillingMonthChange,
  autoFilling,
  contracts,
  parkingSlots,
  selectedSlotIds,
  onSlotToggle,
  onSave,
  saving,
}: BillDialogProps) {
  const { t } = useTranslation()

  const activeContracts = useMemo(() => contracts.filter((c) => c.status === 'active'), [contracts])

  const otherTotal = form.other_items.reduce((sum, item) => sum + (Number(item.amount) || 0), 0)
  const previewTotal =
    (Number(form.rent_amount) || 0) +
    (Number(form.electric_amount) || 0) +
    (Number(form.water_amount) || 0) +
    (Number(form.parking_amount) || 0) +
    otherTotal

  const isSaveDisabled = saving || !form.contract_id || !form.billing_month

  function addOtherItem() {
    onFormChange({ other_items: [...form.other_items, { label: '', amount: '' }] })
  }

  function updateOtherItem(index: number, patch: Partial<OtherItem>) {
    const updated = form.other_items.map((item, i) => (i === index ? { ...item, ...patch } : item))
    onFormChange({ other_items: updated })
  }

  function removeOtherItem(index: number) {
    onFormChange({ other_items: form.other_items.filter((_, i) => i !== index) })
  }

  return (
    <Dialog open={open} onOpenChange={onOpenChange}>
      <DialogContent className="max-w-lg max-h-[90vh] overflow-y-auto">
        <DialogHeader>
          <DialogTitle>{editingBill ? t('bills.editBill') : t('bills.createBill')}</DialogTitle>
          <DialogDescription>{editingBill ? t('bills.editDesc') : t('bills.createDesc')}</DialogDescription>
        </DialogHeader>

        <div className="grid gap-4 py-4">
          <div className="space-y-1.5">
            <Label>{t('bills.contract')} *</Label>
            <Select
              value={form.contract_id}
              onValueChange={(v) => onContractChange(v)}
              disabled={!!editingBill}
            >
              <SelectTrigger><SelectValue placeholder={t('bills.selectContract')} /></SelectTrigger>
              <SelectContent>
                {activeContracts.map((c) => (
                  <SelectItem key={c.id} value={c.id}>
                    {c.tenant_first_name} {c.tenant_last_name} — {t('bills.colRoom')} {c.room_number}
                  </SelectItem>
                ))}
              </SelectContent>
            </Select>
          </div>

          <div className="space-y-1.5">
            <Label>{t('bills.billingMonth')} *</Label>
            <MonthPicker
              value={form.billing_month}
              onChange={(v) => onBillingMonthChange(v)}
              disabled={!!editingBill}
            />
          </div>

          <div className="grid grid-cols-2 gap-4">
            {(['rent_amount', 'electric_amount', 'water_amount'] as const).map((field) => (
              <div key={field} className="space-y-1.5">
                <Label className="flex items-center gap-1.5">
                  {t(`bills.${field.replace(/_([a-z])/g, (_, c) => c.toUpperCase())}`)}
                  {autoFilling && (field === 'electric_amount' || field === 'water_amount') && (
                    <Loader2 className="h-3 w-3 animate-spin text-muted-foreground" />
                  )}
                </Label>
                <Input
                  className="text-right"
                  type="number" min="0" step="0.01"
                  value={form[field]}
                  onChange={(e) => onFormChange({ [field]: e.target.value })}
                  disabled={autoFilling && (field === 'electric_amount' || field === 'water_amount')}
                />
              </div>
            ))}
          </div>

          {/* ค่าจอดรถ — checkbox list หรือ manual input */}
          {parkingSlots.length > 0 ? (
            <div className="space-y-2">
              <Label className="flex items-center gap-1.5">
                {t('bills.parkingAmount')}
                {autoFilling && <Loader2 className="h-3 w-3 animate-spin text-muted-foreground" />}
              </Label>
              <div className="rounded-md border divide-y">
                {parkingSlots.map((slot) => {
                  const checked = selectedSlotIds.includes(slot.id)
                  return (
                    <label
                      key={slot.id}
                      className={cn(
                        'flex items-center gap-3 px-3 py-2.5 cursor-pointer select-none',
                        'hover:bg-muted/40 transition-colors',
                        checked && 'bg-muted/30'
                      )}
                    >
                      <input
                        type="checkbox"
                        checked={checked}
                        onChange={() => onSlotToggle(slot.id)}
                        className="h-4 w-4 rounded border-input accent-primary"
                      />
                      <span className="text-muted-foreground shrink-0">
                        {slot.vehicle_type === 'car'
                          ? <Car className="h-4 w-4" />
                          : <Bike className="h-4 w-4" />}
                      </span>
                      <span className="flex-1 text-sm font-medium">{slot.slot_number}</span>
                      {slot.license_plate && (
                        <span className="text-xs text-muted-foreground">{slot.license_plate}</span>
                      )}
                      <span className="text-sm font-semibold tabular-nums">
                        {slot.monthly_fee > 0 ? `฿${slot.monthly_fee.toLocaleString()}` : '—'}
                      </span>
                    </label>
                  )
                })}
              </div>
              {selectedSlotIds.length > 0 && (
                <p className="text-xs text-muted-foreground text-right">
                  {t('bills.parkingSelectedTotal')}:&nbsp;
                  <span className="font-semibold text-foreground">
                    ฿{(Number(form.parking_amount) || 0).toLocaleString()}
                  </span>
                </p>
              )}
            </div>
          ) : (
            form.contract_id && (
              <div className="space-y-1.5">
                <Label className="flex items-center gap-1.5">
                  {t('bills.parkingAmount')}
                  {autoFilling && <Loader2 className="h-3 w-3 animate-spin text-muted-foreground" />}
                </Label>
                <Input
                  className="text-right"
                  type="number" min="0" step="0.01"
                  value={form.parking_amount}
                  onChange={(e) => onFormChange({ parking_amount: e.target.value })}
                  disabled={autoFilling}
                />
              </div>
            )
          )}

          {/* Other items */}
          <div className="space-y-2">
            <div className="flex items-center justify-between">
              <Label>{t('bills.otherItems')}</Label>
              <Button type="button" variant="outline" size="sm" className="h-7 gap-1 text-xs" onClick={addOtherItem}>
                <Plus className="h-3 w-3" />
                {t('bills.addOtherItem')}
              </Button>
            </div>

            {form.other_items.length > 0 && (
              <div className="space-y-2">
                {form.other_items.map((item, i) => (
                  <div key={i} className="flex gap-2 items-center">
                    <Input
                      placeholder={t('bills.otherItemLabel')}
                      value={item.label}
                      onChange={(e) => updateOtherItem(i, { label: e.target.value })}
                      className="flex-1"
                    />
                    <Input                  
                      type="number" min="0" step="0.01"
                      placeholder="0"
                      value={item.amount}
                      onChange={(e) => updateOtherItem(i, { amount: e.target.value })}
                      className="w-28 text-right"
                    />
                    <Button
                      type="button"
                      variant="ghost"
                      size="icon"
                      className="h-9 w-9 shrink-0 text-muted-foreground hover:text-destructive"
                      onClick={() => removeOtherItem(i)}
                    >
                      <Trash2 className="h-4 w-4" />
                    </Button>
                  </div>
                ))}
              </div>
            )}
          </div>

          <div className="rounded-md border bg-muted/40 px-4 py-3 text-sm flex justify-between">
            <span className="text-muted-foreground">{t('bills.totalAmount')}</span>
            <span className="font-semibold">{previewTotal.toLocaleString()} ฿</span>
          </div>

          {editingBill && (
            <div className="space-y-1.5">
              <Label>{t('bills.status')}</Label>
              <Select
                value={form.status}
                onValueChange={(v) => onFormChange({ status: v as BillStatus })}
              >
                <SelectTrigger><SelectValue /></SelectTrigger>
                <SelectContent>
                  <SelectItem value="unpaid">{t('bills.statuses.unpaid')}</SelectItem>
                  <SelectItem value="paid">{t('bills.statuses.paid')}</SelectItem>
                  <SelectItem value="overdue">{t('bills.statuses.overdue')}</SelectItem>
                </SelectContent>
              </Select>
            </div>
          )}

          <div className="space-y-1.5">
            <Label>{t('bills.dueDate')}</Label>
            <DatePicker
              value={form.due_date}
              onChange={(v) => onFormChange({ due_date: v })}
            />
          </div>

          <div className="space-y-1.5">
            <Label>{t('bills.note')}</Label>
            <textarea
              rows={2}
              placeholder={t('bills.notePlaceholder')}
              value={form.note}
              onChange={(e: React.ChangeEvent<HTMLTextAreaElement>) => onFormChange({ note: e.target.value })}
              className="flex min-h-16 w-full rounded-md border border-input bg-background px-3 py-2 text-sm ring-offset-background placeholder:text-muted-foreground focus-visible:outline-none focus-visible:ring-2 focus-visible:ring-ring focus-visible:ring-offset-2 resize-none"
            />
          </div>
        </div>

        <DialogFooter>
          <Button variant="outline" onClick={() => onOpenChange(false)}>{t('common.cancel')}</Button>
          <Button onClick={onSave} disabled={isSaveDisabled}>
            {saving ? t('common.loading') : editingBill ? t('bills.saveChanges') : t('bills.createBill')}
          </Button>
        </DialogFooter>
      </DialogContent>
    </Dialog>
  )
}
