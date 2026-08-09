import { useState, useEffect, useMemo } from 'react'
import { useTranslation } from 'react-i18next'
import { Link } from 'react-router-dom'
import {
  Home, Users, FileText, AlertCircle, CheckCircle2,
  TrendingUp, DoorOpen, Wrench, Clock, Megaphone, Pin,
} from 'lucide-react'
import { Card, CardContent, CardHeader, CardTitle, CardDescription } from '@/components/ui/card'
import { Badge } from '@/components/ui/badge'
import { Button } from '@/components/ui/button'
import {
  Dialog,
  DialogContent,
  DialogHeader,
  DialogTitle,
  DialogDescription,
} from '@/components/ui/dialog'
import { dashboardService } from '@/features/dashboard/dashboardService'
import { announcementService } from '@/features/announcements/announcementService'
import type { ApiDashboardSummary, ApiAnnouncement } from '@/types'
import { cn } from '@/lib/utils'
import { formatDate } from '@/lib/dateUtils'

function formatBaht(amount: number) {
  return new Intl.NumberFormat('th-TH', { minimumFractionDigits: 0, maximumFractionDigits: 0 }).format(amount)
}

const TYPE_ACCENT: Record<string, string> = {
  general: 'border-l-slate-400',
  maintenance: 'border-l-amber-400',
  payment: 'border-l-blue-400',
  emergency: 'border-l-red-500',
}

const TYPE_BADGE: Record<string, 'secondary' | 'outline' | 'default' | 'destructive'> = {
  general: 'secondary',
  maintenance: 'outline',
  payment: 'default',
  emergency: 'destructive',
}

export function Dashboard() {
  const { t } = useTranslation()
  const [summary, setSummary] = useState<ApiDashboardSummary | null>(null)
  const [loading, setLoading] = useState(true)
  const [announcements, setAnnouncements] = useState<ApiAnnouncement[]>([])
  const [announcementsLoading, setAnnouncementsLoading] = useState(true)
  const [selectedAnnouncement, setSelectedAnnouncement] = useState<ApiAnnouncement | null>(null)

  useEffect(() => {
    dashboardService.summary()
      .then(setSummary)
      .catch(() => {})
      .finally(() => setLoading(false))
    announcementService.list(1, 50)
      .then((r) => setAnnouncements(r.data))
      .catch(() => {})
      .finally(() => setAnnouncementsLoading(false))
  }, [])

  const activeAnnouncements = useMemo(() => {
    const now = new Date()
    return announcements
      .filter((a) => {
        const published = new Date(a.published_at)
        const expired = a.expired_at ? new Date(a.expired_at) : null
        return published <= now && (expired === null || expired > now)
      })
      .sort((a, b) => {
        if (a.is_pinned !== b.is_pinned) return a.is_pinned ? -1 : 1
        return new Date(b.published_at).getTime() - new Date(a.published_at).getTime()
      })
  }, [announcements])

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

      {/* Announcements */}
      <Card>
        <CardHeader className="flex flex-row items-center justify-between pb-3">
          <div className="flex items-center gap-2">
            <Megaphone className="w-4 h-4 text-muted-foreground" />
            <div>
              <CardTitle>{t('dashboard.latestAnnouncements')}</CardTitle>
              <CardDescription className="mt-0.5">{t('dashboard.latestAnnouncementsDesc')}</CardDescription>
            </div>
          </div>
          <Link to="/announcements" className="text-xs text-primary hover:underline shrink-0">
            {t('dashboard.viewAll')}
          </Link>
        </CardHeader>
        <CardContent>
          {announcementsLoading ? (
            <p className="text-sm text-muted-foreground py-4 text-center">{t('common.loading')}</p>
          ) : activeAnnouncements.length === 0 ? (
            <p className="text-sm text-muted-foreground py-6 text-center">{t('dashboard.noActiveAnnouncements')}</p>
          ) : (
            <div className="space-y-2">
              {activeAnnouncements.map((ann) => (
                <button
                  key={ann.id}
                  type="button"
                  onClick={() => setSelectedAnnouncement(ann)}
                  className={cn('w-full text-left p-3 rounded-lg border border-l-4 bg-card hover:bg-muted/40 transition-colors', TYPE_ACCENT[ann.type])}
                >
                  <div className="flex items-start justify-between gap-3">
                    <div className="flex-1 min-w-0">
                      <div className="flex items-center gap-1.5 mb-1 flex-wrap">
                        {ann.is_pinned && <Pin className="w-3 h-3 text-muted-foreground shrink-0" />}
                        <Badge variant={TYPE_BADGE[ann.type]} className="text-xs">
                          {t(`announcements.types.${ann.type}`)}
                        </Badge>
                        <span className="font-medium text-sm">{ann.title}</span>
                      </div>
                      <p className="text-xs text-muted-foreground line-clamp-2">{ann.content}</p>
                    </div>
                    <span className="text-xs text-muted-foreground shrink-0 mt-0.5">
                      {formatDate(ann.published_at)}
                    </span>
                  </div>
                </button>
              ))}
            </div>
          )}
        </CardContent>
      </Card>

      <Dialog open={!!selectedAnnouncement} onOpenChange={(open) => { if (!open) setSelectedAnnouncement(null) }}>
        <DialogContent className="max-w-lg">
          {selectedAnnouncement && (
            <>
              <DialogHeader>
                <div className="flex items-center gap-2 flex-wrap mb-1">
                  {selectedAnnouncement.is_pinned && <Pin className="w-3.5 h-3.5 text-muted-foreground" />}
                  <Badge variant={TYPE_BADGE[selectedAnnouncement.type]}>
                    {t(`announcements.types.${selectedAnnouncement.type}`)}
                  </Badge>
                </div>
                <DialogTitle>{selectedAnnouncement.title}</DialogTitle>
                <DialogDescription className="flex gap-3 text-xs">
                  <span>{t('announcements.publishedAt')}: {formatDate(selectedAnnouncement.published_at)}</span>
                  {selectedAnnouncement.expired_at && (
                    <span>{t('announcements.expiredAt')}: {formatDate(selectedAnnouncement.expired_at)}</span>
                  )}
                </DialogDescription>
              </DialogHeader>
              <p className="text-sm whitespace-pre-wrap leading-relaxed">{selectedAnnouncement.content}</p>
              <div className="flex justify-end pt-2">
                <Button variant="outline" onClick={() => setSelectedAnnouncement(null)}>
                  {t('common.cancel')}
                </Button>
              </div>
            </>
          )}
        </DialogContent>
      </Dialog>
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
      <Badge variant={variant}>
        {value} {t('dashboard.items')}
      </Badge>
    </div>
  )
}
