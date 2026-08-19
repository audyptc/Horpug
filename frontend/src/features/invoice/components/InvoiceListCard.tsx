import { ChevronLeft, ChevronRight, ChevronsLeft, ChevronsRight, Pencil, Trash2 } from 'lucide-react'
import { useLanguage, type TranslationKey } from '@/shared/i18n/language'
import { Badge } from '@/shared/components/ui/badge'
import { Card, CardContent, CardHeader, CardTitle } from '@/shared/components/ui/card'
import { Table, TableBody, TableCell, TableHead, TableHeader, TableRow } from '@/shared/components/ui/table'
import { Button } from '@/shared/components/ui/button'
import type { ApiInvoice, InvoiceStatus } from '../types'
import { INVOICE_PAGE_SIZE_OPTIONS, formatPeriod, toDateInputValue } from '../utils'

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

type InvoiceListCardProps = {
  isLoading: boolean
  loadError: string | null
  deleteError: string | null
  invoices: ApiInvoice[] | null
  query: string
  onQueryChange: (query: string) => void
  filteredInvoices: ApiInvoice[]
  paginatedInvoices: ApiInvoice[]
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
}

export function InvoiceListCard({
  isLoading,
  loadError,
  deleteError,
  invoices,
  query,
  onQueryChange,
  filteredInvoices,
  paginatedInvoices,
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
}: InvoiceListCardProps) {
  const { t } = useLanguage()

  return (
    <Card>
      <CardHeader className="flex flex-row items-start justify-between gap-4">
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

        {!loadError && !isLoading && invoices && invoices.length === 0 && (
          <p className="metric-detail">{t('invoiceNoInvoices')}</p>
        )}

        {!loadError && !isLoading && invoices && invoices.length > 0 && (
          <>
            <label className="flex w-full max-w-md flex-col gap-1.5 text-sm font-medium">
              {t('invoiceSearchLabel')}
              <input
                type="search"
                className="h-10 rounded-md border border-input bg-transparent px-3 text-sm"
                placeholder={t('invoiceSearchPlaceholder')}
                value={query}
                onChange={(event) => onQueryChange(event.target.value)}
              />
            </label>

            {filteredInvoices.length === 0 && <p className="metric-detail">{t('invoiceNoMatching')}</p>}

            {filteredInvoices.length > 0 && (
              <div className="table-wrap invoice-table-wrap">
                <Table>
                  <TableHeader>
                    <TableRow>
                      <TableHead>{t('invoiceTenantColumn')}</TableHead>
                      <TableHead>{t('invoiceRoomColumn')}</TableHead>
                      <TableHead>{t('invoicePeriodColumn')}</TableHead>
                      <TableHead>{t('invoiceDueDateColumn')}</TableHead>
                      <TableHead>{t('invoiceTotalAmountColumn')}</TableHead>
                      <TableHead>{t('invoiceStatusColumn')}</TableHead>
                      <TableHead className="text-right">{t('invoiceActionsColumn')}</TableHead>
                    </TableRow>
                  </TableHeader>
                  <TableBody>
                    {paginatedInvoices.map((invoice) => (
                      <TableRow key={invoice.id}>
                        <TableCell className="font-semibold">{invoice.tenant_name || '—'}</TableCell>
                        <TableCell className="text-muted-foreground">
                          {invoice.room_number || '—'}
                          {invoice.dormitory_name ? ` (${invoice.dormitory_name})` : ''}
                        </TableCell>
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
                          <div className="flex flex-wrap justify-end gap-2">
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
            )}

            {filteredInvoices.length > 0 && (
              <div className="flex flex-col gap-3 sm:flex-row sm:items-center sm:justify-between">
                <p className="text-sm text-muted-foreground">
                  {t('rolePermissionsShowingLabel')} {rangeStart}-{rangeEnd}{' '}
                  {t('rolePermissionsOfLabel')} {filteredInvoices.length} {t('rolePermissionsResultsLabel')}
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
