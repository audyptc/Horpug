import { useEffect, useMemo, useState, type FormEvent } from 'react'
import { api, extractErrorMessage, type ApiPage } from '@/shared/api/client'
import { usePagination } from '@/shared/hooks/use-pagination'
import { useLanguage } from '@/shared/i18n/language'
import { ConfirmDialog } from '@/shared/components/confirm-dialog'
import type { ApiDormitory } from '@/features/dormitory/types'
import { ExpenseListCard } from './components/ExpenseListCard'
import { ExpenseFormSheet } from './components/ExpenseFormSheet'
import type { ApiExpense, ExpenseCategory } from './types'
import { EXPENSE_PAGE_SIZE_OPTIONS, toApiDate, toDateInputValue } from './utils'

export default function ExpensePage() {
  const { t } = useLanguage()

  const [expenses, setExpenses] = useState<ApiExpense[] | null>(null)
  const [dormitories, setDormitories] = useState<ApiDormitory[]>([])
  const [loadError, setLoadError] = useState<string | null>(null)

  const [query, setQuery] = useState('')

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
    let cancelled = false

    Promise.all([
      api.get<ApiPage<ApiExpense[]>>('/expenses', { params: { per_page: 100 } }),
      api.get<ApiDormitory[]>('/dormitories/active', { params: { limit: 100 } }),
    ])
      .then(([expensesRes, dormitoriesRes]) => {
        if (cancelled) return
        setExpenses(expensesRes.data.data)
        setDormitories(dormitoriesRes.data)
      })
      .catch((err) => {
        if (!cancelled) setLoadError(extractErrorMessage(err, t('resourceLoadError')))
      })

    return () => {
      cancelled = true
    }
    // eslint-disable-next-line react-hooks/exhaustive-deps
  }, [])

  const filteredExpenses = useMemo(() => {
    const q = query.trim().toLocaleLowerCase()
    if (!q) return expenses ?? []

    return (expenses ?? []).filter((expense) => {
      return (
        (expense.dormitory_name ?? '').toLocaleLowerCase().includes(q) ||
        expense.description.toLocaleLowerCase().includes(q)
      )
    })
  }, [query, expenses])

  const {
    page: currentPage,
    pageSize,
    setPageSize,
    totalPages,
    rangeStart,
    rangeEnd,
    paginatedItems: paginatedExpenses,
    resetPage,
    prevPage,
    nextPage,
  } = usePagination(filteredExpenses, EXPENSE_PAGE_SIZE_OPTIONS[0])

  const isLoading = !loadError && expenses === null

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
        const { data } = await api.post<ApiExpense>('/expenses', payload)
        setExpenses((prev) => [data, ...(prev ?? [])])
      } else {
        const payload = {
          category: formCategory,
          expense_date: toApiDate(formExpenseDate),
          amount,
          description: formDescription.trim(),
        }
        const { data } = await api.put<ApiExpense>(`/expenses/${formExpenseId}`, payload)
        setExpenses((prev) => prev?.map((item) => (item.id === data.id ? data : item)) ?? prev)
      }
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
      setExpenses((prev) => prev?.filter((item) => item.id !== expense.id) ?? prev)
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
        expenses={expenses}
        query={query}
        onQueryChange={(value) => {
          setQuery(value)
          resetPage()
        }}
        filteredExpenses={filteredExpenses}
        paginatedExpenses={paginatedExpenses}
        currentPage={currentPage}
        totalPages={totalPages}
        rangeStart={rangeStart}
        rangeEnd={rangeEnd}
        pageSize={pageSize}
        onPageSizeChange={setPageSize}
        onPrevPage={prevPage}
        onNextPage={nextPage}
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
