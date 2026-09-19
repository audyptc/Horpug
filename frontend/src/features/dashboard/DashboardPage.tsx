import { useEffect, useState } from 'react'
import {
  BanknoteArrowDown,
  BanknoteArrowUp,
  BedDouble,
  CalendarClock,
  Droplets,
  FileText,
  ReceiptText,
  TriangleAlert,
  Wrench,
  Zap,
  type LucideIcon,
} from 'lucide-react'
import { api, extractErrorMessage, type ApiPage } from '@/shared/api/client'
import { useLanguage } from '@/shared/i18n/language'
import { cn } from '@/shared/lib/utils'
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

const moneyFormat = { maximumFractionDigits: 2 }

// The cards are shown in these groups, in this order, so the page reads as
// "how things stand", "the money", then "what needs doing". A group with no
// card (the role can't read any of its sections) is left out entirely.
const METRIC_GROUPS = [
  { key: 'overview', titleKey: 'dashboardGroupOverview' },
  { key: 'finance', titleKey: 'dashboardGroupFinance' },
  { key: 'attention', titleKey: 'dashboardGroupAttention' },
] as const

type MetricGroup = (typeof METRIC_GROUPS)[number]['key']

// Order of the cards within their group: money in first, then what is still owed.
const METRIC_ORDER = [
  'rooms',
  'contracts',
  'repairs',
  'payments',
  'expenses',
  'invoices',
  'invoices-overdue',
  'contracts-expiring',
  'electricity-readings',
  'water-readings',
]

type Metric = {
  key: string
  group: MetricGroup
  // Draws attention to the card: there is something to act on.
  alert?: boolean
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

  // The month the "this month" figures cover is the server's, so the label
  // can't disagree with the numbers if the browser's clock or zone differs.
  const monthLabel = summary
    ? new Date(summary.period.year, summary.period.month - 1, 1).toLocaleDateString(
        language === 'th' ? 'th-TH' : 'en-US',
        { month: 'long', year: 'numeric' }
      )
    : ''

  // A section the role can't read comes back null and simply gets no card.
  const metrics: Metric[] = []
  if (summary?.rooms) {
    const { total, available, occupied, maintenance } = summary.rooms
    metrics.push({
      key: 'rooms',
      group: 'overview',
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
      group: 'overview',
      title: t('dashboardActiveContracts'),
      value: summary.contracts.active.toLocaleString(),
      detail: `${summary.contracts.total.toLocaleString()} ${t('dashboardContractsTotalLabel')}`,
      icon: FileText,
      iconClass: 'bg-success/10 text-success',
    })
  }
  if (summary?.contracts) {
    metrics.push({
      key: 'contracts-expiring',
      group: 'attention',
      alert: summary.contracts.expiring_soon + summary.contracts.past_end > 0,
      title: t('dashboardExpiringContracts'),
      value: summary.contracts.expiring_soon.toLocaleString(),
      detail: t('dashboardExpiringContractsDetail').replace('{count}', summary.contracts.past_end.toLocaleString()),
      icon: CalendarClock,
      iconClass: 'bg-sky-500/10 text-sky-600 dark:text-sky-400',
    })
  }
  if (summary?.invoices) {
    metrics.push({
      key: 'invoices',
      group: 'finance',
      title: t('dashboardUnpaidInvoices'),
      value: summary.invoices.outstanding_count.toLocaleString(),
      detail: `${summary.invoices.outstanding_amount.toLocaleString(undefined, moneyFormat)} ${t('dashboardBahtUnit')}`,
      icon: ReceiptText,
      iconClass: 'bg-warning/10 text-warning',
    })
    metrics.push({
      key: 'invoices-overdue',
      group: 'finance',
      alert: summary.invoices.overdue_count > 0,
      title: t('dashboardOverdueInvoices'),
      value: summary.invoices.overdue_count.toLocaleString(),
      detail: `${summary.invoices.overdue_amount.toLocaleString(undefined, moneyFormat)} ${t('dashboardOverdueAmountUnit')}`,
      icon: TriangleAlert,
      iconClass: 'bg-destructive/10 text-destructive',
    })
  }
  if (summary?.repairs) {
    metrics.push({
      key: 'repairs',
      group: 'overview',
      title: t('dashboardPendingRepairs'),
      value: summary.repairs.pending.toLocaleString(),
      detail: `${summary.repairs.in_progress.toLocaleString()} ${t('repairStatusInProgress')}`,
      icon: Wrench,
      iconClass: 'bg-violet-500/10 text-violet-600 dark:text-violet-400',
    })
  }
  if (summary?.payments) {
    const received = summary.payments.received_this_month
    // Net needs both sides; without read access to expenses it is left out
    // rather than shown as though expenses were zero.
    const detail = summary.expenses
      ? t('dashboardIncomeNetDetail')
          .replace('{expenses}', summary.expenses.this_month.toLocaleString(undefined, moneyFormat))
          .replace('{net}', (received - summary.expenses.this_month).toLocaleString(undefined, moneyFormat))
      : t('dashboardIncomeDetail').replace('{month}', monthLabel)
    metrics.push({
      key: 'payments',
      group: 'finance',
      title: t('dashboardIncomeThisMonth'),
      value: received.toLocaleString(undefined, moneyFormat),
      detail,
      icon: BanknoteArrowUp,
      iconClass: 'bg-success/10 text-success',
    })
  }
  if (summary?.expenses) {
    metrics.push({
      key: 'expenses',
      group: 'finance',
      title: t('dashboardExpensesThisMonth'),
      value: summary.expenses.this_month.toLocaleString(undefined, moneyFormat),
      detail: t('dashboardExpensesDetail').replace('{month}', monthLabel),
      icon: BanknoteArrowDown,
      iconClass: 'bg-orange-500/10 text-orange-600 dark:text-orange-400',
    })
  }
  if (summary?.electricity_readings) {
    metrics.push({
      key: 'electricity-readings',
      group: 'attention',
      alert: summary.electricity_readings.missing > 0,
      title: t('dashboardMissingElectricity'),
      value: summary.electricity_readings.missing.toLocaleString(),
      detail: t('dashboardMissingReadingsDetail').replace('{month}', monthLabel),
      icon: Zap,
      iconClass: 'bg-yellow-500/10 text-yellow-600 dark:text-yellow-400',
    })
  }
  if (summary?.water_readings) {
    metrics.push({
      key: 'water-readings',
      group: 'attention',
      alert: summary.water_readings.missing > 0,
      title: t('dashboardMissingWater'),
      value: summary.water_readings.missing.toLocaleString(),
      detail: t('dashboardMissingReadingsDetail').replace('{month}', monthLabel),
      icon: Droplets,
      iconClass: 'bg-cyan-500/10 text-cyan-600 dark:text-cyan-400',
    })
  }

  return (
    <main className="content">
      <section className="welcome">
        <h1>{t('dashboard')}</h1>
        <p>{t('dashboardOverview')}</p>
      </section>

      {canChoose && (
        <div className="flex flex-wrap items-center gap-x-3 gap-y-2 px-1 text-sm font-medium">
          <span className="text-muted-foreground">{t('dashboardDormitoryLabel')}</span>
          <div className="w-full sm:w-72">
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
        </div>
      )}

      {loadError && <p className="resource-error">{loadError}</p>}

      {!loadError && isLoading && (
        <section className="metric-grid" aria-busy="true">
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
          {METRIC_GROUPS.map((group) => {
            const items = metrics
              .filter((metric) => metric.group === group.key)
              .sort((x, y) => METRIC_ORDER.indexOf(x.key) - METRIC_ORDER.indexOf(y.key))
            if (items.length === 0) return null

            return (
              <section key={group.key} className="flex flex-col gap-3">
                <h2 className="px-1 text-xs font-semibold uppercase tracking-wider text-muted-foreground">
                  {t(group.titleKey)}
                </h2>
                <div className="metric-grid">
                  {items.map((metric) => (
                    <Card
                      key={metric.key}
                      className={cn(
                        'transition-shadow hover:shadow-md',
                        metric.alert && 'border-warning/60 ring-1 ring-warning/25'
                      )}
                    >
                      <CardHeader className="flex-row items-start justify-between space-y-0">
                        <div>
                          <CardDescription>{metric.title}</CardDescription>
                          <CardTitle className={cn(metric.alert && 'text-warning')}>{metric.value}</CardTitle>
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
                </div>
              </section>
            )
          })}

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
