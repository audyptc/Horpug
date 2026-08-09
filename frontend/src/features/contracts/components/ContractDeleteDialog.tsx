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
import type { ApiContract } from '@/types'

type Props = {
  open: boolean
  onOpenChange: (open: boolean) => void
  contract: ApiContract | null
  onDelete: () => void
}

export function ContractDeleteDialog({ open, onOpenChange, contract, onDelete }: Props) {
  const { t } = useTranslation()

  return (
    <Dialog open={open} onOpenChange={onOpenChange}>
      <DialogContent className="max-w-sm">
        <DialogHeader>
          <DialogTitle>{t('contracts.deleteContract')}</DialogTitle>
          <DialogDescription>
            {t('contracts.deleteConfirm')}{' '}
            <span className="font-semibold text-foreground">
              {contract?.tenant_first_name} {contract?.tenant_last_name}
            </span>?{' '}
            {t('contracts.deleteWarning')}
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
