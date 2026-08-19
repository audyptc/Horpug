import { useEffect, useMemo, useState, type FormEvent } from 'react'
import { api, extractErrorMessage, type ApiPage } from '@/shared/api/client'
import { usePagination } from '@/shared/hooks/use-pagination'
import { useLanguage } from '@/shared/i18n/language'
import { ConfirmDialog } from '@/shared/components/confirm-dialog'
import { InformationDialog } from '@/shared/components/information-dialog'
import { TenantListCard } from './components/TenantListCard'
import { TenantFormSheet } from './components/TenantFormSheet'
import type { ApiTenant, ApiTenantDeletionCheck } from './types'
import { TENANT_PAGE_SIZE_OPTIONS } from './utils'

export default function TenantPage() {
  const { t } = useLanguage()

  const [tenants, setTenants] = useState<ApiTenant[] | null>(null)
  const [loadError, setLoadError] = useState<string | null>(null)

  const [query, setQuery] = useState('')

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

  useEffect(() => {
    let cancelled = false

    api
      .get<ApiPage<ApiTenant[]>>('/tenants', { params: { per_page: 100 } })
      .then(({ data }) => {
        if (!cancelled) setTenants(data.data)
      })
      .catch((err) => {
        if (!cancelled) setLoadError(extractErrorMessage(err, t('resourceLoadError')))
      })

    return () => {
      cancelled = true
    }
    // eslint-disable-next-line react-hooks/exhaustive-deps
  }, [])

  const filteredTenants = useMemo(() => {
    const q = query.trim().toLocaleLowerCase()
    if (!q) return tenants ?? []

    return (tenants ?? []).filter((tenant) => {
      return (
        tenant.first_name.toLocaleLowerCase().includes(q) ||
        tenant.last_name.toLocaleLowerCase().includes(q) ||
        tenant.phone.toLocaleLowerCase().includes(q) ||
        tenant.line_id.toLocaleLowerCase().includes(q) ||
        tenant.id_card.toLocaleLowerCase().includes(q) ||
        tenant.email.toLocaleLowerCase().includes(q)
      )
    })
  }, [query, tenants])

  const {
    page: currentPage,
    pageSize,
    setPageSize,
    totalPages,
    rangeStart,
    rangeEnd,
    paginatedItems: paginatedTenants,
    resetPage,
    firstPage,
    prevPage,
    nextPage,
    lastPage,
  } = usePagination(filteredTenants, TENANT_PAGE_SIZE_OPTIONS[0])

  const isLoading = !loadError && tenants === null

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
        const { data } = await api.post<ApiTenant>('/tenants', payload)
        setTenants((prev) => [...(prev ?? []), data])
      } else {
        const { data } = await api.put<ApiTenant>(`/tenants/${formTenantId}`, payload)
        setTenants((prev) => prev?.map((item) => (item.id === data.id ? data : item)) ?? prev)
      }
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
      setTenants((prev) => prev?.filter((item) => item.id !== tenant.id) ?? prev)
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

    setLineLinkInfo({ tenant, link })
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
        tenants={tenants}
        query={query}
        onQueryChange={(value) => {
          setQuery(value)
          resetPage()
        }}
        filteredTenants={filteredTenants}
        paginatedTenants={paginatedTenants}
        currentPage={currentPage}
        totalPages={totalPages}
        rangeStart={rangeStart}
        rangeEnd={rangeEnd}
        pageSize={pageSize}
        onPageSizeChange={setPageSize}
        onFirstPage={firstPage}
        onPrevPage={prevPage}
        onNextPage={nextPage}
        onLastPage={lastPage}
        deletingTenantId={checkingTenantId ?? deletingTenantId}
        onCreateTenant={openCreateForm}
        onEditTenant={openEditForm}
        onDeleteTenant={handleRequestDeleteTenant}
        onCopyLineLink={handleCopyLineLink}
      />

      <InformationDialog
        open={lineLinkInfo !== null}
        onOpenChange={(open) => !open && setLineLinkInfo(null)}
        title={t('tenantLineLinkDialogTitle')}
        description={
          lineLinkInfo
            ? `${t('tenantLineLinkDialogDescription').replace('{name}', `${lineLinkInfo.tenant.first_name} ${lineLinkInfo.tenant.last_name}`)}\n\n${lineLinkInfo.link}`
            : ''
        }
        actionLabel={t('acknowledge')}
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
