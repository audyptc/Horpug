import type { FormEvent } from 'react'
import { useLanguage } from '@/shared/i18n/language'
import { Button } from '@/shared/components/ui/button'
import { Combobox } from '@/shared/components/ui/combobox'
import { DatePickerField } from '@/shared/components/date-picker-field'
import {
  Sheet,
  SheetContent,
  SheetDescription,
  SheetFooter,
  SheetHeader,
  SheetTitle,
} from '@/shared/components/ui/sheet'
import { DormitorySearchSelect } from '@/features/dormitory/components/DormitorySearchSelect'
import type { ApiDormitory } from '@/features/dormitory/types'
import type { AnnouncementCategory } from '../types'
import { ANNOUNCEMENT_CATEGORIES, announcementCategoryLabelKeys } from '../utils'

type AnnouncementFormSheetProps = {
  open: boolean
  onOpenChange: (open: boolean) => void
  isEdit: boolean
  dormitoryName: string
  onDormitorySelect: (dormitory: ApiDormitory) => void
  title: string
  onTitleChange: (value: string) => void
  content: string
  onContentChange: (value: string) => void
  category: AnnouncementCategory
  onCategoryChange: (value: AnnouncementCategory) => void
  isPinned: boolean
  onIsPinnedChange: (value: boolean) => void
  isPublished: boolean
  onIsPublishedChange: (value: boolean) => void
  publishedDate: string
  onPublishedDateChange: (value: string) => void
  saving: boolean
  error: string | null
  onSubmit: (event: FormEvent<HTMLFormElement>) => void
}

export function AnnouncementFormSheet({
  open,
  onOpenChange,
  isEdit,
  dormitoryName,
  onDormitorySelect,
  title,
  onTitleChange,
  content,
  onContentChange,
  category,
  onCategoryChange,
  isPinned,
  onIsPinnedChange,
  isPublished,
  onIsPublishedChange,
  publishedDate,
  onPublishedDateChange,
  saving,
  error,
  onSubmit,
}: AnnouncementFormSheetProps) {
  const { t } = useLanguage()

  return (
    <Sheet open={open} onOpenChange={onOpenChange}>
      <SheetContent>
        <form className="flex h-full flex-col gap-4" onSubmit={onSubmit}>
          <SheetHeader>
            <SheetTitle>{isEdit ? t('announcementFormEditTitle') : t('announcementFormCreateTitle')}</SheetTitle>
            <SheetDescription>
              {isEdit ? t('announcementFormEditDescription') : t('announcementFormCreateDescription')}
            </SheetDescription>
          </SheetHeader>

          <div className="flex flex-1 flex-col gap-4 overflow-y-auto pr-1">
            <div className="flex flex-col gap-1.5 text-sm font-medium">
              {t('announcementFormDormitoryLabel')}
              <DormitorySearchSelect
                selectedLabel={dormitoryName}
                onSelectDormitory={onDormitorySelect}
                placeholder={t('announcementFormDormitoryPlaceholder')}
                searchPlaceholder={t('announcementFormDormitorySearchPlaceholder')}
                noResultsLabel={t('announcementFormDormitoryNoResults')}
                disabled={isEdit}
              />
            </div>

            <label className="flex flex-col gap-1.5 text-sm font-medium">
              {t('announcementFormTitleLabel')}
              <input
                type="text"
                className="h-10 rounded-md border border-input bg-transparent px-3 text-sm"
                value={title}
                onChange={(event) => onTitleChange(event.target.value)}
              />
            </label>

            <label className="flex flex-col gap-1.5 text-sm font-medium">
              {t('announcementFormCategoryLabel')}
              <Combobox
                options={ANNOUNCEMENT_CATEGORIES.map((value) => ({
                  value,
                  label: t(announcementCategoryLabelKeys[value]),
                }))}
                value={category}
                onChange={(value) => onCategoryChange(value as AnnouncementCategory)}
                placeholder={t('announcementFormCategoryLabel')}
                searchPlaceholder={t('announcementFormCategorySearchPlaceholder')}
                emptyText={t('announcementFormCategoryNoResults')}
              />
            </label>

            <label className="flex flex-col gap-1.5 text-sm font-medium">
              {t('announcementFormContentLabel')}
              <textarea
                className="min-h-32 rounded-md border border-input bg-transparent px-3 py-2 text-sm"
                value={content}
                onChange={(event) => onContentChange(event.target.value)}
              />
            </label>

            <label className="flex flex-col gap-1.5 text-sm font-medium">
              {t('announcementFormStatusLabel')}
              <Combobox
                options={[
                  { value: 'published', label: t('announcementStatusPublished') },
                  { value: 'draft', label: t('announcementStatusDraft') },
                ]}
                value={isPublished ? 'published' : 'draft'}
                onChange={(value) => onIsPublishedChange(value === 'published')}
                placeholder={t('announcementFormStatusLabel')}
                searchPlaceholder={t('announcementFormStatusSearchPlaceholder')}
                emptyText={t('announcementFormStatusNoResults')}
              />
            </label>

            <label className="flex items-center gap-2 text-sm font-medium">
              <input
                type="checkbox"
                className="h-4 w-4 accent-primary"
                checked={isPinned}
                onChange={(event) => onIsPinnedChange(event.target.checked)}
              />
              {t('announcementFormPinnedLabel')}
            </label>

            <label className="flex flex-col gap-1.5 text-sm font-medium">
              {t('announcementFormDateLabel')}
              <DatePickerField
                value={publishedDate}
                onChange={onPublishedDateChange}
                placeholder={t('announcementFormDateLabel')}
              />
            </label>
          </div>

          {error && <p className="resource-error">{error}</p>}

          <SheetFooter>
            <Button type="button" variant="outline" onClick={() => onOpenChange(false)}>
              {t('announcementFormCancel')}
            </Button>
            <Button type="submit" disabled={saving}>
              {saving ? t('announcementSaving') : t('announcementFormSave')}
            </Button>
          </SheetFooter>
        </form>
      </SheetContent>
    </Sheet>
  )
}
