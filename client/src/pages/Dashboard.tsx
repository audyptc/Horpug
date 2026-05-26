import { useState, useEffect, useMemo } from 'react'
import { useTranslation } from 'react-i18next'
import { Users, UserCheck, UserX, TrendingUp, ArrowUpRight, ArrowDownRight } from 'lucide-react'
import { Card, CardContent, CardHeader, CardTitle, CardDescription } from '@/components/ui/card'
import { Avatar, AvatarFallback, AvatarImage } from '@/components/ui/avatar'
import { Badge } from '@/components/ui/badge'
import { userService } from '@/services/userService'
import type { ApiUser } from '@/types/api'
import type { Activity } from '@/types'
import { cn } from '@/lib/utils'
import { formatDistanceToNow } from '../lib/dateUtils'

const staticActivities: Activity[] = [
  { id: '1', user: 'Suphat Siapth', action: 'created user', target: 'Siriporn Mahachai', time: '2026-05-24T08:30:00Z', type: 'create' },
  { id: '2', user: 'Arisa Wongsak', action: 'updated role of', target: 'Wirat Burana', time: '2026-05-23T14:12:00Z', type: 'update' },
  { id: '3', user: 'Suphat Siapth', action: 'suspended', target: 'Natthida Chumphon', time: '2026-05-22T11:00:00Z', type: 'update' },
  { id: '4', user: 'Krit Supakon', action: 'logged in from', target: 'Bangkok, Thailand', time: '2026-05-24T07:00:00Z', type: 'login' },
  { id: '5', user: 'Pimchanok Lertsiri', action: 'deleted record', target: 'Draft Campaign Q1', time: '2026-04-15T11:20:00Z', type: 'delete' },
]

const activityColors = {
  create: 'bg-emerald-500',
  update: 'bg-blue-500',
  delete: 'bg-rose-500',
  login: 'bg-violet-500',
}

function getAvatar(name: string) {
  return `https://api.dicebear.com/9.x/avataaars/svg?seed=${encodeURIComponent(name)}`
}

export function Dashboard() {
  const { t } = useTranslation()
  const [users, setUsers] = useState<ApiUser[]>([])
  const [total, setTotal] = useState(0)
  const [loading, setLoading] = useState(true)

  useEffect(() => {
    userService.list(1, 9999)
      .then((resp) => {
        setUsers(resp.data)
        setTotal(resp.meta.total)
      })
      .catch(() => {})
      .finally(() => setLoading(false))
  }, [])

  const activeCount = useMemo(() => users.filter((u) => u.is_active).length, [users])
  const inactiveCount = useMemo(() => users.filter((u) => !u.is_active).length, [users])
  const newThisMonth = useMemo(() => {
    const now = new Date()
    return users.filter((u) => {
      const d = new Date(u.created_at)
      return d.getMonth() === now.getMonth() && d.getFullYear() === now.getFullYear()
    }).length
  }, [users])

  const recentUsers = useMemo(
    () => [...users].sort((a, b) => new Date(b.created_at).getTime() - new Date(a.created_at).getTime()).slice(0, 5),
    [users]
  )

  const statCards = [
    { title: t('dashboard.totalUsers'), value: total, change: 12.5, icon: Users, color: 'text-violet-500', bg: 'bg-violet-500/10' },
    { title: t('dashboard.activeUsers'), value: activeCount, change: 8.2, icon: UserCheck, color: 'text-emerald-500', bg: 'bg-emerald-500/10' },
    { title: t('dashboard.newThisMonth'), value: newThisMonth, change: -3.1, icon: TrendingUp, color: 'text-blue-500', bg: 'bg-blue-500/10' },
    { title: t('dashboard.inactiveUsers'), value: inactiveCount, change: 0, icon: UserX, color: 'text-rose-500', bg: 'bg-rose-500/10' },
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
            <Card key={stat.title} className="hover:shadow-md transition-shadow overflow-hidden">
              <div className={cn('h-1 w-full', stat.bg.replace('/10', ''))} />
              <CardHeader className="flex flex-row items-center justify-between pb-2">
                <CardTitle className="text-sm font-medium text-muted-foreground">{stat.title}</CardTitle>
                <div className={cn('p-2 rounded-lg', stat.bg)}>
                  <Icon className={cn('w-4 h-4', stat.color)} />
                </div>
              </CardHeader>
              <CardContent>
                <div className="text-3xl font-bold">
                  {loading ? <span className="text-muted-foreground text-lg">—</span> : stat.value}
                </div>
                {!isZero && (
                  <p className={cn('flex items-center gap-1 text-xs mt-1', isPositive ? 'text-emerald-500' : 'text-rose-500')}>
                    {isPositive ? <ArrowUpRight className="w-3 h-3" /> : <ArrowDownRight className="w-3 h-3" />}
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
            {loading ? (
              <div className="flex items-center justify-center py-8 text-sm text-muted-foreground">
                {t('common.loading')}
              </div>
            ) : (
              <div className="space-y-4">
                {recentUsers.map((user) => (
                  <div key={user.id} className="flex items-center gap-3">
                    <Avatar className="h-9 w-9 shrink-0">
                      <AvatarImage src={getAvatar(user.full_name)} alt={user.full_name} />
                      <AvatarFallback className="text-xs">
                        {user.full_name.split(' ').map((n) => n[0]).join('')}
                      </AvatarFallback>
                    </Avatar>
                    <div className="flex-1 min-w-0">
                      <p className="text-sm font-medium truncate">{user.full_name}</p>
                      <p className="text-xs text-muted-foreground truncate">{user.email}</p>
                    </div>
                    <div className="flex items-center gap-2 shrink-0">
                      <Badge variant={user.is_active ? 'success' : 'secondary'}>
                        {user.is_active ? t('users.statuses.active') : t('users.statuses.inactive')}
                      </Badge>
                      {user.role && (
                        <span className="text-xs text-muted-foreground hidden sm:block capitalize">
                          {user.role.name}
                        </span>
                      )}
                    </div>
                  </div>
                ))}
              </div>
            )}
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
              {staticActivities.map((activity) => (
                <div key={activity.id} className="flex gap-3">
                  <div className="relative flex flex-col items-center">
                    <div className={cn('w-2 h-2 rounded-full mt-1.5 shrink-0', activityColors[activity.type])} />
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
    </div>
  )
}
