import { ChevronLeft, ChevronRight, ChevronsLeft, ChevronsRight, Pencil, Trash2 } from 'lucide-react'
import { useLanguage } from '@/shared/i18n/language'
import { Badge } from '@/shared/components/ui/badge'
import { Card, CardContent, CardHeader, CardTitle } from '@/shared/components/ui/card'
import { Table, TableBody, TableCell, TableHead, TableHeader, TableRow } from '@/shared/components/ui/table'
import { Button } from '@/shared/components/ui/button'
import type { ApiAnnouncement } from '../types'
import { ANNOUNCEMENT_PAGE_SIZE_OPTIONS, toDateInputValue } from '../utils'

type AnnouncementListCardProps = {
  isLoading: boolean
  loadError: string | null
  deleteError: string | null
  announcements: ApiAnnouncement[] | null
  query: string
  onQueryChange: (query: string) => void
  filteredAnnouncements: ApiAnnouncement[]
  paginatedAnnouncements: ApiAnnouncement[]
  currentPage: number
  totalPages: number
  rangeStart: number
  rangeEnd: number
  pageSize: number
  onPageSizeChange: (size: number) => void
  onFirstPage: () => void
  onPrevPage: () => void
  onNextPage: () => void
  onLastPage: () => void
  deletingAnnouncementId: string | null
  onCreateAnnouncement: () => void
  onEditAnnouncement: (announcement: ApiAnnouncement) => void
  onDeleteAnnouncement: (announcement: ApiAnnouncement) => void
}

export function AnnouncementListCard({
  isLoading,
  loadError,
  deleteError,
  announcements,
  query,
  onQueryChange,
  filteredAnnouncements,
  paginatedAnnouncements,
  currentPage,
  totalPages,
  rangeStart,
  rangeEnd,
  pageSize,
  onPageSizeChange,
  onFirstPage,
  onPrevPage,
  onNextPage,
  onLastPage,
  deletingAnnouncementId,
  onCreateAnnouncement,
  onEditAnnouncement,
  onDeleteAnnouncement,
}: AnnouncementListCardProps) {
  const { t } = useLanguage()

  return (
    <Card>
      <CardHeader className="flex flex-row items-start justify-between gap-4">
        <div>
          <CardTitle>{t('menuAnnouncements')}</CardTitle>
        </div>
        <Button onClick={onCreateAnnouncement} disabled={isLoading}>
          {t('announcementCreate')}
        </Button>
      </CardHeader>
      <CardContent className="flex flex-col gap-4">
        {loadError && <p className="resource-error">{loadError}</p>}
        {deleteError && <p className="resource-error">{deleteError}</p>}

        {!loadError && isLoading && <p className="metric-detail">{t('loading')}</p>}

        {!loadError && !isLoading && announcements && announcements.length === 0 && (
          <p className="metric-detail">{t('announcementNoAnnouncements')}</p>
        )}

        {!loadError && !isLoading && announcements && announcements.length > 0 && (
          <>
            <label className="flex w-full max-w-md flex-col gap-1.5 text-sm font-medium">
              {t('announcementSearchLabel')}
              <input
                type="search"
                className="h-10 rounded-md border border-input bg-transparent px-3 text-sm"
                placeholder={t('announcementSearchPlaceholder')}
                value={query}
                onChange={(event) => onQueryChange(event.target.value)}
              />
            </label>

            {filteredAnnouncements.length === 0 && <p className="metric-detail">{t('announcementNoMatching')}</p>}

            {filteredAnnouncements.length > 0 && (
              <div className="table-wrap announcement-table-wrap">
                <Table>
                  <TableHeader>
                    <TableRow>
                      <TableHead>{t('announcementDormitoryColumn')}</TableHead>
                      <TableHead>{t('announcementTitleColumn')}</TableHead>
                      <TableHead>{t('announcementStatusColumn')}</TableHead>
                      <TableHead>{t('announcementDateColumn')}</TableHead>
                      <TableHead className="text-right">{t('announcementActionsColumn')}</TableHead>
                    </TableRow>
                  </TableHeader>
                  <TableBody>
                    {paginatedAnnouncements.map((announcement) => (
                      <TableRow key={announcement.id}>
                        <TableCell className="font-semibold">{announcement.dormitory_name || '—'}</TableCell>
                        <TableCell className="text-muted-foreground">{announcement.title}</TableCell>
                        <TableCell>
                          <Badge variant={announcement.is_published ? 'default' : 'outline'}>
                            {t(announcement.is_published ? 'announcementStatusPublished' : 'announcementStatusDraft')}
                          </Badge>
                        </TableCell>
                        <TableCell className="text-muted-foreground">
                          {toDateInputValue(announcement.published_date)}
                        </TableCell>
                        <TableCell className="text-right">
                          <div className="flex flex-wrap justify-end gap-2">
                            <Button
                              type="button"
                              size="icon"
                              variant="outline"
                              title={t('announcementEdit')}
                              aria-label={t('announcementEdit')}
                              onClick={() => onEditAnnouncement(announcement)}
                            >
                              <Pencil />
                            </Button>
                            <Button
                              type="button"
                              size="icon"
                              variant="destructive"
                              title={t('announcementDelete')}
                              aria-label={t('announcementDelete')}
                              onClick={() => onDeleteAnnouncement(announcement)}
                              disabled={deletingAnnouncementId === announcement.id}
                            >
                              <Trash2 />
                            </Button>
                          </div>
                        </TableCell>
                      </TableRow>
                    ))}
                  </TableBody>
                </Table>
              </div>
            )}

            {filteredAnnouncements.length > 0 && (
              <div className="flex flex-col gap-3 sm:flex-row sm:items-center sm:justify-between">
                <p className="text-sm text-muted-foreground">
                  {t('rolePermissionsShowingLabel')} {rangeStart}-{rangeEnd}{' '}
                  {t('rolePermissionsOfLabel')} {filteredAnnouncements.length} {t('rolePermissionsResultsLabel')}
                  {totalPages > 1 && (
                    <>
                      {' '}
                      · {t('rolePermissionsPageLabel')} {currentPage} / {totalPages}
                    </>
                  )}
                </p>
                <div className="flex flex-wrap items-center gap-3">
                  <label className="flex items-center gap-1.5 text-sm text-muted-foreground">
                    {t('rolePermissionsPageSizeLabel')}
                    <select
                      className="h-9 rounded-md border border-input bg-transparent px-2 text-sm"
                      value={pageSize}
                      onChange={(event) => onPageSizeChange(Number(event.target.value))}
                    >
                      {ANNOUNCEMENT_PAGE_SIZE_OPTIONS.map((size) => (
                        <option key={size} value={size}>
                          {size}
                        </option>
                      ))}
                    </select>
                  </label>

                  {totalPages > 1 && (
                    <div className="flex gap-2">
                      <Button
                        type="button"
                        size="icon"
                        variant="outline"
                        title={t('rolePermissionsFirstPage')}
                        aria-label={t('rolePermissionsFirstPage')}
                        disabled={currentPage <= 1}
                        onClick={onFirstPage}
                      >
                        <ChevronsLeft />
                      </Button>
                      <Button
                        type="button"
                        size="icon"
                        variant="outline"
                        title={t('rolePermissionsPrevPage')}
                        aria-label={t('rolePermissionsPrevPage')}
                        disabled={currentPage <= 1}
                        onClick={onPrevPage}
                      >
                        <ChevronLeft />
                      </Button>
                      <Button
                        type="button"
                        size="icon"
                        variant="outline"
                        title={t('rolePermissionsNextPage')}
                        aria-label={t('rolePermissionsNextPage')}
                        disabled={currentPage >= totalPages}
                        onClick={onNextPage}
                      >
                        <ChevronRight />
                      </Button>
                      <Button
                        type="button"
                        size="icon"
                        variant="outline"
                        title={t('rolePermissionsLastPage')}
                        aria-label={t('rolePermissionsLastPage')}
                        disabled={currentPage >= totalPages}
                        onClick={onLastPage}
                      >
                        <ChevronsRight />
                      </Button>
                    </div>
                  )}
                </div>
              </div>
            )}
          </>
        )}
      </CardContent>
    </Card>
  )
}
