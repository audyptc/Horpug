import { useState, useEffect } from 'react'
import { useTranslation } from 'react-i18next'
import { Card, CardContent, CardHeader, CardTitle, CardDescription } from '@/components/ui/card'
import { analyticsService } from '@/services/analyticsService'
import type { ApiAnalyticsSummary } from '@/types/api'
import { TrendingUp, Banknote, BarChart3, Users } from 'lucide-react'
import { cn } from '@/lib/utils'

function formatBaht(v: number) {
  return new Intl.NumberFormat('th-TH', { minimumFractionDigits: 0, maximumFractionDigits: 0 }).format(v)
}

function shortMonth(yyyyMM: string) {
  const [year, month] = yyyyMM.split('-')
  const d = new Date(Number(year), Number(month) - 1, 1)
  return d.toLocaleString('default', { month: 'short' })
}

export function Analytics() {
  const { t } = useTranslation()
  const [data, setData] = useState<ApiAnalyticsSummary | null>(null)
  const [loading, setLoading] = useState(true)

  useEffect(() => {
    analyticsService.summary(12)
      .then(setData)
      .catch(() => {})
      .finally(() => setLoading(false))
  }, [])

  const maxRevenue = data ? Math.max(...data.monthly_revenue.map((m) => m.revenue), 1) : 1
  const maxTenants = data ? Math.max(...data.monthly_tenants.map((m) => m.count), 1) : 1
  const totalBills = data
    ? data.bill_status.unpaid + data.bill_status.paid + data.bill_status.overdue
    : 0

  const summaryCards = [
    {
      label: t('analytics.totalRevenue'),
      value: loading ? '—' : `฿${formatBaht(data?.total_revenue ?? 0)}`,
      sub: t('analytics.last12Months'),
      icon: Banknote,
      color: 'text-emerald-500',
      bg: 'bg-emerald-500/10',
    },
    {
      label: t('analytics.avgMonthly'),
      value: loading ? '—' : `฿${formatBaht(data?.avg_monthly ?? 0)}`,
      sub: t('analytics.perMonth'),
      icon: TrendingUp,
      color: 'text-blue-500',
      bg: 'bg-blue-500/10',
    },
    {
      label: t('analytics.paidBills'),
      value: loading ? '—' : data?.bill_status.paid ?? 0,
      sub: loading ? '' : t('analytics.ofTotal', { total: totalBills }),
      icon: BarChart3,
      color: 'text-violet-500',
      bg: 'bg-violet-500/10',
    },
    {
      label: t('analytics.newTenants'),
      value: loading ? '—' : data?.monthly_tenants.reduce((s, m) => s + m.count, 0) ?? 0,
      sub: t('analytics.last12Months'),
      icon: Users,
      color: 'text-amber-500',
      bg: 'bg-amber-500/10',
    },
  ]

  return (
    <div className="space-y-6">
      <div>
        <h1 className="text-2xl font-bold tracking-tight">{t('analytics.title')}</h1>
        <p className="text-muted-foreground text-sm mt-1">{t('analytics.subtitle')}</p>
      </div>

      {/* Summary cards */}
      <div className="grid grid-cols-2 md:grid-cols-4 gap-4">
        {summaryCards.map((item) => {
          const Icon = item.icon
          return (
            <Card key={item.label} className="hover:shadow-md transition-shadow">
              <CardHeader className="flex flex-row items-center justify-between pb-2">
                <p className="text-sm font-medium text-muted-foreground leading-snug">{item.label}</p>
                <div className={cn('p-2 rounded-lg shrink-0', item.bg)}>
                  <Icon className={cn('w-4 h-4', item.color)} />
                </div>
              </CardHeader>
              <CardContent>
                <p className="text-2xl font-bold">{item.value}</p>
                <p className="text-xs text-muted-foreground mt-1">{item.sub}</p>
              </CardContent>
            </Card>
          )
        })}
      </div>

      <div className="grid grid-cols-1 lg:grid-cols-3 gap-6">
        {/* Monthly revenue bar chart */}
        <Card className="lg:col-span-2">
          <CardHeader className="flex flex-row items-center gap-3 pb-4">
            <div className="p-2 rounded-lg bg-primary/10 shrink-0">
              <BarChart3 className="w-4 h-4 text-primary" />
            </div>
            <div>
              <CardTitle>{t('analytics.monthlyRevenue')}</CardTitle>
              <CardDescription>{t('analytics.monthlyRevenueDesc')}</CardDescription>
            </div>
          </CardHeader>
          <CardContent>
            {loading ? (
              <div className="flex items-center justify-center h-52 text-sm text-muted-foreground">
                {t('common.loading')}
              </div>
            ) : (
              <div className="flex items-end gap-2 h-52 px-1">
                {(data?.monthly_revenue ?? []).map((m) => (
                  <div key={m.month} className="flex-1 flex flex-col items-center gap-1.5">
                    {m.revenue > 0 && (
                      <span className="text-[10px] font-medium text-muted-foreground">
                        {formatBaht(m.revenue)}
                      </span>
                    )}
                    <div
                      className="w-full rounded-t-md bg-primary/70 hover:bg-primary transition-colors cursor-default"
                      style={{ height: `${Math.max((m.revenue / maxRevenue) * 100, m.revenue > 0 ? 4 : 0)}%` }}
                      title={`฿${formatBaht(m.revenue)}`}
                    />
                    <span className="text-[10px] font-medium text-muted-foreground">
                      {shortMonth(m.month)}
                    </span>
                  </div>
                ))}
              </div>
            )}
          </CardContent>
        </Card>

        {/* Bill status breakdown */}
        <Card>
          <CardHeader className="pb-4">
            <CardTitle>{t('analytics.billStatus')}</CardTitle>
            <CardDescription>{t('analytics.billStatusDesc')}</CardDescription>
          </CardHeader>
          <CardContent className="space-y-4">
            {loading ? (
              <div className="flex items-center justify-center py-8 text-sm text-muted-foreground">
                {t('common.loading')}
              </div>
            ) : (
              <>
                {[
                  { key: 'paid', label: t('analytics.statusPaid'), count: data?.bill_status.paid ?? 0, color: 'bg-emerald-500', text: 'text-emerald-500' },
                  { key: 'unpaid', label: t('analytics.statusUnpaid'), count: data?.bill_status.unpaid ?? 0, color: 'bg-amber-500', text: 'text-amber-500' },
                  { key: 'overdue', label: t('analytics.statusOverdue'), count: data?.bill_status.overdue ?? 0, color: 'bg-rose-500', text: 'text-rose-500' },
                ].map((item) => (
                  <div key={item.key} className="space-y-1.5">
                    <div className="flex items-center justify-between">
                      <div className="flex items-center gap-2">
                        <div className={cn('w-2.5 h-2.5 rounded-full', item.color)} />
                        <span className="text-sm font-medium">{item.label}</span>
                      </div>
                      <div className="flex items-center gap-2">
                        <span className={cn('text-xs font-semibold', item.text)}>{item.count}</span>
                        <span className="text-xs text-muted-foreground w-8 text-right">
                          {totalBills > 0 ? Math.round((item.count / totalBills) * 100) : 0}%
                        </span>
                      </div>
                    </div>
                    <div className="h-2 rounded-full bg-muted overflow-hidden">
                      <div
                        className={cn('h-full rounded-full transition-all', item.color)}
                        style={{ width: totalBills > 0 ? `${(item.count / totalBills) * 100}%` : '0%' }}
                      />
                    </div>
                  </div>
                ))}
              </>
            )}
          </CardContent>
        </Card>
      </div>

      {/* Monthly new tenants */}
      <Card>
        <CardHeader className="flex flex-row items-center gap-3 pb-4">
          <div className="p-2 rounded-lg bg-amber-500/10 shrink-0">
            <Users className="w-4 h-4 text-amber-500" />
          </div>
          <div>
            <CardTitle>{t('analytics.monthlyTenants')}</CardTitle>
            <CardDescription>{t('analytics.monthlyTenantsDesc')}</CardDescription>
          </div>
        </CardHeader>
        <CardContent>
          {loading ? (
            <div className="flex items-center justify-center h-32 text-sm text-muted-foreground">
              {t('common.loading')}
            </div>
          ) : (
            <div className="flex items-end gap-2 h-32 px-1">
              {(data?.monthly_tenants ?? []).map((m) => (
                <div key={m.month} className="flex-1 flex flex-col items-center gap-1.5">
                  {m.count > 0 && (
                    <span className="text-[10px] font-medium text-muted-foreground">{m.count}</span>
                  )}
                  <div
                    className="w-full rounded-t-md bg-amber-500/70 hover:bg-amber-500 transition-colors cursor-default"
                    style={{ height: `${Math.max((m.count / maxTenants) * 100, m.count > 0 ? 6 : 0)}%` }}
                    title={`${m.count} tenants`}
                  />
                  <span className="text-[10px] font-medium text-muted-foreground">
                    {shortMonth(m.month)}
                  </span>
                </div>
              ))}
            </div>
          )}
        </CardContent>
      </Card>
    </div>
  )
}
