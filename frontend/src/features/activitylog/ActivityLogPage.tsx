import { useEffect, useState } from 'react'
import { api, extractErrorMessage, type ApiPage } from '@/shared/api/client'
import { useLanguage } from '@/shared/i18n/language'
import { ActivityLogListCard } from './components/ActivityLogListCard'
import type { ApiActivityLog } from './types'
import { ACTIVITY_LOG_PAGE_SIZE_OPTIONS } from './utils'

export default function ActivityLogPage() {
  const { t } = useLanguage()

  const [logs, setLogs] = useState<ApiActivityLog[] | null>(null)
  const [total, setTotal] = useState(0)
  const [totalPages, setTotalPages] = useState(1)
  const [loadError, setLoadError] = useState<string | null>(null)

  const [entityTypeInput, setEntityTypeInput] = useState('')
  const [entityType, setEntityType] = useState('')
  const [page, setPage] = useState(1)
  const [pageSize, setPageSize] = useState<number>(ACTIVITY_LOG_PAGE_SIZE_OPTIONS[0])

  // Debounce the free-text filter so it doesn't fire a request per keystroke.
  useEffect(() => {
    const handle = setTimeout(() => {
      setEntityType(entityTypeInput.trim())
      setPage(1)
    }, 400)

    return () => clearTimeout(handle)
  }, [entityTypeInput])

  useEffect(() => {
    let cancelled = false

    api
      .get<ApiPage<ApiActivityLog[]>>('/activity-logs', {
        params: {
          page,
          per_page: pageSize,
          ...(entityType ? { entity_type: entityType } : {}),
        },
      })
      .then(({ data }) => {
        if (cancelled) return
        setLogs(data.data)
        setTotal(data.meta.total)
        setTotalPages(Math.max(1, data.meta.total_pages))
        setLoadError(null)
      })
      .catch((err) => {
        if (!cancelled) setLoadError(extractErrorMessage(err, t('resourceLoadError')))
      })

    return () => {
      cancelled = true
    }
    // eslint-disable-next-line react-hooks/exhaustive-deps
  }, [page, pageSize, entityType])

  const currentPage = Math.min(page, totalPages)
  const rangeStart = total === 0 ? 0 : (currentPage - 1) * pageSize + 1
  const rangeEnd = Math.min(currentPage * pageSize, total)
  const isLoading = !loadError && logs === null

  return (
    <main className="content">
      <section className="welcome">
        <h1>{t('menuActivityLogs')}</h1>
        <p>{t('menuActivityLogsDescription')}</p>
      </section>

      <ActivityLogListCard
        isLoading={isLoading}
        loadError={loadError}
        logs={logs}
        entityTypeFilter={entityTypeInput}
        onEntityTypeFilterChange={setEntityTypeInput}
        currentPage={currentPage}
        totalPages={totalPages}
        rangeStart={rangeStart}
        rangeEnd={rangeEnd}
        totalItems={total}
        pageSize={pageSize}
        onPageSizeChange={(size) => {
          setPageSize(size)
          setPage(1)
        }}
        onPrevPage={() => setPage((p) => Math.max(1, p - 1))}
        onNextPage={() => setPage((p) => Math.min(totalPages, p + 1))}
      />
    </main>
  )
}
