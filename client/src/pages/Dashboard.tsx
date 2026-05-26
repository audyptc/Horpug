import { useState, useEffect } from 'react'
import { useTranslation } from 'react-i18next'
import {
  Home, Users, FileText, AlertCircle, CheckCircle2,
  TrendingUp, DoorOpen, Wrench, Clock,
} from 'lucide-react'
import { Card, CardContent, CardHeader, CardTitle, CardDescription } from '@/components/ui/card'
import { Badge } from '@/components/ui/badge'
import { dashboardService } from '@/services/dashboardService'
import type { ApiDashboardSummary } from '@/types/api'
import { cn } from '@/lib/utils'

function formatBaht(amount: number) {
  return new Intl.NumberFormat('th-TH', { minimumFractionDigits: 0, maximumFractionDigits: 0 }).format(amount)
}

export function Dashboard() {
  const { t } = useTranslation()
  const [summary, setSummary] = useState<ApiDashboardSummary | null>(null)
  const [loading, setLoading] = useState(true)

  useEffect(() => {
    dashboardService.summary()
      .then(setSummary)
      .catch(() => {})
      .finally(() => setLoading(false))
  }, [])

  const statCards = [
    {
      title: t('dashboard.totalRooms'),
      value: summary?.total_rooms ?? '—',
      sub: loading ? '' : t('dashboard.occupancyRate', { rate: summary?.occupancy_rate.toFixed(1) ?? 0 }),
      icon: Home,
      color: 'text-violet-500',
      bg: 'bg-violet-500/10',
      accent: 'bg-violet-500',
    },
    {
      title: t('dashboard.totalTenants'),
      value: summary?.total_tenants ?? '—',
      sub: loading ? '' : t('dashboard.activeContracts', { count: summary?.active_contracts ?? 0 }),
      icon: Users,
      color: 'text-emerald-500',
      bg: 'bg-emerald-500/10',
      accent: 'bg-emerald-500',
    },
    {
      title: t('dashboard.revenueThisMonth'),
      value: loading ? '—' : `฿${formatBaht(summary?.revenue_this_month ?? 0)}`,
      sub: loading ? '' : t('dashboard.paidBillsThisMonth'),
      icon: TrendingUp,
      color: 'text-blue-500',
      bg: 'bg-blue-500/10',
      accent: 'bg-blue-500',
    },
    {
      title: t('dashboard.pendingBills'),
      value: loading ? '—' : (summary?.unpaid_bills ?? 0) + (summary?.overdue_bills ?? 0),
      sub: loading ? '' : t('dashboard.overdueCount', { count: summary?.overdue_bills ?? 0 }),
      icon: AlertCircle,
      color: 'text-rose-500',
      bg: 'bg-rose-500/10',
      accent: 'bg-rose-500',
    },
  ]

  return (
    <div className="space-y-6">
      <div>
        <h1 className="text-2xl font-bold tracking-tight">{t('dashboard.title')}</h1>
        <p className="text-muted-foreground text-sm mt-1">{t('dashboard.subtitle')}</p>
      </div>

      {/* Stat cards */}
      <div className="grid grid-cols-1 sm:grid-cols-2 xl:grid-cols-4 gap-4">
        {statCards.map((stat) => {
          const Icon = stat.icon
          return (
            <Card key={stat.title} className="hover:shadow-md transition-shadow overflow-hidden">
              <div className={cn('h-1 w-full', stat.accent)} />
              <CardHeader className="flex flex-row items-center justify-between pb-2">
                <CardTitle className="text-sm font-medium text-muted-foreground">{stat.title}</CardTitle>
                <div className={cn('p-2 rounded-lg', stat.bg)}>
                  <Icon className={cn('w-4 h-4', stat.color)} />
                </div>
              </CardHeader>
              <CardContent>
                <div className="text-3xl font-bold">{stat.value}</div>
                {stat.sub && <p className="text-xs text-muted-foreground mt-1">{stat.sub}</p>}
              </CardContent>
            </Card>
          )
        })}
      </div>

      <div className="grid grid-cols-1 xl:grid-cols-3 gap-6">
        {/* Room status */}
        <Card className="xl:col-span-2">
          <CardHeader>
            <CardTitle>{t('dashboard.roomStatus')}</CardTitle>
            <CardDescription>{t('dashboard.roomStatusDesc')}</CardDescription>
          </CardHeader>
          <CardContent className="space-y-4">
            {loading ? (
              <p className="text-sm text-muted-foreground py-4 text-center">{t('common.loading')}</p>
            ) : (
              <>
                {/* Occupancy bar */}
                <div className="space-y-1.5">
                  <div className="flex justify-between text-sm">
                    <span className="font-medium">{t('dashboard.occupancyRate', { rate: summary?.occupancy_rate.toFixed(1) ?? 0 })}</span>
                    <span className="text-muted-foreground">{summary?.occupied_rooms}/{summary?.total_rooms} {t('dashboard.rooms')}</span>
                  </div>
                  <div className="h-3 rounded-full bg-muted overflow-hidden">
                    <div
                      className="h-full rounded-full bg-emerald-500 transition-all"
                      style={{ width: `${summary?.occupancy_rate ?? 0}%` }}
                    />
                  </div>
                </div>

                {/* Room status breakdown */}
                <div className="grid grid-cols-3 gap-3 pt-2">
                  {[
                    { label: t('dashboard.occupied'), value: summary?.occupied_rooms ?? 0, icon: CheckCircle2, color: 'text-emerald-500', bg: 'bg-emerald-500/10' },
                    { label: t('dashboard.available'), value: summary?.available_rooms ?? 0, icon: DoorOpen, color: 'text-blue-500', bg: 'bg-blue-500/10' },
                    { label: t('dashboard.maintenance'), value: summary?.maintenance_rooms ?? 0, icon: Wrench, color: 'text-amber-500', bg: 'bg-amber-500/10' },
                  ].map((item) => {
                    const Icon = item.icon
                    return (
                      <div key={item.label} className={cn('rounded-xl p-4 flex flex-col items-center gap-2', item.bg)}>
                        <Icon className={cn('w-5 h-5', item.color)} />
                        <span className="text-2xl font-bold">{item.value}</span>
                        <span className="text-xs text-muted-foreground text-center">{item.label}</span>
                      </div>
                    )
                  })}
                </div>
              </>
            )}
          </CardContent>
        </Card>

        {/* Alerts */}
        <Card>
          <CardHeader>
            <CardTitle>{t('dashboard.alerts')}</CardTitle>
            <CardDescription>{t('dashboard.alertsDesc')}</CardDescription>
          </CardHeader>
          <CardContent className="space-y-3">
            {loading ? (
              <p className="text-sm text-muted-foreground py-4 text-center">{t('common.loading')}</p>
            ) : (
              <>
                <AlertRow
                  icon={AlertCircle}
                  iconClass="text-rose-500"
                  label={t('dashboard.overdueBills')}
                  value={summary?.overdue_bills ?? 0}
                  amount={summary?.overdue_amount ?? 0}
                  variant={summary?.overdue_bills ? 'destructive' : 'secondary'}
                />
                <AlertRow
                  icon={FileText}
                  iconClass="text-amber-500"
                  label={t('dashboard.unpaidBills')}
                  value={summary?.unpaid_bills ?? 0}
                  amount={summary?.unpaid_amount ?? 0}
                  variant={summary?.unpaid_bills ? 'warning' : 'secondary'}
                />
                <AlertRow
                  icon={Clock}
                  iconClass="text-blue-500"
                  label={t('dashboard.expiringContracts')}
                  value={summary?.expiring_contracts ?? 0}
                  variant={summary?.expiring_contracts ? 'outline' : 'secondary'}
                />
              </>
            )}
          </CardContent>
        </Card>
      </div>
    </div>
  )
}

function AlertRow({
  icon: Icon,
  iconClass,
  label,
  value,
  amount,
  variant,
}: {
  icon: React.ElementType
  iconClass: string
  label: string
  value: number
  amount?: number
  variant: 'destructive' | 'warning' | 'outline' | 'secondary'
}) {
  const { t } = useTranslation()
  return (
    <div className="flex items-center gap-3 p-3 rounded-lg border bg-card">
      <Icon className={cn('w-4 h-4 shrink-0', iconClass)} />
      <div className="flex-1 min-w-0">
        <p className="text-sm font-medium">{label}</p>
        {amount !== undefined && amount > 0 && (
          <p className="text-xs text-muted-foreground">฿{new Intl.NumberFormat('th-TH').format(amount)}</p>
        )}
      </div>
      <Badge variant={variant as any}>
        {value} {t('dashboard.items')}
      </Badge>
    </div>
  )
}
