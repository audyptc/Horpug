import { useEffect, useState } from 'react'
import axios from 'axios'
import { api, extractErrorMessage, type ApiPage } from '@/shared/api/client'
import { useLanguage } from '@/shared/i18n/language'
import { ActivityLogListCard } from './components/ActivityLogListCard'
import type { ApiActivityLog } from './types'
import {
  ACTIVITY_LOG_PAGE_SIZE_OPTIONS,
  type ActivityLogColumnFilters,
  type ActivityLogSortDirection,
  type ActivityLogSortKey,
  type ActivityLogTextFilterKey,
} from './utils'

const SEARCH_DEBOUNCE_MS = 300

export default function ActivityLogPage() {
  const { t } = useLanguage()

  const [logs, setLogs] = useState<ApiActivityLog[] | null>(null)
  const [total, setTotal] = useState(0)
  const [totalPages, setTotalPages] = useState(1)
  const [loadError, setLoadError] = useState<string | null>(null)

  const [query, setQuery] = useState('')
  const [debouncedQuery, setDebouncedQuery] = useState('')
  // Applied on submit from each column's menu, so no debounce is needed here.
  const [columnFilters, setColumnFilters] = useState<ActivityLogColumnFilters>({})
  const [dateFrom, setDateFrom] = useState('')
  const [dateTo, setDateTo] = useState('')
  // Activity logs read most naturally newest-first, unlike the alphabetical
  // default used by the other converted lists.
  const [sortKey, setSortKey] = useState<ActivityLogSortKey>('created_at')
  const [sortDirection, setSortDirection] = useState<ActivityLogSortDirection>('desc')

  const [page, setPage] = useState(1)
  const [pageSize, setPageSizeState] = useState<number>(ACTIVITY_LOG_PAGE_SIZE_OPTIONS[0])

  useEffect(() => {
    const timer = window.setTimeout(() => setDebouncedQuery(query), SEARCH_DEBOUNCE_MS)
    return () => window.clearTimeout(timer)
  }, [query])

  useEffect(() => {
    const controller = new AbortController()

    const columnParams: Record<string, string> = {}
    for (const [key, value] of Object.entries(columnFilters)) {
      if (value) columnParams[`f[${key}]`] = value
    }

    api
      .get<ApiPage<ApiActivityLog[]>>('/activity-logs', {
        signal: controller.signal,
        params: {
          page,
          per_page: pageSize,
          q: debouncedQuery.trim() || undefined,
          date_from: dateFrom || undefined,
          date_to: dateTo || undefined,
          sort: sortKey,
          order: sortDirection,
          ...columnParams,
        },
      })
      .then(({ data }) => {
        setLogs(data.data)
        setTotal(data.meta.total)
        setTotalPages(Math.max(1, data.meta.total_pages))
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
  }, [page, pageSize, debouncedQuery, columnFilters, dateFrom, dateTo, sortKey, sortDirection])

  const isLoading = !loadError && logs === null
  const hasFilters =
    query !== '' || dateFrom !== '' || dateTo !== '' || Object.values(columnFilters).some(Boolean)
  const rangeStart = total === 0 ? 0 : (page - 1) * pageSize + 1
  const rangeEnd = Math.min(page * pageSize, total)

  function setPageSize(size: number) {
    setPageSizeState(size)
    setPage(1)
  }

  function handleSort(key: ActivityLogSortKey) {
    if (key === sortKey) {
      setSortDirection(sortDirection === 'asc' ? 'desc' : 'asc')
    } else {
      setSortKey(key)
      setSortDirection('asc')
    }
    setPage(1)
  }

  return (
    <main className="content">
      <section className="welcome">
        <h1>{t('menuActivityLogs')}</h1>
        <p>{t('menuActivityLogsDescription')}</p>
      </section>

      <ActivityLogListCard
        isLoading={isLoading}
        loadError={loadError}
        logs={logs ?? []}
        query={query}
        onQueryChange={(value) => {
          setQuery(value)
          setPage(1)
        }}
        columnFilters={columnFilters}
        onColumnFilterChange={(key: ActivityLogTextFilterKey, value: string) => {
          setColumnFilters((prev) => ({ ...prev, [key]: value }))
          setPage(1)
        }}
        dateFrom={dateFrom}
        onDateFromChange={(value) => {
          setDateFrom(value)
          setPage(1)
        }}
        dateTo={dateTo}
        onDateToChange={(value) => {
          setDateTo(value)
          setPage(1)
        }}
        hasFilters={hasFilters}
        sortKey={sortKey}
        sortDirection={sortDirection}
        onSort={handleSort}
        currentPage={page}
        totalPages={totalPages}
        rangeStart={rangeStart}
        rangeEnd={rangeEnd}
        totalItems={total}
        pageSize={pageSize}
        onPageSizeChange={setPageSize}
        onFirstPage={() => setPage(1)}
        onPrevPage={() => setPage((value) => Math.max(1, value - 1))}
        onNextPage={() => setPage((value) => Math.min(totalPages, value + 1))}
        onLastPage={() => setPage(totalPages)}
      />
    </main>
  )
}
