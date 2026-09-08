import {
  ArrowDown,
  ArrowUp,
  ArrowUpDown,
  ChevronLeft,
  ChevronRight,
  ChevronsLeft,
  ChevronsRight,
  MessageCircle,
  Pencil,
  Trash2,
  X,
} from 'lucide-react'
import { useLanguage, type TranslationKey } from '@/shared/i18n/language'
import { Badge } from '@/shared/components/ui/badge'
import { Card, CardContent, CardHeader, CardTitle } from '@/shared/components/ui/card'
import { ColumnFilterMenu } from '@/shared/components/column-filter-menu'
import { Table, TableBody, TableCell, TableHead, TableHeader, TableRow } from '@/shared/components/ui/table'
import { Button } from '@/shared/components/ui/button'
import type { ApiInvoice, InvoiceStatus } from '../types'
import {
  INVOICE_PAGE_SIZE_OPTIONS,
  formatPeriod,
  isTextFilterKey,
  toDateInputValue,
  type InvoiceColumnFilters,
  type InvoiceSortDirection,
  type InvoiceSortKey,
  type InvoiceStatusFilter,
  type InvoiceTextFilterKey,
} from '../utils'

const invoiceStatusLabelKeys: Record<InvoiceStatus, TranslationKey> = {
  unpaid: 'invoiceStatusUnpaid',
  paid: 'invoiceStatusPaid',
  overdue: 'invoiceStatusOverdue',
  cancelled: 'invoiceStatusCancelled',
}

const invoiceStatusBadgeVariant: Record<InvoiceStatus, 'default' | 'outline' | 'destructive' | 'secondary'> = {
  unpaid: 'secondary',
  paid: 'default',
  overdue: 'destructive',
  cancelled: 'outline',
}

const SORTABLE_COLUMNS: { key: InvoiceSortKey; labelKey: TranslationKey }[] = [
  { key: 'tenant_name', labelKey: 'invoiceTenantColumn' },
  { key: 'room_number', labelKey: 'invoiceRoomColumn' },
  { key: 'dormitory_name', labelKey: 'invoiceDormitoryColumn' },
  { key: 'period', labelKey: 'invoicePeriodColumn' },
  { key: 'due_date', labelKey: 'invoiceDueDateColumn' },
  { key: 'total_amount', labelKey: 'invoiceTotalAmountColumn' },
  { key: 'status', labelKey: 'invoiceStatusColumn' },
]

type InvoiceListCardProps = {
  isLoading: boolean
  loadError: string | null
  deleteError: string | null
  query: string
  onQueryChange: (query: string) => void
  statusFilter: InvoiceStatusFilter
  onStatusFilterChange: (value: InvoiceStatusFilter) => void
  columnFilters: InvoiceColumnFilters
  onColumnFilterChange: (key: InvoiceTextFilterKey, value: string) => void
  hasFilters: boolean
  sortKey: InvoiceSortKey
  sortDirection: InvoiceSortDirection
  onSort: (key: InvoiceSortKey) => void
  invoices: ApiInvoice[]
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
  deletingInvoiceId: string | null
  onCreateInvoice: () => void
  onEditInvoice: (invoice: ApiInvoice) => void
  onDeleteInvoice: (invoice: ApiInvoice) => void
  sendingLineInvoiceId: string | null
  onSendLineInvoice: (invoice: ApiInvoice) => void
  onOpenLineChat: (invoice: ApiInvoice) => void
}

export function InvoiceListCard({
  isLoading,
  loadError,
  deleteError,
  query,
  onQueryChange,
  statusFilter,
  onStatusFilterChange,
  columnFilters,
  onColumnFilterChange,
  hasFilters,
  sortKey,
  sortDirection,
  onSort,
  invoices,
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
  deletingInvoiceId,
  onCreateInvoice,
  onEditInvoice,
  onDeleteInvoice,
  sendingLineInvoiceId,
  onSendLineInvoice,
  onOpenLineChat,
}: InvoiceListCardProps) {
  const { t } = useLanguage()

  // The column filters live in a table that scrolls, so on a narrow screen
  // they're off to the right and there's no way to tell what's applied.
  // Summarising them here keeps that visible and clearable at any size.
  const activeFilters: { id: string; label: string; onClear: () => void }[] = []

  if (statusFilter !== 'all') {
    activeFilters.push({
      id: 'status',
      label: `${t('invoiceStatusColumn')}: ${t(invoiceStatusLabelKeys[statusFilter])}`,
      onClear: () => onStatusFilterChange('all'),
    })
  }

  for (const column of SORTABLE_COLUMNS) {
    const key = column.key
    if (!isTextFilterKey(key)) continue

    const value = columnFilters[key]
    if (!value) continue

    activeFilters.push({
      id: key,
      label: `${t(column.labelKey)}: ${value}`,
      onClear: () => onColumnFilterChange(key, ''),
    })
  }

  return (
    <Card>
      <CardHeader className="flex flex-col gap-3 sm:flex-row sm:items-start sm:justify-between sm:gap-4">
        <div>
          <CardTitle>{t('menuInvoices')}</CardTitle>
        </div>
        <Button onClick={onCreateInvoice} disabled={isLoading}>
          {t('invoiceCreate')}
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
                  {t('invoiceSearchLabel')}
                  <input
                    type="search"
                    className="h-10 rounded-md border border-input bg-transparent px-3 text-sm"
                    placeholder={t('invoiceSearchPlaceholder')}
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
              <div className="invoice-table-wrap overflow-x-auto">
                <Table>
                  <TableHeader>
                    <TableRow>
                      {SORTABLE_COLUMNS.map((column) => {
                        const isSorted = sortKey === column.key
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

                              {isTextFilterKey(column.key) && (
                                <ColumnFilterMenu
                                  label={t(column.labelKey)}
                                  textValue={columnFilters[column.key] ?? ''}
                                  onTextChange={(value) =>
                                    onColumnFilterChange(column.key as InvoiceTextFilterKey, value)
                                  }
                                />
                              )}

                              {column.key === 'status' && (
                                <ColumnFilterMenu
                                  label={t('invoiceStatusColumn')}
                                  optionValue={statusFilter}
                                  onOptionChange={onStatusFilterChange}
                                  options={[
                                    { value: 'all', label: t('filterAll') },
                                    { value: 'unpaid', label: t('invoiceStatusUnpaid') },
                                    { value: 'paid', label: t('invoiceStatusPaid') },
                                    { value: 'overdue', label: t('invoiceStatusOverdue') },
                                    { value: 'cancelled', label: t('invoiceStatusCancelled') },
                                  ]}
                                />
                              )}
                            </div>
                          </TableHead>
                        )
                      })}
                      <TableHead className="text-right">{t('invoiceActionsColumn')}</TableHead>
                    </TableRow>
                  </TableHeader>
                  <TableBody>
                    {invoices.length === 0 && (
                      <TableRow>
                        {/* The table scrolls, so centring this across every
                            column would push it off a phone screen. Pin it to
                            the left edge instead. */}
                        <TableCell colSpan={SORTABLE_COLUMNS.length + 1} className="p-0">
                          <p className="metric-detail sticky left-0 px-3 py-6">
                            {hasFilters ? t('invoiceNoMatching') : t('invoiceNoInvoices')}
                          </p>
                        </TableCell>
                      </TableRow>
                    )}
                    {invoices.map((invoice) => (
                      <TableRow key={invoice.id}>
                        <TableCell className="font-semibold">{invoice.tenant_name || '—'}</TableCell>
                        <TableCell className="text-muted-foreground">{invoice.room_number || '—'}</TableCell>
                        <TableCell className="text-muted-foreground">{invoice.dormitory_name || '—'}</TableCell>
                        <TableCell className="text-muted-foreground">
                          {formatPeriod(invoice.period_year, invoice.period_month)}
                        </TableCell>
                        <TableCell className="text-muted-foreground">
                          {toDateInputValue(invoice.due_date)}
                        </TableCell>
                        <TableCell className="text-muted-foreground">
                          {invoice.total_amount.toLocaleString()}
                        </TableCell>
                        <TableCell>
                          <Badge variant={invoiceStatusBadgeVariant[invoice.status]}>
                            {t(invoiceStatusLabelKeys[invoice.status])}
                          </Badge>
                        </TableCell>
                        <TableCell className="text-right">
                          {/* Never wrap: a squeezed actions column would stack
                              the buttons and blow up every row's height. The
                              table scrolls horizontally instead. */}
                          <div className="flex flex-nowrap justify-end gap-2">
                            <Button
                              type="button"
                              size="icon"
                              variant="outline"
                              title={
                                invoice.tenant_line_user_id
                                  ? t('invoiceSendLine')
                                  : invoice.tenant_line_id
                                    ? t('invoiceOpenLineChat')
                                    : t('invoiceSendLineUnavailable')
                              }
                              aria-label={t('invoiceSendLine')}
                              disabled={
                                (!invoice.tenant_line_user_id && !invoice.tenant_line_id) ||
                                sendingLineInvoiceId === invoice.id
                              }
                              onClick={() => {
                                if (invoice.tenant_line_user_id) {
                                  onSendLineInvoice(invoice)
                                  return
                                }
                                if (invoice.tenant_line_id) {
                                  onOpenLineChat(invoice)
                                }
                              }}
                            >
                              <MessageCircle />
                            </Button>
                            <Button
                              type="button"
                              size="icon"
                              variant="outline"
                              title={t('invoiceEdit')}
                              aria-label={t('invoiceEdit')}
                              onClick={() => onEditInvoice(invoice)}
                            >
                              <Pencil />
                            </Button>
                            <Button
                              type="button"
                              size="icon"
                              variant="destructive"
                              title={t('invoiceDelete')}
                              aria-label={t('invoiceDelete')}
                              onClick={() => onDeleteInvoice(invoice)}
                              disabled={deletingInvoiceId === invoice.id}
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
                      {INVOICE_PAGE_SIZE_OPTIONS.map((size) => (
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
