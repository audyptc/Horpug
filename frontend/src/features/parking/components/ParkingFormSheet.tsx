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
import { TenantSearchSelect } from '@/features/tenant/components/TenantSearchSelect'
import type { ApiRoom } from '@/features/room/types'
import { RoomSearchSelect } from '@/features/room/components/RoomSearchSelect'
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
  tenantDisplayName: string
  onSelectTenant: (tenant: ApiTenant) => void
  roomDisplayLabel: string
  onSelectRoom: (room: ApiRoom) => void
  onClearRoom: () => void
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
  tenantDisplayName,
  onSelectTenant,
  roomDisplayLabel,
  onSelectRoom,
  onClearRoom,
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
                <TenantSearchSelect
                  selectedLabel={tenantDisplayName}
                  onSelectTenant={onSelectTenant}
                  placeholder={t('parkingFormTenantPlaceholder')}
                  searchPlaceholder={t('pickerSearchTenantPlaceholder')}
                  noResultsLabel={t('pickerNoTenants')}
                />
              </label>
            )}

            <label className="flex flex-col gap-1.5 text-sm font-medium">
              {t('parkingFormRoomLabel')}
              <RoomSearchSelect
                selectedLabel={roomDisplayLabel}
                onSelectRoom={onSelectRoom}
                clearLabel={t('parkingFormRoomPlaceholder')}
                onClear={onClearRoom}
                placeholder={t('parkingFormRoomPlaceholder')}
                searchPlaceholder={t('pickerSearchRoomPlaceholder')}
                noResultsLabel={t('pickerNoRooms')}
              />
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
