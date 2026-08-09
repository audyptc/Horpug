import { useTranslation } from 'react-i18next'
import { Button } from '@/components/ui/button'
import { Dialog, DialogContent, DialogHeader, DialogTitle, DialogDescription, DialogFooter } from '@/components/ui/dialog'
import type { ApiElectricMeter } from '@/types'

type Props = {
  open: boolean
  onOpenChange: (open: boolean) => void
  reading: ApiElectricMeter | null
  onDelete: () => void
}

export function ElectricMeterDeleteDialog({ open, onOpenChange, reading, onDelete }: Props) {
  const { t } = useTranslation()
  return (
    <Dialog open={open} onOpenChange={onOpenChange}>
      <DialogContent className="max-w-sm">
        <DialogHeader>
          <DialogTitle>{t('electricMeters.deleteReading')}</DialogTitle>
          <DialogDescription>
            {t('electricMeters.deleteConfirm')}{' '}
            <span className="font-semibold text-foreground">{t('electricMeters.colRoom')} {reading?.room_number}</span>?{' '}
            {t('electricMeters.deleteWarning')}
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
