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
import type { ApiDormitory } from '@/features/dormitory/types'

type AnnouncementFormSheetProps = {
  open: boolean
  onOpenChange: (open: boolean) => void
  isEdit: boolean
  dormitoryId: string
  onDormitoryIdChange: (dormitoryId: string) => void
  dormitories: ApiDormitory[]
  title: string
  onTitleChange: (value: string) => void
  content: string
  onContentChange: (value: string) => void
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
  dormitoryId,
  onDormitoryIdChange,
  dormitories,
  title,
  onTitleChange,
  content,
  onContentChange,
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
            <label className="flex flex-col gap-1.5 text-sm font-medium">
              {t('announcementFormDormitoryLabel')}
              {dormitories.length === 0 ? (
                <p className="text-xs font-normal text-muted-foreground">{t('announcementFormNoDormitories')}</p>
              ) : (
                <select
                  className="h-10 rounded-md border border-input bg-transparent px-3 text-sm"
                  value={dormitoryId}
                  onChange={(event) => onDormitoryIdChange(event.target.value)}
                  disabled={isEdit}
                >
                  <option value="">{t('announcementFormDormitoryPlaceholder')}</option>
                  {dormitories.map((dormitory) => (
                    <option key={dormitory.id} value={dormitory.id}>
                      {dormitory.name}
                    </option>
                  ))}
                </select>
              )}
            </label>

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
              {t('announcementFormContentLabel')}
              <textarea
                className="min-h-32 rounded-md border border-input bg-transparent px-3 py-2 text-sm"
                value={content}
                onChange={(event) => onContentChange(event.target.value)}
              />
            </label>

            <label className="flex flex-col gap-1.5 text-sm font-medium">
              {t('announcementFormStatusLabel')}
              <select
                className="h-10 rounded-md border border-input bg-transparent px-3 text-sm"
                value={isPublished ? 'published' : 'draft'}
                onChange={(event) => onIsPublishedChange(event.target.value === 'published')}
              >
                <option value="published">{t('announcementStatusPublished')}</option>
                <option value="draft">{t('announcementStatusDraft')}</option>
              </select>
            </label>

            <label className="flex flex-col gap-1.5 text-sm font-medium">
              {t('announcementFormDateLabel')}
              <input
                type="date"
                className="h-10 rounded-md border border-input bg-transparent px-3 text-sm"
                value={publishedDate}
                onChange={(event) => onPublishedDateChange(event.target.value)}
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
