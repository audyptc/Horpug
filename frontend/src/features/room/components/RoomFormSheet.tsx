import type { FormEvent } from 'react'
import { useLanguage, type TranslationKey } from '@/shared/i18n/language'
import { Button } from '@/shared/components/ui/button'
import { Combobox } from '@/shared/components/ui/combobox'
import {
  Sheet,
  SheetContent,
  SheetDescription,
  SheetFooter,
  SheetHeader,
  SheetTitle,
} from '@/shared/components/ui/sheet'
import type { ApiDormitory } from '@/features/dormitory/types'
import type { ApiRoomType } from '@/features/roomtype/types'
import type { RoomStatus } from '../types'
import { ROOM_STATUSES } from '../utils'

const roomStatusLabelKeys: Record<RoomStatus, TranslationKey> = {
  available: 'roomStatusAvailable',
  occupied: 'roomStatusOccupied',
  maintenance: 'roomStatusMaintenance',
}

type RoomFormSheetProps = {
  open: boolean
  onOpenChange: (open: boolean) => void
  isEdit: boolean
  dormitoryId: string
  onDormitoryIdChange: (dormitoryId: string) => void
  dormitories: ApiDormitory[]
  roomTypeId: string
  onRoomTypeIdChange: (roomTypeId: string) => void
  roomTypes: ApiRoomType[]
  roomNumber: string
  onRoomNumberChange: (roomNumber: string) => void
  floor: string
  onFloorChange: (floor: string) => void
  status: RoomStatus
  onStatusChange: (status: RoomStatus) => void
  isActive: boolean
  onIsActiveChange: (isActive: boolean) => void
  saving: boolean
  error: string | null
  onSubmit: (event: FormEvent<HTMLFormElement>) => void
}

export function RoomFormSheet({
  open,
  onOpenChange,
  isEdit,
  dormitoryId,
  onDormitoryIdChange,
  dormitories,
  roomTypeId,
  onRoomTypeIdChange,
  roomTypes,
  roomNumber,
  onRoomNumberChange,
  floor,
  onFloorChange,
  status,
  onStatusChange,
  isActive,
  onIsActiveChange,
  saving,
  error,
  onSubmit,
}: RoomFormSheetProps) {
  const { t } = useLanguage()

  return (
    <Sheet open={open} onOpenChange={onOpenChange}>
      <SheetContent>
        <form className="flex h-full flex-col gap-4" onSubmit={onSubmit}>
          <SheetHeader>
            <SheetTitle>{isEdit ? t('roomFormEditTitle') : t('roomFormCreateTitle')}</SheetTitle>
            <SheetDescription>
              {isEdit ? t('roomFormEditDescription') : t('roomFormCreateDescription')}
            </SheetDescription>
          </SheetHeader>

          <div className="flex flex-1 flex-col gap-4 overflow-y-auto pr-1">
            <label className="flex flex-col gap-1.5 text-sm font-medium">
              {t('roomFormDormitoryLabel')}
              {dormitories.length === 0 ? (
                <p className="text-xs font-normal text-muted-foreground">
                  {t('roomFormNoDormitories')}
                </p>
              ) : (
                <Combobox
                  options={dormitories.map((dormitory) => ({
                    value: dormitory.id,
                    label: dormitory.name,
                  }))}
                  value={dormitoryId}
                  onChange={onDormitoryIdChange}
                  placeholder={t('roomFormDormitoryPlaceholder')}
                  searchPlaceholder={t('roomFormDormitorySearchPlaceholder')}
                  emptyText={t('roomFormDormitoryNoResults')}
                  disabled={isEdit}
                />
              )}
            </label>

            <label className="flex flex-col gap-1.5 text-sm font-medium">
              {t('roomFormRoomTypeLabel')}
              {roomTypes.length === 0 ? (
                <p className="text-xs font-normal text-muted-foreground">
                  {t('roomFormNoRoomTypes')}
                </p>
              ) : (
                <Combobox
                  options={roomTypes.map((roomType) => ({
                    value: roomType.id,
                    label: roomType.name,
                  }))}
                  value={roomTypeId}
                  onChange={onRoomTypeIdChange}
                  placeholder={t('roomFormRoomTypePlaceholder')}
                  searchPlaceholder={t('roomFormRoomTypeSearchPlaceholder')}
                  emptyText={t('roomFormRoomTypeNoResults')}
                  disabled={!dormitoryId}
                />
              )}
            </label>

            <label className="flex flex-col gap-1.5 text-sm font-medium">
              {t('roomFormNumberLabel')}
              <input
                type="text"
                className="h-10 rounded-md border border-input bg-transparent px-3 text-sm"
                value={roomNumber}
                onChange={(event) => onRoomNumberChange(event.target.value)}
                autoFocus
              />
            </label>

            <label className="flex flex-col gap-1.5 text-sm font-medium">
              {t('roomFormFloorLabel')}
              <input
                type="number"
                step="1"
                className="h-10 rounded-md border border-input bg-transparent px-3 text-sm"
                value={floor}
                onChange={(event) => onFloorChange(event.target.value)}
              />
            </label>

            <label className="flex flex-col gap-1.5 text-sm font-medium">
              {t('roomFormStatusLabel')}
              <Combobox
                options={ROOM_STATUSES.map((value) => ({
                  value,
                  label: t(roomStatusLabelKeys[value]),
                }))}
                value={status}
                onChange={(value) => onStatusChange(value as RoomStatus)}
                placeholder={t('roomFormStatusLabel')}
                searchPlaceholder={t('roomFormStatusSearchPlaceholder')}
                emptyText={t('roomFormStatusNoResults')}
              />
            </label>

            <label className="flex items-center gap-2 text-sm font-medium">
              <input
                type="checkbox"
                className="h-4 w-4 accent-primary"
                checked={isActive}
                onChange={(event) => onIsActiveChange(event.target.checked)}
              />
              {t('roomFormActiveLabel')}
            </label>
          </div>

          {error && <p className="resource-error">{error}</p>}

          <SheetFooter>
            <Button type="button" variant="outline" onClick={() => onOpenChange(false)}>
              {t('roomFormCancel')}
            </Button>
            <Button type="submit" disabled={saving}>
              {saving ? t('roomSaving') : t('roomFormSave')}
            </Button>
          </SheetFooter>
        </form>
      </SheetContent>
    </Sheet>
  )
}
