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
  PlayCircle,
  CheckCircle2,
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
import { maintenanceService } from '@/features/maintenance/maintenanceService'
import { MaintenanceDialog } from '@/features/maintenance/components/MaintenanceDialog'
import { MaintenanceDeleteDialog } from '@/features/maintenance/components/MaintenanceDeleteDialog'
import type { ApiMaintenanceRequest, MaintenanceStatus, MaintenancePriority } from '@/types'
import { cn } from '@/lib/utils'
import { formatDate } from '@/lib/dateUtils'

const PER_PAGE_OPTIONS = [10, 20, 50] as const

const STATUS_VARIANTS: Record<MaintenanceStatus, 'default' | 'secondary' | 'destructive' | 'outline'> = {
  open: 'destructive',
  in_progress: 'default',
  done: 'secondary',
  cancelled: 'outline',
}

const PRIORITY_VARIANTS: Record<MaintenancePriority, 'default' | 'secondary' | 'destructive' | 'outline'> = {
  urgent: 'destructive',
  high: 'default',
  normal: 'secondary',
  low: 'outline',
}

function toDateInput(iso: string | null | undefined): string {
  if (!iso) return ''
  return iso.slice(0, 10)
}

const emptyForm = {
  room_id: '',
  title: '',
  description: '',
  status: 'open' as MaintenanceStatus,
  priority: 'normal' as MaintenancePriority,
  reported_date: '',
  resolved_date: '',
  note: '',
}

export function MaintenanceRequests() {
  const { t } = useTranslation()

  const [items, setItems] = useState<ApiMaintenanceRequest[]>([])
  const [loading, setLoading] = useState(true)
  const [error, setError] = useState('')
  const [search, setSearch] = useState('')
  const [filterStatus, setFilterStatus] = useState<string>('all')
  const [filterPriority, setFilterPriority] = useState<string>('all')
  const [page, setPage] = useState(1)
  const [perPage, setPerPage] = useState<number>(20)
  const [total, setTotal] = useState(0)
  const [totalPages, setTotalPages] = useState(1)

  const [dialogOpen, setDialogOpen] = useState(false)
  const [deleteDialogOpen, setDeleteDialogOpen] = useState(false)
  const [editingItem, setEditingItem] = useState<ApiMaintenanceRequest | null>(null)
  const [deletingItem, setDeletingItem] = useState<ApiMaintenanceRequest | null>(null)
  const [form, setForm] = useState(emptyForm)
  const [saving, setSaving] = useState(false)
  const [saveError, setSaveError] = useState('')

  const fetchItems = useCallback(async (p: number, pp: number) => {
    setLoading(true)
    setError('')
    try {
      const resp = await maintenanceService.list(p, pp)
      setItems(resp.data)
      setTotal(resp.meta.total)
      setTotalPages(resp.meta.total_pages)
    } catch {
      setError(t('maintenance.loadError'))
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
      const matchSearch =
        item.title.toLowerCase().includes(q) ||
        item.room_number.toLowerCase().includes(q) ||
        item.description.toLowerCase().includes(q)
      const matchStatus = filterStatus === 'all' || item.status === filterStatus
      const matchPriority = filterPriority === 'all' || item.priority === filterPriority
      return matchSearch && matchStatus && matchPriority
    })
  }, [items, search, filterStatus, filterPriority])

  function openCreate() {
    setEditingItem(null)
    setSaveError('')
    setForm({ ...emptyForm, reported_date: new Date().toISOString().slice(0, 10) })
    setDialogOpen(true)
  }

  function openEdit(item: ApiMaintenanceRequest) {
    setSaveError('')
    setEditingItem(item)
    setForm({
      room_id: item.room_id,
      title: item.title,
      description: item.description,
      status: item.status,
      priority: item.priority,
      reported_date: toDateInput(item.reported_date),
      resolved_date: toDateInput(item.resolved_date),
      note: item.note,
    })
    setDialogOpen(true)
  }

  async function handleSave() {
    if (!form.room_id || !form.title || !form.reported_date) return
    setSaving(true)
    setSaveError('')
    try {
      const payload = {
        room_id: form.room_id,
        title: form.title,
        description: form.description,
        status: form.status,
        priority: form.priority,
        reported_date: form.reported_date + 'T00:00:00Z',
        resolved_date: form.resolved_date ? form.resolved_date + 'T00:00:00Z' : null,
        note: form.note,
      }
      if (editingItem) {
        await maintenanceService.update(editingItem.id, payload)
      } else {
        await maintenanceService.create(payload)
      }
      setDialogOpen(false)
      await fetchItems(page, perPage)
    } catch {
      setSaveError(t('maintenance.saveError'))
    } finally {
      setSaving(false)
    }
  }

  async function handleQuickStatus(item: ApiMaintenanceRequest, status: MaintenanceStatus) {
    try {
      await maintenanceService.update(item.id, {
        room_id: item.room_id,
        title: item.title,
        description: item.description,
        status,
        priority: item.priority,
        reported_date: item.reported_date,
        resolved_date:
          status === 'done'
            ? (item.resolved_date ?? new Date().toISOString())
            : item.resolved_date ?? null,
        note: item.note,
      })
      await fetchItems(page, perPage)
    } catch {
      // ignore
    }
  }

  async function handleDelete() {
    if (!deletingItem) return
    try {
      await maintenanceService.delete(deletingItem.id)
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
          <h1 className="text-2xl font-bold tracking-tight">{t('maintenance.title')}</h1>
          <p className="text-muted-foreground text-sm mt-1">
            {t('maintenance.subtitle', { count: filtered.length, total })}
          </p>
        </div>
        <div className="flex gap-2 shrink-0">
          <Button variant="outline" size="icon" onClick={() => fetchItems(page, perPage)} disabled={loading}>
            <RefreshCw className={cn('w-4 h-4', loading && 'animate-spin')} />
          </Button>
          <Button onClick={openCreate} className="gap-2">
            <Plus className="w-4 h-4" />
            <span className="hidden sm:inline">{t('maintenance.addRequest')}</span>
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
              <CardTitle>{t('maintenance.requestList')}</CardTitle>
              <CardDescription>{t('maintenance.requestListDesc')}</CardDescription>
            </div>
            <div className="flex flex-wrap gap-2">
              <Select value={filterStatus} onValueChange={setFilterStatus}>
                <SelectTrigger className="h-9 w-36">
                  <SelectValue />
                </SelectTrigger>
                <SelectContent>
                  <SelectItem value="all">{t('maintenance.allStatuses')}</SelectItem>
                  <SelectItem value="open">{t('maintenance.statuses.open')}</SelectItem>
                  <SelectItem value="in_progress">{t('maintenance.statuses.in_progress')}</SelectItem>
                  <SelectItem value="done">{t('maintenance.statuses.done')}</SelectItem>
                  <SelectItem value="cancelled">{t('maintenance.statuses.cancelled')}</SelectItem>
                </SelectContent>
              </Select>
              <Select value={filterPriority} onValueChange={setFilterPriority}>
                <SelectTrigger className="h-9 w-36">
                  <SelectValue />
                </SelectTrigger>
                <SelectContent>
                  <SelectItem value="all">{t('maintenance.allPriorities')}</SelectItem>
                  <SelectItem value="urgent">{t('maintenance.priorities.urgent')}</SelectItem>
                  <SelectItem value="high">{t('maintenance.priorities.high')}</SelectItem>
                  <SelectItem value="normal">{t('maintenance.priorities.normal')}</SelectItem>
                  <SelectItem value="low">{t('maintenance.priorities.low')}</SelectItem>
                </SelectContent>
              </Select>
              <div className="relative">
                <Search className="absolute left-3 top-1/2 -translate-y-1/2 w-4 h-4 text-muted-foreground" />
                <Input
                  placeholder={t('maintenance.searchPlaceholder')}
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
                      <th className="text-left px-6 py-3 font-medium text-muted-foreground">{t('maintenance.colRoom')}</th>
                      <th className="text-left px-4 py-3 font-medium text-muted-foreground">{t('maintenance.colTitle')}</th>
                      <th className="text-left px-4 py-3 font-medium text-muted-foreground">{t('maintenance.colStatus')}</th>
                      <th className="text-left px-4 py-3 font-medium text-muted-foreground">{t('maintenance.colPriority')}</th>
                      <th className="text-left px-4 py-3 font-medium text-muted-foreground">{t('maintenance.colDate')}</th>
                      <th className="text-right px-6 py-3 font-medium text-muted-foreground">{t('maintenance.colActions')}</th>
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
                        <td className="px-6 py-4 font-medium">{item.room_number}</td>
                        <td className="px-4 py-4">
                          <p className="font-medium">{item.title}</p>
                          {item.description && (
                            <p className="text-xs text-muted-foreground mt-0.5 truncate max-w-xs">{item.description}</p>
                          )}
                        </td>
                        <td className="px-4 py-4">
                          <Badge variant={STATUS_VARIANTS[item.status]}>
                            {t(`maintenance.statuses.${item.status}`)}
                          </Badge>
                        </td>
                        <td className="px-4 py-4">
                          <Badge variant={PRIORITY_VARIANTS[item.priority]}>
                            {t(`maintenance.priorities.${item.priority}`)}
                          </Badge>
                        </td>
                        <td className="px-4 py-4 text-muted-foreground">{formatDate(item.reported_date)}</td>
                        <td className="px-6 py-4 text-right">
                          <DropdownMenu>
                            <DropdownMenuTrigger asChild>
                              <Button variant="ghost" size="icon" className="h-8 w-8">
                                <MoreHorizontal className="w-4 h-4" />
                              </Button>
                            </DropdownMenuTrigger>
                            <DropdownMenuContent align="end">
                              {item.status === 'open' && (
                                <DropdownMenuItem onClick={() => handleQuickStatus(item, 'in_progress')} className="gap-2">
                                  <PlayCircle className="w-4 h-4" /> {t('maintenance.startProgress')}
                                </DropdownMenuItem>
                              )}
                              {(item.status === 'open' || item.status === 'in_progress') && (
                                <DropdownMenuItem onClick={() => handleQuickStatus(item, 'done')} className="gap-2">
                                  <CheckCircle2 className="w-4 h-4" /> {t('maintenance.markDone')}
                                </DropdownMenuItem>
                              )}
                              {(item.status === 'open' || item.status === 'in_progress') && (
                                <DropdownMenuSeparator />
                              )}
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
                    <p className="text-sm font-medium">{t('maintenance.noRequests')}</p>
                    {(search || filterStatus !== 'all' || filterPriority !== 'all') && (
                      <button
                        type="button"
                        onClick={() => { setSearch(''); setFilterStatus('all'); setFilterPriority('all') }}
                        className="text-xs text-primary hover:underline"
                      >
                        {t('maintenance.clearFilters')}
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
                      <p className="text-sm font-medium">{item.title}</p>
                      <p className="text-xs text-muted-foreground">{t('maintenance.room')} {item.room_number} · {formatDate(item.reported_date)}</p>
                      <div className="flex gap-1.5 flex-wrap">
                        <Badge variant={STATUS_VARIANTS[item.status]} className="text-xs">
                          {t(`maintenance.statuses.${item.status}`)}
                        </Badge>
                        <Badge variant={PRIORITY_VARIANTS[item.priority]} className="text-xs">
                          {t(`maintenance.priorities.${item.priority}`)}
                        </Badge>
                      </div>
                    </div>
                    <DropdownMenu>
                      <DropdownMenuTrigger asChild>
                        <Button variant="ghost" size="icon" className="h-8 w-8 shrink-0">
                          <MoreHorizontal className="w-4 h-4" />
                        </Button>
                      </DropdownMenuTrigger>
                      <DropdownMenuContent align="end">
                        {item.status === 'open' && (
                          <DropdownMenuItem onClick={() => handleQuickStatus(item, 'in_progress')} className="gap-2">
                            <PlayCircle className="w-4 h-4" /> {t('maintenance.startProgress')}
                          </DropdownMenuItem>
                        )}
                        {(item.status === 'open' || item.status === 'in_progress') && (
                          <DropdownMenuItem onClick={() => handleQuickStatus(item, 'done')} className="gap-2">
                            <CheckCircle2 className="w-4 h-4" /> {t('maintenance.markDone')}
                          </DropdownMenuItem>
                        )}
                        {(item.status === 'open' || item.status === 'in_progress') && (
                          <DropdownMenuSeparator />
                        )}
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
                    {t('maintenance.showing', {
                      from: (page - 1) * perPage + 1,
                      to: Math.min(page * perPage, total),
                      total,
                    })}
                  </span>
                  <div className="flex items-center gap-4">
                    <div className="flex items-center gap-2">
                      <span className="hidden sm:inline shrink-0">{t('maintenance.perPage')}</span>
                      <Select value={String(perPage)} onValueChange={handlePerPageChange}>
                        <SelectTrigger className="h-8 w-16"><SelectValue /></SelectTrigger>
                        <SelectContent>
                          {PER_PAGE_OPTIONS.map((n) => (
                            <SelectItem key={n} value={String(n)}>{n}</SelectItem>
                          ))}
                        </SelectContent>
                      </Select>
                    </div>
                    <span className="shrink-0">{t('maintenance.page')} {page} {t('maintenance.of')} {totalPages}</span>
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

      <MaintenanceDialog
        open={dialogOpen}
        onOpenChange={setDialogOpen}
        editingItem={editingItem}
        form={form}
        onFormChange={setForm}
        onSave={handleSave}
        saving={saving}
        error={saveError}
      />

      <MaintenanceDeleteDialog
        open={deleteDialogOpen}
        onOpenChange={setDeleteDialogOpen}
        item={deletingItem}
        onDelete={handleDelete}
      />
    </div>
  )
}
