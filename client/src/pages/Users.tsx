import { useState, useMemo } from 'react'
import { useTranslation } from 'react-i18next'
import {
  Search,
  Plus,
  MoreHorizontal,
  Pencil,
  Trash2,
  ShieldCheck,
  UserCog,
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
import { mockUsers } from '@/data/mockUsers'
import type { User, Role, UserStatus } from '@/types'
import { cn } from '@/lib/utils'
import { formatDate } from '@/lib/dateUtils'

const statusVariant: Record<UserStatus, 'success' | 'secondary' | 'destructive'> = {
  active: 'success',
  inactive: 'secondary',
  suspended: 'destructive',
}

const roleColors: Record<Role, string> = {
  admin: 'text-violet-600 bg-violet-500/10',
  manager: 'text-blue-600 bg-blue-500/10',
  editor: 'text-amber-600 bg-amber-500/10',
  viewer: 'text-slate-600 bg-slate-500/10',
}

const emptyUser: Omit<User, 'id' | 'joinedAt' | 'lastLogin'> = {
  name: '',
  email: '',
  role: 'viewer',
  status: 'active',
  department: '',
  phone: '',
  avatar: '',
}

export function Users() {
  const { t } = useTranslation()
  const [users, setUsers] = useState<User[]>(mockUsers)
  const [search, setSearch] = useState('')
  const [roleFilter, setRoleFilter] = useState<Role | 'all'>('all')
  const [statusFilter, setStatusFilter] = useState<UserStatus | 'all'>('all')
  const [dialogOpen, setDialogOpen] = useState(false)
  const [deleteDialogOpen, setDeleteDialogOpen] = useState(false)
  const [editingUser, setEditingUser] = useState<User | null>(null)
  const [deletingUser, setDeletingUser] = useState<User | null>(null)
  const [form, setForm] = useState(emptyUser)

  const filtered = useMemo(() => {
    return users.filter((u) => {
      const matchSearch =
        u.name.toLowerCase().includes(search.toLowerCase()) ||
        u.email.toLowerCase().includes(search.toLowerCase()) ||
        u.department.toLowerCase().includes(search.toLowerCase())
      const matchRole = roleFilter === 'all' || u.role === roleFilter
      const matchStatus = statusFilter === 'all' || u.status === statusFilter
      return matchSearch && matchRole && matchStatus
    })
  }, [users, search, roleFilter, statusFilter])

  function openCreate() {
    setEditingUser(null)
    setForm(emptyUser)
    setDialogOpen(true)
  }

  function openEdit(user: User) {
    setEditingUser(user)
    setForm({
      name: user.name,
      email: user.email,
      role: user.role,
      status: user.status,
      department: user.department,
      phone: user.phone ?? '',
      avatar: user.avatar ?? '',
    })
    setDialogOpen(true)
  }

  function handleSave() {
    if (!form.name || !form.email) return
    if (editingUser) {
      setUsers((prev) =>
        prev.map((u) => (u.id === editingUser.id ? { ...editingUser, ...form } : u))
      )
    } else {
      const newUser: User = {
        ...form,
        id: String(Date.now()),
        joinedAt: new Date().toISOString().slice(0, 10),
        lastLogin: new Date().toISOString(),
        avatar: `https://api.dicebear.com/9.x/avataaars/svg?seed=${encodeURIComponent(form.name)}`,
      }
      setUsers((prev) => [newUser, ...prev])
    }
    setDialogOpen(false)
  }

  function handleDelete() {
    if (!deletingUser) return
    setUsers((prev) => prev.filter((u) => u.id !== deletingUser.id))
    setDeleteDialogOpen(false)
    setDeletingUser(null)
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
        <Button onClick={openCreate} className="gap-2 shrink-0">
          <Plus className="w-4 h-4" />
          <span className="hidden sm:inline">{t('users.addUser')}</span>
        </Button>
      </div>

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
            <Select value={roleFilter} onValueChange={(v) => setRoleFilter(v as Role | 'all')}>
              <SelectTrigger className="w-full sm:w-36">
                <SelectValue placeholder={t('users.allRoles')} />
              </SelectTrigger>
              <SelectContent>
                <SelectItem value="all">{t('users.allRoles')}</SelectItem>
                <SelectItem value="admin">{t('users.roles.admin')}</SelectItem>
                <SelectItem value="manager">{t('users.roles.manager')}</SelectItem>
                <SelectItem value="editor">{t('users.roles.editor')}</SelectItem>
                <SelectItem value="viewer">{t('users.roles.viewer')}</SelectItem>
              </SelectContent>
            </Select>
            <Select
              value={statusFilter}
              onValueChange={(v) => setStatusFilter(v as UserStatus | 'all')}
            >
              <SelectTrigger className="w-full sm:w-36">
                <SelectValue placeholder={t('users.allStatus')} />
              </SelectTrigger>
              <SelectContent>
                <SelectItem value="all">{t('users.allStatus')}</SelectItem>
                <SelectItem value="active">{t('users.statuses.active')}</SelectItem>
                <SelectItem value="inactive">{t('users.statuses.inactive')}</SelectItem>
                <SelectItem value="suspended">{t('users.statuses.suspended')}</SelectItem>
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
          {/* Desktop table */}
          <div className="hidden md:block overflow-x-auto">
            <table className="w-full text-sm">
              <thead>
                <tr className="border-b bg-muted/40">
                  <th className="text-left px-6 py-3 font-medium text-muted-foreground">{t('users.colUser')}</th>
                  <th className="text-left px-4 py-3 font-medium text-muted-foreground">{t('users.colRole')}</th>
                  <th className="text-left px-4 py-3 font-medium text-muted-foreground">{t('users.colDept')}</th>
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
                          <AvatarImage src={user.avatar} />
                          <AvatarFallback className="text-xs">
                            {user.name.split(' ').map((n) => n[0]).join('')}
                          </AvatarFallback>
                        </Avatar>
                        <div>
                          <p className="font-medium">{user.name}</p>
                          <p className="text-xs text-muted-foreground">{user.email}</p>
                        </div>
                      </div>
                    </td>
                    <td className="px-4 py-4">
                      <span
                        className={cn(
                          'inline-flex items-center gap-1 px-2 py-0.5 rounded-md text-xs font-medium',
                          roleColors[user.role]
                        )}
                      >
                        {user.role === 'admin' && <ShieldCheck className="w-3 h-3" />}
                        {user.role === 'manager' && <UserCog className="w-3 h-3" />}
                        {t(`users.roles.${user.role}`)}
                      </span>
                    </td>
                    <td className="px-4 py-4 text-muted-foreground">{user.department}</td>
                    <td className="px-4 py-4">
                      <Badge variant={statusVariant[user.status]}>
                        {t(`users.statuses.${user.status}`)}
                      </Badge>
                    </td>
                    <td className="px-4 py-4 text-muted-foreground">{formatDate(user.joinedAt)}</td>
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
              <div className="text-center py-12 text-muted-foreground">{t('users.noUsers')}</div>
            )}
          </div>

          {/* Mobile cards */}
          <div className="md:hidden divide-y">
            {filtered.map((user) => (
              <div key={user.id} className="p-4 flex items-center gap-3">
                <Avatar className="h-10 w-10 shrink-0">
                  <AvatarImage src={user.avatar} />
                  <AvatarFallback className="text-xs">
                    {user.name.split(' ').map((n) => n[0]).join('')}
                  </AvatarFallback>
                </Avatar>
                <div className="flex-1 min-w-0">
                  <p className="font-medium truncate">{user.name}</p>
                  <p className="text-xs text-muted-foreground truncate">{user.email}</p>
                  <div className="flex items-center gap-2 mt-1">
                    <Badge variant={statusVariant[user.status]} className="text-xs">
                      {t(`users.statuses.${user.status}`)}
                    </Badge>
                    <span className="text-xs text-muted-foreground">
                      {t(`users.roles.${user.role}`)}
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
                  value={form.name}
                  onChange={(e) => setForm((f) => ({ ...f, name: e.target.value }))}
                />
              </div>
              <div className="col-span-2 space-y-1.5">
                <Label>{t('users.email')} *</Label>
                <Input
                  type="email"
                  placeholder="john@example.com"
                  value={form.email}
                  onChange={(e) => setForm((f) => ({ ...f, email: e.target.value }))}
                />
              </div>
              <div className="space-y-1.5">
                <Label>{t('users.role')}</Label>
                <Select
                  value={form.role}
                  onValueChange={(v) => setForm((f) => ({ ...f, role: v as Role }))}
                >
                  <SelectTrigger>
                    <SelectValue />
                  </SelectTrigger>
                  <SelectContent>
                    <SelectItem value="admin">{t('users.roles.admin')}</SelectItem>
                    <SelectItem value="manager">{t('users.roles.manager')}</SelectItem>
                    <SelectItem value="editor">{t('users.roles.editor')}</SelectItem>
                    <SelectItem value="viewer">{t('users.roles.viewer')}</SelectItem>
                  </SelectContent>
                </Select>
              </div>
              <div className="space-y-1.5">
                <Label>{t('users.status')}</Label>
                <Select
                  value={form.status}
                  onValueChange={(v) => setForm((f) => ({ ...f, status: v as UserStatus }))}
                >
                  <SelectTrigger>
                    <SelectValue />
                  </SelectTrigger>
                  <SelectContent>
                    <SelectItem value="active">{t('users.statuses.active')}</SelectItem>
                    <SelectItem value="inactive">{t('users.statuses.inactive')}</SelectItem>
                    <SelectItem value="suspended">{t('users.statuses.suspended')}</SelectItem>
                  </SelectContent>
                </Select>
              </div>
              <div className="space-y-1.5">
                <Label>{t('users.department')}</Label>
                <Input
                  placeholder="Engineering"
                  value={form.department}
                  onChange={(e) => setForm((f) => ({ ...f, department: e.target.value }))}
                />
              </div>
              <div className="space-y-1.5">
                <Label>{t('users.phone')}</Label>
                <Input
                  placeholder="+66 8x xxx xxxx"
                  value={form.phone}
                  onChange={(e) => setForm((f) => ({ ...f, phone: e.target.value }))}
                />
              </div>
            </div>
          </div>
          <DialogFooter>
            <Button variant="outline" onClick={() => setDialogOpen(false)}>
              {t('common.cancel')}
            </Button>
            <Button onClick={handleSave} disabled={!form.name || !form.email}>
              {editingUser ? t('users.saveChanges') : t('users.createUser')}
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
              <span className="font-semibold text-foreground">{deletingUser?.name}</span>?{' '}
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
