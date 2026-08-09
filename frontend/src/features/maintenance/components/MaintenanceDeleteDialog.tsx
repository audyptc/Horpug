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
import type { ApiMaintenanceRequest } from '@/types'

type Props = {
  open: boolean
  onOpenChange: (open: boolean) => void
  item: ApiMaintenanceRequest | null
  onDelete: () => void
}

export function MaintenanceDeleteDialog({ open, onOpenChange, item, onDelete }: Props) {
  const { t } = useTranslation()

  return (
    <Dialog open={open} onOpenChange={onOpenChange}>
      <DialogContent className="max-w-sm">
        <DialogHeader>
          <DialogTitle>{t('maintenance.deleteRequest')}</DialogTitle>
          <DialogDescription>
            {t('maintenance.deleteConfirm')}{' '}
            <span className="font-semibold text-foreground">{item?.title}</span>?{' '}
            {t('maintenance.deleteWarning')}
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
