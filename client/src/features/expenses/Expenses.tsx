import { useState, useMemo, useEffect, useCallback } from 'react'
import { useTranslation } from 'react-i18next'
import {
  Search,
  Plus,
  MoreHorizontal,
  Pencil,
  Trash2,
  SearchX,
  RefreshCw,
  ChevronLeft,
  ChevronRight,
  ChevronsLeft,
  ChevronsRight,
} from 'lucide-react'
import { Button } from '@/components/ui/button'
import { Input } from '@/components/ui/input'
import { Card, CardContent, CardHeader, CardTitle, CardDescription } from '@/components/ui/card'
import { Label } from '@/components/ui/label'
import { Badge } from '@/components/ui/badge'
import {
  DropdownMenu,
  DropdownMenuContent,
  DropdownMenuItem,
  DropdownMenuSeparator,
  DropdownMenuTrigger,
} from '@/components/ui/dropdown-menu'
import {
  Dialog,
  DialogContent,
  DialogHeader,
  DialogTitle,
  DialogDescription,
  DialogFooter,
} from '@/components/ui/dialog'
import {
  Select,
  SelectContent,
  SelectItem,
  SelectTrigger,
  SelectValue,
} from '@/components/ui/select'
import { expenseService } from '@/features/expenses/expenseService'
import type { ApiExpense, ExpenseCategory } from '@/types/api'
import { cn } from '@/lib/utils'
import { formatDate } from '@/lib/dateUtils'

const PER_PAGE_OPTIONS = [10, 20, 50] as const

const CATEGORY_VARIANTS: Record<ExpenseCategory, 'default' | 'secondary' | 'destructive' | 'outline'> = {
  repair: 'destructive',
  utilities: 'default',
  supplies: 'secondary',
  salary: 'outline',
  other: 'secondary',
}

function toDateInput(iso: string | null | undefined): string {
  if (!iso) return ''
  return iso.slice(0, 10)
}

const emptyForm = {
  expense_date: '',
  category: 'other' as ExpenseCategory,
  description: '',
  amount: '',
  note: '',
}

export function Expenses() {
  const { t } = useTranslation()

  const [expenses, setExpenses] = useState<ApiExpense[]>([])
  const [loading, setLoading] = useState(true)
  const [error, setError] = useState('')
  const [search, setSearch] = useState('')
  const [filterCategory, setFilterCategory] = useState<string>('all')
  const [page, setPage] = useState(1)
  const [perPage, setPerPage] = useState<number>(20)
  const [total, setTotal] = useState(0)
  const [totalPages, setTotalPages] = useState(1)

  const [dialogOpen, setDialogOpen] = useState(false)
  const [deleteDialogOpen, setDeleteDialogOpen] = useState(false)
  const [editingExpense, setEditingExpense] = useState<ApiExpense | null>(null)
  const [deletingExpense, setDeletingExpense] = useState<ApiExpense | null>(null)
  const [form, setForm] = useState(emptyForm)
  const [saving, setSaving] = useState(false)

  const fetchExpenses = useCallback(async (p: number, pp: number) => {
    setLoading(true)
    setError('')
    try {
      const resp = await expenseService.list(p, pp)
      setExpenses(resp.data)
      setTotal(resp.meta.total)
      setTotalPages(resp.meta.total_pages)
    } catch {
      setError(t('expenses.loadError'))
    } finally {
      setLoading(false)
    }
  }, [t])

  useEffect(() => {
    fetchExpenses(page, perPage)
  }, [fetchExpenses, page, perPage])

  function handlePerPageChange(value: string) {
    setPerPage(Number(value))
    setPage(1)
  }

  const filtered = useMemo(() => {
    const q = search.toLowerCase()
    return expenses.filter((e) => {
      const matchSearch = e.description.toLowerCase().includes(q) || e.note.toLowerCase().includes(q)
      const matchCategory = filterCategory === 'all' || e.category === filterCategory
      return matchSearch && matchCategory
    })
  }, [expenses, search, filterCategory])

  function openCreate() {
    setEditingExpense(null)
    setForm({ ...emptyForm, expense_date: new Date().toISOString().slice(0, 10) })
    setDialogOpen(true)
  }

  function openEdit(expense: ApiExpense) {
    setEditingExpense(expense)
    setForm({
      expense_date: toDateInput(expense.expense_date),
      category: expense.category,
      description: expense.description,
      amount: String(expense.amount),
      note: expense.note,
    })
    setDialogOpen(true)
  }

  async function handleSave() {
    if (!form.expense_date || !form.description || !form.amount) return
    setSaving(true)
    try {
      const payload = {
        expense_date: form.expense_date,
        category: form.category,
        description: form.description,
        amount: Number(form.amount),
        note: form.note,
      }
      if (editingExpense) {
        await expenseService.update(editingExpense.id, payload)
      } else {
        await expenseService.create(payload)
      }
      setDialogOpen(false)
      await fetchExpenses(page, perPage)
    } catch {
      // handled silently
    } finally {
      setSaving(false)
    }
  }

  async function handleDelete() {
    if (!deletingExpense) return
    try {
      await expenseService.delete(deletingExpense.id)
      const newPage = expenses.length === 1 && page > 1 ? page - 1 : page
      setPage(newPage)
      await fetchExpenses(newPage, perPage)
    } catch {
      // ignore
    } finally {
      setDeleteDialogOpen(false)
      setDeletingExpense(null)
    }
  }

  const isSaveDisabled = saving || !form.expense_date || !form.description || !form.amount

  const totalAmount = useMemo(
    () => filtered.reduce((sum, e) => sum + e.amount, 0),
    [filtered]
  )

  return (
    <div className="space-y-6">
      <div className="flex flex-wrap items-start justify-between gap-3">
        <div className="min-w-0">
          <h1 className="text-2xl font-bold tracking-tight">{t('expenses.title')}</h1>
          <p className="text-muted-foreground text-sm mt-1">
            {t('expenses.subtitle', { count: filtered.length, total })}
          </p>
        </div>
        <div className="flex gap-2 shrink-0">
          <Button variant="outline" size="icon" onClick={() => fetchExpenses(page, perPage)} disabled={loading}>
            <RefreshCw className={cn('w-4 h-4', loading && 'animate-spin')} />
          </Button>
          <Button onClick={openCreate} className="gap-2">
            <Plus className="w-4 h-4" />
            <span className="hidden sm:inline">{t('expenses.addExpense')}</span>
          </Button>
        </div>
      </div>

      {error && (
        <div className="rounded-md border border-destructive/30 bg-destructive/10 px-4 py-3 text-sm text-destructive">
          {error}
        </div>
      )}

      <Card>
        <CardHeader>
          <div className="flex flex-col sm:flex-row sm:items-center justify-between gap-3">
            <div>
              <CardTitle>{t('expenses.expenseList')}</CardTitle>
              <CardDescription>{t('expenses.expenseListDesc')}</CardDescription>
            </div>
            <div className="flex gap-2">
              <Select value={filterCategory} onValueChange={setFilterCategory}>
                <SelectTrigger className="h-9 w-40">
                  <SelectValue />
                </SelectTrigger>
                <SelectContent>
                  <SelectItem value="all">{t('expenses.allCategories')}</SelectItem>
                  <SelectItem value="repair">{t('expenses.categories.repair')}</SelectItem>
                  <SelectItem value="utilities">{t('expenses.categories.utilities')}</SelectItem>
                  <SelectItem value="supplies">{t('expenses.categories.supplies')}</SelectItem>
                  <SelectItem value="salary">{t('expenses.categories.salary')}</SelectItem>
                  <SelectItem value="other">{t('expenses.categories.other')}</SelectItem>
                </SelectContent>
              </Select>
              <div className="relative">
                <Search className="absolute left-3 top-1/2 -translate-y-1/2 w-4 h-4 text-muted-foreground" />
                <Input
                  placeholder={t('expenses.searchPlaceholder')}
                  className="pl-9 w-full sm:w-52"
                  value={search}
                  onChange={(e) => setSearch(e.target.value)}
                />
              </div>
            </div>
          </div>
        </CardHeader>

        <CardContent className="p-0">
          {loading ? (
            <div className="flex items-center justify-center py-16 text-sm text-muted-foreground">
              {t('common.loading')}
            </div>
          ) : (
            <>
              {/* Desktop table */}
              <div className="hidden md:block overflow-x-auto">
                <table className="w-full text-sm">
                  <thead>
                    <tr className="border-b bg-muted/40">
                      <th className="text-left px-6 py-3 font-medium text-muted-foreground">{t('expenses.colDate')}</th>
                      <th className="text-left px-4 py-3 font-medium text-muted-foreground">{t('expenses.colCategory')}</th>
                      <th className="text-left px-4 py-3 font-medium text-muted-foreground">{t('expenses.colDescription')}</th>
                      <th className="text-right px-4 py-3 font-medium text-muted-foreground">{t('expenses.colAmount')}</th>
                      <th className="text-right px-6 py-3 font-medium text-muted-foreground">{t('expenses.colActions')}</th>
                    </tr>
                  </thead>
                  <tbody>
                    {filtered.map((expense, i) => (
                      <tr
                        key={expense.id}
                        className={cn(
                          'border-b transition-colors hover:bg-muted/30',
                          i === filtered.length - 1 && 'border-0'
                        )}
                      >
                        <td className="px-6 py-4 text-muted-foreground">{formatDate(expense.expense_date)}</td>
                        <td className="px-4 py-4">
                          <Badge variant={CATEGORY_VARIANTS[expense.category]}>
                            {t(`expenses.categories.${expense.category}`)}
                          </Badge>
                        </td>
                        <td className="px-4 py-4 font-medium">
                          {expense.description}
                          {expense.note && (
                            <p className="text-xs text-muted-foreground mt-0.5 truncate max-w-xs">{expense.note}</p>
                          )}
                        </td>
                        <td className="px-4 py-4 text-right font-semibold">
                          {expense.amount.toLocaleString()}
                          <span className="text-xs text-muted-foreground ml-1">฿</span>
                        </td>
                        <td className="px-6 py-4 text-right">
                          <DropdownMenu>
                            <DropdownMenuTrigger asChild>
                              <Button variant="ghost" size="icon" className="h-8 w-8">
                                <MoreHorizontal className="w-4 h-4" />
                              </Button>
                            </DropdownMenuTrigger>
                            <DropdownMenuContent align="end">
                              <DropdownMenuItem onClick={() => openEdit(expense)} className="gap-2">
                                <Pencil className="w-4 h-4" /> {t('common.edit')}
                              </DropdownMenuItem>
                              <DropdownMenuSeparator />
                              <DropdownMenuItem
                                className="gap-2 text-destructive focus:text-destructive"
                                onClick={() => { setDeletingExpense(expense); setDeleteDialogOpen(true) }}
                              >
                                <Trash2 className="w-4 h-4" /> {t('common.delete')}
                              </DropdownMenuItem>
                            </DropdownMenuContent>
                          </DropdownMenu>
                        </td>
                      </tr>
                    ))}
                  </tbody>
                  {filtered.length > 0 && (
                    <tfoot>
                      <tr className="border-t bg-muted/40">
                        <td colSpan={3} className="px-6 py-3 text-sm font-medium text-muted-foreground">
                          {t('expenses.totalShown')}
                        </td>
                        <td className="px-4 py-3 text-right font-bold">
                          {totalAmount.toLocaleString()}
                          <span className="text-xs text-muted-foreground ml-1">฿</span>
                        </td>
                        <td />
                      </tr>
                    </tfoot>
                  )}
                </table>

                {filtered.length === 0 && (
                  <div className="flex flex-col items-center justify-center py-16 gap-3 text-muted-foreground">
                    <div className="p-4 rounded-full bg-muted">
                      <SearchX className="w-6 h-6" />
                    </div>
                    <p className="text-sm font-medium">{t('expenses.noExpenses')}</p>
                    {(search || filterCategory !== 'all') && (
                      <button
                        type="button"
                        onClick={() => { setSearch(''); setFilterCategory('all') }}
                        className="text-xs text-primary hover:underline"
                      >
                        {t('expenses.clearFilters')}
                      </button>
                    )}
                  </div>
                )}
              </div>

              {/* Mobile cards */}
              <div className="md:hidden divide-y">
                {filtered.map((expense) => (
                  <div key={expense.id} className="p-4 flex items-start gap-3">
                    <div className="flex-1 min-w-0 space-y-1">
                      <p className="text-sm font-medium">{expense.description}</p>
                      <p className="text-xs text-muted-foreground">{formatDate(expense.expense_date)}</p>
                      <p className="text-sm font-semibold">{expense.amount.toLocaleString()} ฿</p>
                      <Badge variant={CATEGORY_VARIANTS[expense.category]} className="text-xs">
                        {t(`expenses.categories.${expense.category}`)}
                      </Badge>
                    </div>
                    <DropdownMenu>
                      <DropdownMenuTrigger asChild>
                        <Button variant="ghost" size="icon" className="h-8 w-8 shrink-0">
                          <MoreHorizontal className="w-4 h-4" />
                        </Button>
                      </DropdownMenuTrigger>
                      <DropdownMenuContent align="end">
                        <DropdownMenuItem onClick={() => openEdit(expense)}>{t('common.edit')}</DropdownMenuItem>
                        <DropdownMenuSeparator />
                        <DropdownMenuItem
                          className="text-destructive focus:text-destructive"
                          onClick={() => { setDeletingExpense(expense); setDeleteDialogOpen(true) }}
                        >
                          {t('common.delete')}
                        </DropdownMenuItem>
                      </DropdownMenuContent>
                    </DropdownMenu>
                  </div>
                ))}
              </div>

              {/* Pagination */}
              {!loading && total > 0 && (
                <div className="flex flex-col sm:flex-row items-center justify-between gap-3 px-4 py-3 border-t text-sm text-muted-foreground">
                  <span className="shrink-0">
                    {t('expenses.showing', {
                      from: (page - 1) * perPage + 1,
                      to: Math.min(page * perPage, total),
                      total,
                    })}
                  </span>
                  <div className="flex items-center gap-4">
                    <div className="flex items-center gap-2">
                      <span className="hidden sm:inline shrink-0">{t('expenses.perPage')}</span>
                      <Select value={String(perPage)} onValueChange={handlePerPageChange}>
                        <SelectTrigger className="h-8 w-16"><SelectValue /></SelectTrigger>
                        <SelectContent>
                          {PER_PAGE_OPTIONS.map((n) => (
                            <SelectItem key={n} value={String(n)}>{n}</SelectItem>
                          ))}
                        </SelectContent>
                      </Select>
                    </div>
                    <span className="shrink-0">{t('expenses.page')} {page} {t('expenses.of')} {totalPages}</span>
                    <div className="flex items-center gap-1">
                      <Button variant="outline" size="icon" className="h-8 w-8" onClick={() => setPage(1)} disabled={page === 1}>
                        <ChevronsLeft className="w-4 h-4" />
                      </Button>
                      <Button variant="outline" size="icon" className="h-8 w-8" onClick={() => setPage((p) => p - 1)} disabled={page === 1}>
                        <ChevronLeft className="w-4 h-4" />
                      </Button>
                      <Button variant="outline" size="icon" className="h-8 w-8" onClick={() => setPage((p) => p + 1)} disabled={page >= totalPages}>
                        <ChevronRight className="w-4 h-4" />
                      </Button>
                      <Button variant="outline" size="icon" className="h-8 w-8" onClick={() => setPage(totalPages)} disabled={page >= totalPages}>
                        <ChevronsRight className="w-4 h-4" />
                      </Button>
                    </div>
                  </div>
                </div>
              )}
            </>
          )}
        </CardContent>
      </Card>

      {/* Create/Edit Dialog */}
      <Dialog open={dialogOpen} onOpenChange={setDialogOpen}>
        <DialogContent className="max-w-lg">
          <DialogHeader>
            <DialogTitle>
              {editingExpense ? t('expenses.editExpense') : t('expenses.createExpense')}
            </DialogTitle>
            <DialogDescription>
              {editingExpense ? t('expenses.editDesc') : t('expenses.createDesc')}
            </DialogDescription>
          </DialogHeader>

          <div className="grid gap-4 py-4">
            <div className="grid grid-cols-2 gap-4">
              <div className="space-y-1.5">
                <Label>{t('expenses.expenseDate')} *</Label>
                <Input
                  type="date"
                  value={form.expense_date}
                  onChange={(e) => setForm((f) => ({ ...f, expense_date: e.target.value }))}
                />
              </div>
              <div className="space-y-1.5">
                <Label>{t('expenses.category')} *</Label>
                <Select
                  value={form.category}
                  onValueChange={(v) => setForm((f) => ({ ...f, category: v as ExpenseCategory }))}
                >
                  <SelectTrigger><SelectValue /></SelectTrigger>
                  <SelectContent>
                    <SelectItem value="repair">{t('expenses.categories.repair')}</SelectItem>
                    <SelectItem value="utilities">{t('expenses.categories.utilities')}</SelectItem>
                    <SelectItem value="supplies">{t('expenses.categories.supplies')}</SelectItem>
                    <SelectItem value="salary">{t('expenses.categories.salary')}</SelectItem>
                    <SelectItem value="other">{t('expenses.categories.other')}</SelectItem>
                  </SelectContent>
                </Select>
              </div>
            </div>

            <div className="space-y-1.5">
              <Label>{t('expenses.description')} *</Label>
              <Input
                placeholder={t('expenses.descriptionPlaceholder')}
                value={form.description}
                onChange={(e) => setForm((f) => ({ ...f, description: e.target.value }))}
              />
            </div>

            <div className="space-y-1.5">
              <Label>{t('expenses.amount')} *</Label>
              <Input
                type="number"
                min="0"
                step="0.01"
                placeholder="0.00"
                value={form.amount}
                onChange={(e) => setForm((f) => ({ ...f, amount: e.target.value }))}
              />
            </div>

            <div className="space-y-1.5">
              <Label>{t('expenses.note')}</Label>
              <textarea
                rows={2}
                placeholder={t('expenses.notePlaceholder')}
                value={form.note}
                onChange={(e: React.ChangeEvent<HTMLTextAreaElement>) =>
                  setForm((f) => ({ ...f, note: e.target.value }))
                }
                className="flex min-h-16 w-full rounded-md border border-input bg-background px-3 py-2 text-sm ring-offset-background placeholder:text-muted-foreground focus-visible:outline-none focus-visible:ring-2 focus-visible:ring-ring focus-visible:ring-offset-2 resize-none"
              />
            </div>
          </div>

          <DialogFooter>
            <Button variant="outline" onClick={() => setDialogOpen(false)}>
              {t('common.cancel')}
            </Button>
            <Button onClick={handleSave} disabled={isSaveDisabled}>
              {saving
                ? t('common.loading')
                : editingExpense
                  ? t('expenses.saveChanges')
                  : t('expenses.createExpense')}
            </Button>
          </DialogFooter>
        </DialogContent>
      </Dialog>

      {/* Delete confirm */}
      <Dialog open={deleteDialogOpen} onOpenChange={setDeleteDialogOpen}>
        <DialogContent className="max-w-sm">
          <DialogHeader>
            <DialogTitle>{t('expenses.deleteExpense')}</DialogTitle>
            <DialogDescription>
              {t('expenses.deleteConfirm')}{' '}
              <span className="font-semibold text-foreground">
                {deletingExpense?.description}
              </span>?{' '}
              {t('expenses.deleteWarning')}
            </DialogDescription>
          </DialogHeader>
          <DialogFooter className="gap-2">
            <Button variant="outline" onClick={() => setDeleteDialogOpen(false)}>
              {t('common.cancel')}
            </Button>
            <Button variant="destructive" onClick={handleDelete}>
              {t('common.delete')}
            </Button>
          </DialogFooter>
        </DialogContent>
      </Dialog>
    </div>
  )
}
