import { useState, type FormEvent } from 'react'
import { CalendarIcon, X } from 'lucide-react'
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
import type { ApiTenant } from '@/features/tenant/types'
import type { ApiRoom } from '@/features/room/types'
import { RoomSearchSelect } from '@/features/room/components/RoomSearchSelect'
import type { ContractStatus } from '../types'
import { CONTRACT_STATUSES } from '../utils'

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
          title={t('contractFormDateClear')}
          aria-label={t('contractFormDateClear')}
          onClick={() => onChange('')}
        >
          <X className="size-4" />
        </Button>
      )}
    </div>
  )
}

const contractStatusLabelKeys: Record<ContractStatus, TranslationKey> = {
  active: 'contractStatusActive',
  expired: 'contractStatusExpired',
  terminated: 'contractStatusTerminated',
}

type ContractFormSheetProps = {
  open: boolean
  onOpenChange: (open: boolean) => void
  isEdit: boolean
  tenantId: string
  onTenantIdChange: (tenantId: string) => void
  tenants: ApiTenant[]
  tenantDisplayName: string
  onSelectRoom: (room: ApiRoom) => void
  roomDisplayLabel: string
  startDate: string
  onStartDateChange: (value: string) => void
  endDate: string
  onEndDateChange: (value: string) => void
  rentPrice: string
  onRentPriceChange: (value: string) => void
  deposit: string
  onDepositChange: (value: string) => void
  numOccupants: string
  onNumOccupantsChange: (value: string) => void
  status: ContractStatus
  onStatusChange: (status: ContractStatus) => void
  note: string
  onNoteChange: (value: string) => void
  saving: boolean
  error: string | null
  onSubmit: (event: FormEvent<HTMLFormElement>) => void
}

export function ContractFormSheet({
  open,
  onOpenChange,
  isEdit,
  tenantId,
  onTenantIdChange,
  tenants,
  tenantDisplayName,
  onSelectRoom,
  roomDisplayLabel,
  startDate,
  onStartDateChange,
  endDate,
  onEndDateChange,
  rentPrice,
  onRentPriceChange,
  deposit,
  onDepositChange,
  numOccupants,
  onNumOccupantsChange,
  status,
  onStatusChange,
  note,
  onNoteChange,
  saving,
  error,
  onSubmit,
}: ContractFormSheetProps) {
  const { t } = useLanguage()

  return (
    <Sheet open={open} onOpenChange={onOpenChange}>
      <SheetContent>
        <form className="flex h-full flex-col gap-4" onSubmit={onSubmit}>
          <SheetHeader>
            <SheetTitle>{isEdit ? t('contractFormEditTitle') : t('contractFormCreateTitle')}</SheetTitle>
            <SheetDescription>
              {isEdit ? t('contractFormEditDescription') : t('contractFormCreateDescription')}
            </SheetDescription>
          </SheetHeader>

          <div className="flex flex-1 flex-col gap-4 overflow-y-auto pr-1">
            {isEdit ? (
              <>
                <label className="flex flex-col gap-1.5 text-sm font-medium">
                  {t('contractFormTenantLabel')}
                  <p className="text-sm font-normal text-muted-foreground">{tenantDisplayName || '—'}</p>
                </label>
                <label className="flex flex-col gap-1.5 text-sm font-medium">
                  {t('contractFormRoomLabel')}
                  <p className="text-sm font-normal text-muted-foreground">{roomDisplayLabel || '—'}</p>
                </label>
                <label className="flex flex-col gap-1.5 text-sm font-medium">
                  {t('contractFormStartDateLabel')}
                  <DatePickerField
                    value={startDate}
                    onChange={onStartDateChange}
                    placeholder={t('contractFormStartDateLabel')}
                  />
                </label>
              </>
            ) : (
              <>
                <label className="flex flex-col gap-1.5 text-sm font-medium">
                  {t('contractFormTenantLabel')}
                  {tenants.length === 0 ? (
                    <p className="text-xs font-normal text-muted-foreground">{t('contractFormNoTenants')}</p>
                  ) : (
                    <Combobox
                      options={tenants.map((tenant) => ({
                        value: tenant.id,
                        label: `${tenant.first_name} ${tenant.last_name}`,
                      }))}
                      value={tenantId}
                      onChange={onTenantIdChange}
                      placeholder={t('contractFormTenantPlaceholder')}
                      searchPlaceholder={t('contractFormTenantSearchPlaceholder')}
                      emptyText={t('contractFormTenantNoResults')}
                    />
                  )}
                </label>

                <label className="flex flex-col gap-1.5 text-sm font-medium">
                  {t('contractFormRoomLabel')}
                  <RoomSearchSelect
                    selectedLabel={roomDisplayLabel}
                    onSelectRoom={onSelectRoom}
                    statusFilter="available"
                    placeholder={t('contractFormRoomPlaceholder')}
                    searchPlaceholder={t('contractFormRoomSearchPlaceholder')}
                    noResultsLabel={t('contractFormNoRooms')}
                  />
                </label>

                <label className="flex flex-col gap-1.5 text-sm font-medium">
                  {t('contractFormStartDateLabel')}
                  <DatePickerField
                    value={startDate}
                    onChange={onStartDateChange}
                    placeholder={t('contractFormStartDateLabel')}
                  />
                </label>
              </>
            )}

            <label className="flex flex-col gap-1.5 text-sm font-medium">
              {t('contractFormEndDateLabel')}
              <DatePickerField
                value={endDate}
                onChange={onEndDateChange}
                placeholder={t('contractFormEndDateLabel')}
              />
            </label>

            <label className="flex flex-col gap-1.5 text-sm font-medium">
              {t('contractFormRentPriceLabel')}
              <input
                type="number"
                step="0.01"
                min="0"
                className="h-10 rounded-md border border-input bg-transparent px-3 text-right text-sm"
                value={rentPrice}
                onChange={(event) => onRentPriceChange(event.target.value)}
              />
            </label>

            <label className="flex flex-col gap-1.5 text-sm font-medium">
              {t('contractFormDepositLabel')}
              <input
                type="number"
                step="0.01"
                min="0"
                className="h-10 rounded-md border border-input bg-transparent px-3 text-right text-sm"
                value={deposit}
                onChange={(event) => onDepositChange(event.target.value)}
              />
            </label>

            <label className="flex flex-col gap-1.5 text-sm font-medium">
              {t('contractFormNumOccupantsLabel')}
              <input
                type="number"
                step="1"
                min="1"
                className="h-10 rounded-md border border-input bg-transparent px-3 text-sm"
                value={numOccupants}
                onChange={(event) => onNumOccupantsChange(event.target.value)}
              />
            </label>

            {isEdit && (
              <label className="flex flex-col gap-1.5 text-sm font-medium">
                {t('contractFormStatusLabel')}
                <Combobox
                  options={CONTRACT_STATUSES.map((value) => ({
                    value,
                    label: t(contractStatusLabelKeys[value]),
                  }))}
                  value={status}
                  onChange={(value) => onStatusChange(value as ContractStatus)}
                  placeholder={t('contractFormStatusLabel')}
                  searchPlaceholder={t('contractFormStatusSearchPlaceholder')}
                  emptyText={t('contractFormStatusNoResults')}
                />
              </label>
            )}

            <label className="flex flex-col gap-1.5 text-sm font-medium">
              {t('contractFormNoteLabel')}
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
              {t('contractFormCancel')}
            </Button>
            <Button type="submit" disabled={saving}>
              {saving ? t('contractSaving') : t('contractFormSave')}
            </Button>
          </SheetFooter>
        </form>
      </SheetContent>
    </Sheet>
  )
}
