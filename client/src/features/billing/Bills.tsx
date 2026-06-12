import { useState, useMemo, useEffect, useCallback } from 'react'
import { useTranslation } from 'react-i18next'
import { Plus, Search } from 'lucide-react'
import { Button } from '@/components/ui/button'
import { Input } from '@/components/ui/input'
import { Card, CardContent, CardHeader, CardTitle, CardDescription } from '@/components/ui/card'
import {
  Select,
  SelectContent,
  SelectItem,
  SelectTrigger,
  SelectValue,
} from '@/components/ui/select'
import {
  Dialog,
  DialogContent,
  DialogDescription,
  DialogFooter,
  DialogHeader,
  DialogTitle,
} from '@/components/ui/dialog'
import { PageHeader } from '@/components/shared/PageHeader'
import { Pagination } from '@/components/shared/Pagination'
import { ErrorBanner } from '@/components/shared/ErrorBanner'
import { BillTable } from '@/features/billing/components/BillTable'
import { BillDialog, type BillFormState } from '@/features/billing/components/BillDialog'
import { usePaginatedList } from '@/hooks/usePaginatedList'
import { billService } from '@/features/billing/billService'
import { contractService } from '@/features/contracts/contractService'
import type { ApiBill, ApiContract } from '@/types'

const EMPTY_FORM: BillFormState = {
  contract_id: '',
  billing_month: '',
  rent_amount: '',
  electric_amount: '',
  water_amount: '',
  other_items: [],
  due_date: '',
  note: '',
  status: 'unpaid',
}

function toMonthInput(iso: string | null | undefined) {
  return iso ? iso.slice(0, 7) : ''
}

function toDateInput(iso: string | null | undefined) {
  return iso ? iso.slice(0, 10) : ''
}

export function Bills() {
  const { t } = useTranslation()

  const { items: bills, loading, hasError, page, perPage, total, totalPages, setPage, setPerPage, refresh, deletedOnPage } =
    usePaginatedList(billService.list, { initialPerPage: 20 })

  const [contracts, setContracts] = useState<ApiContract[]>([])
  const [search, setSearch] = useState('')
  const [filterStatus, setFilterStatus] = useState('all')

  const [dialogOpen, setDialogOpen] = useState(false)
  const [deleteDialogOpen, setDeleteDialogOpen] = useState(false)
  const [editingBill, setEditingBill] = useState<ApiBill | null>(null)
  const [deletingBill, setDeletingBill] = useState<ApiBill | null>(null)
  const [form, setForm] = useState<BillFormState>(EMPTY_FORM)
  const [saving, setSaving] = useState(false)

  useEffect(() => {
    contractService.list(1, 200).then((r) => setContracts(r.data)).catch(() => {})
  }, [])

  const filtered = useMemo(() => {
    const q = search.toLowerCase()
    return bills.filter((b) => {
      const matchSearch =
        `${b.tenant_first_name} ${b.tenant_last_name}`.toLowerCase().includes(q) ||
        b.room_number.toLowerCase().includes(q)
      return matchSearch && (filterStatus === 'all' || b.status === filterStatus)
    })
  }, [bills, search, filterStatus])

  function openCreate() {
    setEditingBill(null)
    setForm(EMPTY_FORM)
    setDialogOpen(true)
  }

  async function openEdit(bill: ApiBill) {
    setEditingBill(bill)
    setForm({
      contract_id: bill.contract_id,
      billing_month: toMonthInput(bill.billing_month),
      rent_amount: String(bill.rent_amount),
      electric_amount: String(bill.electric_amount),
      water_amount: String(bill.water_amount),
      other_items: (bill.other_items ?? []).map((item) => ({
        label: item.label,
        amount: String(item.amount),
      })),
      due_date: toDateInput(bill.due_date),
      note: bill.note,
      status: bill.status,
    })
    setDialogOpen(true)
    try {
      const detail = await billService.getById(bill.id)
      setForm((f) => ({
        ...f,
        other_items: detail.other_items.map((item) => ({
          label: item.label,
          amount: String(item.amount),
        })),
      }))
    } catch {
      // keep the empty list from list response
    }
  }

  const handleSave = useCallback(async () => {
    if (!form.contract_id || !form.billing_month) return
    setSaving(true)
    const otherItems = form.other_items
      .map((item) => ({ label: item.label, amount: Number(item.amount) || 0 }))
      .filter((item) => item.label || item.amount > 0)
    try {
      if (editingBill) {
        await billService.update(editingBill.id, {
          rent_amount: Number(form.rent_amount) || 0,
          electric_amount: Number(form.electric_amount) || 0,
          water_amount: Number(form.water_amount) || 0,
          other_items: otherItems,
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
          other_items: otherItems,
          due_date: form.due_date || null,
          note: form.note,
        })
      }
      setDialogOpen(false)
      refresh()
    } catch {
      // silent
    } finally {
      setSaving(false)
    }
  }, [form, editingBill, refresh])

  const handleMarkPaid = useCallback(async (bill: ApiBill) => {
    try {
      await billService.update(bill.id, {
        rent_amount: bill.rent_amount,
        electric_amount: bill.electric_amount,
        water_amount: bill.water_amount,
        status: 'paid',
        due_date: bill.due_date,
        note: bill.note,
      })
      refresh()
    } catch {
      // silent
    }
  }, [refresh])

  async function handleDelete() {
    if (!deletingBill) return
    try {
      await billService.delete(deletingBill.id)
      deletedOnPage()
    } catch {
      // silent
    } finally {
      setDeleteDialogOpen(false)
      setDeletingBill(null)
    }
  }

  function openDelete(bill: ApiBill) {
    setDeletingBill(bill)
    setDeleteDialogOpen(true)
  }

  function clearFilters() {
    setSearch('')
    setFilterStatus('all')
  }

  return (
    <div className="space-y-6">
      <PageHeader
        title={t('bills.title')}
        subtitle={t('bills.subtitle', { count: filtered.length, total })}
        loading={loading}
        onRefresh={refresh}
      >
        <Button onClick={openCreate} className="gap-2">
          <Plus className="w-4 h-4" />
          <span className="hidden sm:inline">{t('bills.addBill')}</span>
        </Button>
      </PageHeader>

      {hasError && <ErrorBanner message={t('bills.loadError')} />}

      <Card>
        <CardHeader>
          <div className="flex flex-col sm:flex-row sm:items-center justify-between gap-3">
            <div>
              <CardTitle>{t('bills.billList')}</CardTitle>
              <CardDescription>{t('bills.billListDesc')}</CardDescription>
            </div>
            <div className="flex gap-2">
              <Select value={filterStatus} onValueChange={setFilterStatus}>
                <SelectTrigger className="h-9 w-36"><SelectValue /></SelectTrigger>
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
              <BillTable
                bills={filtered}
                hasFilters={!!search || filterStatus !== 'all'}
                onEdit={openEdit}
                onMarkPaid={handleMarkPaid}
                onDelete={openDelete}
                onClearFilters={clearFilters}
              />
              <Pagination
                page={page}
                totalPages={totalPages}
                total={total}
                perPage={perPage}
                onPageChange={setPage}
                onPerPageChange={setPerPage}
              />
            </>
          )}
        </CardContent>
      </Card>

      <BillDialog
        open={dialogOpen}
        onOpenChange={setDialogOpen}
        editingBill={editingBill}
        form={form}
        onFormChange={(patch) => setForm((f) => ({ ...f, ...patch }))}
        contracts={contracts}
        onSave={handleSave}
        saving={saving}
      />

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
            <Button variant="outline" onClick={() => setDeleteDialogOpen(false)}>{t('common.cancel')}</Button>
            <Button variant="destructive" onClick={handleDelete}>{t('common.delete')}</Button>
          </DialogFooter>
        </DialogContent>
      </Dialog>
    </div>
  )
}
