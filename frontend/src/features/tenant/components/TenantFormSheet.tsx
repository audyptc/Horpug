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

type TenantFormSheetProps = {
  open: boolean
  onOpenChange: (open: boolean) => void
  isEdit: boolean
  firstName: string
  onFirstNameChange: (value: string) => void
  lastName: string
  onLastNameChange: (value: string) => void
  phone: string
  onPhoneChange: (value: string) => void
  lineId: string
  onLineIdChange: (value: string) => void
  idCard: string
  onIdCardChange: (value: string) => void
  email: string
  onEmailChange: (value: string) => void
  emergencyContact: string
  onEmergencyContactChange: (value: string) => void
  note: string
  onNoteChange: (value: string) => void
  isActive: boolean
  onIsActiveChange: (isActive: boolean) => void
  saving: boolean
  error: string | null
  onSubmit: (event: FormEvent<HTMLFormElement>) => void
}

export function TenantFormSheet({
  open,
  onOpenChange,
  isEdit,
  firstName,
  onFirstNameChange,
  lastName,
  onLastNameChange,
  phone,
  onPhoneChange,
  lineId,
  onLineIdChange,
  idCard,
  onIdCardChange,
  email,
  onEmailChange,
  emergencyContact,
  onEmergencyContactChange,
  note,
  onNoteChange,
  isActive,
  onIsActiveChange,
  saving,
  error,
  onSubmit,
}: TenantFormSheetProps) {
  const { t } = useLanguage()

  return (
    <Sheet open={open} onOpenChange={onOpenChange}>
      <SheetContent>
        <form className="flex h-full flex-col gap-4" onSubmit={onSubmit}>
          <SheetHeader>
            <SheetTitle>{isEdit ? t('tenantFormEditTitle') : t('tenantFormCreateTitle')}</SheetTitle>
            <SheetDescription>
              {isEdit ? t('tenantFormEditDescription') : t('tenantFormCreateDescription')}
            </SheetDescription>
          </SheetHeader>

          <div className="flex flex-1 flex-col gap-4 overflow-y-auto pr-1">
            <label className="flex flex-col gap-1.5 text-sm font-medium">
              {t('tenantFormFirstNameLabel')}
              <input
                type="text"
                className="h-10 rounded-md border border-input bg-transparent px-3 text-sm"
                value={firstName}
                onChange={(event) => onFirstNameChange(event.target.value)}
                autoFocus
              />
            </label>

            <label className="flex flex-col gap-1.5 text-sm font-medium">
              {t('tenantFormLastNameLabel')}
              <input
                type="text"
                className="h-10 rounded-md border border-input bg-transparent px-3 text-sm"
                value={lastName}
                onChange={(event) => onLastNameChange(event.target.value)}
              />
            </label>

            <label className="flex flex-col gap-1.5 text-sm font-medium">
              {t('tenantFormPhoneLabel')}
              <input
                type="text"
                className="h-10 rounded-md border border-input bg-transparent px-3 text-sm"
                value={phone}
                onChange={(event) => onPhoneChange(event.target.value)}
              />
            </label>

            <label className="flex flex-col gap-1.5 text-sm font-medium">
              {t('tenantFormLineIdLabel')}
              <input
                type="text"
                className="h-10 rounded-md border border-input bg-transparent px-3 text-sm"
                value={lineId}
                onChange={(event) => onLineIdChange(event.target.value)}
              />
            </label>

            <label className="flex flex-col gap-1.5 text-sm font-medium">
              {t('tenantFormIdCardLabel')}
              <input
                type="text"
                className="h-10 rounded-md border border-input bg-transparent px-3 text-sm"
                value={idCard}
                onChange={(event) => onIdCardChange(event.target.value)}
              />
            </label>

            <label className="flex flex-col gap-1.5 text-sm font-medium">
              {t('tenantFormEmailLabel')}
              <input
                type="email"
                className="h-10 rounded-md border border-input bg-transparent px-3 text-sm"
                value={email}
                onChange={(event) => onEmailChange(event.target.value)}
              />
            </label>

            <label className="flex flex-col gap-1.5 text-sm font-medium">
              {t('tenantFormEmergencyContactLabel')}
              <input
                type="text"
                className="h-10 rounded-md border border-input bg-transparent px-3 text-sm"
                value={emergencyContact}
                onChange={(event) => onEmergencyContactChange(event.target.value)}
              />
            </label>

            <label className="flex flex-col gap-1.5 text-sm font-medium">
              {t('tenantFormNoteLabel')}
              <textarea
                className="min-h-20 rounded-md border border-input bg-transparent px-3 py-2 text-sm"
                value={note}
                onChange={(event) => onNoteChange(event.target.value)}
              />
            </label>

            <label className="flex items-center gap-2 text-sm font-medium">
              <input
                type="checkbox"
                className="h-4 w-4 accent-primary"
                checked={isActive}
                onChange={(event) => onIsActiveChange(event.target.checked)}
              />
              {t('tenantFormActiveLabel')}
            </label>
          </div>

          {error && <p className="resource-error">{error}</p>}

          <SheetFooter>
            <Button type="button" variant="outline" onClick={() => onOpenChange(false)}>
              {t('tenantFormCancel')}
            </Button>
            <Button type="submit" disabled={saving}>
              {saving ? t('tenantSaving') : t('tenantFormSave')}
            </Button>
          </SheetFooter>
        </form>
      </SheetContent>
    </Sheet>
  )
}
