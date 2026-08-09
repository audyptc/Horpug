import { useTranslation } from 'react-i18next'
import { Button } from '@/components/ui/button'
import { Dialog, DialogContent, DialogHeader, DialogTitle, DialogDescription, DialogFooter } from '@/components/ui/dialog'
import type { ApiWaterMeter } from '@/types'

type Props = {
  open: boolean
  onOpenChange: (open: boolean) => void
  reading: ApiWaterMeter | null
  onDelete: () => void
}

export function WaterMeterDeleteDialog({ open, onOpenChange, reading, onDelete }: Props) {
  const { t } = useTranslation()
  return (
    <Dialog open={open} onOpenChange={onOpenChange}>
      <DialogContent className="max-w-sm">
        <DialogHeader>
          <DialogTitle>{t('waterMeters.deleteReading')}</DialogTitle>
          <DialogDescription>
            {t('waterMeters.deleteConfirm')}{' '}
            <span className="font-semibold text-foreground">{t('waterMeters.colRoom')} {reading?.room_number}</span>?{' '}
            {t('waterMeters.deleteWarning')}
          </DialogDescription>
        </DialogHeader>
        <DialogFooter className="gap-2">
          <Button variant="outline" onClick={() => onOpenChange(false)}>{t('common.cancel')}</Button>
          <Button variant="destructive" onClick={onDelete}>{t('common.delete')}</Button>
        </DialogFooter>
      </DialogContent>
    </Dialog>
  )
}
