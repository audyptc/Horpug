import { useEffect, useMemo, useState, type FormEvent } from 'react'
import { api, extractErrorMessage } from '@/shared/api/client'
import { useLanguage, type TranslationKey } from '@/shared/i18n/language'
import { Card, CardContent, CardDescription, CardHeader, CardTitle } from '@/shared/components/ui/card'
import { Table, TableBody, TableCell, TableHead, TableHeader, TableRow } from '@/shared/components/ui/table'
import { Button } from '@/shared/components/ui/button'
import {
  Sheet,
  SheetContent,
  SheetDescription,
  SheetFooter,
  SheetHeader,
  SheetTitle,
} from '@/shared/components/ui/sheet'
import { menuMeta, type ApiMenu } from '@/features/menu/menus'

type ApiPermission = {
  id: string
  name: string
  description: string
}

type ApiRoleMenuPermission = {
  menu_id: string
  permission_id: string
}

type ApiRole = {
  id: string
  name: string
  description: string
  is_active: boolean
  menu_permissions?: ApiRoleMenuPermission[]
}

const ACTION_ORDER = ['create', 'read', 'update', 'delete']

const actionLabelKeys: Record<string, TranslationKey> = {
  create: 'permissionActionCreate',
  read: 'permissionActionRead',
  update: 'permissionActionUpdate',
  delete: 'permissionActionDelete',
}

function menuLabel(menu: ApiMenu, t: (key: TranslationKey) => string): string {
  const meta = menuMeta[menu.path]
  return meta ? t(meta.labelKey) : menu.name
}

export default function RolePermissionsPage() {
  const { t } = useLanguage()

  const [roles, setRoles] = useState<ApiRole[] | null>(null)
  const [menus, setMenus] = useState<ApiMenu[] | null>(null)
  const [permissions, setPermissions] = useState<ApiPermission[] | null>(null)
  const [loadError, setLoadError] = useState<string | null>(null)

  const [selectedRoleId, setSelectedRoleId] = useState<string | null>(null)
  const [matrix, setMatrix] = useState<Record<string, Set<string>>>({})

  const [saving, setSaving] = useState(false)
  const [saveError, setSaveError] = useState<string | null>(null)
  const [saveSuccess, setSaveSuccess] = useState(false)

  const [formOpen, setFormOpen] = useState(false)
  const [formMode, setFormMode] = useState<'create' | 'edit'>('create')
  const [formName, setFormName] = useState('')
  const [formDescription, setFormDescription] = useState('')
  const [formIsActive, setFormIsActive] = useState(true)
  const [formSaving, setFormSaving] = useState(false)
  const [formError, setFormError] = useState<string | null>(null)

  const [deleting, setDeleting] = useState(false)
  const [deleteError, setDeleteError] = useState<string | null>(null)

  useEffect(() => {
    let cancelled = false
    setLoadError(null)

    Promise.all([
      api.get<ApiRole[]>('/roles'),
      api.get<ApiMenu[]>('/menus'),
      api.get<ApiPermission[]>('/permissions'),
    ])
      .then(([rolesRes, menusRes, permissionsRes]) => {
        if (cancelled) return
        setRoles(rolesRes.data)
        setMenus(menusRes.data)
        setPermissions(permissionsRes.data)
        setSelectedRoleId((current) => current ?? rolesRes.data[0]?.id ?? null)
      })
      .catch((err) => {
        if (!cancelled) setLoadError(extractErrorMessage(err, t('resourceLoadError')))
      })

    return () => {
      cancelled = true
    }
    // eslint-disable-next-line react-hooks/exhaustive-deps
  }, [])

  const selectedRole = useMemo(
    () => roles?.find((role) => role.id === selectedRoleId) ?? null,
    [roles, selectedRoleId]
  )

  useEffect(() => {
    const role = roles?.find((item) => item.id === selectedRoleId) ?? null
    const next: Record<string, Set<string>> = {}
    for (const item of role?.menu_permissions ?? []) {
      if (!next[item.menu_id]) next[item.menu_id] = new Set()
      next[item.menu_id].add(item.permission_id)
    }
    setMatrix(next)
    setSaveError(null)
    setSaveSuccess(false)
    // Deliberately keyed on selectedRoleId (not the role object or `roles`):
    // handleSave replaces the role object in `roles` after a successful save,
    // and depending on that reference would immediately reset saveSuccess/matrix.
    // eslint-disable-next-line react-hooks/exhaustive-deps
  }, [selectedRoleId])

  const sortedMenus = useMemo(
    () => (menus ?? []).filter((menu) => menu.is_active).sort((a, b) => a.path.localeCompare(b.path)),
    [menus]
  )

  const sortedPermissions = useMemo(() => {
    return [...(permissions ?? [])].sort((a, b) => {
      const rankA = ACTION_ORDER.indexOf(a.name)
      const rankB = ACTION_ORDER.indexOf(b.name)
      if (rankA === -1 && rankB === -1) return a.name.localeCompare(b.name)
      if (rankA === -1) return 1
      if (rankB === -1) return -1
      return rankA - rankB
    })
  }, [permissions])

  function toggleCell(menuId: string, permissionId: string) {
    setMatrix((prev) => {
      const next = { ...prev }
      const set = new Set(next[menuId] ?? [])
      if (set.has(permissionId)) {
        set.delete(permissionId)
      } else {
        set.add(permissionId)
      }
      next[menuId] = set
      return next
    })
    setSaveSuccess(false)
  }

  async function handleSave() {
    if (!selectedRoleId) return

    setSaving(true)
    setSaveError(null)
    setSaveSuccess(false)

    const menu_permissions = Object.entries(matrix)
      .filter(([, permissionIds]) => permissionIds.size > 0)
      .map(([menu_id, permissionIds]) => ({ menu_id, permission_ids: Array.from(permissionIds) }))

    try {
      const { data } = await api.put<ApiRole>(`/roles/${selectedRoleId}`, { menu_permissions })
      setRoles((prev) => prev?.map((role) => (role.id === data.id ? data : role)) ?? prev)
      setSaveSuccess(true)
    } catch (err) {
      setSaveError(extractErrorMessage(err, t('rolePermissionsSaveError')))
    } finally {
      setSaving(false)
    }
  }

  function openCreateForm() {
    setFormMode('create')
    setFormName('')
    setFormDescription('')
    setFormIsActive(true)
    setFormError(null)
    setFormOpen(true)
  }

  function openEditForm() {
    if (!selectedRole) return
    setFormMode('edit')
    setFormName(selectedRole.name)
    setFormDescription(selectedRole.description)
    setFormIsActive(selectedRole.is_active)
    setFormError(null)
    setFormOpen(true)
  }

  async function handleFormSubmit(event: FormEvent<HTMLFormElement>) {
    event.preventDefault()

    const name = formName.trim()
    if (!name) {
      setFormError(t('rolePermissionsNameRequired'))
      return
    }

    setFormSaving(true)
    setFormError(null)

    try {
      if (formMode === 'create') {
        const { data } = await api.post<ApiRole>('/roles', {
          name,
          description: formDescription,
          is_active: formIsActive,
        })
        setRoles((prev) => [...(prev ?? []), data])
        setSelectedRoleId(data.id)
      } else if (selectedRoleId) {
        const { data } = await api.put<ApiRole>(`/roles/${selectedRoleId}`, {
          name,
          description: formDescription,
          is_active: formIsActive,
        })
        setRoles((prev) => prev?.map((role) => (role.id === data.id ? data : role)) ?? prev)
      }
      setFormOpen(false)
    } catch (err) {
      const fallback = formMode === 'create' ? t('rolePermissionsCreateError') : t('rolePermissionsUpdateError')
      setFormError(extractErrorMessage(err, fallback))
    } finally {
      setFormSaving(false)
    }
  }

  async function handleDeleteRole() {
    if (!selectedRoleId) return
    if (!window.confirm(t('rolePermissionsDeleteConfirm'))) return

    setDeleting(true)
    setDeleteError(null)

    try {
      await api.delete(`/roles/${selectedRoleId}`)
      setRoles((prev) => {
        const remaining = prev?.filter((role) => role.id !== selectedRoleId) ?? null
        setSelectedRoleId(remaining?.[0]?.id ?? null)
        return remaining
      })
    } catch (err) {
      setDeleteError(extractErrorMessage(err, t('rolePermissionsDeleteError')))
    } finally {
      setDeleting(false)
    }
  }

  const isLoading = !loadError && (roles === null || menus === null || permissions === null)

  return (
    <main className="content">
      <section className="welcome">
        <h1>{t('menuRoles')}</h1>
        <p>{t('menuRolesDescription')}</p>
      </section>

      <Card>
        <CardHeader className="flex flex-row items-start justify-between gap-4">
          <div>
            <CardTitle>{t('menuRoles')}</CardTitle>
            <CardDescription>
              GET /roles · GET /menus · GET /permissions · POST /roles · PUT /roles/:id · DELETE /roles/:id
            </CardDescription>
          </div>
          <Button onClick={openCreateForm} disabled={isLoading}>
            {t('rolePermissionsCreateRole')}
          </Button>
        </CardHeader>
        <CardContent className="flex flex-col gap-4">
          {loadError && <p className="resource-error">{loadError}</p>}

          {!loadError && isLoading && <p className="metric-detail">{t('loading')}</p>}

          {!loadError && !isLoading && roles && roles.length === 0 && (
            <p className="metric-detail">{t('rolePermissionsNoRoles')}</p>
          )}

          {!loadError && !isLoading && roles && roles.length > 0 && (
            <>
              <div className="flex flex-wrap items-end gap-3">
                <label className="flex flex-col gap-1.5 text-sm font-medium sm:max-w-xs">
                  {t('rolePermissionsSelectRoleLabel')}
                  <select
                    className="h-10 rounded-md border border-input bg-transparent px-3 text-sm"
                    value={selectedRoleId ?? ''}
                    onChange={(event) => setSelectedRoleId(event.target.value)}
                  >
                    {roles.map((role) => (
                      <option key={role.id} value={role.id}>
                        {role.name}
                      </option>
                    ))}
                  </select>
                </label>
                <Button variant="outline" onClick={openEditForm} disabled={!selectedRole}>
                  {t('rolePermissionsEditRole')}
                </Button>
                <Button
                  variant="destructive"
                  onClick={handleDeleteRole}
                  disabled={!selectedRole || deleting}
                >
                  {deleting ? t('rolePermissionsDeleting') : t('rolePermissionsDeleteRole')}
                </Button>
              </div>
              {deleteError && <p className="resource-error">{deleteError}</p>}

              <div className="table-wrap">
                <Table>
                  <TableHeader>
                    <TableRow>
                      <TableHead>{t('rolePermissionsMenuColumn')}</TableHead>
                      {sortedPermissions.map((permission) => (
                        <TableHead key={permission.id} className="text-center">
                          {actionLabelKeys[permission.name] ? t(actionLabelKeys[permission.name]) : permission.name}
                        </TableHead>
                      ))}
                    </TableRow>
                  </TableHeader>
                  <TableBody>
                    {sortedMenus.map((menu) => (
                      <TableRow key={menu.id}>
                        <TableCell className="font-semibold">{menuLabel(menu, t)}</TableCell>
                        {sortedPermissions.map((permission) => (
                          <TableCell key={permission.id} className="text-center">
                            <input
                              type="checkbox"
                              aria-label={`${menuLabel(menu, t)} - ${permission.name}`}
                              checked={matrix[menu.id]?.has(permission.id) ?? false}
                              onChange={() => toggleCell(menu.id, permission.id)}
                              className="h-4 w-4 accent-primary"
                            />
                          </TableCell>
                        ))}
                      </TableRow>
                    ))}
                  </TableBody>
                </Table>
              </div>

              <div className="flex items-center gap-3">
                <Button onClick={handleSave} disabled={saving}>
                  {saving ? t('rolePermissionsSaving') : t('rolePermissionsSave')}
                </Button>
                {saveSuccess && <p className="metric-detail">{t('rolePermissionsSaved')}</p>}
                {saveError && <p className="resource-error">{saveError}</p>}
              </div>
            </>
          )}
        </CardContent>
      </Card>

      <Sheet open={formOpen} onOpenChange={setFormOpen}>
        <SheetContent>
          <form className="flex h-full flex-col gap-4" onSubmit={handleFormSubmit}>
            <SheetHeader>
              <SheetTitle>
                {formMode === 'create' ? t('rolePermissionsFormCreateTitle') : t('rolePermissionsFormEditTitle')}
              </SheetTitle>
              <SheetDescription>
                {formMode === 'create'
                  ? t('rolePermissionsFormCreateDescription')
                  : t('rolePermissionsFormEditDescription')}
              </SheetDescription>
            </SheetHeader>

            <label className="flex flex-col gap-1.5 text-sm font-medium">
              {t('rolePermissionsFormNameLabel')}
              <input
                type="text"
                className="h-10 rounded-md border border-input bg-transparent px-3 text-sm"
                value={formName}
                onChange={(event) => setFormName(event.target.value)}
                autoFocus
              />
            </label>

            <label className="flex flex-col gap-1.5 text-sm font-medium">
              {t('rolePermissionsFormDescriptionLabel')}
              <textarea
                className="min-h-20 rounded-md border border-input bg-transparent px-3 py-2 text-sm"
                value={formDescription}
                onChange={(event) => setFormDescription(event.target.value)}
              />
            </label>

            <label className="flex items-center gap-2 text-sm font-medium">
              <input
                type="checkbox"
                className="h-4 w-4 accent-primary"
                checked={formIsActive}
                onChange={(event) => setFormIsActive(event.target.checked)}
              />
              {t('rolePermissionsFormActiveLabel')}
            </label>

            {formError && <p className="resource-error">{formError}</p>}

            <SheetFooter>
              <Button type="button" variant="outline" onClick={() => setFormOpen(false)}>
                {t('rolePermissionsFormCancel')}
              </Button>
              <Button type="submit" disabled={formSaving}>
                {formSaving ? t('rolePermissionsSaving') : t('rolePermissionsFormSave')}
              </Button>
            </SheetFooter>
          </form>
        </SheetContent>
      </Sheet>
    </main>
  )
}
