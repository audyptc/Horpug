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
  Banknote,
  Smartphone,
  CreditCard,
} from 'lucide-react'
import { Button } from '@/components/ui/button'
import { Input } from '@/components/ui/input'
import { Card, CardContent, CardHeader, CardTitle, CardDescription } from '@/components/ui/card'
import { Badge } from '@/components/ui/badge'
import {
  DropdownMenu,
  DropdownMenuContent,
  DropdownMenuItem,
  DropdownMenuSeparator,
  DropdownMenuTrigger,
} from '@/components/ui/dropdown-menu'
import {
  Select,
  SelectContent,
  SelectItem,
  SelectTrigger,
  SelectValue,
} from '@/components/ui/select'
import { paymentService } from '@/features/payments/paymentService'
import { PaymentDialog } from '@/features/payments/components/PaymentDialog'
import { PaymentDeleteDialog } from '@/features/payments/components/PaymentDeleteDialog'
import type { ApiPayment, PaymentMethod } from '@/types'
import { cn } from '@/lib/utils'
import { formatDate } from '@/lib/dateUtils'

const PER_PAGE_OPTIONS = [10, 20, 50] as const

const METHOD_ICONS: Record<PaymentMethod, React.ElementType> = {
  cash: Banknote,
  transfer: CreditCard,
  qr: Smartphone,
}

function toDateInput(iso: string | null | undefined): string {
  if (!iso) return ''
  return iso.slice(0, 10)
}

const emptyForm = {
  bill_id: '',
  amount: '',
  method: 'cash' as PaymentMethod,
  payment_date: '',
  note: '',
}

function formatBaht(n: number) {
  return n.toLocaleString('th-TH', { minimumFractionDigits: 2, maximumFractionDigits: 2 })
}

export function Payments() {
  const { t } = useTranslation()

  const [items, setItems] = useState<ApiPayment[]>([])
  const [loading, setLoading] = useState(true)
  const [error, setError] = useState('')
  const [search, setSearch] = useState('')
  const [filterMethod, setFilterMethod] = useState<string>('all')
  const [page, setPage] = useState(1)
  const [perPage, setPerPage] = useState<number>(20)
  const [total, setTotal] = useState(0)
  const [totalPages, setTotalPages] = useState(1)

  const [dialogOpen, setDialogOpen] = useState(false)
  const [deleteDialogOpen, setDeleteDialogOpen] = useState(false)
  const [editingItem, setEditingItem] = useState<ApiPayment | null>(null)
  const [deletingItem, setDeletingItem] = useState<ApiPayment | null>(null)
  const [form, setForm] = useState(emptyForm)
  const [saving, setSaving] = useState(false)

  const fetchItems = useCallback(async (p: number, pp: number) => {
    setLoading(true)
    setError('')
    try {
      const resp = await paymentService.list(p, pp)
      setItems(resp.data)
      setTotal(resp.meta.total)
      setTotalPages(resp.meta.total_pages)
    } catch {
      setError(t('payments.loadError'))
    } finally {
      setLoading(false)
    }
  }, [t])

  useEffect(() => {
    fetchItems(page, perPage)
  }, [fetchItems, page, perPage])

  function handlePerPageChange(value: string) {
    setPerPage(Number(value))
    setPage(1)
  }

  const filtered = useMemo(() => {
    const q = search.toLowerCase()
    return items.filter((item) => {
      const tenantName = `${item.tenant_first_name} ${item.tenant_last_name}`.toLowerCase()
      const matchSearch =
        tenantName.includes(q) ||
        item.room_number.toLowerCase().includes(q) ||
        item.billing_month.includes(q)
      const matchMethod = filterMethod === 'all' || item.method === filterMethod
      return matchSearch && matchMethod
    })
  }, [items, search, filterMethod])

  function openCreate() {
    setEditingItem(null)
    setForm({ ...emptyForm, payment_date: new Date().toISOString().slice(0, 10) })
    setDialogOpen(true)
  }

  function openEdit(item: ApiPayment) {
    setEditingItem(item)
    setForm({
      bill_id: item.bill_id,
      amount: String(item.amount),
      method: item.method,
      payment_date: toDateInput(item.payment_date),
      note: item.note,
    })
    setDialogOpen(true)
  }

  async function handleSave() {
    if (!form.bill_id || !form.amount || !form.payment_date) return
    setSaving(true)
    try {
      if (editingItem) {
        await paymentService.update(editingItem.id, {
          amount: Number(form.amount),
          method: form.method,
          payment_date: form.payment_date,
          note: form.note,
        })
      } else {
        await paymentService.create({
          bill_id: form.bill_id,
          amount: Number(form.amount),
          method: form.method,
          payment_date: form.payment_date,
          note: form.note,
        })
      }
      setDialogOpen(false)
      await fetchItems(page, perPage)
    } catch {
      // handled silently
    } finally {
      setSaving(false)
    }
  }

  async function handleDelete() {
    if (!deletingItem) return
    try {
      await paymentService.delete(deletingItem.id)
      const newPage = items.length === 1 && page > 1 ? page - 1 : page
      setPage(newPage)
      await fetchItems(newPage, perPage)
    } catch {
      // ignore
    } finally {
      setDeleteDialogOpen(false)
      setDeletingItem(null)
    }
  }

  return (
    <div className="space-y-6">
      <div className="flex flex-wrap items-start justify-between gap-3">
        <div className="min-w-0">
          <h1 className="text-2xl font-bold tracking-tight">{t('payments.title')}</h1>
          <p className="text-muted-foreground text-sm mt-1">
            {t('payments.subtitle', { count: filtered.length, total })}
          </p>
        </div>
        <div className="flex gap-2 shrink-0">
          <Button variant="outline" size="icon" onClick={() => fetchItems(page, perPage)} disabled={loading}>
            <RefreshCw className={cn('w-4 h-4', loading && 'animate-spin')} />
          </Button>
          <Button onClick={openCreate} className="gap-2">
            <Plus className="w-4 h-4" />
            <span className="hidden sm:inline">{t('payments.addPayment')}</span>
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
              <CardTitle>{t('payments.paymentList')}</CardTitle>
              <CardDescription>{t('payments.paymentListDesc')}</CardDescription>
            </div>
            <div className="flex flex-wrap gap-2">
              <Select value={filterMethod} onValueChange={setFilterMethod}>
                <SelectTrigger className="h-9 w-36">
                  <SelectValue />
                </SelectTrigger>
                <SelectContent>
                  <SelectItem value="all">{t('payments.allMethods')}</SelectItem>
                  <SelectItem value="cash">{t('payments.methods.cash')}</SelectItem>
                  <SelectItem value="transfer">{t('payments.methods.transfer')}</SelectItem>
                  <SelectItem value="qr">{t('payments.methods.qr')}</SelectItem>
                </SelectContent>
              </Select>
              <div className="relative">
                <Search className="absolute left-3 top-1/2 -translate-y-1/2 w-4 h-4 text-muted-foreground" />
                <Input
                  placeholder={t('payments.searchPlaceholder')}
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
                      <th className="text-left px-6 py-3 font-medium text-muted-foreground">{t('payments.colTenant')}</th>
                      <th className="text-left px-4 py-3 font-medium text-muted-foreground">{t('payments.colRoom')}</th>
                      <th className="text-left px-4 py-3 font-medium text-muted-foreground">{t('payments.colMonth')}</th>
                      <th className="text-left px-4 py-3 font-medium text-muted-foreground">{t('payments.colMethod')}</th>
                      <th className="text-right px-4 py-3 font-medium text-muted-foreground">{t('payments.colBillTotal')}</th>
                      <th className="text-right px-4 py-3 font-medium text-muted-foreground">{t('payments.colAmount')}</th>
                      <th className="text-left px-4 py-3 font-medium text-muted-foreground">{t('payments.colDate')}</th>
                      <th className="text-right px-6 py-3 font-medium text-muted-foreground">{t('payments.colActions')}</th>
                    </tr>
                  </thead>
                  <tbody>
                    {filtered.map((item, i) => {
                      const MethodIcon = METHOD_ICONS[item.method]
                      return (
                        <tr
                          key={item.id}
                          className={cn(
                            'border-b transition-colors hover:bg-muted/30',
                            i === filtered.length - 1 && 'border-0'
                          )}
                        >
                          <td className="px-6 py-4 font-medium">
                            {item.tenant_first_name} {item.tenant_last_name}
                          </td>
                          <td className="px-4 py-4">{item.room_number}</td>
                          <td className="px-4 py-4 text-muted-foreground">
                            {item.billing_month ? item.billing_month.slice(0, 7) : '-'}
                          </td>
                          <td className="px-4 py-4">
                            <Badge variant="outline" className="gap-1.5">
                              <MethodIcon className="w-3 h-3" />
                              {t(`payments.methods.${item.method}`)}
                            </Badge>
                          </td>
                          <td className="px-4 py-4 text-right text-muted-foreground">
                            ฿{formatBaht(item.bill_total)}
                          </td>
                          <td className="px-4 py-4 text-right font-semibold text-emerald-600">
                            ฿{formatBaht(item.amount)}
                          </td>
                          <td className="px-4 py-4 text-muted-foreground">{formatDate(item.payment_date)}</td>
                          <td className="px-6 py-4 text-right">
                            <DropdownMenu>
                              <DropdownMenuTrigger asChild>
                                <Button variant="ghost" size="icon" className="h-8 w-8">
                                  <MoreHorizontal className="w-4 h-4" />
                                </Button>
                              </DropdownMenuTrigger>
                              <DropdownMenuContent align="end">
                                <DropdownMenuItem onClick={() => openEdit(item)} className="gap-2">
                                  <Pencil className="w-4 h-4" /> {t('common.edit')}
                                </DropdownMenuItem>
                                <DropdownMenuSeparator />
                                <DropdownMenuItem
                                  className="gap-2 text-destructive focus:text-destructive"
                                  onClick={() => { setDeletingItem(item); setDeleteDialogOpen(true) }}
                                >
                                  <Trash2 className="w-4 h-4" /> {t('common.delete')}
                                </DropdownMenuItem>
                              </DropdownMenuContent>
                            </DropdownMenu>
                          </td>
                        </tr>
                      )
                    })}
                  </tbody>
                </table>

                {filtered.length === 0 && (
                  <div className="flex flex-col items-center justify-center py-16 gap-3 text-muted-foreground">
                    <div className="p-4 rounded-full bg-muted">
                      <SearchX className="w-6 h-6" />
                    </div>
                    <p className="text-sm font-medium">{t('payments.noPayments')}</p>
                    {(search || filterMethod !== 'all') && (
                      <button
                        type="button"
                        onClick={() => { setSearch(''); setFilterMethod('all') }}
                        className="text-xs text-primary hover:underline"
                      >
                        {t('payments.clearFilters')}
                      </button>
                    )}
                  </div>
                )}
              </div>

              {/* Mobile cards */}
              <div className="md:hidden divide-y">
                {filtered.map((item) => {
                  const MethodIcon = METHOD_ICONS[item.method]
                  return (
                    <div key={item.id} className="p-4 flex items-start gap-3">
                      <div className="flex-1 min-w-0 space-y-1">
                        <p className="text-sm font-medium">{item.tenant_first_name} {item.tenant_last_name}</p>
                        <p className="text-xs text-muted-foreground">
                          {t('payments.room')} {item.room_number} · {item.billing_month ? item.billing_month.slice(0, 7) : '-'}
                        </p>
                        <div className="flex items-center gap-2">
                          <Badge variant="outline" className="gap-1 text-xs">
                            <MethodIcon className="w-3 h-3" />
                            {t(`payments.methods.${item.method}`)}
                          </Badge>
                          <span className="text-sm font-semibold text-emerald-600">฿{formatBaht(item.amount)}</span>
                        </div>
                      </div>
                      <DropdownMenu>
                        <DropdownMenuTrigger asChild>
                          <Button variant="ghost" size="icon" className="h-8 w-8 shrink-0">
                            <MoreHorizontal className="w-4 h-4" />
                          </Button>
                        </DropdownMenuTrigger>
                        <DropdownMenuContent align="end">
                          <DropdownMenuItem onClick={() => openEdit(item)}>{t('common.edit')}</DropdownMenuItem>
                          <DropdownMenuSeparator />
                          <DropdownMenuItem
                            className="text-destructive focus:text-destructive"
                            onClick={() => { setDeletingItem(item); setDeleteDialogOpen(true) }}
                          >
                            {t('common.delete')}
                          </DropdownMenuItem>
                        </DropdownMenuContent>
                      </DropdownMenu>
                    </div>
                  )
                })}
              </div>

              {/* Pagination */}
              {!loading && total > 0 && (
                <div className="flex flex-col sm:flex-row items-center justify-between gap-3 px-4 py-3 border-t text-sm text-muted-foreground">
                  <span className="shrink-0">
                    {t('payments.showing', {
                      from: (page - 1) * perPage + 1,
                      to: Math.min(page * perPage, total),
                      total,
                    })}
                  </span>
                  <div className="flex items-center gap-4">
                    <div className="flex items-center gap-2">
                      <span className="hidden sm:inline shrink-0">{t('payments.perPage')}</span>
                      <Select value={String(perPage)} onValueChange={handlePerPageChange}>
                        <SelectTrigger className="h-8 w-16"><SelectValue /></SelectTrigger>
                        <SelectContent>
                          {PER_PAGE_OPTIONS.map((n) => (
                            <SelectItem key={n} value={String(n)}>{n}</SelectItem>
                          ))}
                        </SelectContent>
                      </Select>
                    </div>
                    <span className="shrink-0">{t('payments.page')} {page} {t('payments.of')} {totalPages}</span>
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

      <PaymentDialog
        open={dialogOpen}
        onOpenChange={setDialogOpen}
        editingItem={editingItem}
        form={form}
        onFormChange={setForm}
        onSave={handleSave}
        saving={saving}
      />

      <PaymentDeleteDialog
        open={deleteDialogOpen}
        onOpenChange={setDeleteDialogOpen}
        item={deletingItem}
        onDelete={handleDelete}
      />
    </div>
  )
}
