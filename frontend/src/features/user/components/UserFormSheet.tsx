import type { FormEvent } from 'react'
import { useLanguage } from '@/shared/i18n/language'
import { Button } from '@/shared/components/ui/button'
import {
  Sheet,
  SheetContent,
  SheetDescription,
  SheetFooter,
  SheetHeader,
  SheetTitle,
} from '@/shared/components/ui/sheet'
import { RoleSearchSelect } from '@/features/role/components/RoleSearchSelect'
import type { ApiUserRole } from '../types'

type UserFormSheetProps = {
  open: boolean
  onOpenChange: (open: boolean) => void
  isEdit: boolean
  username: string
  onUsernameChange: (username: string) => void
  email: string
  onEmailChange: (email: string) => void
  password: string
  onPasswordChange: (password: string) => void
  roleName: string
  onRoleSelect: (role: ApiUserRole) => void
  isActive: boolean
  onIsActiveChange: (isActive: boolean) => void
  saving: boolean
  error: string | null
  onSubmit: (event: FormEvent<HTMLFormElement>) => void
}

export function UserFormSheet({
  open,
  onOpenChange,
  isEdit,
  username,
  onUsernameChange,
  email,
  onEmailChange,
  password,
  onPasswordChange,
  roleName,
  onRoleSelect,
  isActive,
  onIsActiveChange,
  saving,
  error,
  onSubmit,
}: UserFormSheetProps) {
  const { t } = useLanguage()

  return (
    <Sheet open={open} onOpenChange={onOpenChange}>
      <SheetContent>
        <form className="flex h-full flex-col gap-4" onSubmit={onSubmit}>
          <SheetHeader>
            <SheetTitle>{isEdit ? t('userFormEditTitle') : t('userFormCreateTitle')}</SheetTitle>
            <SheetDescription>
              {isEdit ? t('userFormEditDescription') : t('userFormCreateDescription')}
            </SheetDescription>
          </SheetHeader>

          <div className="flex flex-1 flex-col gap-4 overflow-y-auto pr-1">
            <label className="flex flex-col gap-1.5 text-sm font-medium">
              {t('userFormUsernameLabel')}
              <input
                type="text"
                className="h-10 rounded-md border border-input bg-transparent px-3 text-sm"
                value={username}
                onChange={(event) => onUsernameChange(event.target.value)}
                autoFocus
              />
            </label>

            <label className="flex flex-col gap-1.5 text-sm font-medium">
              {t('userFormEmailLabel')}
              <input
                type="email"
                className="h-10 rounded-md border border-input bg-transparent px-3 text-sm"
                value={email}
                onChange={(event) => onEmailChange(event.target.value)}
              />
            </label>

            <label className="flex flex-col gap-1.5 text-sm font-medium">
              {isEdit ? t('userFormPasswordEditLabel') : t('userFormPasswordLabel')}
              <input
                type="password"
                className="h-10 rounded-md border border-input bg-transparent px-3 text-sm"
                value={password}
                onChange={(event) => onPasswordChange(event.target.value)}
                autoComplete="new-password"
              />
            </label>

            <div className="flex flex-col gap-1.5 text-sm font-medium">
              {t('userFormRoleLabel')}
              <RoleSearchSelect
                selectedLabel={roleName}
                onSelectRole={onRoleSelect}
                placeholder={t('userFormRolePlaceholder')}
                searchPlaceholder={t('userFormRoleSearchPlaceholder')}
                noResultsLabel={t('userFormRoleNoResults')}
              />
            </div>

            <label className="flex items-center gap-2 text-sm font-medium">
              <input
                type="checkbox"
                className="h-4 w-4 accent-primary"
                checked={isActive}
                onChange={(event) => onIsActiveChange(event.target.checked)}
              />
              {t('userFormActiveLabel')}
            </label>
          </div>

          {error && <p className="resource-error">{error}</p>}

          <SheetFooter>
            <Button type="button" variant="outline" onClick={() => onOpenChange(false)}>
              {t('userFormCancel')}
            </Button>
            <Button type="submit" disabled={saving}>
              {saving ? t('userSaving') : t('userFormSave')}
            </Button>
          </SheetFooter>
        </form>
      </SheetContent>
    </Sheet>
  )
}
