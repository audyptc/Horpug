import { useState, useMemo, useEffect, useCallback } from 'react'
import { useTranslation } from 'react-i18next'
import {
  Search,
  Plus,
  MoreHorizontal,
  Pencil,
  Trash2,
  BedDouble,
  SearchX,
  RefreshCw,
  ChevronLeft,
  ChevronRight,
  ChevronsLeft,
  ChevronsRight,
} from 'lucide-react'
import { Button } from '@/components/ui/button'
import { Input } from '@/components/ui/input'
import { Badge } from '@/components/ui/badge'
import { Card, CardContent, CardHeader, CardTitle, CardDescription } from '@/components/ui/card'
import { Label } from '@/components/ui/label'
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
import { roomService } from '@/features/rooms/roomService'
import type { ApiRoom, RoomStatus, RoomType } from '@/types/api'
import { cn } from '@/lib/utils'
import { formatDate } from '@/lib/dateUtils'

const PER_PAGE_OPTIONS = [10, 20, 50] as const

const statusVariantMap: Record<RoomStatus, 'success' | 'warning' | 'destructive'> = {
  available: 'success',
  occupied: 'warning',
  maintenance: 'destructive',
}

const emptyForm = {
  room_number: '',
  floor: 1,
  type: 'standard' as RoomType,
  status: 'available' as RoomStatus,
  rent_price: 0,
  description: '',
}

export function Rooms() {
  const { t } = useTranslation()
  const [rooms, setRooms] = useState<ApiRoom[]>([])
  const [loading, setLoading] = useState(true)
  const [error, setError] = useState('')
  const [search, setSearch] = useState('')
  const [typeFilter, setTypeFilter] = useState<RoomType | 'all'>('all')
  const [statusFilter, setStatusFilter] = useState<RoomStatus | 'all'>('all')
  const [page, setPage] = useState(1)
  const [perPage, setPerPage] = useState<number>(20)
  const [total, setTotal] = useState(0)
  const [totalPages, setTotalPages] = useState(1)
  const [dialogOpen, setDialogOpen] = useState(false)
  const [deleteDialogOpen, setDeleteDialogOpen] = useState(false)
  const [editingRoom, setEditingRoom] = useState<ApiRoom | null>(null)
  const [deletingRoom, setDeletingRoom] = useState<ApiRoom | null>(null)
  const [form, setForm] = useState(emptyForm)
  const [saving, setSaving] = useState(false)

  const fetchRooms = useCallback(async (p: number, pp: number) => {
    setLoading(true)
    setError('')
    try {
      const resp = await roomService.list(p, pp)
      setRooms(resp.data)
      setTotal(resp.meta.total)
      setTotalPages(resp.meta.total_pages)
    } catch {
      setError(t('rooms.loadError'))
    } finally {
      setLoading(false)
    }
  }, [t])

  useEffect(() => {
    fetchRooms(page, perPage)
  }, [fetchRooms, page, perPage])

  function handlePerPageChange(value: string) {
    setPerPage(Number(value))
    setPage(1)
  }

  const filtered = useMemo(() => {
    return rooms.filter((r) => {
      const matchSearch = r.room_number.toLowerCase().includes(search.toLowerCase())
      const matchType = typeFilter === 'all' || r.type === typeFilter
      const matchStatus = statusFilter === 'all' || r.status === statusFilter
      return matchSearch && matchType && matchStatus
    })
  }, [rooms, search, typeFilter, statusFilter])

  function openCreate() {
    setEditingRoom(null)
    setForm(emptyForm)
    setDialogOpen(true)
  }

  function openEdit(room: ApiRoom) {
    setEditingRoom(room)
    setForm({
      room_number: room.room_number,
      floor: room.floor,
      type: room.type,
      status: room.status,
      rent_price: room.rent_price,
      description: room.description,
    })
    setDialogOpen(true)
  }

  async function handleSave() {
    if (!form.room_number || form.floor <= 0 || form.rent_price <= 0) return
    setSaving(true)
    try {
      if (editingRoom) {
        await roomService.update(editingRoom.id, form)
      } else {
        await roomService.create(form)
      }
      setDialogOpen(false)
      await fetchRooms(page, perPage)
    } catch {
      // error handled silently
    } finally {
      setSaving(false)
    }
  }

  async function handleDelete() {
    if (!deletingRoom) return
    try {
      await roomService.delete(deletingRoom.id)
      const newPage = rooms.length === 1 && page > 1 ? page - 1 : page
      setPage(newPage)
      await fetchRooms(newPage, perPage)
    } catch {
      // ignore
    } finally {
      setDeleteDialogOpen(false)
      setDeletingRoom(null)
    }
  }

  const isSaveDisabled = saving || !form.room_number || form.floor <= 0 || form.rent_price <= 0

  return (
    <div className="space-y-6">
      <div className="flex flex-wrap items-start justify-between gap-3">
        <div className="min-w-0">
          <h1 className="text-2xl font-bold tracking-tight">{t('rooms.title')}</h1>
          <p className="text-muted-foreground text-sm mt-1">
            {t('rooms.subtitle', { count: filtered.length, total })}
          </p>
        </div>
        <div className="flex gap-2 shrink-0">
          <Button variant="outline" size="icon" onClick={() => fetchRooms(page, perPage)} disabled={loading}>
            <RefreshCw className={cn('w-4 h-4', loading && 'animate-spin')} />
          </Button>
          <Button onClick={openCreate} className="gap-2">
            <Plus className="w-4 h-4" />
            <span className="hidden sm:inline">{t('rooms.addRoom')}</span>
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
              <CardTitle>{t('rooms.roomList')}</CardTitle>
              <CardDescription>{t('rooms.roomListDesc')}</CardDescription>
            </div>
            <div className="flex flex-col sm:flex-row gap-2 sm:items-center">
              <div className="relative">
                <Search className="absolute left-3 top-1/2 -translate-y-1/2 w-4 h-4 text-muted-foreground" />
                <Input
                  placeholder={t('rooms.searchPlaceholder')}
                  className="pl-9 w-full sm:w-48"
                  value={search}
                  onChange={(e) => setSearch(e.target.value)}
                />
              </div>
              <Select value={typeFilter} onValueChange={(v) => setTypeFilter(v as RoomType | 'all')}>
                <SelectTrigger className="w-full sm:w-32">
                  <SelectValue placeholder={t('rooms.allTypes')} />
                </SelectTrigger>
                <SelectContent>
                  <SelectItem value="all">{t('rooms.allTypes')}</SelectItem>
                  <SelectItem value="standard">{t('rooms.types.standard')}</SelectItem>
                  <SelectItem value="deluxe">{t('rooms.types.deluxe')}</SelectItem>
                  <SelectItem value="suite">{t('rooms.types.suite')}</SelectItem>
                </SelectContent>
              </Select>
              <Select value={statusFilter} onValueChange={(v) => setStatusFilter(v as RoomStatus | 'all')}>
                <SelectTrigger className="w-full sm:w-36">
                  <SelectValue placeholder={t('rooms.allStatuses')} />
                </SelectTrigger>
                <SelectContent>
                  <SelectItem value="all">{t('rooms.allStatuses')}</SelectItem>
                  <SelectItem value="available">{t('rooms.statuses.available')}</SelectItem>
                  <SelectItem value="occupied">{t('rooms.statuses.occupied')}</SelectItem>
                  <SelectItem value="maintenance">{t('rooms.statuses.maintenance')}</SelectItem>
                </SelectContent>
              </Select>
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
                      <th className="text-left px-6 py-3 font-medium text-muted-foreground">{t('rooms.colRoom')}</th>
                      <th className="text-left px-4 py-3 font-medium text-muted-foreground">{t('rooms.colFloor')}</th>
                      <th className="text-left px-4 py-3 font-medium text-muted-foreground">{t('rooms.colType')}</th>
                      <th className="text-left px-4 py-3 font-medium text-muted-foreground">{t('rooms.colStatus')}</th>
                      <th className="text-left px-4 py-3 font-medium text-muted-foreground">{t('rooms.colRentPrice')}</th>
                      <th className="text-left px-4 py-3 font-medium text-muted-foreground">{t('rooms.colCreated')}</th>
                      <th className="text-right px-6 py-3 font-medium text-muted-foreground">{t('rooms.colActions')}</th>
                    </tr>
                  </thead>
                  <tbody>
                    {filtered.map((room, i) => (
                      <tr
                        key={room.id}
                        className={cn(
                          'border-b transition-colors hover:bg-muted/30',
                          i === filtered.length - 1 && 'border-0'
                        )}
                      >
                        <td className="px-6 py-4">
                          <div className="flex items-center gap-2">
                            <div className="h-8 w-8 rounded-lg bg-muted flex items-center justify-center shrink-0">
                              <BedDouble className="w-4 h-4 text-muted-foreground" />
                            </div>
                            <span className="font-medium">{room.room_number}</span>
                          </div>
                        </td>
                        <td className="px-4 py-4 text-muted-foreground">
                          {t('rooms.floorN', { n: room.floor })}
                        </td>
                        <td className="px-4 py-4">
                          <span className="text-xs px-2 py-0.5 rounded-md bg-muted font-medium capitalize">
                            {t(`rooms.types.${room.type}`)}
                          </span>
                        </td>
                        <td className="px-4 py-4">
                          <Badge variant={statusVariantMap[room.status]}>
                            {t(`rooms.statuses.${room.status}`)}
                          </Badge>
                        </td>
                        <td className="px-4 py-4 font-medium">
                          {room.rent_price.toLocaleString('th-TH')} {t('rooms.bahtPerMonth')}
                        </td>
                        <td className="px-4 py-4 text-muted-foreground">{formatDate(room.created_at)}</td>
                        <td className="px-6 py-4 text-right">
                          <DropdownMenu>
                            <DropdownMenuTrigger asChild>
                              <Button variant="ghost" size="icon" className="h-8 w-8">
                                <MoreHorizontal className="w-4 h-4" />
                              </Button>
                            </DropdownMenuTrigger>
                            <DropdownMenuContent align="end">
                              <DropdownMenuItem onClick={() => openEdit(room)} className="gap-2">
                                <Pencil className="w-4 h-4" /> {t('common.edit')}
                              </DropdownMenuItem>
                              <DropdownMenuSeparator />
                              <DropdownMenuItem
                                className="gap-2 text-destructive focus:text-destructive"
                                onClick={() => { setDeletingRoom(room); setDeleteDialogOpen(true) }}
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
                    <p className="text-sm font-medium">{t('rooms.noRooms')}</p>
                    {(search || typeFilter !== 'all' || statusFilter !== 'all') && (
                      <button
                        type="button"
                        onClick={() => { setSearch(''); setTypeFilter('all'); setStatusFilter('all') }}
                        className="text-xs text-primary hover:underline"
                      >
                        {t('rooms.clearFilters')}
                      </button>
                    )}
                  </div>
                )}
              </div>

              {/* Mobile cards */}
              <div className="md:hidden divide-y">
                {filtered.map((room) => (
                  <div key={room.id} className="p-4 flex items-center gap-3">
                    <div className="h-10 w-10 shrink-0 rounded-full bg-muted flex items-center justify-center">
                      <BedDouble className="w-5 h-5 text-muted-foreground" />
                    </div>
                    <div className="flex-1 min-w-0">
                      <p className="text-sm font-medium">{room.room_number}</p>
                      <p className="text-xs text-muted-foreground">
                        {t('rooms.floorN', { n: room.floor })} · {t(`rooms.types.${room.type}`)}
                      </p>
                      <div className="flex items-center gap-2 mt-1">
                        <Badge variant={statusVariantMap[room.status]} className="text-xs">
                          {t(`rooms.statuses.${room.status}`)}
                        </Badge>
                        <span className="text-xs text-muted-foreground">
                          {room.rent_price.toLocaleString('th-TH')} ฿
                        </span>
                      </div>
                    </div>
                    <DropdownMenu>
                      <DropdownMenuTrigger asChild>
                        <Button variant="ghost" size="icon" className="h-8 w-8 shrink-0">
                          <MoreHorizontal className="w-4 h-4" />
                        </Button>
                      </DropdownMenuTrigger>
                      <DropdownMenuContent align="end">
                        <DropdownMenuItem onClick={() => openEdit(room)}>{t('common.edit')}</DropdownMenuItem>
                        <DropdownMenuSeparator />
                        <DropdownMenuItem
                          className="text-destructive focus:text-destructive"
                          onClick={() => { setDeletingRoom(room); setDeleteDialogOpen(true) }}
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
                    {t('rooms.showing', {
                      from: (page - 1) * perPage + 1,
                      to: Math.min(page * perPage, total),
                      total,
                    })}
                  </span>
                  <div className="flex items-center gap-4">
                    <div className="flex items-center gap-2">
                      <span className="hidden sm:inline shrink-0">{t('rooms.perPage')}</span>
                      <Select value={String(perPage)} onValueChange={handlePerPageChange}>
                        <SelectTrigger className="h-8 w-16">
                          <SelectValue />
                        </SelectTrigger>
                        <SelectContent>
                          {PER_PAGE_OPTIONS.map((n) => (
                            <SelectItem key={n} value={String(n)}>{n}</SelectItem>
                          ))}
                        </SelectContent>
                      </Select>
                    </div>
                    <span className="shrink-0">
                      {t('rooms.page')} {page} {t('rooms.of')} {totalPages}
                    </span>
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
        <DialogContent className="max-w-md">
          <DialogHeader>
            <DialogTitle>{editingRoom ? t('rooms.editRoom') : t('rooms.createRoom')}</DialogTitle>
            <DialogDescription>
              {editingRoom ? t('rooms.editDesc') : t('rooms.createDesc')}
            </DialogDescription>
          </DialogHeader>
          <div className="grid gap-4 py-4">
            <div className="grid grid-cols-2 gap-4">
              <div className="space-y-1.5">
                <Label>{t('rooms.roomNumber')} *</Label>
                <Input
                  placeholder="101"
                  value={form.room_number}
                  onChange={(e) => setForm((f) => ({ ...f, room_number: e.target.value }))}
                />
              </div>
              <div className="space-y-1.5">
                <Label>{t('rooms.floor')} *</Label>
                <Input
                  type="number"
                  min={1}
                  placeholder="1"
                  value={form.floor}
                  onChange={(e) => setForm((f) => ({ ...f, floor: Number(e.target.value) }))}
                />
              </div>
            </div>
            <div className="grid grid-cols-2 gap-4">
              <div className="space-y-1.5">
                <Label>{t('rooms.type')}</Label>
                <Select value={form.type} onValueChange={(v) => setForm((f) => ({ ...f, type: v as RoomType }))}>
                  <SelectTrigger><SelectValue /></SelectTrigger>
                  <SelectContent>
                    <SelectItem value="standard">{t('rooms.types.standard')}</SelectItem>
                    <SelectItem value="deluxe">{t('rooms.types.deluxe')}</SelectItem>
                    <SelectItem value="suite">{t('rooms.types.suite')}</SelectItem>
                  </SelectContent>
                </Select>
              </div>
              <div className="space-y-1.5">
                <Label>{t('rooms.status')}</Label>
                <Select value={form.status} onValueChange={(v) => setForm((f) => ({ ...f, status: v as RoomStatus }))}>
                  <SelectTrigger><SelectValue /></SelectTrigger>
                  <SelectContent>
                    <SelectItem value="available">{t('rooms.statuses.available')}</SelectItem>
                    <SelectItem value="occupied">{t('rooms.statuses.occupied')}</SelectItem>
                    <SelectItem value="maintenance">{t('rooms.statuses.maintenance')}</SelectItem>
                  </SelectContent>
                </Select>
              </div>
            </div>
            <div className="space-y-1.5">
              <Label>{t('rooms.rentPrice')} * ({t('rooms.bahtPerMonth')})</Label>
              <Input
                type="number"
                min={0}
                placeholder="3500"
                value={form.rent_price || ''}
                onChange={(e) => setForm((f) => ({ ...f, rent_price: Number(e.target.value) }))}
              />
            </div>
            <div className="space-y-1.5">
              <Label>{t('rooms.description')}</Label>
              <textarea
                rows={2}
                placeholder={t('rooms.descriptionPlaceholder')}
                value={form.description}
                onChange={(e: React.ChangeEvent<HTMLTextAreaElement>) =>
                  setForm((f) => ({ ...f, description: e.target.value }))
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
              {saving ? t('common.loading') : editingRoom ? t('rooms.saveChanges') : t('rooms.createRoom')}
            </Button>
          </DialogFooter>
        </DialogContent>
      </Dialog>

      {/* Delete confirm */}
      <Dialog open={deleteDialogOpen} onOpenChange={setDeleteDialogOpen}>
        <DialogContent className="max-w-sm">
          <DialogHeader>
            <DialogTitle>{t('rooms.deleteRoom')}</DialogTitle>
            <DialogDescription>
              {t('rooms.deleteConfirm')}{' '}
              <span className="font-semibold text-foreground">{deletingRoom?.room_number}</span>?{' '}
              {t('rooms.deleteWarning')}
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
