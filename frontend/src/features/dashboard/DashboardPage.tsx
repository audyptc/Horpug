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
import { activityLogActionVariant } from '@/features/activitylog/utils'

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

  const [summary, setSummary] = useState<ApiDashboardSummary | null>(null)
  // Null both while loading and when the role can't read activity logs; the
  // card is only drawn once it holds a list.
  const [recentActivity, setRecentActivity] = useState<ApiActivityLog[] | null>(null)
  const [loadError, setLoadError] = useState<string | null>(null)

  useEffect(() => {
    let cancelled = false

    Promise.all([
      // Counted by the server, so the figures stay right past any page size.
      api.get<ApiDashboardSummary>('/dashboard/summary'),
      // Needs its own menu permission, so a refusal hides this card rather
      // than failing the whole dashboard.
      api
        .get<ApiPage<ApiActivityLog[]>>('/activity-logs', { params: { per_page: 5 } })
        .then((res) => res.data.data)
        .catch(() => null),
    ])
      .then(([summaryRes, activity]) => {
        if (cancelled) return
        setSummary(summaryRes.data)
        setRecentActivity(activity)
      })
      .catch((err) => {
        if (!cancelled) setLoadError(extractErrorMessage(err, t('resourceLoadError')))
      })

    return () => {
      cancelled = true
    }
    // eslint-disable-next-line react-hooks/exhaustive-deps
  }, [])

  const isLoading = !loadError && summary === null

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
