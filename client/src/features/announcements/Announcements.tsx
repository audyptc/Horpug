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
  Pin,
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
import { announcementService } from '@/services/announcementService'
import type { ApiAnnouncement, AnnouncementType } from '@/types/api'
import { cn } from '@/lib/utils'
import { formatDate } from '@/lib/dateUtils'

const PER_PAGE_OPTIONS = [10, 20, 50] as const

const TYPE_VARIANTS: Record<AnnouncementType, string> = {
  general: 'secondary',
  maintenance: 'outline',
  payment: 'outline',
  emergency: 'destructive',
}

function toDateInput(iso: string | null | undefined): string {
  if (!iso) return ''
  return iso.slice(0, 10)
}

const emptyForm = {
  title: '',
  content: '',
  type: 'general' as AnnouncementType,
  is_pinned: false,
  published_at: '',
  expired_at: '',
}

export function Announcements() {
  const { t } = useTranslation()

  const [items, setItems] = useState<ApiAnnouncement[]>([])
  const [loading, setLoading] = useState(true)
  const [error, setError] = useState('')
  const [search, setSearch] = useState('')
  const [filterType, setFilterType] = useState<string>('all')
  const [page, setPage] = useState(1)
  const [perPage, setPerPage] = useState<number>(20)
  const [total, setTotal] = useState(0)
  const [totalPages, setTotalPages] = useState(1)

  const [dialogOpen, setDialogOpen] = useState(false)
  const [deleteDialogOpen, setDeleteDialogOpen] = useState(false)
  const [editingItem, setEditingItem] = useState<ApiAnnouncement | null>(null)
  const [deletingItem, setDeletingItem] = useState<ApiAnnouncement | null>(null)
  const [form, setForm] = useState(emptyForm)
  const [saving, setSaving] = useState(false)

  const fetchItems = useCallback(async (p: number, pp: number) => {
    setLoading(true)
    setError('')
    try {
      const resp = await announcementService.list(p, pp)
      setItems(resp.data)
      setTotal(resp.meta.total)
      setTotalPages(resp.meta.total_pages)
    } catch {
      setError(t('announcements.loadError'))
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
      const matchSearch = item.title.toLowerCase().includes(q) || item.content.toLowerCase().includes(q)
      const matchType = filterType === 'all' || item.type === filterType
      return matchSearch && matchType
    })
  }, [items, search, filterType])

  function openCreate() {
    setEditingItem(null)
    setForm({ ...emptyForm, published_at: new Date().toISOString().slice(0, 10) })
    setDialogOpen(true)
  }

  function openEdit(item: ApiAnnouncement) {
    setEditingItem(item)
    setForm({
      title: item.title,
      content: item.content,
      type: item.type,
      is_pinned: item.is_pinned,
      published_at: toDateInput(item.published_at),
      expired_at: toDateInput(item.expired_at),
    })
    setDialogOpen(true)
  }

  async function handleSave() {
    if (!form.title || !form.content || !form.published_at) return
    setSaving(true)
    try {
      const payload = {
        title: form.title,
        content: form.content,
        type: form.type,
        is_pinned: form.is_pinned,
        published_at: form.published_at,
        expired_at: form.expired_at || null,
      }
      if (editingItem) {
        await announcementService.update(editingItem.id, payload)
      } else {
        await announcementService.create(payload)
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
      await announcementService.delete(deletingItem.id)
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

  const isSaveDisabled = saving || !form.title || !form.content || !form.published_at

  return (
    <div className="space-y-6">
      <div className="flex flex-wrap items-start justify-between gap-3">
        <div className="min-w-0">
          <h1 className="text-2xl font-bold tracking-tight">{t('announcements.title')}</h1>
          <p className="text-muted-foreground text-sm mt-1">
            {t('announcements.subtitle', { count: filtered.length, total })}
          </p>
        </div>
        <div className="flex gap-2 shrink-0">
          <Button variant="outline" size="icon" onClick={() => fetchItems(page, perPage)} disabled={loading}>
            <RefreshCw className={cn('w-4 h-4', loading && 'animate-spin')} />
          </Button>
          <Button onClick={openCreate} className="gap-2">
            <Plus className="w-4 h-4" />
            <span className="hidden sm:inline">{t('announcements.addAnnouncement')}</span>
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
              <CardTitle>{t('announcements.announcementList')}</CardTitle>
              <CardDescription>{t('announcements.announcementListDesc')}</CardDescription>
            </div>
            <div className="flex flex-wrap gap-2">
              <Select value={filterType} onValueChange={setFilterType}>
                <SelectTrigger className="h-9 w-36">
                  <SelectValue />
                </SelectTrigger>
                <SelectContent>
                  <SelectItem value="all">{t('announcements.allTypes')}</SelectItem>
                  <SelectItem value="general">{t('announcements.types.general')}</SelectItem>
                  <SelectItem value="maintenance">{t('announcements.types.maintenance')}</SelectItem>
                  <SelectItem value="payment">{t('announcements.types.payment')}</SelectItem>
                  <SelectItem value="emergency">{t('announcements.types.emergency')}</SelectItem>
                </SelectContent>
              </Select>
              <div className="relative">
                <Search className="absolute left-3 top-1/2 -translate-y-1/2 w-4 h-4 text-muted-foreground" />
                <Input
                  placeholder={t('announcements.searchPlaceholder')}
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
                      <th className="text-left px-6 py-3 font-medium text-muted-foreground">{t('announcements.colTitle')}</th>
                      <th className="text-left px-4 py-3 font-medium text-muted-foreground">{t('announcements.colType')}</th>
                      <th className="text-left px-4 py-3 font-medium text-muted-foreground">{t('announcements.colPublished')}</th>
                      <th className="text-left px-4 py-3 font-medium text-muted-foreground">{t('announcements.colExpired')}</th>
                      <th className="text-center px-4 py-3 font-medium text-muted-foreground">{t('announcements.colPinned')}</th>
                      <th className="text-right px-6 py-3 font-medium text-muted-foreground">{t('announcements.colActions')}</th>
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
                        <td className="px-6 py-4">
                          <p className="font-medium">{item.title}</p>
                          <p className="text-xs text-muted-foreground line-clamp-1 mt-0.5">{item.content}</p>
                        </td>
                        <td className="px-4 py-4">
                          <Badge variant={TYPE_VARIANTS[item.type] as 'secondary' | 'outline' | 'destructive'}>
                            {t(`announcements.types.${item.type}`)}
                          </Badge>
                        </td>
                        <td className="px-4 py-4 text-muted-foreground">{formatDate(item.published_at)}</td>
                        <td className="px-4 py-4 text-muted-foreground">{item.expired_at ? formatDate(item.expired_at) : '-'}</td>
                        <td className="px-4 py-4 text-center">
                          {item.is_pinned && <Pin className="w-4 h-4 text-primary mx-auto" />}
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
                    <p className="text-sm font-medium">{t('announcements.noAnnouncements')}</p>
                    {(search || filterType !== 'all') && (
                      <button
                        type="button"
                        onClick={() => { setSearch(''); setFilterType('all') }}
                        className="text-xs text-primary hover:underline"
                      >
                        {t('announcements.clearFilters')}
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
                        <p className="text-sm font-medium truncate">{item.title}</p>
                        {item.is_pinned && <Pin className="w-3 h-3 text-primary shrink-0" />}
                      </div>
                      <p className="text-xs text-muted-foreground line-clamp-2">{item.content}</p>
                      <div className="flex items-center gap-2">
                        <Badge variant={TYPE_VARIANTS[item.type] as 'secondary' | 'outline' | 'destructive'} className="text-xs">
                          {t(`announcements.types.${item.type}`)}
                        </Badge>
                        <span className="text-xs text-muted-foreground">{formatDate(item.published_at)}</span>
                      </div>
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
                    {t('announcements.showing', {
                      from: (page - 1) * perPage + 1,
                      to: Math.min(page * perPage, total),
                      total,
                    })}
                  </span>
                  <div className="flex items-center gap-4">
                    <div className="flex items-center gap-2">
                      <span className="hidden sm:inline shrink-0">{t('announcements.perPage')}</span>
                      <Select value={String(perPage)} onValueChange={handlePerPageChange}>
                        <SelectTrigger className="h-8 w-16"><SelectValue /></SelectTrigger>
                        <SelectContent>
                          {PER_PAGE_OPTIONS.map((n) => (
                            <SelectItem key={n} value={String(n)}>{n}</SelectItem>
                          ))}
                        </SelectContent>
                      </Select>
                    </div>
                    <span className="shrink-0">{t('announcements.page')} {page} {t('announcements.of')} {totalPages}</span>
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
              {editingItem ? t('announcements.editAnnouncement') : t('announcements.createAnnouncement')}
            </DialogTitle>
            <DialogDescription>
              {editingItem ? t('announcements.editDesc') : t('announcements.createDesc')}
            </DialogDescription>
          </DialogHeader>

          <div className="grid gap-4 py-4">
            <div className="space-y-1.5">
              <Label>{t('announcements.titleField')} *</Label>
              <Input
                placeholder={t('announcements.titlePlaceholder')}
                value={form.title}
                onChange={(e) => setForm((f) => ({ ...f, title: e.target.value }))}
              />
            </div>

            <div className="space-y-1.5">
              <Label>{t('announcements.contentField')} *</Label>
              <textarea
                rows={4}
                placeholder={t('announcements.contentPlaceholder')}
                value={form.content}
                onChange={(e: React.ChangeEvent<HTMLTextAreaElement>) =>
                  setForm((f) => ({ ...f, content: e.target.value }))
                }
                className="flex min-h-24 w-full rounded-md border border-input bg-background px-3 py-2 text-sm ring-offset-background placeholder:text-muted-foreground focus-visible:outline-none focus-visible:ring-2 focus-visible:ring-ring focus-visible:ring-offset-2 resize-none"
              />
            </div>

            <div className="grid grid-cols-2 gap-4">
              <div className="space-y-1.5">
                <Label>{t('announcements.typeField')} *</Label>
                <Select
                  value={form.type}
                  onValueChange={(v) => setForm((f) => ({ ...f, type: v as AnnouncementType }))}
                >
                  <SelectTrigger><SelectValue /></SelectTrigger>
                  <SelectContent>
                    <SelectItem value="general">{t('announcements.types.general')}</SelectItem>
                    <SelectItem value="maintenance">{t('announcements.types.maintenance')}</SelectItem>
                    <SelectItem value="payment">{t('announcements.types.payment')}</SelectItem>
                    <SelectItem value="emergency">{t('announcements.types.emergency')}</SelectItem>
                  </SelectContent>
                </Select>
              </div>
              <div className="space-y-1.5">
                <Label>{t('announcements.publishedAt')} *</Label>
                <Input
                  type="date"
                  value={form.published_at}
                  onChange={(e) => setForm((f) => ({ ...f, published_at: e.target.value }))}
                />
              </div>
            </div>

            <div className="space-y-1.5">
              <Label>{t('announcements.expiredAt')}</Label>
              <Input
                type="date"
                value={form.expired_at}
                onChange={(e) => setForm((f) => ({ ...f, expired_at: e.target.value }))}
              />
            </div>

            <div className="flex items-center gap-2">
              <input
                id="is_pinned"
                type="checkbox"
                checked={form.is_pinned}
                onChange={(e) => setForm((f) => ({ ...f, is_pinned: e.target.checked }))}
                className="h-4 w-4 rounded border-input accent-primary cursor-pointer"
              />
              <Label htmlFor="is_pinned" className="cursor-pointer">{t('announcements.isPinned')}</Label>
            </div>
          </div>

          <DialogFooter>
            <Button variant="outline" onClick={() => setDialogOpen(false)}>
              {t('common.cancel')}
            </Button>
            <Button onClick={handleSave} disabled={isSaveDisabled}>
              {saving ? t('common.loading') : editingItem ? t('announcements.saveChanges') : t('announcements.createAnnouncement')}
            </Button>
          </DialogFooter>
        </DialogContent>
      </Dialog>

      {/* Delete confirm */}
      <Dialog open={deleteDialogOpen} onOpenChange={setDeleteDialogOpen}>
        <DialogContent className="max-w-sm">
          <DialogHeader>
            <DialogTitle>{t('announcements.deleteAnnouncement')}</DialogTitle>
            <DialogDescription>
              {t('announcements.deleteConfirm')}{' '}
              <span className="font-semibold text-foreground">{deletingItem?.title}</span>?{' '}
              {t('announcements.deleteWarning')}
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
