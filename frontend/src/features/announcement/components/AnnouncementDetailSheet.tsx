import { Pin } from 'lucide-react'
import { useLanguage } from '@/shared/i18n/language'
import { Badge } from '@/shared/components/ui/badge'
import { Button } from '@/shared/components/ui/button'
import {
  Sheet,
  SheetContent,
  SheetDescription,
  SheetFooter,
  SheetHeader,
  SheetTitle,
} from '@/shared/components/ui/sheet'
import type { ApiAnnouncement } from '../types'
import { announcementCategoryLabelKeys, announcementCategoryVariant, toDateInputValue } from '../utils'

type AnnouncementDetailSheetProps = {
  open: boolean
  onOpenChange: (open: boolean) => void
  announcement: ApiAnnouncement | null
}

// Where an announcement is actually read. The list only carries a title, so
// without this a tenant would never see the body.
export function AnnouncementDetailSheet({ open, onOpenChange, announcement }: AnnouncementDetailSheetProps) {
  const { t } = useLanguage()

  return (
    <Sheet open={open} onOpenChange={onOpenChange}>
      <SheetContent>
        {announcement && (
          <div className="flex h-full flex-col gap-4">
            <SheetHeader>
              <div className="flex flex-wrap items-center gap-2">
                <Badge variant={announcementCategoryVariant(announcement.category)}>
                  {t(announcementCategoryLabelKeys[announcement.category] ?? 'announcementCategoryGeneral')}
                </Badge>
                {announcement.is_pinned && (
                  <Badge variant="outline" className="gap-1">
                    <Pin size={12} aria-hidden="true" />
                    {t('announcementPinned')}
                  </Badge>
                )}
                {!announcement.is_published && (
                  <Badge variant="outline">{t('announcementStatusDraft')}</Badge>
                )}
              </div>
              <SheetTitle>{announcement.title}</SheetTitle>
              <SheetDescription>
                {[announcement.dormitory_name, toDateInputValue(announcement.published_date)]
                  .filter(Boolean)
                  .join(' · ')}
              </SheetDescription>
            </SheetHeader>

            <div className="flex-1 overflow-y-auto pr-1">
              {announcement.content ? (
                <p className="whitespace-pre-wrap break-words text-sm leading-relaxed">{announcement.content}</p>
              ) : (
                <p className="metric-detail">{t('announcementNoContent')}</p>
              )}
            </div>

            <SheetFooter>
              <Button type="button" variant="outline" onClick={() => onOpenChange(false)}>
                {t('announcementDetailClose')}
              </Button>
            </SheetFooter>
          </div>
        )}
      </SheetContent>
    </Sheet>
  )
}
