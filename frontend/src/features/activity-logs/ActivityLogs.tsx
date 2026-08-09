import { useState, useEffect, useCallback } from 'react'
import { useTranslation } from 'react-i18next'
import {
  RefreshCw,
  ChevronLeft,
  ChevronRight,
  ChevronsLeft,
  ChevronsRight,
  SearchX,
  History,
} from 'lucide-react'
import { Button } from '@/components/ui/button'
import { Card, CardContent, CardHeader, CardTitle, CardDescription } from '@/components/ui/card'
import {
  Select,
  SelectContent,
  SelectItem,
  SelectTrigger,
  SelectValue,
} from '@/components/ui/select'
import { DatePicker } from '@/components/ui/date-picker'
import { activityLogService } from './activityLogService'
import type { ApiActivityLog, ActivityAction } from '@/types'
import { cn } from '@/lib/utils'

const PER_PAGE_OPTIONS = [20, 50, 100] as const

const ENTITY_TYPES = [
  'payment',
  'contract',
  'role',
  'role_permissions',
  'user',
  'user_role',
] as const

const ACTION_STYLES: Record<ActivityAction, string> = {
  create: 'bg-emerald-100 text-emerald-700 dark:bg-emerald-900/30 dark:text-emerald-400',
  update: 'bg-blue-100 text-blue-700 dark:bg-blue-900/30 dark:text-blue-400',
  delete: 'bg-red-100 text-red-700 dark:bg-red-900/30 dark:text-red-400',
}

function formatDateTime(iso: string) {
  return new Date(iso).toLocaleString('th-TH', {
    day: '2-digit',
    month: 'short',
    year: 'numeric',
    hour: '2-digit',
    minute: '2-digit',
  })
}

export function ActivityLogs() {
  const { t } = useTranslation()

  const [items, setItems] = useState<ApiActivityLog[]>([])
  const [loading, setLoading] = useState(true)
  const [error, setError] = useState('')
  const [page, setPage] = useState(1)
  const [perPage, setPerPage] = useState<number>(20)
  const [total, setTotal] = useState(0)
  const [totalPages, setTotalPages] = useState(1)

  const [filterEntity, setFilterEntity] = useState('all')
  const [filterFrom, setFilterFrom] = useState('')
  const [filterTo, setFilterTo] = useState('')

  const fetchItems = useCallback(
    async (p: number, pp: number, entity: string, from: string, to: string) => {
      setLoading(true)
      setError('')
      try {
        const resp = await activityLogService.list({
          page: p,
          perPage: pp,
          entityType: entity === 'all' ? undefined : entity,
          from: from ? `${from}T00:00:00Z` : undefined,
          to: to ? `${to}T23:59:59Z` : undefined,
        })
        setItems(resp.data)
        setTotal(resp.meta.total)
        setTotalPages(resp.meta.total_pages)
      } catch {
        setError(t('activityLogs.loadError'))
      } finally {
        setLoading(false)
      }
    },
    [t]
  )

  useEffect(() => {
    fetchItems(page, perPage, filterEntity, filterFrom, filterTo)
  }, [fetchItems, page, perPage, filterEntity, filterFrom, filterTo])

  function handlePerPageChange(value: string) {
    setPerPage(Number(value))
    setPage(1)
  }

  function handleEntityChange(value: string) {
    setFilterEntity(value)
    setPage(1)
  }

  function handleFromChange(value: string) {
    setFilterFrom(value)
    setPage(1)
  }

  function handleToChange(value: string) {
    setFilterTo(value)
    setPage(1)
  }

  function clearFilters() {
    setFilterEntity('all')
    setFilterFrom('')
    setFilterTo('')
    setPage(1)
  }

  const hasFilters = filterEntity !== 'all' || filterFrom || filterTo

  return (
    <div className="space-y-6">
      <div className="flex flex-wrap items-start justify-between gap-3">
        <div className="min-w-0">
          <h1 className="text-2xl font-bold tracking-tight">{t('activityLogs.title')}</h1>
          <p className="text-muted-foreground text-sm mt-1">
            {t('activityLogs.subtitle', { total })}
          </p>
        </div>
        <Button
          variant="outline"
          size="icon"
          onClick={() => fetchItems(page, perPage, filterEntity, filterFrom, filterTo)}
          disabled={loading}
        >
          <RefreshCw className={cn('w-4 h-4', loading && 'animate-spin')} />
        </Button>
      </div>

      {error && (
        <div className="rounded-md border border-destructive/30 bg-destructive/10 px-4 py-3 text-sm text-destructive">
          {error}
        </div>
      )}

      <Card>
        <CardHeader>
          <div className="flex flex-col sm:flex-row sm:items-center justify-between gap-3">
            <div>
              <CardTitle>{t('activityLogs.logList')}</CardTitle>
              <CardDescription>{t('activityLogs.logListDesc')}</CardDescription>
            </div>
            <div className="flex flex-wrap items-center gap-2">
              {/* Entity type filter */}
              <Select value={filterEntity} onValueChange={handleEntityChange}>
                <SelectTrigger className="h-9 w-44">
                  <SelectValue />
                </SelectTrigger>
                <SelectContent>
                  <SelectItem value="all">{t('activityLogs.allEntities')}</SelectItem>
                  {ENTITY_TYPES.map((et) => (
                    <SelectItem key={et} value={et}>
                      {t(`activityLogs.entityTypes.${et}`)}
                    </SelectItem>
                  ))}
                </SelectContent>
              </Select>

              {/* Date range */}
              <DatePicker
                value={filterFrom}
                onChange={handleFromChange}
                placeholder={t('activityLogs.from')}
                className="h-9 w-36"
              />
              <span className="text-muted-foreground text-sm">–</span>
              <DatePicker
                value={filterTo}
                onChange={handleToChange}
                placeholder={t('activityLogs.to')}
                className="h-9 w-36"
              />

              {hasFilters && (
                <Button variant="ghost" size="sm" onClick={clearFilters} className="h-9 text-xs">
                  {t('activityLogs.clearFilters')}
                </Button>
              )}
            </div>
          </div>
        </CardHeader>

        <CardContent className="p-0">
          {loading ? (
            <div className="flex items-center justify-center py-16 text-sm text-muted-foreground">
              {t('common.loading')}
            </div>
          ) : (
            <>
              {/* Desktop table */}
              <div className="hidden md:block overflow-x-auto">
                <table className="w-full text-sm">
                  <thead>
                    <tr className="border-b bg-muted/40">
                      <th className="text-left px-6 py-3 font-medium text-muted-foreground">
                        {t('activityLogs.colDateTime')}
                      </th>
                      <th className="text-left px-4 py-3 font-medium text-muted-foreground">
                        {t('activityLogs.colActor')}
                      </th>
                      <th className="text-left px-4 py-3 font-medium text-muted-foreground">
                        {t('activityLogs.colAction')}
                      </th>
                      <th className="text-left px-4 py-3 font-medium text-muted-foreground">
                        {t('activityLogs.colEntityType')}
                      </th>
                      <th className="text-left px-4 py-3 font-medium text-muted-foreground">
                        {t('activityLogs.colEntityId')}
                      </th>
                    </tr>
                  </thead>
                  <tbody>
                    {items.map((item, i) => (
                      <tr
                        key={item.id}
                        className={cn(
                          'border-b transition-colors hover:bg-muted/30',
                          i === items.length - 1 && 'border-0'
                        )}
                      >
                        <td className="px-6 py-3 text-muted-foreground whitespace-nowrap">
                          {formatDateTime(item.created_at)}
                        </td>
                        <td className="px-4 py-3 font-medium">
                          {item.actor_name || (
                            <span className="text-muted-foreground italic">
                              {t('activityLogs.unknownActor')}
                            </span>
                          )}
                        </td>
                        <td className="px-4 py-3">
                          <span
                            className={cn(
                              'inline-flex items-center rounded-full px-2.5 py-0.5 text-xs font-medium',
                              ACTION_STYLES[item.action]
                            )}
                          >
                            {t(`activityLogs.actions.${item.action}`)}
                          </span>
                        </td>
                        <td className="px-4 py-3">
                          {t(`activityLogs.entityTypes.${item.entity_type}`, {
                            defaultValue: item.entity_type,
                          })}
                        </td>
                        <td className="px-4 py-3 font-mono text-xs text-muted-foreground">
                          {item.entity_id.slice(0, 8)}…
                        </td>
                      </tr>
                    ))}
                  </tbody>
                </table>

                {items.length === 0 && (
                  <div className="flex flex-col items-center justify-center py-16 gap-3 text-muted-foreground">
                    <div className="p-4 rounded-full bg-muted">
                      <SearchX className="w-6 h-6" />
                    </div>
                    <p className="text-sm font-medium">{t('activityLogs.noLogs')}</p>
                    {hasFilters && (
                      <button
                        type="button"
                        onClick={clearFilters}
                        className="text-xs text-primary hover:underline"
                      >
                        {t('activityLogs.clearFilters')}
                      </button>
                    )}
                  </div>
                )}
              </div>

              {/* Mobile cards */}
              <div className="md:hidden divide-y">
                {items.map((item) => (
                  <div key={item.id} className="p-4 space-y-1.5">
                    <div className="flex items-center justify-between gap-2">
                      <span className="font-medium text-sm">
                        {item.actor_name || (
                          <span className="text-muted-foreground italic">
                            {t('activityLogs.unknownActor')}
                          </span>
                        )}
                      </span>
                      <span
                        className={cn(
                          'inline-flex items-center rounded-full px-2 py-0.5 text-xs font-medium',
                          ACTION_STYLES[item.action]
                        )}
                      >
                        {t(`activityLogs.actions.${item.action}`)}
                      </span>
                    </div>
                    <p className="text-xs text-muted-foreground">
                      {t(`activityLogs.entityTypes.${item.entity_type}`, {
                        defaultValue: item.entity_type,
                      })}{' '}
                      · <span className="font-mono">{item.entity_id.slice(0, 8)}…</span>
                    </p>
                    <p className="text-xs text-muted-foreground">{formatDateTime(item.created_at)}</p>
                  </div>
                ))}

                {items.length === 0 && (
                  <div className="flex flex-col items-center justify-center py-12 gap-2 text-muted-foreground">
                    <History className="w-6 h-6" />
                    <p className="text-sm">{t('activityLogs.noLogs')}</p>
                  </div>
                )}
              </div>

              {/* Pagination */}
              {!loading && total > 0 && (
                <div className="flex flex-col sm:flex-row items-center justify-between gap-3 px-4 py-3 border-t text-sm text-muted-foreground">
                  <span className="shrink-0">
                    {t('common.pagination.showing', {
                      from: (page - 1) * perPage + 1,
                      to: Math.min(page * perPage, total),
                      total,
                    })}
                  </span>
                  <div className="flex items-center gap-4">
                    <div className="flex items-center gap-2">
                      <span className="hidden sm:inline shrink-0">{t('common.pagination.perPage')}</span>
                      <Select value={String(perPage)} onValueChange={handlePerPageChange}>
                        <SelectTrigger className="h-8 w-16">
                          <SelectValue />
                        </SelectTrigger>
                        <SelectContent>
                          {PER_PAGE_OPTIONS.map((n) => (
                            <SelectItem key={n} value={String(n)}>
                              {n}
                            </SelectItem>
                          ))}
                        </SelectContent>
                      </Select>
                    </div>
                    <span className="shrink-0">
                      {t('common.pagination.page')} {page} {t('common.pagination.of')} {totalPages}
                    </span>
                    <div className="flex items-center gap-1">
                      <Button
                        variant="outline"
                        size="icon"
                        className="h-8 w-8"
                        onClick={() => setPage(1)}
                        disabled={page === 1}
                      >
                        <ChevronsLeft className="w-4 h-4" />
                      </Button>
                      <Button
                        variant="outline"
                        size="icon"
                        className="h-8 w-8"
                        onClick={() => setPage((p) => p - 1)}
                        disabled={page === 1}
                      >
                        <ChevronLeft className="w-4 h-4" />
                      </Button>
                      <Button
                        variant="outline"
                        size="icon"
                        className="h-8 w-8"
                        onClick={() => setPage((p) => p + 1)}
                        disabled={page >= totalPages}
                      >
                        <ChevronRight className="w-4 h-4" />
                      </Button>
                      <Button
                        variant="outline"
                        size="icon"
                        className="h-8 w-8"
                        onClick={() => setPage(totalPages)}
                        disabled={page >= totalPages}
                      >
                        <ChevronsRight className="w-4 h-4" />
                      </Button>
                    </div>
                  </div>
                </div>
              )}
            </>
          )}
        </CardContent>
      </Card>
    </div>
  )
}
