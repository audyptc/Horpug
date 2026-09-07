import { useEffect, useState, type FormEvent } from 'react'
import axios from 'axios'
import { api, extractErrorMessage, type ApiPage } from '@/shared/api/client'
import { useLanguage } from '@/shared/i18n/language'
import { ConfirmDialog } from '@/shared/components/confirm-dialog'
import { InformationDialog } from '@/shared/components/information-dialog'
import { TenantListCard } from './components/TenantListCard'
import { TenantFormSheet } from './components/TenantFormSheet'
import { TenantLineLinkDialog } from './components/TenantLineLinkDialog'
import type { ApiTenant, ApiTenantDeletionCheck } from './types'
import {
  TENANT_PAGE_SIZE_OPTIONS,
  type TenantLineFilter,
  type TenantSortDirection,
  type TenantSortKey,
  type TenantStatusFilter,
} from './utils'

const SEARCH_DEBOUNCE_MS = 300

export default function TenantPage() {
  const { t } = useLanguage()

  const [tenants, setTenants] = useState<ApiTenant[] | null>(null)
  const [total, setTotal] = useState(0)
  const [totalPages, setTotalPages] = useState(1)
  const [loadError, setLoadError] = useState<string | null>(null)

  const [query, setQuery] = useState('')
  const [debouncedQuery, setDebouncedQuery] = useState('')
  const [statusFilter, setStatusFilter] = useState<TenantStatusFilter>('all')
  const [lineFilter, setLineFilter] = useState<TenantLineFilter>('all')
  const [sortKey, setSortKey] = useState<TenantSortKey>('first_name')
  const [sortDirection, setSortDirection] = useState<TenantSortDirection>('asc')

  const [page, setPage] = useState(1)
  const [pageSize, setPageSizeState] = useState<number>(TENANT_PAGE_SIZE_OPTIONS[0])
  // Bumped by mutations so the list refetches; the server owns the ordering
  // and page boundaries now, so patching rows locally would misplace them.
  const [refreshToken, setRefreshToken] = useState(0)

  const [formOpen, setFormOpen] = useState(false)
  const [formTenantId, setFormTenantId] = useState<string | null>(null)
  const [formFirstName, setFormFirstName] = useState('')
  const [formLastName, setFormLastName] = useState('')
  const [formPhone, setFormPhone] = useState('')
  const [formLineId, setFormLineId] = useState('')
  const [formIdCard, setFormIdCard] = useState('')
  const [formEmail, setFormEmail] = useState('')
  const [formEmergencyContact, setFormEmergencyContact] = useState('')
  const [formNote, setFormNote] = useState('')
  const [formIsActive, setFormIsActive] = useState(true)
  const [formSaving, setFormSaving] = useState(false)
  const [formError, setFormError] = useState<string | null>(null)

  const [deletingTenantId, setDeletingTenantId] = useState<string | null>(null)
  const [checkingTenantId, setCheckingTenantId] = useState<string | null>(null)
  const [deleteError, setDeleteError] = useState<string | null>(null)
  const [confirmDeleteTenant, setConfirmDeleteTenant] = useState<ApiTenant | null>(null)
  const [blockedDeletionContractCount, setBlockedDeletionContractCount] = useState<number | null>(null)

  const [lineLinkInfo, setLineLinkInfo] = useState<{ tenant: ApiTenant; link: string } | null>(null)
  const [addFriendUrl, setAddFriendUrl] = useState<string | null>(null)
  const [lineStatus, setLineStatus] = useState<{ linked: boolean; is_friend: boolean } | null>(null)

  const [confirmUnlinkTenant, setConfirmUnlinkTenant] = useState<ApiTenant | null>(null)
  const [unlinkingTenantId, setUnlinkingTenantId] = useState<string | null>(null)
  const [unlinkError, setUnlinkError] = useState<string | null>(null)

  useEffect(() => {
    const timer = window.setTimeout(() => setDebouncedQuery(query), SEARCH_DEBOUNCE_MS)
    return () => window.clearTimeout(timer)
  }, [query])

  useEffect(() => {
    const controller = new AbortController()

    api
      .get<ApiPage<ApiTenant[]>>('/tenants', {
        signal: controller.signal,
        params: {
          page,
          per_page: pageSize,
          q: debouncedQuery.trim() || undefined,
          is_active: statusFilter === 'all' ? undefined : statusFilter === 'active',
          line_linked: lineFilter === 'all' ? undefined : lineFilter === 'linked',
          sort: sortKey,
          order: sortDirection,
        },
      })
      .then(({ data }) => {
        setTenants(data.data)
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
  }, [page, pageSize, debouncedQuery, statusFilter, lineFilter, sortKey, sortDirection, refreshToken])

  const isLoading = !loadError && tenants === null
  const hasFilters = query !== '' || statusFilter !== 'all' || lineFilter !== 'all'
  const rangeStart = total === 0 ? 0 : (page - 1) * pageSize + 1
  const rangeEnd = Math.min(page * pageSize, total)

  function refresh() {
    setRefreshToken((value) => value + 1)
  }

  function setPageSize(size: number) {
    setPageSizeState(size)
    setPage(1)
  }

  function handleSort(key: TenantSortKey) {
    if (key === sortKey) {
      setSortDirection(sortDirection === 'asc' ? 'desc' : 'asc')
    } else {
      setSortKey(key)
      setSortDirection('asc')
    }
    setPage(1)
  }

  function openCreateForm() {
    setFormTenantId(null)
    setFormFirstName('')
    setFormLastName('')
    setFormPhone('')
    setFormLineId('')
    setFormIdCard('')
    setFormEmail('')
    setFormEmergencyContact('')
    setFormNote('')
    setFormIsActive(true)
    setFormError(null)
    setFormOpen(true)
  }

  function openEditForm(tenant: ApiTenant) {
    setFormTenantId(tenant.id)
    setFormFirstName(tenant.first_name)
    setFormLastName(tenant.last_name)
    setFormPhone(tenant.phone)
    setFormLineId(tenant.line_id)
    setFormIdCard(tenant.id_card)
    setFormEmail(tenant.email)
    setFormEmergencyContact(tenant.emergency_contact)
    setFormNote(tenant.note)
    setFormIsActive(tenant.is_active)
    setFormError(null)
    setFormOpen(true)
  }

  async function handleFormSubmit(event: FormEvent<HTMLFormElement>) {
    event.preventDefault()

    const firstName = formFirstName.trim()
    const lastName = formLastName.trim()
    const isEdit = formTenantId !== null

    if (!firstName) {
      setFormError(t('tenantFirstNameRequired'))
      return
    }
    if (!lastName) {
      setFormError(t('tenantLastNameRequired'))
      return
    }

    setFormSaving(true)
    setFormError(null)

    const payload = {
      first_name: firstName,
      last_name: lastName,
      phone: formPhone.trim(),
      line_id: formLineId.trim(),
      id_card: formIdCard.trim(),
      email: formEmail.trim(),
      emergency_contact: formEmergencyContact.trim(),
      note: formNote.trim(),
      is_active: formIsActive,
    }

    try {
      if (!isEdit) {
        await api.post<ApiTenant>('/tenants', payload)
      } else {
        await api.put<ApiTenant>(`/tenants/${formTenantId}`, payload)
      }
      refresh()
      setFormOpen(false)
    } catch (err) {
      const fallback = isEdit ? t('tenantUpdateError') : t('tenantCreateError')
      setFormError(extractErrorMessage(err, fallback))
    } finally {
      setFormSaving(false)
    }
  }

  async function handleDeleteTenant() {
    if (!confirmDeleteTenant) return
    const tenant = confirmDeleteTenant

    setDeletingTenantId(tenant.id)
    setDeleteError(null)

    try {
      await api.delete(`/tenants/${tenant.id}`)
      // Deleting the last row of the final page would otherwise strand the
      // view on a page the server no longer has.
      setPage((value) => (tenants?.length === 1 ? Math.max(1, value - 1) : value))
      refresh()
      setConfirmDeleteTenant(null)
    } catch (err) {
      setDeleteError(extractErrorMessage(err, t('tenantDeleteError')))
    } finally {
      setDeletingTenantId(null)
    }
  }

  async function handleRequestDeleteTenant(tenant: ApiTenant) {
    setCheckingTenantId(tenant.id)
    setDeleteError(null)

    try {
      const { data } = await api.get<ApiTenantDeletionCheck>(`/tenants/${tenant.id}/deletion-check`)
      if (data.can_delete) {
        setConfirmDeleteTenant(tenant)
      } else {
        setBlockedDeletionContractCount(data.contract_count)
      }
    } catch (err) {
      setDeleteError(extractErrorMessage(err, t('tenantDeleteError')))
    } finally {
      setCheckingTenantId(null)
    }
  }

  async function handleCopyLineLink(tenant: ApiTenant) {
    const liffId = import.meta.env.VITE_LIFF_ID as string | undefined
    const link = liffId
      ? `https://liff.line.me/${liffId}?tenant_id=${tenant.id}`
      : `${window.location.origin}/liff/link-tenant?tenant_id=${tenant.id}`

    try {
      await navigator.clipboard.writeText(link)
    } catch {
      // Clipboard API may be unavailable (e.g. insecure context); the dialog
      // below still shows the link so it can be copied by hand.
    }

    setLineStatus(null)
    setLineLinkInfo({ tenant, link })

    // Linking alone doesn't make a tenant reachable by push — they must also
    // have the OA as a friend. Show both conditions plus the add-friend link,
    // so staff can see why an invoice would fail before they try to send it.
    try {
      const { data } = await api.get<{ linked: boolean; is_friend: boolean }>(
        `/tenants/${tenant.id}/line/status`,
      )
      setLineStatus(data)
    } catch {
      // Ignore — the dialog just omits the status line.
    }

    if (addFriendUrl === null) {
      try {
        const { data } = await api.get<{ add_friend_url: string }>('/public/line/oa')
        setAddFriendUrl(data.add_friend_url)
      } catch {
        // Ignore — the dialog just omits the add-friend link.
      }
    }
  }

  async function handleUnlinkLine() {
    if (!confirmUnlinkTenant) return
    const tenant = confirmUnlinkTenant

    setUnlinkingTenantId(tenant.id)
    setUnlinkError(null)

    try {
      await api.delete<ApiTenant>(`/tenants/${tenant.id}/line`)
      refresh()
      setConfirmUnlinkTenant(null)
    } catch (err) {
      setUnlinkError(extractErrorMessage(err, t('tenantUnlinkLineError')))
    } finally {
      setUnlinkingTenantId(null)
    }
  }

  return (
    <main className="content">
      <section className="welcome">
        <h1>{t('menuTenants')}</h1>
        <p>{t('menuTenantsDescription')}</p>
      </section>

      <TenantListCard
        isLoading={isLoading}
        loadError={loadError}
        deleteError={deleteError}
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
        lineFilter={lineFilter}
        onLineFilterChange={(value) => {
          setLineFilter(value)
          setPage(1)
        }}
        hasFilters={hasFilters}
        sortKey={sortKey}
        sortDirection={sortDirection}
        onSort={handleSort}
        tenants={tenants ?? []}
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
        deletingTenantId={checkingTenantId ?? deletingTenantId}
        onCreateTenant={openCreateForm}
        onEditTenant={openEditForm}
        onDeleteTenant={handleRequestDeleteTenant}
        onCopyLineLink={handleCopyLineLink}
        onUnlinkLine={setConfirmUnlinkTenant}
      />

      <TenantLineLinkDialog
        open={lineLinkInfo !== null}
        onOpenChange={(open) => !open && setLineLinkInfo(null)}
        tenant={lineLinkInfo?.tenant ?? null}
        link={lineLinkInfo?.link ?? ''}
        addFriendUrl={addFriendUrl}
        lineStatus={lineStatus}
      />

      <ConfirmDialog
        open={confirmDeleteTenant !== null}
        onOpenChange={(open) => !open && setConfirmDeleteTenant(null)}
        title={t('confirmDeleteTitle')}
        description={t('tenantDeleteConfirm')}
        confirmLabel={t('tenantDelete')}
        cancelLabel={t('cancel')}
        loading={deletingTenantId === confirmDeleteTenant?.id}
        error={deleteError}
        onConfirm={handleDeleteTenant}
      />

      <ConfirmDialog
        open={confirmUnlinkTenant !== null}
        onOpenChange={(open) => !open && setConfirmUnlinkTenant(null)}
        title={t('tenantUnlinkLine')}
        description={t('tenantUnlinkLineConfirm')}
        confirmLabel={t('tenantUnlinkLine')}
        cancelLabel={t('cancel')}
        loading={unlinkingTenantId === confirmUnlinkTenant?.id}
        error={unlinkError}
        onConfirm={handleUnlinkLine}
      />

      <InformationDialog
        open={blockedDeletionContractCount !== null}
        onOpenChange={(open) => !open && setBlockedDeletionContractCount(null)}
        title={t('tenantDeleteBlockedTitle')}
        description={t('tenantDeleteBlockedDescription').replace('{count}', String(blockedDeletionContractCount ?? 0))}
        actionLabel={t('acknowledge')}
      />

      <TenantFormSheet
        open={formOpen}
        onOpenChange={setFormOpen}
        isEdit={formTenantId !== null}
        firstName={formFirstName}
        onFirstNameChange={setFormFirstName}
        lastName={formLastName}
        onLastNameChange={setFormLastName}
        phone={formPhone}
        onPhoneChange={setFormPhone}
        lineId={formLineId}
        onLineIdChange={setFormLineId}
        idCard={formIdCard}
        onIdCardChange={setFormIdCard}
        email={formEmail}
        onEmailChange={setFormEmail}
        emergencyContact={formEmergencyContact}
        onEmergencyContactChange={setFormEmergencyContact}
        note={formNote}
        onNoteChange={setFormNote}
        isActive={formIsActive}
        onIsActiveChange={setFormIsActive}
        saving={formSaving}
        error={formError}
        onSubmit={handleFormSubmit}
      />
    </main>
  )
}
