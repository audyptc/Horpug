import { useMemo } from 'react'
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

export function DashboardPage() {
  const { t, language } = useLanguage()

  const metrics = useMemo(
    () => [
      { title: t('totalViews'), value: '1,284,920', detail: t('totalViewsDetail') },
      { title: t('subscribers'), value: '98,432', detail: t('subscribersDetail') },
      { title: t('watchTime'), value: '3,920 hrs', detail: t('watchTimeDetail') },
    ],
    // eslint-disable-next-line react-hooks/exhaustive-deps
    [language]
  )

  const recentUploads = useMemo(
    () => [
      {
        title: 'Campus Tour 2026',
        status: t('statusPublished'),
        reach: '24.8K',
        updated: t('timeHoursAgo'),
      },
      {
        title: 'Student Interview #12',
        status: t('statusProcessing'),
        reach: t('reachPending'),
        updated: t('timeMinutesAgo'),
      },
      {
        title: 'Welcome Freshmen Highlights',
        status: t('statusDraft'),
        reach: t('reachNotPublished'),
        updated: t('timeYesterday'),
      },
    ],
    // eslint-disable-next-line react-hooks/exhaustive-deps
    [language]
  )

  return (
    <main className="content">
      <section className="welcome">
        <h1>{t('dashboard')}</h1>
        <p>{t('dashboardOverview')}</p>
      </section>

      <section className="metric-grid">
        {metrics.map((metric) => (
          <Card key={metric.title}>
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
          <CardTitle>{t('recentUploadQueue')}</CardTitle>
          <CardDescription>{t('recentUploadQueueDescription')}</CardDescription>
        </CardHeader>
        <CardContent>
          <div className="table-wrap">
            <Table>
              <TableHeader>
                <TableRow>
                  <TableHead>{t('tableVideo')}</TableHead>
                  <TableHead>{t('tableStatus')}</TableHead>
                  <TableHead>{t('tableReach')}</TableHead>
                  <TableHead className="text-right">{t('tableLastUpdate')}</TableHead>
                </TableRow>
              </TableHeader>
              <TableBody>
                {recentUploads.map((upload) => (
                  <TableRow key={upload.title}>
                    <TableCell className="font-semibold">{upload.title}</TableCell>
                    <TableCell>
                      <Badge
                        variant={
                          upload.status === t('statusPublished')
                            ? 'default'
                            : upload.status === t('statusProcessing')
                              ? 'secondary'
                              : 'outline'
                        }
                      >
                        {upload.status}
                      </Badge>
                    </TableCell>
                    <TableCell>{upload.reach}</TableCell>
                    <TableCell className="text-right text-muted-foreground">{upload.updated}</TableCell>
                  </TableRow>
                ))}
              </TableBody>
            </Table>
          </div>
        </CardContent>
      </Card>
    </main>
  )
}
