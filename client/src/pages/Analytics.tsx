import { useState, useEffect, useMemo } from 'react'
import { useTranslation } from 'react-i18next'
import { Card, CardContent, CardHeader, CardTitle, CardDescription } from '@/components/ui/card'
import { Badge } from '@/components/ui/badge'
import { userService } from '@/services/userService'
import type { ApiUser } from '@/types/api'
import { TrendingUp, Clock, Eye, Users, BarChart3 } from 'lucide-react'
import { cn } from '@/lib/utils'

const months = ['Jan', 'Feb', 'Mar', 'Apr', 'May', 'Jun']
const signups = [3, 5, 2, 7, 4, 8]
const maxSignup = Math.max(...signups)

const roleColors: Record<string, string> = {
  admin: 'bg-violet-500',
  manager: 'bg-blue-500',
  editor: 'bg-amber-500',
  viewer: 'bg-slate-400',
}

const roleTextColors: Record<string, string> = {
  admin: 'text-violet-500',
  manager: 'text-blue-500',
  editor: 'text-amber-500',
  viewer: 'text-slate-400',
}

export function Analytics() {
  const { t } = useTranslation()
  const [users, setUsers] = useState<ApiUser[]>([])
  const [loading, setLoading] = useState(true)

  useEffect(() => {
    userService.list(1, 9999)
      .then((resp) => setUsers(resp.data))
      .catch(() => {})
      .finally(() => setLoading(false))
  }, [])

  const byRole = useMemo(() => {
    return users.reduce<Record<string, number>>((acc, u) => {
      const name = u.role?.name.toLowerCase() ?? 'unassigned'
      acc[name] = (acc[name] ?? 0) + 1
      return acc
    }, {})
  }, [users])

  const summaryCards = [
    { label: t('analytics.avgSession'), value: '24m 18s', sub: t('analytics.perUser'), icon: Clock, color: 'text-violet-500', bg: 'bg-violet-500/10' },
    { label: t('analytics.pageViews'), value: '1,284', sub: t('analytics.thisMonth'), icon: Eye, color: 'text-blue-500', bg: 'bg-blue-500/10' },
    { label: t('analytics.retention'), value: '76%', sub: t('analytics.monthlyActive'), icon: Users, color: 'text-emerald-500', bg: 'bg-emerald-500/10' },
    { label: t('analytics.growthRate'), value: '+12.5%', sub: t('analytics.vsLastMonth'), icon: TrendingUp, color: 'text-amber-500', bg: 'bg-amber-500/10' },
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
        {/* Bar chart */}
        <Card className="lg:col-span-2">
          <CardHeader className="flex flex-row items-center gap-3 pb-4">
            <div className="p-2 rounded-lg bg-primary/10 shrink-0">
              <BarChart3 className="w-4 h-4 text-primary" />
            </div>
            <div>
              <CardTitle>{t('analytics.monthlySignups')}</CardTitle>
              <CardDescription>{t('analytics.monthlySignupsDesc')}</CardDescription>
            </div>
          </CardHeader>
          <CardContent>
            <div className="flex items-end gap-3 h-52 px-2">
              {signups.map((val, i) => (
                <div key={i} className="flex-1 flex flex-col items-center gap-2">
                  <span className="text-xs font-semibold text-muted-foreground">{val}</span>
                  <div
                    className="w-full rounded-t-lg bg-primary/70 hover:bg-primary transition-colors cursor-pointer"
                    style={{ height: `${Math.max((val / maxSignup) * 100, 8)}%` }}
                  />
                  <span className="text-xs font-medium text-muted-foreground">{months[i]}</span>
                </div>
              ))}
            </div>
          </CardContent>
        </Card>

        {/* Role breakdown */}
        <Card>
          <CardHeader className="pb-4">
            <CardTitle>{t('analytics.roleDistribution')}</CardTitle>
            <CardDescription>{t('analytics.roleDistributionDesc')}</CardDescription>
          </CardHeader>
          <CardContent className="space-y-4">
            {loading ? (
              <div className="flex items-center justify-center py-8 text-sm text-muted-foreground">
                {t('common.loading')}
              </div>
            ) : (
              Object.entries(byRole).map(([role, count]) => (
                <div key={role} className="space-y-1.5">
                  <div className="flex items-center justify-between">
                    <div className="flex items-center gap-2">
                      <div className={cn('w-2.5 h-2.5 rounded-full', roleColors[role] ?? 'bg-slate-400')} />
                      <span className="text-sm font-medium capitalize">
                        {t(`users.roles.${role}`, { defaultValue: role })}
                      </span>
                    </div>
                    <div className="flex items-center gap-2">
                      <Badge variant="secondary">{count}</Badge>
                      <span className={cn('text-xs font-semibold w-8 text-right', roleTextColors[role] ?? 'text-slate-400')}>
                        {users.length > 0 ? Math.round((count / users.length) * 100) : 0}%
                      </span>
                    </div>
                  </div>
                  <div className="h-2 rounded-full bg-muted overflow-hidden">
                    <div
                      className={cn('h-full rounded-full transition-all', roleColors[role] ?? 'bg-slate-400')}
                      style={{ width: users.length > 0 ? `${(count / users.length) * 100}%` : '0%' }}
                    />
                  </div>
                </div>
              ))
            )}
          </CardContent>
        </Card>
      </div>
    </div>
  )
}
