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
import type { VehicleType } from '../types'
import { VEHICLE_TYPES } from '../utils'

const vehicleTypeLabelKeys: Record<VehicleType, TranslationKey> = {
  car: 'parkingVehicleTypeCar',
  motorcycle: 'parkingVehicleTypeMotorcycle',
  other: 'parkingVehicleTypeOther',
}

type ParkingFormSheetProps = {
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
  vehicleType: VehicleType
  onVehicleTypeChange: (vehicleType: VehicleType) => void
  licensePlate: string
  onLicensePlateChange: (value: string) => void
  parkingSpot: string
  onParkingSpotChange: (value: string) => void
  saving: boolean
  error: string | null
  onSubmit: (event: FormEvent<HTMLFormElement>) => void
}

export function ParkingFormSheet({
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
  vehicleType,
  onVehicleTypeChange,
  licensePlate,
  onLicensePlateChange,
  parkingSpot,
  onParkingSpotChange,
  saving,
  error,
  onSubmit,
}: ParkingFormSheetProps) {
  const { t } = useLanguage()

  return (
    <Sheet open={open} onOpenChange={onOpenChange}>
      <SheetContent>
        <form className="flex h-full flex-col gap-4" onSubmit={onSubmit}>
          <SheetHeader>
            <SheetTitle>{isEdit ? t('parkingFormEditTitle') : t('parkingFormCreateTitle')}</SheetTitle>
            <SheetDescription>
              {isEdit ? t('parkingFormEditDescription') : t('parkingFormCreateDescription')}
            </SheetDescription>
          </SheetHeader>

          <div className="flex flex-1 flex-col gap-4 overflow-y-auto pr-1">
            {isEdit ? (
              <label className="flex flex-col gap-1.5 text-sm font-medium">
                {t('parkingFormTenantLabel')}
                <p className="text-sm font-normal text-muted-foreground">{tenantDisplayName || '—'}</p>
              </label>
            ) : (
              <label className="flex flex-col gap-1.5 text-sm font-medium">
                {t('parkingFormTenantLabel')}
                {tenants.length === 0 ? (
                  <p className="text-xs font-normal text-muted-foreground">{t('parkingFormNoTenants')}</p>
                ) : (
                  <select
                    className="h-10 rounded-md border border-input bg-transparent px-3 text-sm"
                    value={tenantId}
                    onChange={(event) => onTenantIdChange(event.target.value)}
                  >
                    <option value="">{t('parkingFormTenantPlaceholder')}</option>
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
              {t('parkingFormRoomLabel')}
              {rooms.length === 0 ? (
                <p className="text-xs font-normal text-muted-foreground">{t('parkingFormNoRooms')}</p>
              ) : (
                <select
                  className="h-10 rounded-md border border-input bg-transparent px-3 text-sm"
                  value={roomId}
                  onChange={(event) => onRoomIdChange(event.target.value)}
                >
                  <option value="">{t('parkingFormRoomPlaceholder')}</option>
                  {rooms.map((room) => (
                    <option key={room.id} value={room.id}>
                      {room.room_number} {room.dormitory_name ? `(${room.dormitory_name})` : ''}
                    </option>
                  ))}
                </select>
              )}
              <span className="text-xs font-normal text-muted-foreground">{t('parkingFormRoomHint')}</span>
            </label>

            <label className="flex flex-col gap-1.5 text-sm font-medium">
              {t('parkingFormVehicleTypeLabel')}
              <select
                className="h-10 rounded-md border border-input bg-transparent px-3 text-sm"
                value={vehicleType}
                onChange={(event) => onVehicleTypeChange(event.target.value as VehicleType)}
              >
                {VEHICLE_TYPES.map((value) => (
                  <option key={value} value={value}>
                    {t(vehicleTypeLabelKeys[value])}
                  </option>
                ))}
              </select>
            </label>

            <label className="flex flex-col gap-1.5 text-sm font-medium">
              {t('parkingFormLicensePlateLabel')}
              <input
                type="text"
                className="h-10 rounded-md border border-input bg-transparent px-3 text-sm"
                value={licensePlate}
                onChange={(event) => onLicensePlateChange(event.target.value)}
              />
            </label>

            <label className="flex flex-col gap-1.5 text-sm font-medium">
              {t('parkingFormParkingSpotLabel')}
              <input
                type="text"
                className="h-10 rounded-md border border-input bg-transparent px-3 text-sm"
                value={parkingSpot}
                onChange={(event) => onParkingSpotChange(event.target.value)}
              />
            </label>
          </div>

          {error && <p className="resource-error">{error}</p>}

          <SheetFooter>
            <Button type="button" variant="outline" onClick={() => onOpenChange(false)}>
              {t('parkingFormCancel')}
            </Button>
            <Button type="submit" disabled={saving}>
              {saving ? t('parkingSaving') : t('parkingFormSave')}
            </Button>
          </SheetFooter>
        </form>
      </SheetContent>
    </Sheet>
  )
}
