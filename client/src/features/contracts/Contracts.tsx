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
import { contractService } from '@/features/contracts/contractService'
import { tenantService } from '@/features/tenants/tenantService'
import { roomService } from '@/features/rooms/roomService'
import { ContractDialog } from '@/features/contracts/components/ContractDialog'
import { ContractDeleteDialog } from '@/features/contracts/components/ContractDeleteDialog'
import type { ApiContract, ApiTenant, ApiRoom, ContractStatus } from '@/types'
import { cn } from '@/lib/utils'
import { formatDate } from '@/lib/dateUtils'

const PER_PAGE_OPTIONS = [10, 20, 50] as const

const STATUS_VARIANTS: Record<ContractStatus, 'default' | 'secondary' | 'destructive'> = {
  active: 'default',
  expired: 'secondary',
  terminated: 'destructive',
}

function toDateInput(iso: string | null | undefined): string {
  if (!iso) return ''
  return iso.slice(0, 10)
}

const emptyForm = {
  tenant_id: '',
  room_id: '',
  start_date: '',
  end_date: '',
  rent_price: '',
  deposit: '',
  note: '',
  status: 'active' as ContractStatus,
}

export function Contracts() {
  const { t } = useTranslation()

  const [contracts, setContracts] = useState<ApiContract[]>([])
  const [loading, setLoading] = useState(true)
  const [error, setError] = useState('')
  const [search, setSearch] = useState('')
  const [filterStatus, setFilterStatus] = useState<string>('all')
  const [page, setPage] = useState(1)
  const [perPage, setPerPage] = useState<number>(20)
  const [total, setTotal] = useState(0)
  const [totalPages, setTotalPages] = useState(1)

  const [tenants, setTenants] = useState<ApiTenant[]>([])
  const [rooms, setRooms] = useState<ApiRoom[]>([])

  const [dialogOpen, setDialogOpen] = useState(false)
  const [deleteDialogOpen, setDeleteDialogOpen] = useState(false)
  const [editingContract, setEditingContract] = useState<ApiContract | null>(null)
  const [deletingContract, setDeletingContract] = useState<ApiContract | null>(null)
  const [form, setForm] = useState(emptyForm)
  const [saving, setSaving] = useState(false)
  const [saveError, setSaveError] = useState('')

  const fetchContracts = useCallback(async (p: number, pp: number) => {
    setLoading(true)
    setError('')
    try {
      const resp = await contractService.list(p, pp)
      setContracts(resp.data)
      setTotal(resp.meta.total)
      setTotalPages(resp.meta.total_pages)
    } catch {
      setError(t('contracts.loadError'))
    } finally {
      setLoading(false)
    }
  }, [t])

  useEffect(() => {
    fetchContracts(page, perPage)
  }, [fetchContracts, page, perPage])

  useEffect(() => {
    tenantService.list(1, 200).then((r) => setTenants(r.data)).catch(() => {})
    roomService.list(1, 200).then((r) => setRooms(r.data)).catch(() => {})
  }, [])

  function handlePerPageChange(value: string) {
    setPerPage(Number(value))
    setPage(1)
  }

  const filtered = useMemo(() => {
    const q = search.toLowerCase()
    return contracts.filter((c) => {
      const matchSearch =
        `${c.tenant_first_name} ${c.tenant_last_name}`.toLowerCase().includes(q) ||
        c.room_number.toLowerCase().includes(q)
      const matchStatus = filterStatus === 'all' || c.status === filterStatus
      return matchSearch && matchStatus
    })
  }, [contracts, search, filterStatus])

  const availableRooms = useMemo(() => {
    if (editingContract) return rooms
    return rooms.filter((r) => r.status === 'available')
  }, [rooms, editingContract])

  function openCreate() {
    setEditingContract(null)
    setForm(emptyForm)
    setSaveError('')
    setDialogOpen(true)
  }

  function openEdit(contract: ApiContract) {
    setSaveError('')
    setEditingContract(contract)
    setForm({
      tenant_id: contract.tenant_id,
      room_id: contract.room_id,
      start_date: toDateInput(contract.start_date),
      end_date: toDateInput(contract.end_date),
      rent_price: String(contract.rent_price),
      deposit: String(contract.deposit),
      note: contract.note,
      status: contract.status,
    })
    setDialogOpen(true)
  }

  async function handleSave() {
    if (!form.tenant_id || !form.room_id || !form.start_date || !form.rent_price) return
    setSaving(true)
    setSaveError('')
    try {
      if (editingContract) {
        await contractService.update(editingContract.id, {
          end_date: form.end_date ? `${form.end_date}T00:00:00Z` : null,
          rent_price: Number(form.rent_price),
          deposit: Number(form.deposit),
          status: form.status,
          note: form.note,
        })
      } else {
        await contractService.create({
          tenant_id: form.tenant_id,
          room_id: form.room_id,
          start_date: `${form.start_date}T00:00:00Z`,
          end_date: form.end_date ? `${form.end_date}T00:00:00Z` : null,
          rent_price: Number(form.rent_price),
          deposit: Number(form.deposit),
          note: form.note,
        })
      }
      roomService.list(1, 200).then((r) => setRooms(r.data)).catch(() => {})
      setDialogOpen(false)
      await fetchContracts(page, perPage)
    } catch (err: unknown) {
      const msg = (err as { response?: { data?: { message?: string } } })?.response?.data?.message
      setSaveError(msg || t('contracts.saveError'))
    } finally {
      setSaving(false)
    }
  }

  async function handleDelete() {
    if (!deletingContract) return
    try {
      await contractService.delete(deletingContract.id)
      const newPage = contracts.length === 1 && page > 1 ? page - 1 : page
      setPage(newPage)
      await fetchContracts(newPage, perPage)
      roomService.list(1, 200).then((r) => setRooms(r.data)).catch(() => {})
    } catch {
      // ignore
    } finally {
      setDeleteDialogOpen(false)
      setDeletingContract(null)
    }
  }

  return (
    <div className="space-y-6">
      <div className="flex flex-wrap items-start justify-between gap-3">
        <div className="min-w-0">
          <h1 className="text-2xl font-bold tracking-tight">{t('contracts.title')}</h1>
          <p className="text-muted-foreground text-sm mt-1">
            {t('contracts.subtitle', { count: filtered.length, total })}
          </p>
        </div>
        <div className="flex gap-2 shrink-0">
          <Button variant="outline" size="icon" onClick={() => fetchContracts(page, perPage)} disabled={loading}>
            <RefreshCw className={cn('w-4 h-4', loading && 'animate-spin')} />
          </Button>
          <Button onClick={openCreate} className="gap-2">
            <Plus className="w-4 h-4" />
            <span className="hidden sm:inline">{t('contracts.addContract')}</span>
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
              <CardTitle>{t('contracts.contractList')}</CardTitle>
              <CardDescription>{t('contracts.contractListDesc')}</CardDescription>
            </div>
            <div className="flex gap-2">
              <Select value={filterStatus} onValueChange={setFilterStatus}>
                <SelectTrigger className="h-9 w-36">
                  <SelectValue />
                </SelectTrigger>
                <SelectContent>
                  <SelectItem value="all">{t('contracts.allStatuses')}</SelectItem>
                  <SelectItem value="active">{t('contracts.statuses.active')}</SelectItem>
                  <SelectItem value="expired">{t('contracts.statuses.expired')}</SelectItem>
                  <SelectItem value="terminated">{t('contracts.statuses.terminated')}</SelectItem>
                </SelectContent>
              </Select>
              <div className="relative">
                <Search className="absolute left-3 top-1/2 -translate-y-1/2 w-4 h-4 text-muted-foreground" />
                <Input
                  placeholder={t('contracts.searchPlaceholder')}
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
                      <th className="text-left px-6 py-3 font-medium text-muted-foreground">{t('contracts.colTenant')}</th>
                      <th className="text-left px-4 py-3 font-medium text-muted-foreground">{t('contracts.colRoom')}</th>
                      <th className="text-left px-4 py-3 font-medium text-muted-foreground">{t('contracts.colStartDate')}</th>
                      <th className="text-left px-4 py-3 font-medium text-muted-foreground">{t('contracts.colEndDate')}</th>
                      <th className="text-right px-4 py-3 font-medium text-muted-foreground">{t('contracts.colRentPrice')}</th>
                      <th className="text-right px-4 py-3 font-medium text-muted-foreground">{t('contracts.colDeposit')}</th>
                      <th className="text-center px-4 py-3 font-medium text-muted-foreground">{t('contracts.colStatus')}</th>
                      <th className="text-right px-6 py-3 font-medium text-muted-foreground">{t('contracts.colActions')}</th>
                    </tr>
                  </thead>
                  <tbody>
                    {filtered.map((contract, i) => (
                      <tr
                        key={contract.id}
                        className={cn(
                          'border-b transition-colors hover:bg-muted/30',
                          i === filtered.length - 1 && 'border-0'
                        )}
                      >
                        <td className="px-6 py-4 font-medium">
                          {contract.tenant_first_name} {contract.tenant_last_name}
                        </td>
                        <td className="px-4 py-4 text-muted-foreground">{contract.room_number}</td>
                        <td className="px-4 py-4 text-muted-foreground">{formatDate(contract.start_date)}</td>
                        <td className="px-4 py-4 text-muted-foreground">
                          {contract.end_date ? formatDate(contract.end_date) : (
                            <span className="text-xs text-muted-foreground/60">{t('contracts.openEnded')}</span>
                          )}
                        </td>
                        <td className="px-4 py-4 text-right">
                          {contract.rent_price.toLocaleString()}
                          <span className="text-xs text-muted-foreground ml-1">{t('contracts.bahtPerMonth')}</span>
                        </td>
                        <td className="px-4 py-4 text-right text-muted-foreground">
                          {contract.deposit.toLocaleString()}
                        </td>
                        <td className="px-4 py-4 text-center">
                          <Badge variant={STATUS_VARIANTS[contract.status]}>
                            {t(`contracts.statuses.${contract.status}`)}
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
                              <DropdownMenuItem onClick={() => openEdit(contract)} className="gap-2">
                                <Pencil className="w-4 h-4" /> {t('common.edit')}
                              </DropdownMenuItem>
                              <DropdownMenuSeparator />
                              <DropdownMenuItem
                                className="gap-2 text-destructive focus:text-destructive"
                                onClick={() => { setDeletingContract(contract); setDeleteDialogOpen(true) }}
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
                    <p className="text-sm font-medium">{t('contracts.noContracts')}</p>
                    {(search || filterStatus !== 'all') && (
                      <button
                        type="button"
                        onClick={() => { setSearch(''); setFilterStatus('all') }}
                        className="text-xs text-primary hover:underline"
                      >
                        {t('contracts.clearFilters')}
                      </button>
                    )}
                  </div>
                )}
              </div>

              {/* Mobile cards */}
              <div className="md:hidden divide-y">
                {filtered.map((contract) => (
                  <div key={contract.id} className="p-4 flex items-start gap-3">
                    <div className="flex-1 min-w-0 space-y-1">
                      <p className="text-sm font-medium">
                        {contract.tenant_first_name} {contract.tenant_last_name}
                      </p>
                      <p className="text-xs text-muted-foreground">ห้อง {contract.room_number}</p>
                      <p className="text-xs text-muted-foreground">
                        {formatDate(contract.start_date)} →{' '}
                        {contract.end_date ? formatDate(contract.end_date) : t('contracts.openEnded')}
                      </p>
                      <Badge variant={STATUS_VARIANTS[contract.status]} className="text-xs">
                        {t(`contracts.statuses.${contract.status}`)}
                      </Badge>
                    </div>
                    <DropdownMenu>
                      <DropdownMenuTrigger asChild>
                        <Button variant="ghost" size="icon" className="h-8 w-8 shrink-0">
                          <MoreHorizontal className="w-4 h-4" />
                        </Button>
                      </DropdownMenuTrigger>
                      <DropdownMenuContent align="end">
                        <DropdownMenuItem onClick={() => openEdit(contract)}>{t('common.edit')}</DropdownMenuItem>
                        <DropdownMenuSeparator />
                        <DropdownMenuItem
                          className="text-destructive focus:text-destructive"
                          onClick={() => { setDeletingContract(contract); setDeleteDialogOpen(true) }}
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
                    {t('contracts.showing', {
                      from: (page - 1) * perPage + 1,
                      to: Math.min(page * perPage, total),
                      total,
                    })}
                  </span>
                  <div className="flex items-center gap-4">
                    <div className="flex items-center gap-2">
                      <span className="hidden sm:inline shrink-0">{t('contracts.perPage')}</span>
                      <Select value={String(perPage)} onValueChange={handlePerPageChange}>
                        <SelectTrigger className="h-8 w-16"><SelectValue /></SelectTrigger>
                        <SelectContent>
                          {PER_PAGE_OPTIONS.map((n) => (
                            <SelectItem key={n} value={String(n)}>{n}</SelectItem>
                          ))}
                        </SelectContent>
                      </Select>
                    </div>
                    <span className="shrink-0">{t('contracts.page')} {page} {t('contracts.of')} {totalPages}</span>
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

      <ContractDialog
        open={dialogOpen}
        onOpenChange={setDialogOpen}
        editingContract={editingContract}
        form={form}
        onFormChange={setForm}
        onSave={handleSave}
        saving={saving}
        tenants={tenants}
        availableRooms={availableRooms}
        saveError={saveError}
      />

      <ContractDeleteDialog
        open={deleteDialogOpen}
        onOpenChange={setDeleteDialogOpen}
        contract={deletingContract}
        onDelete={handleDelete}
      />
    </div>
  )
}
