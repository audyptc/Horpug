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
import type { ApiRoom } from '@/features/room/types'
import type { ApiTenant } from '@/features/tenant/types'
import type { RepairCategory, RepairStatus } from '../types'
import { REPAIR_CATEGORIES, REPAIR_STATUSES } from '../utils'

const repairCategoryLabelKeys: Record<RepairCategory, TranslationKey> = {
  electrical: 'repairCategoryElectrical',
  plumbing: 'repairCategoryPlumbing',
  furniture: 'repairCategoryFurniture',
  aircon: 'repairCategoryAircon',
  other: 'repairCategoryOther',
}

const repairStatusLabelKeys: Record<RepairStatus, TranslationKey> = {
  pending: 'repairStatusPending',
  in_progress: 'repairStatusInProgress',
  completed: 'repairStatusCompleted',
  cancelled: 'repairStatusCancelled',
}

type RepairRequestFormSheetProps = {
  open: boolean
  onOpenChange: (open: boolean) => void
  isEdit: boolean
  roomId: string
  onRoomIdChange: (roomId: string) => void
  rooms: ApiRoom[]
  roomDisplayLabel: string
  tenantId: string
  onTenantIdChange: (tenantId: string) => void
  tenants: ApiTenant[]
  category: RepairCategory
  onCategoryChange: (category: RepairCategory) => void
  description: string
  onDescriptionChange: (value: string) => void
  status: RepairStatus
  onStatusChange: (status: RepairStatus) => void
  reportedDate: string
  onReportedDateChange: (value: string) => void
  saving: boolean
  error: string | null
  onSubmit: (event: FormEvent<HTMLFormElement>) => void
}

export function RepairRequestFormSheet({
  open,
  onOpenChange,
  isEdit,
  roomId,
  onRoomIdChange,
  rooms,
  roomDisplayLabel,
  tenantId,
  onTenantIdChange,
  tenants,
  category,
  onCategoryChange,
  description,
  onDescriptionChange,
  status,
  onStatusChange,
  reportedDate,
  onReportedDateChange,
  saving,
  error,
  onSubmit,
}: RepairRequestFormSheetProps) {
  const { t } = useLanguage()

  return (
    <Sheet open={open} onOpenChange={onOpenChange}>
      <SheetContent>
        <form className="flex h-full flex-col gap-4" onSubmit={onSubmit}>
          <SheetHeader>
            <SheetTitle>{isEdit ? t('repairFormEditTitle') : t('repairFormCreateTitle')}</SheetTitle>
            <SheetDescription>
              {isEdit ? t('repairFormEditDescription') : t('repairFormCreateDescription')}
            </SheetDescription>
          </SheetHeader>

          <div className="flex flex-1 flex-col gap-4 overflow-y-auto pr-1">
            {isEdit ? (
              <label className="flex flex-col gap-1.5 text-sm font-medium">
                {t('repairFormRoomLabel')}
                <p className="text-sm font-normal text-muted-foreground">{roomDisplayLabel || '—'}</p>
              </label>
            ) : (
              <label className="flex flex-col gap-1.5 text-sm font-medium">
                {t('repairFormRoomLabel')}
                {rooms.length === 0 ? (
                  <p className="text-xs font-normal text-muted-foreground">{t('repairFormNoRooms')}</p>
                ) : (
                  <select
                    className="h-10 rounded-md border border-input bg-transparent px-3 text-sm"
                    value={roomId}
                    onChange={(event) => onRoomIdChange(event.target.value)}
                  >
                    <option value="">{t('repairFormRoomPlaceholder')}</option>
                    {rooms.map((room) => (
                      <option key={room.id} value={room.id}>
                        {room.room_number} {room.dormitory_name ? `(${room.dormitory_name})` : ''}
                      </option>
                    ))}
                  </select>
                )}
              </label>
            )}

            <label className="flex flex-col gap-1.5 text-sm font-medium">
              {t('repairFormTenantLabel')}
              <select
                className="h-10 rounded-md border border-input bg-transparent px-3 text-sm"
                value={tenantId}
                onChange={(event) => onTenantIdChange(event.target.value)}
              >
                <option value="">{t('repairFormTenantPlaceholder')}</option>
                {tenants.map((tenant) => (
                  <option key={tenant.id} value={tenant.id}>
                    {tenant.first_name} {tenant.last_name}
                  </option>
                ))}
              </select>
            </label>

            <label className="flex flex-col gap-1.5 text-sm font-medium">
              {t('repairFormCategoryLabel')}
              <select
                className="h-10 rounded-md border border-input bg-transparent px-3 text-sm"
                value={category}
                onChange={(event) => onCategoryChange(event.target.value as RepairCategory)}
              >
                {REPAIR_CATEGORIES.map((value) => (
                  <option key={value} value={value}>
                    {t(repairCategoryLabelKeys[value])}
                  </option>
                ))}
              </select>
            </label>

            <label className="flex flex-col gap-1.5 text-sm font-medium">
              {t('repairFormStatusLabel')}
              <select
                className="h-10 rounded-md border border-input bg-transparent px-3 text-sm"
                value={status}
                onChange={(event) => onStatusChange(event.target.value as RepairStatus)}
              >
                {REPAIR_STATUSES.map((value) => (
                  <option key={value} value={value}>
                    {t(repairStatusLabelKeys[value])}
                  </option>
                ))}
              </select>
            </label>

            <label className="flex flex-col gap-1.5 text-sm font-medium">
              {t('repairFormReportedDateLabel')}
              <input
                type="date"
                className="h-10 rounded-md border border-input bg-transparent px-3 text-sm"
                value={reportedDate}
                onChange={(event) => onReportedDateChange(event.target.value)}
              />
            </label>

            <label className="flex flex-col gap-1.5 text-sm font-medium">
              {t('repairFormDescriptionLabel')}
              <textarea
                className="min-h-20 rounded-md border border-input bg-transparent px-3 py-2 text-sm"
                value={description}
                onChange={(event) => onDescriptionChange(event.target.value)}
              />
            </label>
          </div>

          {error && <p className="resource-error">{error}</p>}

          <SheetFooter>
            <Button type="button" variant="outline" onClick={() => onOpenChange(false)}>
              {t('repairFormCancel')}
            </Button>
            <Button type="submit" disabled={saving}>
              {saving ? t('repairSaving') : t('repairFormSave')}
            </Button>
          </SheetFooter>
        </form>
      </SheetContent>
    </Sheet>
  )
}
