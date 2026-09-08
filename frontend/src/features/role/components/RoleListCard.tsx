import {
  ArrowDown,
  ArrowUp,
  ArrowUpDown,
  ChevronLeft,
  ChevronRight,
  ChevronsLeft,
  ChevronsRight,
  KeyRound,
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
import type { ApiRole } from '../types'
import {
  ROLE_PAGE_SIZE_OPTIONS,
  isTextFilterKey,
  type RoleColumnFilters,
  type RoleSortDirection,
  type RoleSortKey,
  type RoleStatusFilter,
  type RoleTextFilterKey,
} from '../utils'

const TEXT_COLUMNS: { key: RoleSortKey; labelKey: TranslationKey }[] = [
  { key: 'name', labelKey: 'rolePermissionsNameColumn' },
  { key: 'description', labelKey: 'rolePermissionsDescriptionColumn' },
]

const STATUS_COLUMN: { key: RoleSortKey; labelKey: TranslationKey } = {
  key: 'is_active',
  labelKey: 'rolePermissionsStatusColumn',
}

const TOTAL_COLUMN_COUNT = TEXT_COLUMNS.length + 2 // + status + actions

type RoleListCardProps = {
  isLoading: boolean
  loadError: string | null
  deleteError: string | null
  roleQuery: string
  onRoleQueryChange: (query: string) => void
  statusFilter: RoleStatusFilter
  onStatusFilterChange: (value: RoleStatusFilter) => void
  columnFilters: RoleColumnFilters
  onColumnFilterChange: (key: RoleTextFilterKey, value: string) => void
  hasFilters: boolean
  sortKey: RoleSortKey
  sortDirection: RoleSortDirection
  onSort: (key: RoleSortKey) => void
  roles: ApiRole[]
  total: number
  currentRolePage: number
  totalRolePages: number
  rolesRangeStart: number
  rolesRangeEnd: number
  rolePageSize: number
  onRolePageSizeChange: (size: number) => void
  onFirstPage: () => void
  onPrevPage: () => void
  onNextPage: () => void
  onLastPage: () => void
  deletingRoleId: string | null
  onCreateRole: () => void
  onManageRole: (role: ApiRole) => void
  onEditRole: (role: ApiRole) => void
  onDeleteRole: (role: ApiRole) => void
}

export function RoleListCard({
  isLoading,
  loadError,
  deleteError,
  roleQuery,
  onRoleQueryChange,
  statusFilter,
  onStatusFilterChange,
  columnFilters,
  onColumnFilterChange,
  hasFilters,
  sortKey,
  sortDirection,
  onSort,
  roles,
  total,
  currentRolePage,
  totalRolePages,
  rolesRangeStart,
  rolesRangeEnd,
  rolePageSize,
  onRolePageSizeChange,
  onFirstPage,
  onPrevPage,
  onNextPage,
  onLastPage,
  deletingRoleId,
  onCreateRole,
  onManageRole,
  onEditRole,
  onDeleteRole,
}: RoleListCardProps) {
  const { t } = useLanguage()

  // The column filters live in a table that scrolls, so on a narrow screen
  // they're off to the right and there's no way to tell what's applied.
  // Summarising them here keeps that visible and clearable at any size.
  const activeFilters: { id: string; label: string; onClear: () => void }[] = []

  if (statusFilter !== 'all') {
    activeFilters.push({
      id: 'is_active',
      label: `${t('rolePermissionsStatusColumn')}: ${statusFilter === 'active' ? t('statusActive') : t('statusInactive')}`,
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

  function renderSortableHead(column: { key: RoleSortKey; labelKey: TranslationKey }) {
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
              onTextChange={(value) => onColumnFilterChange(column.key as RoleTextFilterKey, value)}
            />
          )}

          {column.key === 'is_active' && (
            <ColumnFilterMenu
              label={t('rolePermissionsStatusColumn')}
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
          <CardTitle>{t('menuRoles')}</CardTitle>
        </div>
        <Button onClick={onCreateRole} disabled={isLoading}>
          {t('rolePermissionsCreateRole')}
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
                  {t('rolePermissionsRoleSearchLabel')}
                  <input
                    type="search"
                    className="h-10 rounded-md border border-input bg-transparent px-3 text-sm"
                    placeholder={t('rolePermissionsRoleSearchPlaceholder')}
                    value={roleQuery}
                    onChange={(event) => onRoleQueryChange(event.target.value)}
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
              <div className="role-table-wrap overflow-x-auto">
                <Table>
                  <TableHeader>
                    <TableRow>
                      {TEXT_COLUMNS.map((column) => renderSortableHead(column))}
                      {renderSortableHead(STATUS_COLUMN)}
                      <TableHead className="text-right">{t('rolePermissionsActionsColumn')}</TableHead>
                    </TableRow>
                  </TableHeader>
                  <TableBody>
                    {roles.length === 0 && (
                      <TableRow>
                        {/* The table scrolls, so centring this across every
                            column would push it off a phone screen. Pin it to
                            the left edge instead. */}
                        <TableCell colSpan={TOTAL_COLUMN_COUNT} className="p-0">
                          <p className="metric-detail sticky left-0 px-3 py-6">
                            {hasFilters ? t('rolePermissionsNoMatchingRoles') : t('rolePermissionsNoRoles')}
                          </p>
                        </TableCell>
                      </TableRow>
                    )}
                    {roles.map((role) => (
                      <TableRow key={role.id}>
                        <TableCell className="font-semibold">
                          <div className="flex items-center gap-2">
                            {role.name}
                            {role.is_protected && (
                              <Badge variant="outline">{t('rolePermissionsProtected')}</Badge>
                            )}
                          </div>
                        </TableCell>
                        <TableCell className="text-muted-foreground">
                          {role.description || t('rolePermissionsDescriptionEmpty')}
                        </TableCell>
                        <TableCell>
                          <Badge variant={role.is_active ? 'default' : 'outline'}>
                            {role.is_active ? t('statusActive') : t('statusInactive')}
                          </Badge>
                        </TableCell>
                        <TableCell className="text-right">
                          <div className="flex flex-wrap justify-end gap-2">
                            <Button
                              type="button"
                              size="icon"
                              variant="outline"
                              title={t('rolePermissionsManage')}
                              aria-label={t('rolePermissionsManage')}
                              onClick={() => onManageRole(role)}
                            >
                              <KeyRound />
                            </Button>
                            <Button
                              type="button"
                              size="icon"
                              variant="outline"
                              title={role.is_protected ? t('rolePermissionsProtectedHint') : t('rolePermissionsEditRole')}
                              aria-label={t('rolePermissionsEditRole')}
                              onClick={() => onEditRole(role)}
                              disabled={role.is_protected}
                            >
                              <Pencil />
                            </Button>
                            <Button
                              type="button"
                              size="icon"
                              variant="destructive"
                              title={role.is_protected ? t('rolePermissionsProtectedHint') : t('rolePermissionsDeleteRole')}
                              aria-label={t('rolePermissionsDeleteRole')}
                              onClick={() => onDeleteRole(role)}
                              disabled={deletingRoleId === role.id || role.is_protected}
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
                  {t('rolePermissionsShowingLabel')} {rolesRangeStart}-{rolesRangeEnd}{' '}
                  {t('rolePermissionsOfLabel')} {total} {t('rolePermissionsResultsLabel')}
                  {totalRolePages > 1 && (
                    <>
                      {' '}
                      · {t('rolePermissionsPageLabel')} {currentRolePage} / {totalRolePages}
                    </>
                  )}
                </p>
                <div className="flex flex-wrap items-center gap-3">
                  <label className="flex items-center gap-1.5 text-sm text-muted-foreground">
                    {t('rolePermissionsPageSizeLabel')}
                    <select
                      className="h-9 rounded-md border border-input bg-transparent px-2 text-sm"
                      value={rolePageSize}
                      onChange={(event) => onRolePageSizeChange(Number(event.target.value))}
                    >
                      {ROLE_PAGE_SIZE_OPTIONS.map((size) => (
                        <option key={size} value={size}>
                          {size}
                        </option>
                      ))}
                    </select>
                  </label>

                  {totalRolePages > 1 && (
                    <div className="flex gap-2">
                      <Button
                        type="button"
                        size="icon"
                        variant="outline"
                        title={t('rolePermissionsFirstPage')}
                        aria-label={t('rolePermissionsFirstPage')}
                        disabled={currentRolePage <= 1}
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
                        disabled={currentRolePage <= 1}
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
                        disabled={currentRolePage >= totalRolePages}
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
                        disabled={currentRolePage >= totalRolePages}
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
