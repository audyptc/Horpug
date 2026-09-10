import { ArrowDown, ArrowUp, ArrowUpDown, ChevronLeft, ChevronRight, ChevronsLeft, ChevronsRight, Pencil, Trash2, X } from 'lucide-react'
import { useLanguage, type TranslationKey } from '@/shared/i18n/language'
import { Badge } from '@/shared/components/ui/badge'
import { Card, CardContent, CardHeader, CardTitle } from '@/shared/components/ui/card'
import { ColumnFilterMenu } from '@/shared/components/column-filter-menu'
import { Table, TableBody, TableCell, TableHead, TableHeader, TableRow } from '@/shared/components/ui/table'
import { Button } from '@/shared/components/ui/button'
import type { ApiRepairRequest, RepairCategory, RepairStatus } from '../types'
import {
  REPAIR_CATEGORIES,
  REPAIR_REQUEST_PAGE_SIZE_OPTIONS,
  REPAIR_STATUSES,
  isRepairTextFilterKey,
  toDateInputValue,
  type RepairCategoryFilter,
  type RepairColumnFilters,
  type RepairSortDirection,
  type RepairSortKey,
  type RepairStatusFilter,
  type RepairTextFilterKey,
} from '../utils'

const repairCategoryLabelKeys: Record<RepairCategory, TranslationKey> = {
  electrical: 'repairCategoryElectrical',
  plumbing: 'repairCategoryPlumbing',
  furniture: 'repairCategoryFurniture',
  aircon: 'repairCategoryAircon',
  other: 'repairCategoryOther',
}

const repairStatusLabelKeys: Record<RepairStatus, TranslationKey> = {
  pending: 'repairStatusPending',
  in_progress: 'repairStatusInProgress',
  completed: 'repairStatusCompleted',
  cancelled: 'repairStatusCancelled',
}

const repairStatusBadgeVariant: Record<RepairStatus, 'default' | 'outline' | 'destructive' | 'secondary'> = {
  pending: 'outline',
  in_progress: 'secondary',
  completed: 'default',
  cancelled: 'destructive',
}

const SORTABLE_COLUMNS: { key: RepairSortKey; labelKey: TranslationKey }[] = [
  { key: 'room_number', labelKey: 'repairRoomColumn' },
  { key: 'tenant_name', labelKey: 'repairTenantColumn' },
  { key: 'category', labelKey: 'repairCategoryColumn' },
  { key: 'description', labelKey: 'repairDescriptionColumn' },
  { key: 'status', labelKey: 'repairStatusColumn' },
  { key: 'reported_date', labelKey: 'repairReportedDateColumn' },
]

type RepairRequestListCardProps = {
  isLoading: boolean
  loadError: string | null
  deleteError: string | null
  query: string
  onQueryChange: (query: string) => void
  categoryFilter: RepairCategoryFilter
  onCategoryFilterChange: (value: RepairCategoryFilter) => void
  statusFilter: RepairStatusFilter
  onStatusFilterChange: (value: RepairStatusFilter) => void
  columnFilters: RepairColumnFilters
  onColumnFilterChange: (key: RepairTextFilterKey, value: string) => void
  hasFilters: boolean
  sortKey: RepairSortKey
  sortDirection: RepairSortDirection
  onSort: (key: RepairSortKey) => void
  repairRequests: ApiRepairRequest[]
  total: number
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
  deletingRepairRequestId: string | null
  onCreateRepairRequest: () => void
  onEditRepairRequest: (repairRequest: ApiRepairRequest) => void
  onDeleteRepairRequest: (repairRequest: ApiRepairRequest) => void
}

export function RepairRequestListCard({
  isLoading,
  loadError,
  deleteError,
  query,
  onQueryChange,
  categoryFilter,
  onCategoryFilterChange,
  statusFilter,
  onStatusFilterChange,
  columnFilters,
  onColumnFilterChange,
  hasFilters,
  sortKey,
  sortDirection,
  onSort,
  repairRequests,
  total,
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
  deletingRepairRequestId,
  onCreateRepairRequest,
  onEditRepairRequest,
  onDeleteRepairRequest,
}: RepairRequestListCardProps) {
  const { t } = useLanguage()

  // The column filters live in a table that scrolls, so on a narrow screen
  // they're off to the right and there's no way to tell what's applied.
  // Summarising them here keeps that visible and clearable at any size.
  const activeFilters: { id: string; label: string; onClear: () => void }[] = []

  if (categoryFilter !== 'all') {
    activeFilters.push({
      id: 'category',
      label: `${t('repairFilterCategoryLabel')}: ${t(repairCategoryLabelKeys[categoryFilter])}`,
      onClear: () => onCategoryFilterChange('all'),
    })
  }

  if (statusFilter !== 'all') {
    activeFilters.push({
      id: 'status',
      label: `${t('repairFilterStatusLabel')}: ${t(repairStatusLabelKeys[statusFilter])}`,
      onClear: () => onStatusFilterChange('all'),
    })
  }

  for (const column of SORTABLE_COLUMNS) {
    const key = column.key
    if (!isRepairTextFilterKey(key)) continue

    const value = columnFilters[key]
    if (!value) continue

    activeFilters.push({
      id: key,
      label: `${t(column.labelKey)}: ${value}`,
      onClear: () => onColumnFilterChange(key, ''),
    })
  }

  return (
    <Card>
      <CardHeader className="flex flex-col gap-3 sm:flex-row sm:items-start sm:justify-between sm:gap-4">
        <div>
          <CardTitle>{t('menuRepairRequests')}</CardTitle>
        </div>
        <Button onClick={onCreateRepairRequest} disabled={isLoading}>
          {t('repairCreate')}
        </Button>
      </CardHeader>
      <CardContent className="flex flex-col gap-4">
        {loadError && <p className="resource-error">{loadError}</p>}
        {deleteError && <p className="resource-error">{deleteError}</p>}

        {!loadError && isLoading && <p className="metric-detail">{t('loading')}</p>}

        {!loadError && !isLoading && (
          <>
            <div className="overflow-hidden rounded-md border border-border">
              <div className="flex flex-col gap-3 border-b border-border bg-muted/40 p-3">
                <label className="flex w-full flex-col gap-1.5 text-sm font-medium sm:max-w-md">
                  {t('repairSearchLabel')}
                  <input
                    type="search"
                    className="h-10 rounded-md border border-input bg-transparent px-3 text-sm"
                    placeholder={t('repairSearchPlaceholder')}
                    value={query}
                    onChange={(event) => onQueryChange(event.target.value)}
                  />
                </label>

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
              <div className="repairrequest-table-wrap overflow-x-auto">
                <Table>
                  <TableHeader>
                    <TableRow>
                      {SORTABLE_COLUMNS.map((column) => {
                        const isSorted = sortKey === column.key
                        return (
                          <TableHead
                            key={column.key}
                            aria-sort={
                              isSorted ? (sortDirection === 'asc' ? 'ascending' : 'descending') : 'none'
                            }
                          >
                            <div className="flex items-center gap-1">
                              <button
                                type="button"
                                onClick={() => onSort(column.key)}
                                title={
                                  isSorted && sortDirection === 'asc'
                                    ? t('sortAscending')
                                    : t('sortDescending')
                                }
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

                              {column.key === 'category' && (
                                <ColumnFilterMenu
                                  label={t('repairFilterCategoryLabel')}
                                  optionValue={categoryFilter}
                                  onOptionChange={onCategoryFilterChange}
                                  options={[
                                    { value: 'all', label: t('filterAll') },
                                    ...REPAIR_CATEGORIES.map((category) => ({
                                      value: category,
                                      label: t(repairCategoryLabelKeys[category]),
                                    })),
                                  ]}
                                />
                              )}

                              {column.key === 'status' && (
                                <ColumnFilterMenu
                                  label={t('repairFilterStatusLabel')}
                                  optionValue={statusFilter}
                                  onOptionChange={onStatusFilterChange}
                                  options={[
                                    { value: 'all', label: t('filterAll') },
                                    ...REPAIR_STATUSES.map((status) => ({
                                      value: status,
                                      label: t(repairStatusLabelKeys[status]),
                                    })),
                                  ]}
                                />
                              )}

                              {isRepairTextFilterKey(column.key) && (
                                <ColumnFilterMenu
                                  label={t(column.labelKey)}
                                  textValue={columnFilters[column.key] ?? ''}
                                  onTextChange={(value) =>
                                    onColumnFilterChange(column.key as RepairTextFilterKey, value)
                                  }
                                />
                              )}
                            </div>
                          </TableHead>
                        )
                      })}
                      <TableHead className="text-right">{t('repairActionsColumn')}</TableHead>
                    </TableRow>
                  </TableHeader>
                  <TableBody>
                    {repairRequests.length === 0 && (
                      <TableRow>
                        {/* The table scrolls horizontally, so centring this
                            across every column would push it off a phone
                            screen. Pin it to the left edge instead. */}
                        <TableCell colSpan={SORTABLE_COLUMNS.length + 1} className="p-0">
                          <p className="metric-detail sticky left-0 px-3 py-6">
                            {hasFilters ? t('repairNoMatching') : t('repairNoRequests')}
                          </p>
                        </TableCell>
                      </TableRow>
                    )}
                    {repairRequests.map((repairRequest) => (
                      <TableRow key={repairRequest.id}>
                        <TableCell className="font-semibold">
                          {repairRequest.room_number || '—'}
                          {repairRequest.dormitory_name ? ` (${repairRequest.dormitory_name})` : ''}
                        </TableCell>
                        <TableCell className="text-muted-foreground">
                          {repairRequest.tenant_name || '—'}
                        </TableCell>
                        <TableCell>
                          <Badge variant="secondary">{t(repairCategoryLabelKeys[repairRequest.category])}</Badge>
                        </TableCell>
                        <TableCell className="text-muted-foreground">
                          {repairRequest.description || '—'}
                        </TableCell>
                        <TableCell>
                          <Badge variant={repairStatusBadgeVariant[repairRequest.status]}>
                            {t(repairStatusLabelKeys[repairRequest.status])}
                          </Badge>
                        </TableCell>
                        <TableCell className="text-muted-foreground">
                          {toDateInputValue(repairRequest.reported_date)}
                        </TableCell>
                        <TableCell className="text-right">
                          <div className="flex flex-nowrap justify-end gap-2">
                            <Button
                              type="button"
                              size="icon"
                              variant="outline"
                              title={t('repairEdit')}
                              aria-label={t('repairEdit')}
                              onClick={() => onEditRepairRequest(repairRequest)}
                            >
                              <Pencil />
                            </Button>
                            <Button
                              type="button"
                              size="icon"
                              variant="destructive"
                              title={t('repairDelete')}
                              aria-label={t('repairDelete')}
                              onClick={() => onDeleteRepairRequest(repairRequest)}
                              disabled={deletingRepairRequestId === repairRequest.id}
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
            </div>

            {total > 0 && (
              <div className="flex flex-col gap-3 sm:flex-row sm:items-center sm:justify-between">
                <p className="text-sm text-muted-foreground">
                  {t('rolePermissionsShowingLabel')} {rangeStart}-{rangeEnd}{' '}
                  {t('rolePermissionsOfLabel')} {total} {t('rolePermissionsResultsLabel')}
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
                      {REPAIR_REQUEST_PAGE_SIZE_OPTIONS.map((size) => (
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
