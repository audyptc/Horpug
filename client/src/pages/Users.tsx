import { useState, useMemo, useEffect, useCallback } from 'react'
import { useTranslation } from 'react-i18next'
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
import { userService } from '@/services/userService'
import { roleService } from '@/services/roleService'
import type { ApiUser, ApiRole } from '@/types/api'
import { cn } from '@/lib/utils'
import { formatDate } from '@/lib/dateUtils'

type StatusFilter = 'all' | 'active' | 'inactive'

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
  return `https://api.dicebear.com/9.x/avataaars/svg?seed=${encodeURIComponent(name)}`
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
  const [users, setUsers] = useState<ApiUser[]>([])
  const [roles, setRoles] = useState<ApiRole[]>([])
  const [loading, setLoading] = useState(true)
  const [error, setError] = useState('')
  const [search, setSearch] = useState('')
  const [roleFilter, setRoleFilter] = useState('all')
  const [statusFilter, setStatusFilter] = useState<StatusFilter>('all')
  const [dialogOpen, setDialogOpen] = useState(false)
  const [deleteDialogOpen, setDeleteDialogOpen] = useState(false)
  const [editingUser, setEditingUser] = useState<ApiUser | null>(null)
  const [deletingUser, setDeletingUser] = useState<ApiUser | null>(null)
  const [form, setForm] = useState(emptyForm)
  const [saving, setSaving] = useState(false)

  const fetchUsers = useCallback(async () => {
    setLoading(true)
    setError('')
    try {
      const resp = await userService.list(1, 100)
      setUsers(resp.data)
    } catch {
      setError(t('users.loadError'))
    } finally {
      setLoading(false)
    }
  }, [t])

  useEffect(() => {
    fetchUsers()
    roleService.listActive().then(setRoles).catch(() => {})
  }, [fetchUsers])

  const filtered = useMemo(() => {
    return users.filter((u) => {
      const name = u.full_name.toLowerCase()
      const matchSearch =
        name.includes(search.toLowerCase()) ||
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
    setDialogOpen(true)
  }

  async function handleSave() {
    if (!form.full_name || !form.email) return
    setSaving(true)
    try {
      if (editingUser) {
        const payload: Record<string, unknown> = {
          full_name: form.full_name,
          is_active: form.is_active,
        }
        if (form.password) payload.password = form.password

        const updated = await userService.update(editingUser.id, payload)

        if (form.role_id && form.role_id !== editingUser.role?.id) {
          await userService.assignRole(editingUser.id, { role_id: form.role_id })
          updated.role = roles.find((r) => r.id === form.role_id)
        }

        setUsers((prev) => prev.map((u) => (u.id === editingUser.id ? updated : u)))
      } else {
        if (!form.password) return
        const created = await userService.create({
          full_name: form.full_name,
          email: form.email,
          password: form.password,
        })

        if (form.role_id) {
          await userService.assignRole(created.id, { role_id: form.role_id })
          created.role = roles.find((r) => r.id === form.role_id)
        }

        setUsers((prev) => [created, ...prev])
      }
      setDialogOpen(false)
    } catch {
      // error handled silently; server validation will show in console
    } finally {
      setSaving(false)
    }
  }

  async function handleDelete() {
    if (!deletingUser) return
    try {
      await userService.delete(deletingUser.id)
      setUsers((prev) => prev.filter((u) => u.id !== deletingUser.id))
    } catch {
      // ignore
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
            {t('users.subtitle_other', { count: filtered.length, total: users.length })}
          </p>
        </div>
        <div className="flex gap-2 shrink-0">
          <Button variant="outline" size="icon" onClick={fetchUsers} disabled={loading}>
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

      {/* Filters */}
      <Card>
        <CardContent className="pt-4">
          <div className="flex flex-col sm:flex-row gap-3">
            <div className="relative flex-1">
              <Search className="absolute left-3 top-1/2 -translate-y-1/2 w-4 h-4 text-muted-foreground" />
              <Input
                placeholder={t('users.searchPlaceholder')}
                className="pl-9"
                value={search}
                onChange={(e) => setSearch(e.target.value)}
              />
            </div>
            <Select value={roleFilter} onValueChange={setRoleFilter}>
              <SelectTrigger className="w-full sm:w-40">
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
              <SelectTrigger className="w-full sm:w-36">
                <SelectValue placeholder={t('users.allStatus')} />
              </SelectTrigger>
              <SelectContent>
                <SelectItem value="all">{t('users.allStatus')}</SelectItem>
                <SelectItem value="active">{t('users.statuses.active')}</SelectItem>
                <SelectItem value="inactive">{t('users.statuses.inactive')}</SelectItem>
              </SelectContent>
            </Select>
          </div>
        </CardContent>
      </Card>

      {/* Table */}
      <Card>
        <CardHeader>
          <CardTitle>{t('users.userList')}</CardTitle>
          <CardDescription>{t('users.userListDesc')}</CardDescription>
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
                            <Avatar className="h-9 w-9">
                              <AvatarImage src={getAvatar(user.full_name)} />
                              <AvatarFallback className="text-xs">
                                {user.full_name.split(' ').map((n) => n[0]).join('')}
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
                    <Avatar className="h-10 w-10 shrink-0">
                      <AvatarImage src={getAvatar(user.full_name)} />
                      <AvatarFallback className="text-xs">
                        {user.full_name.split(' ').map((n) => n[0]).join('')}
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
