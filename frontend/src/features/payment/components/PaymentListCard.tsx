import { ArrowDown, ArrowUp, ArrowUpDown, ChevronLeft, ChevronRight, ChevronsLeft, ChevronsRight, Trash2, X } from 'lucide-react'
import { useLanguage, type TranslationKey } from '@/shared/i18n/language'
import { Badge } from '@/shared/components/ui/badge'
import { Card, CardContent, CardHeader, CardTitle } from '@/shared/components/ui/card'
import { ColumnFilterMenu } from '@/shared/components/column-filter-menu'
import { Table, TableBody, TableCell, TableHead, TableHeader, TableRow } from '@/shared/components/ui/table'
import { Button } from '@/shared/components/ui/button'
import type { ApiPayment, PaymentMethod } from '../types'
import {
  PAYMENT_METHODS,
  PAYMENT_PAGE_SIZE_OPTIONS,
  toDateInputValue,
  type PaymentColumnFilters,
  type PaymentMethodFilter,
  type PaymentSortDirection,
  type PaymentSortKey,
  type PaymentTextFilterKey,
} from '../utils'

const paymentMethodLabelKeys: Record<PaymentMethod, TranslationKey> = {
  cash: 'paymentMethodCash',
  transfer: 'paymentMethodTransfer',
  credit_card: 'paymentMethodCreditCard',
  other: 'paymentMethodOther',
}

const paymentMethodBadgeVariant: Record<PaymentMethod, 'default' | 'outline' | 'destructive' | 'secondary'> = {
  cash: 'secondary',
  transfer: 'default',
  credit_card: 'outline',
  other: 'outline',
}

const SORTABLE_COLUMNS: { key: PaymentSortKey; labelKey: TranslationKey }[] = [
  { key: 'tenant_name', labelKey: 'paymentTenantColumn' },
  { key: 'room_number', labelKey: 'paymentRoomColumn' },
  { key: 'amount', labelKey: 'paymentAmountColumn' },
  { key: 'payment_date', labelKey: 'paymentDateColumn' },
]

type PaymentListCardProps = {
  isLoading: boolean
  loadError: string | null
  deleteError: string | null
  query: string
  onQueryChange: (query: string) => void
  methodFilter: PaymentMethodFilter
  onMethodFilterChange: (value: PaymentMethodFilter) => void
  columnFilters: PaymentColumnFilters
  onColumnFilterChange: (key: PaymentTextFilterKey, value: string) => void
  hasFilters: boolean
  sortKey: PaymentSortKey
  sortDirection: PaymentSortDirection
  onSort: (key: PaymentSortKey) => void
  payments: ApiPayment[]
  total: number
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
  deletingPaymentId: string | null
  onCreatePayment: () => void
  onDeletePayment: (payment: ApiPayment) => void
}

export function PaymentListCard({
  isLoading,
  loadError,
  deleteError,
  query,
  onQueryChange,
  methodFilter,
  onMethodFilterChange,
  columnFilters,
  onColumnFilterChange,
  hasFilters,
  sortKey,
  sortDirection,
  onSort,
  payments,
  total,
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
  deletingPaymentId,
  onCreatePayment,
  onDeletePayment,
}: PaymentListCardProps) {
  const { t } = useLanguage()

  // The column filters live in a table that scrolls, so on a narrow screen
  // they're off to the right and there's no way to tell what's applied.
  // Summarising them here keeps that visible and clearable at any size.
  const activeFilters: { id: string; label: string; onClear: () => void }[] = []

  if (methodFilter !== 'all') {
    activeFilters.push({
      id: 'payment_method',
      label: `${t('paymentFilterMethodLabel')}: ${t(paymentMethodLabelKeys[methodFilter])}`,
      onClear: () => onMethodFilterChange('all'),
    })
  }

  const textFilterLabelKeys: Record<PaymentTextFilterKey, TranslationKey> = {
    tenant_name: 'paymentTenantColumn',
    room_number: 'paymentRoomColumn',
    reference_no: 'paymentReferenceColumn',
  }

  for (const key of Object.keys(textFilterLabelKeys) as PaymentTextFilterKey[]) {
    const value = columnFilters[key]
    if (!value) continue

    activeFilters.push({
      id: key,
      label: `${t(textFilterLabelKeys[key])}: ${value}`,
      onClear: () => onColumnFilterChange(key, ''),
    })
  }

  return (
    <Card>
      <CardHeader className="flex flex-col gap-3 sm:flex-row sm:items-start sm:justify-between sm:gap-4">
        <div>
          <CardTitle>{t('menuPayments')}</CardTitle>
        </div>
        <Button onClick={onCreatePayment} disabled={isLoading}>
          {t('paymentCreate')}
        </Button>
      </CardHeader>
      <CardContent className="flex flex-col gap-4">
        {loadError && <p className="resource-error">{loadError}</p>}
        {deleteError && <p className="resource-error">{deleteError}</p>}

        {!loadError && isLoading && <p className="metric-detail">{t('loading')}</p>}

        {!loadError && !isLoading && (
          <>
            <div className="overflow-hidden rounded-md border border-border">
              <div className="flex flex-col gap-3 border-b border-border bg-muted/40 p-3">
                <label className="flex w-full flex-col gap-1.5 text-sm font-medium sm:max-w-md">
                  {t('paymentSearchLabel')}
                  <input
                    type="search"
                    className="h-10 rounded-md border border-input bg-transparent px-3 text-sm"
                    placeholder={t('paymentSearchPlaceholder')}
                    value={query}
                    onChange={(event) => onQueryChange(event.target.value)}
                  />
                </label>

                {activeFilters.length > 0 && (
                  <div className="flex flex-wrap items-center gap-1.5">
                    {activeFilters.map((filter) => (
                      <span
                        key={filter.id}
                        className="inline-flex max-w-full items-center gap-1 rounded-full border border-border bg-background py-0.5 pl-2.5 pr-1 text-xs"
                      >
                        <span className="truncate">{filter.label}</span>
                        <button
                          type="button"
                          onClick={filter.onClear}
                          title={t('filterClear')}
                          aria-label={`${t('filterClear')}: ${filter.label}`}
                          className="inline-flex size-5 shrink-0 items-center justify-center rounded-full text-muted-foreground transition-colors hover:bg-accent hover:text-foreground"
                        >
                          <X size={12} />
                        </button>
                      </span>
                    ))}
                  </div>
                )}
              </div>

              {/* Rendered even with no matches: the column filters live in the
                  header, so hiding it would strand the user with no way to
                  widen the filter again. */}
              <div className="payment-table-wrap overflow-x-auto">
                <Table>
                  <TableHeader>
                    <TableRow>
                      {SORTABLE_COLUMNS.map((column) => {
                        const isSorted = sortKey === column.key
                        const filterKey =
                          column.key === 'tenant_name' || column.key === 'room_number'
                            ? (column.key as PaymentTextFilterKey)
                            : null

                        return (
                          <TableHead
                            key={column.key}
                            aria-sort={
                              isSorted ? (sortDirection === 'asc' ? 'ascending' : 'descending') : 'none'
                            }
                          >
                            <div className="flex items-center gap-1">
                              <button
                                type="button"
                                onClick={() => onSort(column.key)}
                                title={
                                  isSorted && sortDirection === 'asc'
                                    ? t('sortAscending')
                                    : t('sortDescending')
                                }
                                className="inline-flex items-center gap-1 whitespace-nowrap transition-colors hover:text-foreground"
                              >
                                {t(column.labelKey)}
                                {isSorted ? (
                                  sortDirection === 'asc' ? (
                                    <ArrowUp size={13} className="shrink-0" />
                                  ) : (
                                    <ArrowDown size={13} className="shrink-0" />
                                  )
                                ) : (
                                  <ArrowUpDown size={13} className="shrink-0 opacity-40" />
                                )}
                              </button>

                              {filterKey && (
                                <ColumnFilterMenu
                                  label={t(column.labelKey)}
                                  textValue={columnFilters[filterKey] ?? ''}
                                  onTextChange={(value) => onColumnFilterChange(filterKey, value)}
                                />
                              )}
                            </div>
                          </TableHead>
                        )
                      })}
                      <TableHead>
                        <div className="flex items-center gap-1">
                          {t('paymentMethodColumn')}
                          <ColumnFilterMenu
                            label={t('paymentFilterMethodLabel')}
                            optionValue={methodFilter}
                            onOptionChange={onMethodFilterChange}
                            options={[
                              { value: 'all', label: t('filterAll') },
                              ...PAYMENT_METHODS.map((method) => ({
                                value: method,
                                label: t(paymentMethodLabelKeys[method]),
                              })),
                            ]}
                          />
                        </div>
                      </TableHead>
                      <TableHead>
                        <div className="flex items-center gap-1">
                          {t('paymentReferenceColumn')}
                          <ColumnFilterMenu
                            label={t('paymentReferenceColumn')}
                            textValue={columnFilters.reference_no ?? ''}
                            onTextChange={(value) => onColumnFilterChange('reference_no', value)}
                          />
                        </div>
                      </TableHead>
                      <TableHead className="text-right">{t('paymentActionsColumn')}</TableHead>
                    </TableRow>
                  </TableHeader>
                  <TableBody>
                    {payments.length === 0 && (
                      <TableRow>
                        {/* The table scrolls horizontally, so centring this
                            across every column would push it off a phone
                            screen. Pin it to the left edge instead. */}
                        <TableCell colSpan={SORTABLE_COLUMNS.length + 3} className="p-0">
                          <p className="metric-detail sticky left-0 px-3 py-6">
                            {hasFilters ? t('paymentNoMatching') : t('paymentNoPayments')}
                          </p>
                        </TableCell>
                      </TableRow>
                    )}
                    {payments.map((payment) => {
                      const references = payment.items
                        .map((item) => item.reference_no)
                        .filter((value) => value.trim().length > 0)

                      return (
                        <TableRow key={payment.id}>
                          <TableCell className="font-semibold">{payment.tenant_name || '—'}</TableCell>
                          <TableCell className="text-muted-foreground">
                            {payment.room_number || '—'}
                            {payment.dormitory_name ? ` (${payment.dormitory_name})` : ''}
                          </TableCell>
                          <TableCell className="text-muted-foreground">
                            {payment.total_amount.toLocaleString()}
                          </TableCell>
                          <TableCell className="text-muted-foreground">
                            {toDateInputValue(payment.payment_date)}
                          </TableCell>
                          <TableCell>
                            <div className="flex flex-wrap gap-1">
                              {payment.items.map((item) => (
                                <Badge key={item.id} variant={paymentMethodBadgeVariant[item.payment_method]}>
                                  {t(paymentMethodLabelKeys[item.payment_method])} · {item.amount.toLocaleString()}
                                </Badge>
                              ))}
                            </div>
                          </TableCell>
                          <TableCell className="text-muted-foreground">
                            {references.length > 0 ? references.join(', ') : '—'}
                          </TableCell>
                          <TableCell className="text-right">
                            <div className="flex flex-nowrap justify-end gap-2">
                              <Button
                                type="button"
                                size="icon"
                                variant="destructive"
                                title={t('paymentDelete')}
                                aria-label={t('paymentDelete')}
                                onClick={() => onDeletePayment(payment)}
                                disabled={deletingPaymentId === payment.id}
                              >
                                <Trash2 />
                              </Button>
                            </div>
                          </TableCell>
                        </TableRow>
                      )
                    })}
                  </TableBody>
                </Table>
              </div>
            </div>

            {total > 0 && (
              <div className="flex flex-col gap-3 sm:flex-row sm:items-center sm:justify-between">
                <p className="text-sm text-muted-foreground">
                  {t('rolePermissionsShowingLabel')} {rangeStart}-{rangeEnd}{' '}
                  {t('rolePermissionsOfLabel')} {total} {t('rolePermissionsResultsLabel')}
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
                      {PAYMENT_PAGE_SIZE_OPTIONS.map((size) => (
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
