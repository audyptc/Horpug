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
import type { ApiPayment } from '@/types'

type Props = {
  open: boolean
  onOpenChange: (open: boolean) => void
  item: ApiPayment | null
  onDelete: () => void
}

function formatBaht(n: number) {
  return n.toLocaleString('th-TH', { minimumFractionDigits: 2, maximumFractionDigits: 2 })
}

export function PaymentDeleteDialog({ open, onOpenChange, item, onDelete }: Props) {
  const { t } = useTranslation()

  return (
    <Dialog open={open} onOpenChange={onOpenChange}>
      <DialogContent className="max-w-sm">
        <DialogHeader>
          <DialogTitle>{t('payments.deletePayment')}</DialogTitle>
          <DialogDescription>
            {t('payments.deleteConfirm')}{' '}
            <span className="font-semibold text-foreground">
              ฿{item ? formatBaht(item.amount) : ''}
            </span>?{' '}
            {t('payments.deleteWarning')}
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
