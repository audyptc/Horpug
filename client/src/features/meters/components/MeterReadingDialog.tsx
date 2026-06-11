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
import type { ApiMeterReading, ApiRoom, MeterType } from '@/types'

type MeterReadingForm = {
  room_id: string
  meter_type: MeterType
  reading_date: string
  previous_reading: string
  current_reading: string
  unit_price: string
  note: string
}

type Props = {
  open: boolean
  onOpenChange: (open: boolean) => void
  editingReading: ApiMeterReading | null
  form: MeterReadingForm
  onFormChange: (form: MeterReadingForm) => void
  onSave: () => void
  saving: boolean
  rooms: ApiRoom[]
}

export function MeterReadingDialog({
  open,
  onOpenChange,
  editingReading,
  form,
  onFormChange,
  onSave,
  saving,
  rooms,
}: Props) {
  const { t } = useTranslation()

  const unitUsed =
    form.current_reading && form.previous_reading
      ? Number(form.current_reading) - Number(form.previous_reading)
      : null

  const totalAmount =
    unitUsed !== null && form.unit_price ? unitUsed * Number(form.unit_price) : null

  const isSaveDisabled =
    saving ||
    !form.room_id ||
    !form.reading_date ||
    !form.current_reading ||
    !form.unit_price ||
    Number(form.current_reading) < Number(form.previous_reading)

  function set(patch: Partial<MeterReadingForm>) {
    onFormChange({ ...form, ...patch })
  }

  return (
    <Dialog open={open} onOpenChange={onOpenChange}>
      <DialogContent className="max-w-lg">
        <DialogHeader>
          <DialogTitle>
            {editingReading ? t('meters.editReading') : t('meters.createReading')}
          </DialogTitle>
          <DialogDescription>
            {editingReading ? t('meters.editDesc') : t('meters.createDesc')}
          </DialogDescription>
        </DialogHeader>

        <div className="grid gap-4 py-4">
          <div className="space-y-1.5">
            <Label>{t('meters.room')} *</Label>
            <Select
              value={form.room_id}
              onValueChange={(v) => set({ room_id: v })}
              disabled={!!editingReading}
            >
              <SelectTrigger><SelectValue placeholder={t('meters.selectRoom')} /></SelectTrigger>
              <SelectContent>
                {rooms.map((rm) => (
                  <SelectItem key={rm.id} value={rm.id}>
                    {rm.room_number}
                  </SelectItem>
                ))}
              </SelectContent>
            </Select>
          </div>

          <div className="space-y-1.5">
            <Label>{t('meters.meterType')} *</Label>
            <Select
              value={form.meter_type}
              onValueChange={(v) => set({ meter_type: v as MeterType })}
              disabled={!!editingReading}
            >
              <SelectTrigger><SelectValue /></SelectTrigger>
              <SelectContent>
                <SelectItem value="electric">{t('meters.types.electric')}</SelectItem>
                <SelectItem value="water">{t('meters.types.water')}</SelectItem>
              </SelectContent>
            </Select>
          </div>

          <div className="space-y-1.5">
            <Label>{t('meters.readingDate')} *</Label>
            <Input
              type="date"
              value={form.reading_date}
              onChange={(e) => set({ reading_date: e.target.value })}
            />
          </div>

          <div className="grid grid-cols-2 gap-4">
            <div className="space-y-1.5">
              <Label>{t('meters.previousReading')}</Label>
              <Input
                type="number"
                min="0"
                step="0.01"
                value={form.previous_reading}
                onChange={(e) => set({ previous_reading: e.target.value })}
              />
            </div>
            <div className="space-y-1.5">
              <Label>{t('meters.currentReading')} *</Label>
              <Input
                type="number"
                min="0"
                step="0.01"
                value={form.current_reading}
                onChange={(e) => set({ current_reading: e.target.value })}
              />
            </div>
          </div>

          <div className="space-y-1.5">
            <Label>{t('meters.unitPrice')} *</Label>
            <Input
              type="number"
              min="0"
              step="0.0001"
              value={form.unit_price}
              onChange={(e) => set({ unit_price: e.target.value })}
            />
          </div>

          {unitUsed !== null && unitUsed >= 0 && (
            <div className="rounded-md border bg-muted/40 px-4 py-3 text-sm space-y-1">
              <div className="flex justify-between">
                <span className="text-muted-foreground">{t('meters.unitUsed')}</span>
                <span className="font-medium">{unitUsed.toLocaleString()} {t('meters.unit')}</span>
              </div>
              {totalAmount !== null && (
                <div className="flex justify-between">
                  <span className="text-muted-foreground">{t('meters.totalAmount')}</span>
                  <span className="font-semibold">{totalAmount.toLocaleString()} ฿</span>
                </div>
              )}
            </div>
          )}

          <div className="space-y-1.5">
            <Label>{t('meters.note')}</Label>
            <textarea
              rows={2}
              placeholder={t('meters.notePlaceholder')}
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
            {saving
              ? t('common.loading')
              : editingReading
                ? t('meters.saveChanges')
                : t('meters.createReading')}
          </Button>
        </DialogFooter>
      </DialogContent>
    </Dialog>
  )
}
