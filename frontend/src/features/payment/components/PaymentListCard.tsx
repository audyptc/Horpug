import { ChevronLeft, ChevronRight, Trash2 } from 'lucide-react'
import { useLanguage, type TranslationKey } from '@/shared/i18n/language'
import { Badge } from '@/shared/components/ui/badge'
import { Card, CardContent, CardHeader, CardTitle } from '@/shared/components/ui/card'
import { Table, TableBody, TableCell, TableHead, TableHeader, TableRow } from '@/shared/components/ui/table'
import { Button } from '@/shared/components/ui/button'
import type { ApiPayment, PaymentMethod } from '../types'
import { PAYMENT_PAGE_SIZE_OPTIONS, toDateInputValue } from '../utils'

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

type PaymentListCardProps = {
  isLoading: boolean
  loadError: string | null
  deleteError: string | null
  payments: ApiPayment[] | null
  query: string
  onQueryChange: (query: string) => void
  filteredPayments: ApiPayment[]
  paginatedPayments: ApiPayment[]
  currentPage: number
  totalPages: number
  rangeStart: number
  rangeEnd: number
  pageSize: number
  onPageSizeChange: (size: number) => void
  onPrevPage: () => void
  onNextPage: () => void
  deletingPaymentId: string | null
  onCreatePayment: () => void
  onDeletePayment: (payment: ApiPayment) => void
}

export function PaymentListCard({
  isLoading,
  loadError,
  deleteError,
  payments,
  query,
  onQueryChange,
  filteredPayments,
  paginatedPayments,
  currentPage,
  totalPages,
  rangeStart,
  rangeEnd,
  pageSize,
  onPageSizeChange,
  onPrevPage,
  onNextPage,
  deletingPaymentId,
  onCreatePayment,
  onDeletePayment,
}: PaymentListCardProps) {
  const { t } = useLanguage()

  return (
    <Card>
      <CardHeader className="flex flex-row items-start justify-between gap-4">
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

        {!loadError && !isLoading && payments && payments.length === 0 && (
          <p className="metric-detail">{t('paymentNoPayments')}</p>
        )}

        {!loadError && !isLoading && payments && payments.length > 0 && (
          <>
            <label className="flex w-full max-w-md flex-col gap-1.5 text-sm font-medium">
              {t('paymentSearchLabel')}
              <input
                type="search"
                className="h-10 rounded-md border border-input bg-transparent px-3 text-sm"
                placeholder={t('paymentSearchPlaceholder')}
                value={query}
                onChange={(event) => onQueryChange(event.target.value)}
              />
            </label>

            {filteredPayments.length === 0 && <p className="metric-detail">{t('paymentNoMatching')}</p>}

            {filteredPayments.length > 0 && (
              <div className="table-wrap">
                <Table>
                  <TableHeader>
                    <TableRow>
                      <TableHead>{t('paymentTenantColumn')}</TableHead>
                      <TableHead>{t('paymentRoomColumn')}</TableHead>
                      <TableHead>{t('paymentAmountColumn')}</TableHead>
                      <TableHead>{t('paymentMethodColumn')}</TableHead>
                      <TableHead>{t('paymentDateColumn')}</TableHead>
                      <TableHead>{t('paymentReferenceColumn')}</TableHead>
                      <TableHead className="text-right">{t('paymentActionsColumn')}</TableHead>
                    </TableRow>
                  </TableHeader>
                  <TableBody>
                    {paginatedPayments.map((payment) => {
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
                            {toDateInputValue(payment.payment_date)}
                          </TableCell>
                          <TableCell className="text-muted-foreground">
                            {references.length > 0 ? references.join(', ') : '—'}
                          </TableCell>
                          <TableCell className="text-right">
                            <div className="flex flex-wrap justify-end gap-2">
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
            )}

            {filteredPayments.length > 0 && (
              <div className="flex items-center justify-between gap-3">
                <p className="text-sm text-muted-foreground">
                  {t('rolePermissionsShowingLabel')} {rangeStart}-{rangeEnd}{' '}
                  {t('rolePermissionsOfLabel')} {filteredPayments.length} {t('rolePermissionsResultsLabel')}
                  {totalPages > 1 && (
                    <>
                      {' '}
                      · {t('rolePermissionsPageLabel')} {currentPage} / {totalPages}
                    </>
                  )}
                </p>
                <div className="flex items-center gap-3">
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
