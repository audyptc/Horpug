import { useState, type FormEvent } from 'react'
import { CalendarIcon, X } from 'lucide-react'
import { useLanguage } from '@/shared/i18n/language'
import { cn } from '@/shared/lib/utils'
import { Button } from '@/shared/components/ui/button'
import { Calendar } from '@/shared/components/ui/calendar'
import { Combobox } from '@/shared/components/ui/combobox'
import { Popover, PopoverContent, PopoverTrigger } from '@/shared/components/ui/popover'
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
import type { BillingMethod } from '../types'

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
          title={t('meterFormDateClear')}
          aria-label={t('meterFormDateClear')}
          onClick={() => onChange('')}
        >
          <X className="size-4" />
        </Button>
      )}
    </div>
  )
}

type MeterFormSheetProps = {
  open: boolean
  onOpenChange: (open: boolean) => void
  isEdit: boolean
  roomId: string
  onRoomIdChange: (roomId: string) => void
  rooms: ApiRoom[]
  roomDisplayLabel: string
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

export function MeterFormSheet({
  open,
  onOpenChange,
  isEdit,
  roomId,
  onRoomIdChange,
  rooms,
  roomDisplayLabel,
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
}: MeterFormSheetProps) {
  const { t } = useLanguage()

  return (
    <Sheet open={open} onOpenChange={onOpenChange}>
      <SheetContent>
        <form className="flex h-full flex-col gap-4" onSubmit={onSubmit}>
          <SheetHeader>
            <SheetTitle>{isEdit ? t('meterFormEditTitle') : t('meterFormCreateTitle')}</SheetTitle>
            <SheetDescription>
              {isEdit ? t('meterFormEditDescription') : t('meterFormCreateDescription')}
            </SheetDescription>
          </SheetHeader>

          <div className="flex flex-1 flex-col gap-4 overflow-y-auto pr-1">
            {isEdit ? (
              <label className="flex flex-col gap-1.5 text-sm font-medium">
                {t('meterFormRoomLabel')}
                <p className="text-sm font-normal text-muted-foreground">{roomDisplayLabel || '—'}</p>
              </label>
            ) : (
              <label className="flex flex-col gap-1.5 text-sm font-medium">
                {t('meterFormRoomLabel')}
                {rooms.length === 0 ? (
                  <p className="text-xs font-normal text-muted-foreground">{t('meterFormNoRooms')}</p>
                ) : (
                  <Combobox
                    options={rooms.map((room) => ({
                      value: room.id,
                      label: `${room.room_number} ${room.dormitory_name ? `(${room.dormitory_name})` : ''}`,
                    }))}
                    value={roomId}
                    onChange={onRoomIdChange}
                    placeholder={t('meterFormRoomPlaceholder')}
                    searchPlaceholder={t('meterFormRoomSearchPlaceholder')}
                    emptyText={t('meterFormRoomNoResults')}
                  />
                )}
              </label>
            )}

            <label className="flex flex-col gap-1.5 text-sm font-medium">
              {t('meterFormReadingDateLabel')}
              <DatePickerField
                value={readingDate}
                onChange={onReadingDateChange}
                placeholder={t('meterFormReadingDateLabel')}
              />
            </label>

            <label className="flex items-center gap-2 text-sm font-medium">
              {t('meterFormBillingMethodLabel')}
              <span className="flex items-center gap-2 font-normal text-muted-foreground">
                <Switch
                  checked={billingMethod === 'metered'}
                  onCheckedChange={(checked) => onBillingMethodChange(checked ? 'metered' : 'flat')}
                />
                {t(billingMethod === 'metered' ? 'meterBillingMethodMetered' : 'meterBillingMethodFlat')}
              </span>
            </label>

            {billingMethod === 'metered' ? (
              <>
                <label className="flex flex-col gap-1.5 text-sm font-medium">
                  {t('meterFormPreviousUnitLabel')}
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
                  {t('meterFormCurrentUnitLabel')}
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
                  {t('meterFormPricePerUnitLabel')}
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
                {t('meterFormFlatAmountLabel')}
                <input
                  type="number"
                  step="0.01"
                  min="0"
                  className="h-10 rounded-md border border-input bg-transparent px-3 text-right text-sm"
                  value={flatAmount}
                  onChange={(event) => onFlatAmountChange(event.target.value)}
                />
                <span className="text-xs font-normal text-muted-foreground">
                  {t('meterFormFlatAmountHint')}
                </span>
              </label>
            )}

            <label className="flex flex-col gap-1.5 text-sm font-medium">
              {t('meterFormNoteLabel')}
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
              {t('meterFormCancel')}
            </Button>
            <Button type="submit" disabled={saving}>
              {saving ? t('meterSaving') : t('meterFormSave')}
            </Button>
          </SheetFooter>
        </form>
      </SheetContent>
    </Sheet>
  )
}
