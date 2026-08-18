import { useEffect, useMemo, useState, type FormEvent } from 'react'
import { api, extractErrorMessage, type ApiPage } from '@/shared/api/client'
import { usePagination } from '@/shared/hooks/use-pagination'
import { useLanguage } from '@/shared/i18n/language'
import { ConfirmDialog } from '@/shared/components/confirm-dialog'
import { InformationDialog } from '@/shared/components/information-dialog'
import { UserListCard } from './components/UserListCard'
import { UserFormSheet } from './components/UserFormSheet'
import type { ApiUser, ApiUserDeletionCheck, ApiUserRole } from './types'
import { USER_PAGE_SIZE_OPTIONS } from './utils'

export default function UserPage() {
  const { t } = useLanguage()

  const [users, setUsers] = useState<ApiUser[] | null>(null)
  const [roles, setRoles] = useState<ApiUserRole[]>([])
  const [loadError, setLoadError] = useState<string | null>(null)

  const [query, setQuery] = useState('')

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
    let cancelled = false

    Promise.all([
      api.get<ApiPage<ApiUser[]>>('/users', { params: { per_page: 100 } }),
      api.get<ApiPage<ApiUserRole[]>>('/roles', { params: { per_page: 100 } }),
    ])
      .then(([usersRes, rolesRes]) => {
        if (cancelled) return
        setUsers(usersRes.data.data)
        setRoles(rolesRes.data.data)
      })
      .catch((err) => {
        if (!cancelled) setLoadError(extractErrorMessage(err, t('resourceLoadError')))
      })

    return () => {
      cancelled = true
    }
    // eslint-disable-next-line react-hooks/exhaustive-deps
  }, [])

  const filteredUsers = useMemo(() => {
    const q = query.trim().toLocaleLowerCase()
    if (!q) return users ?? []

    return (users ?? []).filter((user) => {
      return (
        user.username.toLocaleLowerCase().includes(q) ||
        user.email.toLocaleLowerCase().includes(q) ||
        (user.role?.name ?? '').toLocaleLowerCase().includes(q)
      )
    })
  }, [query, users])

  const {
    page: currentPage,
    pageSize,
    setPageSize,
    totalPages,
    rangeStart,
    rangeEnd,
    paginatedItems: paginatedUsers,
    resetPage,
    prevPage,
    nextPage,
  } = usePagination(filteredUsers, USER_PAGE_SIZE_OPTIONS[0])

  const isLoading = !loadError && users === null

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
        const { data } = await api.post<ApiUser>('/users', {
          username,
          email,
          password: formPassword,
          role_id: formRoleId,
          is_active: formIsActive,
        })
        setUsers((prev) => [...(prev ?? []), data])
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
        const { data } = await api.put<ApiUser>(`/users/${formUserId}`, payload)
        setUsers((prev) => prev?.map((item) => (item.id === data.id ? data : item)) ?? prev)
      }
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
      const { data } = await api.put<ApiUser>(`/users/${user.id}`, {
        is_active: !user.is_active,
      })
      setUsers((prev) => prev?.map((item) => (item.id === data.id ? data : item)) ?? prev)
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
      setUsers((prev) => prev?.filter((item) => item.id !== user.id) ?? prev)
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
        users={users}
        query={query}
        onQueryChange={(value) => {
          setQuery(value)
          resetPage()
        }}
        filteredUsers={filteredUsers}
        paginatedUsers={paginatedUsers}
        currentPage={currentPage}
        totalPages={totalPages}
        rangeStart={rangeStart}
        rangeEnd={rangeEnd}
        pageSize={pageSize}
        onPageSizeChange={setPageSize}
        onPrevPage={prevPage}
        onNextPage={nextPage}
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
