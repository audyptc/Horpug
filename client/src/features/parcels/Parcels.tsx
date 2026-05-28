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
import { parcelService } from '@/features/parcels/parcelService'
import type { ApiParcel, ParcelStatus } from '@/types'
import { cn } from '@/lib/utils'
import { formatDate } from '@/lib/dateUtils'

const PER_PAGE_OPTIONS = [10, 20, 50] as const

function toDateInput(val: string | null | undefined): string {
  if (!val) return ''
  return val.slice(0, 10)
}

const emptyForm = {
  tracking_number: '',
  recipient_name: '',
  room_number: '',
  status: 'pending' as ParcelStatus,
  received_date: new Date().toISOString().slice(0, 10),
  picked_up_date: '',
  note: '',
}

export function Parcels() {
  const { t } = useTranslation()

  const [items, setItems] = useState<ApiParcel[]>([])
  const [loading, setLoading] = useState(true)
  const [error, setError] = useState('')
  const [search, setSearch] = useState('')
  const [filterStatus, setFilterStatus] = useState<string>('all')
  const [page, setPage] = useState(1)
  const [perPage, setPerPage] = useState<number>(20)
  const [total, setTotal] = useState(0)
  const [totalPages, setTotalPages] = useState(1)

  const [dialogOpen, setDialogOpen] = useState(false)
  const [deleteDialogOpen, setDeleteDialogOpen] = useState(false)
  const [editingItem, setEditingItem] = useState<ApiParcel | null>(null)
  const [deletingItem, setDeletingItem] = useState<ApiParcel | null>(null)
  const [form, setForm] = useState(emptyForm)
  const [saving, setSaving] = useState(false)

  const fetchItems = useCallback(
    async (p: number, pp: number) => {
      setLoading(true)
      setError('')
      try {
        const resp = await parcelService.list(p, pp)
        setItems(resp.data)
        setTotal(resp.meta.total)
        setTotalPages(resp.meta.total_pages)
      } catch {
        setError(t('parcels.loadError'))
      } finally {
        setLoading(false)
      }
    },
    [t]
  )

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
        item.tracking_number.toLowerCase().includes(q) ||
        item.recipient_name.toLowerCase().includes(q) ||
        item.room_number.toLowerCase().includes(q)
      const matchStatus = filterStatus === 'all' || item.status === filterStatus
      return matchSearch && matchStatus
    })
  }, [items, search, filterStatus])

  function openCreate() {
    setEditingItem(null)
    setForm(emptyForm)
    setDialogOpen(true)
  }

  function openEdit(item: ApiParcel) {
    setEditingItem(item)
    setForm({
      tracking_number: item.tracking_number,
      recipient_name: item.recipient_name,
      room_number: item.room_number,
      status: item.status,
      received_date: toDateInput(item.received_date),
      picked_up_date: toDateInput(item.picked_up_date),
      note: item.note,
    })
    setDialogOpen(true)
  }

  async function handleSave() {
    if (!form.recipient_name || !form.received_date) return
    setSaving(true)
    try {
      const payload = {
        tracking_number: form.tracking_number,
        recipient_name: form.recipient_name,
        room_number: form.room_number,
        status: form.status,
        received_date: form.received_date,
        picked_up_date: form.picked_up_date || null,
        note: form.note,
      }
      if (editingItem) {
        await parcelService.update(editingItem.id, payload)
      } else {
        await parcelService.create(payload)
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
      await parcelService.delete(deletingItem.id)
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

  const isSaveDisabled = saving || !form.recipient_name || !form.received_date

  return (
    <div className="space-y-6">
      <div className="flex flex-wrap items-start justify-between gap-3">
        <div className="min-w-0">
          <h1 className="text-2xl font-bold tracking-tight">{t('parcels.title')}</h1>
          <p className="text-muted-foreground text-sm mt-1">
            {t('parcels.subtitle', { count: filtered.length, total })}
          </p>
        </div>
        <div className="flex gap-2 shrink-0">
          <Button variant="outline" size="icon" onClick={() => fetchItems(page, perPage)} disabled={loading}>
            <RefreshCw className={cn('w-4 h-4', loading && 'animate-spin')} />
          </Button>
          <Button onClick={openCreate} className="gap-2">
            <Plus className="w-4 h-4" />
            <span className="hidden sm:inline">{t('parcels.addParcel')}</span>
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
              <CardTitle>{t('parcels.parcelList')}</CardTitle>
              <CardDescription>{t('parcels.parcelListDesc')}</CardDescription>
            </div>
            <div className="flex flex-wrap gap-2">
              <Select value={filterStatus} onValueChange={setFilterStatus}>
                <SelectTrigger className="h-9 w-36">
                  <SelectValue />
                </SelectTrigger>
                <SelectContent>
                  <SelectItem value="all">{t('parcels.allStatuses')}</SelectItem>
                  <SelectItem value="pending">{t('parcels.statuses.pending')}</SelectItem>
                  <SelectItem value="picked_up">{t('parcels.statuses.picked_up')}</SelectItem>
                </SelectContent>
              </Select>
              <div className="relative">
                <Search className="absolute left-3 top-1/2 -translate-y-1/2 w-4 h-4 text-muted-foreground" />
                <Input
                  placeholder={t('parcels.searchPlaceholder')}
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
                      <th className="text-left px-6 py-3 font-medium text-muted-foreground">{t('parcels.colTracking')}</th>
                      <th className="text-left px-4 py-3 font-medium text-muted-foreground">{t('parcels.colRecipient')}</th>
                      <th className="text-left px-4 py-3 font-medium text-muted-foreground">{t('parcels.colRoom')}</th>
                      <th className="text-left px-4 py-3 font-medium text-muted-foreground">{t('parcels.colStatus')}</th>
                      <th className="text-left px-4 py-3 font-medium text-muted-foreground">{t('parcels.colReceivedDate')}</th>
                      <th className="text-left px-4 py-3 font-medium text-muted-foreground">{t('parcels.colPickedUpDate')}</th>
                      <th className="text-right px-6 py-3 font-medium text-muted-foreground">{t('parcels.colActions')}</th>
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
                        <td className="px-6 py-4 font-medium font-mono text-xs">
                          {item.tracking_number || '-'}
                        </td>
                        <td className="px-4 py-4">{item.recipient_name}</td>
                        <td className="px-4 py-4 text-muted-foreground">{item.room_number || '-'}</td>
                        <td className="px-4 py-4">
                          <Badge variant={item.status === 'pending' ? 'secondary' : 'default'}>
                            {t(`parcels.statuses.${item.status}`)}
                          </Badge>
                        </td>
                        <td className="px-4 py-4 text-muted-foreground">{formatDate(item.received_date)}</td>
                        <td className="px-4 py-4 text-muted-foreground">
                          {item.picked_up_date ? formatDate(item.picked_up_date) : '-'}
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
                    <p className="text-sm font-medium">{t('parcels.noParcels')}</p>
                    {(search || filterStatus !== 'all') && (
                      <button
                        type="button"
                        onClick={() => { setSearch(''); setFilterStatus('all') }}
                        className="text-xs text-primary hover:underline"
                      >
                        {t('parcels.clearFilters')}
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
                        <p className="text-sm font-medium">{item.recipient_name}</p>
                        <Badge variant={item.status === 'pending' ? 'secondary' : 'default'} className="text-xs">
                          {t(`parcels.statuses.${item.status}`)}
                        </Badge>
                      </div>
                      {item.tracking_number && (
                        <p className="text-xs text-muted-foreground font-mono">{item.tracking_number}</p>
                      )}
                      <p className="text-xs text-muted-foreground">
                        {item.room_number ? `${t('parcels.colRoom')}: ${item.room_number} · ` : ''}
                        {formatDate(item.received_date)}
                      </p>
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
                    {t('parcels.showing', {
                      from: (page - 1) * perPage + 1,
                      to: Math.min(page * perPage, total),
                      total,
                    })}
                  </span>
                  <div className="flex items-center gap-4">
                    <div className="flex items-center gap-2">
                      <span className="hidden sm:inline shrink-0">{t('parcels.perPage')}</span>
                      <Select value={String(perPage)} onValueChange={handlePerPageChange}>
                        <SelectTrigger className="h-8 w-16"><SelectValue /></SelectTrigger>
                        <SelectContent>
                          {PER_PAGE_OPTIONS.map((n) => (
                            <SelectItem key={n} value={String(n)}>{n}</SelectItem>
                          ))}
                        </SelectContent>
                      </Select>
                    </div>
                    <span className="shrink-0">{t('parcels.page')} {page} {t('parcels.of')} {totalPages}</span>
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
              {editingItem ? t('parcels.editParcel') : t('parcels.createParcel')}
            </DialogTitle>
            <DialogDescription>
              {editingItem ? t('parcels.editDesc') : t('parcels.createDesc')}
            </DialogDescription>
          </DialogHeader>

          <div className="grid gap-4 py-4">
            <div className="space-y-1.5">
              <Label>{t('parcels.trackingNumber')}</Label>
              <Input
                placeholder={t('parcels.trackingNumberPlaceholder')}
                value={form.tracking_number}
                onChange={(e) => setForm((f) => ({ ...f, tracking_number: e.target.value }))}
              />
            </div>

            <div className="grid grid-cols-2 gap-4">
              <div className="space-y-1.5">
                <Label>{t('parcels.recipientName')} *</Label>
                <Input
                  placeholder={t('parcels.recipientNamePlaceholder')}
                  value={form.recipient_name}
                  onChange={(e) => setForm((f) => ({ ...f, recipient_name: e.target.value }))}
                />
              </div>
              <div className="space-y-1.5">
                <Label>{t('parcels.roomNumber')}</Label>
                <Input
                  placeholder={t('parcels.roomNumberPlaceholder')}
                  value={form.room_number}
                  onChange={(e) => setForm((f) => ({ ...f, room_number: e.target.value }))}
                />
              </div>
            </div>

            <div className="grid grid-cols-2 gap-4">
              <div className="space-y-1.5">
                <Label>{t('parcels.status')}</Label>
                <Select
                  value={form.status}
                  onValueChange={(v) => setForm((f) => ({ ...f, status: v as ParcelStatus }))}
                >
                  <SelectTrigger><SelectValue /></SelectTrigger>
                  <SelectContent>
                    <SelectItem value="pending">{t('parcels.statuses.pending')}</SelectItem>
                    <SelectItem value="picked_up">{t('parcels.statuses.picked_up')}</SelectItem>
                  </SelectContent>
                </Select>
              </div>
              <div className="space-y-1.5">
                <Label>{t('parcels.receivedDate')} *</Label>
                <Input
                  type="date"
                  value={form.received_date}
                  onChange={(e) => setForm((f) => ({ ...f, received_date: e.target.value }))}
                />
              </div>
            </div>

            <div className="space-y-1.5">
              <Label>{t('parcels.pickedUpDate')}</Label>
              <Input
                type="date"
                value={form.picked_up_date}
                onChange={(e) => setForm((f) => ({ ...f, picked_up_date: e.target.value }))}
              />
            </div>

            <div className="space-y-1.5">
              <Label>{t('parcels.note')}</Label>
              <Input
                placeholder={t('parcels.notePlaceholder')}
                value={form.note}
                onChange={(e) => setForm((f) => ({ ...f, note: e.target.value }))}
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
                : editingItem
                ? t('parcels.saveChanges')
                : t('parcels.createParcel')}
            </Button>
          </DialogFooter>
        </DialogContent>
      </Dialog>

      {/* Delete confirm */}
      <Dialog open={deleteDialogOpen} onOpenChange={setDeleteDialogOpen}>
        <DialogContent className="max-w-sm">
          <DialogHeader>
            <DialogTitle>{t('parcels.deleteParcel')}</DialogTitle>
            <DialogDescription>
              {t('parcels.deleteConfirm')}{' '}
              <span className="font-semibold text-foreground">{deletingItem?.recipient_name}</span>?{' '}
              {t('parcels.deleteWarning')}
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
