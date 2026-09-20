import type { FormEvent } from 'react'
import { useLanguage, type TranslationKey } from '@/shared/i18n/language'
import { Button } from '@/shared/components/ui/button'
import { Combobox } from '@/shared/components/ui/combobox'
import { DatePickerField } from '@/shared/components/date-picker-field'
import {
  Sheet,
  SheetContent,
  SheetDescription,
  SheetFooter,
  SheetHeader,
  SheetTitle,
} from '@/shared/components/ui/sheet'
import type { ApiTenant } from '@/features/tenant/types'
import { TenantSearchSelect } from '@/features/tenant/components/TenantSearchSelect'
import type { ApiRoom } from '@/features/room/types'
import { RoomSearchSelect } from '@/features/room/components/RoomSearchSelect'
import type { ParcelStatus } from '../types'
import { PARCEL_STATUSES } from '../utils'

const parcelStatusLabelKeys: Record<ParcelStatus, TranslationKey> = {
  pending: 'parcelStatusPending',
  picked_up: 'parcelStatusPickedUp',
  returned: 'parcelStatusReturned',
}

type ParcelFormSheetProps = {
  open: boolean
  onOpenChange: (open: boolean) => void
  isEdit: boolean
  tenantDisplayName: string
  onSelectTenant: (tenant: ApiTenant) => void
  roomDisplayLabel: string
  onSelectRoom: (room: ApiRoom) => void
  onClearRoom: () => void
  courier: string
  onCourierChange: (value: string) => void
  trackingNumber: string
  onTrackingNumberChange: (value: string) => void
  status: ParcelStatus
  onStatusChange: (status: ParcelStatus) => void
  receivedDate: string
  onReceivedDateChange: (value: string) => void
  note: string
  onNoteChange: (value: string) => void
  saving: boolean
  error: string | null
  onSubmit: (event: FormEvent<HTMLFormElement>) => void
}

export function ParcelFormSheet({
  open,
  onOpenChange,
  isEdit,
  tenantDisplayName,
  onSelectTenant,
  roomDisplayLabel,
  onSelectRoom,
  onClearRoom,
  courier,
  onCourierChange,
  trackingNumber,
  onTrackingNumberChange,
  status,
  onStatusChange,
  receivedDate,
  onReceivedDateChange,
  note,
  onNoteChange,
  saving,
  error,
  onSubmit,
}: ParcelFormSheetProps) {
  const { t } = useLanguage()

  return (
    <Sheet open={open} onOpenChange={onOpenChange}>
      <SheetContent>
        <form className="flex h-full flex-col gap-4" onSubmit={onSubmit}>
          <SheetHeader>
            <SheetTitle>{isEdit ? t('parcelFormEditTitle') : t('parcelFormCreateTitle')}</SheetTitle>
            <SheetDescription>
              {isEdit ? t('parcelFormEditDescription') : t('parcelFormCreateDescription')}
            </SheetDescription>
          </SheetHeader>

          <div className="flex flex-1 flex-col gap-4 overflow-y-auto pr-1">
            {isEdit ? (
              <label className="flex flex-col gap-1.5 text-sm font-medium">
                {t('parcelFormTenantLabel')}
                <p className="text-sm font-normal text-muted-foreground">{tenantDisplayName || '—'}</p>
              </label>
            ) : (
              <label className="flex flex-col gap-1.5 text-sm font-medium">
                {t('parcelFormTenantLabel')}
                <TenantSearchSelect
                  selectedLabel={tenantDisplayName}
                  onSelectTenant={onSelectTenant}
                  placeholder={t('parcelFormTenantPlaceholder')}
                  searchPlaceholder={t('pickerSearchTenantPlaceholder')}
                  noResultsLabel={t('pickerNoTenants')}
                />
              </label>
            )}

            <label className="flex flex-col gap-1.5 text-sm font-medium">
              {t('parcelFormRoomLabel')}
              <RoomSearchSelect
                selectedLabel={roomDisplayLabel}
                onSelectRoom={onSelectRoom}
                clearLabel={t('parcelFormRoomPlaceholder')}
                onClear={onClearRoom}
                placeholder={t('parcelFormRoomPlaceholder')}
                searchPlaceholder={t('pickerSearchRoomPlaceholder')}
                noResultsLabel={t('pickerNoRooms')}
              />
              <span className="text-xs font-normal text-muted-foreground">{t('parcelFormRoomHint')}</span>
            </label>

            <label className="flex flex-col gap-1.5 text-sm font-medium">
              {t('parcelFormCourierLabel')}
              <input
                type="text"
                className="h-10 rounded-md border border-input bg-transparent px-3 text-sm"
                value={courier}
                onChange={(event) => onCourierChange(event.target.value)}
              />
            </label>

            <label className="flex flex-col gap-1.5 text-sm font-medium">
              {t('parcelFormTrackingNumberLabel')}
              <input
                type="text"
                className="h-10 rounded-md border border-input bg-transparent px-3 text-sm"
                value={trackingNumber}
                onChange={(event) => onTrackingNumberChange(event.target.value)}
              />
            </label>

            <label className="flex flex-col gap-1.5 text-sm font-medium">
              {t('parcelFormStatusLabel')}
              <Combobox
                options={PARCEL_STATUSES.map((value) => ({
                  value,
                  label: t(parcelStatusLabelKeys[value]),
                }))}
                value={status}
                onChange={(value) => onStatusChange(value as ParcelStatus)}
                showCheck={false}
                placeholder={t('parcelFormStatusPlaceholder')}
                searchPlaceholder={t('parcelFormStatusSearchPlaceholder')}
                emptyText={t('parcelFormStatusNoResults')}
              />
            </label>

            <label className="flex flex-col gap-1.5 text-sm font-medium">
              {t('parcelFormReceivedDateLabel')}
              <DatePickerField
                value={receivedDate}
                onChange={onReceivedDateChange}
                placeholder={t('parcelFormReceivedDateLabel')}
              />
            </label>

            <label className="flex flex-col gap-1.5 text-sm font-medium">
              {t('parcelFormNoteLabel')}
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
              {t('parcelFormCancel')}
            </Button>
            <Button type="submit" disabled={saving}>
              {saving ? t('parcelSaving') : t('parcelFormSave')}
            </Button>
          </SheetFooter>
        </form>
      </SheetContent>
    </Sheet>
  )
}
