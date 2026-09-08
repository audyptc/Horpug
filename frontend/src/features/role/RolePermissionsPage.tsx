import { useEffect, useMemo, useState, type FormEvent } from 'react'
import axios from 'axios'
import { api, extractErrorMessage, type ApiPage } from '@/shared/api/client'
import { useLanguage } from '@/shared/i18n/language'
import { ConfirmDialog } from '@/shared/components/confirm-dialog'
import { InformationDialog } from '@/shared/components/information-dialog'
import type { ApiMenu } from '@/features/menu/menus'
import type { ApiPermission, ApiRole, ApiRoleDeletionCheck } from './types'
import {
  ACTION_ORDER,
  ROLE_PAGE_SIZE_OPTIONS,
  areMatricesEqual,
  buildRoleMatrix,
  menuLabel,
  type RoleColumnFilters,
  type RoleSortDirection,
  type RoleSortKey,
  type RoleStatusFilter,
  type RoleTextFilterKey,
} from './utils'
import { RoleListCard } from './components/RoleListCard'
import { RolePermissionMatrixCard } from './components/RolePermissionMatrixCard'
import { RoleFormSheet } from './components/RoleFormSheet'

type View = 'list' | 'permissions'

const SEARCH_DEBOUNCE_MS = 300

export default function RolePermissionsPage() {
  const { t } = useLanguage()

  const [roles, setRoles] = useState<ApiRole[] | null>(null)
  const [total, setTotal] = useState(0)
  const [totalPages, setTotalPages] = useState(1)
  const [menus, setMenus] = useState<ApiMenu[] | null>(null)
  const [permissions, setPermissions] = useState<ApiPermission[] | null>(null)
  const [loadError, setLoadError] = useState<string | null>(null)

  const [roleQuery, setRoleQuery] = useState('')
  const [debouncedRoleQuery, setDebouncedRoleQuery] = useState('')
  const [statusFilter, setStatusFilter] = useState<RoleStatusFilter>('all')
  // Applied on submit from each column's menu, so no debounce is needed here.
  const [columnFilters, setColumnFilters] = useState<RoleColumnFilters>({})
  const [sortKey, setSortKey] = useState<RoleSortKey>('name')
  const [sortDirection, setSortDirection] = useState<RoleSortDirection>('asc')

  const [page, setPage] = useState(1)
  const [pageSize, setPageSizeState] = useState<number>(ROLE_PAGE_SIZE_OPTIONS[0])
  // Bumped by mutations so the list refetches; the server owns the ordering
  // and page boundaries now, so patching rows locally would misplace them.
  const [refreshToken, setRefreshToken] = useState(0)

  const [view, setView] = useState<View>('list')
  // Held directly rather than looked up by id from `roles`, since `roles` now
  // only holds the current page — the selected role may not be on it.
  const [selectedRole, setSelectedRole] = useState<ApiRole | null>(null)
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
  const [checkingRoleId, setCheckingRoleId] = useState<string | null>(null)
  const [deleteError, setDeleteError] = useState<string | null>(null)
  const [confirmDeleteRole, setConfirmDeleteRole] = useState<ApiRole | null>(null)
  const [blockedRoleDeletion, setBlockedRoleDeletion] = useState<ApiRoleDeletionCheck | null>(null)

  useEffect(() => {
    const timer = window.setTimeout(() => setDebouncedRoleQuery(roleQuery), SEARCH_DEBOUNCE_MS)
    return () => window.clearTimeout(timer)
  }, [roleQuery])

  // The permission matrix's menus and permissions aren't affected by the
  // role list's filters, so they're loaded once rather than on every refetch.
  useEffect(() => {
    let cancelled = false

    Promise.all([
      api.get<ApiPage<ApiMenu[]>>('/menus', { params: { per_page: 100 } }),
      api.get<ApiPage<ApiPermission[]>>('/permissions', { params: { per_page: 100 } }),
    ])
      .then(([menusRes, permissionsRes]) => {
        if (cancelled) return
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

  useEffect(() => {
    const controller = new AbortController()

    const columnParams: Record<string, string> = {}
    for (const [key, value] of Object.entries(columnFilters)) {
      if (value) columnParams[`f[${key}]`] = value
    }

    api
      .get<ApiPage<ApiRole[]>>('/roles', {
        signal: controller.signal,
        params: {
          page,
          per_page: pageSize,
          q: debouncedRoleQuery.trim() || undefined,
          is_active: statusFilter === 'all' ? undefined : statusFilter === 'active',
          sort: sortKey,
          order: sortDirection,
          ...columnParams,
        },
      })
      .then(({ data }) => {
        setRoles(data.data)
        setTotal(data.meta.total)
        setTotalPages(data.meta.total_pages)
        setLoadError(null)
      })
      .catch((err) => {
        // A cancelled request is this effect being superseded by a newer one,
        // not a failure — letting it through would overwrite the newer result.
        if (axios.isCancel(err)) return
        setLoadError(extractErrorMessage(err, t('resourceLoadError')))
      })

    return () => controller.abort()
    // eslint-disable-next-line react-hooks/exhaustive-deps
  }, [page, pageSize, debouncedRoleQuery, statusFilter, columnFilters, sortKey, sortDirection, refreshToken])

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

  const isLoading = !loadError && (roles === null || menus === null || permissions === null)
  const hasFilters =
    roleQuery !== '' || statusFilter !== 'all' || Object.values(columnFilters).some(Boolean)
  const rangeStart = total === 0 ? 0 : (page - 1) * pageSize + 1
  const rangeEnd = Math.min(page * pageSize, total)

  function refresh() {
    setRefreshToken((value) => value + 1)
  }

  function setPageSize(size: number) {
    setPageSizeState(size)
    setPage(1)
  }

  function handleSort(key: RoleSortKey) {
    if (key === sortKey) {
      setSortDirection(sortDirection === 'asc' ? 'desc' : 'asc')
    } else {
      setSortKey(key)
      setSortDirection('asc')
    }
    setPage(1)
  }

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
    if (!selectedRole || !hasUnsavedChanges) return

    setSaving(true)
    setSaveError(null)
    setSaveSuccess(false)

    const menu_permissions = Object.entries(matrix)
      .filter(([, permissionIds]) => permissionIds.size > 0)
      .map(([menu_id, permissionIds]) => ({ menu_id, permission_ids: Array.from(permissionIds) }))

    try {
      await api.put<ApiRole>(`/roles/${selectedRole.id}`, { menu_permissions })
      refresh()
      setSaveSuccess(true)
      setView('list')
    } catch (err) {
      setSaveError(extractErrorMessage(err, t('rolePermissionsSaveError')))
    } finally {
      setSaving(false)
    }
  }

  function openPermissions(role: ApiRole) {
    setSelectedRole(role)
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
      } else {
        const { data } = await api.put<ApiRole>(`/roles/${formRoleId}`, {
          name,
          description: formDescription,
          is_active: formIsActive,
        })
        savedRole = data
      }
      refresh()
      setFormOpen(false)
      openPermissions(savedRole)
    } catch (err) {
      const fallback = formRoleId === null ? t('rolePermissionsCreateError') : t('rolePermissionsUpdateError')
      setFormError(extractErrorMessage(err, fallback))
    } finally {
      setFormSaving(false)
    }
  }

  async function handleRequestDeleteRole(role: ApiRole) {
    setCheckingRoleId(role.id)
    setDeleteError(null)

    try {
      const { data } = await api.get<ApiRoleDeletionCheck>(`/roles/${role.id}/deletion-check`)
      if (data.can_delete) {
        setConfirmDeleteRole(role)
      } else {
        setBlockedRoleDeletion(data)
      }
    } catch (err) {
      setDeleteError(extractErrorMessage(err, t('rolePermissionsDeleteError')))
    } finally {
      setCheckingRoleId(null)
    }
  }

  async function handleDeleteRole() {
    if (!confirmDeleteRole) return
    const role = confirmDeleteRole

    setDeletingRoleId(role.id)
    setDeleteError(null)

    try {
      await api.delete(`/roles/${role.id}`)
      // Deleting the last row of the final page would otherwise strand the
      // view on a page the server no longer has.
      setPage((value) => (roles?.length === 1 ? Math.max(1, value - 1) : value))
      refresh()
      if (selectedRole?.id === role.id) {
        setView('list')
        setSelectedRole(null)
      }
      setConfirmDeleteRole(null)
    } catch (err) {
      setDeleteError(extractErrorMessage(err, t('rolePermissionsDeleteError')))
    } finally {
      setDeletingRoleId(null)
    }
  }

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
          roleQuery={roleQuery}
          onRoleQueryChange={(query) => {
            setRoleQuery(query)
            setPage(1)
          }}
          statusFilter={statusFilter}
          onStatusFilterChange={(value) => {
            setStatusFilter(value)
            setPage(1)
          }}
          columnFilters={columnFilters}
          onColumnFilterChange={(key: RoleTextFilterKey, value: string) => {
            setColumnFilters((prev) => ({ ...prev, [key]: value }))
            setPage(1)
          }}
          hasFilters={hasFilters}
          sortKey={sortKey}
          sortDirection={sortDirection}
          onSort={handleSort}
          roles={roles ?? []}
          total={total}
          currentRolePage={page}
          totalRolePages={totalPages}
          rolesRangeStart={rangeStart}
          rolesRangeEnd={rangeEnd}
          rolePageSize={pageSize}
          onRolePageSizeChange={setPageSize}
          onFirstPage={() => setPage(1)}
          onPrevPage={() => setPage((value) => Math.max(1, value - 1))}
          onNextPage={() => setPage((value) => Math.min(totalPages, value + 1))}
          onLastPage={() => setPage(totalPages)}
          deletingRoleId={checkingRoleId ?? deletingRoleId}
          onCreateRole={openCreateForm}
          onManageRole={openPermissions}
          onEditRole={openEditForm}
          onDeleteRole={handleRequestDeleteRole}
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
        error={deleteError}
        onConfirm={handleDeleteRole}
      />

      <InformationDialog
        open={blockedRoleDeletion !== null}
        onOpenChange={(open) => !open && setBlockedRoleDeletion(null)}
        title={t('rolePermissionsDeleteBlockedTitle')}
        description={
          blockedRoleDeletion?.is_protected
            ? t('rolePermissionsProtectedHint')
            : t('rolePermissionsDeleteBlockedDescription').replace(
                '{count}',
                String(blockedRoleDeletion?.user_count ?? 0)
              )
        }
        actionLabel={t('acknowledge')}
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
