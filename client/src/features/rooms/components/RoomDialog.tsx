import { useEffect, useState } from 'react'
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
import { roomTypeService } from '@/features/rooms/roomTypeService'
import type { ApiRoom, ApiRoomType, RoomStatus, RoomType } from '@/types'

type RoomForm = {
  room_number: string
  floor: number
  type: RoomType
  status: RoomStatus
  rent_price: number
  description: string
}

type Props = {
  open: boolean
  onOpenChange: (open: boolean) => void
  editingRoom: ApiRoom | null
  form: RoomForm
  onFormChange: (form: RoomForm) => void
  onSave: () => void
  saving: boolean
}

export function RoomDialog({
  open,
  onOpenChange,
  editingRoom,
  form,
  onFormChange,
  onSave,
  saving,
}: Props) {
  const { t } = useTranslation()
  const [roomTypes, setRoomTypes] = useState<ApiRoomType[]>([])

  useEffect(() => {
    if (!open) return
    roomTypeService.list().then(setRoomTypes).catch(() => {})
  }, [open])

  const isSaveDisabled = saving || !form.room_number || form.floor <= 0 || form.rent_price <= 0

  function set(patch: Partial<RoomForm>) {
    onFormChange({ ...form, ...patch })
  }

  return (
    <Dialog open={open} onOpenChange={onOpenChange}>
      <DialogContent className="max-w-md">
        <DialogHeader>
          <DialogTitle>{editingRoom ? t('rooms.editRoom') : t('rooms.createRoom')}</DialogTitle>
          <DialogDescription>
            {editingRoom ? t('rooms.editDesc') : t('rooms.createDesc')}
          </DialogDescription>
        </DialogHeader>
        <div className="grid gap-4 py-4">
          <div className="grid grid-cols-2 gap-4">
            <div className="space-y-1.5">
              <Label>{t('rooms.roomNumber')} *</Label>
              <Input
                placeholder="101"
                value={form.room_number}
                onChange={(e) => set({ room_number: e.target.value })}
              />
            </div>
            <div className="space-y-1.5">
              <Label>{t('rooms.floor')} *</Label>
              <Input
                type="number"
                min={1}
                placeholder="1"
                value={form.floor}
                onChange={(e) => set({ floor: Number(e.target.value) })}
              />
            </div>
          </div>
          <div className="grid grid-cols-2 gap-4">
            <div className="space-y-1.5">
              <Label>{t('rooms.type')}</Label>
              <Select value={form.type} onValueChange={(v) => set({ type: v as RoomType })}>
                <SelectTrigger><SelectValue /></SelectTrigger>
                <SelectContent>
                  {roomTypes.length === 0 ? (
                    <SelectItem value={form.type || 'standard'}>{form.type || 'standard'}</SelectItem>
                  ) : (
                    roomTypes.map((rt) => (
                      <SelectItem key={rt.id} value={rt.id}>{rt.name}</SelectItem>
                    ))
                  )}
                </SelectContent>
              </Select>
            </div>
            <div className="space-y-1.5">
              <Label>{t('rooms.status')}</Label>
              <Select value={form.status} onValueChange={(v) => set({ status: v as RoomStatus })}>
                <SelectTrigger><SelectValue /></SelectTrigger>
                <SelectContent>
                  <SelectItem value="available">{t('rooms.statuses.available')}</SelectItem>
                  <SelectItem value="occupied">{t('rooms.statuses.occupied')}</SelectItem>
                  <SelectItem value="maintenance">{t('rooms.statuses.maintenance')}</SelectItem>
                </SelectContent>
              </Select>
            </div>
          </div>
          <div className="space-y-1.5">
            <Label>{t('rooms.rentPrice')} * ({t('rooms.bahtPerMonth')})</Label>
            <Input
              type="number"
              min={0}
              placeholder="3500"
              value={form.rent_price || ''}
              onChange={(e) => set({ rent_price: Number(e.target.value) })}
            />
          </div>
          <div className="space-y-1.5">
            <Label>{t('rooms.description')}</Label>
            <textarea
              rows={2}
              placeholder={t('rooms.descriptionPlaceholder')}
              value={form.description}
              onChange={(e: React.ChangeEvent<HTMLTextAreaElement>) => set({ description: e.target.value })}
              className="flex min-h-16 w-full rounded-md border border-input bg-background px-3 py-2 text-sm ring-offset-background placeholder:text-muted-foreground focus-visible:outline-none focus-visible:ring-2 focus-visible:ring-ring focus-visible:ring-offset-2 resize-none"
            />
          </div>
        </div>
        <DialogFooter>
          <Button variant="outline" onClick={() => onOpenChange(false)}>
            {t('common.cancel')}
          </Button>
          <Button onClick={onSave} disabled={isSaveDisabled}>
            {saving ? t('common.loading') : editingRoom ? t('rooms.saveChanges') : t('rooms.createRoom')}
          </Button>
        </DialogFooter>
      </DialogContent>
    </Dialog>
  )
}
