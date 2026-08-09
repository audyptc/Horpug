import { useTranslation } from 'react-i18next'
import { Button } from '@/components/ui/button'
import {
  Dialog,
  DialogContent,
  DialogHeader,
  DialogTitle,
  DialogDescription,
  DialogFooter,
} from '@/components/ui/dialog'
import type { ApiRoom } from '@/types'

type Props = {
  open: boolean
  onOpenChange: (open: boolean) => void
  room: ApiRoom | null
  onDelete: () => void
}

export function RoomDeleteDialog({ open, onOpenChange, room, onDelete }: Props) {
  const { t } = useTranslation()

  return (
    <Dialog open={open} onOpenChange={onOpenChange}>
      <DialogContent className="max-w-sm">
        <DialogHeader>
          <DialogTitle>{t('rooms.deleteRoom')}</DialogTitle>
          <DialogDescription>
            {t('rooms.deleteConfirm')}{' '}
            <span className="font-semibold text-foreground">{room?.room_number}</span>?{' '}
            {t('rooms.deleteWarning')}
          </DialogDescription>
        </DialogHeader>
        <DialogFooter className="gap-2">
          <Button variant="outline" onClick={() => onOpenChange(false)}>
            {t('common.cancel')}
          </Button>
          <Button variant="destructive" onClick={onDelete}>
            {t('common.delete')}
          </Button>
        </DialogFooter>
      </DialogContent>
    </Dialog>
  )
}
