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
import { DatePicker } from '@/components/ui/date-picker'
import { cn } from '@/lib/utils'
import type { ApiWaterMeter, ApiRoom, WaterBillingType } from '@/types'

export type WaterMeterForm = {
  room_id: string
  billing_type: WaterBillingType
  reading_date: string
  previous_reading: string
  current_reading: string
  unit_price: string
  flat_amount: string
  note: string
}

type Props = {
  open: boolean
  onOpenChange: (open: boolean) => void
  editing: ApiWaterMeter | null
  form: WaterMeterForm
  onFormChange: (form: WaterMeterForm) => void
  onSave: () => void
  saving: boolean
  rooms: ApiRoom[]
}

export function WaterMeterDialog({ open, onOpenChange, editing, form, onFormChange, onSave, saving, rooms }: Props) {
  const { t } = useTranslation()
  const isMeter = form.billing_type === 'meter'

  const unitUsed = isMeter && form.current_reading && form.previous_reading
    ? Number(form.current_reading) - Number(form.previous_reading)
    : null
  const totalAmount = isMeter
    ? (unitUsed !== null && form.unit_price ? unitUsed * Number(form.unit_price) : null)
    : (form.flat_amount ? Number(form.flat_amount) : null)

  const isSaveDisabled = saving || !form.room_id || !form.reading_date || (
    isMeter
      ? (!form.current_reading || !form.unit_price || Number(form.current_reading) < Number(form.previous_reading))
      : !form.flat_amount
  )

  function set(patch: Partial<WaterMeterForm>) {
    onFormChange({ ...form, ...patch })
  }

  return (
    <Dialog open={open} onOpenChange={onOpenChange}>
      <DialogContent className="max-w-lg">
        <DialogHeader>
          <DialogTitle>{editing ? t('waterMeters.editReading') : t('waterMeters.createReading')}</DialogTitle>
          <DialogDescription>{editing ? t('waterMeters.editDesc') : t('waterMeters.createDesc')}</DialogDescription>
        </DialogHeader>

        <div className="grid gap-4 py-4">
          <div className="space-y-1.5">
            <Label>{t('waterMeters.room')} *</Label>
            <Select value={form.room_id} onValueChange={(v) => set({ room_id: v })} disabled={!!editing}>
              <SelectTrigger><SelectValue placeholder={t('waterMeters.selectRoom')} /></SelectTrigger>
              <SelectContent>
                {rooms.map((rm) => (
                  <SelectItem key={rm.id} value={rm.id}>{rm.room_number}</SelectItem>
                ))}
              </SelectContent>
            </Select>
          </div>

          {/* Billing type toggle */}
          <div className="space-y-1.5">
            <Label>{t('waterMeters.billingType')} *</Label>
            <div className="flex rounded-md border overflow-hidden">
              {(['meter', 'flat'] as WaterBillingType[]).map((type) => (
                <button
                  key={type}
                  type="button"
                  onClick={() => set({ billing_type: type })}
                  className={cn(
                    'flex-1 py-2 text-sm font-medium transition-colors',
                    form.billing_type === type
                      ? 'bg-primary text-primary-foreground'
                      : 'bg-background text-muted-foreground hover:bg-muted'
                  )}
                >
                  {t(`waterMeters.billingTypes.${type}`)}
                </button>
              ))}
            </div>
          </div>

          <div className="space-y-1.5">
            <Label>{t('waterMeters.readingDate')} *</Label>
            <DatePicker value={form.reading_date} onChange={(v) => set({ reading_date: v })} />
          </div>

          {isMeter ? (
            <>
              <div className="grid grid-cols-2 gap-4">
                <div className="space-y-1.5">
                  <Label>{t('waterMeters.previousReading')}</Label>
                  <Input type="number" min="0" step="0.01" value={form.previous_reading}
                    onChange={(e) => set({ previous_reading: e.target.value })} />
                </div>
                <div className="space-y-1.5">
                  <Label>{t('waterMeters.currentReading')} *</Label>
                  <Input type="number" min="0" step="0.01" value={form.current_reading}
                    onChange={(e) => set({ current_reading: e.target.value })} />
                </div>
              </div>
              <div className="space-y-1.5">
                <Label>{t('waterMeters.unitPrice')} *</Label>
                <Input type="number" min="0" step="0.0001" value={form.unit_price}
                  onChange={(e) => set({ unit_price: e.target.value })} />
              </div>
            </>
          ) : (
            <div className="space-y-1.5">
              <Label>{t('waterMeters.flatAmount')} *</Label>
              <Input type="number" min="0" step="0.01" value={form.flat_amount}
                onChange={(e) => set({ flat_amount: e.target.value })} />
            </div>
          )}

          {totalAmount !== null && totalAmount >= 0 && (
            <div className="rounded-md border bg-muted/40 px-4 py-3 text-sm space-y-1">
              {isMeter && unitUsed !== null && (
                <div className="flex justify-between">
                  <span className="text-muted-foreground">{t('waterMeters.unitUsed')}</span>
                  <span className="font-medium">{unitUsed.toLocaleString()} {t('waterMeters.unit')}</span>
                </div>
              )}
              <div className="flex justify-between">
                <span className="text-muted-foreground">{t('waterMeters.totalAmount')}</span>
                <span className="font-semibold">{totalAmount.toLocaleString()} ฿</span>
              </div>
            </div>
          )}

          <div className="space-y-1.5">
            <Label>{t('waterMeters.note')}</Label>
            <textarea rows={2} placeholder={t('waterMeters.notePlaceholder')} value={form.note}
              onChange={(e: React.ChangeEvent<HTMLTextAreaElement>) => set({ note: e.target.value })}
              className="flex min-h-16 w-full rounded-md border border-input bg-background px-3 py-2 text-sm ring-offset-background placeholder:text-muted-foreground focus-visible:outline-none focus-visible:ring-2 focus-visible:ring-ring focus-visible:ring-offset-2 resize-none"
            />
          </div>
        </div>

        <DialogFooter>
          <Button variant="outline" onClick={() => onOpenChange(false)}>{t('common.cancel')}</Button>
          <Button onClick={onSave} disabled={isSaveDisabled}>
            {saving ? t('common.loading') : editing ? t('waterMeters.saveChanges') : t('waterMeters.createReading')}
          </Button>
        </DialogFooter>
      </DialogContent>
    </Dialog>
  )
}
