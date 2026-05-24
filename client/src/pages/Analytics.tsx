import { useTranslation } from 'react-i18next'
import { Card, CardContent, CardHeader, CardTitle, CardDescription } from '@/components/ui/card'
import { Badge } from '@/components/ui/badge'
import { mockUsers } from '@/data/mockUsers'

const months = ['Jan', 'Feb', 'Mar', 'Apr', 'May', 'Jun']
const signups = [3, 5, 2, 7, 4, 8]
const maxSignup = Math.max(...signups)

export function Analytics() {
  const { t } = useTranslation()

  const byRole = mockUsers.reduce<Record<string, number>>((acc, u) => {
    acc[u.role] = (acc[u.role] ?? 0) + 1
    return acc
  }, {})

  const roleColors: Record<string, string> = {
    admin: 'bg-violet-500',
    manager: 'bg-blue-500',
    editor: 'bg-amber-500',
    viewer: 'bg-slate-400',
  }

  const summaryCards = [
    { label: t('analytics.avgSession'), value: '24m 18s', sub: t('analytics.perUser') },
    { label: t('analytics.pageViews'), value: '1,284', sub: t('analytics.thisMonth') },
    { label: t('analytics.retention'), value: '76%', sub: t('analytics.monthlyActive') },
    { label: t('analytics.growthRate'), value: '+12.5%', sub: t('analytics.vsLastMonth') },
  ]

  return (
    <div className="space-y-6">
      <div>
        <h1 className="text-2xl font-bold tracking-tight">{t('analytics.title')}</h1>
        <p className="text-muted-foreground text-sm mt-1">{t('analytics.subtitle')}</p>
      </div>

      <div className="grid grid-cols-1 lg:grid-cols-3 gap-6">
        {/* Bar chart */}
        <Card className="lg:col-span-2">
          <CardHeader>
            <CardTitle>{t('analytics.monthlySignups')}</CardTitle>
            <CardDescription>{t('analytics.monthlySignupsDesc')}</CardDescription>
          </CardHeader>
          <CardContent>
            <div className="flex items-end gap-3 h-48">
              {signups.map((val, i) => (
                <div key={i} className="flex-1 flex flex-col items-center gap-2">
                  <span className="text-xs font-medium text-muted-foreground">{val}</span>
                  <div
                    className="w-full rounded-t-md bg-primary/80 hover:bg-primary transition-colors"
                    style={{ height: `${(val / maxSignup) * 100}%` }}
                  />
                  <span className="text-xs text-muted-foreground">{months[i]}</span>
                </div>
              ))}
            </div>
          </CardContent>
        </Card>

        {/* Role breakdown */}
        <Card>
          <CardHeader>
            <CardTitle>{t('analytics.roleDistribution')}</CardTitle>
            <CardDescription>{t('analytics.roleDistributionDesc')}</CardDescription>
          </CardHeader>
          <CardContent className="space-y-4">
            {Object.entries(byRole).map(([role, count]) => (
              <div key={role} className="space-y-1.5">
                <div className="flex items-center justify-between">
                  <div className="flex items-center gap-2">
                    <div className={`w-2.5 h-2.5 rounded-full ${roleColors[role]}`} />
                    <span className="text-sm">{t(`users.roles.${role}`)}</span>
                  </div>
                  <div className="flex items-center gap-2">
                    <Badge variant="secondary">{count}</Badge>
                    <span className="text-xs text-muted-foreground w-8 text-right">
                      {Math.round((count / mockUsers.length) * 100)}%
                    </span>
                  </div>
                </div>
                <div className="h-1.5 rounded-full bg-muted overflow-hidden">
                  <div
                    className={`h-full rounded-full ${roleColors[role]}`}
                    style={{ width: `${(count / mockUsers.length) * 100}%` }}
                  />
                </div>
              </div>
            ))}
          </CardContent>
        </Card>
      </div>

      {/* Summary cards */}
      <div className="grid grid-cols-2 md:grid-cols-4 gap-4">
        {summaryCards.map((item) => (
          <Card key={item.label}>
            <CardContent className="pt-6">
              <p className="text-2xl font-bold">{item.value}</p>
              <p className="text-sm font-medium mt-1">{item.label}</p>
              <p className="text-xs text-muted-foreground">{item.sub}</p>
            </CardContent>
          </Card>
        ))}
      </div>
    </div>
  )
}
