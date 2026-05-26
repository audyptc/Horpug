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
  CheckCircle2,
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
import { billService } from '@/services/billService'
import { contractService } from '@/services/contractService'
import type { ApiBill, ApiContract, BillStatus } from '@/types/api'
import { cn } from '@/lib/utils'
import { formatDate } from '@/lib/dateUtils'

const PER_PAGE_OPTIONS = [10, 20, 50] as const

const STATUS_VARIANTS: Record<BillStatus, 'default' | 'secondary' | 'destructive'> = {
  unpaid: 'secondary',
  paid: 'default',
  overdue: 'destructive',
}

function toMonthInput(iso: string | null | undefined): string {
  if (!iso) return ''
  return iso.slice(0, 7)
}

function toDateInput(iso: string | null | undefined): string {
  if (!iso) return ''
  return iso.slice(0, 10)
}

const emptyForm = {
  contract_id: '',
  billing_month: '',
  rent_amount: '',
  electric_amount: '',
  water_amount: '',
  other_amount: '',
  other_note: '',
  due_date: '',
  note: '',
  status: 'unpaid' as BillStatus,
}

export function Bills() {
  const { t } = useTranslation()

  const [bills, setBills] = useState<ApiBill[]>([])
  const [loading, setLoading] = useState(true)
  const [error, setError] = useState('')
  const [search, setSearch] = useState('')
  const [filterStatus, setFilterStatus] = useState<string>('all')
  const [page, setPage] = useState(1)
  const [perPage, setPerPage] = useState<number>(20)
  const [total, setTotal] = useState(0)
  const [totalPages, setTotalPages] = useState(1)

  const [contracts, setContracts] = useState<ApiContract[]>([])

  const [dialogOpen, setDialogOpen] = useState(false)
  const [deleteDialogOpen, setDeleteDialogOpen] = useState(false)
  const [editingBill, setEditingBill] = useState<ApiBill | null>(null)
  const [deletingBill, setDeletingBill] = useState<ApiBill | null>(null)
  const [form, setForm] = useState(emptyForm)
  const [saving, setSaving] = useState(false)

  const fetchBills = useCallback(async (p: number, pp: number) => {
    setLoading(true)
    setError('')
    try {
      const resp = await billService.list(p, pp)
      setBills(resp.data)
      setTotal(resp.meta.total)
      setTotalPages(resp.meta.total_pages)
    } catch {
      setError(t('bills.loadError'))
    } finally {
      setLoading(false)
    }
  }, [t])

  useEffect(() => {
    fetchBills(page, perPage)
  }, [fetchBills, page, perPage])

  useEffect(() => {
    contractService.list(1, 200).then((r) => setContracts(r.data)).catch(() => {})
  }, [])

  function handlePerPageChange(value: string) {
    setPerPage(Number(value))
    setPage(1)
  }

  const filtered = useMemo(() => {
    const q = search.toLowerCase()
    return bills.filter((b) => {
      const matchSearch =
        `${b.tenant_first_name} ${b.tenant_last_name}`.toLowerCase().includes(q) ||
        b.room_number.toLowerCase().includes(q)
      const matchStatus = filterStatus === 'all' || b.status === filterStatus
      return matchSearch && matchStatus
    })
  }, [bills, search, filterStatus])

  function openCreate() {
    setEditingBill(null)
    setForm(emptyForm)
    setDialogOpen(true)
  }

  function openEdit(bill: ApiBill) {
    setEditingBill(bill)
    setForm({
      contract_id: bill.contract_id,
      billing_month: toMonthInput(bill.billing_month),
      rent_amount: String(bill.rent_amount),
      electric_amount: String(bill.electric_amount),
      water_amount: String(bill.water_amount),
      other_amount: String(bill.other_amount),
      other_note: bill.other_note,
      due_date: toDateInput(bill.due_date),
      note: bill.note,
      status: bill.status,
    })
    setDialogOpen(true)
  }

  const previewTotal =
    (Number(form.rent_amount) || 0) +
    (Number(form.electric_amount) || 0) +
    (Number(form.water_amount) || 0) +
    (Number(form.other_amount) || 0)

  async function handleSave() {
    if (!form.contract_id || !form.billing_month) return
    setSaving(true)
    try {
      if (editingBill) {
        await billService.update(editingBill.id, {
          rent_amount: Number(form.rent_amount) || 0,
          electric_amount: Number(form.electric_amount) || 0,
          water_amount: Number(form.water_amount) || 0,
          other_amount: Number(form.other_amount) || 0,
          other_note: form.other_note,
          status: form.status,
          due_date: form.due_date || null,
          note: form.note,
        })
      } else {
        await billService.create({
          contract_id: form.contract_id,
          billing_month: form.billing_month + '-01',
          rent_amount: Number(form.rent_amount) || 0,
          electric_amount: Number(form.electric_amount) || 0,
          water_amount: Number(form.water_amount) || 0,
          other_amount: Number(form.other_amount) || 0,
          other_note: form.other_note,
          due_date: form.due_date || null,
          note: form.note,
        })
      }
      setDialogOpen(false)
      await fetchBills(page, perPage)
    } catch {
      // handled silently
    } finally {
      setSaving(false)
    }
  }

  async function handleMarkPaid(bill: ApiBill) {
    try {
      await billService.update(bill.id, {
        rent_amount: bill.rent_amount,
        electric_amount: bill.electric_amount,
        water_amount: bill.water_amount,
        other_amount: bill.other_amount,
        other_note: bill.other_note,
        status: 'paid',
        due_date: bill.due_date,
        note: bill.note,
      })
      await fetchBills(page, perPage)
    } catch {
      // ignore
    }
  }

  async function handleDelete() {
    if (!deletingBill) return
    try {
      await billService.delete(deletingBill.id)
      const newPage = bills.length === 1 && page > 1 ? page - 1 : page
      setPage(newPage)
      await fetchBills(newPage, perPage)
    } catch {
      // ignore
    } finally {
      setDeleteDialogOpen(false)
      setDeletingBill(null)
    }
  }

  const isSaveDisabled = saving || !form.contract_id || !form.billing_month

  const activeContracts = useMemo(
    () => contracts.filter((c) => c.status === 'active'),
    [contracts]
  )

  return (
    <div className="space-y-6">
      <div className="flex flex-wrap items-start justify-between gap-3">
        <div className="min-w-0">
          <h1 className="text-2xl font-bold tracking-tight">{t('bills.title')}</h1>
          <p className="text-muted-foreground text-sm mt-1">
            {t('bills.subtitle', { count: filtered.length, total })}
          </p>
        </div>
        <div className="flex gap-2 shrink-0">
          <Button variant="outline" size="icon" onClick={() => fetchBills(page, perPage)} disabled={loading}>
            <RefreshCw className={cn('w-4 h-4', loading && 'animate-spin')} />
          </Button>
          <Button onClick={openCreate} className="gap-2">
            <Plus className="w-4 h-4" />
            <span className="hidden sm:inline">{t('bills.addBill')}</span>
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
              <CardTitle>{t('bills.billList')}</CardTitle>
              <CardDescription>{t('bills.billListDesc')}</CardDescription>
            </div>
            <div className="flex gap-2">
              <Select value={filterStatus} onValueChange={setFilterStatus}>
                <SelectTrigger className="h-9 w-36">
                  <SelectValue />
                </SelectTrigger>
                <SelectContent>
                  <SelectItem value="all">{t('bills.allStatuses')}</SelectItem>
                  <SelectItem value="unpaid">{t('bills.statuses.unpaid')}</SelectItem>
                  <SelectItem value="paid">{t('bills.statuses.paid')}</SelectItem>
                  <SelectItem value="overdue">{t('bills.statuses.overdue')}</SelectItem>
                </SelectContent>
              </Select>
              <div className="relative">
                <Search className="absolute left-3 top-1/2 -translate-y-1/2 w-4 h-4 text-muted-foreground" />
                <Input
                  placeholder={t('bills.searchPlaceholder')}
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
                      <th className="text-left px-6 py-3 font-medium text-muted-foreground">{t('bills.colTenant')}</th>
                      <th className="text-left px-4 py-3 font-medium text-muted-foreground">{t('bills.colRoom')}</th>
                      <th className="text-left px-4 py-3 font-medium text-muted-foreground">{t('bills.colMonth')}</th>
                      <th className="text-right px-4 py-3 font-medium text-muted-foreground">{t('bills.colRent')}</th>
                      <th className="text-right px-4 py-3 font-medium text-muted-foreground">{t('bills.colElectric')}</th>
                      <th className="text-right px-4 py-3 font-medium text-muted-foreground">{t('bills.colWater')}</th>
                      <th className="text-right px-4 py-3 font-medium text-muted-foreground">{t('bills.colTotal')}</th>
                      <th className="text-center px-4 py-3 font-medium text-muted-foreground">{t('bills.colStatus')}</th>
                      <th className="text-right px-6 py-3 font-medium text-muted-foreground">{t('bills.colActions')}</th>
                    </tr>
                  </thead>
                  <tbody>
                    {filtered.map((bill, i) => (
                      <tr
                        key={bill.id}
                        className={cn(
                          'border-b transition-colors hover:bg-muted/30',
                          i === filtered.length - 1 && 'border-0'
                        )}
                      >
                        <td className="px-6 py-4 font-medium">
                          {bill.tenant_first_name} {bill.tenant_last_name}
                        </td>
                        <td className="px-4 py-4 text-muted-foreground">{bill.room_number}</td>
                        <td className="px-4 py-4 text-muted-foreground">{formatDate(bill.billing_month)}</td>
                        <td className="px-4 py-4 text-right text-muted-foreground">{bill.rent_amount.toLocaleString()}</td>
                        <td className="px-4 py-4 text-right text-muted-foreground">{bill.electric_amount.toLocaleString()}</td>
                        <td className="px-4 py-4 text-right text-muted-foreground">{bill.water_amount.toLocaleString()}</td>
                        <td className="px-4 py-4 text-right font-semibold">
                          {bill.total_amount.toLocaleString()}
                          <span className="text-xs text-muted-foreground ml-1">฿</span>
                        </td>
                        <td className="px-4 py-4 text-center">
                          <Badge variant={STATUS_VARIANTS[bill.status]}>
                            {t(`bills.statuses.${bill.status}`)}
                          </Badge>
                        </td>
                        <td className="px-6 py-4 text-right">
                          <DropdownMenu>
                            <DropdownMenuTrigger asChild>
                              <Button variant="ghost" size="icon" className="h-8 w-8">
                                <MoreHorizontal className="w-4 h-4" />
                              </Button>
                            </DropdownMenuTrigger>
                            <DropdownMenuContent align="end">
                              {bill.status !== 'paid' && (
                                <DropdownMenuItem onClick={() => handleMarkPaid(bill)} className="gap-2">
                                  <CheckCircle2 className="w-4 h-4" /> {t('bills.markPaid')}
                                </DropdownMenuItem>
                              )}
                              <DropdownMenuItem onClick={() => openEdit(bill)} className="gap-2">
                                <Pencil className="w-4 h-4" /> {t('common.edit')}
                              </DropdownMenuItem>
                              <DropdownMenuSeparator />
                              <DropdownMenuItem
                                className="gap-2 text-destructive focus:text-destructive"
                                onClick={() => { setDeletingBill(bill); setDeleteDialogOpen(true) }}
                              >
                                <Trash2 className="w-4 h-4" /> {t('common.delete')}
                              </DropdownMenuItem>
                            </DropdownMenuContent>
                          </DropdownMenu>
                        </td>
                      </tr>
                    ))}
                  </tbody>
                </table>

                {filtered.length === 0 && (
                  <div className="flex flex-col items-center justify-center py-16 gap-3 text-muted-foreground">
                    <div className="p-4 rounded-full bg-muted">
                      <SearchX className="w-6 h-6" />
                    </div>
                    <p className="text-sm font-medium">{t('bills.noBills')}</p>
                    {(search || filterStatus !== 'all') && (
                      <button
                        type="button"
                        onClick={() => { setSearch(''); setFilterStatus('all') }}
                        className="text-xs text-primary hover:underline"
                      >
                        {t('bills.clearFilters')}
                      </button>
                    )}
                  </div>
                )}
              </div>

              {/* Mobile cards */}
              <div className="md:hidden divide-y">
                {filtered.map((bill) => (
                  <div key={bill.id} className="p-4 flex items-start gap-3">
                    <div className="flex-1 min-w-0 space-y-1">
                      <p className="text-sm font-medium">
                        {bill.tenant_first_name} {bill.tenant_last_name}
                      </p>
                      <p className="text-xs text-muted-foreground">
                        {t('bills.colRoom')} {bill.room_number} · {formatDate(bill.billing_month)}
                      </p>
                      <p className="text-sm font-semibold">{bill.total_amount.toLocaleString()} ฿</p>
                      <Badge variant={STATUS_VARIANTS[bill.status]} className="text-xs">
                        {t(`bills.statuses.${bill.status}`)}
                      </Badge>
                    </div>
                    <DropdownMenu>
                      <DropdownMenuTrigger asChild>
                        <Button variant="ghost" size="icon" className="h-8 w-8 shrink-0">
                          <MoreHorizontal className="w-4 h-4" />
                        </Button>
                      </DropdownMenuTrigger>
                      <DropdownMenuContent align="end">
                        {bill.status !== 'paid' && (
                          <DropdownMenuItem onClick={() => handleMarkPaid(bill)}>
                            {t('bills.markPaid')}
                          </DropdownMenuItem>
                        )}
                        <DropdownMenuItem onClick={() => openEdit(bill)}>{t('common.edit')}</DropdownMenuItem>
                        <DropdownMenuSeparator />
                        <DropdownMenuItem
                          className="text-destructive focus:text-destructive"
                          onClick={() => { setDeletingBill(bill); setDeleteDialogOpen(true) }}
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
                    {t('bills.showing', {
                      from: (page - 1) * perPage + 1,
                      to: Math.min(page * perPage, total),
                      total,
                    })}
                  </span>
                  <div className="flex items-center gap-4">
                    <div className="flex items-center gap-2">
                      <span className="hidden sm:inline shrink-0">{t('bills.perPage')}</span>
                      <Select value={String(perPage)} onValueChange={handlePerPageChange}>
                        <SelectTrigger className="h-8 w-16"><SelectValue /></SelectTrigger>
                        <SelectContent>
                          {PER_PAGE_OPTIONS.map((n) => (
                            <SelectItem key={n} value={String(n)}>{n}</SelectItem>
                          ))}
                        </SelectContent>
                      </Select>
                    </div>
                    <span className="shrink-0">{t('bills.page')} {page} {t('bills.of')} {totalPages}</span>
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
              {editingBill ? t('bills.editBill') : t('bills.createBill')}
            </DialogTitle>
            <DialogDescription>
              {editingBill ? t('bills.editDesc') : t('bills.createDesc')}
            </DialogDescription>
          </DialogHeader>

          <div className="grid gap-4 py-4">
            {/* Contract */}
            <div className="space-y-1.5">
              <Label>{t('bills.contract')} *</Label>
              <Select
                value={form.contract_id}
                onValueChange={(v) => setForm((f) => ({ ...f, contract_id: v }))}
                disabled={!!editingBill}
              >
                <SelectTrigger><SelectValue placeholder={t('bills.selectContract')} /></SelectTrigger>
                <SelectContent>
                  {activeContracts.map((c) => (
                    <SelectItem key={c.id} value={c.id}>
                      {c.tenant_first_name} {c.tenant_last_name} — {t('bills.colRoom')} {c.room_number}
                    </SelectItem>
                  ))}
                </SelectContent>
              </Select>
            </div>

            {/* Billing month */}
            <div className="space-y-1.5">
              <Label>{t('bills.billingMonth')} *</Label>
              <Input
                type="month"
                value={form.billing_month}
                onChange={(e) => setForm((f) => ({ ...f, billing_month: e.target.value }))}
                disabled={!!editingBill}
              />
            </div>

            {/* Amounts */}
            <div className="grid grid-cols-2 gap-4">
              <div className="space-y-1.5">
                <Label>{t('bills.rentAmount')}</Label>
                <Input
                  type="number" min="0" step="0.01"
                  value={form.rent_amount}
                  onChange={(e) => setForm((f) => ({ ...f, rent_amount: e.target.value }))}
                />
              </div>
              <div className="space-y-1.5">
                <Label>{t('bills.electricAmount')}</Label>
                <Input
                  type="number" min="0" step="0.01"
                  value={form.electric_amount}
                  onChange={(e) => setForm((f) => ({ ...f, electric_amount: e.target.value }))}
                />
              </div>
              <div className="space-y-1.5">
                <Label>{t('bills.waterAmount')}</Label>
                <Input
                  type="number" min="0" step="0.01"
                  value={form.water_amount}
                  onChange={(e) => setForm((f) => ({ ...f, water_amount: e.target.value }))}
                />
              </div>
              <div className="space-y-1.5">
                <Label>{t('bills.otherAmount')}</Label>
                <Input
                  type="number" min="0" step="0.01"
                  value={form.other_amount}
                  onChange={(e) => setForm((f) => ({ ...f, other_amount: e.target.value }))}
                />
              </div>
            </div>

            {/* Other note */}
            <div className="space-y-1.5">
              <Label>{t('bills.otherNote')}</Label>
              <Input
                placeholder={t('bills.otherNotePlaceholder')}
                value={form.other_note}
                onChange={(e) => setForm((f) => ({ ...f, other_note: e.target.value }))}
              />
            </div>

            {/* Total preview */}
            <div className="rounded-md border bg-muted/40 px-4 py-3 text-sm flex justify-between">
              <span className="text-muted-foreground">{t('bills.totalAmount')}</span>
              <span className="font-semibold">{previewTotal.toLocaleString()} ฿</span>
            </div>

            {/* Status (edit only) */}
            {editingBill && (
              <div className="space-y-1.5">
                <Label>{t('bills.status')}</Label>
                <Select
                  value={form.status}
                  onValueChange={(v) => setForm((f) => ({ ...f, status: v as BillStatus }))}
                >
                  <SelectTrigger><SelectValue /></SelectTrigger>
                  <SelectContent>
                    <SelectItem value="unpaid">{t('bills.statuses.unpaid')}</SelectItem>
                    <SelectItem value="paid">{t('bills.statuses.paid')}</SelectItem>
                    <SelectItem value="overdue">{t('bills.statuses.overdue')}</SelectItem>
                  </SelectContent>
                </Select>
              </div>
            )}

            {/* Due date */}
            <div className="space-y-1.5">
              <Label>{t('bills.dueDate')}</Label>
              <Input
                type="date"
                value={form.due_date}
                onChange={(e) => setForm((f) => ({ ...f, due_date: e.target.value }))}
              />
            </div>

            {/* Note */}
            <div className="space-y-1.5">
              <Label>{t('bills.note')}</Label>
              <textarea
                rows={2}
                placeholder={t('bills.notePlaceholder')}
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
                : editingBill
                  ? t('bills.saveChanges')
                  : t('bills.createBill')}
            </Button>
          </DialogFooter>
        </DialogContent>
      </Dialog>

      {/* Delete confirm */}
      <Dialog open={deleteDialogOpen} onOpenChange={setDeleteDialogOpen}>
        <DialogContent className="max-w-sm">
          <DialogHeader>
            <DialogTitle>{t('bills.deleteBill')}</DialogTitle>
            <DialogDescription>
              {t('bills.deleteConfirm')}{' '}
              <span className="font-semibold text-foreground">
                {deletingBill?.tenant_first_name} {deletingBill?.tenant_last_name}
              </span>?{' '}
              {t('bills.deleteWarning')}
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
