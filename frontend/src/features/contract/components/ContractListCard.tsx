import { ChevronLeft, ChevronRight, Pencil, Trash2 } from 'lucide-react'
import { useLanguage, type TranslationKey } from '@/shared/i18n/language'
import { Badge } from '@/shared/components/ui/badge'
import { Card, CardContent, CardHeader, CardTitle } from '@/shared/components/ui/card'
import { Table, TableBody, TableCell, TableHead, TableHeader, TableRow } from '@/shared/components/ui/table'
import { Button } from '@/shared/components/ui/button'
import type { ApiContract, ContractStatus } from '../types'
import { CONTRACT_PAGE_SIZE_OPTIONS, toDateInputValue } from '../utils'

const contractStatusLabelKeys: Record<ContractStatus, TranslationKey> = {
  active: 'contractStatusActive',
  expired: 'contractStatusExpired',
  terminated: 'contractStatusTerminated',
}

const contractStatusBadgeVariant: Record<ContractStatus, 'default' | 'outline' | 'destructive'> = {
  active: 'default',
  expired: 'outline',
  terminated: 'destructive',
}

type ContractListCardProps = {
  isLoading: boolean
  loadError: string | null
  deleteError: string | null
  contracts: ApiContract[] | null
  query: string
  onQueryChange: (query: string) => void
  filteredContracts: ApiContract[]
  paginatedContracts: ApiContract[]
  currentPage: number
  totalPages: number
  rangeStart: number
  rangeEnd: number
  pageSize: number
  onPageSizeChange: (size: number) => void
  onPrevPage: () => void
  onNextPage: () => void
  deletingContractId: string | null
  onCreateContract: () => void
  onEditContract: (contract: ApiContract) => void
  onDeleteContract: (contract: ApiContract) => void
}

export function ContractListCard({
  isLoading,
  loadError,
  deleteError,
  contracts,
  query,
  onQueryChange,
  filteredContracts,
  paginatedContracts,
  currentPage,
  totalPages,
  rangeStart,
  rangeEnd,
  pageSize,
  onPageSizeChange,
  onPrevPage,
  onNextPage,
  deletingContractId,
  onCreateContract,
  onEditContract,
  onDeleteContract,
}: ContractListCardProps) {
  const { t } = useLanguage()

  return (
    <Card>
      <CardHeader className="flex flex-row items-start justify-between gap-4">
        <div>
          <CardTitle>{t('menuContracts')}</CardTitle>
        </div>
        <Button onClick={onCreateContract} disabled={isLoading}>
          {t('contractCreate')}
        </Button>
      </CardHeader>
      <CardContent className="flex flex-col gap-4">
        {loadError && <p className="resource-error">{loadError}</p>}
        {deleteError && <p className="resource-error">{deleteError}</p>}

        {!loadError && isLoading && <p className="metric-detail">{t('loading')}</p>}

        {!loadError && !isLoading && contracts && contracts.length === 0 && (
          <p className="metric-detail">{t('contractNoContracts')}</p>
        )}

        {!loadError && !isLoading && contracts && contracts.length > 0 && (
          <>
            <label className="flex w-full max-w-md flex-col gap-1.5 text-sm font-medium">
              {t('contractSearchLabel')}
              <input
                type="search"
                className="h-10 rounded-md border border-input bg-transparent px-3 text-sm"
                placeholder={t('contractSearchPlaceholder')}
                value={query}
                onChange={(event) => onQueryChange(event.target.value)}
              />
            </label>

            {filteredContracts.length === 0 && <p className="metric-detail">{t('contractNoMatching')}</p>}

            {filteredContracts.length > 0 && (
              <div className="table-wrap">
                <Table>
                  <TableHeader>
                    <TableRow>
                      <TableHead>{t('contractTenantColumn')}</TableHead>
                      <TableHead>{t('contractRoomColumn')}</TableHead>
                      <TableHead>{t('contractStartDateColumn')}</TableHead>
                      <TableHead>{t('contractEndDateColumn')}</TableHead>
                      <TableHead>{t('contractRentPriceColumn')}</TableHead>
                      <TableHead>{t('contractDepositColumn')}</TableHead>
                      <TableHead>{t('contractStatusColumn')}</TableHead>
                      <TableHead className="text-right">{t('contractActionsColumn')}</TableHead>
                    </TableRow>
                  </TableHeader>
                  <TableBody>
                    {paginatedContracts.map((contract) => (
                      <TableRow key={contract.id}>
                        <TableCell className="font-semibold">{contract.tenant_name || '—'}</TableCell>
                        <TableCell className="text-muted-foreground">
                          {contract.room_number || '—'}
                          {contract.dormitory_name ? ` (${contract.dormitory_name})` : ''}
                        </TableCell>
                        <TableCell className="text-muted-foreground">
                          {toDateInputValue(contract.start_date)}
                        </TableCell>
                        <TableCell className="text-muted-foreground">
                          {toDateInputValue(contract.end_date) || '—'}
                        </TableCell>
                        <TableCell className="text-muted-foreground">
                          {contract.rent_price.toLocaleString()}
                        </TableCell>
                        <TableCell className="text-muted-foreground">
                          {contract.deposit.toLocaleString()}
                        </TableCell>
                        <TableCell>
                          <Badge variant={contractStatusBadgeVariant[contract.status]}>
                            {t(contractStatusLabelKeys[contract.status])}
                          </Badge>
                        </TableCell>
                        <TableCell className="text-right">
                          <div className="flex flex-wrap justify-end gap-2">
                            <Button
                              type="button"
                              size="icon"
                              variant="outline"
                              title={t('contractEdit')}
                              aria-label={t('contractEdit')}
                              onClick={() => onEditContract(contract)}
                            >
                              <Pencil />
                            </Button>
                            <Button
                              type="button"
                              size="icon"
                              variant="destructive"
                              title={t('contractDelete')}
                              aria-label={t('contractDelete')}
                              onClick={() => onDeleteContract(contract)}
                              disabled={deletingContractId === contract.id}
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

            {filteredContracts.length > 0 && (
              <div className="flex items-center justify-between gap-3">
                <p className="text-sm text-muted-foreground">
                  {t('rolePermissionsShowingLabel')} {rangeStart}-{rangeEnd}{' '}
                  {t('rolePermissionsOfLabel')} {filteredContracts.length} {t('rolePermissionsResultsLabel')}
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
                      {CONTRACT_PAGE_SIZE_OPTIONS.map((size) => (
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
