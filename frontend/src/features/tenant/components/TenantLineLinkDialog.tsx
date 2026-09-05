import { QRCodeSVG } from 'qrcode.react'
import {
  AlertDialog,
  AlertDialogCancel,
  AlertDialogContent,
  AlertDialogDescription,
  AlertDialogFooter,
  AlertDialogHeader,
  AlertDialogTitle,
} from '@/shared/components/ui/alert-dialog'
import { Button } from '@/shared/components/ui/button'
import { useLanguage } from '@/shared/i18n/language'
import type { ApiTenant } from '../types'

type TenantLineLinkDialogProps = {
  open: boolean
  onOpenChange: (open: boolean) => void
  tenant: ApiTenant | null
  link: string
}

export function TenantLineLinkDialog({ open, onOpenChange, tenant, link }: TenantLineLinkDialogProps) {
  const { t } = useLanguage()

  return (
    <AlertDialog open={open} onOpenChange={onOpenChange}>
      <AlertDialogContent>
        <AlertDialogHeader>
          <AlertDialogTitle>{t('tenantLineLinkDialogTitle')}</AlertDialogTitle>
          <AlertDialogDescription>
            {tenant
              ? t('tenantLineLinkDialogDescription').replace(
                  '{name}',
                  `${tenant.first_name} ${tenant.last_name}`
                )
              : ''}
          </AlertDialogDescription>
        </AlertDialogHeader>

        <div className="flex flex-col items-center gap-3 py-2">
          <div className="rounded-lg border border-border bg-white p-3">
            <QRCodeSVG value={link} size={180} />
          </div>
          <p className="text-center text-xs text-muted-foreground">{t('tenantLineLinkQrHint')}</p>
          <p className="w-full break-all rounded-md bg-muted px-3 py-2 text-center text-xs text-muted-foreground">
            {link}
          </p>
        </div>

        <AlertDialogFooter>
          <AlertDialogCancel asChild>
            <Button type="button">{t('acknowledge')}</Button>
          </AlertDialogCancel>
        </AlertDialogFooter>
      </AlertDialogContent>
    </AlertDialog>
  )
}
