import { useEffect, useMemo, useState, type FormEvent } from 'react'
import { ChevronLeft, ChevronRight, KeyRound, Pencil, Trash2 } from 'lucide-react'
import { api, extractErrorMessage, type ApiPage } from '@/shared/api/client'
import { useLanguage, type TranslationKey } from '@/shared/i18n/language'
import { Badge } from '@/shared/components/ui/badge'
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

function buildRoleMatrix(role: ApiRole | null): Record<string, Set<string>> {
  const next: Record<string, Set<string>> = {}
  for (const item of role?.menu_permissions ?? []) {
    if (!next[item.menu_id]) next[item.menu_id] = new Set()
    next[item.menu_id].add(item.permission_id)
  }
  return next
}

function areMatricesEqual(
  left: Record<string, Set<string>>,
  right: Record<string, Set<string>>
): boolean {
  const menuIds = new Set([...Object.keys(left), ...Object.keys(right)])
  for (const menuId of menuIds) {
    const leftIds = left[menuId] ?? new Set<string>()
    const rightIds = right[menuId] ?? new Set<string>()
    if (leftIds.size !== rightIds.size) return false
    for (const permissionId of leftIds) {
      if (!rightIds.has(permissionId)) return false
    }
  }
  return true
}

type View = 'list' | 'permissions'

const ROLES_PAGE_SIZE = 10

export default function RolePermissionsPage() {
  const { t } = useLanguage()

  const [roles, setRoles] = useState<ApiRole[] | null>(null)
  const [menus, setMenus] = useState<ApiMenu[] | null>(null)
  const [permissions, setPermissions] = useState<ApiPermission[] | null>(null)
  const [loadError, setLoadError] = useState<string | null>(null)

  const [view, setView] = useState<View>('list')
  const [roleQuery, setRoleQuery] = useState('')
  const [rolePage, setRolePage] = useState(1)
  const [selectedRoleId, setSelectedRoleId] = useState<string | null>(null)
  const [matrix, setMatrix] = useState<Record<string, Set<string>>>({})
  const [menuQuery, setMenuQuery] = useState('')

  const [saving, setSaving] = useState(false)
  const [saveError, setSaveError] = useState<string | null>(null)
  const [saveSuccess, setSaveSuccess] = useState(false)

  const [formOpen, setFormOpen] = useState(false)
  const [formRoleId, setFormRoleId] = useState<string | null>(null)
  const [formName, setFormName] = useState('')
  const [formDescription, setFormDescription] = useState('')
  const [formIsActive, setFormIsActive] = useState(true)
  const [formSaving, setFormSaving] = useState(false)
  const [formError, setFormError] = useState<string | null>(null)

  const [deletingRoleId, setDeletingRoleId] = useState<string | null>(null)
  const [deleteError, setDeleteError] = useState<string | null>(null)

  useEffect(() => {
    let cancelled = false

    Promise.all([
      api.get<ApiPage<ApiRole[]>>('/roles', { params: { per_page: 100 } }),
      api.get<ApiPage<ApiMenu[]>>('/menus', { params: { per_page: 100 } }),
      api.get<ApiPage<ApiPermission[]>>('/permissions', { params: { per_page: 100 } }),
    ])
      .then(([rolesRes, menusRes, permissionsRes]) => {
        if (cancelled) return
        setRoles(rolesRes.data.data)
        setMenus(menusRes.data.data)
        setPermissions(permissionsRes.data.data)
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

  const baselineMatrix = useMemo(() => buildRoleMatrix(selectedRole), [selectedRole])

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

  const filteredMenus = useMemo(() => {
    const query = menuQuery.trim().toLocaleLowerCase()
    if (!query) return sortedMenus

    return sortedMenus.filter((menu) => {
      const label = menuLabel(menu, t).toLocaleLowerCase()
      return (
        label.includes(query) ||
        menu.name.toLocaleLowerCase().includes(query) ||
        menu.path.toLocaleLowerCase().includes(query)
      )
    })
  }, [menuQuery, sortedMenus, t])

  const hasUnsavedChanges = useMemo(
    () => !areMatricesEqual(matrix, baselineMatrix),
    [baselineMatrix, matrix]
  )

  const filteredRoles = useMemo(() => {
    const query = roleQuery.trim().toLocaleLowerCase()
    if (!query) return roles ?? []

    return (roles ?? []).filter((role) => {
      return (
        role.name.toLocaleLowerCase().includes(query) ||
        role.description.toLocaleLowerCase().includes(query)
      )
    })
  }, [roleQuery, roles])

  const totalRolePages = Math.max(1, Math.ceil(filteredRoles.length / ROLES_PAGE_SIZE))
  const currentRolePage = Math.min(rolePage, totalRolePages)
  const rolesRangeStart = filteredRoles.length === 0 ? 0 : (currentRolePage - 1) * ROLES_PAGE_SIZE + 1
  const rolesRangeEnd = Math.min(currentRolePage * ROLES_PAGE_SIZE, filteredRoles.length)
  const paginatedRoles = filteredRoles.slice(
    (currentRolePage - 1) * ROLES_PAGE_SIZE,
    currentRolePage * ROLES_PAGE_SIZE
  )

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

  function setMenuPermissions(menuId: string, permissionIds: string[]) {
    setMatrix((prev) => ({
      ...prev,
      [menuId]: new Set(permissionIds),
    }))
    setSaveSuccess(false)
  }

  function setVisiblePermissions(grantAll: boolean) {
    const nextPermissionIds = grantAll ? sortedPermissions.map((permission) => permission.id) : []

    setMatrix((prev) => {
      const next = { ...prev }
      for (const menu of filteredMenus) {
        next[menu.id] = new Set(nextPermissionIds)
      }
      return next
    })
    setSaveSuccess(false)
  }

  async function handleSave() {
    if (!selectedRoleId || !hasUnsavedChanges) return

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

  function openPermissions(role: ApiRole) {
    setSelectedRoleId(role.id)
    setMatrix(buildRoleMatrix(role))
    setSaveError(null)
    setSaveSuccess(false)
    setMenuQuery('')
    setView('permissions')
  }

  function openCreateForm() {
    setFormRoleId(null)
    setFormName('')
    setFormDescription('')
    setFormIsActive(true)
    setFormError(null)
    setFormOpen(true)
  }

  function openEditForm(role: ApiRole) {
    setFormRoleId(role.id)
    setFormName(role.name)
    setFormDescription(role.description)
    setFormIsActive(role.is_active)
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
      let savedRole: ApiRole
      if (formRoleId === null) {
        const { data } = await api.post<ApiRole>('/roles', {
          name,
          description: formDescription,
          is_active: formIsActive,
        })
        savedRole = data
        setRoles((prev) => [...(prev ?? []), data])
      } else {
        const { data } = await api.put<ApiRole>(`/roles/${formRoleId}`, {
          name,
          description: formDescription,
          is_active: formIsActive,
        })
        savedRole = data
        setRoles((prev) => prev?.map((role) => (role.id === data.id ? data : role)) ?? prev)
      }
      setFormOpen(false)
      openPermissions(savedRole)
    } catch (err) {
      const fallback = formRoleId === null ? t('rolePermissionsCreateError') : t('rolePermissionsUpdateError')
      setFormError(extractErrorMessage(err, fallback))
    } finally {
      setFormSaving(false)
    }
  }

  async function handleDeleteRole(role: ApiRole) {
    if (!window.confirm(t('rolePermissionsDeleteConfirm'))) return

    setDeletingRoleId(role.id)
    setDeleteError(null)

    try {
      await api.delete(`/roles/${role.id}`)
      setRoles((prev) => prev?.filter((item) => item.id !== role.id) ?? prev)
      if (selectedRoleId === role.id) {
        setView('list')
        setSelectedRoleId(null)
      }
    } catch (err) {
      setDeleteError(extractErrorMessage(err, t('rolePermissionsDeleteError')))
    } finally {
      setDeletingRoleId(null)
    }
  }

  const isLoading = !loadError && (roles === null || menus === null || permissions === null)

  return (
    <main className="content">
      <section className="welcome">
        <h1>{t('menuRoles')}</h1>
        <p>{t('menuRolesDescription')}</p>
      </section>

      {view === 'list' && (
        <Card>
          <CardHeader className="flex flex-row items-start justify-between gap-4">
            <div>
              <CardTitle>{t('menuRoles')}</CardTitle>             
            </div>
            <Button onClick={openCreateForm} disabled={isLoading}>
              {t('rolePermissionsCreateRole')}
            </Button>
          </CardHeader>
          <CardContent className="flex flex-col gap-4">
            {loadError && <p className="resource-error">{loadError}</p>}
            {deleteError && <p className="resource-error">{deleteError}</p>}

            {!loadError && isLoading && <p className="metric-detail">{t('loading')}</p>}

            {!loadError && !isLoading && roles && roles.length === 0 && (
              <p className="metric-detail">{t('rolePermissionsNoRoles')}</p>
            )}

            {!loadError && !isLoading && roles && roles.length > 0 && (
              <>
                <label className="flex w-full max-w-md flex-col gap-1.5 text-sm font-medium">
                  {t('rolePermissionsRoleSearchLabel')}
                  <input
                    type="search"
                    className="h-10 rounded-md border border-input bg-transparent px-3 text-sm"
                    placeholder={t('rolePermissionsRoleSearchPlaceholder')}
                    value={roleQuery}
                    onChange={(event) => {
                      setRoleQuery(event.target.value)
                      setRolePage(1)
                    }}
                  />
                </label>

                {filteredRoles.length === 0 && (
                  <p className="metric-detail">{t('rolePermissionsNoMatchingRoles')}</p>
                )}

                {filteredRoles.length > 0 && (
                  <div className="table-wrap">
                    <Table>
                      <TableHeader>
                        <TableRow>
                          <TableHead>{t('rolePermissionsNameColumn')}</TableHead>
                          <TableHead>{t('rolePermissionsDescriptionColumn')}</TableHead>
                          <TableHead>{t('rolePermissionsStatusColumn')}</TableHead>
                          <TableHead className="text-right">{t('rolePermissionsActionsColumn')}</TableHead>
                        </TableRow>
                      </TableHeader>
                      <TableBody>
                        {paginatedRoles.map((role) => (
                          <TableRow key={role.id}>
                            <TableCell className="font-semibold">{role.name}</TableCell>
                            <TableCell className="text-muted-foreground">
                              {role.description || t('rolePermissionsDescriptionEmpty')}
                            </TableCell>
                            <TableCell>
                              <Badge variant={role.is_active ? 'default' : 'outline'}>
                                {role.is_active ? t('statusActive') : t('statusInactive')}
                              </Badge>
                            </TableCell>
                            <TableCell className="text-right">
                              <div className="flex flex-wrap justify-end gap-2">
                                <Button
                                  type="button"
                                  size="icon"
                                  variant="outline"
                                  title={t('rolePermissionsManage')}
                                  aria-label={t('rolePermissionsManage')}
                                  onClick={() => openPermissions(role)}
                                >
                                  <KeyRound />
                                </Button>
                                <Button
                                  type="button"
                                  size="icon"
                                  variant="outline"
                                  title={t('rolePermissionsEditRole')}
                                  aria-label={t('rolePermissionsEditRole')}
                                  onClick={() => openEditForm(role)}
                                >
                                  <Pencil />
                                </Button>
                                <Button
                                  type="button"
                                  size="icon"
                                  variant="destructive"
                                  title={t('rolePermissionsDeleteRole')}
                                  aria-label={t('rolePermissionsDeleteRole')}
                                  onClick={() => handleDeleteRole(role)}
                                  disabled={deletingRoleId === role.id}
                                >
                                  <Trash2 />
                                </Button>
                              </div>
                            </TableCell>
                          </TableRow>
                        ))}
                      </TableBody>
                    </Table>
                  </div>
                )}

                {filteredRoles.length > 0 && (
                  <div className="flex items-center justify-between gap-3">
                    <p className="text-sm text-muted-foreground">
                      {t('rolePermissionsShowingLabel')} {rolesRangeStart}-{rolesRangeEnd}{' '}
                      {t('rolePermissionsOfLabel')} {filteredRoles.length} {t('rolePermissionsResultsLabel')}
                      {totalRolePages > 1 && (
                        <>
                          {' '}
                          · {t('rolePermissionsPageLabel')} {currentRolePage} / {totalRolePages}
                        </>
                      )}
                    </p>
                    {totalRolePages > 1 && (
                      <div className="flex gap-2">
                        <Button
                          type="button"
                          size="icon"
                          variant="outline"
                          title={t('rolePermissionsPrevPage')}
                          aria-label={t('rolePermissionsPrevPage')}
                          disabled={currentRolePage <= 1}
                          onClick={() => setRolePage((page) => Math.max(1, page - 1))}
                        >
                          <ChevronLeft />
                        </Button>
                        <Button
                          type="button"
                          size="icon"
                          variant="outline"
                          title={t('rolePermissionsNextPage')}
                          aria-label={t('rolePermissionsNextPage')}
                          disabled={currentRolePage >= totalRolePages}
                          onClick={() => setRolePage((page) => Math.min(totalRolePages, page + 1))}
                        >
                          <ChevronRight />
                        </Button>
                      </div>
                    )}
                  </div>
                )}
              </>
            )}
          </CardContent>
        </Card>
      )}

      {view === 'permissions' && selectedRole && (
        <Card>
          <CardHeader className="flex flex-row items-start justify-between gap-4">
            <div>
              <Button
                type="button"
                variant="ghost"
                size="sm"
                className="-ml-3 mb-2"
                onClick={() => setView('list')}
              >
                {t('rolePermissionsBackToRoles')}
              </Button>
              <CardTitle className="flex items-center gap-2">
                {selectedRole.name}
                <Badge variant={selectedRole.is_active ? 'default' : 'outline'}>
                  {selectedRole.is_active ? t('statusActive') : t('statusInactive')}
                </Badge>
              </CardTitle>
              <CardDescription>
                {selectedRole.description || t('rolePermissionsDescriptionEmpty')}
              </CardDescription>
            </div>
            <Button type="button" variant="outline" onClick={() => openEditForm(selectedRole)}>
              {t('rolePermissionsEditRole')}
            </Button>
          </CardHeader>
          <CardContent className="flex flex-col gap-4">
            <div className="flex flex-col gap-3 rounded-lg border border-border bg-muted/30 p-3 lg:flex-row lg:items-end lg:justify-between">
              <label className="flex w-full max-w-md flex-col gap-1.5 text-sm font-medium">
                {t('rolePermissionsSearchLabel')}
                <input
                  type="search"
                  className="h-10 rounded-md border border-input bg-transparent px-3 text-sm"
                  placeholder={t('rolePermissionsSearchPlaceholder')}
                  value={menuQuery}
                  onChange={(event) => setMenuQuery(event.target.value)}
                />
              </label>

              <div className="flex flex-wrap gap-2">
                <Button
                  type="button"
                  variant="outline"
                  onClick={() => setVisiblePermissions(true)}
                  disabled={filteredMenus.length === 0 || sortedPermissions.length === 0}
                >
                  {t('rolePermissionsSelectAllVisible')}
                </Button>
                <Button
                  type="button"
                  variant="outline"
                  onClick={() => setVisiblePermissions(false)}
                  disabled={filteredMenus.length === 0}
                >
                  {t('rolePermissionsClearVisible')}
                </Button>
              </div>
            </div>

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
                  {filteredMenus.map((menu) => (
                    <TableRow key={menu.id}>
                      <TableCell className="align-top">
                        <div className="flex min-w-48 flex-col gap-2">
                          <div>
                            <p className="font-semibold">{menuLabel(menu, t)}</p>
                            <p className="text-xs text-muted-foreground">{menu.path}</p>
                          </div>
                          <div className="flex flex-wrap gap-2">
                            <Button
                              type="button"
                              size="sm"
                              variant="outline"
                              onClick={() =>
                                setMenuPermissions(
                                  menu.id,
                                  sortedPermissions.map((permission) => permission.id)
                                )
                              }
                              disabled={sortedPermissions.length === 0}
                            >
                              {t('rolePermissionsSelectAllRow')}
                            </Button>
                            <Button
                              type="button"
                              size="sm"
                              variant="ghost"
                              onClick={() => setMenuPermissions(menu.id, [])}
                            >
                              {t('rolePermissionsClearRow')}
                            </Button>
                          </div>
                        </div>
                      </TableCell>
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
                  {filteredMenus.length === 0 && (
                    <TableRow>
                      <TableCell colSpan={sortedPermissions.length + 1} className="py-6 text-center text-muted-foreground">
                        {t('rolePermissionsNoVisibleMenus')}
                      </TableCell>
                    </TableRow>
                  )}
                </TableBody>
              </Table>
            </div>

            <div className="flex items-center gap-3">
              <Button onClick={handleSave} disabled={saving || !hasUnsavedChanges}>
                {saving ? t('rolePermissionsSaving') : t('rolePermissionsSave')}
              </Button>
              {saveSuccess && <p className="metric-detail">{t('rolePermissionsSaved')}</p>}
              {!saveSuccess && hasUnsavedChanges && <p className="metric-detail">{t('rolePermissionsUnsaved')}</p>}
              {saveError && <p className="resource-error">{saveError}</p>}
            </div>
          </CardContent>
        </Card>
      )}

      <Sheet open={formOpen} onOpenChange={setFormOpen}>
        <SheetContent>
          <form className="flex h-full flex-col gap-4" onSubmit={handleFormSubmit}>
            <SheetHeader>
              <SheetTitle>
                {formRoleId === null ? t('rolePermissionsFormCreateTitle') : t('rolePermissionsFormEditTitle')}
              </SheetTitle>
              <SheetDescription>
                {formRoleId === null
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
