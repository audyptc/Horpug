import { useEffect, useMemo, useState, type FormEvent } from 'react'
import { api, extractErrorMessage, type ApiPage } from '@/shared/api/client'
import { useLanguage } from '@/shared/i18n/language'
import { ConfirmDialog } from '@/shared/components/confirm-dialog'
import type { ApiMenu } from '@/features/menu/menus'
import type { ApiPermission, ApiRole } from './types'
import { ACTION_ORDER, ROLE_PAGE_SIZE_OPTIONS, areMatricesEqual, buildRoleMatrix, menuLabel } from './utils'
import { RoleListCard } from './components/RoleListCard'
import { RolePermissionMatrixCard } from './components/RolePermissionMatrixCard'
import { RoleFormSheet } from './components/RoleFormSheet'

type View = 'list' | 'permissions'

export default function RolePermissionsPage() {
  const { t } = useLanguage()

  const [roles, setRoles] = useState<ApiRole[] | null>(null)
  const [menus, setMenus] = useState<ApiMenu[] | null>(null)
  const [permissions, setPermissions] = useState<ApiPermission[] | null>(null)
  const [loadError, setLoadError] = useState<string | null>(null)

  const [view, setView] = useState<View>('list')
  const [roleQuery, setRoleQuery] = useState('')
  const [rolePage, setRolePage] = useState(1)
  const [rolePageSize, setRolePageSize] = useState<number>(ROLE_PAGE_SIZE_OPTIONS[0])
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
  const [confirmDeleteRole, setConfirmDeleteRole] = useState<ApiRole | null>(null)

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

  const totalRolePages = Math.max(1, Math.ceil(filteredRoles.length / rolePageSize))
  const currentRolePage = Math.min(rolePage, totalRolePages)
  const rolesRangeStart = filteredRoles.length === 0 ? 0 : (currentRolePage - 1) * rolePageSize + 1
  const rolesRangeEnd = Math.min(currentRolePage * rolePageSize, filteredRoles.length)
  const paginatedRoles = filteredRoles.slice(
    (currentRolePage - 1) * rolePageSize,
    currentRolePage * rolePageSize
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
      setView('list')
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

  async function handleDeleteRole() {
    if (!confirmDeleteRole) return
    const role = confirmDeleteRole

    setDeletingRoleId(role.id)
    setDeleteError(null)

    try {
      await api.delete(`/roles/${role.id}`)
      setRoles((prev) => prev?.filter((item) => item.id !== role.id) ?? prev)
      if (selectedRoleId === role.id) {
        setView('list')
        setSelectedRoleId(null)
      }
      setConfirmDeleteRole(null)
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
        <RoleListCard
          isLoading={isLoading}
          loadError={loadError}
          deleteError={deleteError}
          roles={roles}
          roleQuery={roleQuery}
          onRoleQueryChange={(query) => {
            setRoleQuery(query)
            setRolePage(1)
          }}
          filteredRoles={filteredRoles}
          paginatedRoles={paginatedRoles}
          currentRolePage={currentRolePage}
          totalRolePages={totalRolePages}
          rolesRangeStart={rolesRangeStart}
          rolesRangeEnd={rolesRangeEnd}
          rolePageSize={rolePageSize}
          onRolePageSizeChange={(size) => {
            setRolePageSize(size)
            setRolePage(1)
          }}
          onPrevPage={() => setRolePage((page) => Math.max(1, page - 1))}
          onNextPage={() => setRolePage((page) => Math.min(totalRolePages, page + 1))}
          deletingRoleId={deletingRoleId}
          onCreateRole={openCreateForm}
          onManageRole={openPermissions}
          onEditRole={openEditForm}
          onDeleteRole={setConfirmDeleteRole}
        />
      )}

      {view === 'permissions' && selectedRole && (
        <RolePermissionMatrixCard
          selectedRole={selectedRole}
          menuQuery={menuQuery}
          onMenuQueryChange={setMenuQuery}
          filteredMenus={filteredMenus}
          sortedPermissions={sortedPermissions}
          matrix={matrix}
          onToggleCell={toggleCell}
          onSetMenuPermissions={setMenuPermissions}
          onSetVisiblePermissions={setVisiblePermissions}
          onBack={() => setView('list')}
          onEditRole={() => openEditForm(selectedRole)}
          onSave={handleSave}
          saving={saving}
          saveError={saveError}
          saveSuccess={saveSuccess}
          hasUnsavedChanges={hasUnsavedChanges}
        />
      )}

      <ConfirmDialog
        open={confirmDeleteRole !== null}
        onOpenChange={(open) => !open && setConfirmDeleteRole(null)}
        title={t('confirmDeleteTitle')}
        description={t('rolePermissionsDeleteConfirm')}
        confirmLabel={t('rolePermissionsDeleteRole')}
        cancelLabel={t('cancel')}
        loading={deletingRoleId === confirmDeleteRole?.id}
        onConfirm={handleDeleteRole}
      />

      <RoleFormSheet
        open={formOpen}
        onOpenChange={setFormOpen}
        isEdit={formRoleId !== null}
        name={formName}
        onNameChange={setFormName}
        description={formDescription}
        onDescriptionChange={setFormDescription}
        isActive={formIsActive}
        onIsActiveChange={setFormIsActive}
        saving={formSaving}
        error={formError}
        onSubmit={handleFormSubmit}
      />
    </main>
  )
}
