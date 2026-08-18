import { useState } from 'react'
import type { DateRange } from 'react-day-picker'
import { CalendarIcon, ChevronLeft, ChevronRight, ChevronsLeft, ChevronsRight, X } from 'lucide-react'
import { useLanguage } from '@/shared/i18n/language'
import { cn } from '@/shared/lib/utils'
import { Badge } from '@/shared/components/ui/badge'
import { Calendar } from '@/shared/components/ui/calendar'
import { Card, CardContent, CardHeader, CardTitle } from '@/shared/components/ui/card'
import { Popover, PopoverContent, PopoverTrigger } from '@/shared/components/ui/popover'
import { Table, TableBody, TableCell, TableHead, TableHeader, TableRow } from '@/shared/components/ui/table'
import { Button } from '@/shared/components/ui/button'
import type { ApiActivityLog } from '../types'
import { ACTIVITY_LOG_PAGE_SIZE_OPTIONS, activityLogActionVariant } from '../utils'

function parseDateInput(value: string): Date | undefined {
  if (!value) return undefined
  const [year, month, day] = value.split('-').map(Number)
  return new Date(year, month - 1, day)
}

function formatDateInput(date: Date | undefined): string {
  if (!date) return ''
  const year = date.getFullYear()
  const month = String(date.getMonth() + 1).padStart(2, '0')
  const day = String(date.getDate()).padStart(2, '0')
  return `${year}-${month}-${day}`
}

type ActivityLogListCardProps = {
  isLoading: boolean
  loadError: string | null
  logs: ApiActivityLog[] | null
  entityTypeFilter: string
  onEntityTypeFilterChange: (value: string) => void
  dateFrom: string
  onDateFromChange: (value: string) => void
  dateTo: string
  onDateToChange: (value: string) => void
  currentPage: number
  totalPages: number
  rangeStart: number
  rangeEnd: number
  totalItems: number
  pageSize: number
  onPageSizeChange: (size: number) => void
  onFirstPage: () => void
  onPrevPage: () => void
  onNextPage: () => void
  onLastPage: () => void
}

export function ActivityLogListCard({
  isLoading,
  loadError,
  logs,
  entityTypeFilter,
  onEntityTypeFilterChange,
  dateFrom,
  onDateFromChange,
  dateTo,
  onDateToChange,
  currentPage,
  totalPages,
  rangeStart,
  rangeEnd,
  totalItems,
  pageSize,
  onPageSizeChange,
  onFirstPage,
  onPrevPage,
  onNextPage,
  onLastPage,
}: ActivityLogListCardProps) {
  const { t, language } = useLanguage()
  const [datePickerOpen, setDatePickerOpen] = useState(false)

  const selectedRange: DateRange = {
    from: parseDateInput(dateFrom),
    to: parseDateInput(dateTo),
  }

  function handleRangeSelect(range: DateRange | undefined) {
    onDateFromChange(formatDateInput(range?.from))
    onDateToChange(formatDateInput(range?.to))
    // react-day-picker reports a single click as a same-day range (from === to),
    // so only auto-close once the user has actually picked two distinct ends.
    if (range?.from && range?.to && range.from.getTime() !== range.to.getTime()) {
      setDatePickerOpen(false)
    }
  }

  function clearDateRange() {
    onDateFromChange('')
    onDateToChange('')
  }

  const dateLocale = language === 'th' ? 'th-TH' : 'en-US'
  const dateRangeLabel =
    selectedRange.from || selectedRange.to
      ? [selectedRange.from, selectedRange.to]
          .map((date) => (date ? date.toLocaleDateString(dateLocale, { dateStyle: 'medium' }) : '?'))
          .join(' – ')
      : t('activityLogDateRangeLabel')

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
            <div className="flex flex-wrap items-end gap-4">
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

              <div className="flex flex-col gap-1.5 text-sm font-medium">
                {t('activityLogDateRangeLabel')}
                <div className="flex items-center gap-1.5">
                  <Popover open={datePickerOpen} onOpenChange={setDatePickerOpen}>
                    <PopoverTrigger asChild>
                      <Button
                        type="button"
                        variant="outline"
                        className={cn(
                          'h-10 justify-start gap-2 px-3 text-sm font-normal',
                          !(selectedRange.from || selectedRange.to) && 'text-muted-foreground'
                        )}
                      >
                        <CalendarIcon className="size-4 shrink-0" />
                        {dateRangeLabel}
                      </Button>
                    </PopoverTrigger>
                    <PopoverContent className="w-auto p-0" align="start">
                      <Calendar
                        mode="range"
                        numberOfMonths={2}
                        defaultMonth={selectedRange.from ?? new Date()}
                        selected={selectedRange}
                        onSelect={handleRangeSelect}
                      />
                    </PopoverContent>
                  </Popover>

                  {(dateFrom || dateTo) && (
                    <Button
                      type="button"
                      size="icon"
                      variant="ghost"
                      className="h-10 w-10 text-muted-foreground"
                      title={t('activityLogDateClear')}
                      aria-label={t('activityLogDateClear')}
                      onClick={clearDateRange}
                    >
                      <X className="size-4" />
                    </Button>
                  )}
                </div>
              </div>
            </div>

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
