import { useEffect, useState, type FormEvent } from 'react'
import axios from 'axios'
import { api, extractErrorMessage, type ApiPage } from '@/shared/api/client'
import { useLanguage } from '@/shared/i18n/language'
import { ConfirmDialog } from '@/shared/components/confirm-dialog'
import type { ApiDormitory } from '@/features/dormitory/types'
import { ExpenseListCard } from './components/ExpenseListCard'
import { ExpenseFormSheet } from './components/ExpenseFormSheet'
import type { ApiExpense, ExpenseCategory } from './types'
import {
  EXPENSE_PAGE_SIZE_OPTIONS,
  toApiDate,
  toDateInputValue,
  type ExpenseCategoryFilter,
  type ExpenseColumnFilters,
  type ExpenseSortDirection,
  type ExpenseSortKey,
  type ExpenseTextFilterKey,
} from './utils'

const SEARCH_DEBOUNCE_MS = 300

export default function ExpensePage() {
  const { t } = useLanguage()

  const [expenses, setExpenses] = useState<ApiExpense[] | null>(null)
  const [total, setTotal] = useState(0)
  const [totalPages, setTotalPages] = useState(1)
  const [dormitories, setDormitories] = useState<ApiDormitory[]>([])
  const [loadError, setLoadError] = useState<string | null>(null)

  const [query, setQuery] = useState('')
  const [debouncedQuery, setDebouncedQuery] = useState('')
  const [categoryFilter, setCategoryFilter] = useState<ExpenseCategoryFilter>('all')
  // Applied on submit from each column's menu, so no debounce is needed here.
  const [columnFilters, setColumnFilters] = useState<ExpenseColumnFilters>({})
  const [sortKey, setSortKey] = useState<ExpenseSortKey>('expense_date')
  const [sortDirection, setSortDirection] = useState<ExpenseSortDirection>('desc')

  const [page, setPage] = useState(1)
  const [pageSize, setPageSizeState] = useState<number>(EXPENSE_PAGE_SIZE_OPTIONS[0])
  // Bumped by mutations so the list refetches; the server owns the ordering
  // and page boundaries now, so patching rows locally would misplace them.
  const [refreshToken, setRefreshToken] = useState(0)

  const [formOpen, setFormOpen] = useState(false)
  const [formExpenseId, setFormExpenseId] = useState<string | null>(null)
  const [formDormitoryId, setFormDormitoryId] = useState('')
  const [formCategory, setFormCategory] = useState<ExpenseCategory>('other')
  const [formExpenseDate, setFormExpenseDate] = useState('')
  const [formAmount, setFormAmount] = useState('')
  const [formDescription, setFormDescription] = useState('')
  const [formSaving, setFormSaving] = useState(false)
  const [formError, setFormError] = useState<string | null>(null)

  const [deletingExpenseId, setDeletingExpenseId] = useState<string | null>(null)
  const [deleteError, setDeleteError] = useState<string | null>(null)
  const [confirmDeleteExpense, setConfirmDeleteExpense] = useState<ApiExpense | null>(null)

  useEffect(() => {
    const timer = window.setTimeout(() => setDebouncedQuery(query), SEARCH_DEBOUNCE_MS)
    return () => window.clearTimeout(timer)
  }, [query])

  useEffect(() => {
    let cancelled = false

    api
      .get<ApiDormitory[]>('/dormitories/active', { params: { limit: 100 } })
      .then(({ data }) => {
        if (cancelled) return
        setDormitories(data)
      })
      .catch(() => {
        // Ignore — the create form just shows no dormitory choices.
      })

    return () => {
      cancelled = true
    }
  }, [])

  useEffect(() => {
    const controller = new AbortController()

    const columnParams: Record<string, string> = {}
    for (const [key, value] of Object.entries(columnFilters)) {
      if (value) columnParams[`f[${key}]`] = value
    }

    api
      .get<ApiPage<ApiExpense[]>>('/expenses', {
        signal: controller.signal,
        params: {
          page,
          per_page: pageSize,
          q: debouncedQuery.trim() || undefined,
          category: categoryFilter === 'all' ? undefined : categoryFilter,
          sort: sortKey,
          order: sortDirection,
          ...columnParams,
        },
      })
      .then(({ data }) => {
        setExpenses(data.data)
        setTotal(data.meta.total)
        setTotalPages(data.meta.total_pages)
        setLoadError(null)
      })
      .catch((err) => {
        // A cancelled request is this effect being superseded by a newer one,
        // not a failure — letting it through would overwrite the newer result.
        if (axios.isCancel(err)) return
        setLoadError(extractErrorMessage(err, t('resourceLoadError')))
      })

    return () => controller.abort()
    // eslint-disable-next-line react-hooks/exhaustive-deps
  }, [page, pageSize, debouncedQuery, categoryFilter, columnFilters, sortKey, sortDirection, refreshToken])

  const isLoading = !loadError && expenses === null
  const hasFilters = query !== '' || categoryFilter !== 'all' || Object.values(columnFilters).some(Boolean)
  const rangeStart = total === 0 ? 0 : (page - 1) * pageSize + 1
  const rangeEnd = Math.min(page * pageSize, total)

  function refresh() {
    setRefreshToken((value) => value + 1)
  }

  function setPageSize(size: number) {
    setPageSizeState(size)
    setPage(1)
  }

  function handleSort(key: ExpenseSortKey) {
    if (key === sortKey) {
      setSortDirection(sortDirection === 'asc' ? 'desc' : 'asc')
    } else {
      setSortKey(key)
      setSortDirection('asc')
    }
    setPage(1)
  }

  function openCreateForm() {
    setFormExpenseId(null)
    setFormDormitoryId('')
    setFormCategory('other')
    setFormExpenseDate(toDateInputValue(new Date().toISOString()))
    setFormAmount('')
    setFormDescription('')
    setFormError(null)
    setFormOpen(true)
  }

  function openEditForm(expense: ApiExpense) {
    setFormExpenseId(expense.id)
    setFormDormitoryId(expense.dormitory_id)
    setFormCategory(expense.category)
    setFormExpenseDate(toDateInputValue(expense.expense_date))
    setFormAmount(String(expense.amount))
    setFormDescription(expense.description)
    setFormError(null)
    setFormOpen(true)
  }

  async function handleFormSubmit(event: FormEvent<HTMLFormElement>) {
    event.preventDefault()

    if (formExpenseId === null && !formDormitoryId) {
      setFormError(t('expenseDormitoryRequired'))
      return
    }
    if (!formExpenseDate) {
      setFormError(t('expenseDateRequired'))
      return
    }
    const amount = Number(formAmount)
    if (!formAmount || Number.isNaN(amount) || amount <= 0) {
      setFormError(t('expenseAmountInvalid'))
      return
    }

    setFormSaving(true)
    setFormError(null)

    try {
      if (formExpenseId === null) {
        const payload = {
          dormitory_id: formDormitoryId,
          category: formCategory,
          expense_date: toApiDate(formExpenseDate),
          amount,
          description: formDescription.trim(),
        }
        await api.post<ApiExpense>('/expenses', payload)
      } else {
        const payload = {
          category: formCategory,
          expense_date: toApiDate(formExpenseDate),
          amount,
          description: formDescription.trim(),
        }
        await api.put<ApiExpense>(`/expenses/${formExpenseId}`, payload)
      }
      refresh()
      setFormOpen(false)
    } catch (err) {
      const fallback = formExpenseId === null ? t('expenseCreateError') : t('expenseUpdateError')
      setFormError(extractErrorMessage(err, fallback))
    } finally {
      setFormSaving(false)
    }
  }

  async function handleDeleteExpense() {
    if (!confirmDeleteExpense) return
    const expense = confirmDeleteExpense

    setDeletingExpenseId(expense.id)
    setDeleteError(null)

    try {
      await api.delete(`/expenses/${expense.id}`)
      // Deleting the last row of the final page would otherwise strand the
      // view on a page the server no longer has.
      setPage((value) => (expenses?.length === 1 ? Math.max(1, value - 1) : value))
      refresh()
      setConfirmDeleteExpense(null)
    } catch (err) {
      setDeleteError(extractErrorMessage(err, t('expenseDeleteError')))
    } finally {
      setDeletingExpenseId(null)
    }
  }

  return (
    <main className="content">
      <section className="welcome">
        <h1>{t('menuExpenses')}</h1>
        <p>{t('menuExpensesDescription')}</p>
      </section>

      <ExpenseListCard
        isLoading={isLoading}
        loadError={loadError}
        deleteError={deleteError}
        query={query}
        onQueryChange={(value) => {
          setQuery(value)
          setPage(1)
        }}
        categoryFilter={categoryFilter}
        onCategoryFilterChange={(value) => {
          setCategoryFilter(value)
          setPage(1)
        }}
        columnFilters={columnFilters}
        onColumnFilterChange={(key: ExpenseTextFilterKey, value: string) => {
          setColumnFilters((prev) => ({ ...prev, [key]: value }))
          setPage(1)
        }}
        hasFilters={hasFilters}
        sortKey={sortKey}
        sortDirection={sortDirection}
        onSort={handleSort}
        expenses={expenses ?? []}
        total={total}
        currentPage={page}
        totalPages={totalPages}
        rangeStart={rangeStart}
        rangeEnd={rangeEnd}
        pageSize={pageSize}
        onPageSizeChange={setPageSize}
        onFirstPage={() => setPage(1)}
        onPrevPage={() => setPage((value) => Math.max(1, value - 1))}
        onNextPage={() => setPage((value) => Math.min(totalPages, value + 1))}
        onLastPage={() => setPage(totalPages)}
        deletingExpenseId={deletingExpenseId}
        onCreateExpense={openCreateForm}
        onEditExpense={openEditForm}
        onDeleteExpense={setConfirmDeleteExpense}
      />

      <ConfirmDialog
        open={confirmDeleteExpense !== null}
        onOpenChange={(open) => !open && setConfirmDeleteExpense(null)}
        title={t('confirmDeleteTitle')}
        description={t('expenseDeleteConfirm')}
        confirmLabel={t('expenseDelete')}
        cancelLabel={t('cancel')}
        loading={deletingExpenseId === confirmDeleteExpense?.id}
        error={deleteError}
        onConfirm={handleDeleteExpense}
      />

      <ExpenseFormSheet
        open={formOpen}
        onOpenChange={setFormOpen}
        isEdit={formExpenseId !== null}
        dormitoryId={formDormitoryId}
        onDormitoryIdChange={setFormDormitoryId}
        dormitories={dormitories}
        category={formCategory}
        onCategoryChange={setFormCategory}
        expenseDate={formExpenseDate}
        onExpenseDateChange={setFormExpenseDate}
        amount={formAmount}
        onAmountChange={setFormAmount}
        description={formDescription}
        onDescriptionChange={setFormDescription}
        saving={formSaving}
        error={formError}
        onSubmit={handleFormSubmit}
      />
    </main>
  )
}
