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
  Zap,
  Droplets,
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
import { meterReadingService } from '@/services/meterReadingService'
import { roomService } from '@/services/roomService'
import type { ApiMeterReading, ApiRoom, MeterType } from '@/types/api'
import { cn } from '@/lib/utils'
import { formatDate } from '@/lib/dateUtils'

const PER_PAGE_OPTIONS = [10, 20, 50] as const

const emptyForm = {
  room_id: '',
  meter_type: 'electric' as MeterType,
  reading_date: '',
  previous_reading: '',
  current_reading: '',
  unit_price: '',
  note: '',
}

function toDateInput(iso: string | null | undefined): string {
  if (!iso) return ''
  return iso.slice(0, 10)
}

export function MeterReadings() {
  const { t } = useTranslation()

  const [readings, setReadings] = useState<ApiMeterReading[]>([])
  const [loading, setLoading] = useState(true)
  const [error, setError] = useState('')
  const [search, setSearch] = useState('')
  const [filterType, setFilterType] = useState<string>('all')
  const [page, setPage] = useState(1)
  const [perPage, setPerPage] = useState<number>(20)
  const [total, setTotal] = useState(0)
  const [totalPages, setTotalPages] = useState(1)

  const [rooms, setRooms] = useState<ApiRoom[]>([])

  const [dialogOpen, setDialogOpen] = useState(false)
  const [deleteDialogOpen, setDeleteDialogOpen] = useState(false)
  const [editingReading, setEditingReading] = useState<ApiMeterReading | null>(null)
  const [deletingReading, setDeletingReading] = useState<ApiMeterReading | null>(null)
  const [form, setForm] = useState(emptyForm)
  const [saving, setSaving] = useState(false)

  const fetchReadings = useCallback(async (p: number, pp: number) => {
    setLoading(true)
    setError('')
    try {
      const resp = await meterReadingService.list(p, pp)
      setReadings(resp.data)
      setTotal(resp.meta.total)
      setTotalPages(resp.meta.total_pages)
    } catch {
      setError(t('meters.loadError'))
    } finally {
      setLoading(false)
    }
  }, [t])

  useEffect(() => {
    fetchReadings(page, perPage)
  }, [fetchReadings, page, perPage])

  useEffect(() => {
    roomService.list(1, 200).then((r) => setRooms(r.data)).catch(() => {})
  }, [])

  function handlePerPageChange(value: string) {
    setPerPage(Number(value))
    setPage(1)
  }

  const filtered = useMemo(() => {
    const q = search.toLowerCase()
    return readings.filter((r) => {
      const matchSearch = r.room_number.toLowerCase().includes(q)
      const matchType = filterType === 'all' || r.meter_type === filterType
      return matchSearch && matchType
    })
  }, [readings, search, filterType])

  function openCreate() {
    setEditingReading(null)
    setForm(emptyForm)
    setDialogOpen(true)
  }

  function openEdit(reading: ApiMeterReading) {
    setEditingReading(reading)
    setForm({
      room_id: reading.room_id,
      meter_type: reading.meter_type,
      reading_date: toDateInput(reading.reading_date),
      previous_reading: String(reading.previous_reading),
      current_reading: String(reading.current_reading),
      unit_price: String(reading.unit_price),
      note: reading.note,
    })
    setDialogOpen(true)
  }

  async function handleSave() {
    if (!form.room_id || !form.reading_date || !form.current_reading || !form.unit_price) return
    setSaving(true)
    try {
      if (editingReading) {
        await meterReadingService.update(editingReading.id, {
          reading_date: form.reading_date,
          previous_reading: Number(form.previous_reading),
          current_reading: Number(form.current_reading),
          unit_price: Number(form.unit_price),
          note: form.note,
        })
      } else {
        await meterReadingService.create({
          room_id: form.room_id,
          meter_type: form.meter_type,
          reading_date: form.reading_date,
          previous_reading: Number(form.previous_reading),
          current_reading: Number(form.current_reading),
          unit_price: Number(form.unit_price),
          note: form.note,
        })
      }
      setDialogOpen(false)
      await fetchReadings(page, perPage)
    } catch {
      // handled silently
    } finally {
      setSaving(false)
    }
  }

  async function handleDelete() {
    if (!deletingReading) return
    try {
      await meterReadingService.delete(deletingReading.id)
      const newPage = readings.length === 1 && page > 1 ? page - 1 : page
      setPage(newPage)
      await fetchReadings(newPage, perPage)
    } catch {
      // ignore
    } finally {
      setDeleteDialogOpen(false)
      setDeletingReading(null)
    }
  }

  const unitUsed =
    form.current_reading && form.previous_reading
      ? Number(form.current_reading) - Number(form.previous_reading)
      : null

  const totalAmount =
    unitUsed !== null && form.unit_price ? unitUsed * Number(form.unit_price) : null

  const isSaveDisabled =
    saving ||
    !form.room_id ||
    !form.reading_date ||
    !form.current_reading ||
    !form.unit_price ||
    Number(form.current_reading) < Number(form.previous_reading)

  return (
    <div className="space-y-6">
      <div className="flex flex-wrap items-start justify-between gap-3">
        <div className="min-w-0">
          <h1 className="text-2xl font-bold tracking-tight">{t('meters.title')}</h1>
          <p className="text-muted-foreground text-sm mt-1">
            {t('meters.subtitle', { count: filtered.length, total })}
          </p>
        </div>
        <div className="flex gap-2 shrink-0">
          <Button variant="outline" size="icon" onClick={() => fetchReadings(page, perPage)} disabled={loading}>
            <RefreshCw className={cn('w-4 h-4', loading && 'animate-spin')} />
          </Button>
          <Button onClick={openCreate} className="gap-2">
            <Plus className="w-4 h-4" />
            <span className="hidden sm:inline">{t('meters.addReading')}</span>
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
              <CardTitle>{t('meters.readingList')}</CardTitle>
              <CardDescription>{t('meters.readingListDesc')}</CardDescription>
            </div>
            <div className="flex gap-2">
              <Select value={filterType} onValueChange={setFilterType}>
                <SelectTrigger className="h-9 w-36">
                  <SelectValue />
                </SelectTrigger>
                <SelectContent>
                  <SelectItem value="all">{t('meters.allTypes')}</SelectItem>
                  <SelectItem value="electric">{t('meters.types.electric')}</SelectItem>
                  <SelectItem value="water">{t('meters.types.water')}</SelectItem>
                </SelectContent>
              </Select>
              <div className="relative">
                <Search className="absolute left-3 top-1/2 -translate-y-1/2 w-4 h-4 text-muted-foreground" />
                <Input
                  placeholder={t('meters.searchPlaceholder')}
                  className="pl-9 w-full sm:w-48"
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
                      <th className="text-left px-6 py-3 font-medium text-muted-foreground">{t('meters.colRoom')}</th>
                      <th className="text-left px-4 py-3 font-medium text-muted-foreground">{t('meters.colType')}</th>
                      <th className="text-left px-4 py-3 font-medium text-muted-foreground">{t('meters.colDate')}</th>
                      <th className="text-right px-4 py-3 font-medium text-muted-foreground">{t('meters.colPrev')}</th>
                      <th className="text-right px-4 py-3 font-medium text-muted-foreground">{t('meters.colCurr')}</th>
                      <th className="text-right px-4 py-3 font-medium text-muted-foreground">{t('meters.colUsed')}</th>
                      <th className="text-right px-4 py-3 font-medium text-muted-foreground">{t('meters.colUnitPrice')}</th>
                      <th className="text-right px-4 py-3 font-medium text-muted-foreground">{t('meters.colTotal')}</th>
                      <th className="text-right px-6 py-3 font-medium text-muted-foreground">{t('meters.colActions')}</th>
                    </tr>
                  </thead>
                  <tbody>
                    {filtered.map((r, i) => (
                      <tr
                        key={r.id}
                        className={cn(
                          'border-b transition-colors hover:bg-muted/30',
                          i === filtered.length - 1 && 'border-0'
                        )}
                      >
                        <td className="px-6 py-4 font-medium">{r.room_number}</td>
                        <td className="px-4 py-4">
                          <Badge variant={r.meter_type === 'electric' ? 'default' : 'secondary'} className="gap-1">
                            {r.meter_type === 'electric'
                              ? <Zap className="w-3 h-3" />
                              : <Droplets className="w-3 h-3" />}
                            {t(`meters.types.${r.meter_type}`)}
                          </Badge>
                        </td>
                        <td className="px-4 py-4 text-muted-foreground">{formatDate(r.reading_date)}</td>
                        <td className="px-4 py-4 text-right text-muted-foreground">{r.previous_reading.toLocaleString()}</td>
                        <td className="px-4 py-4 text-right text-muted-foreground">{r.current_reading.toLocaleString()}</td>
                        <td className="px-4 py-4 text-right font-medium">{r.unit_used.toLocaleString()}</td>
                        <td className="px-4 py-4 text-right text-muted-foreground">{r.unit_price.toLocaleString()}</td>
                        <td className="px-4 py-4 text-right font-semibold">
                          {r.total_amount.toLocaleString()}
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
                              <DropdownMenuItem onClick={() => openEdit(r)} className="gap-2">
                                <Pencil className="w-4 h-4" /> {t('common.edit')}
                              </DropdownMenuItem>
                              <DropdownMenuSeparator />
                              <DropdownMenuItem
                                className="gap-2 text-destructive focus:text-destructive"
                                onClick={() => { setDeletingReading(r); setDeleteDialogOpen(true) }}
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
                    <p className="text-sm font-medium">{t('meters.noReadings')}</p>
                    {(search || filterType !== 'all') && (
                      <button
                        type="button"
                        onClick={() => { setSearch(''); setFilterType('all') }}
                        className="text-xs text-primary hover:underline"
                      >
                        {t('meters.clearFilters')}
                      </button>
                    )}
                  </div>
                )}
              </div>

              {/* Mobile cards */}
              <div className="md:hidden divide-y">
                {filtered.map((r) => (
                  <div key={r.id} className="p-4 flex items-start gap-3">
                    <div className="flex-1 min-w-0 space-y-1">
                      <div className="flex items-center gap-2">
                        <p className="text-sm font-medium">{t('meters.colRoom')} {r.room_number}</p>
                        <Badge variant={r.meter_type === 'electric' ? 'default' : 'secondary'} className="gap-1 text-xs">
                          {r.meter_type === 'electric' ? <Zap className="w-3 h-3" /> : <Droplets className="w-3 h-3" />}
                          {t(`meters.types.${r.meter_type}`)}
                        </Badge>
                      </div>
                      <p className="text-xs text-muted-foreground">{formatDate(r.reading_date)}</p>
                      <p className="text-xs text-muted-foreground">
                        {r.previous_reading} → {r.current_reading} ({t('meters.used')} {r.unit_used})
                      </p>
                      <p className="text-sm font-semibold">{r.total_amount.toLocaleString()} ฿</p>
                    </div>
                    <DropdownMenu>
                      <DropdownMenuTrigger asChild>
                        <Button variant="ghost" size="icon" className="h-8 w-8 shrink-0">
                          <MoreHorizontal className="w-4 h-4" />
                        </Button>
                      </DropdownMenuTrigger>
                      <DropdownMenuContent align="end">
                        <DropdownMenuItem onClick={() => openEdit(r)}>{t('common.edit')}</DropdownMenuItem>
                        <DropdownMenuSeparator />
                        <DropdownMenuItem
                          className="text-destructive focus:text-destructive"
                          onClick={() => { setDeletingReading(r); setDeleteDialogOpen(true) }}
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
                    {t('meters.showing', {
                      from: (page - 1) * perPage + 1,
                      to: Math.min(page * perPage, total),
                      total,
                    })}
                  </span>
                  <div className="flex items-center gap-4">
                    <div className="flex items-center gap-2">
                      <span className="hidden sm:inline shrink-0">{t('meters.perPage')}</span>
                      <Select value={String(perPage)} onValueChange={handlePerPageChange}>
                        <SelectTrigger className="h-8 w-16"><SelectValue /></SelectTrigger>
                        <SelectContent>
                          {PER_PAGE_OPTIONS.map((n) => (
                            <SelectItem key={n} value={String(n)}>{n}</SelectItem>
                          ))}
                        </SelectContent>
                      </Select>
                    </div>
                    <span className="shrink-0">{t('meters.page')} {page} {t('meters.of')} {totalPages}</span>
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
              {editingReading ? t('meters.editReading') : t('meters.createReading')}
            </DialogTitle>
            <DialogDescription>
              {editingReading ? t('meters.editDesc') : t('meters.createDesc')}
            </DialogDescription>
          </DialogHeader>

          <div className="grid gap-4 py-4">
            {/* Room */}
            <div className="space-y-1.5">
              <Label>{t('meters.room')} *</Label>
              <Select
                value={form.room_id}
                onValueChange={(v) => setForm((f) => ({ ...f, room_id: v }))}
                disabled={!!editingReading}
              >
                <SelectTrigger><SelectValue placeholder={t('meters.selectRoom')} /></SelectTrigger>
                <SelectContent>
                  {rooms.map((rm) => (
                    <SelectItem key={rm.id} value={rm.id}>
                      {rm.room_number}
                    </SelectItem>
                  ))}
                </SelectContent>
              </Select>
            </div>

            {/* Meter Type */}
            <div className="space-y-1.5">
              <Label>{t('meters.meterType')} *</Label>
              <Select
                value={form.meter_type}
                onValueChange={(v) => setForm((f) => ({ ...f, meter_type: v as MeterType }))}
                disabled={!!editingReading}
              >
                <SelectTrigger><SelectValue /></SelectTrigger>
                <SelectContent>
                  <SelectItem value="electric">{t('meters.types.electric')}</SelectItem>
                  <SelectItem value="water">{t('meters.types.water')}</SelectItem>
                </SelectContent>
              </Select>
            </div>

            {/* Date */}
            <div className="space-y-1.5">
              <Label>{t('meters.readingDate')} *</Label>
              <Input
                type="date"
                value={form.reading_date}
                onChange={(e) => setForm((f) => ({ ...f, reading_date: e.target.value }))}
              />
            </div>

            {/* Readings */}
            <div className="grid grid-cols-2 gap-4">
              <div className="space-y-1.5">
                <Label>{t('meters.previousReading')}</Label>
                <Input
                  type="number"
                  min="0"
                  step="0.01"
                  value={form.previous_reading}
                  onChange={(e) => setForm((f) => ({ ...f, previous_reading: e.target.value }))}
                />
              </div>
              <div className="space-y-1.5">
                <Label>{t('meters.currentReading')} *</Label>
                <Input
                  type="number"
                  min="0"
                  step="0.01"
                  value={form.current_reading}
                  onChange={(e) => setForm((f) => ({ ...f, current_reading: e.target.value }))}
                />
              </div>
            </div>

            {/* Unit price */}
            <div className="space-y-1.5">
              <Label>{t('meters.unitPrice')} *</Label>
              <Input
                type="number"
                min="0"
                step="0.0001"
                value={form.unit_price}
                onChange={(e) => setForm((f) => ({ ...f, unit_price: e.target.value }))}
              />
            </div>

            {/* Preview */}
            {unitUsed !== null && unitUsed >= 0 && (
              <div className="rounded-md border bg-muted/40 px-4 py-3 text-sm space-y-1">
                <div className="flex justify-between">
                  <span className="text-muted-foreground">{t('meters.unitUsed')}</span>
                  <span className="font-medium">{unitUsed.toLocaleString()} {t('meters.unit')}</span>
                </div>
                {totalAmount !== null && (
                  <div className="flex justify-between">
                    <span className="text-muted-foreground">{t('meters.totalAmount')}</span>
                    <span className="font-semibold">{totalAmount.toLocaleString()} ฿</span>
                  </div>
                )}
              </div>
            )}

            {/* Note */}
            <div className="space-y-1.5">
              <Label>{t('meters.note')}</Label>
              <textarea
                rows={2}
                placeholder={t('meters.notePlaceholder')}
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
                : editingReading
                  ? t('meters.saveChanges')
                  : t('meters.createReading')}
            </Button>
          </DialogFooter>
        </DialogContent>
      </Dialog>

      {/* Delete confirm */}
      <Dialog open={deleteDialogOpen} onOpenChange={setDeleteDialogOpen}>
        <DialogContent className="max-w-sm">
          <DialogHeader>
            <DialogTitle>{t('meters.deleteReading')}</DialogTitle>
            <DialogDescription>
              {t('meters.deleteConfirm')}{' '}
              <span className="font-semibold text-foreground">
                {t('meters.colRoom')} {deletingReading?.room_number}
              </span>?{' '}
              {t('meters.deleteWarning')}
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
