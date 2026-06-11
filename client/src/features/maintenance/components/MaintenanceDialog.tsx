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
import { DatePicker } from '@/components/ui/date-picker'
import type { ApiMaintenanceRequest, MaintenanceStatus, MaintenancePriority } from '@/types'

type MaintenanceForm = {
  room_id: string
  title: string
  description: string
  status: MaintenanceStatus
  priority: MaintenancePriority
  reported_date: string
  resolved_date: string
  note: string
}

type Props = {
  open: boolean
  onOpenChange: (open: boolean) => void
  editingItem: ApiMaintenanceRequest | null
  form: MaintenanceForm
  onFormChange: (form: MaintenanceForm) => void
  onSave: () => void
  saving: boolean
}

export function MaintenanceDialog({
  open,
  onOpenChange,
  editingItem,
  form,
  onFormChange,
  onSave,
  saving,
}: Props) {
  const { t } = useTranslation()

  const isSaveDisabled = saving || !form.room_id || !form.title || !form.reported_date

  function set(patch: Partial<MaintenanceForm>) {
    onFormChange({ ...form, ...patch })
  }

  return (
    <Dialog open={open} onOpenChange={onOpenChange}>
      <DialogContent className="max-w-lg">
        <DialogHeader>
          <DialogTitle>
            {editingItem ? t('maintenance.editRequest') : t('maintenance.createRequest')}
          </DialogTitle>
          <DialogDescription>
            {editingItem ? t('maintenance.editDesc') : t('maintenance.createDesc')}
          </DialogDescription>
        </DialogHeader>

        <div className="grid gap-4 py-4">
          <div className="space-y-1.5">
            <Label>{t('maintenance.roomId')} *</Label>
            <Input
              placeholder={t('maintenance.roomIdPlaceholder')}
              value={form.room_id}
              onChange={(e) => set({ room_id: e.target.value })}
            />
          </div>

          <div className="space-y-1.5">
            <Label>{t('maintenance.titleField')} *</Label>
            <Input
              placeholder={t('maintenance.titlePlaceholder')}
              value={form.title}
              onChange={(e) => set({ title: e.target.value })}
            />
          </div>

          <div className="space-y-1.5">
            <Label>{t('maintenance.descriptionField')}</Label>
            <textarea
              rows={2}
              placeholder={t('maintenance.descriptionPlaceholder')}
              value={form.description}
              onChange={(e: React.ChangeEvent<HTMLTextAreaElement>) => set({ description: e.target.value })}
              className="flex min-h-16 w-full rounded-md border border-input bg-background px-3 py-2 text-sm ring-offset-background placeholder:text-muted-foreground focus-visible:outline-none focus-visible:ring-2 focus-visible:ring-ring focus-visible:ring-offset-2 resize-none"
            />
          </div>

          <div className="grid grid-cols-2 gap-4">
            <div className="space-y-1.5">
              <Label>{t('maintenance.status')} *</Label>
              <Select
                value={form.status}
                onValueChange={(v) => set({ status: v as MaintenanceStatus })}
              >
                <SelectTrigger><SelectValue /></SelectTrigger>
                <SelectContent>
                  <SelectItem value="open">{t('maintenance.statuses.open')}</SelectItem>
                  <SelectItem value="in_progress">{t('maintenance.statuses.in_progress')}</SelectItem>
                  <SelectItem value="done">{t('maintenance.statuses.done')}</SelectItem>
                  <SelectItem value="cancelled">{t('maintenance.statuses.cancelled')}</SelectItem>
                </SelectContent>
              </Select>
            </div>
            <div className="space-y-1.5">
              <Label>{t('maintenance.priority')} *</Label>
              <Select
                value={form.priority}
                onValueChange={(v) => set({ priority: v as MaintenancePriority })}
              >
                <SelectTrigger><SelectValue /></SelectTrigger>
                <SelectContent>
                  <SelectItem value="urgent">{t('maintenance.priorities.urgent')}</SelectItem>
                  <SelectItem value="high">{t('maintenance.priorities.high')}</SelectItem>
                  <SelectItem value="normal">{t('maintenance.priorities.normal')}</SelectItem>
                  <SelectItem value="low">{t('maintenance.priorities.low')}</SelectItem>
                </SelectContent>
              </Select>
            </div>
          </div>

          <div className="grid grid-cols-2 gap-4">
            <div className="space-y-1.5">
              <Label>{t('maintenance.reportedDate')} *</Label>
              <DatePicker
                value={form.reported_date}
                onChange={(v) => set({ reported_date: v })}
              />
            </div>
            <div className="space-y-1.5">
              <Label>{t('maintenance.resolvedDate')}</Label>
              <DatePicker
                value={form.resolved_date}
                onChange={(v) => set({ resolved_date: v })}
              />
            </div>
          </div>

          <div className="space-y-1.5">
            <Label>{t('maintenance.note')}</Label>
            <textarea
              rows={2}
              placeholder={t('maintenance.notePlaceholder')}
              value={form.note}
              onChange={(e: React.ChangeEvent<HTMLTextAreaElement>) => set({ note: e.target.value })}
              className="flex min-h-16 w-full rounded-md border border-input bg-background px-3 py-2 text-sm ring-offset-background placeholder:text-muted-foreground focus-visible:outline-none focus-visible:ring-2 focus-visible:ring-ring focus-visible:ring-offset-2 resize-none"
            />
          </div>
        </div>

        <DialogFooter>
          <Button variant="outline" onClick={() => onOpenChange(false)}>
            {t('common.cancel')}
          </Button>
          <Button onClick={onSave} disabled={isSaveDisabled}>
            {saving
              ? t('common.loading')
              : editingItem
                ? t('maintenance.saveChanges')
                : t('maintenance.createRequest')}
          </Button>
        </DialogFooter>
      </DialogContent>
    </Dialog>
  )
}
