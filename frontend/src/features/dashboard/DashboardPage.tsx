import { useEffect, useMemo, useState } from 'react'
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
import type { ApiRoom } from '@/features/room/types'
import type { ApiContract } from '@/features/contract/types'
import type { ApiInvoice } from '@/features/invoice/types'
import type { ApiRepairRequest } from '@/features/repairrequest/types'
import type { ApiActivityLog } from '@/features/activitylog/types'
import { activityLogActionVariant } from '@/features/activitylog/utils'

export function DashboardPage() {
  const { t, language } = useLanguage()

  const [rooms, setRooms] = useState<ApiRoom[] | null>(null)
  const [contracts, setContracts] = useState<ApiContract[] | null>(null)
  const [invoices, setInvoices] = useState<ApiInvoice[] | null>(null)
  const [repairRequests, setRepairRequests] = useState<ApiRepairRequest[] | null>(null)
  const [recentActivity, setRecentActivity] = useState<ApiActivityLog[] | null>(null)
  const [loadError, setLoadError] = useState<string | null>(null)

  useEffect(() => {
    let cancelled = false

    Promise.all([
      api.get<ApiPage<ApiRoom[]>>('/rooms', { params: { per_page: 100 } }),
      api.get<ApiPage<ApiContract[]>>('/contracts', { params: { per_page: 100 } }),
      api.get<ApiPage<ApiInvoice[]>>('/invoices', { params: { per_page: 100 } }),
      api.get<ApiPage<ApiRepairRequest[]>>('/repair-requests', { params: { per_page: 100 } }),
      api.get<ApiPage<ApiActivityLog[]>>('/activity-logs', { params: { per_page: 5 } }),
    ])
      .then(([roomsRes, contractsRes, invoicesRes, repairRequestsRes, activityRes]) => {
        if (cancelled) return
        setRooms(roomsRes.data.data)
        setContracts(contractsRes.data.data)
        setInvoices(invoicesRes.data.data)
        setRepairRequests(repairRequestsRes.data.data)
        setRecentActivity(activityRes.data.data)
      })
      .catch((err) => {
        if (!cancelled) setLoadError(extractErrorMessage(err, t('resourceLoadError')))
      })

    return () => {
      cancelled = true
    }
    // eslint-disable-next-line react-hooks/exhaustive-deps
  }, [])

  const isLoading = !loadError && (rooms === null || contracts === null || invoices === null || repairRequests === null)

  const roomStats = useMemo(() => {
    const list = rooms ?? []
    return {
      total: list.length,
      available: list.filter((room) => room.status === 'available').length,
      occupied: list.filter((room) => room.status === 'occupied').length,
      maintenance: list.filter((room) => room.status === 'maintenance').length,
    }
  }, [rooms])

  const contractStats = useMemo(() => {
    const list = contracts ?? []
    return {
      total: list.length,
      active: list.filter((contract) => contract.status === 'active').length,
    }
  }, [contracts])

  const invoiceStats = useMemo(() => {
    const list = invoices ?? []
    const outstanding = list.filter((invoice) => invoice.status === 'unpaid' || invoice.status === 'overdue')
    return {
      count: outstanding.length,
      amount: outstanding.reduce((sum, invoice) => sum + invoice.total_amount, 0),
    }
  }, [invoices])

  const repairStats = useMemo(() => {
    const list = repairRequests ?? []
    return {
      pending: list.filter((request) => request.status === 'pending').length,
      inProgress: list.filter((request) => request.status === 'in_progress').length,
    }
  }, [repairRequests])

  const metrics = [
    {
      key: 'rooms',
      title: t('dashboardTotalRooms'),
      value: roomStats.total.toLocaleString(),
      detail: `${roomStats.available.toLocaleString()} ${t('roomStatusAvailable')} · ${roomStats.occupied.toLocaleString()} ${t('roomStatusOccupied')} · ${roomStats.maintenance.toLocaleString()} ${t('roomStatusMaintenance')}`,
    },
    {
      key: 'contracts',
      title: t('dashboardActiveContracts'),
      value: contractStats.active.toLocaleString(),
      detail: `${contractStats.total.toLocaleString()} ${t('dashboardContractsTotalLabel')}`,
    },
    {
      key: 'invoices',
      title: t('dashboardUnpaidInvoices'),
      value: invoiceStats.count.toLocaleString(),
      detail: `${invoiceStats.amount.toLocaleString()} ${t('dashboardBahtUnit')}`,
    },
    {
      key: 'repairs',
      title: t('dashboardPendingRepairs'),
      value: repairStats.pending.toLocaleString(),
      detail: `${repairStats.inProgress.toLocaleString()} ${t('repairStatusInProgress')}`,
    },
  ]

  return (
    <main className="content">
      <section className="welcome">
        <h1>{t('dashboard')}</h1>
        <p>{t('dashboardOverview')}</p>
      </section>

      {loadError && <p className="resource-error">{loadError}</p>}

      {!loadError && isLoading && <p className="metric-detail">{t('loading')}</p>}

      {!loadError && !isLoading && (
        <>
          <section className="metric-grid">
            {metrics.map((metric) => (
              <Card key={metric.key}>
                <CardHeader>
                  <CardDescription>{metric.title}</CardDescription>
                  <CardTitle>{metric.value}</CardTitle>
                </CardHeader>
                <CardContent>
                  <p className="metric-detail">{metric.detail}</p>
                </CardContent>
              </Card>
            ))}
          </section>

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
        </>
      )}
    </main>
  )
}
