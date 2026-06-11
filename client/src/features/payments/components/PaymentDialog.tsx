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
import type { ApiPayment, PaymentMethod } from '@/types'

type PaymentForm = {
  bill_id: string
  amount: string
  method: PaymentMethod
  payment_date: string
  note: string
}

type Props = {
  open: boolean
  onOpenChange: (open: boolean) => void
  editingItem: ApiPayment | null
  form: PaymentForm
  onFormChange: (form: PaymentForm) => void
  onSave: () => void
  saving: boolean
}

export function PaymentDialog({
  open,
  onOpenChange,
  editingItem,
  form,
  onFormChange,
  onSave,
  saving,
}: Props) {
  const { t } = useTranslation()

  const isSaveDisabled = saving || !form.bill_id || !form.amount || Number(form.amount) <= 0 || !form.payment_date

  function set(patch: Partial<PaymentForm>) {
    onFormChange({ ...form, ...patch })
  }

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
              <Label>{t('payments.billId')} *</Label>
              <Input
                placeholder={t('payments.billIdPlaceholder')}
                value={form.bill_id}
                onChange={(e) => set({ bill_id: e.target.value })}
              />
            </div>
          )}

          <div className="grid grid-cols-2 gap-4">
            <div className="space-y-1.5">
              <Label>{t('payments.amount')} *</Label>
              <Input
                type="number"
                min="0"
                step="0.01"
                placeholder="0.00"
                value={form.amount}
                onChange={(e) => set({ amount: e.target.value })}
              />
            </div>
            <div className="space-y-1.5">
              <Label>{t('payments.method')} *</Label>
              <Select
                value={form.method}
                onValueChange={(v) => set({ method: v as PaymentMethod })}
              >
                <SelectTrigger><SelectValue /></SelectTrigger>
                <SelectContent>
                  <SelectItem value="cash">{t('payments.methods.cash')}</SelectItem>
                  <SelectItem value="transfer">{t('payments.methods.transfer')}</SelectItem>
                  <SelectItem value="qr">{t('payments.methods.qr')}</SelectItem>
                </SelectContent>
              </Select>
            </div>
          </div>

          <div className="space-y-1.5">
            <Label>{t('payments.paymentDate')} *</Label>
            <Input
              type="date"
              value={form.payment_date}
              onChange={(e) => set({ payment_date: e.target.value })}
            />
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

        <DialogFooter>
          <Button variant="outline" onClick={() => onOpenChange(false)}>
            {t('common.cancel')}
          </Button>
          <Button onClick={onSave} disabled={isSaveDisabled}>
            {saving ? t('common.loading') : editingItem ? t('payments.saveChanges') : t('payments.createPayment')}
          </Button>
        </DialogFooter>
      </DialogContent>
    </Dialog>
  )
}
