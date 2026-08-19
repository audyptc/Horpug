import { ChevronLeft, ChevronRight, ChevronsLeft, ChevronsRight, Pencil, Trash2 } from 'lucide-react'
import { useLanguage, type TranslationKey } from '@/shared/i18n/language'
import { Badge } from '@/shared/components/ui/badge'
import { Card, CardContent, CardHeader, CardTitle } from '@/shared/components/ui/card'
import { Table, TableBody, TableCell, TableHead, TableHeader, TableRow } from '@/shared/components/ui/table'
import { Button } from '@/shared/components/ui/button'
import type { ApiMeter, BillingMethod } from '../types'
import { METER_PAGE_SIZE_OPTIONS, toDateInputValue } from '../utils'

const billingMethodLabelKeys: Record<BillingMethod, TranslationKey> = {
  metered: 'meterBillingMethodMetered',
  flat: 'meterBillingMethodFlat',
}

const billingMethodBadgeVariant: Record<BillingMethod, 'default' | 'outline'> = {
  metered: 'default',
  flat: 'outline',
}

type MeterListCardProps = {
  isLoading: boolean
  loadError: string | null
  deleteError: string | null
  meters: ApiMeter[] | null
  query: string
  onQueryChange: (query: string) => void
  filteredMeters: ApiMeter[]
  paginatedMeters: ApiMeter[]
  currentPage: number
  totalPages: number
  rangeStart: number
  rangeEnd: number
  pageSize: number
  onPageSizeChange: (size: number) => void
  onFirstPage: () => void
  onPrevPage: () => void
  onNextPage: () => void
  onLastPage: () => void
  deletingMeterId: string | null
  onCreateMeter: () => void
  onEditMeter: (meter: ApiMeter) => void
  onDeleteMeter: (meter: ApiMeter) => void
}

export function MeterListCard({
  isLoading,
  loadError,
  deleteError,
  meters,
  query,
  onQueryChange,
  filteredMeters,
  paginatedMeters,
  currentPage,
  totalPages,
  rangeStart,
  rangeEnd,
  pageSize,
  onPageSizeChange,
  onFirstPage,
  onPrevPage,
  onNextPage,
  onLastPage,
  deletingMeterId,
  onCreateMeter,
  onEditMeter,
  onDeleteMeter,
}: MeterListCardProps) {
  const { t } = useLanguage()

  return (
    <Card>
      <CardHeader className="flex flex-row items-start justify-between gap-4">
        <div>
          <CardTitle>{t('menuMeters')}</CardTitle>
        </div>
        <Button onClick={onCreateMeter} disabled={isLoading}>
          {t('meterCreate')}
        </Button>
      </CardHeader>
      <CardContent className="flex flex-col gap-4">
        {loadError && <p className="resource-error">{loadError}</p>}
        {deleteError && <p className="resource-error">{deleteError}</p>}

        {!loadError && isLoading && <p className="metric-detail">{t('loading')}</p>}

        {!loadError && !isLoading && meters && meters.length === 0 && (
          <p className="metric-detail">{t('meterNoMeters')}</p>
        )}

        {!loadError && !isLoading && meters && meters.length > 0 && (
          <>
            <label className="flex w-full max-w-md flex-col gap-1.5 text-sm font-medium">
              {t('meterSearchLabel')}
              <input
                type="search"
                className="h-10 rounded-md border border-input bg-transparent px-3 text-sm"
                placeholder={t('meterSearchPlaceholder')}
                value={query}
                onChange={(event) => onQueryChange(event.target.value)}
              />
            </label>

            {filteredMeters.length === 0 && <p className="metric-detail">{t('meterNoMatching')}</p>}

            {filteredMeters.length > 0 && (
              <div className="table-wrap meter-table-wrap">
                <Table>
                  <TableHeader>
                    <TableRow>
                      <TableHead>{t('meterRoomColumn')}</TableHead>
                      <TableHead>{t('meterReadingDateColumn')}</TableHead>
                      <TableHead>{t('meterBillingMethodColumn')}</TableHead>
                      <TableHead>{t('meterPreviousUnitColumn')}</TableHead>
                      <TableHead>{t('meterCurrentUnitColumn')}</TableHead>
                      <TableHead>{t('meterUnitUsedColumn')}</TableHead>
                      <TableHead>{t('meterPricePerUnitColumn')}</TableHead>
                      <TableHead>{t('meterTotalAmountColumn')}</TableHead>
                      <TableHead className="text-right">{t('meterActionsColumn')}</TableHead>
                    </TableRow>
                  </TableHeader>
                  <TableBody>
                    {paginatedMeters.map((meter) => (
                      <TableRow key={meter.id}>
                        <TableCell className="font-semibold">
                          {meter.room_number || '—'}
                          {meter.dormitory_name ? ` (${meter.dormitory_name})` : ''}
                        </TableCell>
                        <TableCell className="text-muted-foreground">
                          {toDateInputValue(meter.reading_date)}
                        </TableCell>
                        <TableCell>
                          <div className="flex flex-wrap items-center gap-1.5">
                            <Badge variant={billingMethodBadgeVariant[meter.billing_method]}>
                              {t(billingMethodLabelKeys[meter.billing_method])}
                            </Badge>
                            {!meter.is_billed && (
                              <Badge variant="secondary">{t('meterNotBilledBadge')}</Badge>
                            )}
                          </div>
                        </TableCell>
                        <TableCell className="text-muted-foreground">
                          {meter.billing_method === 'metered' ? meter.previous_unit.toLocaleString() : '—'}
                        </TableCell>
                        <TableCell className="text-muted-foreground">
                          {meter.billing_method === 'metered' ? meter.current_unit.toLocaleString() : '—'}
                        </TableCell>
                        <TableCell className="text-muted-foreground">
                          {meter.billing_method === 'metered' ? meter.unit_used.toLocaleString() : '—'}
                        </TableCell>
                        <TableCell className="text-muted-foreground">
                          {meter.billing_method === 'metered' ? meter.price_per_unit.toLocaleString() : '—'}
                        </TableCell>
                        <TableCell className="font-semibold">{meter.total_amount.toLocaleString()}</TableCell>
                        <TableCell className="text-right">
                          <div className="flex flex-wrap justify-end gap-2">
                            <Button
                              type="button"
                              size="icon"
                              variant="outline"
                              title={t('meterEdit')}
                              aria-label={t('meterEdit')}
                              onClick={() => onEditMeter(meter)}
                            >
                              <Pencil />
                            </Button>
                            <Button
                              type="button"
                              size="icon"
                              variant="destructive"
                              title={t('meterDelete')}
                              aria-label={t('meterDelete')}
                              onClick={() => onDeleteMeter(meter)}
                              disabled={deletingMeterId === meter.id}
                            >
                              <Trash2 />
                            </Button>
                          </div>
                        </TableCell>
                      </TableRow>
                    ))}
                  </TableBody>
                </Table>
              </div>
            )}

            {filteredMeters.length > 0 && (
              <div className="flex flex-col gap-3 sm:flex-row sm:items-center sm:justify-between">
                <p className="text-sm text-muted-foreground">
                  {t('rolePermissionsShowingLabel')} {rangeStart}-{rangeEnd}{' '}
                  {t('rolePermissionsOfLabel')} {filteredMeters.length} {t('rolePermissionsResultsLabel')}
                  {totalPages > 1 && (
                    <>
                      {' '}
                      · {t('rolePermissionsPageLabel')} {currentPage} / {totalPages}
                    </>
                  )}
                </p>
                <div className="flex flex-wrap items-center gap-3">
                  <label className="flex items-center gap-1.5 text-sm text-muted-foreground">
                    {t('rolePermissionsPageSizeLabel')}
                    <select
                      className="h-9 rounded-md border border-input bg-transparent px-2 text-sm"
                      value={pageSize}
                      onChange={(event) => onPageSizeChange(Number(event.target.value))}
                    >
                      {METER_PAGE_SIZE_OPTIONS.map((size) => (
                        <option key={size} value={size}>
                          {size}
                        </option>
                      ))}
                    </select>
                  </label>

                  {totalPages > 1 && (
                    <div className="flex gap-2">
                      <Button
                        type="button"
                        size="icon"
                        variant="outline"
                        title={t('rolePermissionsFirstPage')}
                        aria-label={t('rolePermissionsFirstPage')}
                        disabled={currentPage <= 1}
                        onClick={onFirstPage}
                      >
                        <ChevronsLeft />
                      </Button>
                      <Button
                        type="button"
                        size="icon"
                        variant="outline"
                        title={t('rolePermissionsPrevPage')}
                        aria-label={t('rolePermissionsPrevPage')}
                        disabled={currentPage <= 1}
                        onClick={onPrevPage}
                      >
                        <ChevronLeft />
                      </Button>
                      <Button
                        type="button"
                        size="icon"
                        variant="outline"
                        title={t('rolePermissionsNextPage')}
                        aria-label={t('rolePermissionsNextPage')}
                        disabled={currentPage >= totalPages}
                        onClick={onNextPage}
                      >
                        <ChevronRight />
                      </Button>
                      <Button
                        type="button"
                        size="icon"
                        variant="outline"
                        title={t('rolePermissionsLastPage')}
                        aria-label={t('rolePermissionsLastPage')}
                        disabled={currentPage >= totalPages}
                        onClick={onLastPage}
                      >
                        <ChevronsRight />
                      </Button>
                    </div>
                  )}
                </div>
              </div>
            )}
          </>
        )}
      </CardContent>
    </Card>
  )
}
