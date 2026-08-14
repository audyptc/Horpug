import type { FormEvent } from 'react'
import { useLanguage, type TranslationKey } from '@/shared/i18n/language'
import { Button } from '@/shared/components/ui/button'
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
  tenantId: string
  onTenantIdChange: (tenantId: string) => void
  tenants: ApiTenant[]
  tenantDisplayName: string
  roomId: string
  onRoomIdChange: (roomId: string) => void
  rooms: ApiRoom[]
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
  tenantId,
  onTenantIdChange,
  tenants,
  tenantDisplayName,
  roomId,
  onRoomIdChange,
  rooms,
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
                {tenants.length === 0 ? (
                  <p className="text-xs font-normal text-muted-foreground">{t('parcelFormNoTenants')}</p>
                ) : (
                  <select
                    className="h-10 rounded-md border border-input bg-transparent px-3 text-sm"
                    value={tenantId}
                    onChange={(event) => onTenantIdChange(event.target.value)}
                  >
                    <option value="">{t('parcelFormTenantPlaceholder')}</option>
                    {tenants.map((tenant) => (
                      <option key={tenant.id} value={tenant.id}>
                        {tenant.first_name} {tenant.last_name}
                      </option>
                    ))}
                  </select>
                )}
              </label>
            )}

            <label className="flex flex-col gap-1.5 text-sm font-medium">
              {t('parcelFormRoomLabel')}
              {rooms.length === 0 ? (
                <p className="text-xs font-normal text-muted-foreground">{t('parcelFormNoRooms')}</p>
              ) : (
                <select
                  className="h-10 rounded-md border border-input bg-transparent px-3 text-sm"
                  value={roomId}
                  onChange={(event) => onRoomIdChange(event.target.value)}
                >
                  <option value="">{t('parcelFormRoomPlaceholder')}</option>
                  {rooms.map((room) => (
                    <option key={room.id} value={room.id}>
                      {room.room_number} {room.dormitory_name ? `(${room.dormitory_name})` : ''}
                    </option>
                  ))}
                </select>
              )}
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
              <select
                className="h-10 rounded-md border border-input bg-transparent px-3 text-sm"
                value={status}
                onChange={(event) => onStatusChange(event.target.value as ParcelStatus)}
              >
                {PARCEL_STATUSES.map((value) => (
                  <option key={value} value={value}>
                    {t(parcelStatusLabelKeys[value])}
                  </option>
                ))}
              </select>
            </label>

            <label className="flex flex-col gap-1.5 text-sm font-medium">
              {t('parcelFormReceivedDateLabel')}
              <input
                type="date"
                className="h-10 rounded-md border border-input bg-transparent px-3 text-sm"
                value={receivedDate}
                onChange={(event) => onReceivedDateChange(event.target.value)}
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
