import { useState, useMemo, useEffect, useCallback } from 'react'
import { useTranslation } from 'react-i18next'
import {
  Search, Plus, MoreHorizontal, Pencil, Trash2, SearchX, RefreshCw,
  ChevronLeft, ChevronRight, ChevronsLeft, ChevronsRight,
} from 'lucide-react'
import { Button } from '@/components/ui/button'
import { Input } from '@/components/ui/input'
import { Card, CardContent, CardHeader, CardTitle, CardDescription } from '@/components/ui/card'
import { Badge } from '@/components/ui/badge'
import {
  DropdownMenu, DropdownMenuContent, DropdownMenuItem,
  DropdownMenuSeparator, DropdownMenuTrigger,
} from '@/components/ui/dropdown-menu'
import { Select, SelectContent, SelectItem, SelectTrigger, SelectValue } from '@/components/ui/select'
import { waterMeterService } from '@/features/water-meters/waterMeterService'
import { roomService } from '@/features/rooms/roomService'
import { WaterMeterDialog, type WaterMeterForm } from '@/features/water-meters/components/WaterMeterDialog'
import { WaterMeterDeleteDialog } from '@/features/water-meters/components/WaterMeterDeleteDialog'
import type { ApiWaterMeter, ApiRoom } from '@/types'
import { cn } from '@/lib/utils'
import { formatDate, formatMonth } from '@/lib/dateUtils'
import { usePermission } from '@/hooks/usePermission'

const PER_PAGE_OPTIONS = [10, 20, 50] as const

const emptyForm: WaterMeterForm = {
  room_id: '',
  billing_type: 'meter',
  billing_month: '',
  reading_date: '',
  previous_reading: '',
  current_reading: '',
  unit_price: '',
  flat_amount: '',
  note: '',
}

function toDateInput(iso: string | null | undefined): string {
  if (!iso) return ''
  return iso.slice(0, 10)
}

export function WaterMeters() {
  const { t } = useTranslation()
  const { canCreate, canUpdate, canDelete } = usePermission('/water-meters')

  const [readings, setReadings] = useState<ApiWaterMeter[]>([])
  const [loading, setLoading] = useState(true)
  const [error, setError] = useState('')
  const [search, setSearch] = useState('')
  const [page, setPage] = useState(1)
  const [perPage, setPerPage] = useState(20)
  const [total, setTotal] = useState(0)
  const [totalPages, setTotalPages] = useState(1)
  const [rooms, setRooms] = useState<ApiRoom[]>([])

  const [dialogOpen, setDialogOpen] = useState(false)
  const [deleteDialogOpen, setDeleteDialogOpen] = useState(false)
  const [editing, setEditing] = useState<ApiWaterMeter | null>(null)
  const [deleting, setDeleting] = useState<ApiWaterMeter | null>(null)
  const [form, setForm] = useState(emptyForm)
  const [saving, setSaving] = useState(false)

  const fetchReadings = useCallback(async (p: number, pp: number) => {
    setLoading(true)
    setError('')
    try {
      const resp = await waterMeterService.list(p, pp)
      setReadings(resp.data)
      setTotal(resp.meta.total)
      setTotalPages(resp.meta.total_pages)
    } catch {
      setError(t('waterMeters.loadError'))
    } finally {
      setLoading(false)
    }
  }, [t])

  useEffect(() => { fetchReadings(page, perPage) }, [fetchReadings, page, perPage])
  useEffect(() => { roomService.list(1, 200).then((r) => setRooms(r.data)).catch(() => {}) }, [])

  const filtered = useMemo(() => {
    const q = search.toLowerCase()
    return readings.filter((r) => r.room_number.toLowerCase().includes(q))
  }, [readings, search])

  function openCreate() {
    setEditing(null)
    setForm(emptyForm)
    setDialogOpen(true)
  }

  function openEdit(r: ApiWaterMeter) {
    setEditing(r)
    setForm({
      room_id: r.room_id,
      billing_type: r.billing_type,
      billing_month: r.billing_month ? r.billing_month.slice(0, 7) : '',
      reading_date: toDateInput(r.reading_date),
      previous_reading: r.previous_reading != null ? String(r.previous_reading) : '',
      current_reading: r.current_reading != null ? String(r.current_reading) : '',
      unit_price: r.unit_price != null ? String(r.unit_price) : '',
      flat_amount: r.flat_amount != null ? String(r.flat_amount) : '',
      note: r.note,
    })
    setDialogOpen(true)
  }

  async function handleSave() {
    if (!form.room_id || !form.reading_date) return
    setSaving(true)
    try {
      if (editing) {
        await waterMeterService.update(editing.id, {
          billing_type: form.billing_type,
          billing_month: form.billing_month ? `${form.billing_month}-01T00:00:00Z` : null,
          reading_date: `${form.reading_date}T00:00:00Z`,
          previous_reading: form.previous_reading ? Number(form.previous_reading) : undefined,
          current_reading: form.current_reading ? Number(form.current_reading) : undefined,
          unit_price: form.unit_price ? Number(form.unit_price) : undefined,
          flat_amount: form.flat_amount ? Number(form.flat_amount) : undefined,
          note: form.note,
        })
      } else {
        await waterMeterService.create({
          room_id: form.room_id,
          billing_type: form.billing_type,
          billing_month: form.billing_month ? `${form.billing_month}-01T00:00:00Z` : undefined,
          reading_date: `${form.reading_date}T00:00:00Z`,
          previous_reading: form.previous_reading ? Number(form.previous_reading) : undefined,
          current_reading: form.current_reading ? Number(form.current_reading) : undefined,
          unit_price: form.unit_price ? Number(form.unit_price) : undefined,
          flat_amount: form.flat_amount ? Number(form.flat_amount) : undefined,
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
    if (!deleting) return
    try {
      await waterMeterService.delete(deleting.id)
      const newPage = readings.length === 1 && page > 1 ? page - 1 : page
      setPage(newPage)
      await fetchReadings(newPage, perPage)
    } catch {
      // ignore
    } finally {
      setDeleteDialogOpen(false)
      setDeleting(null)
    }
  }

  return (
    <div className="space-y-6">
      <div className="flex flex-wrap items-start justify-between gap-3">
        <div className="min-w-0">
          <h1 className="text-2xl font-bold tracking-tight">{t('waterMeters.title')}</h1>
          <p className="text-muted-foreground text-sm mt-1">{t('waterMeters.subtitle', { count: filtered.length, total })}</p>
        </div>
        <div className="flex gap-2 shrink-0">
          <Button variant="outline" size="icon" onClick={() => fetchReadings(page, perPage)} disabled={loading}>
            <RefreshCw className={cn('w-4 h-4', loading && 'animate-spin')} />
          </Button>
          {canCreate && (
            <Button onClick={openCreate} className="gap-2">
              <Plus className="w-4 h-4" />
              <span className="hidden sm:inline">{t('waterMeters.addReading')}</span>
            </Button>
          )}
        </div>
      </div>

      {error && (
        <div className="rounded-md border border-destructive/30 bg-destructive/10 px-4 py-3 text-sm text-destructive">{error}</div>
      )}

      <Card>
        <CardHeader>
          <div className="flex flex-col sm:flex-row sm:items-center justify-between gap-3">
            <div>
              <CardTitle>{t('waterMeters.readingList')}</CardTitle>
              <CardDescription>{t('waterMeters.readingListDesc')}</CardDescription>
            </div>
            <div className="relative">
              <Search className="absolute left-3 top-1/2 -translate-y-1/2 w-4 h-4 text-muted-foreground" />
              <Input placeholder={t('waterMeters.searchPlaceholder')} className="pl-9 w-full sm:w-48"
                value={search} onChange={(e) => setSearch(e.target.value)} />
            </div>
          </div>
        </CardHeader>

        <CardContent className="p-0">
          {loading ? (
            <div className="flex items-center justify-center py-16 text-sm text-muted-foreground">{t('common.loading')}</div>
          ) : (
            <>
              <div className="hidden md:block overflow-x-auto">
                <table className="w-full text-sm">
                  <thead>
                    <tr className="border-b bg-muted/40">
                      <th className="text-left px-6 py-3 font-medium text-muted-foreground">{t('waterMeters.colRoom')}</th>
                      <th className="text-left px-4 py-3 font-medium text-muted-foreground">{t('waterMeters.colBillingType')}</th>
                      <th className="text-left px-4 py-3 font-medium text-muted-foreground">{t('waterMeters.colBillingMonth')}</th>
                      <th className="text-left px-4 py-3 font-medium text-muted-foreground">{t('waterMeters.colDate')}</th>
                      <th className="text-right px-4 py-3 font-medium text-muted-foreground">{t('waterMeters.colPreviousReading')}</th>
                      <th className="text-right px-4 py-3 font-medium text-muted-foreground">{t('waterMeters.colCurrentReading')}</th>
                      <th className="text-right px-4 py-3 font-medium text-muted-foreground">{t('waterMeters.colUsed')}</th>
                      <th className="text-right px-4 py-3 font-medium text-muted-foreground">{t('waterMeters.colTotal')}</th>
                      {(canUpdate || canDelete) && (
                        <th className="text-right px-6 py-3 font-medium text-muted-foreground">{t('waterMeters.colActions')}</th>
                      )}
                    </tr>
                  </thead>
                  <tbody>
                    {filtered.map((r, i) => (
                      <tr key={r.id} className={cn('border-b transition-colors hover:bg-muted/30', i === filtered.length - 1 && 'border-0')}>
                        <td className="px-6 py-4 font-medium">{r.room_number}</td>
                        <td className="px-4 py-4">
                          <Badge variant={r.billing_type === 'flat' ? 'secondary' : 'outline'}>
                            {t(`waterMeters.billingTypes.${r.billing_type}`)}
                          </Badge>
                        </td>
                        <td className="px-4 py-4 text-muted-foreground">
                          {r.billing_month ? formatMonth(r.billing_month) : <span className="text-muted-foreground/40">—</span>}
                        </td>
                        <td className="px-4 py-4 text-muted-foreground">{formatDate(r.reading_date)}</td>
                        <td className="px-4 py-4 text-right text-muted-foreground">
                          {r.previous_reading != null ? r.previous_reading.toLocaleString() : '-'}
                        </td>
                        <td className="px-4 py-4 text-right text-muted-foreground">
                          {r.current_reading != null ? r.current_reading.toLocaleString() : '-'}
                        </td>
                        <td className="px-4 py-4 text-right font-medium">
                          {r.unit_used != null ? r.unit_used.toLocaleString() : '-'}
                        </td>
                        <td className="px-4 py-4 text-right font-semibold">
                          {r.total_amount.toLocaleString()}<span className="text-xs text-muted-foreground ml-1">฿</span>
                        </td>
                        {(canUpdate || canDelete) && (
                          <td className="px-6 py-4 text-right">
                            <DropdownMenu>
                              <DropdownMenuTrigger asChild>
                                <Button variant="ghost" size="icon" className="h-8 w-8"><MoreHorizontal className="w-4 h-4" /></Button>
                              </DropdownMenuTrigger>
                              <DropdownMenuContent align="end">
                                {canUpdate && (
                                  <DropdownMenuItem onClick={() => openEdit(r)} className="gap-2">
                                    <Pencil className="w-4 h-4" /> {t('common.edit')}
                                  </DropdownMenuItem>
                                )}
                                {canUpdate && canDelete && <DropdownMenuSeparator />}
                                {canDelete && (
                                  <DropdownMenuItem className="gap-2 text-destructive focus:text-destructive"
                                    onClick={() => { setDeleting(r); setDeleteDialogOpen(true) }}>
                                    <Trash2 className="w-4 h-4" /> {t('common.delete')}
                                  </DropdownMenuItem>
                                )}
                              </DropdownMenuContent>
                            </DropdownMenu>
                          </td>
                        )}
                      </tr>
                    ))}
                  </tbody>
                </table>
                {filtered.length === 0 && (
                  <div className="flex flex-col items-center justify-center py-16 gap-3 text-muted-foreground">
                    <div className="p-4 rounded-full bg-muted"><SearchX className="w-6 h-6" /></div>
                    <p className="text-sm font-medium">{t('waterMeters.noReadings')}</p>
                    {search && (
                      <button type="button" onClick={() => setSearch('')} className="text-xs text-primary hover:underline">
                        {t('waterMeters.clearFilters')}
                      </button>
                    )}
                  </div>
                )}
              </div>

              <div className="md:hidden divide-y">
                {filtered.map((r) => (
                  <div key={r.id} className="p-4 flex items-start gap-3">
                    <div className="flex-1 min-w-0 space-y-1">
                      <div className="flex items-center gap-2">
                        <p className="text-sm font-medium">{t('waterMeters.colRoom')} {r.room_number}</p>
                        <Badge variant={r.billing_type === 'flat' ? 'secondary' : 'outline'} className="text-xs">
                          {t(`waterMeters.billingTypes.${r.billing_type}`)}
                        </Badge>
                      </div>
                      {r.billing_month && (
                        <p className="text-xs font-medium">{formatMonth(r.billing_month)}</p>
                      )}
                      <p className="text-xs text-muted-foreground">{formatDate(r.reading_date)}</p>
                      {r.billing_type === 'meter' && r.previous_reading != null && r.current_reading != null && (
                        <p className="text-xs text-muted-foreground">
                          {r.previous_reading} → {r.current_reading} ({t('waterMeters.used')} {r.unit_used})
                        </p>
                      )}
                      <p className="text-sm font-semibold">{r.total_amount.toLocaleString()} ฿</p>
                    </div>
                    {(canUpdate || canDelete) && (
                      <DropdownMenu>
                        <DropdownMenuTrigger asChild>
                          <Button variant="ghost" size="icon" className="h-8 w-8 shrink-0"><MoreHorizontal className="w-4 h-4" /></Button>
                        </DropdownMenuTrigger>
                        <DropdownMenuContent align="end">
                          {canUpdate && (
                            <DropdownMenuItem onClick={() => openEdit(r)}>{t('common.edit')}</DropdownMenuItem>
                          )}
                          {canUpdate && canDelete && <DropdownMenuSeparator />}
                          {canDelete && (
                            <DropdownMenuItem className="text-destructive focus:text-destructive"
                              onClick={() => { setDeleting(r); setDeleteDialogOpen(true) }}>
                              {t('common.delete')}
                            </DropdownMenuItem>
                          )}
                        </DropdownMenuContent>
                      </DropdownMenu>
                    )}
                  </div>
                ))}
              </div>

              {!loading && total > 0 && (
                <div className="flex flex-col sm:flex-row items-center justify-between gap-3 px-4 py-3 border-t text-sm text-muted-foreground">
                  <span className="shrink-0">{t('waterMeters.showing', { from: (page - 1) * perPage + 1, to: Math.min(page * perPage, total), total })}</span>
                  <div className="flex items-center gap-4">
                    <div className="flex items-center gap-2">
                      <span className="hidden sm:inline shrink-0">{t('waterMeters.perPage')}</span>
                      <Select value={String(perPage)} onValueChange={(v) => { setPerPage(Number(v)); setPage(1) }}>
                        <SelectTrigger className="h-8 w-16"><SelectValue /></SelectTrigger>
                        <SelectContent>{PER_PAGE_OPTIONS.map((n) => <SelectItem key={n} value={String(n)}>{n}</SelectItem>)}</SelectContent>
                      </Select>
                    </div>
                    <span className="shrink-0">{t('waterMeters.page')} {page} {t('waterMeters.of')} {totalPages}</span>
                    <div className="flex items-center gap-1">
                      <Button variant="outline" size="icon" className="h-8 w-8" onClick={() => setPage(1)} disabled={page === 1}><ChevronsLeft className="w-4 h-4" /></Button>
                      <Button variant="outline" size="icon" className="h-8 w-8" onClick={() => setPage((p) => p - 1)} disabled={page === 1}><ChevronLeft className="w-4 h-4" /></Button>
                      <Button variant="outline" size="icon" className="h-8 w-8" onClick={() => setPage((p) => p + 1)} disabled={page >= totalPages}><ChevronRight className="w-4 h-4" /></Button>
                      <Button variant="outline" size="icon" className="h-8 w-8" onClick={() => setPage(totalPages)} disabled={page >= totalPages}><ChevronsRight className="w-4 h-4" /></Button>
                    </div>
                  </div>
                </div>
              )}
            </>
          )}
        </CardContent>
      </Card>

      <WaterMeterDialog open={dialogOpen} onOpenChange={setDialogOpen} editing={editing}
        form={form} onFormChange={setForm} onSave={handleSave} saving={saving} rooms={rooms} />
      <WaterMeterDeleteDialog open={deleteDialogOpen} onOpenChange={setDeleteDialogOpen}
        reading={deleting} onDelete={handleDelete} />
    </div>
  )
}
