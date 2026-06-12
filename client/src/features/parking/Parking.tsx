import { useState, useMemo, useEffect, useCallback, useRef } from 'react'
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
  ChevronDown,
  X,
  User,
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
import { Popover, PopoverContent, PopoverTrigger } from '@/components/ui/popover'
import { DatePicker } from '@/components/ui/date-picker'
import { parkingService } from '@/features/parking/parkingService'
import { tenantService } from '@/features/tenants/tenantService'
import type { ApiParkingSlot, ApiTenant, ParkingStatus, VehicleType } from '@/types'
import { cn } from '@/lib/utils'
import { formatDate } from '@/lib/dateUtils'

const PER_PAGE_OPTIONS = [10, 20, 50] as const

function toDateInput(val: string | null | undefined): string {
  if (!val) return ''
  return val.slice(0, 10)
}

const emptyForm = {
  slot_number: '',
  vehicle_type: 'car' as VehicleType,
  status: 'available' as ParkingStatus,
  tenant_id: '' as string,
  license_plate: '',
  monthly_fee: '' as string | number,
  start_date: '',
  note: '',
}

export function Parking() {
  const { t } = useTranslation()

  const [items, setItems] = useState<ApiParkingSlot[]>([])
  const [loading, setLoading] = useState(true)
  const [error, setError] = useState('')
  const [search, setSearch] = useState('')
  const [filterStatus, setFilterStatus] = useState<string>('all')
  const [filterVehicle, setFilterVehicle] = useState<string>('all')
  const [page, setPage] = useState(1)
  const [perPage, setPerPage] = useState<number>(20)
  const [total, setTotal] = useState(0)
  const [totalPages, setTotalPages] = useState(1)

  const [dialogOpen, setDialogOpen] = useState(false)
  const [deleteDialogOpen, setDeleteDialogOpen] = useState(false)
  const [editingItem, setEditingItem] = useState<ApiParkingSlot | null>(null)
  const [deletingItem, setDeletingItem] = useState<ApiParkingSlot | null>(null)
  const [form, setForm] = useState(emptyForm)
  const [saving, setSaving] = useState(false)

  const [tenants, setTenants] = useState<ApiTenant[]>([])
  const [tenantPickerOpen, setTenantPickerOpen] = useState(false)
  const [tenantSearch, setTenantSearch] = useState('')
  const tenantSearchRef = useRef<HTMLInputElement>(null)

  const fetchItems = useCallback(
    async (p: number, pp: number) => {
      setLoading(true)
      setError('')
      try {
        const resp = await parkingService.list(p, pp)
        setItems(resp.data)
        setTotal(resp.meta.total)
        setTotalPages(resp.meta.total_pages)
      } catch {
        setError(t('parking.loadError'))
      } finally {
        setLoading(false)
      }
    },
    [t]
  )

  useEffect(() => {
    fetchItems(page, perPage)
  }, [fetchItems, page, perPage])

  useEffect(() => {
    if (!dialogOpen) return
    tenantService.list(1, 200).then((r) => setTenants(r.data)).catch(() => {})
  }, [dialogOpen])

  function handlePerPageChange(value: string) {
    setPerPage(Number(value))
    setPage(1)
  }

  const filtered = useMemo(() => {
    const q = search.toLowerCase()
    return items.filter((item) => {
      const tenantName = `${item.tenant_first_name} ${item.tenant_last_name}`.toLowerCase()
      const matchSearch =
        item.slot_number.toLowerCase().includes(q) ||
        item.license_plate.toLowerCase().includes(q) ||
        tenantName.includes(q)
      const matchStatus = filterStatus === 'all' || item.status === filterStatus
      const matchVehicle = filterVehicle === 'all' || item.vehicle_type === filterVehicle
      return matchSearch && matchStatus && matchVehicle
    })
  }, [items, search, filterStatus, filterVehicle])

  function openCreate() {
    setEditingItem(null)
    setForm(emptyForm)
    setTenantSearch('')
    setDialogOpen(true)
  }

  function openEdit(item: ApiParkingSlot) {
    setEditingItem(item)
    setForm({
      slot_number: item.slot_number,
      vehicle_type: item.vehicle_type,
      status: item.status,
      tenant_id: item.tenant_id ?? '',
      license_plate: item.license_plate,
      monthly_fee: item.monthly_fee > 0 ? item.monthly_fee : '',
      start_date: toDateInput(item.start_date),
      note: item.note,
    })
    setTenantSearch('')
    setDialogOpen(true)
  }

  function handleStatusChange(v: string) {
    const newStatus = v as ParkingStatus
    setForm((f) => ({
      ...f,
      status: newStatus,
      ...(newStatus === 'available'
        ? { tenant_id: '', license_plate: '', start_date: '' }
        : {}),
    }))
  }

  const selectedTenant = tenants.find((t) => t.id === form.tenant_id) ?? null

  const filteredTenants = useMemo(() => {
    const q = tenantSearch.toLowerCase()
    if (!q) return tenants
    return tenants.filter(
      (t) =>
        t.first_name.toLowerCase().includes(q) ||
        t.last_name.toLowerCase().includes(q) ||
        `${t.first_name} ${t.last_name}`.toLowerCase().includes(q)
    )
  }, [tenants, tenantSearch])

  async function handleSave() {
    if (!form.slot_number) return
    setSaving(true)
    try {
      const payload = {
        slot_number: form.slot_number,
        vehicle_type: form.vehicle_type,
        status: form.status,
        tenant_id: form.status === 'occupied' ? form.tenant_id || null : null,
        license_plate: form.status === 'occupied' ? form.license_plate : '',
        monthly_fee: form.monthly_fee === '' ? 0 : Number(form.monthly_fee),
        start_date: form.status === 'occupied' ? form.start_date || null : null,
        note: form.note,
      }
      if (editingItem) {
        await parkingService.update(editingItem.id, payload)
      } else {
        await parkingService.create(payload)
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
      await parkingService.delete(deletingItem.id)
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

  const isSaveDisabled = saving || !form.slot_number

  return (
    <div className="space-y-6">
      <div className="flex flex-wrap items-start justify-between gap-3">
        <div className="min-w-0">
          <h1 className="text-2xl font-bold tracking-tight">{t('parking.title')}</h1>
          <p className="text-muted-foreground text-sm mt-1">
            {t('parking.subtitle', { count: filtered.length, total })}
          </p>
        </div>
        <div className="flex gap-2 shrink-0">
          <Button variant="outline" size="icon" onClick={() => fetchItems(page, perPage)} disabled={loading}>
            <RefreshCw className={cn('w-4 h-4', loading && 'animate-spin')} />
          </Button>
          <Button onClick={openCreate} className="gap-2">
            <Plus className="w-4 h-4" />
            <span className="hidden sm:inline">{t('parking.addSlot')}</span>
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
              <CardTitle>{t('parking.slotList')}</CardTitle>
              <CardDescription>{t('parking.slotListDesc')}</CardDescription>
            </div>
            <div className="flex flex-wrap gap-2">
              <Select value={filterVehicle} onValueChange={setFilterVehicle}>
                <SelectTrigger className="h-9 w-36">
                  <SelectValue />
                </SelectTrigger>
                <SelectContent>
                  <SelectItem value="all">{t('parking.allVehicles')}</SelectItem>
                  <SelectItem value="car">{t('parking.vehicleTypes.car')}</SelectItem>
                  <SelectItem value="motorcycle">{t('parking.vehicleTypes.motorcycle')}</SelectItem>
                </SelectContent>
              </Select>
              <Select value={filterStatus} onValueChange={setFilterStatus}>
                <SelectTrigger className="h-9 w-36">
                  <SelectValue />
                </SelectTrigger>
                <SelectContent>
                  <SelectItem value="all">{t('parking.allStatuses')}</SelectItem>
                  <SelectItem value="available">{t('parking.statuses.available')}</SelectItem>
                  <SelectItem value="occupied">{t('parking.statuses.occupied')}</SelectItem>
                </SelectContent>
              </Select>
              <div className="relative">
                <Search className="absolute left-3 top-1/2 -translate-y-1/2 w-4 h-4 text-muted-foreground" />
                <Input
                  placeholder={t('parking.searchPlaceholder')}
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
                      <th className="text-left px-6 py-3 font-medium text-muted-foreground">{t('parking.colSlot')}</th>
                      <th className="text-left px-4 py-3 font-medium text-muted-foreground">{t('parking.colVehicle')}</th>
                      <th className="text-left px-4 py-3 font-medium text-muted-foreground">{t('parking.colStatus')}</th>
                      <th className="text-left px-4 py-3 font-medium text-muted-foreground">{t('parking.colTenant')}</th>
                      <th className="text-left px-4 py-3 font-medium text-muted-foreground">{t('parking.colLicense')}</th>
                      <th className="text-left px-4 py-3 font-medium text-muted-foreground">{t('parking.colFee')}</th>
                      <th className="text-left px-4 py-3 font-medium text-muted-foreground">{t('parking.colStartDate')}</th>
                      <th className="text-right px-6 py-3 font-medium text-muted-foreground">{t('parking.colActions')}</th>
                    </tr>
                  </thead>
                  <tbody>
                    {filtered.map((item, i) => (
                      <tr
                        key={item.id}
                        className={cn(
                          'border-b transition-colors hover:bg-muted/30',
                          i === filtered.length - 1 && 'border-0'
                        )}
                      >
                        <td className="px-6 py-4 font-medium">{item.slot_number}</td>
                        <td className="px-4 py-4">
                          <Badge variant="outline">
                            {t(`parking.vehicleTypes.${item.vehicle_type}`)}
                          </Badge>
                        </td>
                        <td className="px-4 py-4">
                          <Badge variant={item.status === 'available' ? 'secondary' : 'default'}>
                            {t(`parking.statuses.${item.status}`)}
                          </Badge>
                        </td>
                        <td className="px-4 py-4 text-muted-foreground">
                          {item.tenant_first_name
                            ? `${item.tenant_first_name} ${item.tenant_last_name}`
                            : '-'}
                        </td>
                        <td className="px-4 py-4 text-muted-foreground">{item.license_plate || '-'}</td>
                        <td className="px-4 py-4 text-muted-foreground">
                          {item.monthly_fee > 0 ? `฿${item.monthly_fee.toLocaleString()}` : '-'}
                        </td>
                        <td className="px-4 py-4 text-muted-foreground">
                          {item.start_date ? formatDate(item.start_date) : '-'}
                        </td>
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
                    ))}
                  </tbody>
                </table>

                {filtered.length === 0 && (
                  <div className="flex flex-col items-center justify-center py-16 gap-3 text-muted-foreground">
                    <div className="p-4 rounded-full bg-muted">
                      <SearchX className="w-6 h-6" />
                    </div>
                    <p className="text-sm font-medium">{t('parking.noSlots')}</p>
                    {(search || filterStatus !== 'all' || filterVehicle !== 'all') && (
                      <button
                        type="button"
                        onClick={() => { setSearch(''); setFilterStatus('all'); setFilterVehicle('all') }}
                        className="text-xs text-primary hover:underline"
                      >
                        {t('parking.clearFilters')}
                      </button>
                    )}
                  </div>
                )}
              </div>

              {/* Mobile cards */}
              <div className="md:hidden divide-y">
                {filtered.map((item) => (
                  <div key={item.id} className="p-4 flex items-start gap-3">
                    <div className="flex-1 min-w-0 space-y-1">
                      <div className="flex items-center gap-2">
                        <p className="text-sm font-medium">{item.slot_number}</p>
                        <Badge variant={item.status === 'available' ? 'secondary' : 'default'} className="text-xs">
                          {t(`parking.statuses.${item.status}`)}
                        </Badge>
                      </div>
                      <p className="text-xs text-muted-foreground">
                        {t(`parking.vehicleTypes.${item.vehicle_type}`)}
                        {item.license_plate ? ` · ${item.license_plate}` : ''}
                      </p>
                      {item.tenant_first_name && (
                        <p className="text-xs text-muted-foreground">
                          {item.tenant_first_name} {item.tenant_last_name}
                        </p>
                      )}
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
                ))}
              </div>

              {/* Pagination */}
              {!loading && total > 0 && (
                <div className="flex flex-col sm:flex-row items-center justify-between gap-3 px-4 py-3 border-t text-sm text-muted-foreground">
                  <span className="shrink-0">
                    {t('parking.showing', {
                      from: (page - 1) * perPage + 1,
                      to: Math.min(page * perPage, total),
                      total,
                    })}
                  </span>
                  <div className="flex items-center gap-4">
                    <div className="flex items-center gap-2">
                      <span className="hidden sm:inline shrink-0">{t('parking.perPage')}</span>
                      <Select value={String(perPage)} onValueChange={handlePerPageChange}>
                        <SelectTrigger className="h-8 w-16"><SelectValue /></SelectTrigger>
                        <SelectContent>
                          {PER_PAGE_OPTIONS.map((n) => (
                            <SelectItem key={n} value={String(n)}>{n}</SelectItem>
                          ))}
                        </SelectContent>
                      </Select>
                    </div>
                    <span className="shrink-0">{t('parking.page')} {page} {t('parking.of')} {totalPages}</span>
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
              {editingItem ? t('parking.editSlot') : t('parking.createSlot')}
            </DialogTitle>
            <DialogDescription>
              {editingItem ? t('parking.editDesc') : t('parking.createDesc')}
            </DialogDescription>
          </DialogHeader>

          <div className="grid gap-4 py-4">
            {/* ข้อมูลช่องจอด */}
            <div className="grid grid-cols-2 gap-4">
              <div className="space-y-1.5">
                <Label>{t('parking.slotNumber')} *</Label>
                <Input
                  placeholder={t('parking.slotNumberPlaceholder')}
                  value={form.slot_number}
                  onChange={(e) => setForm((f) => ({ ...f, slot_number: e.target.value }))}
                />
              </div>
              <div className="space-y-1.5">
                <Label>{t('parking.vehicleType')}</Label>
                <Select
                  value={form.vehicle_type}
                  onValueChange={(v) => setForm((f) => ({ ...f, vehicle_type: v as VehicleType }))}
                >
                  <SelectTrigger><SelectValue /></SelectTrigger>
                  <SelectContent>
                    <SelectItem value="car">{t('parking.vehicleTypes.car')}</SelectItem>
                    <SelectItem value="motorcycle">{t('parking.vehicleTypes.motorcycle')}</SelectItem>
                  </SelectContent>
                </Select>
              </div>
            </div>

            <div className="grid grid-cols-2 gap-4">
              <div className="space-y-1.5">
                <Label>{t('parking.status')}</Label>
                <Select value={form.status} onValueChange={handleStatusChange}>
                  <SelectTrigger><SelectValue /></SelectTrigger>
                  <SelectContent>
                    <SelectItem value="available">{t('parking.statuses.available')}</SelectItem>
                    <SelectItem value="occupied">{t('parking.statuses.occupied')}</SelectItem>
                  </SelectContent>
                </Select>
              </div>
              <div className="space-y-1.5">
                <Label>{t('parking.monthlyFee')}</Label>
                <Input
                  type="number"
                  min={0}
                  placeholder={t('parking.monthlyFeePlaceholder')}
                  value={form.monthly_fee}
                  onChange={(e) => setForm((f) => ({ ...f, monthly_fee: e.target.value }))}
                />
              </div>
            </div>

            <div className="space-y-1.5">
              <Label>{t('parking.note')}</Label>
              <Input
                placeholder={t('parking.notePlaceholder')}
                value={form.note}
                onChange={(e) => setForm((f) => ({ ...f, note: e.target.value }))}
              />
            </div>

            {/* ข้อมูลผู้ใช้ช่องจอด (แสดงเฉพาะเมื่อ occupied) */}
            {form.status === 'occupied' && (
              <div className="space-y-4 rounded-lg border bg-muted/30 p-4">
                <p className="text-xs font-medium text-muted-foreground uppercase tracking-wide">
                  {t('parking.occupancySection')}
                </p>

                {/* Tenant picker */}
                <div className="space-y-1.5">
                  <Label>{t('parking.tenant')}</Label>
                  <Popover
                    open={tenantPickerOpen}
                    onOpenChange={(open) => {
                      setTenantPickerOpen(open)
                      if (open) {
                        setTenantSearch('')
                        setTimeout(() => tenantSearchRef.current?.focus(), 50)
                      }
                    }}
                  >
                    <PopoverTrigger asChild>
                      <button
                        type="button"
                        className={cn(
                          'flex w-full items-center justify-between rounded-md border bg-background px-3 py-2 text-sm ring-offset-background',
                          'hover:bg-accent focus:outline-none focus:ring-2 focus:ring-ring focus:ring-offset-2',
                          !selectedTenant && 'text-muted-foreground'
                        )}
                      >
                        <span className="flex items-center gap-2 min-w-0">
                          <User className="w-4 h-4 shrink-0 text-muted-foreground" />
                          <span className="truncate">
                            {selectedTenant
                              ? `${selectedTenant.first_name} ${selectedTenant.last_name}`
                              : t('parking.selectTenantPlaceholder')}
                          </span>
                        </span>
                        <span className="flex items-center gap-1 shrink-0 ml-2">
                          {selectedTenant && (
                            <span
                              role="button"
                              tabIndex={0}
                              className="rounded p-0.5 hover:bg-muted"
                              onClick={(e) => {
                                e.stopPropagation()
                                setForm((f) => ({ ...f, tenant_id: '' }))
                              }}
                              onKeyDown={(e) => {
                                if (e.key === 'Enter' || e.key === ' ') {
                                  e.stopPropagation()
                                  setForm((f) => ({ ...f, tenant_id: '' }))
                                }
                              }}
                            >
                              <X className="w-3.5 h-3.5 text-muted-foreground" />
                            </span>
                          )}
                          <ChevronDown className="w-4 h-4 text-muted-foreground" />
                        </span>
                      </button>
                    </PopoverTrigger>
                    <PopoverContent align="start" className="w-72 p-0">
                      <div className="p-2 border-b">
                        <Input
                          ref={tenantSearchRef}
                          placeholder={t('parking.searchTenant')}
                          value={tenantSearch}
                          onChange={(e) => setTenantSearch(e.target.value)}
                          className="h-8"
                        />
                      </div>
                      <div className="max-h-48 overflow-y-auto py-1">
                        {filteredTenants.length === 0 ? (
                          <p className="px-3 py-2 text-sm text-muted-foreground">
                            {t('parking.noTenantFound')}
                          </p>
                        ) : (
                          filteredTenants.map((tenant) => (
                            <button
                              key={tenant.id}
                              type="button"
                              className={cn(
                                'flex w-full items-center gap-2 px-3 py-2 text-sm hover:bg-accent text-left',
                                form.tenant_id === tenant.id && 'bg-accent font-medium'
                              )}
                              onClick={() => {
                                setForm((f) => ({ ...f, tenant_id: tenant.id }))
                                setTenantPickerOpen(false)
                              }}
                            >
                              <User className="w-4 h-4 shrink-0 text-muted-foreground" />
                              {tenant.first_name} {tenant.last_name}
                            </button>
                          ))
                        )}
                      </div>
                    </PopoverContent>
                  </Popover>
                </div>

                <div className="grid grid-cols-2 gap-4">
                  <div className="space-y-1.5">
                    <Label>{t('parking.licensePlate')}</Label>
                    <Input
                      placeholder={t('parking.licensePlatePlaceholder')}
                      value={form.license_plate}
                      onChange={(e) => setForm((f) => ({ ...f, license_plate: e.target.value }))}
                    />
                  </div>
                  <div className="space-y-1.5">
                    <Label>{t('parking.startDate')}</Label>
                    <DatePicker
                      value={form.start_date}
                      onChange={(v) => setForm((f) => ({ ...f, start_date: v }))}
                    />
                  </div>
                </div>
              </div>
            )}
          </div>

          <DialogFooter>
            <Button variant="outline" onClick={() => setDialogOpen(false)}>
              {t('common.cancel')}
            </Button>
            <Button onClick={handleSave} disabled={isSaveDisabled}>
              {saving
                ? t('common.loading')
                : editingItem
                ? t('parking.saveChanges')
                : t('parking.createSlot')}
            </Button>
          </DialogFooter>
        </DialogContent>
      </Dialog>

      {/* Delete confirm */}
      <Dialog open={deleteDialogOpen} onOpenChange={setDeleteDialogOpen}>
        <DialogContent className="max-w-sm">
          <DialogHeader>
            <DialogTitle>{t('parking.deleteSlot')}</DialogTitle>
            <DialogDescription>
              {t('parking.deleteConfirm')}{' '}
              <span className="font-semibold text-foreground">{deletingItem?.slot_number}</span>?{' '}
              {t('parking.deleteWarning')}
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
