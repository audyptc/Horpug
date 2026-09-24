import { useEffect, useState } from 'react'
import axios from 'axios'
import { Download } from 'lucide-react'
import { api, extractErrorMessage } from '@/shared/api/client'
import { useLanguage, type TranslationKey } from '@/shared/i18n/language'
import { Button } from '@/shared/components/ui/button'
import { Card, CardContent, CardDescription, CardHeader, CardTitle } from '@/shared/components/ui/card'
import { Table, TableBody, TableCell, TableHead, TableHeader, TableRow } from '@/shared/components/ui/table'
import type { ApiDormitory } from '@/features/dormitory/types'
import { DormitorySearchSelect } from '@/features/dormitory/components/DormitorySearchSelect'
import { MonthPickerField } from '@/features/invoice/components/InvoiceFormSheet'
import { parsePeriodInputValue, toPeriodInputValue } from '@/features/invoice/utils'
import type { ApiMonthlyReport, ReportAmount } from './types'

const methodLabelKeys: Record<string, TranslationKey> = {
  cash: 'paymentMethodCash',
  transfer: 'paymentMethodTransfer',
  credit_card: 'paymentMethodCreditCard',
  other: 'paymentMethodOther',
  deposit: 'paymentMethodDeposit',
}

const itemTypeLabelKeys: Record<string, TranslationKey> = {
  rent: 'invoiceItemTypeRent',
  electricity: 'invoiceItemTypeElectricity',
  water: 'invoiceItemTypeWater',
  other: 'invoiceItemTypeOther',
}

const categoryLabelKeys: Record<string, TranslationKey> = {
  maintenance: 'reportCategoryMaintenance',
  utility: 'reportCategoryUtility',
  salary: 'reportCategorySalary',
  supplies: 'reportCategorySupplies',
  other: 'reportCategoryOther',
}

function formatMoney(value: number): string {
  return value.toLocaleString('th-TH', { minimumFractionDigits: 2, maximumFractionDigits: 2 })
}

type BreakdownProps = {
  title: string
  rows: ReportAmount[]
  labelKeys: Record<string, TranslationKey>
  emptyLabel: string
}

function Breakdown({ title, rows, labelKeys, emptyLabel }: BreakdownProps) {
  const { t } = useLanguage()
  return (
    <Card>
      <CardHeader>
        <CardTitle className="text-base">{title}</CardTitle>
      </CardHeader>
      <CardContent>
        {rows.length === 0 ? (
          <p className="metric-detail">{emptyLabel}</p>
        ) : (
          <ul className="divide-y divide-border text-sm">
            {rows.map((row) => (
              <li key={row.key} className="flex justify-between gap-2 py-2">
                <span>{labelKeys[row.key] ? t(labelKeys[row.key]) : row.key}</span>
                <span className="tabular-nums">{formatMoney(row.amount)}</span>
              </li>
            ))}
          </ul>
        )}
      </CardContent>
    </Card>
  )
}

// Monthly income / billing / expense report over the dormitories the user
// manages, with an Excel download of the same month.
export default function ReportPage() {
  const { t } = useLanguage()
  const [period, setPeriod] = useState(() => {
    const now = new Date()
    return toPeriodInputValue(now.getFullYear(), now.getMonth() + 1)
  })
  const [dormitory, setDormitory] = useState<ApiDormitory | null>(null)
  // Stored with the request it answers, so a stale report reads as loading.
  const [report, setReport] = useState<{ key: string; data: ApiMonthlyReport } | null>(null)
  const [loadError, setLoadError] = useState<{ key: string; message: string } | null>(null)
  const [exporting, setExporting] = useState(false)
  const [exportError, setExportError] = useState<string | null>(null)

  const parsed = parsePeriodInputValue(period)
  const params = parsed
    ? { year: parsed.year, month: parsed.month, dormitory_id: dormitory?.id || undefined }
    : null
  const key = params ? JSON.stringify(params) : ''

  useEffect(() => {
    if (!key || !params) return
    const controller = new AbortController()
    api
      .get<ApiMonthlyReport>('/reports/monthly', { params, signal: controller.signal })
      .then(({ data }) => setReport({ key, data }))
      .catch((err) => {
        if (axios.isCancel(err)) return
        setLoadError({ key, message: extractErrorMessage(err, t('resourceLoadError')) })
      })
    return () => controller.abort()
    // eslint-disable-next-line react-hooks/exhaustive-deps
  }, [key])

  const data = report && report.key === key ? report.data : null
  const error = loadError && loadError.key === key ? loadError.message : null

  async function handleExport() {
    if (!params) return
    setExporting(true)
    setExportError(null)
    try {
      const { data: blob } = await api.get<Blob>('/reports/monthly/export', { params, responseType: 'blob' })
      const url = URL.createObjectURL(blob)
      const link = document.createElement('a')
      link.href = url
      link.download = `report-${params.year}-${String(params.month).padStart(2, '0')}.xlsx`
      document.body.appendChild(link)
      link.click()
      link.remove()
      // Revoked on the next tick: revoking at once can cancel the download.
      window.setTimeout(() => URL.revokeObjectURL(url), 0)
    } catch (err) {
      setExportError(extractErrorMessage(err, t('reportExportError')))
    } finally {
      setExporting(false)
    }
  }

  const tiles = data
    ? [
        {
          label: t('reportIncome'),
          value: formatMoney(data.income.total),
          detail: t('reportIncomeDetail').replace('{count}', String(data.income.count)),
        },
        {
          label: t('reportExpenses'),
          value: formatMoney(data.expenses.total),
          detail: t('reportExpensesDetail'),
        },
        {
          label: t('reportNet'),
          value: formatMoney(data.net),
          detail: data.net < 0 ? t('reportNetNegative') : t('reportNetDetail'),
        },
        {
          label: t('reportBilled'),
          value: formatMoney(data.billing.billed),
          detail: t('reportBilledDetail')
            .replace('{count}', String(data.billing.invoice_count))
            .replace('{collected}', formatMoney(data.billing.collected))
            .replace('{outstanding}', formatMoney(data.billing.outstanding)),
        },
        {
          label: t('reportArrears'),
          value: formatMoney(data.arrears.amount),
          detail: t('reportArrearsDetail').replace('{count}', String(data.arrears.count)),
        },
        {
          label: t('reportOccupancy'),
          value: `${data.occupancy.occupied} / ${data.occupancy.rooms}`,
          detail: t('reportOccupancyDetail'),
        },
      ]
    : []

  return (
    <main className="content">
      <section className="welcome">
        <h1>{t('menuReports')}</h1>
        <p>{t('menuReportsDescription')}</p>
      </section>

      <Card>
        <CardContent className="flex flex-col gap-3 pt-6 sm:flex-row sm:items-end">
          <label className="flex flex-col gap-1.5 text-sm font-medium sm:w-56">
            {t('reportPeriodLabel')}
            <MonthPickerField value={period} onChange={setPeriod} placeholder={t('reportPeriodLabel')} />
          </label>
          <label className="flex flex-1 flex-col gap-1.5 text-sm font-medium">
            {t('reportDormitoryLabel')}
            <DormitorySearchSelect
              selectedLabel={dormitory?.name ?? ''}
              onSelectDormitory={setDormitory}
              placeholder={t('reportAllDormitories')}
              searchPlaceholder={t('invoiceGenerateDormitoryPlaceholder')}
              noResultsLabel={t('invoiceGenerateDormitoryNoResults')}
              clearLabel={t('reportAllDormitories')}
              onClear={() => setDormitory(null)}
            />
          </label>
          <Button onClick={handleExport} disabled={!params || exporting}>
            <Download />
            {exporting ? t('reportExporting') : t('reportExport')}
          </Button>
        </CardContent>
        {exportError && <p className="resource-error px-6 pb-4">{exportError}</p>}
      </Card>

      {error ? (
        <p className="resource-error">{error}</p>
      ) : !data ? (
        <p className="metric-detail">{t('loading')}</p>
      ) : (
        <>
          <div className="metric-grid">
            {tiles.map((tile) => (
              <Card key={tile.label}>
                <CardHeader>
                  <CardDescription>{tile.label}</CardDescription>
                  <CardTitle className="tabular-nums">{tile.value}</CardTitle>
                </CardHeader>
                <CardContent>
                  <p className="metric-detail">{tile.detail}</p>
                </CardContent>
              </Card>
            ))}
          </div>

          <div className="grid gap-4 lg:grid-cols-3">
            <Breakdown
              title={t('reportIncomeByMethod')}
              rows={data.income.by_method}
              labelKeys={methodLabelKeys}
              emptyLabel={t('reportNoData')}
            />
            <Breakdown
              title={t('reportBilledByType')}
              rows={data.billing.by_item_type}
              labelKeys={itemTypeLabelKeys}
              emptyLabel={t('reportNoData')}
            />
            <Breakdown
              title={t('reportExpensesByCategory')}
              rows={data.expenses.by_category}
              labelKeys={categoryLabelKeys}
              emptyLabel={t('reportNoData')}
            />
          </div>

          {data.dormitories.length > 1 && (
            <Card>
              <CardHeader>
                <CardTitle className="text-base">{t('reportByDormitory')}</CardTitle>
              </CardHeader>
              <CardContent>
                <div className="overflow-x-auto">
                  <Table>
                    <TableHeader>
                      <TableRow>
                        <TableHead>{t('reportDormitoryLabel')}</TableHead>
                        <TableHead className="text-right">{t('reportIncome')}</TableHead>
                        <TableHead className="text-right">{t('reportExpenses')}</TableHead>
                        <TableHead className="text-right">{t('reportNet')}</TableHead>
                        <TableHead className="text-right">{t('reportBilled')}</TableHead>
                        <TableHead className="text-right">{t('reportOutstanding')}</TableHead>
                      </TableRow>
                    </TableHeader>
                    <TableBody>
                      {data.dormitories.map((row) => (
                        <TableRow key={row.id}>
                          <TableCell className="font-medium">{row.name}</TableCell>
                          <TableCell className="text-right tabular-nums">{formatMoney(row.income)}</TableCell>
                          <TableCell className="text-right tabular-nums">{formatMoney(row.expenses)}</TableCell>
                          <TableCell className="text-right tabular-nums">{formatMoney(row.net)}</TableCell>
                          <TableCell className="text-right tabular-nums">{formatMoney(row.billed)}</TableCell>
                          <TableCell className="text-right tabular-nums">{formatMoney(row.outstanding)}</TableCell>
                        </TableRow>
                      ))}
                    </TableBody>
                  </Table>
                </div>
              </CardContent>
            </Card>
          )}
        </>
      )}
    </main>
  )
}
