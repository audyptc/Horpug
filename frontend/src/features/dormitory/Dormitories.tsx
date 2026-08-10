import { useState, useMemo, useEffect, useCallback } from 'react'
import { useTranslation } from 'react-i18next'
import {
  Search,
  Plus,
  MoreHorizontal,
  Pencil,
  Trash2,
  Building2,
  SearchX,
  RefreshCw,
} from 'lucide-react'
import { Button } from '@/components/ui/button'
import { Input } from '@/components/ui/input'
import { Label } from '@/components/ui/label'
import { Badge } from '@/components/ui/badge'
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
import {
  Select,
  SelectContent,
  SelectItem,
  SelectTrigger,
  SelectValue,
} from '@/components/ui/select'
import { dormitoryService } from '@/features/dormitory/dormitoryService'
import type { ApiDormitory } from '@/types'
import { cn } from '@/lib/utils'
import { formatDate } from '@/lib/dateUtils'
import { usePermission } from '@/hooks/usePermission'

const emptyForm = { name: '', address: '', is_active: true }

export function Dormitories() {
  const { t } = useTranslation()
  const { canCreate, canUpdate, canDelete } = usePermission('/settings/dormitories')
  const [dormitories, setDormitories] = useState<ApiDormitory[]>([])
  const [loading, setLoading] = useState(true)
  const [error, setError] = useState('')
  const [search, setSearch] = useState('')
  const [dialogOpen, setDialogOpen] = useState(false)
  const [deleteDialogOpen, setDeleteDialogOpen] = useState(false)
  const [editingDormitory, setEditingDormitory] = useState<ApiDormitory | null>(null)
  const [deletingDormitory, setDeletingDormitory] = useState<ApiDormitory | null>(null)
  const [form, setForm] = useState(emptyForm)
  const [saving, setSaving] = useState(false)
  const [saveError, setSaveError] = useState('')

  const load = useCallback(async () => {
    setLoading(true)
    setError('')
    try {
      setDormitories(await dormitoryService.list())
    } catch {
      setError(t('dormitories.loadError'))
    } finally {
      setLoading(false)
    }
  }, [t])

  useEffect(() => { load() }, [load])

  const filtered = useMemo(() => {
    if (!search) return dormitories
    return dormitories.filter(
      (d) =>
        d.name.toLowerCase().includes(search.toLowerCase()) ||
        d.address.toLowerCase().includes(search.toLowerCase())
    )
  }, [dormitories, search])

  function openCreate() {
    setEditingDormitory(null)
    setForm(emptyForm)
    setSaveError('')
    setDialogOpen(true)
  }

  function openEdit(d: ApiDormitory) {
    setEditingDormitory(d)
    setForm({ name: d.name, address: d.address, is_active: d.is_active })
    setSaveError('')
    setDialogOpen(true)
  }

  async function handleSave() {
    if (!form.name.trim()) return
    setSaving(true)
    setSaveError('')
    try {
      if (editingDormitory) {
        await dormitoryService.update(editingDormitory.id, form)
      } else {
        await dormitoryService.create({ name: form.name, address: form.address })
      }
      setDialogOpen(false)
      await load()
    } catch (err: unknown) {
      const msg = (err as { response?: { data?: { message?: string } } })?.response?.data?.message
      setSaveError(msg ?? t('dormitories.saveError'))
    } finally {
      setSaving(false)
    }
  }

  async function handleDelete() {
    if (!deletingDormitory) return
    try {
      await dormitoryService.delete(deletingDormitory.id)
      await load()
    } catch {
      // ignore
    } finally {
      setDeleteDialogOpen(false)
      setDeletingDormitory(null)
    }
  }

  const isSaveDisabled = saving || !form.name.trim()

  return (
    <div className="space-y-6">
      <div className="flex flex-wrap items-start justify-between gap-3">
        <div className="min-w-0">
          <h1 className="text-2xl font-bold tracking-tight">{t('dormitories.title')}</h1>
          <p className="text-muted-foreground text-sm mt-1">
            {t('dormitories.subtitle', { count: filtered.length, total: dormitories.length })}
          </p>
        </div>
        <div className="flex gap-2 shrink-0">
          <Button variant="outline" size="icon" onClick={load} disabled={loading}>
            <RefreshCw className={cn('w-4 h-4', loading && 'animate-spin')} />
          </Button>
          {canCreate && (
            <Button onClick={openCreate} className="gap-2">
              <Plus className="w-4 h-4" />
              <span className="hidden sm:inline">{t('dormitories.addDormitory')}</span>
            </Button>
          )}
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
              <CardTitle>{t('dormitories.listTitle')}</CardTitle>
              <CardDescription>{t('dormitories.listDesc')}</CardDescription>
            </div>
            <div className="relative">
              <Search className="absolute left-3 top-1/2 -translate-y-1/2 w-4 h-4 text-muted-foreground" />
              <Input
                placeholder={t('dormitories.searchPlaceholder')}
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
                      <th className="text-left px-6 py-3 font-medium text-muted-foreground">{t('dormitories.colName')}</th>
                      <th className="text-left px-4 py-3 font-medium text-muted-foreground">{t('dormitories.colAddress')}</th>
                      <th className="text-left px-4 py-3 font-medium text-muted-foreground">{t('dormitories.colStatus')}</th>
                      <th className="text-left px-4 py-3 font-medium text-muted-foreground">{t('dormitories.colCreated')}</th>
                      {(canUpdate || canDelete) && (
                        <th className="text-right px-6 py-3 font-medium text-muted-foreground">{t('dormitories.colActions')}</th>
                      )}
                    </tr>
                  </thead>
                  <tbody>
                    {filtered.map((d, i) => (
                      <tr
                        key={d.id}
                        className={cn(
                          'border-b transition-colors hover:bg-muted/30',
                          i === filtered.length - 1 && 'border-0'
                        )}
                      >
                        <td className="px-6 py-4">
                          <div className="flex items-center gap-2">
                            <div className="h-8 w-8 rounded-lg bg-muted flex items-center justify-center shrink-0">
                              <Building2 className="w-4 h-4 text-muted-foreground" />
                            </div>
                            <span className="font-medium">{d.name}</span>
                          </div>
                        </td>
                        <td className="px-4 py-4 text-muted-foreground">{d.address || '—'}</td>
                        <td className="px-4 py-4">
                          <Badge variant={d.is_active ? 'success' : 'secondary'}>
                            {d.is_active ? t('dormitories.statusActive') : t('dormitories.statusInactive')}
                          </Badge>
                        </td>
                        <td className="px-4 py-4 text-muted-foreground">{formatDate(d.created_at)}</td>
                        {(canUpdate || canDelete) && (
                          <td className="px-6 py-4 text-right">
                            <DropdownMenu>
                              <DropdownMenuTrigger asChild>
                                <Button variant="ghost" size="icon" className="h-8 w-8">
                                  <MoreHorizontal className="w-4 h-4" />
                                </Button>
                              </DropdownMenuTrigger>
                              <DropdownMenuContent align="end">
                                {canUpdate && (
                                  <DropdownMenuItem onClick={() => openEdit(d)} className="gap-2">
                                    <Pencil className="w-4 h-4" /> {t('common.edit')}
                                  </DropdownMenuItem>
                                )}
                                {canUpdate && canDelete && <DropdownMenuSeparator />}
                                {canDelete && (
                                  <DropdownMenuItem
                                    className="gap-2 text-destructive focus:text-destructive"
                                    onClick={() => { setDeletingDormitory(d); setDeleteDialogOpen(true) }}
                                  >
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
                    <div className="p-4 rounded-full bg-muted">
                      <SearchX className="w-6 h-6" />
                    </div>
                    <p className="text-sm font-medium">{t('dormitories.noDormitories')}</p>
                    {search && (
                      <button
                        type="button"
                        onClick={() => setSearch('')}
                        className="text-xs text-primary hover:underline"
                      >
                        {t('dormitories.clearSearch')}
                      </button>
                    )}
                  </div>
                )}
              </div>

              {/* Mobile cards */}
              <div className="md:hidden divide-y">
                {filtered.map((d) => (
                  <div key={d.id} className="p-4 flex items-center gap-3">
                    <div className="h-10 w-10 shrink-0 rounded-full bg-muted flex items-center justify-center">
                      <Building2 className="w-5 h-5 text-muted-foreground" />
                    </div>
                    <div className="flex-1 min-w-0">
                      <p className="text-sm font-medium">{d.name}</p>
                      <p className="text-xs text-muted-foreground mt-0.5 truncate">{d.address || '—'}</p>
                      <div className="mt-1">
                        <Badge variant={d.is_active ? 'success' : 'secondary'} className="text-xs">
                          {d.is_active ? t('dormitories.statusActive') : t('dormitories.statusInactive')}
                        </Badge>
                      </div>
                    </div>
                    {(canUpdate || canDelete) && (
                      <DropdownMenu>
                        <DropdownMenuTrigger asChild>
                          <Button variant="ghost" size="icon" className="h-8 w-8 shrink-0">
                            <MoreHorizontal className="w-4 h-4" />
                          </Button>
                        </DropdownMenuTrigger>
                        <DropdownMenuContent align="end">
                          {canUpdate && (
                            <DropdownMenuItem onClick={() => openEdit(d)}>{t('common.edit')}</DropdownMenuItem>
                          )}
                          {canUpdate && canDelete && <DropdownMenuSeparator />}
                          {canDelete && (
                            <DropdownMenuItem
                              className="text-destructive focus:text-destructive"
                              onClick={() => { setDeletingDormitory(d); setDeleteDialogOpen(true) }}
                            >
                              {t('common.delete')}
                            </DropdownMenuItem>
                          )}
                        </DropdownMenuContent>
                      </DropdownMenu>
                    )}
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
              {editingDormitory ? t('dormitories.editDormitory') : t('dormitories.createDormitory')}
            </DialogTitle>
            <DialogDescription>
              {editingDormitory ? t('dormitories.editDesc') : t('dormitories.createDesc')}
            </DialogDescription>
          </DialogHeader>
          <div className="grid gap-4 py-4">
            <div className="space-y-1.5">
              <Label>{t('dormitories.nameField')} *</Label>
              <Input
                placeholder={t('dormitories.namePlaceholder')}
                value={form.name}
                onChange={(e) => setForm((f) => ({ ...f, name: e.target.value }))}
              />
            </div>
            <div className="space-y-1.5">
              <Label>{t('dormitories.addressField')}</Label>
              <Input
                placeholder={t('dormitories.addressPlaceholder')}
                value={form.address}
                onChange={(e) => setForm((f) => ({ ...f, address: e.target.value }))}
              />
            </div>
            {editingDormitory && (
              <div className="space-y-1.5">
                <Label>{t('dormitories.colStatus')}</Label>
                <Select
                  value={form.is_active ? 'active' : 'inactive'}
                  onValueChange={(v) => setForm((f) => ({ ...f, is_active: v === 'active' }))}
                >
                  <SelectTrigger>
                    <SelectValue />
                  </SelectTrigger>
                  <SelectContent>
                    <SelectItem value="active">{t('dormitories.statusActive')}</SelectItem>
                    <SelectItem value="inactive">{t('dormitories.statusInactive')}</SelectItem>
                  </SelectContent>
                </Select>
              </div>
            )}
            {saveError && (
              <p className="text-sm text-destructive">{saveError}</p>
            )}
          </div>
          <DialogFooter>
            <Button variant="outline" onClick={() => setDialogOpen(false)}>{t('common.cancel')}</Button>
            <Button onClick={handleSave} disabled={isSaveDisabled}>
              {saving ? t('common.loading') : editingDormitory ? t('dormitories.saveChanges') : t('dormitories.createDormitory')}
            </Button>
          </DialogFooter>
        </DialogContent>
      </Dialog>

      {/* Delete confirm */}
      <Dialog open={deleteDialogOpen} onOpenChange={setDeleteDialogOpen}>
        <DialogContent className="max-w-sm">
          <DialogHeader>
            <DialogTitle>{t('dormitories.deleteDormitory')}</DialogTitle>
            <DialogDescription>
              {t('dormitories.deleteConfirm')}{' '}
              <span className="font-semibold text-foreground">{deletingDormitory?.name}</span>?{' '}
              {t('dormitories.deleteWarning')}
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
