import { useEffect, useState, type FormEvent } from 'react'
import axios from 'axios'
import { api, extractErrorMessage, type ApiPage } from '@/shared/api/client'
import { useLanguage } from '@/shared/i18n/language'
import { ConfirmDialog } from '@/shared/components/confirm-dialog'
import { InformationDialog } from '@/shared/components/information-dialog'
import { UserListCard } from './components/UserListCard'
import { UserFormSheet } from './components/UserFormSheet'
import type { ApiUser, ApiUserDeletionCheck, ApiUserRole } from './types'
import {
  USER_PAGE_SIZE_OPTIONS,
  type UserColumnFilters,
  type UserSortDirection,
  type UserSortKey,
  type UserStatusFilter,
  type UserTextFilterKey,
} from './utils'

const SEARCH_DEBOUNCE_MS = 300

export default function UserPage() {
  const { t } = useLanguage()

  const [users, setUsers] = useState<ApiUser[] | null>(null)
  const [roles, setRoles] = useState<ApiUserRole[]>([])
  const [total, setTotal] = useState(0)
  const [totalPages, setTotalPages] = useState(1)
  const [loadError, setLoadError] = useState<string | null>(null)

  const [query, setQuery] = useState('')
  const [debouncedQuery, setDebouncedQuery] = useState('')
  const [statusFilter, setStatusFilter] = useState<UserStatusFilter>('all')
  // Applied on submit from each column's menu, so no debounce is needed here.
  const [columnFilters, setColumnFilters] = useState<UserColumnFilters>({})
  const [sortKey, setSortKey] = useState<UserSortKey>('username')
  const [sortDirection, setSortDirection] = useState<UserSortDirection>('asc')

  const [page, setPage] = useState(1)
  const [pageSize, setPageSizeState] = useState<number>(USER_PAGE_SIZE_OPTIONS[0])
  // Bumped by mutations so the list refetches; the server owns the ordering
  // and page boundaries now, so patching rows locally would misplace them.
  const [refreshToken, setRefreshToken] = useState(0)

  const [formOpen, setFormOpen] = useState(false)
  const [formUserId, setFormUserId] = useState<string | null>(null)
  const [formUsername, setFormUsername] = useState('')
  const [formEmail, setFormEmail] = useState('')
  const [formPassword, setFormPassword] = useState('')
  const [formRoleId, setFormRoleId] = useState('')
  const [formIsActive, setFormIsActive] = useState(true)
  const [formSaving, setFormSaving] = useState(false)
  const [formError, setFormError] = useState<string | null>(null)

  const [deletingUserId, setDeletingUserId] = useState<string | null>(null)
  const [deleteError, setDeleteError] = useState<string | null>(null)
  const [confirmDeleteUser, setConfirmDeleteUser] = useState<ApiUser | null>(null)
  const [checkingUserId, setCheckingUserId] = useState<string | null>(null)
  const [blockedUserDeletion, setBlockedUserDeletion] = useState<ApiUserDeletionCheck | null>(null)

  const [togglingUserId, setTogglingUserId] = useState<string | null>(null)
  const [toggleError, setToggleError] = useState<string | null>(null)

  useEffect(() => {
    const timer = window.setTimeout(() => setDebouncedQuery(query), SEARCH_DEBOUNCE_MS)
    return () => window.clearTimeout(timer)
  }, [query])

  // The role selector in the form isn't affected by the list's filters, so
  // it's loaded once rather than on every refetch.
  useEffect(() => {
    let cancelled = false

    api
      .get<ApiPage<ApiUserRole[]>>('/roles', { params: { per_page: 100 } })
      .then(({ data }) => {
        if (!cancelled) setRoles(data.data)
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
      .get<ApiPage<ApiUser[]>>('/users', {
        signal: controller.signal,
        params: {
          page,
          per_page: pageSize,
          q: debouncedQuery.trim() || undefined,
          is_active: statusFilter === 'all' ? undefined : statusFilter === 'active',
          sort: sortKey,
          order: sortDirection,
          ...columnParams,
        },
      })
      .then(({ data }) => {
        setUsers(data.data)
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
  }, [page, pageSize, debouncedQuery, statusFilter, columnFilters, sortKey, sortDirection, refreshToken])

  const isLoading = !loadError && users === null
  const hasFilters =
    query !== '' || statusFilter !== 'all' || Object.values(columnFilters).some(Boolean)
  const rangeStart = total === 0 ? 0 : (page - 1) * pageSize + 1
  const rangeEnd = Math.min(page * pageSize, total)

  function refresh() {
    setRefreshToken((value) => value + 1)
  }

  function setPageSize(size: number) {
    setPageSizeState(size)
    setPage(1)
  }

  function handleSort(key: UserSortKey) {
    if (key === sortKey) {
      setSortDirection(sortDirection === 'asc' ? 'desc' : 'asc')
    } else {
      setSortKey(key)
      setSortDirection('asc')
    }
    setPage(1)
  }

  function openCreateForm() {
    setFormUserId(null)
    setFormUsername('')
    setFormEmail('')
    setFormPassword('')
    setFormRoleId('')
    setFormIsActive(true)
    setFormError(null)
    setFormOpen(true)
  }

  function openEditForm(user: ApiUser) {
    setFormUserId(user.id)
    setFormUsername(user.username)
    setFormEmail(user.email)
    setFormPassword('')
    setFormRoleId(user.role_id)
    setFormIsActive(user.is_active)
    setFormError(null)
    setFormOpen(true)
  }

  async function handleFormSubmit(event: FormEvent<HTMLFormElement>) {
    event.preventDefault()

    const username = formUsername.trim()
    const email = formEmail.trim()
    const isEdit = formUserId !== null

    if (!username || !email) {
      setFormError(t('userRequiredFields'))
      return
    }
    if (!formRoleId) {
      setFormError(t('userRoleRequired'))
      return
    }
    if (!isEdit && !formPassword.trim()) {
      setFormError(t('userPasswordRequired'))
      return
    }

    setFormSaving(true)
    setFormError(null)

    try {
      if (!isEdit) {
        await api.post<ApiUser>('/users', {
          username,
          email,
          password: formPassword,
          role_id: formRoleId,
          is_active: formIsActive,
        })
      } else {
        const payload: Record<string, unknown> = {
          username,
          email,
          role_id: formRoleId,
          is_active: formIsActive,
        }
        if (formPassword.trim()) {
          payload.password = formPassword
        }
        await api.put<ApiUser>(`/users/${formUserId}`, payload)
      }
      refresh()
      setFormOpen(false)
    } catch (err) {
      const fallback = isEdit ? t('userUpdateError') : t('userCreateError')
      setFormError(extractErrorMessage(err, fallback))
    } finally {
      setFormSaving(false)
    }
  }

  async function handleRequestDeleteUser(user: ApiUser) {
    setCheckingUserId(user.id)
    setDeleteError(null)

    try {
      const { data } = await api.get<ApiUserDeletionCheck>(`/users/${user.id}/deletion-check`)
      if (data.can_delete) {
        setConfirmDeleteUser(user)
      } else {
        setBlockedUserDeletion(data)
      }
    } catch (err) {
      setDeleteError(extractErrorMessage(err, t('userDeleteError')))
    } finally {
      setCheckingUserId(null)
    }
  }

  async function handleToggleActiveUser(user: ApiUser) {
    setTogglingUserId(user.id)
    setToggleError(null)

    try {
      await api.put<ApiUser>(`/users/${user.id}`, {
        is_active: !user.is_active,
      })
      refresh()
    } catch (err) {
      setToggleError(extractErrorMessage(err, t('userToggleActiveError')))
    } finally {
      setTogglingUserId(null)
    }
  }

  async function handleDeleteUser() {
    if (!confirmDeleteUser) return
    const user = confirmDeleteUser

    setDeletingUserId(user.id)
    setDeleteError(null)

    try {
      await api.delete(`/users/${user.id}`)
      // Deleting the last row of the final page would otherwise strand the
      // view on a page the server no longer has.
      setPage((value) => (users?.length === 1 ? Math.max(1, value - 1) : value))
      refresh()
      setConfirmDeleteUser(null)
    } catch (err) {
      setDeleteError(extractErrorMessage(err, t('userDeleteError')))
    } finally {
      setDeletingUserId(null)
    }
  }

  return (
    <main className="content">
      <section className="welcome">
        <h1>{t('menuUsers')}</h1>
        <p>{t('menuUsersDescription')}</p>
      </section>

      <UserListCard
        isLoading={isLoading}
        loadError={loadError}
        deleteError={deleteError}
        toggleError={toggleError}
        query={query}
        onQueryChange={(value) => {
          setQuery(value)
          setPage(1)
        }}
        statusFilter={statusFilter}
        onStatusFilterChange={(value) => {
          setStatusFilter(value)
          setPage(1)
        }}
        columnFilters={columnFilters}
        onColumnFilterChange={(key: UserTextFilterKey, value: string) => {
          setColumnFilters((prev) => ({ ...prev, [key]: value }))
          setPage(1)
        }}
        hasFilters={hasFilters}
        sortKey={sortKey}
        sortDirection={sortDirection}
        onSort={handleSort}
        users={users ?? []}
        total={total}
        currentPage={page}
        totalPages={totalPages}
        rangeStart={rangeStart}
        rangeEnd={rangeEnd}
        pageSize={pageSize}
        onPageSizeChange={setPageSize}
        onFirstPage={() => setPage(1)}
        onPrevPage={() => setPage((value) => Math.max(1, value - 1))}
        onNextPage={() => setPage((value) => Math.min(totalPages, value + 1))}
        onLastPage={() => setPage(totalPages)}
        deletingUserId={checkingUserId ?? deletingUserId}
        togglingUserId={togglingUserId}
        onCreateUser={openCreateForm}
        onEditUser={openEditForm}
        onDeleteUser={handleRequestDeleteUser}
        onToggleActiveUser={handleToggleActiveUser}
      />

      <ConfirmDialog
        open={confirmDeleteUser !== null}
        onOpenChange={(open) => !open && setConfirmDeleteUser(null)}
        title={t('confirmDeleteTitle')}
        description={t('userDeleteConfirm')}
        confirmLabel={t('userDelete')}
        cancelLabel={t('cancel')}
        loading={deletingUserId === confirmDeleteUser?.id}
        error={deleteError}
        onConfirm={handleDeleteUser}
      />

      <InformationDialog
        open={blockedUserDeletion !== null}
        onOpenChange={(open) => !open && setBlockedUserDeletion(null)}
        title={t('userDeleteBlockedTitle')}
        description={
          blockedUserDeletion?.is_protected
            ? t('userProtectedHint')
            : t('userDeleteBlockedDescription').replace(
                '{count}',
                String(blockedUserDeletion?.record_count ?? 0)
              )
        }
        actionLabel={t('acknowledge')}
      />

      <UserFormSheet
        open={formOpen}
        onOpenChange={setFormOpen}
        isEdit={formUserId !== null}
        username={formUsername}
        onUsernameChange={setFormUsername}
        email={formEmail}
        onEmailChange={setFormEmail}
        password={formPassword}
        onPasswordChange={setFormPassword}
        roles={roles}
        roleId={formRoleId}
        onRoleIdChange={setFormRoleId}
        isActive={formIsActive}
        onIsActiveChange={setFormIsActive}
        saving={formSaving}
        error={formError}
        onSubmit={handleFormSubmit}
      />
    </main>
  )
}
