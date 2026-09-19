import { useEffect, useState } from 'react'
import { BedDouble, FileText, ReceiptText, Wrench, type LucideIcon } from 'lucide-react'
import { api, extractErrorMessage, type ApiPage } from '@/shared/api/client'
import { useLanguage } from '@/shared/i18n/language'
import { Badge } from '@/shared/components/ui/badge'
import {
  Card,
  CardContent,
  CardDescription,
  CardHeader,
  CardTitle,
} from '@/shared/components/ui/card'
import {
  Table,
  TableBody,
  TableCell,
  TableHead,
  TableHeader,
  TableRow,
} from '@/shared/components/ui/table'
import type { ApiDashboardSummary } from './types'
import type { ApiActivityLog } from '@/features/activitylog/types'
import { DormitorySearchSelect } from '@/features/dormitory/components/DormitorySearchSelect'
import type { ApiDormitory } from '@/features/dormitory/types'
import { activityLogActionVariant } from '@/features/activitylog/utils'

type SelectedDormitory = { id: string; name: string }

// A per-browser convenience only: it is re-checked against the server on every
// visit, so a dormitory the user has since lost access to is dropped rather
// than trusted.
const DORMITORY_STORAGE_KEY = 'dashboard.dormitory'

function readStoredDormitory(): SelectedDormitory | null {
  try {
    const parsed = JSON.parse(localStorage.getItem(DORMITORY_STORAGE_KEY) ?? 'null')
    return typeof parsed?.id === 'string' && typeof parsed?.name === 'string' ? parsed : null
  } catch {
    return null
  }
}

function storeDormitory(dormitory: SelectedDormitory | null) {
  try {
    if (dormitory) localStorage.setItem(DORMITORY_STORAGE_KEY, JSON.stringify(dormitory))
    else localStorage.removeItem(DORMITORY_STORAGE_KEY)
  } catch {
    // Storage can be unavailable (private mode, blocked site data); the
    // selection just won't be remembered.
  }
}

type Metric = {
  key: string
  title: string
  value: string
  detail: string
  icon: LucideIcon
  iconClass: string
}

export function DashboardPage() {
  const { t, language } = useLanguage()

  // The dormitory the figures are narrowed to; null means every dormitory the
  // user can access. Not applied until `ready`, once the remembered choice has
  // been checked.
  const [dormitory, setDormitory] = useState<SelectedDormitory | null>(null)
  const [ready, setReady] = useState(false)
  // Whether there is anything to choose between: with one dormitory (or none)
  // the selector would only be noise.
  const [canChoose, setCanChoose] = useState(false)
  // The response is stored with the dormitory it was fetched for, so switching
  // shows the loading state without having to reset anything from the effect.
  const [loaded, setLoaded] = useState<{
    key: string
    summary: ApiDashboardSummary
    // Null when the role can't read activity logs; the card is only drawn once
    // it holds a list.
    recentActivity: ApiActivityLog[] | null
  } | null>(null)
  const [loadError, setLoadError] = useState<string | null>(null)

  useEffect(() => {
    let cancelled = false

    async function init() {
      let choices = 0
      try {
        const { data } = await api.get<ApiDormitory[]>('/dormitories/active', { params: { limit: 2 } })
        choices = data.length
      } catch {
        // No access to the dormitory list (or it failed): no selector.
      }

      const stored = choices > 1 ? readStoredDormitory() : null
      let verified: SelectedDormitory | null = null
      if (stored) {
        try {
          // Scoped to the user's access, so this fails for a dormitory they
          // no longer have.
          await api.get(`/dormitories/${stored.id}`)
          verified = stored
        } catch {
          // Fall through to the default and forget the stale choice.
        }
      }
      if (!verified) storeDormitory(null)

      if (cancelled) return
      setCanChoose(choices > 1)
      setDormitory(verified)
      setReady(true)
    }

    void init()

    return () => {
      cancelled = true
    }
  }, [])

  const dormitoryKey = dormitory?.id ?? ''

  useEffect(() => {
    if (!ready) return

    let cancelled = false
    const dormitoryParams = dormitory ? { dormitory_id: dormitory.id } : {}

    Promise.all([
      // Counted by the server, so the figures stay right past any page size.
      api.get<ApiDashboardSummary>('/dashboard/summary', { params: dormitoryParams }),
      // Needs its own menu permission, so a refusal hides this card rather
      // than failing the whole dashboard.
      api
        .get<ApiPage<ApiActivityLog[]>>('/activity-logs', { params: { per_page: 5, ...dormitoryParams } })
        .then((res) => res.data.data)
        .catch(() => null),
    ])
      .then(([summaryRes, activity]) => {
        if (cancelled) return
        setLoaded({ key: dormitoryKey, summary: summaryRes.data, recentActivity: activity })
        setLoadError(null)
      })
      .catch((err) => {
        if (!cancelled) setLoadError(extractErrorMessage(err, t('resourceLoadError')))
      })

    return () => {
      cancelled = true
    }
    // eslint-disable-next-line react-hooks/exhaustive-deps
  }, [ready, dormitoryKey])

  function selectDormitory(next: SelectedDormitory | null) {
    setDormitory(next)
    storeDormitory(next)
  }

  const summary = loaded?.summary ?? null
  const recentActivity = loaded?.recentActivity ?? null
  const isLoading = !loadError && (!ready || loaded?.key !== dormitoryKey)

  // A section the role can't read comes back null and simply gets no card.
  const metrics: Metric[] = []
  if (summary?.rooms) {
    const { total, available, occupied, maintenance } = summary.rooms
    metrics.push({
      key: 'rooms',
      title: t('dashboardTotalRooms'),
      value: total.toLocaleString(),
      detail: `${available.toLocaleString()} ${t('roomStatusAvailable')} · ${occupied.toLocaleString()} ${t('roomStatusOccupied')} · ${maintenance.toLocaleString()} ${t('roomStatusMaintenance')}`,
      icon: BedDouble,
      iconClass: 'bg-primary/10 text-primary',
    })
  }
  if (summary?.contracts) {
    metrics.push({
      key: 'contracts',
      title: t('dashboardActiveContracts'),
      value: summary.contracts.active.toLocaleString(),
      detail: `${summary.contracts.total.toLocaleString()} ${t('dashboardContractsTotalLabel')}`,
      icon: FileText,
      iconClass: 'bg-success/10 text-success',
    })
  }
  if (summary?.invoices) {
    metrics.push({
      key: 'invoices',
      title: t('dashboardUnpaidInvoices'),
      value: summary.invoices.outstanding_count.toLocaleString(),
      detail: `${summary.invoices.outstanding_amount.toLocaleString()} ${t('dashboardBahtUnit')}`,
      icon: ReceiptText,
      iconClass: 'bg-warning/10 text-warning',
    })
  }
  if (summary?.repairs) {
    metrics.push({
      key: 'repairs',
      title: t('dashboardPendingRepairs'),
      value: summary.repairs.pending.toLocaleString(),
      detail: `${summary.repairs.in_progress.toLocaleString()} ${t('repairStatusInProgress')}`,
      icon: Wrench,
      iconClass: 'bg-violet-500/10 text-violet-600 dark:text-violet-400',
    })
  }

  return (
    <main className="content">
      <section className="welcome">
        <h1>{t('dashboard')}</h1>
        <p>{t('dashboardOverview')}</p>
      </section>

      {canChoose && (
        <div className="flex max-w-sm flex-col gap-1.5 text-sm font-medium">
          {t('dashboardDormitoryLabel')}
          <DormitorySearchSelect
            selectedLabel={dormitory?.name ?? ''}
            onSelectDormitory={(selected) => selectDormitory({ id: selected.id, name: selected.name })}
            clearLabel={t('dashboardAllDormitories')}
            onClear={() => selectDormitory(null)}
            placeholder={t('dashboardAllDormitories')}
            searchPlaceholder={t('dashboardDormitorySearchPlaceholder')}
            noResultsLabel={t('dashboardDormitoryNoResults')}
          />
        </div>
      )}

      {loadError && <p className="resource-error">{loadError}</p>}

      {!loadError && isLoading && (
        <section className="metric-grid">
          {Array.from({ length: 4 }).map((_, index) => (
            <Card key={index}>
              <CardHeader>
                <div className="h-3 w-24 animate-pulse rounded bg-muted" />
                <div className="mt-2 h-7 w-16 animate-pulse rounded bg-muted" />
              </CardHeader>
              <CardContent>
                <div className="h-3 w-full animate-pulse rounded bg-muted" />
              </CardContent>
            </Card>
          ))}
        </section>
      )}

      {!loadError && !isLoading && (
        <>
          <section className="metric-grid">
            {metrics.map((metric) => (
              <Card key={metric.key} className="transition-shadow hover:shadow-md">
                <CardHeader className="flex-row items-start justify-between space-y-0">
                  <div>
                    <CardDescription>{metric.title}</CardDescription>
                    <CardTitle>{metric.value}</CardTitle>
                  </div>
                  <span className={`flex h-10 w-10 shrink-0 items-center justify-center rounded-full ${metric.iconClass}`}>
                    <metric.icon size={20} strokeWidth={2} />
                  </span>
                </CardHeader>
                <CardContent>
                  <p className="metric-detail">{metric.detail}</p>
                </CardContent>
              </Card>
            ))}
          </section>

          {recentActivity && (
          <Card>
            <CardHeader>
              <CardTitle>{t('dashboardRecentActivity')}</CardTitle>
              <CardDescription>{t('dashboardRecentActivityDescription')}</CardDescription>
            </CardHeader>
            <CardContent>
              {recentActivity && recentActivity.length === 0 && (
                <p className="metric-detail">{t('dashboardNoRecentActivity')}</p>
              )}

              {recentActivity && recentActivity.length > 0 && (
                <div className="table-wrap">
                  <Table>
                    <TableHeader>
                      <TableRow>
                        <TableHead>{t('activityLogTimeColumn')}</TableHead>
                        <TableHead>{t('activityLogUserColumn')}</TableHead>
                        <TableHead>{t('activityLogActionColumn')}</TableHead>
                        <TableHead className="text-right">{t('activityLogDescriptionColumn')}</TableHead>
                      </TableRow>
                    </TableHeader>
                    <TableBody>
                      {recentActivity.map((log) => (
                        <TableRow key={log.id}>
                          <TableCell className="text-muted-foreground">
                            {new Date(log.created_at).toLocaleString(language === 'th' ? 'th-TH' : 'en-US')}
                          </TableCell>
                          <TableCell className="font-semibold">
                            {log.username || t('activityLogSystemUser')}
                          </TableCell>
                          <TableCell>
                            <Badge variant={activityLogActionVariant(log.action)}>{log.action}</Badge>
                          </TableCell>
                          <TableCell className="text-right text-muted-foreground">
                            {log.description || '—'}
                          </TableCell>
                        </TableRow>
                      ))}
                    </TableBody>
                  </Table>
                </div>
              )}
            </CardContent>
          </Card>
          )}
        </>
      )}
    </main>
  )
}
