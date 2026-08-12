import { useEffect, useState } from 'react'
import { api, extractErrorMessage } from '@/shared/api/client'
import { useLanguage, type TranslationKey } from '@/shared/i18n/language'
import { Badge } from '@/shared/components/ui/badge'
import { Card, CardContent, CardDescription, CardHeader, CardTitle } from '@/shared/components/ui/card'
import { Table, TableBody, TableCell, TableRow } from '@/shared/components/ui/table'

type ApiRecord = Record<string, unknown>

type Props = {
  titleKey: TranslationKey
  descriptionKey: TranslationKey
  endpoint: string
}

function recordLabel(record: ApiRecord): string {
  const label = record.name ?? record.username ?? record.action ?? record.id
  return typeof label === 'string' ? label : String(label ?? '')
}

export default function ResourcePage({ titleKey, descriptionKey, endpoint }: Props) {
  const { t, language } = useLanguage()
  const [resourceState, setResourceState] = useState<{
    endpoint: string
    records: ApiRecord[] | null
    error: string | null
  }>(() => ({
    endpoint,
    records: null,
    error: null,
  }))

  const records = resourceState.endpoint === endpoint ? resourceState.records : null
  const error = resourceState.endpoint === endpoint ? resourceState.error : null

  useEffect(() => {
    let cancelled = false

    api
      .get<ApiRecord[]>(endpoint)
      .then(({ data }) => {
        if (!cancelled) {
          setResourceState({
            endpoint,
            records: data,
            error: null,
          })
        }
      })
      .catch((err) => {
        if (!cancelled) {
          setResourceState({
            endpoint,
            records: null,
            error: extractErrorMessage(err, t('resourceLoadError')),
          })
        }
      })

    return () => {
      cancelled = true
    }
  }, [endpoint, t])

  return (
    <main className="content">
      <section className="welcome">
        <h1>{t(titleKey)}</h1>
        <p>{t(descriptionKey)}</p>
      </section>

      <Card>
        <CardHeader>
          <CardTitle>{t('liveApiData')}</CardTitle>
          <CardDescription>
            GET {endpoint} · {records ? `${records.length} ${t('recordsLoaded')}` : t('loading')}
          </CardDescription>
        </CardHeader>
        <CardContent>
          {error && <p className="resource-error">{error}</p>}
          {!error && records && records.length === 0 && <p className="metric-detail">{t('noRecords')}</p>}
          {!error && records && records.length > 0 && (
            <div className="table-wrap">
              <Table>
                <TableBody>
                  {records.slice(0, 20).map((record, index) => (
                    <TableRow key={String(record.id ?? index)}>
                      <TableCell className="font-semibold">{recordLabel(record)}</TableCell>
                      <TableCell>
                        {typeof record.is_active === 'boolean' && (
                          <Badge variant={record.is_active ? 'default' : 'outline'}>
                            {record.is_active ? t('statusActive') : t('statusInactive')}
                          </Badge>
                        )}
                      </TableCell>
                      <TableCell className="text-right text-muted-foreground">
                        {typeof record.created_at === 'string'
                          ? new Date(record.created_at).toLocaleString(
                              language === 'th' ? 'th-TH' : 'en-US'
                            )
                          : ''}
                      </TableCell>
                    </TableRow>
                  ))}
                </TableBody>
              </Table>
            </div>
          )}
        </CardContent>
      </Card>
    </main>
  )
}
