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

type RoleFormSheetProps = {
  open: boolean
  onOpenChange: (open: boolean) => void
  isEdit: boolean
  name: string
  onNameChange: (name: string) => void
  description: string
  onDescriptionChange: (description: string) => void
  isActive: boolean
  onIsActiveChange: (isActive: boolean) => void
  saving: boolean
  error: string | null
  onSubmit: (event: FormEvent<HTMLFormElement>) => void
}

export function RoleFormSheet({
  open,
  onOpenChange,
  isEdit,
  name,
  onNameChange,
  description,
  onDescriptionChange,
  isActive,
  onIsActiveChange,
  saving,
  error,
  onSubmit,
}: RoleFormSheetProps) {
  const { t } = useLanguage()

  return (
    <Sheet open={open} onOpenChange={onOpenChange}>
      <SheetContent>
        <form className="flex h-full flex-col gap-4" onSubmit={onSubmit}>
          <SheetHeader>
            <SheetTitle>
              {isEdit ? t('rolePermissionsFormEditTitle') : t('rolePermissionsFormCreateTitle')}
            </SheetTitle>
            <SheetDescription>
              {isEdit ? t('rolePermissionsFormEditDescription') : t('rolePermissionsFormCreateDescription')}
            </SheetDescription>
          </SheetHeader>

          <label className="flex flex-col gap-1.5 text-sm font-medium">
            {t('rolePermissionsFormNameLabel')}
            <input
              type="text"
              className="h-10 rounded-md border border-input bg-transparent px-3 text-sm"
              value={name}
              onChange={(event) => onNameChange(event.target.value)}
              autoFocus
            />
          </label>

          <label className="flex flex-col gap-1.5 text-sm font-medium">
            {t('rolePermissionsFormDescriptionLabel')}
            <textarea
              className="min-h-20 rounded-md border border-input bg-transparent px-3 py-2 text-sm"
              value={description}
              onChange={(event) => onDescriptionChange(event.target.value)}
            />
          </label>

          <label className="flex items-center gap-2 text-sm font-medium">
            <input
              type="checkbox"
              className="h-4 w-4 accent-primary"
              checked={isActive}
              onChange={(event) => onIsActiveChange(event.target.checked)}
            />
            {t('rolePermissionsFormActiveLabel')}
          </label>

          {error && <p className="resource-error">{error}</p>}

          <SheetFooter>
            <Button type="button" variant="outline" onClick={() => onOpenChange(false)}>
              {t('rolePermissionsFormCancel')}
            </Button>
            <Button type="submit" disabled={saving}>
              {saving ? t('rolePermissionsSaving') : t('rolePermissionsFormSave')}
            </Button>
          </SheetFooter>
        </form>
      </SheetContent>
    </Sheet>
  )
}
