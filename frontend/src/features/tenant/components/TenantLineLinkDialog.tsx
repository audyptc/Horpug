import { useState } from 'react'
import { QRCodeSVG } from 'qrcode.react'
import { Check, CheckCircle2, Copy, QrCode, TriangleAlert, UserPlus } from 'lucide-react'
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
  addFriendUrl: string | null
  lineStatus: { linked: boolean; is_friend: boolean } | null
}

function CopyLinkRow({ value, copiedLabel }: { value: string; copiedLabel: string }) {
  const [copied, setCopied] = useState(false)

  async function handleCopy() {
    try {
      await navigator.clipboard.writeText(value)
      setCopied(true)
      window.setTimeout(() => setCopied(false), 1500)
    } catch {
      // Clipboard API may be unavailable (e.g. insecure context); the link
      // below is still tappable/selectable by hand.
    }
  }

  return (
    <div className="flex w-full items-center gap-1.5 rounded-md border border-border bg-muted/50 py-1 pl-3 pr-1.5">
      <a
        href={value}
        target="_blank"
        rel="noreferrer"
        className="min-w-0 flex-1 truncate text-xs text-muted-foreground underline decoration-muted-foreground/40 underline-offset-2 hover:text-foreground"
      >
        {value}
      </a>
      <button
        type="button"
        onClick={handleCopy}
        className="flex shrink-0 items-center gap-1 rounded-md px-2 py-1.5 text-xs font-medium text-muted-foreground transition-colors hover:bg-accent hover:text-accent-foreground"
      >
        {copied ? (
          <>
            <Check size={13} className="text-emerald-600" />
            <span className="text-emerald-600">{copiedLabel}</span>
          </>
        ) : (
          <Copy size={13} />
        )}
      </button>
    </div>
  )
}

export function TenantLineLinkDialog({
  open,
  onOpenChange,
  tenant,
  link,
  addFriendUrl,
  lineStatus,
}: TenantLineLinkDialogProps) {
  const { t } = useLanguage()

  // Both conditions must hold before an invoice can be pushed, so name the
  // one that's actually missing rather than just saying it won't work.
  const statusKey = !lineStatus
    ? null
    : !lineStatus.linked
      ? 'tenantLineStatusNotLinked'
      : !lineStatus.is_friend
        ? 'tenantLineStatusNotFriend'
        : 'tenantLineStatusReady'
  const isReady = statusKey === 'tenantLineStatusReady'

  return (
    <AlertDialog open={open} onOpenChange={onOpenChange}>
      <AlertDialogContent className="max-w-sm">
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

        <div className="flex flex-col gap-4 py-1">
          {statusKey && (
            <div
              className={`flex items-start gap-2 rounded-md px-3 py-2.5 text-xs leading-relaxed ${
                isReady
                  ? 'bg-emerald-50 text-emerald-800 dark:bg-emerald-950/40 dark:text-emerald-300'
                  : 'bg-amber-50 text-amber-900 dark:bg-amber-950/40 dark:text-amber-300'
              }`}
            >
              {isReady ? (
                <CheckCircle2 size={16} className="mt-0.5 shrink-0" />
              ) : (
                <TriangleAlert size={16} className="mt-0.5 shrink-0" />
              )}
              <span>{t(statusKey)}</span>
            </div>
          )}

          <div className="flex flex-col items-center gap-3 rounded-lg border border-border bg-muted/30 p-4">
            <div className="rounded-lg border border-border bg-white p-3">
              <QRCodeSVG value={link} size={168} />
            </div>
            <p className="flex items-center justify-center gap-1.5 text-center text-xs text-muted-foreground">
              <QrCode size={13} className="shrink-0" />
              {t('tenantLineLinkQrHint')}
            </p>
            <CopyLinkRow value={link} copiedLabel={t('copied')} />
          </div>

          {addFriendUrl && (
            <div className="flex flex-col gap-1.5 border-t border-border pt-2">
              <p className="flex items-center gap-1.5 text-xs text-muted-foreground">
                <UserPlus size={13} className="shrink-0" />
                {t('tenantLineAddFriendHint')}
              </p>
              <CopyLinkRow value={addFriendUrl} copiedLabel={t('copied')} />
            </div>
          )}
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
