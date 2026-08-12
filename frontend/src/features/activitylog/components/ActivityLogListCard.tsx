import { ChevronLeft, ChevronRight } from 'lucide-react'
import { useLanguage } from '@/shared/i18n/language'
import { Badge } from '@/shared/components/ui/badge'
import { Card, CardContent, CardHeader, CardTitle } from '@/shared/components/ui/card'
import { Table, TableBody, TableCell, TableHead, TableHeader, TableRow } from '@/shared/components/ui/table'
import { Button } from '@/shared/components/ui/button'
import type { ApiActivityLog } from '../types'
import { ACTIVITY_LOG_PAGE_SIZE_OPTIONS, activityLogActionVariant } from '../utils'

type ActivityLogListCardProps = {
  isLoading: boolean
  loadError: string | null
  logs: ApiActivityLog[] | null
  entityTypeFilter: string
  onEntityTypeFilterChange: (value: string) => void
  currentPage: number
  totalPages: number
  rangeStart: number
  rangeEnd: number
  totalItems: number
  pageSize: number
  onPageSizeChange: (size: number) => void
  onPrevPage: () => void
  onNextPage: () => void
}

export function ActivityLogListCard({
  isLoading,
  loadError,
  logs,
  entityTypeFilter,
  onEntityTypeFilterChange,
  currentPage,
  totalPages,
  rangeStart,
  rangeEnd,
  totalItems,
  pageSize,
  onPageSizeChange,
  onPrevPage,
  onNextPage,
}: ActivityLogListCardProps) {
  const { t, language } = useLanguage()

  return (
    <Card>
      <CardHeader>
        <CardTitle>{t('menuActivityLogs')}</CardTitle>
      </CardHeader>
      <CardContent className="flex flex-col gap-4">
        {loadError && <p className="resource-error">{loadError}</p>}

        {!loadError && isLoading && <p className="metric-detail">{t('loading')}</p>}

        {!loadError && !isLoading && (
          <>
            <label className="flex w-full max-w-md flex-col gap-1.5 text-sm font-medium">
              {t('activityLogEntityTypeLabel')}
              <input
                type="search"
                className="h-10 rounded-md border border-input bg-transparent px-3 text-sm"
                placeholder={t('activityLogEntityTypePlaceholder')}
                value={entityTypeFilter}
                onChange={(event) => onEntityTypeFilterChange(event.target.value)}
              />
            </label>

            {logs && logs.length === 0 && <p className="metric-detail">{t('activityLogNoLogs')}</p>}

            {logs && logs.length > 0 && (
              <div className="table-wrap">
                <Table>
                  <TableHeader>
                    <TableRow>
                      <TableHead>{t('activityLogTimeColumn')}</TableHead>
                      <TableHead>{t('activityLogUserColumn')}</TableHead>
                      <TableHead>{t('activityLogActionColumn')}</TableHead>
                      <TableHead>{t('activityLogEntityTypeColumn')}</TableHead>
                      <TableHead>{t('activityLogDescriptionColumn')}</TableHead>
                      <TableHead>{t('activityLogIpColumn')}</TableHead>
                    </TableRow>
                  </TableHeader>
                  <TableBody>
                    {logs.map((log) => (
                      <TableRow key={log.id}>
                        <TableCell className="text-muted-foreground">
                          {new Date(log.created_at).toLocaleString(language === 'th' ? 'th-TH' : 'en-US')}
                        </TableCell>
                        <TableCell className="font-semibold">
                          {log.username || t('activityLogSystemUser')}
                        </TableCell>
                        <TableCell>
                          <Badge variant={activityLogActionVariant(log.action)}>{log.action}</Badge>
                        </TableCell>
                        <TableCell className="text-muted-foreground">{log.entity_type}</TableCell>
                        <TableCell className="text-muted-foreground">{log.description || '—'}</TableCell>
                        <TableCell className="text-muted-foreground">{log.ip_address || '—'}</TableCell>
                      </TableRow>
                    ))}
                  </TableBody>
                </Table>
              </div>
            )}

            {totalItems > 0 && (
              <div className="flex items-center justify-between gap-3">
                <p className="text-sm text-muted-foreground">
                  {t('rolePermissionsShowingLabel')} {rangeStart}-{rangeEnd}{' '}
                  {t('rolePermissionsOfLabel')} {totalItems} {t('rolePermissionsResultsLabel')}
                  {totalPages > 1 && (
                    <>
                      {' '}
                      · {t('rolePermissionsPageLabel')} {currentPage} / {totalPages}
                    </>
                  )}
                </p>
                <div className="flex items-center gap-3">
                  <label className="flex items-center gap-1.5 text-sm text-muted-foreground">
                    {t('rolePermissionsPageSizeLabel')}
                    <select
                      className="h-9 rounded-md border border-input bg-transparent px-2 text-sm"
                      value={pageSize}
                      onChange={(event) => onPageSizeChange(Number(event.target.value))}
                    >
                      {ACTIVITY_LOG_PAGE_SIZE_OPTIONS.map((size) => (
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
