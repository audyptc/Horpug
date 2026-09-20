import { type FormEvent } from 'react'
import { useLanguage } from '@/shared/i18n/language'
import { Button } from '@/shared/components/ui/button'
import { DatePickerField } from '@/shared/components/date-picker-field'
import { Switch } from '@/shared/components/ui/switch'
import {
  Sheet,
  SheetContent,
  SheetDescription,
  SheetFooter,
  SheetHeader,
  SheetTitle,
} from '@/shared/components/ui/sheet'
import type { ApiRoom } from '@/features/room/types'
import { RoomSearchSelect } from '@/features/room/components/RoomSearchSelect'
import type { BillingMethod } from '../types'

type WaterMeterFormSheetProps = {
  open: boolean
  onOpenChange: (open: boolean) => void
  isEdit: boolean
  roomDisplayLabel: string
  onSelectRoom: (room: ApiRoom) => void
  billingMethod: BillingMethod
  onBillingMethodChange: (method: BillingMethod) => void
  readingDate: string
  onReadingDateChange: (value: string) => void
  previousUnit: string
  onPreviousUnitChange: (value: string) => void
  currentUnit: string
  onCurrentUnitChange: (value: string) => void
  pricePerUnit: string
  onPricePerUnitChange: (value: string) => void
  flatAmount: string
  onFlatAmountChange: (value: string) => void
  note: string
  onNoteChange: (value: string) => void
  saving: boolean
  error: string | null
  onSubmit: (event: FormEvent<HTMLFormElement>) => void
}

export function WaterMeterFormSheet({
  open,
  onOpenChange,
  isEdit,
  roomDisplayLabel,
  onSelectRoom,
  billingMethod,
  onBillingMethodChange,
  readingDate,
  onReadingDateChange,
  previousUnit,
  onPreviousUnitChange,
  currentUnit,
  onCurrentUnitChange,
  pricePerUnit,
  onPricePerUnitChange,
  flatAmount,
  onFlatAmountChange,
  note,
  onNoteChange,
  saving,
  error,
  onSubmit,
}: WaterMeterFormSheetProps) {
  const { t } = useLanguage()

  return (
    <Sheet open={open} onOpenChange={onOpenChange}>
      <SheetContent>
        <form className="flex h-full flex-col gap-4" onSubmit={onSubmit}>
          <SheetHeader>
            <SheetTitle>{isEdit ? t('waterMeterFormEditTitle') : t('waterMeterFormCreateTitle')}</SheetTitle>
            <SheetDescription>
              {isEdit ? t('waterMeterFormEditDescription') : t('waterMeterFormCreateDescription')}
            </SheetDescription>
          </SheetHeader>

          <div className="flex flex-1 flex-col gap-4 overflow-y-auto pr-1">
            {isEdit ? (
              <label className="flex flex-col gap-1.5 text-sm font-medium">
                {t('waterMeterFormRoomLabel')}
                <p className="text-sm font-normal text-muted-foreground">{roomDisplayLabel || '—'}</p>
              </label>
            ) : (
              <label className="flex flex-col gap-1.5 text-sm font-medium">
                {t('waterMeterFormRoomLabel')}
                <RoomSearchSelect
                  selectedLabel={roomDisplayLabel}
                  onSelectRoom={onSelectRoom}
                  placeholder={t('waterMeterFormRoomPlaceholder')}
                  searchPlaceholder={t('waterMeterFormRoomSearchPlaceholder')}
                  noResultsLabel={t('waterMeterFormRoomNoResults')}
                />
              </label>
            )}

            <label className="flex flex-col gap-1.5 text-sm font-medium">
              {t('waterMeterFormReadingDateLabel')}
              <DatePickerField
                value={readingDate}
                onChange={onReadingDateChange}
                placeholder={t('waterMeterFormReadingDateLabel')}
              />
            </label>

            <label className="flex items-center gap-2 text-sm font-medium">
              {t('waterMeterFormBillingMethodLabel')}
              <span className="flex items-center gap-2 font-normal text-muted-foreground">
                <Switch
                  checked={billingMethod === 'metered'}
                  onCheckedChange={(checked) => onBillingMethodChange(checked ? 'metered' : 'flat')}
                />
                {t(billingMethod === 'metered' ? 'waterMeterBillingMethodMetered' : 'waterMeterBillingMethodFlat')}
              </span>
            </label>

            {billingMethod === 'metered' ? (
              <>
                <label className="flex flex-col gap-1.5 text-sm font-medium">
                  {t('waterMeterFormPreviousUnitLabel')}
                  <input
                    type="number"
                    step="0.01"
                    min="0"
                    className="h-10 rounded-md border border-input bg-transparent px-3 text-right text-sm"
                    value={previousUnit}
                    onChange={(event) => onPreviousUnitChange(event.target.value)}
                  />
                </label>

                <label className="flex flex-col gap-1.5 text-sm font-medium">
                  {t('waterMeterFormCurrentUnitLabel')}
                  <input
                    type="number"
                    step="0.01"
                    min="0"
                    className="h-10 rounded-md border border-input bg-transparent px-3 text-right text-sm"
                    value={currentUnit}
                    onChange={(event) => onCurrentUnitChange(event.target.value)}
                  />
                </label>

                <label className="flex flex-col gap-1.5 text-sm font-medium">
                  {t('waterMeterFormPricePerUnitLabel')}
                  <input
                    type="number"
                    step="0.01"
                    min="0"
                    className="h-10 rounded-md border border-input bg-transparent px-3 text-right text-sm"
                    value={pricePerUnit}
                    onChange={(event) => onPricePerUnitChange(event.target.value)}
                  />
                </label>
              </>
            ) : (
              <label className="flex flex-col gap-1.5 text-sm font-medium">
                {t('waterMeterFormFlatAmountLabel')}
                <input
                  type="number"
                  step="0.01"
                  min="0"
                  className="h-10 rounded-md border border-input bg-transparent px-3 text-right text-sm"
                  value={flatAmount}
                  onChange={(event) => onFlatAmountChange(event.target.value)}
                />
                <span className="text-xs font-normal text-muted-foreground">
                  {t('waterMeterFormFlatAmountHint')}
                </span>
              </label>
            )}

            <label className="flex flex-col gap-1.5 text-sm font-medium">
              {t('waterMeterFormNoteLabel')}
              <textarea
                className="min-h-20 rounded-md border border-input bg-transparent px-3 py-2 text-sm"
                value={note}
                onChange={(event) => onNoteChange(event.target.value)}
              />
            </label>
          </div>

          {error && <p className="resource-error">{error}</p>}

          <SheetFooter>
            <Button type="button" variant="outline" onClick={() => onOpenChange(false)}>
              {t('waterMeterFormCancel')}
            </Button>
            <Button type="submit" disabled={saving}>
              {saving ? t('waterMeterSaving') : t('waterMeterFormSave')}
            </Button>
          </SheetFooter>
        </form>
      </SheetContent>
    </Sheet>
  )
}
