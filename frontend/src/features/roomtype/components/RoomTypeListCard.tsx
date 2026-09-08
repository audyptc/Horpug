import {
  ArrowDown,
  ArrowUp,
  ArrowUpDown,
  ChevronLeft,
  ChevronRight,
  ChevronsLeft,
  ChevronsRight,
  Pencil,
  Trash2,
  X,
} from 'lucide-react'
import { useLanguage, type TranslationKey } from '@/shared/i18n/language'
import { Badge } from '@/shared/components/ui/badge'
import { Card, CardContent, CardHeader, CardTitle } from '@/shared/components/ui/card'
import { ColumnFilterMenu } from '@/shared/components/column-filter-menu'
import { Table, TableBody, TableCell, TableHead, TableHeader, TableRow } from '@/shared/components/ui/table'
import { Button } from '@/shared/components/ui/button'
import type { ApiRoomType } from '../types'
import {
  ROOM_TYPE_PAGE_SIZE_OPTIONS,
  isTextFilterKey,
  type RoomTypeColumnFilters,
  type RoomTypeSortDirection,
  type RoomTypeSortKey,
  type RoomTypeStatusFilter,
  type RoomTypeTextFilterKey,
} from '../utils'

const TEXT_COLUMNS: { key: RoomTypeSortKey; labelKey: TranslationKey }[] = [
  { key: 'name', labelKey: 'roomTypeNameColumn' },
  { key: 'dormitory', labelKey: 'roomTypeDormitoryColumn' },
]

const PRICE_COLUMN: { key: RoomTypeSortKey; labelKey: TranslationKey } = {
  key: 'price',
  labelKey: 'roomTypePriceColumn',
}

const STATUS_COLUMN: { key: RoomTypeSortKey; labelKey: TranslationKey } = {
  key: 'is_active',
  labelKey: 'roomTypeStatusColumn',
}

const TOTAL_COLUMN_COUNT = TEXT_COLUMNS.length + 3 // + price + status + actions

type RoomTypeListCardProps = {
  isLoading: boolean
  loadError: string | null
  deleteError: string | null
  query: string
  onQueryChange: (query: string) => void
  statusFilter: RoomTypeStatusFilter
  onStatusFilterChange: (value: RoomTypeStatusFilter) => void
  columnFilters: RoomTypeColumnFilters
  onColumnFilterChange: (key: RoomTypeTextFilterKey, value: string) => void
  hasFilters: boolean
  sortKey: RoomTypeSortKey
  sortDirection: RoomTypeSortDirection
  onSort: (key: RoomTypeSortKey) => void
  roomTypes: ApiRoomType[]
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
  deletingRoomTypeId: string | null
  onCreateRoomType: () => void
  onEditRoomType: (roomType: ApiRoomType) => void
  onDeleteRoomType: (roomType: ApiRoomType) => void
}

export function RoomTypeListCard({
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
  roomTypes,
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
  deletingRoomTypeId,
  onCreateRoomType,
  onEditRoomType,
  onDeleteRoomType,
}: RoomTypeListCardProps) {
  const { t } = useLanguage()

  // The column filters live in a table that scrolls, so on a narrow screen
  // they're off to the right and there's no way to tell what's applied.
  // Summarising them here keeps that visible and clearable at any size.
  const activeFilters: { id: string; label: string; onClear: () => void }[] = []

  if (statusFilter !== 'all') {
    activeFilters.push({
      id: 'is_active',
      label: `${t('roomTypeStatusColumn')}: ${statusFilter === 'active' ? t('statusActive') : t('statusInactive')}`,
      onClear: () => onStatusFilterChange('all'),
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

  function renderSortableHead(column: { key: RoomTypeSortKey; labelKey: TranslationKey }) {
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
              onTextChange={(value) => onColumnFilterChange(column.key as RoomTypeTextFilterKey, value)}
            />
          )}

          {column.key === 'is_active' && (
            <ColumnFilterMenu
              label={t('roomTypeStatusColumn')}
              optionValue={statusFilter}
              onOptionChange={onStatusFilterChange}
              options={[
                { value: 'all', label: t('filterAll') },
                { value: 'active', label: t('statusActive') },
                { value: 'inactive', label: t('statusInactive') },
              ]}
            />
          )}
        </div>
      </TableHead>
    )
  }

  return (
    <Card>
      <CardHeader className="flex flex-col gap-3 sm:flex-row sm:items-start sm:justify-between sm:gap-4">
        <div>
          <CardTitle>{t('menuRoomTypes')}</CardTitle>
        </div>
        <Button onClick={onCreateRoomType} disabled={isLoading}>
          {t('roomTypeCreate')}
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
                  {t('roomTypeSearchLabel')}
                  <input
                    type="search"
                    className="h-10 rounded-md border border-input bg-transparent px-3 text-sm"
                    placeholder={t('roomTypeSearchPlaceholder')}
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
              <div className="roomtype-table-wrap overflow-x-auto">
                <Table>
                  <TableHeader>
                    <TableRow>
                      {TEXT_COLUMNS.map((column) => renderSortableHead(column))}
                      {renderSortableHead(PRICE_COLUMN)}
                      {renderSortableHead(STATUS_COLUMN)}
                      <TableHead className="text-right">{t('roomTypeActionsColumn')}</TableHead>
                    </TableRow>
                  </TableHeader>
                  <TableBody>
                    {roomTypes.length === 0 && (
                      <TableRow>
                        {/* The table scrolls, so centring this across every
                            column would push it off a phone screen. Pin it to
                            the left edge instead. */}
                        <TableCell colSpan={TOTAL_COLUMN_COUNT} className="p-0">
                          <p className="metric-detail sticky left-0 px-3 py-6">
                            {hasFilters ? t('roomTypeNoMatching') : t('roomTypeNoRoomTypes')}
                          </p>
                        </TableCell>
                      </TableRow>
                    )}
                    {roomTypes.map((roomType) => (
                      <TableRow key={roomType.id}>
                        <TableCell className="font-semibold">{roomType.name}</TableCell>
                        <TableCell className="text-muted-foreground">
                          {roomType.dormitory_name || '—'}
                        </TableCell>
                        <TableCell className="text-muted-foreground">
                          {roomType.price.toLocaleString()}
                        </TableCell>
                        <TableCell>
                          <Badge variant={roomType.is_active ? 'default' : 'outline'}>
                            {roomType.is_active ? t('statusActive') : t('statusInactive')}
                          </Badge>
                        </TableCell>
                        <TableCell className="text-right">
                          <div className="flex flex-wrap justify-end gap-2">
                            <Button
                              type="button"
                              size="icon"
                              variant="outline"
                              title={t('roomTypeEdit')}
                              aria-label={t('roomTypeEdit')}
                              onClick={() => onEditRoomType(roomType)}
                            >
                              <Pencil />
                            </Button>
                            <Button
                              type="button"
                              size="icon"
                              variant="destructive"
                              title={t('roomTypeDelete')}
                              aria-label={t('roomTypeDelete')}
                              onClick={() => onDeleteRoomType(roomType)}
                              disabled={deletingRoomTypeId === roomType.id}
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
                      {ROOM_TYPE_PAGE_SIZE_OPTIONS.map((size) => (
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
