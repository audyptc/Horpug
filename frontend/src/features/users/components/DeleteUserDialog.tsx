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
import { userService } from '@/features/users/userService'
import { useToast } from '@/components/ui/toast'
import type { ApiUser } from '@/types'

interface Props {
  open: boolean
  onOpenChange: (open: boolean) => void
  user: ApiUser | null
  onDeleted: () => void
}

export function DeleteUserDialog({ open, onOpenChange, user, onDeleted }: Props) {
  const { t } = useTranslation()
  const toast = useToast()

  async function handleDelete() {
    if (!user) return
    try {
      await userService.delete(user.id)
      toast.success(t('users.deleteSuccess'))
      onDeleted()
    } catch (err: unknown) {
      const reason = (err as { response?: { data?: { message?: string } } })?.response?.data?.message
      toast.error(reason ? `${t('users.deleteError')}: ${reason}` : t('users.deleteError'))
    } finally {
      onOpenChange(false)
    }
  }

  return (
    <Dialog open={open} onOpenChange={onOpenChange}>
      <DialogContent className="max-w-sm">
        <DialogHeader>
          <DialogTitle>{t('users.deleteUser')}</DialogTitle>
          <DialogDescription>
            {t('users.deleteConfirm')}{' '}
            <span className="font-semibold text-foreground">{user?.full_name}</span>?{' '}
            {t('users.deleteWarning')}
          </DialogDescription>
        </DialogHeader>
        <DialogFooter className="gap-2">
          <Button variant="outline" onClick={() => onOpenChange(false)}>
            {t('common.cancel')}
          </Button>
          <Button variant="destructive" onClick={handleDelete}>
            {t('common.delete')}
          </Button>
        </DialogFooter>
      </DialogContent>
    </Dialog>
  )
}
