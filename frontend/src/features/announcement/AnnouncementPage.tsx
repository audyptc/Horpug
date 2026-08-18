import { useEffect, useMemo, useState, type FormEvent } from 'react'
import { api, extractErrorMessage, type ApiPage } from '@/shared/api/client'
import { usePagination } from '@/shared/hooks/use-pagination'
import { useLanguage } from '@/shared/i18n/language'
import { ConfirmDialog } from '@/shared/components/confirm-dialog'
import type { ApiDormitory } from '@/features/dormitory/types'
import { AnnouncementListCard } from './components/AnnouncementListCard'
import { AnnouncementFormSheet } from './components/AnnouncementFormSheet'
import type { ApiAnnouncement } from './types'
import { ANNOUNCEMENT_PAGE_SIZE_OPTIONS, toApiDate, toDateInputValue } from './utils'

export default function AnnouncementPage() {
  const { t } = useLanguage()

  const [announcements, setAnnouncements] = useState<ApiAnnouncement[] | null>(null)
  const [dormitories, setDormitories] = useState<ApiDormitory[]>([])
  const [loadError, setLoadError] = useState<string | null>(null)

  const [query, setQuery] = useState('')

  const [formOpen, setFormOpen] = useState(false)
  const [formAnnouncementId, setFormAnnouncementId] = useState<string | null>(null)
  const [formDormitoryId, setFormDormitoryId] = useState('')
  const [formTitle, setFormTitle] = useState('')
  const [formContent, setFormContent] = useState('')
  const [formIsPublished, setFormIsPublished] = useState(true)
  const [formPublishedDate, setFormPublishedDate] = useState('')
  const [formSaving, setFormSaving] = useState(false)
  const [formError, setFormError] = useState<string | null>(null)

  const [deletingAnnouncementId, setDeletingAnnouncementId] = useState<string | null>(null)
  const [deleteError, setDeleteError] = useState<string | null>(null)
  const [confirmDeleteAnnouncement, setConfirmDeleteAnnouncement] = useState<ApiAnnouncement | null>(null)

  useEffect(() => {
    let cancelled = false

    Promise.all([
      api.get<ApiPage<ApiAnnouncement[]>>('/announcements', { params: { per_page: 100 } }),
      api.get<ApiDormitory[]>('/dormitories/active', { params: { limit: 100 } }),
    ])
      .then(([announcementsRes, dormitoriesRes]) => {
        if (cancelled) return
        setAnnouncements(announcementsRes.data.data)
        setDormitories(dormitoriesRes.data)
      })
      .catch((err) => {
        if (!cancelled) setLoadError(extractErrorMessage(err, t('resourceLoadError')))
      })

    return () => {
      cancelled = true
    }
    // eslint-disable-next-line react-hooks/exhaustive-deps
  }, [])

  const filteredAnnouncements = useMemo(() => {
    const q = query.trim().toLocaleLowerCase()
    if (!q) return announcements ?? []

    return (announcements ?? []).filter((announcement) => {
      return (
        (announcement.dormitory_name ?? '').toLocaleLowerCase().includes(q) ||
        announcement.title.toLocaleLowerCase().includes(q) ||
        announcement.content.toLocaleLowerCase().includes(q)
      )
    })
  }, [query, announcements])

  const {
    page: currentPage,
    pageSize,
    setPageSize,
    totalPages,
    rangeStart,
    rangeEnd,
    paginatedItems: paginatedAnnouncements,
    resetPage,
    firstPage,
    prevPage,
    nextPage,
    lastPage,
  } = usePagination(filteredAnnouncements, ANNOUNCEMENT_PAGE_SIZE_OPTIONS[0])

  const isLoading = !loadError && announcements === null

  function openCreateForm() {
    setFormAnnouncementId(null)
    setFormDormitoryId('')
    setFormTitle('')
    setFormContent('')
    setFormIsPublished(true)
    setFormPublishedDate(toDateInputValue(new Date().toISOString()))
    setFormError(null)
    setFormOpen(true)
  }

  function openEditForm(announcement: ApiAnnouncement) {
    setFormAnnouncementId(announcement.id)
    setFormDormitoryId(announcement.dormitory_id)
    setFormTitle(announcement.title)
    setFormContent(announcement.content)
    setFormIsPublished(announcement.is_published)
    setFormPublishedDate(toDateInputValue(announcement.published_date))
    setFormError(null)
    setFormOpen(true)
  }

  async function handleFormSubmit(event: FormEvent<HTMLFormElement>) {
    event.preventDefault()

    const isEdit = formAnnouncementId !== null

    if (!isEdit && !formDormitoryId) {
      setFormError(t('announcementDormitoryRequired'))
      return
    }
    if (!formTitle.trim()) {
      setFormError(t('announcementTitleRequired'))
      return
    }

    setFormSaving(true)
    setFormError(null)

    try {
      if (!isEdit) {
        const payload = {
          dormitory_id: formDormitoryId,
          title: formTitle.trim(),
          content: formContent.trim(),
          is_published: formIsPublished,
          published_date: formPublishedDate ? toApiDate(formPublishedDate) : undefined,
        }
        const { data } = await api.post<ApiAnnouncement>('/announcements', payload)
        setAnnouncements((prev) => [data, ...(prev ?? [])])
      } else {
        const payload = {
          title: formTitle.trim(),
          content: formContent.trim(),
          is_published: formIsPublished,
          published_date: formPublishedDate ? toApiDate(formPublishedDate) : undefined,
        }
        const { data } = await api.put<ApiAnnouncement>(`/announcements/${formAnnouncementId}`, payload)
        setAnnouncements((prev) => prev?.map((item) => (item.id === data.id ? data : item)) ?? prev)
      }
      setFormOpen(false)
    } catch (err) {
      const fallback = isEdit ? t('announcementUpdateError') : t('announcementCreateError')
      setFormError(extractErrorMessage(err, fallback))
    } finally {
      setFormSaving(false)
    }
  }

  async function handleDeleteAnnouncement() {
    if (!confirmDeleteAnnouncement) return
    const announcement = confirmDeleteAnnouncement

    setDeletingAnnouncementId(announcement.id)
    setDeleteError(null)

    try {
      await api.delete(`/announcements/${announcement.id}`)
      setAnnouncements((prev) => prev?.filter((item) => item.id !== announcement.id) ?? prev)
      setConfirmDeleteAnnouncement(null)
    } catch (err) {
      setDeleteError(extractErrorMessage(err, t('announcementDeleteError')))
    } finally {
      setDeletingAnnouncementId(null)
    }
  }

  return (
    <main className="content">
      <section className="welcome">
        <h1>{t('menuAnnouncements')}</h1>
        <p>{t('menuAnnouncementsDescription')}</p>
      </section>

      <AnnouncementListCard
        isLoading={isLoading}
        loadError={loadError}
        deleteError={deleteError}
        announcements={announcements}
        query={query}
        onQueryChange={(value) => {
          setQuery(value)
          resetPage()
        }}
        filteredAnnouncements={filteredAnnouncements}
        paginatedAnnouncements={paginatedAnnouncements}
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
        deletingAnnouncementId={deletingAnnouncementId}
        onCreateAnnouncement={openCreateForm}
        onEditAnnouncement={openEditForm}
        onDeleteAnnouncement={setConfirmDeleteAnnouncement}
      />

      <ConfirmDialog
        open={confirmDeleteAnnouncement !== null}
        onOpenChange={(open) => !open && setConfirmDeleteAnnouncement(null)}
        title={t('confirmDeleteTitle')}
        description={t('announcementDeleteConfirm')}
        confirmLabel={t('announcementDelete')}
        cancelLabel={t('cancel')}
        loading={deletingAnnouncementId === confirmDeleteAnnouncement?.id}
        error={deleteError}
        onConfirm={handleDeleteAnnouncement}
      />

      <AnnouncementFormSheet
        open={formOpen}
        onOpenChange={setFormOpen}
        isEdit={formAnnouncementId !== null}
        dormitoryId={formDormitoryId}
        onDormitoryIdChange={setFormDormitoryId}
        dormitories={dormitories}
        title={formTitle}
        onTitleChange={setFormTitle}
        content={formContent}
        onContentChange={setFormContent}
        isPublished={formIsPublished}
        onIsPublishedChange={setFormIsPublished}
        publishedDate={formPublishedDate}
        onPublishedDateChange={setFormPublishedDate}
        saving={formSaving}
        error={formError}
        onSubmit={handleFormSubmit}
      />
    </main>
  )
}
