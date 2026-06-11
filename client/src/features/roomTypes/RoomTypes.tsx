import { useState, useMemo, useEffect, useCallback } from 'react'
import { useTranslation } from 'react-i18next'
import {
  Search,
  Plus,
  MoreHorizontal,
  Pencil,
  Trash2,
  LayoutList,
  SearchX,
  RefreshCw,
} from 'lucide-react'
import { Button } from '@/components/ui/button'
import { Input } from '@/components/ui/input'
import { Label } from '@/components/ui/label'
import { Card, CardContent, CardHeader, CardTitle, CardDescription } from '@/components/ui/card'
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
import { roomTypeService } from '@/features/roomTypes/roomTypeService'
import type { ApiRoomType } from '@/types'
import { cn } from '@/lib/utils'
import { formatDate } from '@/lib/dateUtils'

const emptyForm = { id: '', name: '', sort_order: 0 }

export function RoomTypes() {
  const { t } = useTranslation()
  const [types, setTypes] = useState<ApiRoomType[]>([])
  const [loading, setLoading] = useState(true)
  const [error, setError] = useState('')
  const [search, setSearch] = useState('')
  const [dialogOpen, setDialogOpen] = useState(false)
  const [deleteDialogOpen, setDeleteDialogOpen] = useState(false)
  const [editingType, setEditingType] = useState<ApiRoomType | null>(null)
  const [deletingType, setDeletingType] = useState<ApiRoomType | null>(null)
  const [form, setForm] = useState(emptyForm)
  const [saving, setSaving] = useState(false)
  const [saveError, setSaveError] = useState('')

  const load = useCallback(async () => {
    setLoading(true)
    setError('')
    try {
      setTypes(await roomTypeService.list())
    } catch {
      setError(t('roomTypes.loadError'))
    } finally {
      setLoading(false)
    }
  }, [t])

  useEffect(() => { load() }, [load])

  const filtered = useMemo(() => {
    if (!search) return types
    return types.filter(
      (rt) =>
        rt.name.toLowerCase().includes(search.toLowerCase()) ||
        rt.id.toLowerCase().includes(search.toLowerCase())
    )
  }, [types, search])

  function openCreate() {
    setEditingType(null)
    setForm({ ...emptyForm, sort_order: types.length + 1 })
    setSaveError('')
    setDialogOpen(true)
  }

  function openEdit(rt: ApiRoomType) {
    setEditingType(rt)
    setForm({ id: rt.id, name: rt.name, sort_order: rt.sort_order })
    setSaveError('')
    setDialogOpen(true)
  }

  async function handleSave() {
    if (!form.name.trim()) return
    if (!editingType && !form.id.trim()) return
    setSaving(true)
    setSaveError('')
    try {
      if (editingType) {
        await roomTypeService.update(editingType.id, { name: form.name, sort_order: form.sort_order })
      } else {
        await roomTypeService.create({ id: form.id, name: form.name, sort_order: form.sort_order })
      }
      setDialogOpen(false)
      await load()
    } catch (err: unknown) {
      const msg = (err as { response?: { data?: { message?: string } } })?.response?.data?.message
      setSaveError(msg ?? t('roomTypes.saveError'))
    } finally {
      setSaving(false)
    }
  }

  async function handleDelete() {
    if (!deletingType) return
    try {
      await roomTypeService.delete(deletingType.id)
      await load()
    } catch {
      // ignore
    } finally {
      setDeleteDialogOpen(false)
      setDeletingType(null)
    }
  }

  const isSaveDisabled = saving || !form.name.trim() || (!editingType && !form.id.trim())

  return (
    <div className="space-y-6">
      <div className="flex flex-wrap items-start justify-between gap-3">
        <div className="min-w-0">
          <h1 className="text-2xl font-bold tracking-tight">{t('roomTypes.title')}</h1>
          <p className="text-muted-foreground text-sm mt-1">
            {t('roomTypes.subtitle', { count: filtered.length, total: types.length })}
          </p>
        </div>
        <div className="flex gap-2 shrink-0">
          <Button variant="outline" size="icon" onClick={load} disabled={loading}>
            <RefreshCw className={cn('w-4 h-4', loading && 'animate-spin')} />
          </Button>
          <Button onClick={openCreate} className="gap-2">
            <Plus className="w-4 h-4" />
            <span className="hidden sm:inline">{t('roomTypes.addType')}</span>
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
              <CardTitle>{t('roomTypes.listTitle')}</CardTitle>
              <CardDescription>{t('roomTypes.listDesc')}</CardDescription>
            </div>
            <div className="relative">
              <Search className="absolute left-3 top-1/2 -translate-y-1/2 w-4 h-4 text-muted-foreground" />
              <Input
                placeholder={t('roomTypes.searchPlaceholder')}
                className="pl-9 w-full sm:w-56"
                value={search}
                onChange={(e) => setSearch(e.target.value)}
              />
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
                      <th className="text-left px-4 py-3 font-medium text-muted-foreground">{t('roomTypes.colId')}</th>
                          <th className="text-left px-4 py-3 font-medium text-muted-foreground">{t('roomTypes.colOrder')}</th>
                      <th className="text-left px-6 py-3 font-medium text-muted-foreground">{t('roomTypes.colName')}</th>
                    
                  
                      <th className="text-left px-4 py-3 font-medium text-muted-foreground">{t('roomTypes.colCreated')}</th>
                      <th className="text-right px-6 py-3 font-medium text-muted-foreground">{t('roomTypes.colActions')}</th>
                    </tr>
                  </thead>
                  <tbody>
                    {filtered.map((rt, i) => (
                      <tr
                        key={rt.id}
                        className={cn(
                          'border-b transition-colors hover:bg-muted/30',
                          i === filtered.length - 1 && 'border-0'
                        )}
                      >
                        <td className="px-6 py-4">
                          <div className="flex items-center gap-2">
                            <div className="h-8 w-8 rounded-lg bg-muted flex items-center justify-center shrink-0">
                              <LayoutList className="w-4 h-4 text-muted-foreground" />
                            </div>
                            <span className="font-medium">{rt.name}</span>
                          </div>
                        </td>
                        <td className="px-4 py-4">
                          <code className="text-xs font-mono bg-muted px-1.5 py-0.5 rounded">{rt.id}</code>
                        </td>
                        <td className="px-4 py-4 text-muted-foreground">{rt.sort_order}</td>
                        <td className="px-4 py-4 text-muted-foreground">{formatDate(rt.created_at)}</td>
                        <td className="px-6 py-4 text-right">
                          <DropdownMenu>
                            <DropdownMenuTrigger asChild>
                              <Button variant="ghost" size="icon" className="h-8 w-8">
                                <MoreHorizontal className="w-4 h-4" />
                              </Button>
                            </DropdownMenuTrigger>
                            <DropdownMenuContent align="end">
                              <DropdownMenuItem onClick={() => openEdit(rt)} className="gap-2">
                                <Pencil className="w-4 h-4" /> {t('common.edit')}
                              </DropdownMenuItem>
                              <DropdownMenuSeparator />
                              <DropdownMenuItem
                                className="gap-2 text-destructive focus:text-destructive"
                                onClick={() => { setDeletingType(rt); setDeleteDialogOpen(true) }}
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
                    <p className="text-sm font-medium">{t('roomTypes.noTypes')}</p>
                    {search && (
                      <button
                        type="button"
                        onClick={() => setSearch('')}
                        className="text-xs text-primary hover:underline"
                      >
                        {t('roomTypes.clearSearch')}
                      </button>
                    )}
                  </div>
                )}
              </div>

              {/* Mobile cards */}
              <div className="md:hidden divide-y">
                {filtered.map((rt) => (
                  <div key={rt.id} className="p-4 flex items-center gap-3">
                    <div className="h-10 w-10 shrink-0 rounded-full bg-muted flex items-center justify-center">
                      <LayoutList className="w-5 h-5 text-muted-foreground" />
                    </div>
                    <div className="flex-1 min-w-0">
                      <p className="text-sm font-medium">{rt.name}</p>
                      <p className="text-xs text-muted-foreground mt-0.5">
                        <code className="font-mono">{rt.id}</code>
                        {' · '}{t('roomTypes.orderLabel')}: {rt.sort_order}
                      </p>
                    </div>
                    <DropdownMenu>
                      <DropdownMenuTrigger asChild>
                        <Button variant="ghost" size="icon" className="h-8 w-8 shrink-0">
                          <MoreHorizontal className="w-4 h-4" />
                        </Button>
                      </DropdownMenuTrigger>
                      <DropdownMenuContent align="end">
                        <DropdownMenuItem onClick={() => openEdit(rt)}>{t('common.edit')}</DropdownMenuItem>
                        <DropdownMenuSeparator />
                        <DropdownMenuItem
                          className="text-destructive focus:text-destructive"
                          onClick={() => { setDeletingType(rt); setDeleteDialogOpen(true) }}
                        >
                          {t('common.delete')}
                        </DropdownMenuItem>
                      </DropdownMenuContent>
                    </DropdownMenu>
                  </div>
                ))}
              </div>
            </>
          )}
        </CardContent>
      </Card>

      {/* Create / Edit Dialog */}
      <Dialog open={dialogOpen} onOpenChange={setDialogOpen}>
        <DialogContent className="max-w-sm">
          <DialogHeader>
            <DialogTitle>
              {editingType ? t('roomTypes.editType') : t('roomTypes.createType')}
            </DialogTitle>
            <DialogDescription>
              {editingType ? t('roomTypes.editDesc') : t('roomTypes.createDesc')}
            </DialogDescription>
          </DialogHeader>
          <div className="grid gap-4 py-4">
            {!editingType && (
              <div className="space-y-1.5">
                <Label>{t('roomTypes.idField')} *</Label>
                <Input
                  placeholder="standard"
                  value={form.id}
                  onChange={(e) => setForm((f) => ({ ...f, id: e.target.value }))}
                />
                <p className="text-xs text-muted-foreground">{t('roomTypes.idHint')}</p>
              </div>
            )}
            <div className="space-y-1.5">
              <Label>{t('roomTypes.nameField')} *</Label>
              <Input
                placeholder={t('roomTypes.namePlaceholder')}
                value={form.name}
                onChange={(e) => setForm((f) => ({ ...f, name: e.target.value }))}
              />
            </div>
            <div className="space-y-1.5">
              <Label>{t('roomTypes.sortOrderField')}</Label>
              <Input
                type="number"
                min={0}
                value={form.sort_order}
                onChange={(e) => setForm((f) => ({ ...f, sort_order: Number(e.target.value) }))}
              />
            </div>
            {saveError && (
              <p className="text-sm text-destructive">{saveError}</p>
            )}
          </div>
          <DialogFooter>
            <Button variant="outline" onClick={() => setDialogOpen(false)}>{t('common.cancel')}</Button>
            <Button onClick={handleSave} disabled={isSaveDisabled}>
              {saving ? t('common.loading') : editingType ? t('roomTypes.saveChanges') : t('roomTypes.createType')}
            </Button>
          </DialogFooter>
        </DialogContent>
      </Dialog>

      {/* Delete confirm */}
      <Dialog open={deleteDialogOpen} onOpenChange={setDeleteDialogOpen}>
        <DialogContent className="max-w-sm">
          <DialogHeader>
            <DialogTitle>{t('roomTypes.deleteType')}</DialogTitle>
            <DialogDescription>
              {t('roomTypes.deleteConfirm')}{' '}
              <span className="font-semibold text-foreground">{deletingType?.name}</span>?{' '}
              {t('roomTypes.deleteWarning')}
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
