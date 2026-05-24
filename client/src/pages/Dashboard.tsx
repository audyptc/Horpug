import { useTranslation } from 'react-i18next'
import { Users, UserCheck, UserX, TrendingUp, ArrowUpRight, ArrowDownRight } from 'lucide-react'
import { Card, CardContent, CardHeader, CardTitle, CardDescription } from '@/components/ui/card'
import { Avatar, AvatarFallback, AvatarImage } from '@/components/ui/avatar'
import { Badge } from '@/components/ui/badge'
import { mockUsers, mockActivities, mockStats } from '@/data/mockUsers'
import type { UserStatus } from '@/types'
import { cn } from '@/lib/utils'
import { formatDistanceToNow } from '../lib/dateUtils'

const activityColors = {
  create: 'bg-emerald-500',
  update: 'bg-blue-500',
  delete: 'bg-rose-500',
  login: 'bg-violet-500',
}

const statusVariant: Record<UserStatus, 'success' | 'secondary' | 'destructive'> = {
  active: 'success',
  inactive: 'secondary',
  suspended: 'destructive',
}

export function Dashboard() {
  const { t } = useTranslation()

  const statCards = [
    {
      title: t('dashboard.totalUsers'),
      value: mockStats.totalUsers,
      change: 12.5,
      icon: Users,
      color: 'text-violet-500',
      bg: 'bg-violet-500/10',
    },
    {
      title: t('dashboard.activeUsers'),
      value: mockStats.activeUsers,
      change: 8.2,
      icon: UserCheck,
      color: 'text-emerald-500',
      bg: 'bg-emerald-500/10',
    },
    {
      title: t('dashboard.newThisMonth'),
      value: mockStats.newUsersThisMonth,
      change: -3.1,
      icon: TrendingUp,
      color: 'text-blue-500',
      bg: 'bg-blue-500/10',
    },
    {
      title: t('dashboard.suspended'),
      value: mockStats.suspendedUsers,
      change: 0,
      icon: UserX,
      color: 'text-rose-500',
      bg: 'bg-rose-500/10',
    },
  ]

  return (
    <div className="space-y-6">
      <div>
        <h1 className="text-2xl font-bold tracking-tight">{t('dashboard.title')}</h1>
        <p className="text-muted-foreground text-sm mt-1">{t('dashboard.subtitle')}</p>
      </div>

      {/* Stats */}
      <div className="grid grid-cols-1 sm:grid-cols-2 xl:grid-cols-4 gap-4">
        {statCards.map((stat) => {
          const Icon = stat.icon
          const isPositive = stat.change > 0
          const isZero = stat.change === 0
          return (
            <Card key={stat.title} className="hover:shadow-md transition-shadow">
              <CardHeader className="flex flex-row items-center justify-between pb-2">
                <CardTitle className="text-sm font-medium text-muted-foreground">
                  {stat.title}
                </CardTitle>
                <div className={cn('p-2 rounded-lg', stat.bg)}>
                  <Icon className={cn('w-4 h-4', stat.color)} />
                </div>
              </CardHeader>
              <CardContent>
                <div className="text-3xl font-bold">{stat.value}</div>
                {!isZero && (
                  <p
                    className={cn(
                      'flex items-center gap-1 text-xs mt-1',
                      isPositive ? 'text-emerald-500' : 'text-rose-500'
                    )}
                  >
                    {isPositive ? (
                      <ArrowUpRight className="w-3 h-3" />
                    ) : (
                      <ArrowDownRight className="w-3 h-3" />
                    )}
                    {Math.abs(stat.change)}{t('dashboard.fromLastMonth')}
                  </p>
                )}
              </CardContent>
            </Card>
          )
        })}
      </div>

      <div className="grid grid-cols-1 xl:grid-cols-3 gap-6">
        {/* Recent users */}
        <Card className="xl:col-span-2">
          <CardHeader>
            <CardTitle>{t('dashboard.recentUsers')}</CardTitle>
            <CardDescription>{t('dashboard.recentUsersDesc')}</CardDescription>
          </CardHeader>
          <CardContent>
            <div className="space-y-4">
              {mockUsers.slice(0, 5).map((user) => (
                <div key={user.id} className="flex items-center gap-3">
                  <Avatar className="h-9 w-9 shrink-0">
                    <AvatarImage src={user.avatar} alt={user.name} />
                    <AvatarFallback className="text-xs">
                      {user.name.split(' ').map((n) => n[0]).join('')}
                    </AvatarFallback>
                  </Avatar>
                  <div className="flex-1 min-w-0">
                    <p className="text-sm font-medium truncate">{user.name}</p>
                    <p className="text-xs text-muted-foreground truncate">{user.email}</p>
                  </div>
                  <div className="flex items-center gap-2 shrink-0">
                    <Badge variant={statusVariant[user.status]}>
                      {user.status}
                    </Badge>
                    <span className="text-xs text-muted-foreground hidden sm:block capitalize">
                      {user.role}
                    </span>
                  </div>
                </div>
              ))}
            </div>
          </CardContent>
        </Card>

        {/* Activity feed */}
        <Card>
          <CardHeader>
            <CardTitle>{t('dashboard.recentActivity')}</CardTitle>
            <CardDescription>{t('dashboard.activityDesc')}</CardDescription>
          </CardHeader>
          <CardContent>
            <div className="space-y-4">
              {mockActivities.map((activity) => (
                <div key={activity.id} className="flex gap-3">
                  <div className="relative flex flex-col items-center">
                    <div
                      className={cn(
                        'w-2 h-2 rounded-full mt-1.5 shrink-0',
                        activityColors[activity.type]
                      )}
                    />
                    <div className="w-px bg-border flex-1 mt-1" />
                  </div>
                  <div className="pb-4 min-w-0">
                    <p className="text-sm leading-snug">
                      <span className="font-medium">{activity.user}</span>{' '}
                      <span className="text-muted-foreground">{activity.action}</span>{' '}
                      <span className="font-medium">{activity.target}</span>
                    </p>
                    <p className="text-xs text-muted-foreground mt-0.5">
                      {formatDistanceToNow(new Date(activity.time))}
                    </p>
                  </div>
                </div>
              ))}
            </div>
          </CardContent>
        </Card>
      </div>

      {/* Department breakdown */}
      <Card>
        <CardHeader>
          <CardTitle>{t('dashboard.deptBreakdown')}</CardTitle>
          <CardDescription>{t('dashboard.deptBreakdownDesc')}</CardDescription>
        </CardHeader>
        <CardContent>
          <div className="grid grid-cols-2 sm:grid-cols-4 gap-4">
            {Object.entries(
              mockUsers.reduce<Record<string, number>>((acc, u) => {
                acc[u.department] = (acc[u.department] ?? 0) + 1
                return acc
              }, {})
            ).map(([dept, count]) => (
              <div
                key={dept}
                className="flex flex-col gap-1 p-4 rounded-lg bg-muted/50 hover:bg-muted transition-colors"
              >
                <span className="text-2xl font-bold">{count}</span>
                <span className="text-sm text-muted-foreground">{dept}</span>
                <div className="mt-2 h-1.5 rounded-full bg-border overflow-hidden">
                  <div
                    className="h-full bg-primary rounded-full"
                    style={{ width: `${(count / mockUsers.length) * 100}%` }}
                  />
                </div>
              </div>
            ))}
          </div>
        </CardContent>
      </Card>
    </div>
  )
}
