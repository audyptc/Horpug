import { useTranslation } from 'react-i18next'
import { Button } from '@/components/ui/button'
import { Input } from '@/components/ui/input'
import { Label } from '@/components/ui/label'
import {
  Dialog,
  DialogContent,
  DialogHeader,
  DialogTitle,
  DialogDescription,
  DialogFooter,
} from '@/components/ui/dialog'
import {
  Select,
  SelectContent,
  SelectItem,
  SelectTrigger,
  SelectValue,
} from '@/components/ui/select'
import type { ApiContract, ApiTenant, ApiRoom, ContractStatus } from '@/types'

type ContractForm = {
  tenant_id: string
  room_id: string
  start_date: string
  end_date: string
  rent_price: string
  deposit: string
  note: string
  status: ContractStatus
}

type Props = {
  open: boolean
  onOpenChange: (open: boolean) => void
  editingContract: ApiContract | null
  form: ContractForm
  onFormChange: (form: ContractForm) => void
  onSave: () => void
  saving: boolean
  tenants: ApiTenant[]
  availableRooms: ApiRoom[]
  saveError?: string
}

export function ContractDialog({
  open,
  onOpenChange,
  editingContract,
  form,
  onFormChange,
  onSave,
  saving,
  tenants,
  availableRooms,
  saveError,
}: Props) {
  const { t } = useTranslation()

  const isSaveDisabled =
    saving || !form.tenant_id || !form.room_id || !form.start_date || !form.rent_price

  function set(patch: Partial<ContractForm>) {
    onFormChange({ ...form, ...patch })
  }

  return (
    <Dialog open={open} onOpenChange={onOpenChange}>
      <DialogContent className="max-w-lg">
        <DialogHeader>
          <DialogTitle>
            {editingContract ? t('contracts.editContract') : t('contracts.createContract')}
          </DialogTitle>
          <DialogDescription>
            {editingContract ? t('contracts.editDesc') : t('contracts.createDesc')}
          </DialogDescription>
        </DialogHeader>

        <div className="grid gap-4 py-4">
          <div className="space-y-1.5">
            <Label>{t('contracts.tenant')} *</Label>
            <Select
              value={form.tenant_id}
              onValueChange={(v) => set({ tenant_id: v })}
              disabled={!!editingContract}
            >
              <SelectTrigger><SelectValue placeholder={t('contracts.selectTenant')} /></SelectTrigger>
              <SelectContent>
                {tenants.map((tn) => (
                  <SelectItem key={tn.id} value={tn.id}>
                    {tn.first_name} {tn.last_name}
                  </SelectItem>
                ))}
              </SelectContent>
            </Select>
          </div>

          <div className="space-y-1.5">
            <Label>{t('contracts.room')} *</Label>
            <Select
              value={form.room_id}
              onValueChange={(v) => set({ room_id: v })}
              disabled={!!editingContract}
            >
              <SelectTrigger><SelectValue placeholder={t('contracts.selectRoom')} /></SelectTrigger>
              <SelectContent>
                {availableRooms.map((rm) => (
                  <SelectItem key={rm.id} value={rm.id}>
                    {rm.room_number} — {rm.rent_price.toLocaleString()} ฿/เดือน
                  </SelectItem>
                ))}
              </SelectContent>
            </Select>
          </div>

          <div className="grid grid-cols-2 gap-4">
            <div className="space-y-1.5">
              <Label>{t('contracts.startDate')} *</Label>
              <Input
                type="date"
                value={form.start_date}
                onChange={(e) => set({ start_date: e.target.value })}
                disabled={!!editingContract}
              />
            </div>
            <div className="space-y-1.5">
              <Label>{t('contracts.endDate')}</Label>
              <Input
                type="date"
                value={form.end_date}
                onChange={(e) => set({ end_date: e.target.value })}
                placeholder={t('contracts.endDateOptional')}
              />
            </div>
          </div>

          <div className="grid grid-cols-2 gap-4">
            <div className="space-y-1.5">
              <Label>{t('contracts.rentPrice')} *</Label>
              <Input
                type="number"
                min="0"
                value={form.rent_price}
                onChange={(e) => set({ rent_price: e.target.value })}
              />
            </div>
            <div className="space-y-1.5">
              <Label>{t('contracts.deposit')}</Label>
              <Input
                type="number"
                min="0"
                value={form.deposit}
                onChange={(e) => set({ deposit: e.target.value })}
              />
            </div>
          </div>

          {editingContract && (
            <div className="space-y-1.5">
              <Label>{t('contracts.status')}</Label>
              <Select
                value={form.status}
                onValueChange={(v) => set({ status: v as ContractStatus })}
              >
                <SelectTrigger><SelectValue /></SelectTrigger>
                <SelectContent>
                  <SelectItem value="active">{t('contracts.statuses.active')}</SelectItem>
                  <SelectItem value="expired">{t('contracts.statuses.expired')}</SelectItem>
                  <SelectItem value="terminated">{t('contracts.statuses.terminated')}</SelectItem>
                </SelectContent>
              </Select>
            </div>
          )}

          <div className="space-y-1.5">
            <Label>{t('contracts.note')}</Label>
            <textarea
              rows={2}
              placeholder={t('contracts.notePlaceholder')}
              value={form.note}
              onChange={(e: React.ChangeEvent<HTMLTextAreaElement>) => set({ note: e.target.value })}
              className="flex min-h-16 w-full rounded-md border border-input bg-background px-3 py-2 text-sm ring-offset-background placeholder:text-muted-foreground focus-visible:outline-none focus-visible:ring-2 focus-visible:ring-ring focus-visible:ring-offset-2 resize-none"
            />
          </div>
        </div>

        {saveError && (
          <div className="rounded-md border border-destructive/30 bg-destructive/10 px-4 py-3 text-sm text-destructive">
            {saveError}
          </div>
        )}

        <DialogFooter>
          <Button variant="outline" onClick={() => onOpenChange(false)}>
            {t('common.cancel')}
          </Button>
          <Button onClick={onSave} disabled={isSaveDisabled}>
            {saving
              ? t('common.loading')
              : editingContract
                ? t('contracts.saveChanges')
                : t('contracts.createContract')}
          </Button>
        </DialogFooter>
      </DialogContent>
    </Dialog>
  )
}
