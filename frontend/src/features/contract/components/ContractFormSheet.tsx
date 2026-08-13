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
import { RoomSearchSelect } from '@/features/room/components/RoomSearchSelect'
import type { ContractStatus } from '../types'
import { CONTRACT_STATUSES } from '../utils'

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
  onClearRoomSelection: () => void
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
  onClearRoomSelection,
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
                  <p className="text-sm font-normal text-muted-foreground">{startDate || '—'}</p>
                </label>
              </>
            ) : (
              <>
                <label className="flex flex-col gap-1.5 text-sm font-medium">
                  {t('contractFormTenantLabel')}
                  {tenants.length === 0 ? (
                    <p className="text-xs font-normal text-muted-foreground">{t('contractFormNoTenants')}</p>
                  ) : (
                    <select
                      className="h-10 rounded-md border border-input bg-transparent px-3 text-sm"
                      value={tenantId}
                      onChange={(event) => onTenantIdChange(event.target.value)}
                    >
                      <option value="">{t('contractFormTenantPlaceholder')}</option>
                      {tenants.map((tenant) => (
                        <option key={tenant.id} value={tenant.id}>
                          {tenant.first_name} {tenant.last_name}
                        </option>
                      ))}
                    </select>
                  )}
                </label>

                <label className="flex flex-col gap-1.5 text-sm font-medium">
                  {t('contractFormRoomLabel')}
                  <RoomSearchSelect
                    selectedLabel={roomDisplayLabel}
                    onSelectRoom={onSelectRoom}
                    onClearSelection={onClearRoomSelection}
                    statusFilter="available"
                    placeholder={t('contractFormRoomPlaceholder')}
                    changeLabel={t('contractFormRoomChange')}
                    noResultsLabel={t('contractFormNoRooms')}
                  />
                </label>

                <label className="flex flex-col gap-1.5 text-sm font-medium">
                  {t('contractFormStartDateLabel')}
                  <input
                    type="date"
                    className="h-10 rounded-md border border-input bg-transparent px-3 text-sm"
                    value={startDate}
                    onChange={(event) => onStartDateChange(event.target.value)}
                  />
                </label>
              </>
            )}

            <label className="flex flex-col gap-1.5 text-sm font-medium">
              {t('contractFormEndDateLabel')}
              <input
                type="date"
                className="h-10 rounded-md border border-input bg-transparent px-3 text-sm"
                value={endDate}
                onChange={(event) => onEndDateChange(event.target.value)}
              />
            </label>

            <label className="flex flex-col gap-1.5 text-sm font-medium">
              {t('contractFormRentPriceLabel')}
              <input
                type="number"
                step="0.01"
                min="0"
                className="h-10 rounded-md border border-input bg-transparent px-3 text-sm"
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
                className="h-10 rounded-md border border-input bg-transparent px-3 text-sm"
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
                <select
                  className="h-10 rounded-md border border-input bg-transparent px-3 text-sm"
                  value={status}
                  onChange={(event) => onStatusChange(event.target.value as ContractStatus)}
                >
                  {CONTRACT_STATUSES.map((value) => (
                    <option key={value} value={value}>
                      {t(contractStatusLabelKeys[value])}
                    </option>
                  ))}
                </select>
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
