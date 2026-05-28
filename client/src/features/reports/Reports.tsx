import { useState, useEffect, useCallback } from 'react'
import { useTranslation } from 'react-i18next'
import { Card, CardContent, CardHeader, CardTitle, CardDescription } from '@/components/ui/card'
import { reportService } from '@/features/reports/reportService'
import type { ApiIncomeReport, ApiExpenseReport, ApiOccupancyReport } from '@/types'
import { Banknote, TrendingDown, BedDouble, FileBarChart2 } from 'lucide-react'
import { cn } from '@/lib/utils'

type Tab = 'income' | 'expenses' | 'occupancy'

function formatBaht(v: number) {
  return new Intl.NumberFormat('th-TH', { minimumFractionDigits: 0, maximumFractionDigits: 0 }).format(v)
}

function currentYearMonth() {
  const now = new Date()
  return `${now.getFullYear()}-${String(now.getMonth() + 1).padStart(2, '0')}`
}

function sixMonthsAgo() {
  const d = new Date()
  d.setMonth(d.getMonth() - 5)
  return `${d.getFullYear()}-${String(d.getMonth() + 1).padStart(2, '0')}`
}

const CATEGORY_COLORS: Record<string, string> = {
  repair: 'bg-orange-500',
  utilities: 'bg-blue-500',
  supplies: 'bg-violet-500',
  salary: 'bg-emerald-500',
  other: 'bg-gray-400',
}

export function Reports() {
  const { t } = useTranslation()
  const [tab, setTab] = useState<Tab>('income')
  const [from, setFrom] = useState(sixMonthsAgo())
  const [to, setTo] = useState(currentYearMonth())

  const [income, setIncome] = useState<ApiIncomeReport | null>(null)
  const [expenses, setExpenses] = useState<ApiExpenseReport | null>(null)
  const [occupancy, setOccupancy] = useState<ApiOccupancyReport | null>(null)
  const [loading, setLoading] = useState(false)

  const fetchIncome = useCallback(() => {
    setLoading(true)
    reportService.income(from, to)
      .then(setIncome)
      .catch(() => {})
      .finally(() => setLoading(false))
  }, [from, to])

  const fetchExpenses = useCallback(() => {
    setLoading(true)
    reportService.expenses(from, to)
      .then(setExpenses)
      .catch(() => {})
      .finally(() => setLoading(false))
  }, [from, to])

  const fetchOccupancy = useCallback(() => {
    setLoading(true)
    reportService.occupancy()
      .then(setOccupancy)
      .catch(() => {})
      .finally(() => setLoading(false))
  }, [])

  useEffect(() => {
    if (tab === 'income') fetchIncome()
    else if (tab === 'expenses') fetchExpenses()
    else fetchOccupancy()
  }, [tab, fetchIncome, fetchExpenses, fetchOccupancy])

  const tabs: { key: Tab; label: string; icon: React.ElementType }[] = [
    { key: 'income', label: t('reports.tabIncome'), icon: Banknote },
    { key: 'expenses', label: t('reports.tabExpenses'), icon: TrendingDown },
    { key: 'occupancy', label: t('reports.tabOccupancy'), icon: BedDouble },
  ]

  return (
    <div className="space-y-6">
      <div>
        <h1 className="text-2xl font-bold tracking-tight">{t('reports.title')}</h1>
        <p className="text-muted-foreground text-sm mt-1">{t('reports.subtitle')}</p>
      </div>

      {/* Tabs */}
      <div className="flex gap-1 p-1 bg-muted rounded-lg w-fit">
        {tabs.map(({ key, label, icon: Icon }) => (
          <button
            key={key}
            type="button"
            onClick={() => setTab(key)}
            className={cn(
              'flex items-center gap-2 px-4 py-2 rounded-md text-sm font-medium transition-colors',
              tab === key
                ? 'bg-background text-foreground shadow-sm'
                : 'text-muted-foreground hover:text-foreground'
            )}
          >
            <Icon className="w-4 h-4" />
            {label}
          </button>
        ))}
      </div>

      {/* Date range filter (income & expenses only) */}
      {tab !== 'occupancy' && (
        <div className="flex flex-wrap items-center gap-3">
          <div className="flex items-center gap-2">
            <label className="text-sm font-medium text-muted-foreground">{t('reports.from')}</label>
            <input
              type="month"
              value={from}
              max={to}
              onChange={(e) => setFrom(e.target.value)}
              className="h-9 rounded-md border border-input bg-background px-3 text-sm focus:outline-none focus:ring-1 focus:ring-ring"
            />
          </div>
          <div className="flex items-center gap-2">
            <label className="text-sm font-medium text-muted-foreground">{t('reports.to')}</label>
            <input
              type="month"
              value={to}
              min={from}
              max={currentYearMonth()}
              onChange={(e) => setTo(e.target.value)}
              className="h-9 rounded-md border border-input bg-background px-3 text-sm focus:outline-none focus:ring-1 focus:ring-ring"
            />
          </div>
        </div>
      )}

      {/* Income Report */}
      {tab === 'income' && (
        <div className="space-y-6">
          {/* Summary cards */}
          <div className="grid grid-cols-2 md:grid-cols-4 gap-4">
            {[
              { label: t('reports.totalRent'), value: income?.total_rent ?? 0, color: 'text-emerald-500', bg: 'bg-emerald-500/10' },
              { label: t('reports.totalElectric'), value: income?.total_electric ?? 0, color: 'text-amber-500', bg: 'bg-amber-500/10' },
              { label: t('reports.totalWater'), value: income?.total_water ?? 0, color: 'text-blue-500', bg: 'bg-blue-500/10' },
              { label: t('reports.grandTotal'), value: income?.grand_total ?? 0, color: 'text-violet-500', bg: 'bg-violet-500/10' },
            ].map((item) => (
              <Card key={item.label} className="hover:shadow-md transition-shadow">
                <CardHeader className="pb-2">
                  <p className="text-sm font-medium text-muted-foreground">{item.label}</p>
                </CardHeader>
                <CardContent>
                  <p className={cn('text-2xl font-bold', item.color)}>
                    {loading ? '—' : `฿${formatBaht(item.value)}`}
                  </p>
                </CardContent>
              </Card>
            ))}
          </div>

          {/* Table */}
          <Card>
            <CardHeader>
              <CardTitle>{t('reports.incomeByMonth')}</CardTitle>
              <CardDescription>{t('reports.incomeByMonthDesc')}</CardDescription>
            </CardHeader>
            <CardContent>
              {loading ? (
                <p className="text-sm text-muted-foreground py-8 text-center">{t('common.loading')}</p>
              ) : !income?.rows?.length ? (
                <p className="text-sm text-muted-foreground py-8 text-center">{t('common.noResults')}</p>
              ) : (
                <div className="overflow-x-auto">
                  <table className="w-full text-sm">
                    <thead>
                      <tr className="border-b">
                        <th className="text-left py-3 px-2 font-medium text-muted-foreground">{t('reports.colMonth')}</th>
                        <th className="text-right py-3 px-2 font-medium text-muted-foreground">{t('reports.colRent')}</th>
                        <th className="text-right py-3 px-2 font-medium text-muted-foreground">{t('reports.colElectric')}</th>
                        <th className="text-right py-3 px-2 font-medium text-muted-foreground">{t('reports.colWater')}</th>
                        <th className="text-right py-3 px-2 font-medium text-muted-foreground">{t('reports.colOther')}</th>
                        <th className="text-right py-3 px-2 font-medium text-muted-foreground">{t('reports.colTotal')}</th>
                        <th className="text-right py-3 px-2 font-medium text-muted-foreground">{t('reports.colBills')}</th>
                      </tr>
                    </thead>
                    <tbody>
                      {income.rows.map((row) => (
                        <tr key={row.month} className="border-b last:border-0 hover:bg-muted/40 transition-colors">
                          <td className="py-3 px-2 font-medium">{row.month}</td>
                          <td className="py-3 px-2 text-right">฿{formatBaht(row.rent_amount)}</td>
                          <td className="py-3 px-2 text-right">฿{formatBaht(row.electric_amount)}</td>
                          <td className="py-3 px-2 text-right">฿{formatBaht(row.water_amount)}</td>
                          <td className="py-3 px-2 text-right">฿{formatBaht(row.other_amount)}</td>
                          <td className="py-3 px-2 text-right font-semibold">฿{formatBaht(row.total_amount)}</td>
                          <td className="py-3 px-2 text-right text-muted-foreground">{row.bill_count}</td>
                        </tr>
                      ))}
                    </tbody>
                    <tfoot>
                      <tr className="bg-muted/50 font-semibold">
                        <td className="py-3 px-2">{t('reports.total')}</td>
                        <td className="py-3 px-2 text-right">฿{formatBaht(income.total_rent)}</td>
                        <td className="py-3 px-2 text-right">฿{formatBaht(income.total_electric)}</td>
                        <td className="py-3 px-2 text-right">฿{formatBaht(income.total_water)}</td>
                        <td className="py-3 px-2 text-right">฿{formatBaht(income.total_other)}</td>
                        <td className="py-3 px-2 text-right text-violet-600">฿{formatBaht(income.grand_total)}</td>
                        <td className="py-3 px-2 text-right text-muted-foreground">
                          {income.rows.reduce((s, r) => s + r.bill_count, 0)}
                        </td>
                      </tr>
                    </tfoot>
                  </table>
                </div>
              )}
            </CardContent>
          </Card>
        </div>
      )}

      {/* Expense Report */}
      {tab === 'expenses' && (
        <div className="space-y-6">
          <div className="grid grid-cols-1 md:grid-cols-2 gap-4">
            <Card className="hover:shadow-md transition-shadow">
              <CardHeader className="pb-2">
                <p className="text-sm font-medium text-muted-foreground">{t('reports.grandTotal')}</p>
              </CardHeader>
              <CardContent>
                <p className="text-2xl font-bold text-rose-500">
                  {loading ? '—' : `฿${formatBaht(expenses?.grand_total ?? 0)}`}
                </p>
              </CardContent>
            </Card>
            <Card className="hover:shadow-md transition-shadow">
              <CardHeader className="pb-2">
                <p className="text-sm font-medium text-muted-foreground">{t('reports.totalTransactions')}</p>
              </CardHeader>
              <CardContent>
                <p className="text-2xl font-bold">
                  {loading ? '—' : (expenses?.rows?.reduce((s, r) => s + r.count, 0) ?? 0)}
                </p>
              </CardContent>
            </Card>
          </div>

          <Card>
            <CardHeader>
              <CardTitle>{t('reports.expensesByMonthCategory')}</CardTitle>
              <CardDescription>{t('reports.expensesByMonthCategoryDesc')}</CardDescription>
            </CardHeader>
            <CardContent>
              {loading ? (
                <p className="text-sm text-muted-foreground py-8 text-center">{t('common.loading')}</p>
              ) : !expenses?.rows?.length ? (
                <p className="text-sm text-muted-foreground py-8 text-center">{t('common.noResults')}</p>
              ) : (
                <div className="overflow-x-auto">
                  <table className="w-full text-sm">
                    <thead>
                      <tr className="border-b">
                        <th className="text-left py-3 px-2 font-medium text-muted-foreground">{t('reports.colMonth')}</th>
                        <th className="text-left py-3 px-2 font-medium text-muted-foreground">{t('reports.colCategory')}</th>
                        <th className="text-right py-3 px-2 font-medium text-muted-foreground">{t('reports.colTotal')}</th>
                        <th className="text-right py-3 px-2 font-medium text-muted-foreground">{t('reports.colCount')}</th>
                      </tr>
                    </thead>
                    <tbody>
                      {expenses.rows.map((row, i) => (
                        <tr key={`${row.month}-${row.category}-${i}`} className="border-b last:border-0 hover:bg-muted/40 transition-colors">
                          <td className="py-3 px-2 font-medium">{row.month}</td>
                          <td className="py-3 px-2">
                            <div className="flex items-center gap-2">
                              <span className={cn('w-2 h-2 rounded-full shrink-0', CATEGORY_COLORS[row.category] ?? 'bg-gray-400')} />
                              {t(`expenses.categories.${row.category}`, { defaultValue: row.category })}
                            </div>
                          </td>
                          <td className="py-3 px-2 text-right font-semibold text-rose-600">฿{formatBaht(row.total_amount)}</td>
                          <td className="py-3 px-2 text-right text-muted-foreground">{row.count}</td>
                        </tr>
                      ))}
                    </tbody>
                    <tfoot>
                      <tr className="bg-muted/50 font-semibold">
                        <td colSpan={2} className="py-3 px-2">{t('reports.total')}</td>
                        <td className="py-3 px-2 text-right text-rose-600">฿{formatBaht(expenses.grand_total)}</td>
                        <td className="py-3 px-2 text-right text-muted-foreground">
                          {expenses.rows.reduce((s, r) => s + r.count, 0)}
                        </td>
                      </tr>
                    </tfoot>
                  </table>
                </div>
              )}
            </CardContent>
          </Card>
        </div>
      )}

      {/* Occupancy Report */}
      {tab === 'occupancy' && (
        <div className="space-y-6">
          <div className="grid grid-cols-2 md:grid-cols-4 gap-4">
            {[
              { label: t('reports.totalRooms'), value: occupancy?.total ?? 0, color: 'text-foreground', bg: 'bg-muted' },
              { label: t('reports.occupiedRooms'), value: occupancy?.occupied ?? 0, color: 'text-emerald-600', bg: 'bg-emerald-500/10' },
              { label: t('reports.availableRooms'), value: occupancy?.available ?? 0, color: 'text-blue-600', bg: 'bg-blue-500/10' },
              {
                label: t('reports.occupancyRate'),
                value: loading ? '—' : `${(occupancy?.occupancy_rate ?? 0).toFixed(1)}%`,
                color: 'text-violet-600',
                bg: 'bg-violet-500/10',
                raw: true,
              },
            ].map((item) => (
              <Card key={item.label} className="hover:shadow-md transition-shadow">
                <CardHeader className="pb-2">
                  <p className="text-sm font-medium text-muted-foreground">{item.label}</p>
                </CardHeader>
                <CardContent>
                  <p className={cn('text-2xl font-bold', item.color)}>
                    {loading ? '—' : (item.raw ? item.value : item.value)}
                  </p>
                </CardContent>
              </Card>
            ))}
          </div>

          <Card>
            <CardHeader>
              <CardTitle>{t('reports.roomList')}</CardTitle>
              <CardDescription>{t('reports.roomListDesc')}</CardDescription>
            </CardHeader>
            <CardContent>
              {loading ? (
                <p className="text-sm text-muted-foreground py-8 text-center">{t('common.loading')}</p>
              ) : !occupancy?.rows?.length ? (
                <p className="text-sm text-muted-foreground py-8 text-center">{t('common.noResults')}</p>
              ) : (
                <div className="overflow-x-auto">
                  <table className="w-full text-sm">
                    <thead>
                      <tr className="border-b">
                        <th className="text-left py-3 px-2 font-medium text-muted-foreground">{t('reports.colRoom')}</th>
                        <th className="text-left py-3 px-2 font-medium text-muted-foreground">{t('reports.colFloor')}</th>
                        <th className="text-left py-3 px-2 font-medium text-muted-foreground">{t('reports.colType')}</th>
                        <th className="text-left py-3 px-2 font-medium text-muted-foreground">{t('reports.colStatus')}</th>
                        <th className="text-left py-3 px-2 font-medium text-muted-foreground">{t('reports.colTenant')}</th>
                        <th className="text-right py-3 px-2 font-medium text-muted-foreground">{t('reports.colRentPrice')}</th>
                      </tr>
                    </thead>
                    <tbody>
                      {occupancy.rows.map((row) => (
                        <tr key={row.room_number} className="border-b last:border-0 hover:bg-muted/40 transition-colors">
                          <td className="py-3 px-2 font-medium">{row.room_number}</td>
                          <td className="py-3 px-2">{row.floor}</td>
                          <td className="py-3 px-2">{t(`rooms.types.${row.type}`, { defaultValue: row.type })}</td>
                          <td className="py-3 px-2">
                            <span className={cn(
                              'inline-flex items-center px-2 py-0.5 rounded-full text-xs font-medium',
                              row.status === 'occupied' && 'bg-emerald-100 text-emerald-700 dark:bg-emerald-900/30 dark:text-emerald-400',
                              row.status === 'available' && 'bg-blue-100 text-blue-700 dark:bg-blue-900/30 dark:text-blue-400',
                              row.status === 'maintenance' && 'bg-amber-100 text-amber-700 dark:bg-amber-900/30 dark:text-amber-400',
                            )}>
                              {t(`rooms.statuses.${row.status}`, { defaultValue: row.status })}
                            </span>
                          </td>
                          <td className="py-3 px-2 text-muted-foreground">{row.tenant_name || '—'}</td>
                          <td className="py-3 px-2 text-right">฿{formatBaht(row.rent_price)}</td>
                        </tr>
                      ))}
                    </tbody>
                  </table>
                </div>
              )}
            </CardContent>
          </Card>
        </div>
      )}
    </div>
  )
}
