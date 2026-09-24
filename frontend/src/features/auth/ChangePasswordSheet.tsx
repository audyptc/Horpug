import { useState, type FormEvent } from 'react'
import { api, extractErrorCode, extractErrorMessage } from '@/shared/api/client'
import { Button } from '@/shared/components/ui/button'
import {
  Sheet,
  SheetContent,
  SheetDescription,
  SheetFooter,
  SheetHeader,
  SheetTitle,
} from '@/shared/components/ui/sheet'
import { useLanguage, type TranslationKey } from '@/shared/i18n/language'

// Must match auth usecase MinPasswordLength on the backend.
const MIN_PASSWORD_LENGTH = 8

const errorKeys: Record<string, TranslationKey> = {
  wrong_password: 'changePasswordErrorWrong',
  weak_password: 'changePasswordErrorWeak',
  same_password: 'changePasswordErrorSame',
  too_many_requests: 'changePasswordErrorTooMany',
}

type ChangePasswordSheetProps = {
  open: boolean
  onOpenChange: (open: boolean) => void
  onChanged: () => void
}

export function ChangePasswordSheet({ open, onOpenChange, onChanged }: ChangePasswordSheetProps) {
  const { t } = useLanguage()
  const [currentPassword, setCurrentPassword] = useState('')
  const [newPassword, setNewPassword] = useState('')
  const [confirmPassword, setConfirmPassword] = useState('')
  const [saving, setSaving] = useState(false)
  const [error, setError] = useState<string | null>(null)

  const handleOpenChange = (next: boolean) => {
    if (!next) {
      setCurrentPassword('')
      setNewPassword('')
      setConfirmPassword('')
      setError(null)
    }
    onOpenChange(next)
  }

  const handleSubmit = async (event: FormEvent<HTMLFormElement>) => {
    event.preventDefault()
    if (newPassword.length < MIN_PASSWORD_LENGTH) {
      setError(t('changePasswordErrorWeak'))
      return
    }
    if (newPassword !== confirmPassword) {
      setError(t('changePasswordErrorMismatch'))
      return
    }

    setSaving(true)
    setError(null)
    try {
      await api.post('/auth/change-password', {
        current_password: currentPassword,
        new_password: newPassword,
      })
      handleOpenChange(false)
      onChanged()
    } catch (err) {
      const key = errorKeys[extractErrorCode(err) ?? '']
      setError(key ? t(key) : extractErrorMessage(err, t('changePasswordErrorGeneric')))
    } finally {
      setSaving(false)
    }
  }

  const inputClass = 'h-10 rounded-md border border-input bg-transparent px-3 text-sm'

  return (
    <Sheet open={open} onOpenChange={handleOpenChange}>
      <SheetContent>
        <form className="flex h-full flex-col gap-4" onSubmit={handleSubmit}>
          <SheetHeader>
            <SheetTitle>{t('changePasswordTitle')}</SheetTitle>
            <SheetDescription>{t('changePasswordDescription')}</SheetDescription>
          </SheetHeader>

          <div className="flex flex-1 flex-col gap-4 overflow-y-auto pr-1">
            <label className="flex flex-col gap-1.5 text-sm font-medium">
              {t('changePasswordCurrentLabel')}
              <input
                type="password"
                autoComplete="current-password"
                className={inputClass}
                value={currentPassword}
                onChange={(event) => setCurrentPassword(event.target.value)}
                required
              />
            </label>
            <label className="flex flex-col gap-1.5 text-sm font-medium">
              {t('changePasswordNewLabel')}
              <input
                type="password"
                autoComplete="new-password"
                minLength={MIN_PASSWORD_LENGTH}
                className={inputClass}
                value={newPassword}
                onChange={(event) => setNewPassword(event.target.value)}
                required
              />
              <span className="text-xs font-normal text-muted-foreground">
                {t('changePasswordHint')}
              </span>
            </label>
            <label className="flex flex-col gap-1.5 text-sm font-medium">
              {t('changePasswordConfirmLabel')}
              <input
                type="password"
                autoComplete="new-password"
                className={inputClass}
                value={confirmPassword}
                onChange={(event) => setConfirmPassword(event.target.value)}
                required
              />
            </label>
          </div>

          {error && <p className="resource-error">{error}</p>}

          <SheetFooter>
            <Button type="button" variant="outline" onClick={() => handleOpenChange(false)}>
              {t('changePasswordCancel')}
            </Button>
            <Button type="submit" disabled={saving}>
              {saving ? t('changePasswordSaving') : t('changePasswordSubmit')}
            </Button>
          </SheetFooter>
        </form>
      </SheetContent>
    </Sheet>
  )
}
