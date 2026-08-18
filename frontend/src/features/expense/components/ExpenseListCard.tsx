import { ChevronLeft, ChevronRight, ChevronsLeft, ChevronsRight, Pencil, Trash2 } from 'lucide-react'
import { useLanguage, type TranslationKey } from '@/shared/i18n/language'
import { Badge } from '@/shared/components/ui/badge'
import { Card, CardContent, CardHeader, CardTitle } from '@/shared/components/ui/card'
import { Table, TableBody, TableCell, TableHead, TableHeader, TableRow } from '@/shared/components/ui/table'
import { Button } from '@/shared/components/ui/button'
import type { ApiExpense, ExpenseCategory } from '../types'
import { EXPENSE_PAGE_SIZE_OPTIONS, toDateInputValue } from '../utils'

const expenseCategoryLabelKeys: Record<ExpenseCategory, TranslationKey> = {
  maintenance: 'expenseCategoryMaintenance',
  utility: 'expenseCategoryUtility',
  salary: 'expenseCategorySalary',
  supplies: 'expenseCategorySupplies',
  other: 'expenseCategoryOther',
}

type ExpenseListCardProps = {
  isLoading: boolean
  loadError: string | null
  deleteError: string | null
  expenses: ApiExpense[] | null
  query: string
  onQueryChange: (query: string) => void
  filteredExpenses: ApiExpense[]
  paginatedExpenses: ApiExpense[]
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
  expenses,
  query,
  onQueryChange,
  filteredExpenses,
  paginatedExpenses,
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

  return (
    <Card>
      <CardHeader className="flex flex-row items-start justify-between gap-4">
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

        {!loadError && !isLoading && expenses && expenses.length === 0 && (
          <p className="metric-detail">{t('expenseNoExpenses')}</p>
        )}

        {!loadError && !isLoading && expenses && expenses.length > 0 && (
          <>
            <label className="flex w-full max-w-md flex-col gap-1.5 text-sm font-medium">
              {t('expenseSearchLabel')}
              <input
                type="search"
                className="h-10 rounded-md border border-input bg-transparent px-3 text-sm"
                placeholder={t('expenseSearchPlaceholder')}
                value={query}
                onChange={(event) => onQueryChange(event.target.value)}
              />
            </label>

            {filteredExpenses.length === 0 && <p className="metric-detail">{t('expenseNoMatching')}</p>}

            {filteredExpenses.length > 0 && (
              <div className="table-wrap">
                <Table>
                  <TableHeader>
                    <TableRow>
                      <TableHead>{t('expenseDormitoryColumn')}</TableHead>
                      <TableHead>{t('expenseCategoryColumn')}</TableHead>
                      <TableHead>{t('expenseDateColumn')}</TableHead>
                      <TableHead>{t('expenseAmountColumn')}</TableHead>
                      <TableHead>{t('expenseDescriptionColumn')}</TableHead>
                      <TableHead className="text-right">{t('expenseActionsColumn')}</TableHead>
                    </TableRow>
                  </TableHeader>
                  <TableBody>
                    {paginatedExpenses.map((expense) => (
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
                          <div className="flex flex-wrap justify-end gap-2">
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
            )}

            {filteredExpenses.length > 0 && (
              <div className="flex items-center justify-between gap-3">
                <p className="text-sm text-muted-foreground">
                  {t('rolePermissionsShowingLabel')} {rangeStart}-{rangeEnd}{' '}
                  {t('rolePermissionsOfLabel')} {filteredExpenses.length} {t('rolePermissionsResultsLabel')}
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
