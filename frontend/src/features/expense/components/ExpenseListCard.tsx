import { ArrowDown, ArrowUp, ArrowUpDown, ChevronLeft, ChevronRight, ChevronsLeft, ChevronsRight, Pencil, Trash2, X } from 'lucide-react'
import { useLanguage, type TranslationKey } from '@/shared/i18n/language'
import { Badge } from '@/shared/components/ui/badge'
import { Card, CardContent, CardHeader, CardTitle } from '@/shared/components/ui/card'
import { ColumnFilterMenu } from '@/shared/components/column-filter-menu'
import { Table, TableBody, TableCell, TableHead, TableHeader, TableRow } from '@/shared/components/ui/table'
import { Button } from '@/shared/components/ui/button'
import type { ApiExpense, ExpenseCategory } from '../types'
import {
  EXPENSE_CATEGORIES,
  EXPENSE_PAGE_SIZE_OPTIONS,
  isExpenseTextFilterKey,
  toDateInputValue,
  type ExpenseCategoryFilter,
  type ExpenseColumnFilters,
  type ExpenseSortDirection,
  type ExpenseSortKey,
  type ExpenseTextFilterKey,
} from '../utils'

const expenseCategoryLabelKeys: Record<ExpenseCategory, TranslationKey> = {
  maintenance: 'expenseCategoryMaintenance',
  utility: 'expenseCategoryUtility',
  salary: 'expenseCategorySalary',
  supplies: 'expenseCategorySupplies',
  other: 'expenseCategoryOther',
}

const SORTABLE_COLUMNS: { key: ExpenseSortKey; labelKey: TranslationKey }[] = [
  { key: 'dormitory_name', labelKey: 'expenseDormitoryColumn' },
  { key: 'category', labelKey: 'expenseCategoryColumn' },
  { key: 'expense_date', labelKey: 'expenseDateColumn' },
  { key: 'amount', labelKey: 'expenseAmountColumn' },
  { key: 'description', labelKey: 'expenseDescriptionColumn' },
]

type ExpenseListCardProps = {
  isLoading: boolean
  loadError: string | null
  deleteError: string | null
  query: string
  onQueryChange: (query: string) => void
  categoryFilter: ExpenseCategoryFilter
  onCategoryFilterChange: (value: ExpenseCategoryFilter) => void
  columnFilters: ExpenseColumnFilters
  onColumnFilterChange: (key: ExpenseTextFilterKey, value: string) => void
  hasFilters: boolean
  sortKey: ExpenseSortKey
  sortDirection: ExpenseSortDirection
  onSort: (key: ExpenseSortKey) => void
  expenses: ApiExpense[]
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
  deletingExpenseId: string | null
  onCreateExpense: () => void
  onEditExpense: (expense: ApiExpense) => void
  onDeleteExpense: (expense: ApiExpense) => void
}

export function ExpenseListCard({
  isLoading,
  loadError,
  deleteError,
  query,
  onQueryChange,
  categoryFilter,
  onCategoryFilterChange,
  columnFilters,
  onColumnFilterChange,
  hasFilters,
  sortKey,
  sortDirection,
  onSort,
  expenses,
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
  deletingExpenseId,
  onCreateExpense,
  onEditExpense,
  onDeleteExpense,
}: ExpenseListCardProps) {
  const { t } = useLanguage()

  // The column filters live in a table that scrolls, so on a narrow screen
  // they're off to the right and there's no way to tell what's applied.
  // Summarising them here keeps that visible and clearable at any size.
  const activeFilters: { id: string; label: string; onClear: () => void }[] = []

  if (categoryFilter !== 'all') {
    activeFilters.push({
      id: 'category',
      label: `${t('expenseFilterCategoryLabel')}: ${t(expenseCategoryLabelKeys[categoryFilter])}`,
      onClear: () => onCategoryFilterChange('all'),
    })
  }

  for (const column of SORTABLE_COLUMNS) {
    const key = column.key
    if (!isExpenseTextFilterKey(key)) continue

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
          <CardTitle>{t('menuExpenses')}</CardTitle>
        </div>
        <Button onClick={onCreateExpense} disabled={isLoading}>
          {t('expenseCreate')}
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
                  {t('expenseSearchLabel')}
                  <input
                    type="search"
                    className="h-10 rounded-md border border-input bg-transparent px-3 text-sm"
                    placeholder={t('expenseSearchPlaceholder')}
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
              <div className="expense-table-wrap overflow-x-auto">
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
                                  label={t('expenseFilterCategoryLabel')}
                                  optionValue={categoryFilter}
                                  onOptionChange={onCategoryFilterChange}
                                  options={[
                                    { value: 'all', label: t('filterAll') },
                                    ...EXPENSE_CATEGORIES.map((category) => ({
                                      value: category,
                                      label: t(expenseCategoryLabelKeys[category]),
                                    })),
                                  ]}
                                />
                              )}

                              {isExpenseTextFilterKey(column.key) && (
                                <ColumnFilterMenu
                                  label={t(column.labelKey)}
                                  textValue={columnFilters[column.key] ?? ''}
                                  onTextChange={(value) =>
                                    onColumnFilterChange(column.key as ExpenseTextFilterKey, value)
                                  }
                                />
                              )}
                            </div>
                          </TableHead>
                        )
                      })}
                      <TableHead className="text-right">{t('expenseActionsColumn')}</TableHead>
                    </TableRow>
                  </TableHeader>
                  <TableBody>
                    {expenses.length === 0 && (
                      <TableRow>
                        {/* The table scrolls horizontally, so centring this
                            across every column would push it off a phone
                            screen. Pin it to the left edge instead. */}
                        <TableCell colSpan={SORTABLE_COLUMNS.length + 1} className="p-0">
                          <p className="metric-detail sticky left-0 px-3 py-6">
                            {hasFilters ? t('expenseNoMatching') : t('expenseNoExpenses')}
                          </p>
                        </TableCell>
                      </TableRow>
                    )}
                    {expenses.map((expense) => (
                      <TableRow key={expense.id}>
                        <TableCell className="font-semibold">{expense.dormitory_name || '—'}</TableCell>
                        <TableCell>
                          <Badge variant="secondary">{t(expenseCategoryLabelKeys[expense.category])}</Badge>
                        </TableCell>
                        <TableCell className="text-muted-foreground">
                          {toDateInputValue(expense.expense_date)}
                        </TableCell>
                        <TableCell className="text-muted-foreground">{expense.amount.toLocaleString()}</TableCell>
                        <TableCell className="text-muted-foreground">{expense.description || '—'}</TableCell>
                        <TableCell className="text-right">
                          <div className="flex flex-nowrap justify-end gap-2">
                            <Button
                              type="button"
                              size="icon"
                              variant="outline"
                              title={t('expenseEdit')}
                              aria-label={t('expenseEdit')}
                              onClick={() => onEditExpense(expense)}
                            >
                              <Pencil />
                            </Button>
                            <Button
                              type="button"
                              size="icon"
                              variant="destructive"
                              title={t('expenseDelete')}
                              aria-label={t('expenseDelete')}
                              onClick={() => onDeleteExpense(expense)}
                              disabled={deletingExpenseId === expense.id}
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
                      {EXPENSE_PAGE_SIZE_OPTIONS.map((size) => (
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
