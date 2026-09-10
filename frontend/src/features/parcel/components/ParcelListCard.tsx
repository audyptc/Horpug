import { ArrowDown, ArrowUp, ArrowUpDown, ChevronLeft, ChevronRight, ChevronsLeft, ChevronsRight, Pencil, Trash2, X } from 'lucide-react'
import { useLanguage, type TranslationKey } from '@/shared/i18n/language'
import { Badge } from '@/shared/components/ui/badge'
import { Card, CardContent, CardHeader, CardTitle } from '@/shared/components/ui/card'
import { ColumnFilterMenu } from '@/shared/components/column-filter-menu'
import { Table, TableBody, TableCell, TableHead, TableHeader, TableRow } from '@/shared/components/ui/table'
import { Button } from '@/shared/components/ui/button'
import type { ApiParcel, ParcelStatus } from '../types'
import {
  PARCEL_PAGE_SIZE_OPTIONS,
  PARCEL_STATUSES,
  isParcelTextFilterKey,
  toDateInputValue,
  type ParcelColumnFilters,
  type ParcelSortDirection,
  type ParcelSortKey,
  type ParcelStatusFilter,
  type ParcelTextFilterKey,
} from '../utils'

const parcelStatusLabelKeys: Record<ParcelStatus, TranslationKey> = {
  pending: 'parcelStatusPending',
  picked_up: 'parcelStatusPickedUp',
  returned: 'parcelStatusReturned',
}

const parcelStatusBadgeVariant: Record<ParcelStatus, 'default' | 'outline' | 'destructive' | 'secondary'> = {
  pending: 'outline',
  picked_up: 'default',
  returned: 'destructive',
}

const SORTABLE_COLUMNS: { key: ParcelSortKey; labelKey: TranslationKey }[] = [
  { key: 'tenant_name', labelKey: 'parcelTenantColumn' },
  { key: 'room_number', labelKey: 'parcelRoomColumn' },
  { key: 'courier', labelKey: 'parcelCourierColumn' },
  { key: 'tracking_number', labelKey: 'parcelTrackingNumberColumn' },
  { key: 'status', labelKey: 'parcelStatusColumn' },
  { key: 'received_date', labelKey: 'parcelReceivedDateColumn' },
]

type ParcelListCardProps = {
  isLoading: boolean
  loadError: string | null
  deleteError: string | null
  query: string
  onQueryChange: (query: string) => void
  statusFilter: ParcelStatusFilter
  onStatusFilterChange: (value: ParcelStatusFilter) => void
  columnFilters: ParcelColumnFilters
  onColumnFilterChange: (key: ParcelTextFilterKey, value: string) => void
  hasFilters: boolean
  sortKey: ParcelSortKey
  sortDirection: ParcelSortDirection
  onSort: (key: ParcelSortKey) => void
  parcels: ApiParcel[]
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
  deletingParcelId: string | null
  onCreateParcel: () => void
  onEditParcel: (parcel: ApiParcel) => void
  onDeleteParcel: (parcel: ApiParcel) => void
}

export function ParcelListCard({
  isLoading,
  loadError,
  deleteError,
  query,
  onQueryChange,
  statusFilter,
  onStatusFilterChange,
  columnFilters,
  onColumnFilterChange,
  hasFilters,
  sortKey,
  sortDirection,
  onSort,
  parcels,
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
  deletingParcelId,
  onCreateParcel,
  onEditParcel,
  onDeleteParcel,
}: ParcelListCardProps) {
  const { t } = useLanguage()

  // The column filters live in a table that scrolls, so on a narrow screen
  // they're off to the right and there's no way to tell what's applied.
  // Summarising them here keeps that visible and clearable at any size.
  const activeFilters: { id: string; label: string; onClear: () => void }[] = []

  if (statusFilter !== 'all') {
    activeFilters.push({
      id: 'status',
      label: `${t('parcelFilterStatusLabel')}: ${t(parcelStatusLabelKeys[statusFilter])}`,
      onClear: () => onStatusFilterChange('all'),
    })
  }

  for (const column of SORTABLE_COLUMNS) {
    const key = column.key
    if (!isParcelTextFilterKey(key)) continue

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
          <CardTitle>{t('menuParcels')}</CardTitle>
        </div>
        <Button onClick={onCreateParcel} disabled={isLoading}>
          {t('parcelCreate')}
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
                  {t('parcelSearchLabel')}
                  <input
                    type="search"
                    className="h-10 rounded-md border border-input bg-transparent px-3 text-sm"
                    placeholder={t('parcelSearchPlaceholder')}
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
              <div className="parcel-table-wrap overflow-x-auto">
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

                              {column.key === 'status' && (
                                <ColumnFilterMenu
                                  label={t('parcelFilterStatusLabel')}
                                  optionValue={statusFilter}
                                  onOptionChange={onStatusFilterChange}
                                  options={[
                                    { value: 'all', label: t('filterAll') },
                                    ...PARCEL_STATUSES.map((status) => ({
                                      value: status,
                                      label: t(parcelStatusLabelKeys[status]),
                                    })),
                                  ]}
                                />
                              )}

                              {isParcelTextFilterKey(column.key) && (
                                <ColumnFilterMenu
                                  label={t(column.labelKey)}
                                  textValue={columnFilters[column.key] ?? ''}
                                  onTextChange={(value) =>
                                    onColumnFilterChange(column.key as ParcelTextFilterKey, value)
                                  }
                                />
                              )}
                            </div>
                          </TableHead>
                        )
                      })}
                      <TableHead className="text-right">{t('parcelActionsColumn')}</TableHead>
                    </TableRow>
                  </TableHeader>
                  <TableBody>
                    {parcels.length === 0 && (
                      <TableRow>
                        {/* The table scrolls horizontally, so centring this
                            across every column would push it off a phone
                            screen. Pin it to the left edge instead. */}
                        <TableCell colSpan={SORTABLE_COLUMNS.length + 1} className="p-0">
                          <p className="metric-detail sticky left-0 px-3 py-6">
                            {hasFilters ? t('parcelNoMatching') : t('parcelNoParcels')}
                          </p>
                        </TableCell>
                      </TableRow>
                    )}
                    {parcels.map((parcel) => (
                      <TableRow key={parcel.id}>
                        <TableCell className="font-semibold">{parcel.tenant_name || '—'}</TableCell>
                        <TableCell className="text-muted-foreground">
                          {parcel.room_number
                            ? `${parcel.room_number}${parcel.dormitory_name ? ` (${parcel.dormitory_name})` : ''}`
                            : '—'}
                        </TableCell>
                        <TableCell className="text-muted-foreground">{parcel.courier || '—'}</TableCell>
                        <TableCell className="text-muted-foreground">{parcel.tracking_number || '—'}</TableCell>
                        <TableCell>
                          <Badge variant={parcelStatusBadgeVariant[parcel.status]}>
                            {t(parcelStatusLabelKeys[parcel.status])}
                          </Badge>
                        </TableCell>
                        <TableCell className="text-muted-foreground">
                          {toDateInputValue(parcel.received_date)}
                        </TableCell>
                        <TableCell className="text-right">
                          <div className="flex flex-nowrap justify-end gap-2">
                            <Button
                              type="button"
                              size="icon"
                              variant="outline"
                              title={t('parcelEdit')}
                              aria-label={t('parcelEdit')}
                              onClick={() => onEditParcel(parcel)}
                            >
                              <Pencil />
                            </Button>
                            <Button
                              type="button"
                              size="icon"
                              variant="destructive"
                              title={t('parcelDelete')}
                              aria-label={t('parcelDelete')}
                              onClick={() => onDeleteParcel(parcel)}
                              disabled={deletingParcelId === parcel.id}
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
                      {PARCEL_PAGE_SIZE_OPTIONS.map((size) => (
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
