import { ArrowDown, ArrowUp, ArrowUpDown, CheckCircle2, ChevronLeft, ChevronRight, ChevronsLeft, ChevronsRight, Link2, Pencil, Trash2, Unlink, X } from 'lucide-react'
import { useLanguage } from '@/shared/i18n/language'
import type { TranslationKey } from '@/shared/i18n/language'
import { Badge } from '@/shared/components/ui/badge'
import { Card, CardContent, CardHeader, CardTitle } from '@/shared/components/ui/card'
import { ColumnFilterMenu } from '@/shared/components/column-filter-menu'
import { Table, TableBody, TableCell, TableHead, TableHeader, TableRow } from '@/shared/components/ui/table'
import { Button } from '@/shared/components/ui/button'
import type { ApiTenant } from '../types'
import {
  TENANT_PAGE_SIZE_OPTIONS,
  isTextFilterKey,
  type TenantColumnFilters,
  type TenantLineFilter,
  type TenantSortDirection,
  type TenantSortKey,
  type TenantStatusFilter,
  type TenantTextFilterKey,
} from '../utils'

const SORTABLE_COLUMNS: { key: TenantSortKey; labelKey: TranslationKey }[] = [
  { key: 'first_name', labelKey: 'tenantFirstNameColumn' },
  { key: 'last_name', labelKey: 'tenantLastNameColumn' },
  { key: 'phone', labelKey: 'tenantPhoneColumn' },
  { key: 'line_id', labelKey: 'tenantLineIdColumn' },
  { key: 'id_card', labelKey: 'tenantIdCardColumn' },
  { key: 'email', labelKey: 'tenantEmailColumn' },
  { key: 'is_active', labelKey: 'tenantActiveColumn' },
]

type TenantListCardProps = {
  isLoading: boolean
  loadError: string | null
  deleteError: string | null
  query: string
  onQueryChange: (query: string) => void
  statusFilter: TenantStatusFilter
  onStatusFilterChange: (value: TenantStatusFilter) => void
  lineFilter: TenantLineFilter
  onLineFilterChange: (value: TenantLineFilter) => void
  columnFilters: TenantColumnFilters
  onColumnFilterChange: (key: TenantTextFilterKey, value: string) => void
  hasFilters: boolean
  sortKey: TenantSortKey
  sortDirection: TenantSortDirection
  onSort: (key: TenantSortKey) => void
  tenants: ApiTenant[]
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
  deletingTenantId: string | null
  onCreateTenant: () => void
  onEditTenant: (tenant: ApiTenant) => void
  onDeleteTenant: (tenant: ApiTenant) => void
  onCopyLineLink: (tenant: ApiTenant) => void
  onUnlinkLine: (tenant: ApiTenant) => void
}

export function TenantListCard({
  isLoading,
  loadError,
  deleteError,
  query,
  onQueryChange,
  statusFilter,
  onStatusFilterChange,
  lineFilter,
  onLineFilterChange,
  columnFilters,
  onColumnFilterChange,
  hasFilters,
  sortKey,
  sortDirection,
  onSort,
  tenants,
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
  deletingTenantId,
  onCreateTenant,
  onEditTenant,
  onDeleteTenant,
  onCopyLineLink,
  onUnlinkLine,
}: TenantListCardProps) {
  const { t } = useLanguage()

  // The column filters live in a table that is 60rem wide and scrolls, so on a
  // narrow screen they're off to the right and there's no way to tell what's
  // applied. Summarising them here keeps that visible and clearable at any size.
  const activeFilters: { id: string; label: string; onClear: () => void }[] = []

  if (statusFilter !== 'all') {
    activeFilters.push({
      id: 'is_active',
      label: `${t('tenantFilterStatusLabel')}: ${statusFilter === 'active' ? t('statusActive') : t('statusInactive')}`,
      onClear: () => onStatusFilterChange('all'),
    })
  }

  if (lineFilter !== 'all') {
    activeFilters.push({
      id: 'line_linked',
      label: `${t('tenantFilterLineLabel')}: ${lineFilter === 'linked' ? t('tenantFilterLineLinked') : t('tenantFilterLineUnlinked')}`,
      onClear: () => onLineFilterChange('all'),
    })
  }

  for (const column of SORTABLE_COLUMNS) {
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

  return (
    <Card>
      <CardHeader className="flex flex-col gap-3 sm:flex-row sm:items-start sm:justify-between sm:gap-4">
        <div>
          <CardTitle>{t('menuTenants')}</CardTitle>
        </div>
        <Button onClick={onCreateTenant} disabled={isLoading}>
          {t('tenantCreate')}
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
                  {t('tenantSearchLabel')}
                  <input
                    type="search"
                    className="h-10 rounded-md border border-input bg-transparent px-3 text-sm"
                    placeholder={t('tenantSearchPlaceholder')}
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
              <div className="tenant-table-wrap overflow-x-auto">
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

                              {column.key === 'line_id' && (
                                <ColumnFilterMenu
                                  label={t('tenantFilterLineLabel')}
                                  optionValue={lineFilter}
                                  onOptionChange={onLineFilterChange}
                                  options={[
                                    { value: 'all', label: t('filterAll') },
                                    { value: 'linked', label: t('tenantFilterLineLinked') },
                                    { value: 'unlinked', label: t('tenantFilterLineUnlinked') },
                                  ]}
                                  textValue={columnFilters.line_id ?? ''}
                                  onTextChange={(value) => onColumnFilterChange('line_id', value)}
                                />
                              )}

                              {isTextFilterKey(column.key) && column.key !== 'line_id' && (
                                <ColumnFilterMenu
                                  label={t(column.labelKey)}
                                  textValue={columnFilters[column.key] ?? ''}
                                  onTextChange={(value) => onColumnFilterChange(column.key as TenantTextFilterKey, value)}
                                />
                              )}

                              {column.key === 'is_active' && (
                                <ColumnFilterMenu
                                  label={t('tenantFilterStatusLabel')}
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
                      })}
                      <TableHead className="text-right">{t('tenantActionsColumn')}</TableHead>
                    </TableRow>
                  </TableHeader>
                  <TableBody>
                    {tenants.length === 0 && (
                      <TableRow>
                        {/* The table is 60rem wide and scrolls, so centring
                            this across every column would push it off a phone
                            screen. Pin it to the left edge instead. */}
                        <TableCell colSpan={SORTABLE_COLUMNS.length + 1} className="p-0">
                          <p className="metric-detail sticky left-0 px-3 py-6">
                            {hasFilters ? t('tenantNoMatching') : t('tenantNoTenants')}
                          </p>
                        </TableCell>
                      </TableRow>
                    )}
                    {tenants.map((tenant) => (
                      <TableRow key={tenant.id}>
                        <TableCell className="font-semibold">{tenant.first_name}</TableCell>
                        <TableCell className="font-semibold">{tenant.last_name}</TableCell>
                        <TableCell className="text-muted-foreground">{tenant.phone || '—'}</TableCell>
                        <TableCell className="text-muted-foreground">
                          <div className="flex items-center gap-1.5">
                            {tenant.line_id || '—'}
                            {tenant.line_user_id && (
                              <CheckCircle2
                                className="h-4 w-4 text-emerald-600"
                                aria-label={t('tenantLineLinked')}
                              >
                                <title>{t('tenantLineLinked')}</title>
                              </CheckCircle2>
                            )}
                          </div>
                        </TableCell>
                        <TableCell className="text-muted-foreground">{tenant.id_card || '—'}</TableCell>
                        <TableCell className="text-muted-foreground">{tenant.email || '—'}</TableCell>
                        <TableCell>
                          <Badge variant={tenant.is_active ? 'default' : 'outline'}>
                            {tenant.is_active ? t('statusActive') : t('statusInactive')}
                          </Badge>
                        </TableCell>
                        <TableCell className="text-right">
                          {/* Never wrap: a squeezed actions column would stack
                              the buttons and blow up every row's height. The
                              table scrolls horizontally instead. */}
                          <div className="flex flex-nowrap justify-end gap-2">
                            <Button
                              type="button"
                              size="icon"
                              variant="outline"
                              title={tenant.line_id ? t('tenantCopyLineLink') : t('tenantLineLinkRequiresLineId')}
                              aria-label={tenant.line_id ? t('tenantCopyLineLink') : t('tenantLineLinkRequiresLineId')}
                              disabled={!tenant.line_id}
                              onClick={() => onCopyLineLink(tenant)}
                            >
                              <Link2 />
                            </Button>
                            {tenant.line_user_id && (
                              <Button
                                type="button"
                                size="icon"
                                variant="outline"
                                title={t('tenantUnlinkLine')}
                                aria-label={t('tenantUnlinkLine')}
                                onClick={() => onUnlinkLine(tenant)}
                              >
                                <Unlink />
                              </Button>
                            )}
                            <Button
                              type="button"
                              size="icon"
                              variant="outline"
                              title={t('tenantEdit')}
                              aria-label={t('tenantEdit')}
                              onClick={() => onEditTenant(tenant)}
                            >
                              <Pencil />
                            </Button>
                            <Button
                              type="button"
                              size="icon"
                              variant="destructive"
                              title={t('tenantDelete')}
                              aria-label={t('tenantDelete')}
                              onClick={() => onDeleteTenant(tenant)}
                              disabled={deletingTenantId === tenant.id}
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
                      {TENANT_PAGE_SIZE_OPTIONS.map((size) => (
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
