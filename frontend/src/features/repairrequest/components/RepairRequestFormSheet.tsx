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
import type { ApiRoom } from '@/features/room/types'
import { RoomSearchSelect } from '@/features/room/components/RoomSearchSelect'
import type { ApiTenant } from '@/features/tenant/types'
import { TenantSearchSelect } from '@/features/tenant/components/TenantSearchSelect'
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
  roomDisplayLabel: string
  onSelectRoom: (room: ApiRoom) => void
  tenantDisplayName: string
  onSelectTenant: (tenant: ApiTenant) => void
  onClearTenant: () => void
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
  roomDisplayLabel,
  onSelectRoom,
  tenantDisplayName,
  onSelectTenant,
  onClearTenant,
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
                <RoomSearchSelect
                  selectedLabel={roomDisplayLabel}
                  onSelectRoom={onSelectRoom}
                  placeholder={t('repairFormRoomPlaceholder')}
                  searchPlaceholder={t('pickerSearchRoomPlaceholder')}
                  noResultsLabel={t('pickerNoRooms')}
                />
              </label>
            )}

            <label className="flex flex-col gap-1.5 text-sm font-medium">
              {t('repairFormTenantLabel')}
              <TenantSearchSelect
                selectedLabel={tenantDisplayName}
                onSelectTenant={onSelectTenant}
                clearLabel={t('repairFormTenantPlaceholder')}
                onClear={onClearTenant}
                placeholder={t('repairFormTenantPlaceholder')}
                searchPlaceholder={t('pickerSearchTenantPlaceholder')}
                noResultsLabel={t('pickerNoTenants')}
              />
            </label>

            <label className="flex flex-col gap-1.5 text-sm font-medium">
              {t('repairFormCategoryLabel')}
              <Combobox
                options={REPAIR_CATEGORIES.map((value) => ({
                  value,
                  label: t(repairCategoryLabelKeys[value]),
                }))}
                value={category}
                onChange={(value) => onCategoryChange(value as RepairCategory)}
                showCheck={false}
                placeholder={t('repairFormCategoryPlaceholder')}
                searchPlaceholder={t('repairFormCategorySearchPlaceholder')}
                emptyText={t('repairFormCategoryNoResults')}
              />
            </label>

            <label className="flex flex-col gap-1.5 text-sm font-medium">
              {t('repairFormStatusLabel')}
              <Combobox
                options={REPAIR_STATUSES.map((value) => ({
                  value,
                  label: t(repairStatusLabelKeys[value]),
                }))}
                value={status}
                onChange={(value) => onStatusChange(value as RepairStatus)}
                showCheck={false}
                placeholder={t('repairFormStatusPlaceholder')}
                searchPlaceholder={t('repairFormStatusSearchPlaceholder')}
                emptyText={t('repairFormStatusNoResults')}
              />
            </label>

            <label className="flex flex-col gap-1.5 text-sm font-medium">
              {t('repairFormReportedDateLabel')}
              <DatePickerField
                value={reportedDate}
                onChange={onReportedDateChange}
                placeholder={t('repairFormReportedDateLabel')}
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
