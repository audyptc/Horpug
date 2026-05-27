import { useMemo } from 'react'
import { useTranslation } from 'react-i18next'
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
import type { ApiBill, ApiContract, BillStatus } from '@/types/api'

export type BillFormState = {
  contract_id: string
  billing_month: string
  rent_amount: string
  electric_amount: string
  water_amount: string
  other_amount: string
  other_note: string
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
  contracts: ApiContract[]
  onSave: () => void
  saving: boolean
}

export function BillDialog({
  open,
  onOpenChange,
  editingBill,
  form,
  onFormChange,
  contracts,
  onSave,
  saving,
}: BillDialogProps) {
  const { t } = useTranslation()

  const activeContracts = useMemo(() => contracts.filter((c) => c.status === 'active'), [contracts])
  const previewTotal =
    (Number(form.rent_amount) || 0) +
    (Number(form.electric_amount) || 0) +
    (Number(form.water_amount) || 0) +
    (Number(form.other_amount) || 0)

  const isSaveDisabled = saving || !form.contract_id || !form.billing_month

  return (
    <Dialog open={open} onOpenChange={onOpenChange}>
      <DialogContent className="max-w-lg">
        <DialogHeader>
          <DialogTitle>{editingBill ? t('bills.editBill') : t('bills.createBill')}</DialogTitle>
          <DialogDescription>{editingBill ? t('bills.editDesc') : t('bills.createDesc')}</DialogDescription>
        </DialogHeader>

        <div className="grid gap-4 py-4">
          <div className="space-y-1.5">
            <Label>{t('bills.contract')} *</Label>
            <Select
              value={form.contract_id}
              onValueChange={(v) => onFormChange({ contract_id: v })}
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
            <Input
              type="month"
              value={form.billing_month}
              onChange={(e) => onFormChange({ billing_month: e.target.value })}
              disabled={!!editingBill}
            />
          </div>

          <div className="grid grid-cols-2 gap-4">
            {(['rent_amount', 'electric_amount', 'water_amount', 'other_amount'] as const).map((field) => (
              <div key={field} className="space-y-1.5">
                <Label>{t(`bills.${field}`)}</Label>
                <Input
                  type="number" min="0" step="0.01"
                  value={form[field]}
                  onChange={(e) => onFormChange({ [field]: e.target.value })}
                />
              </div>
            ))}
          </div>

          <div className="space-y-1.5">
            <Label>{t('bills.otherNote')}</Label>
            <Input
              placeholder={t('bills.otherNotePlaceholder')}
              value={form.other_note}
              onChange={(e) => onFormChange({ other_note: e.target.value })}
            />
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
            <Input
              type="date"
              value={form.due_date}
              onChange={(e) => onFormChange({ due_date: e.target.value })}
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
