import { useState, useEffect, useMemo } from 'react'
import { useTranslation } from 'react-i18next'
import { Check, ChevronDown, Search } from 'lucide-react'
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
import { Popover, PopoverContent, PopoverTrigger } from '@/components/ui/popover'
import { ScrollArea } from '@/components/ui/scroll-area'
import { roomService } from '@/features/rooms/roomService'
import type { ApiMaintenanceRequest, MaintenanceStatus, MaintenancePriority, ApiRoom } from '@/types'
import { cn } from '@/lib/utils'

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
  error?: string
}

export function MaintenanceDialog({
  open,
  onOpenChange,
  editingItem,
  form,
  onFormChange,
  onSave,
  saving,
  error,
}: Props) {
  const { t } = useTranslation()

  const [rooms, setRooms] = useState<ApiRoom[]>([])
  const [roomsLoading, setRoomsLoading] = useState(false)
  const [roomSearch, setRoomSearch] = useState('')
  const [roomPickerOpen, setRoomPickerOpen] = useState(false)

  const isSaveDisabled = saving || !form.room_id || !form.title || !form.reported_date

  function set(patch: Partial<MaintenanceForm>) {
    onFormChange({ ...form, ...patch })
  }

  useEffect(() => {
    if (!open) return
    setRoomsLoading(true)
    roomService
      .list(1, 200)
      .then((r) => setRooms(r.data))
      .catch(() => {})
      .finally(() => setRoomsLoading(false))
  }, [open])

  useEffect(() => {
    if (!open) {
      setRoomSearch('')
      setRoomPickerOpen(false)
    }
  }, [open])

  const selectedRoom = rooms.find((r) => r.id === form.room_id) ?? null

  const displayRoom =
    selectedRoom?.room_number ??
    (editingItem?.room_number && form.room_id === editingItem.room_id
      ? editingItem.room_number
      : null)

  const filteredRooms = useMemo(() => {
    const q = roomSearch.toLowerCase()
    if (!q) return rooms
    return rooms.filter((r) => r.room_number.toLowerCase().includes(q))
  }, [rooms, roomSearch])

  function selectRoom(room: ApiRoom) {
    set({ room_id: room.id })
    setRoomPickerOpen(false)
    setRoomSearch('')
  }

  function handleStatusChange(v: string) {
    const patch: Partial<MaintenanceForm> = { status: v as MaintenanceStatus }
    if (v === 'done' && !form.resolved_date) {
      patch.resolved_date = new Date().toISOString().slice(0, 10)
    }
    set(patch)
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
            <Label>{t('maintenance.selectRoom')} *</Label>
            <Popover open={roomPickerOpen} onOpenChange={setRoomPickerOpen}>
              <PopoverTrigger asChild>
                <Button
                  variant="outline"
                  role="combobox"
                  className={cn(
                    'w-full justify-between font-normal',
                    !displayRoom && 'text-muted-foreground'
                  )}
                >
                  {displayRoom ?? t('maintenance.selectRoomPlaceholder')}
                  <ChevronDown className="ml-2 h-4 w-4 shrink-0 opacity-50" />
                </Button>
              </PopoverTrigger>
              <PopoverContent className="w-full p-0" align="start">
                <div className="p-2 border-b">
                  <div className="relative">
                    <Search className="absolute left-2.5 top-1/2 -translate-y-1/2 w-3.5 h-3.5 text-muted-foreground" />
                    <Input
                      placeholder={t('maintenance.roomSearch')}
                      className="pl-8 h-8 text-sm"
                      value={roomSearch}
                      onChange={(e) => setRoomSearch(e.target.value)}
                    />
                  </div>
                </div>
                <ScrollArea className="max-h-48">
                  {roomsLoading ? (
                    <p className="py-6 text-center text-sm text-muted-foreground">
                      {t('common.loading')}
                    </p>
                  ) : filteredRooms.length === 0 ? (
                    <p className="py-6 text-center text-sm text-muted-foreground">
                      {t('maintenance.noRooms')}
                    </p>
                  ) : (
                    filteredRooms.map((room) => (
                      <button
                        key={room.id}
                        type="button"
                        onClick={() => selectRoom(room)}
                        className="flex w-full items-center gap-2 px-3 py-2 text-sm hover:bg-muted transition-colors"
                      >
                        <Check
                          className={cn(
                            'h-4 w-4 shrink-0',
                            form.room_id === room.id ? 'opacity-100' : 'opacity-0'
                          )}
                        />
                        {room.room_number}
                      </button>
                    ))
                  )}
                </ScrollArea>
              </PopoverContent>
            </Popover>
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
              <Select value={form.status} onValueChange={handleStatusChange}>
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
            {(form.status === 'done' || form.status === 'cancelled' || form.resolved_date) && (
              <div className="space-y-1.5">
                <Label>{t('maintenance.resolvedDate')}</Label>
                <DatePicker
                  value={form.resolved_date}
                  onChange={(v) => set({ resolved_date: v })}
                />
              </div>
            )}
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

          {error && (
            <p className="text-sm text-destructive">{error}</p>
          )}
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
