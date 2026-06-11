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
import { tenantService } from '@/features/tenants/tenantService'
import { useToast } from '@/components/ui/toast'
import type { ApiTenant } from '@/types'
import { cn } from '@/lib/utils'
import { formatDate } from '@/lib/dateUtils'

const PER_PAGE_OPTIONS = [10, 20, 50] as const

const emptyForm = {
  first_name: '',
  last_name: '',
  phone: '',
  id_card: '',
  email: '',
  emergency_contact: '',
  note: '',
}

function getInitials(first: string, last: string) {
  return `${first[0] ?? ''}${last[0] ?? ''}`.toUpperCase()
}

export function Tenants() {
  const { t } = useTranslation()
  const toast = useToast()
  const [tenants, setTenants] = useState<ApiTenant[]>([])
  const [loading, setLoading] = useState(true)
  const [error, setError] = useState('')
  const [search, setSearch] = useState('')
  const [page, setPage] = useState(1)
  const [perPage, setPerPage] = useState<number>(20)
  const [total, setTotal] = useState(0)
  const [totalPages, setTotalPages] = useState(1)
  const [dialogOpen, setDialogOpen] = useState(false)
  const [deleteDialogOpen, setDeleteDialogOpen] = useState(false)
  const [editingTenant, setEditingTenant] = useState<ApiTenant | null>(null)
  const [deletingTenant, setDeletingTenant] = useState<ApiTenant | null>(null)
  const [form, setForm] = useState(emptyForm)
  const [saving, setSaving] = useState(false)
  const [saveError, setSaveError] = useState('')

  const fetchTenants = useCallback(async (p: number, pp: number) => {
    setLoading(true)
    setError('')
    try {
      const resp = await tenantService.list(p, pp)
      setTenants(resp.data)
      setTotal(resp.meta.total)
      setTotalPages(resp.meta.total_pages)
    } catch {
      setError(t('tenants.loadError'))
    } finally {
      setLoading(false)
    }
  }, [t])

  useEffect(() => {
    fetchTenants(page, perPage)
  }, [fetchTenants, page, perPage])

  function handlePerPageChange(value: string) {
    setPerPage(Number(value))
    setPage(1)
  }

  const filtered = useMemo(() => {
    const q = search.toLowerCase()
    return tenants.filter((t) =>
      `${t.first_name} ${t.last_name}`.toLowerCase().includes(q) ||
      t.phone.includes(q) ||
      t.id_card.includes(q)
    )
  }, [tenants, search])

  function openCreate() {
    setEditingTenant(null)
    setForm(emptyForm)
    setSaveError('')
    setDialogOpen(true)
  }

  function openEdit(tenant: ApiTenant) {
    setEditingTenant(tenant)
    setForm({
      first_name: tenant.first_name,
      last_name: tenant.last_name,
      phone: tenant.phone,
      id_card: tenant.id_card,
      email: tenant.email,
      emergency_contact: tenant.emergency_contact,
      note: tenant.note,
    })
    setSaveError('')
    setDialogOpen(true)
  }

  async function handleSave() {
    if (!form.first_name || !form.last_name || !form.phone || !form.id_card) return
    setSaving(true)
    setSaveError('')
    try {
      if (editingTenant) {
        await tenantService.update(editingTenant.id, form)
        toast.success(t('tenants.editSuccess'))
      } else {
        await tenantService.create(form)
        toast.success(t('tenants.createSuccess'))
      }
      setDialogOpen(false)
      await fetchTenants(page, perPage)
    } catch (err: unknown) {
      const msg =
        (err as { response?: { data?: { message?: string } } })?.response?.data?.message ??
        t('tenants.saveError')
      setSaveError(msg)
    } finally {
      setSaving(false)
    }
  }

  async function handleDelete() {
    if (!deletingTenant) return
    try {
      await tenantService.delete(deletingTenant.id)
      toast.success(t('tenants.deleteSuccess'))
      const newPage = tenants.length === 1 && page > 1 ? page - 1 : page
      setPage(newPage)
      await fetchTenants(newPage, perPage)
    } catch (err: unknown) {
      const reason = (err as { response?: { data?: { message?: string } } })?.response?.data?.message
      toast.error(reason ? `${t('tenants.deleteError')}: ${reason}` : t('tenants.deleteError'))
    } finally {
      setDeleteDialogOpen(false)
      setDeletingTenant(null)
    }
  }

  const isSaveDisabled = saving || !form.first_name || !form.last_name || !form.phone || !form.id_card

  return (
    <div className="space-y-6">
      <div className="flex flex-wrap items-start justify-between gap-3">
        <div className="min-w-0">
          <h1 className="text-2xl font-bold tracking-tight">{t('tenants.title')}</h1>
          <p className="text-muted-foreground text-sm mt-1">
            {t('tenants.subtitle', { count: filtered.length, total })}
          </p>
        </div>
        <div className="flex gap-2 shrink-0">
          <Button variant="outline" size="icon" onClick={() => fetchTenants(page, perPage)} disabled={loading}>
            <RefreshCw className={cn('w-4 h-4', loading && 'animate-spin')} />
          </Button>
          <Button onClick={openCreate} className="gap-2">
            <Plus className="w-4 h-4" />
            <span className="hidden sm:inline">{t('tenants.addTenant')}</span>
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
              <CardTitle>{t('tenants.tenantList')}</CardTitle>
              <CardDescription>{t('tenants.tenantListDesc')}</CardDescription>
            </div>
            <div className="relative">
              <Search className="absolute left-3 top-1/2 -translate-y-1/2 w-4 h-4 text-muted-foreground" />
              <Input
                placeholder={t('tenants.searchPlaceholder')}
                className="pl-9 w-full sm:w-64"
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
                      <th className="text-left px-6 py-3 font-medium text-muted-foreground">{t('tenants.colName')}</th>
                      <th className="text-left px-4 py-3 font-medium text-muted-foreground">{t('tenants.colPhone')}</th>
                      <th className="text-left px-4 py-3 font-medium text-muted-foreground">{t('tenants.colIdCard')}</th>
                      <th className="text-left px-4 py-3 font-medium text-muted-foreground">{t('tenants.colEmail')}</th>
                      <th className="text-left px-4 py-3 font-medium text-muted-foreground">{t('tenants.colCreated')}</th>
                      <th className="text-left px-4 py-3 font-medium text-muted-foreground">{t('tenants.colUpdatedAt')}</th>
                      <th className="text-left px-4 py-3 font-medium text-muted-foreground">{t('tenants.colUpdatedBy')}</th>
                      <th className="text-right px-6 py-3 font-medium text-muted-foreground">{t('tenants.colActions')}</th>
                    </tr>
                  </thead>
                  <tbody>
                    {filtered.map((tenant, i) => (
                      <tr
                        key={tenant.id}
                        className={cn(
                          'border-b transition-colors hover:bg-muted/30',
                          i === filtered.length - 1 && 'border-0'
                        )}
                      >
                        <td className="px-6 py-4">
                          <div className="flex items-center gap-3">
                            <div className="h-8 w-8 rounded-full bg-primary/10 flex items-center justify-center shrink-0 text-xs font-semibold text-primary">
                              {getInitials(tenant.first_name, tenant.last_name)}
                            </div>
                            <span className="font-medium">{tenant.first_name} {tenant.last_name}</span>
                          </div>
                        </td>
                        <td className="px-4 py-4 text-muted-foreground">{tenant.phone}</td>
                        <td className="px-4 py-4 text-muted-foreground font-mono text-xs">{tenant.id_card}</td>
                        <td className="px-4 py-4 text-muted-foreground">
                          {tenant.email || <span className="text-xs">—</span>}
                        </td>
                        <td className="px-4 py-4 text-muted-foreground">{formatDate(tenant.created_at)}</td>
                        <td className="px-4 py-4 text-muted-foreground">{formatDate(tenant.updated_at)}</td>
                        <td className="px-4 py-4 text-muted-foreground">
                          {tenant.updated_by_name || <span className="text-xs">—</span>}
                        </td>
                        <td className="px-6 py-4 text-right">
                          <DropdownMenu>
                            <DropdownMenuTrigger asChild>
                              <Button variant="ghost" size="icon" className="h-8 w-8">
                                <MoreHorizontal className="w-4 h-4" />
                              </Button>
                            </DropdownMenuTrigger>
                            <DropdownMenuContent align="end">
                              <DropdownMenuItem onClick={() => openEdit(tenant)} className="gap-2">
                                <Pencil className="w-4 h-4" /> {t('common.edit')}
                              </DropdownMenuItem>
                              <DropdownMenuSeparator />
                              <DropdownMenuItem
                                className="gap-2 text-destructive focus:text-destructive"
                                onClick={() => { setDeletingTenant(tenant); setDeleteDialogOpen(true) }}
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
                    <p className="text-sm font-medium">{t('tenants.noTenants')}</p>
                    {search && (
                      <button
                        type="button"
                        onClick={() => setSearch('')}
                        className="text-xs text-primary hover:underline"
                      >
                        {t('tenants.clearFilters')}
                      </button>
                    )}
                  </div>
                )}
              </div>

              {/* Mobile cards */}
              <div className="md:hidden divide-y">
                {filtered.map((tenant) => (
                  <div key={tenant.id} className="p-4 flex items-center gap-3">
                    <div className="h-10 w-10 shrink-0 rounded-full bg-primary/10 flex items-center justify-center text-sm font-semibold text-primary">
                      {getInitials(tenant.first_name, tenant.last_name)}
                    </div>
                    <div className="flex-1 min-w-0">
                      <p className="text-sm font-medium">{tenant.first_name} {tenant.last_name}</p>
                      <p className="text-xs text-muted-foreground">{tenant.phone}</p>
                      <p className="text-xs text-muted-foreground font-mono">{tenant.id_card}</p>
                    </div>
                    <DropdownMenu>
                      <DropdownMenuTrigger asChild>
                        <Button variant="ghost" size="icon" className="h-8 w-8 shrink-0">
                          <MoreHorizontal className="w-4 h-4" />
                        </Button>
                      </DropdownMenuTrigger>
                      <DropdownMenuContent align="end">
                        <DropdownMenuItem onClick={() => openEdit(tenant)}>{t('common.edit')}</DropdownMenuItem>
                        <DropdownMenuSeparator />
                        <DropdownMenuItem
                          className="text-destructive focus:text-destructive"
                          onClick={() => { setDeletingTenant(tenant); setDeleteDialogOpen(true) }}
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
                    {t('tenants.showing', {
                      from: (page - 1) * perPage + 1,
                      to: Math.min(page * perPage, total),
                      total,
                    })}
                  </span>
                  <div className="flex items-center gap-4">
                    <div className="flex items-center gap-2">
                      <span className="hidden sm:inline shrink-0">{t('tenants.perPage')}</span>
                      <Select value={String(perPage)} onValueChange={handlePerPageChange}>
                        <SelectTrigger className="h-8 w-16"><SelectValue /></SelectTrigger>
                        <SelectContent>
                          {PER_PAGE_OPTIONS.map((n) => (
                            <SelectItem key={n} value={String(n)}>{n}</SelectItem>
                          ))}
                        </SelectContent>
                      </Select>
                    </div>
                    <span className="shrink-0">{t('tenants.page')} {page} {t('tenants.of')} {totalPages}</span>
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
            <DialogTitle>{editingTenant ? t('tenants.editTenant') : t('tenants.createTenant')}</DialogTitle>
            <DialogDescription>
              {editingTenant ? t('tenants.editDesc') : t('tenants.createDesc')}
            </DialogDescription>
          </DialogHeader>
          {saveError && (
            <div className="rounded-md border border-destructive/30 bg-destructive/10 px-4 py-3 text-sm text-destructive">
              {saveError}
            </div>
          )}
          <div className="grid gap-4 py-4">
            <div className="grid grid-cols-2 gap-4">
              <div className="space-y-1.5">
                <Label>{t('tenants.firstName')} *</Label>
                <Input
                  placeholder={t('tenants.firstNamePlaceholder')}
                  value={form.first_name}
                  onChange={(e) => setForm((f) => ({ ...f, first_name: e.target.value }))}
                />
              </div>
              <div className="space-y-1.5">
                <Label>{t('tenants.lastName')} *</Label>
                <Input
                  placeholder={t('tenants.lastNamePlaceholder')}
                  value={form.last_name}
                  onChange={(e) => setForm((f) => ({ ...f, last_name: e.target.value }))}
                />
              </div>
            </div>
            <div className="grid grid-cols-2 gap-4">
              <div className="space-y-1.5">
                <Label>{t('tenants.phone')} *</Label>
                <Input
                  placeholder="0812345678"
                  value={form.phone}
                  onChange={(e) => setForm((f) => ({ ...f, phone: e.target.value }))}
                />
              </div>
              <div className="space-y-1.5">
                <Label>{t('tenants.idCard')} *</Label>
                <Input
                  placeholder="1234567890123"
                  value={form.id_card}
                  onChange={(e) => setForm((f) => ({ ...f, id_card: e.target.value }))}
                />
              </div>
            </div>
            <div className="space-y-1.5">
              <Label>{t('tenants.email')}</Label>
              <Input
                type="email"
                placeholder="example@email.com"
                value={form.email}
                onChange={(e) => setForm((f) => ({ ...f, email: e.target.value }))}
              />
            </div>
            <div className="space-y-1.5">
              <Label>{t('tenants.emergencyContact')}</Label>
              <Input
                placeholder={t('tenants.emergencyContactPlaceholder')}
                value={form.emergency_contact}
                onChange={(e) => setForm((f) => ({ ...f, emergency_contact: e.target.value }))}
              />
            </div>
            <div className="space-y-1.5">
              <Label>{t('tenants.note')}</Label>
              <textarea
                rows={2}
                placeholder={t('tenants.notePlaceholder')}
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
              {saving ? t('common.loading') : editingTenant ? t('tenants.saveChanges') : t('tenants.createTenant')}
            </Button>
          </DialogFooter>
        </DialogContent>
      </Dialog>

      {/* Delete confirm */}
      <Dialog open={deleteDialogOpen} onOpenChange={setDeleteDialogOpen}>
        <DialogContent className="max-w-sm">
          <DialogHeader>
            <DialogTitle>{t('tenants.deleteTenant')}</DialogTitle>
            <DialogDescription>
              {t('tenants.deleteConfirm')}{' '}
              <span className="font-semibold text-foreground">
                {deletingTenant?.first_name} {deletingTenant?.last_name}
              </span>?{' '}
              {t('tenants.deleteWarning')}
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
