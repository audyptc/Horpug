import { useState, useMemo, useEffect, useCallback } from 'react'
import { useTranslation } from 'react-i18next'
import {
  Search,
  Plus,
  MoreHorizontal,
  Pencil,
  Trash2,
  ShieldCheck,
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
import { roleService } from '@/features/roles/roleService'
import { useToast } from '@/components/ui/toast'
import type { ApiRole, ApiMenu, ApiPermission } from '@/types'
import { cn } from '@/lib/utils'
import { formatDate } from '@/lib/dateUtils'

type StatusFilter = 'all' | 'active' | 'inactive'

const PER_PAGE_OPTIONS = [10, 20, 50] as const

const statusVariant: Record<'active' | 'inactive', 'success' | 'secondary'> = {
  active: 'success',
  inactive: 'secondary',
}

const roleColorMap: Record<string, string> = {
  admin: 'text-violet-600 bg-violet-500/10',
  manager: 'text-blue-600 bg-blue-500/10',
  editor: 'text-amber-600 bg-amber-500/10',
  viewer: 'text-slate-600 bg-slate-500/10',
}

function getRoleColor(roleName: string) {
  return roleColorMap[roleName.toLowerCase()] ?? 'text-slate-600 bg-slate-500/10'
}

const emptyForm = {
  name: '',
  description: '',
  is_active: true,
}

// matrix: menuId -> Set of permissionIds
type PermMatrix = Record<string, Set<string>>

export function Roles() {
  const { t } = useTranslation()
  const toast = useToast()

  const [roles, setRoles] = useState<ApiRole[]>([])
  const [loading, setLoading] = useState(true)
  const [error, setError] = useState('')
  const [search, setSearch] = useState('')
  const [statusFilter, setStatusFilter] = useState<StatusFilter>('all')
  const [page, setPage] = useState(1)
  const [perPage, setPerPage] = useState<number>(20)
  const [total, setTotal] = useState(0)
  const [totalPages, setTotalPages] = useState(1)
  const [dialogOpen, setDialogOpen] = useState(false)
  const [deleteDialogOpen, setDeleteDialogOpen] = useState(false)
  const [editingRole, setEditingRole] = useState<ApiRole | null>(null)
  const [deletingRole, setDeletingRole] = useState<ApiRole | null>(null)
  const [form, setForm] = useState(emptyForm)
  const [saving, setSaving] = useState(false)

  // Permission dialog state
  const [permDialogOpen, setPermDialogOpen] = useState(false)
  const [permRole, setPermRole] = useState<ApiRole | null>(null)
  const [allMenus, setAllMenus] = useState<ApiMenu[]>([])
  const [allPermissions, setAllPermissions] = useState<ApiPermission[]>([])
  const [matrix, setMatrix] = useState<PermMatrix>({})
  const [permLoading, setPermLoading] = useState(false)
  const [permSaving, setPermSaving] = useState(false)
  const [permError, setPermError] = useState('')

  const fetchRoles = useCallback(async (p: number, pp: number) => {
    setLoading(true)
    setError('')
    try {
      const resp = await roleService.list(p, pp)
      setRoles(resp.data)
      setTotal(resp.meta.total)
      setTotalPages(resp.meta.total_pages)
    } catch {
      setError(t('roles.loadError'))
    } finally {
      setLoading(false)
    }
  }, [t])

  useEffect(() => {
    fetchRoles(page, perPage)
  }, [fetchRoles, page, perPage])

  function handlePerPageChange(value: string) {
    setPerPage(Number(value))
    setPage(1)
  }

  const filtered = useMemo(() => {
    return roles.filter((r) => {
      const matchSearch =
        r.name.toLowerCase().includes(search.toLowerCase()) ||
        r.description.toLowerCase().includes(search.toLowerCase())
      const matchStatus =
        statusFilter === 'all' ||
        (statusFilter === 'active' ? r.is_active : !r.is_active)
      return matchSearch && matchStatus
    })
  }, [roles, search, statusFilter])

  function openCreate() {
    setEditingRole(null)
    setForm(emptyForm)
    setDialogOpen(true)
  }

  function openEdit(role: ApiRole) {
    setEditingRole(role)
    setForm({
      name: role.name,
      description: role.description,
      is_active: role.is_active,
    })
    setDialogOpen(true)
  }

  async function handleSave() {
    if (!form.name) return
    setSaving(true)
    try {
      if (editingRole) {
        await roleService.update(editingRole.id, {
          name: form.name,
          description: form.description,
          is_active: form.is_active,
        })
      } else {
        await roleService.create({
          name: form.name,
          description: form.description,
          is_active: form.is_active,
        })
      }
      setDialogOpen(false)
      await fetchRoles(page, perPage)
    } catch {
      // error handled silently
    } finally {
      setSaving(false)
    }
  }

  async function handleDelete() {
    if (!deletingRole) return
    try {
      await roleService.delete(deletingRole.id)
      const newPage = roles.length === 1 && page > 1 ? page - 1 : page
      setPage(newPage)
      await fetchRoles(newPage, perPage)
    } catch {
      // ignore
    } finally {
      setDeleteDialogOpen(false)
      setDeletingRole(null)
    }
  }

  async function openPermissions(role: ApiRole) {
    setPermRole(role)
    setPermDialogOpen(true)
    setPermLoading(true)
    setPermError('')
    try {
      const [menus, permissions, roleDetail] = await Promise.all([
        roleService.listMenus(),
        roleService.listPermissions(),
        roleService.getById(role.id),
      ])
      setAllMenus(menus)
      setAllPermissions(permissions)

      // Build matrix from role's current menu_permissions
      const m: PermMatrix = {}
      for (const menu of menus) {
        m[menu.id] = new Set<string>()
      }
      for (const rmp of roleDetail.menu_permissions ?? []) {
        if (!m[rmp.menu_id]) m[rmp.menu_id] = new Set()
        for (const p of rmp.permissions) {
          m[rmp.menu_id].add(p.id)
        }
      }
      setMatrix(m)
    } catch {
      setPermError(t('roles.permLoadError'))
    } finally {
      setPermLoading(false)
    }
  }

  function togglePermission(menuId: string, permId: string) {
    setMatrix((prev) => {
      const next = { ...prev }
      const set = new Set(next[menuId] ?? [])
      if (set.has(permId)) {
        set.delete(permId)
      } else {
        set.add(permId)
      }
      next[menuId] = set
      return next
    })
  }

  function toggleRowAll(menuId: string) {
    setMatrix((prev) => {
      const next = { ...prev }
      const currentSet = next[menuId] ?? new Set()
      const allChecked = allPermissions.every((p) => currentSet.has(p.id))
      if (allChecked) {
        next[menuId] = new Set()
      } else {
        next[menuId] = new Set(allPermissions.map((p) => p.id))
      }
      return next
    })
  }

  function toggleColAll(permId: string) {
    setMatrix((prev) => {
      const next = { ...prev }
      const allChecked = allMenus.every((m) => (next[m.id] ?? new Set()).has(permId))
      for (const menu of allMenus) {
        const set = new Set(next[menu.id] ?? [])
        if (allChecked) {
          set.delete(permId)
        } else {
          set.add(permId)
        }
        next[menu.id] = set
      }
      return next
    })
  }

  async function handleSavePermissions() {
    if (!permRole) return
    setPermSaving(true)
    try {
      const items = Object.entries(matrix)
        .filter(([, permSet]) => permSet.size > 0)
        .map(([menuId, permSet]) => ({
          menu_id: menuId,
          permission_ids: Array.from(permSet),
        }))
      await roleService.assignPermissions(permRole.id, { items })
      toast.success(t('roles.permSaveSuccess'))
      setPermDialogOpen(false)
    } catch {
      toast.error(t('roles.permSaveError'))
    } finally {
      setPermSaving(false)
    }
  }

  return (
    <div className="space-y-6">
      <div className="flex flex-wrap items-start justify-between gap-3">
        <div className="min-w-0">
          <h1 className="text-2xl font-bold tracking-tight">{t('roles.title')}</h1>
          <p className="text-muted-foreground text-sm mt-1">
            {t('roles.subtitle_other', { count: filtered.length, total })}
          </p>
        </div>
        <div className="flex gap-2 shrink-0">
          <Button variant="outline" size="icon" onClick={() => fetchRoles(page, perPage)} disabled={loading}>
            <RefreshCw className={cn('w-4 h-4', loading && 'animate-spin')} />
          </Button>
          <Button onClick={openCreate} className="gap-2">
            <Plus className="w-4 h-4" />
            <span className="hidden sm:inline">{t('roles.addRole')}</span>
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
              <CardTitle>{t('roles.roleList')}</CardTitle>
              <CardDescription>{t('roles.roleListDesc')}</CardDescription>
            </div>
            <div className="flex flex-col sm:flex-row gap-2 sm:items-center">
              <div className="relative">
                <Search className="absolute left-3 top-1/2 -translate-y-1/2 w-4 h-4 text-muted-foreground" />
                <Input
                  placeholder={t('roles.searchPlaceholder')}
                  className="pl-9 w-full sm:w-64"
                  value={search}
                  onChange={(e) => setSearch(e.target.value)}
                />
              </div>
              <Select
                value={statusFilter}
                onValueChange={(v) => setStatusFilter(v as StatusFilter)}
              >
                <SelectTrigger className="w-full sm:w-32">
                  <SelectValue placeholder={t('roles.allStatus')} />
                </SelectTrigger>
                <SelectContent>
                  <SelectItem value="all">{t('roles.allStatus')}</SelectItem>
                  <SelectItem value="active">{t('roles.statuses.active')}</SelectItem>
                  <SelectItem value="inactive">{t('roles.statuses.inactive')}</SelectItem>
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
                      <th className="text-left px-6 py-3 font-medium text-muted-foreground">{t('roles.colName')}</th>
                      <th className="text-left px-4 py-3 font-medium text-muted-foreground">{t('roles.colDescription')}</th>
                      <th className="text-left px-4 py-3 font-medium text-muted-foreground">{t('roles.colStatus')}</th>
                      <th className="text-left px-4 py-3 font-medium text-muted-foreground">{t('roles.colCreated')}</th>
                      <th className="text-right px-6 py-3 font-medium text-muted-foreground">{t('roles.colActions')}</th>
                    </tr>
                  </thead>
                  <tbody>
                    {filtered.map((role, i) => (
                      <tr
                        key={role.id}
                        className={cn(
                          'border-b transition-colors hover:bg-muted/30',
                          i === filtered.length - 1 && 'border-0'
                        )}
                      >
                        <td className="px-6 py-4">
                          <div className="flex items-center gap-2">
                            <span
                              className={cn(
                                'inline-flex items-center gap-1 px-2 py-0.5 rounded-md text-xs font-medium',
                                getRoleColor(role.name)
                              )}
                            >
                              {role.name.toLowerCase() === 'admin' && <ShieldCheck className="w-3 h-3" />}
                              {role.name}
                            </span>
                          </div>
                        </td>
                        <td className="px-4 py-4 text-muted-foreground max-w-xs truncate">
                          {role.description || <span className="text-xs">—</span>}
                        </td>
                        <td className="px-4 py-4">
                          <Badge variant={role.is_active ? statusVariant.active : statusVariant.inactive}>
                            {role.is_active ? t('roles.statuses.active') : t('roles.statuses.inactive')}
                          </Badge>
                        </td>
                        <td className="px-4 py-4 text-muted-foreground">{formatDate(role.created_at)}</td>
                        <td className="px-6 py-4 text-right">
                          <DropdownMenu>
                            <DropdownMenuTrigger asChild>
                              <Button variant="ghost" size="icon" className="h-8 w-8">
                                <MoreHorizontal className="w-4 h-4" />
                              </Button>
                            </DropdownMenuTrigger>
                            <DropdownMenuContent align="end">
                              <DropdownMenuItem onClick={() => openPermissions(role)} className="gap-2">
                                <ShieldCheck className="w-4 h-4" /> {t('roles.managePermissions')}
                              </DropdownMenuItem>
                              <DropdownMenuSeparator />
                              <DropdownMenuItem onClick={() => openEdit(role)} className="gap-2">
                                <Pencil className="w-4 h-4" /> {t('common.edit')}
                              </DropdownMenuItem>
                              <DropdownMenuSeparator />
                              <DropdownMenuItem
                                className="gap-2 text-destructive focus:text-destructive"
                                onClick={() => {
                                  setDeletingRole(role)
                                  setDeleteDialogOpen(true)
                                }}
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
                    <p className="text-sm font-medium">{t('roles.noRoles')}</p>
                    {(search || statusFilter !== 'all') && (
                      <button
                        type="button"
                        onClick={() => { setSearch(''); setStatusFilter('all') }}
                        className="text-xs text-primary hover:underline"
                      >
                        {t('roles.clearFilters')}
                      </button>
                    )}
                  </div>
                )}
              </div>

              {/* Mobile cards */}
              <div className="md:hidden divide-y">
                {filtered.map((role) => (
                  <div key={role.id} className="p-4 flex items-center gap-3">
                    <div className="h-10 w-10 shrink-0 rounded-full bg-muted flex items-center justify-center">
                      <ShieldCheck className={cn('w-5 h-5', getRoleColor(role.name).split(' ')[0])} />
                    </div>
                    <div className="flex-1 min-w-0">
                      <div className="flex items-center gap-2">
                        <span
                          className={cn(
                            'inline-flex items-center gap-1 px-2 py-0.5 rounded-md text-xs font-medium',
                            getRoleColor(role.name)
                          )}
                        >
                          {role.name}
                        </span>
                      </div>
                      {role.description && (
                        <p className="text-xs text-muted-foreground truncate mt-0.5">{role.description}</p>
                      )}
                      <div className="mt-1">
                        <Badge variant={role.is_active ? statusVariant.active : statusVariant.inactive} className="text-xs">
                          {role.is_active ? t('roles.statuses.active') : t('roles.statuses.inactive')}
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
                        <DropdownMenuItem onClick={() => openPermissions(role)} className="gap-2">
                          <ShieldCheck className="w-4 h-4" /> {t('roles.managePermissions')}
                        </DropdownMenuItem>
                        <DropdownMenuSeparator />
                        <DropdownMenuItem onClick={() => openEdit(role)}>{t('common.edit')}</DropdownMenuItem>
                        <DropdownMenuSeparator />
                        <DropdownMenuItem
                          className="text-destructive focus:text-destructive"
                          onClick={() => {
                            setDeletingRole(role)
                            setDeleteDialogOpen(true)
                          }}
                        >
                          {t('common.delete')}
                        </DropdownMenuItem>
                      </DropdownMenuContent>
                    </DropdownMenu>
                  </div>
                ))}
              </div>

              {/* Pagination bar */}
              {!loading && total > 0 && (
                <div className="flex flex-col sm:flex-row items-center justify-between gap-3 px-4 py-3 border-t text-sm text-muted-foreground">
                  <span className="shrink-0">
                    {t('roles.showing', {
                      from: (page - 1) * perPage + 1,
                      to: Math.min(page * perPage, total),
                      total,
                    })}
                  </span>
                  <div className="flex items-center gap-4">
                    <div className="flex items-center gap-2">
                      <span className="hidden sm:inline shrink-0">{t('roles.perPage')}</span>
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
                      {t('roles.page')} {page} {t('roles.of')} {totalPages}
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
            <DialogTitle>{editingRole ? t('roles.editRole') : t('roles.createRole')}</DialogTitle>
            <DialogDescription>
              {editingRole ? t('roles.editDesc') : t('roles.createDesc')}
            </DialogDescription>
          </DialogHeader>
          <div className="grid gap-4 py-4">
            <div className="space-y-1.5">
              <Label>{t('roles.name')} *</Label>
              <Input
                placeholder="Admin"
                value={form.name}
                onChange={(e) => setForm((f) => ({ ...f, name: e.target.value }))}
              />
            </div>
            <div className="space-y-1.5">
              <Label>{t('roles.description')}</Label>
              <textarea
                placeholder={t('roles.descriptionPlaceholder')}
                rows={3}
                value={form.description}
                onChange={(e: React.ChangeEvent<HTMLTextAreaElement>) => setForm((f) => ({ ...f, description: e.target.value }))}
                className="flex min-h-20 w-full rounded-md border border-input bg-background px-3 py-2 text-sm ring-offset-background placeholder:text-muted-foreground focus-visible:outline-none focus-visible:ring-2 focus-visible:ring-ring focus-visible:ring-offset-2 disabled:cursor-not-allowed disabled:opacity-50 resize-none"
              />
            </div>
            <div className="space-y-1.5">
              <Label>{t('roles.status')}</Label>
              <Select
                value={form.is_active ? 'active' : 'inactive'}
                onValueChange={(v) => setForm((f) => ({ ...f, is_active: v === 'active' }))}
              >
                <SelectTrigger>
                  <SelectValue />
                </SelectTrigger>
                <SelectContent>
                  <SelectItem value="active">{t('roles.statuses.active')}</SelectItem>
                  <SelectItem value="inactive">{t('roles.statuses.inactive')}</SelectItem>
                </SelectContent>
              </Select>
            </div>
          </div>
          <DialogFooter>
            <Button variant="outline" onClick={() => setDialogOpen(false)}>
              {t('common.cancel')}
            </Button>
            <Button onClick={handleSave} disabled={saving || !form.name}>
              {saving ? t('common.loading') : editingRole ? t('roles.saveChanges') : t('roles.createRole')}
            </Button>
          </DialogFooter>
        </DialogContent>
      </Dialog>

      {/* Delete confirm */}
      <Dialog open={deleteDialogOpen} onOpenChange={setDeleteDialogOpen}>
        <DialogContent className="max-w-sm">
          <DialogHeader>
            <DialogTitle>{t('roles.deleteRole')}</DialogTitle>
            <DialogDescription>
              {t('roles.deleteConfirm')}{' '}
              <span className="font-semibold text-foreground">{deletingRole?.name}</span>?{' '}
              {t('roles.deleteWarning')}
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

      {/* Permission Matrix Dialog */}
      <Dialog open={permDialogOpen} onOpenChange={setPermDialogOpen}>
        <DialogContent className="max-w-3xl max-h-[85vh] flex flex-col">
          <DialogHeader>
            <DialogTitle className="flex items-center gap-2">
              <ShieldCheck className="w-5 h-5 text-primary" />
              {t('roles.permissionsFor')}{' '}
              <span
                className={cn(
                  'px-2 py-0.5 rounded-md text-sm font-medium',
                  permRole ? getRoleColor(permRole.name) : ''
                )}
              >
                {permRole?.name}
              </span>
            </DialogTitle>
            <DialogDescription>{t('roles.permissionsDesc')}</DialogDescription>
          </DialogHeader>

          <div className="flex-1 overflow-auto min-h-0">
            {permLoading ? (
              <div className="flex items-center justify-center py-16 text-sm text-muted-foreground">
                {t('common.loading')}
              </div>
            ) : permError ? (
              <div className="rounded-md border border-destructive/30 bg-destructive/10 px-4 py-3 text-sm text-destructive">
                {permError}
              </div>
            ) : allMenus.length === 0 ? (
              <p className="text-sm text-muted-foreground py-8 text-center">{t('roles.noMenus')}</p>
            ) : (
              <table className="w-full text-sm border-collapse">
                <thead className="sticky top-0 bg-background z-10">
                  <tr className="border-b bg-muted/40">
                    <th className="text-left px-4 py-3 font-medium text-muted-foreground min-w-36">เมนู</th>
                    {allPermissions.map((perm) => (
                      <th key={perm.id} className="px-3 py-3 font-medium text-muted-foreground text-center min-w-20">
                        <div className="flex flex-col items-center gap-1">
                          <span className="capitalize">{perm.name}</span>
                          <input
                            type="checkbox"
                            title={`เลือกทั้งหมด: ${perm.name}`}
                            className="h-4 w-4 cursor-pointer accent-primary"
                            checked={allMenus.length > 0 && allMenus.every((m) => (matrix[m.id] ?? new Set()).has(perm.id))}
                            onChange={() => toggleColAll(perm.id)}
                          />
                        </div>
                      </th>
                    ))}
                    <th className="px-3 py-3 font-medium text-muted-foreground text-center min-w-20">
                      <div className="flex flex-col items-center gap-1">
                        <span>{t('roles.selectAll')}</span>
                        <span className="h-4 w-4 block" />
                      </div>
                    </th>
                  </tr>
                </thead>
                <tbody>
                  {allMenus.map((menu, i) => {
                    const menuPerms = matrix[menu.id] ?? new Set()
                    const allChecked = allPermissions.every((p) => menuPerms.has(p.id))
                    const someChecked = allPermissions.some((p) => menuPerms.has(p.id))
                    return (
                      <tr
                        key={menu.id}
                        className={cn(
                          'border-b transition-colors hover:bg-muted/20',
                          i === allMenus.length - 1 && 'border-0'
                        )}
                      >
                        <td className="px-4 py-3 font-medium">{menu.name}</td>
                        {allPermissions.map((perm) => (
                          <td key={perm.id} className="px-3 py-3 text-center">
                            <input
                              type="checkbox"
                              className="h-4 w-4 cursor-pointer accent-primary"
                              checked={menuPerms.has(perm.id)}
                              onChange={() => togglePermission(menu.id, perm.id)}
                            />
                          </td>
                        ))}
                        <td className="px-3 py-3 text-center">
                          <input
                            type="checkbox"
                            title={t('roles.selectAll')}
                            className="h-4 w-4 cursor-pointer accent-primary"
                            checked={allChecked}
                            ref={(el) => {
                              if (el) el.indeterminate = !allChecked && someChecked
                            }}
                            onChange={() => toggleRowAll(menu.id)}
                          />
                        </td>
                      </tr>
                    )
                  })}
                </tbody>
              </table>
            )}
          </div>

          <DialogFooter className="border-t pt-4">
            <Button variant="outline" onClick={() => setPermDialogOpen(false)}>
              {t('common.cancel')}
            </Button>
            <Button onClick={handleSavePermissions} disabled={permSaving || permLoading}>
              {permSaving ? t('common.loading') : t('roles.saveChanges')}
            </Button>
          </DialogFooter>
        </DialogContent>
      </Dialog>
    </div>
  )
}
