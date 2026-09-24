import type { FormEvent } from 'react'
import { X } from 'lucide-react'
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
import type { ApiUser, FormManager } from '../types'
import { ManagerSearchSelect } from './ManagerSearchSelect'
import { DeductionPresetsEditor } from '@/features/moveout/DeductionPresetsEditor'
import type { PresetRow } from '@/features/moveout/utils'

type DormitoryFormSheetProps = {
  open: boolean
  onOpenChange: (open: boolean) => void
  isEdit: boolean
  name: string
  onNameChange: (name: string) => void
  address: string
  onAddressChange: (address: string) => void
  phone: string
  onPhoneChange: (phone: string) => void
  promptPayId: string
  onPromptPayIdChange: (promptPayId: string) => void
  description: string
  onDescriptionChange: (description: string) => void
  isActive: boolean
  onIsActiveChange: (isActive: boolean) => void
  overdueReminderEnabled: boolean
  onOverdueReminderEnabledChange: (enabled: boolean) => void
  presets: PresetRow[]
  onPresetsChange: (rows: PresetRow[]) => void
  presetsLoading: boolean
  managers: FormManager[]
  onAddManager: (user: ApiUser) => void
  onRemoveManager: (userId: string) => void
  saving: boolean
  error: string | null
  onSubmit: (event: FormEvent<HTMLFormElement>) => void
}

export function DormitoryFormSheet({
  open,
  onOpenChange,
  isEdit,
  name,
  onNameChange,
  address,
  onAddressChange,
  phone,
  onPhoneChange,
  promptPayId,
  onPromptPayIdChange,
  description,
  onDescriptionChange,
  isActive,
  onIsActiveChange,
  overdueReminderEnabled,
  onOverdueReminderEnabledChange,
  presets,
  onPresetsChange,
  presetsLoading,
  managers,
  onAddManager,
  onRemoveManager,
  saving,
  error,
  onSubmit,
}: DormitoryFormSheetProps) {
  const { t } = useLanguage()

  return (
    <Sheet open={open} onOpenChange={onOpenChange}>
      <SheetContent>
        <form className="flex h-full flex-col gap-4" onSubmit={onSubmit}>
          <SheetHeader>
            <SheetTitle>
              {isEdit ? t('dormitoryFormEditTitle') : t('dormitoryFormCreateTitle')}
            </SheetTitle>
            <SheetDescription>
              {isEdit ? t('dormitoryFormEditDescription') : t('dormitoryFormCreateDescription')}
            </SheetDescription>
          </SheetHeader>

          <div className="flex flex-1 flex-col gap-4 overflow-y-auto pr-1">
            <label className="flex flex-col gap-1.5 text-sm font-medium">
              {t('dormitoryFormNameLabel')}
              <input
                type="text"
                className="h-10 rounded-md border border-input bg-transparent px-3 text-sm"
                value={name}
                onChange={(event) => onNameChange(event.target.value)}
                autoFocus
              />
            </label>

            <label className="flex flex-col gap-1.5 text-sm font-medium">
              {t('dormitoryFormAddressLabel')}
              <textarea
                className="min-h-16 rounded-md border border-input bg-transparent px-3 py-2 text-sm"
                value={address}
                onChange={(event) => onAddressChange(event.target.value)}
              />
            </label>

            <label className="flex flex-col gap-1.5 text-sm font-medium">
              {t('dormitoryFormPhoneLabel')}
              <input
                type="text"
                className="h-10 rounded-md border border-input bg-transparent px-3 text-sm"
                value={phone}
                onChange={(event) => onPhoneChange(event.target.value)}
              />
            </label>

            <label className="flex flex-col gap-1.5 text-sm font-medium">
              {t('dormitoryFormPromptPayLabel')}
              <input
                type="text"
                inputMode="numeric"
                className="h-10 rounded-md border border-input bg-transparent px-3 text-sm"
                placeholder={t('dormitoryFormPromptPayPlaceholder')}
                value={promptPayId}
                onChange={(event) => onPromptPayIdChange(event.target.value)}
              />
              <span className="text-xs font-normal text-muted-foreground">{t('dormitoryFormPromptPayHint')}</span>
            </label>

            <label className="flex flex-col gap-1.5 text-sm font-medium">
              {t('dormitoryFormDescriptionLabel')}
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
              {t('dormitoryFormActiveLabel')}
            </label>

            <label className="flex items-start gap-2 text-sm font-medium">
              <input
                type="checkbox"
                className="mt-0.5 h-4 w-4 accent-primary"
                checked={overdueReminderEnabled}
                onChange={(event) => onOverdueReminderEnabledChange(event.target.checked)}
              />
              <span className="flex flex-col gap-0.5">
                {t('dormitoryFormOverdueReminderLabel')}
                <span className="text-xs font-normal text-muted-foreground">
                  {t('dormitoryFormOverdueReminderHint')}
                </span>
              </span>
            </label>

            <DeductionPresetsEditor rows={presets} onChange={onPresetsChange} loading={presetsLoading} />

            <div className="flex flex-col gap-1.5 text-sm font-medium">
              {t('dormitoryFormManagersLabel')}
              {managers.length > 0 && (
                <ul className="flex flex-col gap-1 rounded-md border border-input p-2">
                  {managers.map((manager) => (
                    <li
                      key={manager.id}
                      className="flex items-center gap-2 rounded px-1.5 py-1 text-sm font-normal hover:bg-muted/50"
                    >
                      <span>{manager.username}</span>
                      <span className="min-w-0 flex-1 truncate text-xs text-muted-foreground">{manager.email}</span>
                      <button
                        type="button"
                        className="shrink-0 rounded p-1 text-muted-foreground hover:bg-muted hover:text-foreground"
                        aria-label={t('dormitoryFormRemoveManager')}
                        onClick={() => onRemoveManager(manager.id)}
                      >
                        <X className="size-4" />
                      </button>
                    </li>
                  ))}
                </ul>
              )}
              <ManagerSearchSelect
                excludeIds={managers.map((manager) => manager.id)}
                onSelectUser={onAddManager}
                placeholder={t('dormitoryFormAddManager')}
                searchPlaceholder={t('pickerSearchUserPlaceholder')}
                noResultsLabel={t('pickerNoUsers')}
              />
            </div>
          </div>

          {error && <p className="resource-error">{error}</p>}

          <SheetFooter>
            <Button type="button" variant="outline" onClick={() => onOpenChange(false)}>
              {t('dormitoryFormCancel')}
            </Button>
            <Button type="submit" disabled={saving}>
              {saving ? t('dormitorySaving') : t('dormitoryFormSave')}
            </Button>
          </SheetFooter>
        </form>
      </SheetContent>
    </Sheet>
  )
}
