import { useState, useMemo, useEffect, useCallback } from 'react'
import { useTranslation } from 'react-i18next'
import * as XLSX from 'xlsx'
import {
  Search,
  Plus,
  MoreHorizontal,
  Pencil,
  Trash2,
  ShieldCheck,
  UserCog,
  SearchX,
  RefreshCw,
  ChevronLeft,
  ChevronRight,
  ChevronsLeft,
  ChevronsRight,
  Download,
} from 'lucide-react'
import { Button } from '@/components/ui/button'
import { Input } from '@/components/ui/input'
import { Badge } from '@/components/ui/badge'
import { Avatar, AvatarFallback, AvatarImage } from '@/components/ui/avatar'
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
import { userService } from '@/features/users/userService'
import { roleService } from '@/features/roles/roleService'
import { useToast } from '@/components/ui/toast'
import type { ApiUser, ApiRole } from '@/types'
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

function getAvatar(name: string) {
  return `https://api.dicebear.com/9.x/micah/svg?seed=${encodeURIComponent(name)}&backgroundColor=b6e3f4,c0aede,d1d4f9,ffd5dc,ffdfbf`
}

const avatarGradients = [
  'from-violet-500 to-purple-600',
  'from-blue-500 to-cyan-500',
  'from-emerald-500 to-teal-600',
  'from-orange-500 to-amber-500',
  'from-rose-500 to-pink-600',
  'from-indigo-500 to-blue-600',
] as const

function getAvatarGradient(name: string) {
  const hash = name.split('').reduce((acc, c) => acc + c.charCodeAt(0), 0)
  return avatarGradients[hash % avatarGradients.length]
}

function getInitials(name: string) {
  return name.split(' ').map((n) => n[0]).join('').slice(0, 2).toUpperCase()
}

const emptyForm = {
  full_name: '',
  email: '',
  password: '',
  role_id: '',
  is_active: true,
}

export function Users() {
  const { t } = useTranslation()
  const toast = useToast()
  const [users, setUsers] = useState<ApiUser[]>([])
  const [roles, setRoles] = useState<ApiRole[]>([])
  const [loading, setLoading] = useState(true)
  const [error, setError] = useState('')
  const [search, setSearch] = useState('')
  const [roleFilter, setRoleFilter] = useState('all')
  const [statusFilter, setStatusFilter] = useState<StatusFilter>('all')
  const [page, setPage] = useState(1)
  const [perPage, setPerPage] = useState<number>(20)
  const [total, setTotal] = useState(0)
  const [totalPages, setTotalPages] = useState(1)
  const [dialogOpen, setDialogOpen] = useState(false)
  const [deleteDialogOpen, setDeleteDialogOpen] = useState(false)
  const [editingUser, setEditingUser] = useState<ApiUser | null>(null)
  const [deletingUser, setDeletingUser] = useState<ApiUser | null>(null)
  const [form, setForm] = useState(emptyForm)
  const [saving, setSaving] = useState(false)
  const [saveError, setSaveError] = useState('')
  const [exporting, setExporting] = useState(false)

  const fetchUsers = useCallback(async (p: number, pp: number) => {
    setLoading(true)
    setError('')
    try {
      const resp = await userService.list(p, pp)
      setUsers(resp.data)
      setTotal(resp.meta.total)
      setTotalPages(resp.meta.total_pages)
    } catch {
      setError(t('users.loadError'))
    } finally {
      setLoading(false)
    }
  }, [t])

  useEffect(() => {
    fetchUsers(page, perPage)
  }, [fetchUsers, page, perPage])

  useEffect(() => {
    roleService.listActive().then(setRoles).catch(() => {})
  }, [])

  function handlePerPageChange(value: string) {
    setPerPage(Number(value))
    setPage(1)
  }

  async function handleExport() {
    setExporting(true)
    try {
      const resp = await userService.list(1, 9999)
      const rows = resp.data.map((u) => ({
        [t('users.exportColName')]: u.full_name,
        [t('users.exportColEmail')]: u.email,
        [t('users.exportColRole')]: u.role?.name ?? '',
        [t('users.exportColStatus')]: u.is_active ? t('users.statuses.active') : t('users.statuses.inactive'),
        [t('users.exportColCreated')]: u.created_at ? new Date(u.created_at).toLocaleDateString() : '',
      }))
      const ws = XLSX.utils.json_to_sheet(rows)
      const wb = XLSX.utils.book_new()
      XLSX.utils.book_append_sheet(wb, ws, t('users.title'))
      XLSX.writeFile(wb, `users_${new Date().toISOString().slice(0, 10)}.xlsx`)
    } finally {
      setExporting(false)
    }
  }

  const filtered = useMemo(() => {
    return users.filter((u) => {
      const matchSearch =
        u.full_name.toLowerCase().includes(search.toLowerCase()) ||
        u.email.toLowerCase().includes(search.toLowerCase())
      const matchRole =
        roleFilter === 'all' || u.role?.name.toLowerCase() === roleFilter.toLowerCase()
      const matchStatus =
        statusFilter === 'all' ||
        (statusFilter === 'active' ? u.is_active : !u.is_active)
      return matchSearch && matchRole && matchStatus
    })
  }, [users, search, roleFilter, statusFilter])

  function openCreate() {
    setEditingUser(null)
    setForm(emptyForm)
    setSaveError('')
    setDialogOpen(true)
  }

  function openEdit(user: ApiUser) {
    setEditingUser(user)
    setForm({
      full_name: user.full_name,
      email: user.email,
      password: '',
      role_id: user.role?.id ?? '',
      is_active: user.is_active,
    })
    setSaveError('')
    setDialogOpen(true)
  }

  async function handleSave() {
    if (!form.full_name || !form.email) return
    setSaving(true)
    setSaveError('')
    try {
      if (editingUser) {
        const payload: Record<string, unknown> = {
          full_name: form.full_name,
          is_active: form.is_active,
        }
        if (form.password) payload.password = form.password

        await userService.update(editingUser.id, payload)

        if (form.role_id && form.role_id !== editingUser.role?.id) {
          await userService.assignRole(editingUser.id, { role_id: form.role_id })
        }
        toast.success(t('users.editSuccess'))
      } else {
        if (!form.password) return
        const created = await userService.create({
          full_name: form.full_name,
          email: form.email,
          password: form.password,
        })

        if (form.role_id) {
          await userService.assignRole(created.id, { role_id: form.role_id })
        }
        toast.success(t('users.createSuccess'))
      }
      setDialogOpen(false)
      await fetchUsers(page, perPage)
    } catch (err: unknown) {
      const msg =
        (err as { response?: { data?: { message?: string } } })?.response?.data?.message ??
        t('users.saveError')
      setSaveError(msg)
    } finally {
      setSaving(false)
    }
  }

  async function handleDelete() {
    if (!deletingUser) return
    try {
      await userService.delete(deletingUser.id)
      toast.success(t('users.deleteSuccess'))
      const newPage = users.length === 1 && page > 1 ? page - 1 : page
      setPage(newPage)
      await fetchUsers(newPage, perPage)
    } catch (err: unknown) {
      const reason = (err as { response?: { data?: { message?: string } } })?.response?.data?.message
      toast.error(reason ? `${t('users.deleteError')}: ${reason}` : t('users.deleteError'))
    } finally {
      setDeleteDialogOpen(false)
      setDeletingUser(null)
    }
  }

  return (
    <div className="space-y-6">
      <div className="flex flex-wrap items-start justify-between gap-3">
        <div className="min-w-0">
          <h1 className="text-2xl font-bold tracking-tight">{t('users.title')}</h1>
          <p className="text-muted-foreground text-sm mt-1">
            {t('users.subtitle_other', { count: filtered.length, total })}
          </p>
        </div>
        <div className="flex gap-2 shrink-0">
          <Button variant="outline" size="icon" onClick={() => fetchUsers(page, perPage)} disabled={loading}>
            <RefreshCw className={cn('w-4 h-4', loading && 'animate-spin')} />
          </Button>
          <Button onClick={openCreate} className="gap-2">
            <Plus className="w-4 h-4" />
            <span className="hidden sm:inline">{t('users.addUser')}</span>
          </Button>
        </div>
      </div>

      {error && (
        <div className="rounded-md border border-destructive/30 bg-destructive/10 px-4 py-3 text-sm text-destructive">
          {error}
        </div>
      )}

      {/* Table */}
      <Card>
        <CardHeader>
          <div className="flex flex-col sm:flex-row sm:items-center justify-between gap-3">
            <div>
              <CardTitle>{t('users.userList')}</CardTitle>
              <CardDescription>{t('users.userListDesc')}</CardDescription>
            </div>
            <div className="flex flex-col sm:flex-row gap-2 sm:items-center">
              <div className="relative">
                <Search className="absolute left-3 top-1/2 -translate-y-1/2 w-4 h-4 text-muted-foreground" />
                <Input
                  placeholder={t('users.searchPlaceholder')}
                  className="pl-9 w-full sm:w-56"
                  value={search}
                  onChange={(e) => setSearch(e.target.value)}
                />
              </div>
              <Select value={roleFilter} onValueChange={setRoleFilter}>
                <SelectTrigger className="w-full sm:w-36">
                  <SelectValue placeholder={t('users.allRoles')} />
                </SelectTrigger>
                <SelectContent>
                  <SelectItem value="all">{t('users.allRoles')}</SelectItem>
                  {roles.map((r) => (
                    <SelectItem key={r.id} value={r.name.toLowerCase()}>
                      {r.name}
                    </SelectItem>
                  ))}
                </SelectContent>
              </Select>
              <Select
                value={statusFilter}
                onValueChange={(v) => setStatusFilter(v as StatusFilter)}
              >
                <SelectTrigger className="w-full sm:w-32">
                  <SelectValue placeholder={t('users.allStatus')} />
                </SelectTrigger>
                <SelectContent>
                  <SelectItem value="all">{t('users.allStatus')}</SelectItem>
                  <SelectItem value="active">{t('users.statuses.active')}</SelectItem>
                  <SelectItem value="inactive">{t('users.statuses.inactive')}</SelectItem>
                </SelectContent>
              </Select>
              <Button variant="outline" onClick={handleExport} disabled={exporting} className="gap-2 shrink-0">
                <Download className="w-4 h-4" />
                <span className="hidden sm:inline">
                  {exporting ? t('users.exporting') : t('users.exportExcel')}
                </span>
              </Button>
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
                      <th className="text-left px-6 py-3 font-medium text-muted-foreground">{t('users.colUser')}</th>
                      <th className="text-left px-4 py-3 font-medium text-muted-foreground">{t('users.colRole')}</th>
                      <th className="text-left px-4 py-3 font-medium text-muted-foreground">{t('users.colStatus')}</th>
                      <th className="text-left px-4 py-3 font-medium text-muted-foreground">{t('users.colJoined')}</th>
                      <th className="text-right px-6 py-3 font-medium text-muted-foreground">{t('users.colActions')}</th>
                    </tr>
                  </thead>
                  <tbody>
                    {filtered.map((user, i) => (
                      <tr
                        key={user.id}
                        className={cn(
                          'border-b transition-colors hover:bg-muted/30',
                          i === filtered.length - 1 && 'border-0'
                        )}
                      >
                        <td className="px-6 py-4">
                          <div className="flex items-center gap-3">
                            <Avatar className="h-10 w-10 ring-2 ring-offset-2 ring-muted">
                              <AvatarImage src={getAvatar(user.full_name)} className="object-cover" />
                              <AvatarFallback className={cn('text-xs font-semibold text-white bg-linear-to-br', getAvatarGradient(user.full_name))}>
                                {getInitials(user.full_name)}
                              </AvatarFallback>
                            </Avatar>
                            <div>
                              <p className="font-medium">{user.full_name}</p>
                              <p className="text-xs text-muted-foreground">{user.email}</p>
                            </div>
                          </div>
                        </td>
                        <td className="px-4 py-4">
                          {user.role ? (
                            <span
                              className={cn(
                                'inline-flex items-center gap-1 px-2 py-0.5 rounded-md text-xs font-medium',
                                getRoleColor(user.role.name)
                              )}
                            >
                              {user.role.name.toLowerCase() === 'admin' && <ShieldCheck className="w-3 h-3" />}
                              {user.role.name.toLowerCase() === 'manager' && <UserCog className="w-3 h-3" />}
                              {user.role.name}
                            </span>
                          ) : (
                            <span className="text-xs text-muted-foreground">—</span>
                          )}
                        </td>
                        <td className="px-4 py-4">
                          <Badge variant={user.is_active ? statusVariant.active : statusVariant.inactive}>
                            {user.is_active ? t('users.statuses.active') : t('users.statuses.inactive')}
                          </Badge>
                        </td>
                        <td className="px-4 py-4 text-muted-foreground">{formatDate(user.created_at)}</td>
                        <td className="px-6 py-4 text-right">
                          <DropdownMenu>
                            <DropdownMenuTrigger asChild>
                              <Button variant="ghost" size="icon" className="h-8 w-8">
                                <MoreHorizontal className="w-4 h-4" />
                              </Button>
                            </DropdownMenuTrigger>
                            <DropdownMenuContent align="end">
                              <DropdownMenuItem onClick={() => openEdit(user)} className="gap-2">
                                <Pencil className="w-4 h-4" /> {t('common.edit')}
                              </DropdownMenuItem>
                              <DropdownMenuSeparator />
                              <DropdownMenuItem
                                className="gap-2 text-destructive focus:text-destructive"
                                onClick={() => {
                                  setDeletingUser(user)
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
                    <p className="text-sm font-medium">{t('users.noUsers')}</p>
                    {(search || roleFilter !== 'all' || statusFilter !== 'all') && (
                      <button
                        type="button"
                        onClick={() => { setSearch(''); setRoleFilter('all'); setStatusFilter('all') }}
                        className="text-xs text-primary hover:underline"
                      >
                        {t('users.clearFilters')}
                      </button>
                    )}
                  </div>
                )}
              </div>

              {/* Mobile cards */}
              <div className="md:hidden divide-y">
                {filtered.map((user) => (
                  <div key={user.id} className="p-4 flex items-center gap-3">
                    <Avatar className="h-11 w-11 shrink-0 ring-2 ring-offset-2 ring-muted">
                      <AvatarImage src={getAvatar(user.full_name)} className="object-cover" />
                      <AvatarFallback className={cn('text-xs font-semibold text-white bg-linear-to-br', getAvatarGradient(user.full_name))}>
                        {getInitials(user.full_name)}
                      </AvatarFallback>
                    </Avatar>
                    <div className="flex-1 min-w-0">
                      <p className="font-medium truncate">{user.full_name}</p>
                      <p className="text-xs text-muted-foreground truncate">{user.email}</p>
                      <div className="flex items-center gap-2 mt-1">
                        <Badge variant={user.is_active ? statusVariant.active : statusVariant.inactive} className="text-xs">
                          {user.is_active ? t('users.statuses.active') : t('users.statuses.inactive')}
                        </Badge>
                        {user.role && (
                          <span className="text-xs text-muted-foreground">{user.role.name}</span>
                        )}
                      </div>
                    </div>
                    <DropdownMenu>
                      <DropdownMenuTrigger asChild>
                        <Button variant="ghost" size="icon" className="h-8 w-8 shrink-0">
                          <MoreHorizontal className="w-4 h-4" />
                        </Button>
                      </DropdownMenuTrigger>
                      <DropdownMenuContent align="end">
                        <DropdownMenuItem onClick={() => openEdit(user)}>{t('common.edit')}</DropdownMenuItem>
                        <DropdownMenuSeparator />
                        <DropdownMenuItem
                          className="text-destructive focus:text-destructive"
                          onClick={() => {
                            setDeletingUser(user)
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
                    {t('users.showing', {
                      from: (page - 1) * perPage + 1,
                      to: Math.min(page * perPage, total),
                      total,
                    })}
                  </span>

                  <div className="flex items-center gap-4">
                    {/* Per-page selector */}
                    <div className="flex items-center gap-2">
                      <span className="hidden sm:inline shrink-0">{t('users.perPage')}</span>
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

                    {/* Page info */}
                    <span className="shrink-0">
                      {t('users.page')} {page} {t('users.of')} {totalPages}
                    </span>

                    {/* Nav buttons */}
                    <div className="flex items-center gap-1">
                      <Button
                        variant="outline"
                        size="icon"
                        className="h-8 w-8"
                        onClick={() => setPage(1)}
                        disabled={page === 1}
                      >
                        <ChevronsLeft className="w-4 h-4" />
                      </Button>
                      <Button
                        variant="outline"
                        size="icon"
                        className="h-8 w-8"
                        onClick={() => setPage((p) => p - 1)}
                        disabled={page === 1}
                      >
                        <ChevronLeft className="w-4 h-4" />
                      </Button>
                      <Button
                        variant="outline"
                        size="icon"
                        className="h-8 w-8"
                        onClick={() => setPage((p) => p + 1)}
                        disabled={page >= totalPages}
                      >
                        <ChevronRight className="w-4 h-4" />
                      </Button>
                      <Button
                        variant="outline"
                        size="icon"
                        className="h-8 w-8"
                        onClick={() => setPage(totalPages)}
                        disabled={page >= totalPages}
                      >
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
            <DialogTitle>{editingUser ? t('users.editUser') : t('users.createUser')}</DialogTitle>
            <DialogDescription>
              {editingUser ? t('users.editDesc') : t('users.createDesc')}
            </DialogDescription>
          </DialogHeader>
          <div className="grid gap-4 py-4">
            <div className="grid grid-cols-2 gap-4">
              <div className="col-span-2 space-y-1.5">
                <Label>{t('users.fullName')} *</Label>
                <Input
                  placeholder="John Doe"
                  value={form.full_name}
                  onChange={(e) => setForm((f) => ({ ...f, full_name: e.target.value }))}
                />
              </div>
              {!editingUser && (
                <div className="col-span-2 space-y-1.5">
                  <Label>{t('users.email')} *</Label>
                  <Input
                    type="email"
                    placeholder="john@example.com"
                    value={form.email}
                    onChange={(e) => setForm((f) => ({ ...f, email: e.target.value }))}
                  />
                </div>
              )}
              <div className="col-span-2 space-y-1.5">
                <Label>
                  {t('users.password')}
                  {!editingUser && ' *'}
                  {editingUser && (
                    <span className="ml-1 text-xs text-muted-foreground">
                      ({t('users.passwordOptional')})
                    </span>
                  )}
                </Label>
                <Input
                  type="password"
                  placeholder="••••••••"
                  value={form.password}
                  onChange={(e) => setForm((f) => ({ ...f, password: e.target.value }))}
                />
              </div>
              <div className="space-y-1.5">
                <Label>{t('users.role')}</Label>
                <Select
                  value={form.role_id}
                  onValueChange={(v) => setForm((f) => ({ ...f, role_id: v }))}
                >
                  <SelectTrigger>
                    <SelectValue placeholder={t('users.selectRole')} />
                  </SelectTrigger>
                  <SelectContent>
                    {roles.map((r) => (
                      <SelectItem key={r.id} value={r.id}>
                        {r.name}
                      </SelectItem>
                    ))}
                  </SelectContent>
                </Select>
              </div>
              <div className="space-y-1.5">
                <Label>{t('users.status')}</Label>
                <Select
                  value={form.is_active ? 'active' : 'inactive'}
                  onValueChange={(v) => setForm((f) => ({ ...f, is_active: v === 'active' }))}
                >
                  <SelectTrigger>
                    <SelectValue />
                  </SelectTrigger>
                  <SelectContent>
                    <SelectItem value="active">{t('users.statuses.active')}</SelectItem>
                    <SelectItem value="inactive">{t('users.statuses.inactive')}</SelectItem>
                  </SelectContent>
                </Select>
              </div>
            </div>
          </div>
          {saveError && (
            <p className="text-sm text-destructive px-1 pb-2">{saveError}</p>
          )}
          <DialogFooter>
            <Button variant="outline" onClick={() => setDialogOpen(false)}>
              {t('common.cancel')}
            </Button>
            <Button
              onClick={handleSave}
              disabled={saving || !form.full_name || (!editingUser && (!form.email || !form.password))}
            >
              {saving ? t('common.loading') : editingUser ? t('users.saveChanges') : t('users.createUser')}
            </Button>
          </DialogFooter>
        </DialogContent>
      </Dialog>

      {/* Delete confirm */}
      <Dialog open={deleteDialogOpen} onOpenChange={setDeleteDialogOpen}>
        <DialogContent className="max-w-sm">
          <DialogHeader>
            <DialogTitle>{t('users.deleteUser')}</DialogTitle>
            <DialogDescription>
              {t('users.deleteConfirm')}{' '}
              <span className="font-semibold text-foreground">{deletingUser?.full_name}</span>?{' '}
              {t('users.deleteWarning')}
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
