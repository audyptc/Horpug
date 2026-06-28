import { useState, useMemo, useEffect, useCallback, useRef } from 'react'
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
import { usePermission } from '@/hooks/usePermission'
import { billService } from '@/features/billing/billService'
import { contractService } from '@/features/contracts/contractService'
import { electricMeterService } from '@/features/electric-meters/electricMeterService'
import { waterMeterService } from '@/features/water-meters/waterMeterService'
import { parkingService } from '@/features/parking/parkingService'
import { paymentService } from '@/features/payments/paymentService'
import { PaymentDialog } from '@/features/payments/components/PaymentDialog'
import type { ApiBill, ApiContract, ApiParkingSlot, PaymentMethod } from '@/types'

const EMPTY_FORM: BillFormState = {
  contract_id: '',
  billing_month: '',
  rent_amount: '',
  electric_amount: '',
  water_amount: '',
  parking_amount: '',
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
  const { canCreate, canUpdate, canDelete } = usePermission('/bills')

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
  const [autoFilling, setAutoFilling] = useState(false)

  const [tenantParkingSlots, setTenantParkingSlots] = useState<ApiParkingSlot[]>([])
  const [selectedSlotIds, setSelectedSlotIds] = useState<string[]>([])

  const [paymentDialogOpen, setPaymentDialogOpen] = useState(false)
  const [paymentForm, setPaymentForm] = useState({
    bill_id: '',
    amount: '',
    method: 'cash' as PaymentMethod,
    payment_date: '',
    note: '',
  })
  const [paymentSaving, setPaymentSaving] = useState(false)
  const [paymentError, setPaymentError] = useState('')

  const contractIdRef = useRef('')
  const billingMonthRef = useRef('')
  contractIdRef.current = form.contract_id
  billingMonthRef.current = form.billing_month

  useEffect(() => {
    contractService.list(1, 200).then((r) => setContracts(r.data)).catch(() => {})
  }, [])

  function onFormChange(patch: Partial<BillFormState>) {
    setForm((f) => ({ ...f, ...patch }))
  }

  const handleContractChange = useCallback(async (contractId: string) => {
    const contract = contracts.find((c) => c.id === contractId)
    if (!contract) return
    setForm((f) => ({ ...f, contract_id: contractId, rent_amount: String(contract.rent_price) }))
    setAutoFilling(true)
    const billingMonth = billingMonthRef.current || undefined
    const [electric, water, parkingResp] = await Promise.all([
      electricMeterService.getLatestByRoomId(contract.room_id, billingMonth),
      waterMeterService.getLatestByRoomId(contract.room_id, billingMonth),
      parkingService.list(1, 200),
    ])
    setAutoFilling(false)
    const slots = parkingResp.data.filter(
      (s) => s.tenant_id === contract.tenant_id && s.status === 'occupied'
    )
    setTenantParkingSlots(slots)
    const allIds = slots.map((s) => s.id)
    setSelectedSlotIds(allIds)
    const parkingTotal = slots.reduce((sum, s) => sum + s.monthly_fee, 0)
    setForm((f) => ({
      ...f,
      electric_amount: electric ? String(electric.total_amount) : f.electric_amount,
      water_amount: water ? String(water.total_amount) : f.water_amount,
      parking_amount: parkingTotal > 0 ? String(parkingTotal) : '',
    }))
  }, [contracts])

  const handleBillingMonthChange = useCallback(async (billingMonth: string) => {
    setForm((f) => ({ ...f, billing_month: billingMonth }))
    const contractId = contractIdRef.current
    if (!contractId) return
    const contract = contracts.find((c) => c.id === contractId)
    if (!contract) return
    setAutoFilling(true)
    try {
      const [electric, water] = await Promise.all([
        electricMeterService.getLatestByRoomId(contract.room_id, billingMonth || undefined),
        waterMeterService.getLatestByRoomId(contract.room_id, billingMonth || undefined),
      ])
      setForm((f) => ({
        ...f,
        electric_amount: electric ? String(electric.total_amount) : f.electric_amount,
        water_amount: water ? String(water.total_amount) : f.water_amount,
      }))
    } finally {
      setAutoFilling(false)
    }
  }, [contracts])

  const handleSlotToggle = useCallback((slotId: string) => {
    setSelectedSlotIds((prev) => {
      const next = prev.includes(slotId) ? prev.filter((id) => id !== slotId) : [...prev, slotId]
      const total = tenantParkingSlots
        .filter((s) => next.includes(s.id))
        .reduce((sum, s) => sum + s.monthly_fee, 0)
      setForm((f) => ({ ...f, parking_amount: total > 0 ? String(total) : '' }))
      return next
    })
  }, [tenantParkingSlots])

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
    setTenantParkingSlots([])
    setSelectedSlotIds([])
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
      parking_amount: bill.parking_amount > 0 ? String(bill.parking_amount) : '',
      other_items: (bill.other_items ?? []).map((item) => ({
        label: item.label,
        amount: String(item.amount),
      })),
      due_date: toDateInput(bill.due_date),
      note: bill.note,
      status: bill.status,
    })
    setTenantParkingSlots([])
    setSelectedSlotIds([])
    setDialogOpen(true)

    // โหลด slots และ other_items พร้อมกัน
    const contract = contracts.find((c) => c.id === bill.contract_id)
    const [detail, parkingResp] = await Promise.all([
      billService.getById(bill.id).catch(() => null),
      contract ? parkingService.list(1, 200).catch(() => null) : Promise.resolve(null),
    ])
    if (detail) {
      setForm((f) => ({
        ...f,
        other_items: detail.other_items.map((item) => ({
          label: item.label,
          amount: String(item.amount),
        })),
      }))
    }
    if (parkingResp && contract) {
      const slots = parkingResp.data.filter(
        (s) => s.tenant_id === contract.tenant_id && s.status === 'occupied'
      )
      setTenantParkingSlots(slots)
      // pre-select ช่องที่ผลรวมตรงกับที่บันทึกไว้
      const stored = bill.parking_amount
      if (stored > 0) {
        const matching = slots.filter((s) => s.monthly_fee === stored)
        if (matching.length === 1) {
          setSelectedSlotIds([matching[0].id])
        } else {
          const allSum = slots.reduce((sum, s) => sum + s.monthly_fee, 0)
          if (allSum === stored) setSelectedSlotIds(slots.map((s) => s.id))
        }
      }
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
          parking_amount: Number(form.parking_amount) || 0,
          other_items: otherItems,
          status: form.status,
          due_date: form.due_date ? `${form.due_date}T00:00:00Z` : null,
          note: form.note,
        })
      } else {
        await billService.create({
          contract_id: form.contract_id,
          billing_month: `${form.billing_month}-01T00:00:00Z`,
          rent_amount: Number(form.rent_amount) || 0,
          electric_amount: Number(form.electric_amount) || 0,
          water_amount: Number(form.water_amount) || 0,
          parking_amount: Number(form.parking_amount) || 0,
          other_items: otherItems,
          due_date: form.due_date ? `${form.due_date}T00:00:00Z` : null,
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

  const handleMarkPaid = useCallback((bill: ApiBill) => {
    setPaymentForm({
      bill_id: bill.id,
      amount: String(bill.total_amount),
      method: 'cash',
      payment_date: new Date().toISOString().slice(0, 10),
      note: '',
    })
    setPaymentError('')
    setPaymentDialogOpen(true)
  }, [])

  const handlePaymentSave = useCallback(async () => {
    if (!paymentForm.bill_id || !paymentForm.amount || !paymentForm.payment_date) return
    setPaymentSaving(true)
    setPaymentError('')
    try {
      await paymentService.create({
        bill_id: paymentForm.bill_id,
        amount: Number(paymentForm.amount),
        method: paymentForm.method,
        payment_date: paymentForm.payment_date + 'T00:00:00Z',
        note: paymentForm.note,
      })
      setPaymentDialogOpen(false)
      refresh()
    } catch {
      setPaymentError('เกิดข้อผิดพลาด กรุณาลองใหม่')
    } finally {
      setPaymentSaving(false)
    }
  }, [paymentForm, refresh])

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
        {canCreate && (
          <Button onClick={openCreate} className="gap-2">
            <Plus className="w-4 h-4" />
            <span className="hidden sm:inline">{t('bills.addBill')}</span>
          </Button>
        )}
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
                canUpdate={canUpdate}
                canDelete={canDelete}
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
        onFormChange={onFormChange}
        onContractChange={handleContractChange}
        onBillingMonthChange={handleBillingMonthChange}
        autoFilling={autoFilling}
        contracts={contracts}
        parkingSlots={tenantParkingSlots}
        selectedSlotIds={selectedSlotIds}
        onSlotToggle={handleSlotToggle}
        onSave={handleSave}
        saving={saving}
      />

      <PaymentDialog
        open={paymentDialogOpen}
        onOpenChange={setPaymentDialogOpen}
        editingItem={null}
        form={paymentForm}
        onFormChange={setPaymentForm}
        onSave={handlePaymentSave}
        saving={paymentSaving}
        error={paymentError}
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
