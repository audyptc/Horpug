import { useState } from 'react'
import type { DateRange } from 'react-day-picker'
import {
  ArrowDown,
  ArrowUp,
  ArrowUpDown,
  CalendarIcon,
  ChevronLeft,
  ChevronRight,
  ChevronsLeft,
  ChevronsRight,
  X,
} from 'lucide-react'
import { useLanguage, type TranslationKey } from '@/shared/i18n/language'
import { cn } from '@/shared/lib/utils'
import { Badge } from '@/shared/components/ui/badge'
import { Calendar } from '@/shared/components/ui/calendar'
import { Card, CardContent, CardHeader, CardTitle } from '@/shared/components/ui/card'
import { ColumnFilterMenu } from '@/shared/components/column-filter-menu'
import { Popover, PopoverContent, PopoverTrigger } from '@/shared/components/ui/popover'
import { Table, TableBody, TableCell, TableHead, TableHeader, TableRow } from '@/shared/components/ui/table'
import { Button } from '@/shared/components/ui/button'
import type { ApiActivityLog } from '../types'
import {
  ACTIVITY_LOG_PAGE_SIZE_OPTIONS,
  activityLogActionVariant,
  isTextFilterKey,
  type ActivityLogColumnFilters,
  type ActivityLogSortDirection,
  type ActivityLogSortKey,
  type ActivityLogTextFilterKey,
} from '../utils'

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

const TIME_COLUMN: { key: ActivityLogSortKey; labelKey: TranslationKey } = {
  key: 'created_at',
  labelKey: 'activityLogTimeColumn',
}

const TEXT_COLUMNS: { key: ActivityLogSortKey; labelKey: TranslationKey }[] = [
  { key: 'username', labelKey: 'activityLogUserColumn' },
  { key: 'action', labelKey: 'activityLogActionColumn' },
  { key: 'entity_type', labelKey: 'activityLogEntityTypeColumn' },
  { key: 'description', labelKey: 'activityLogDescriptionColumn' },
  { key: 'ip_address', labelKey: 'activityLogIpColumn' },
]

const TOTAL_COLUMN_COUNT = 1 + TEXT_COLUMNS.length // time + the rest; no actions column

type ActivityLogListCardProps = {
  isLoading: boolean
  loadError: string | null
  logs: ApiActivityLog[]
  query: string
  onQueryChange: (query: string) => void
  columnFilters: ActivityLogColumnFilters
  onColumnFilterChange: (key: ActivityLogTextFilterKey, value: string) => void
  dateFrom: string
  onDateFromChange: (value: string) => void
  dateTo: string
  onDateToChange: (value: string) => void
  hasFilters: boolean
  sortKey: ActivityLogSortKey
  sortDirection: ActivityLogSortDirection
  onSort: (key: ActivityLogSortKey) => void
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
  query,
  onQueryChange,
  columnFilters,
  onColumnFilterChange,
  dateFrom,
  onDateFromChange,
  dateTo,
  onDateToChange,
  hasFilters,
  sortKey,
  sortDirection,
  onSort,
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

  // The column filters live in a table that scrolls, so on a narrow screen
  // they're off to the right and there's no way to tell what's applied.
  // Summarising them here keeps that visible and clearable at any size.
  const activeFilters: { id: string; label: string; onClear: () => void }[] = []

  if (dateFrom || dateTo) {
    activeFilters.push({
      id: 'date_range',
      label: `${t('activityLogDateRangeLabel')}: ${dateRangeLabel}`,
      onClear: clearDateRange,
    })
  }

  for (const column of TEXT_COLUMNS) {
    const key = column.key
    if (!isTextFilterKey(key)) continue

    const value = columnFilters[key]
    if (!value) continue

    activeFilters.push({
      id: key,
      label: `${t(column.labelKey)}: ${value}`,
      onClear: () => onColumnFilterChange(key, ''),
    })
  }

  function renderSortableHead(column: { key: ActivityLogSortKey; labelKey: TranslationKey }) {
    const isSorted = sortKey === column.key
    return (
      <TableHead
        key={column.key}
        aria-sort={isSorted ? (sortDirection === 'asc' ? 'ascending' : 'descending') : 'none'}
      >
        <div className="flex items-center gap-1">
          <button
            type="button"
            onClick={() => onSort(column.key)}
            title={isSorted && sortDirection === 'asc' ? t('sortAscending') : t('sortDescending')}
            className="inline-flex items-center gap-1 whitespace-nowrap transition-colors hover:text-foreground"
          >
            {t(column.labelKey)}
            {isSorted ? (
              sortDirection === 'asc' ? (
                <ArrowUp size={13} className="shrink-0" />
              ) : (
                <ArrowDown size={13} className="shrink-0" />
              )
            ) : (
              <ArrowUpDown size={13} className="shrink-0 opacity-40" />
            )}
          </button>

          {isTextFilterKey(column.key) && (
            <ColumnFilterMenu
              label={t(column.labelKey)}
              textValue={columnFilters[column.key] ?? ''}
              onTextChange={(value) => onColumnFilterChange(column.key as ActivityLogTextFilterKey, value)}
            />
          )}
        </div>
      </TableHead>
    )
  }

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
            <div className="overflow-hidden rounded-md border border-border">
              <div className="flex flex-col gap-3 border-b border-border bg-muted/40 p-3">
                <div className="flex flex-wrap items-end gap-4">
                  <label className="flex w-full flex-col gap-1.5 text-sm font-medium sm:max-w-md">
                    {t('activityLogSearchLabel')}
                    <input
                      type="search"
                      className="h-10 rounded-md border border-input bg-transparent px-3 text-sm"
                      placeholder={t('activityLogSearchPlaceholder')}
                      value={query}
                      onChange={(event) => onQueryChange(event.target.value)}
                    />
                  </label>

                  <div className="flex min-w-0 flex-col gap-1.5 text-sm font-medium">
                    {t('activityLogDateRangeLabel')}
                    <div className="flex min-w-0 items-center gap-1.5">
                      <Popover open={datePickerOpen} onOpenChange={setDatePickerOpen}>
                        <PopoverTrigger asChild>
                          <Button
                            type="button"
                            variant="outline"
                            className={cn(
                              'h-10 min-w-0 shrink justify-start gap-2 px-3 text-sm font-normal',
                              !(selectedRange.from || selectedRange.to) && 'text-muted-foreground'
                            )}
                          >
                            <CalendarIcon className="size-4 shrink-0" />
                            <span className="truncate">{dateRangeLabel}</span>
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

                {activeFilters.length > 0 && (
                  <div className="flex flex-wrap items-center gap-1.5">
                    {activeFilters.map((filter) => (
                      <span
                        key={filter.id}
                        className="inline-flex max-w-full items-center gap-1 rounded-full border border-border bg-background py-0.5 pl-2.5 pr-1 text-xs"
                      >
                        <span className="truncate">{filter.label}</span>
                        <button
                          type="button"
                          onClick={filter.onClear}
                          title={t('filterClear')}
                          aria-label={`${t('filterClear')}: ${filter.label}`}
                          className="inline-flex size-5 shrink-0 items-center justify-center rounded-full text-muted-foreground transition-colors hover:bg-accent hover:text-foreground"
                        >
                          <X size={12} />
                        </button>
                      </span>
                    ))}
                  </div>
                )}
              </div>

              {/* Rendered even with no matches: the column filters live in the
                  header, so hiding it would strand the user with no way to
                  widen the filter again. */}
              <div className="activitylog-table-wrap overflow-x-auto">
                <Table>
                  <TableHeader>
                    <TableRow>
                      {renderSortableHead(TIME_COLUMN)}
                      {TEXT_COLUMNS.map((column) => renderSortableHead(column))}
                    </TableRow>
                  </TableHeader>
                  <TableBody>
                    {logs.length === 0 && (
                      <TableRow>
                        {/* The table scrolls, so centring this across every
                            column would push it off a phone screen. Pin it to
                            the left edge instead. */}
                        <TableCell colSpan={TOTAL_COLUMN_COUNT} className="p-0">
                          <p className="metric-detail sticky left-0 px-3 py-6">
                            {hasFilters ? t('activityLogNoMatching') : t('activityLogNoLogs')}
                          </p>
                        </TableCell>
                      </TableRow>
                    )}
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
            </div>

            {totalItems > 0 && (
              <div className="flex flex-col gap-3 sm:flex-row sm:items-center sm:justify-between">
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
                <div className="flex flex-wrap items-center gap-3">
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
