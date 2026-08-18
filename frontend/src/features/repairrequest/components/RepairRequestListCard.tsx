import { ChevronLeft, ChevronRight, ChevronsLeft, ChevronsRight, Pencil, Trash2 } from 'lucide-react'
import { useLanguage, type TranslationKey } from '@/shared/i18n/language'
import { Badge } from '@/shared/components/ui/badge'
import { Card, CardContent, CardHeader, CardTitle } from '@/shared/components/ui/card'
import { Table, TableBody, TableCell, TableHead, TableHeader, TableRow } from '@/shared/components/ui/table'
import { Button } from '@/shared/components/ui/button'
import type { ApiRepairRequest, RepairCategory, RepairStatus } from '../types'
import { REPAIR_REQUEST_PAGE_SIZE_OPTIONS, toDateInputValue } from '../utils'

const repairCategoryLabelKeys: Record<RepairCategory, TranslationKey> = {
  electrical: 'repairCategoryElectrical',
  plumbing: 'repairCategoryPlumbing',
  furniture: 'repairCategoryFurniture',
  aircon: 'repairCategoryAircon',
  other: 'repairCategoryOther',
}

const repairStatusLabelKeys: Record<RepairStatus, TranslationKey> = {
  pending: 'repairStatusPending',
  in_progress: 'repairStatusInProgress',
  completed: 'repairStatusCompleted',
  cancelled: 'repairStatusCancelled',
}

const repairStatusBadgeVariant: Record<RepairStatus, 'default' | 'outline' | 'destructive' | 'secondary'> = {
  pending: 'outline',
  in_progress: 'secondary',
  completed: 'default',
  cancelled: 'destructive',
}

type RepairRequestListCardProps = {
  isLoading: boolean
  loadError: string | null
  deleteError: string | null
  repairRequests: ApiRepairRequest[] | null
  query: string
  onQueryChange: (query: string) => void
  filteredRepairRequests: ApiRepairRequest[]
  paginatedRepairRequests: ApiRepairRequest[]
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
  deletingRepairRequestId: string | null
  onCreateRepairRequest: () => void
  onEditRepairRequest: (repairRequest: ApiRepairRequest) => void
  onDeleteRepairRequest: (repairRequest: ApiRepairRequest) => void
}

export function RepairRequestListCard({
  isLoading,
  loadError,
  deleteError,
  repairRequests,
  query,
  onQueryChange,
  filteredRepairRequests,
  paginatedRepairRequests,
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
  deletingRepairRequestId,
  onCreateRepairRequest,
  onEditRepairRequest,
  onDeleteRepairRequest,
}: RepairRequestListCardProps) {
  const { t } = useLanguage()

  return (
    <Card>
      <CardHeader className="flex flex-row items-start justify-between gap-4">
        <div>
          <CardTitle>{t('menuRepairRequests')}</CardTitle>
        </div>
        <Button onClick={onCreateRepairRequest} disabled={isLoading}>
          {t('repairCreate')}
        </Button>
      </CardHeader>
      <CardContent className="flex flex-col gap-4">
        {loadError && <p className="resource-error">{loadError}</p>}
        {deleteError && <p className="resource-error">{deleteError}</p>}

        {!loadError && isLoading && <p className="metric-detail">{t('loading')}</p>}

        {!loadError && !isLoading && repairRequests && repairRequests.length === 0 && (
          <p className="metric-detail">{t('repairNoRequests')}</p>
        )}

        {!loadError && !isLoading && repairRequests && repairRequests.length > 0 && (
          <>
            <label className="flex w-full max-w-md flex-col gap-1.5 text-sm font-medium">
              {t('repairSearchLabel')}
              <input
                type="search"
                className="h-10 rounded-md border border-input bg-transparent px-3 text-sm"
                placeholder={t('repairSearchPlaceholder')}
                value={query}
                onChange={(event) => onQueryChange(event.target.value)}
              />
            </label>

            {filteredRepairRequests.length === 0 && <p className="metric-detail">{t('repairNoMatching')}</p>}

            {filteredRepairRequests.length > 0 && (
              <div className="table-wrap">
                <Table>
                  <TableHeader>
                    <TableRow>
                      <TableHead>{t('repairRoomColumn')}</TableHead>
                      <TableHead>{t('repairTenantColumn')}</TableHead>
                      <TableHead>{t('repairCategoryColumn')}</TableHead>
                      <TableHead>{t('repairDescriptionColumn')}</TableHead>
                      <TableHead>{t('repairStatusColumn')}</TableHead>
                      <TableHead>{t('repairReportedDateColumn')}</TableHead>
                      <TableHead className="text-right">{t('repairActionsColumn')}</TableHead>
                    </TableRow>
                  </TableHeader>
                  <TableBody>
                    {paginatedRepairRequests.map((repairRequest) => (
                      <TableRow key={repairRequest.id}>
                        <TableCell className="font-semibold">
                          {repairRequest.room_number || '—'}
                          {repairRequest.dormitory_name ? ` (${repairRequest.dormitory_name})` : ''}
                        </TableCell>
                        <TableCell className="text-muted-foreground">
                          {repairRequest.tenant_name || '—'}
                        </TableCell>
                        <TableCell>
                          <Badge variant="secondary">{t(repairCategoryLabelKeys[repairRequest.category])}</Badge>
                        </TableCell>
                        <TableCell className="text-muted-foreground">
                          {repairRequest.description || '—'}
                        </TableCell>
                        <TableCell>
                          <Badge variant={repairStatusBadgeVariant[repairRequest.status]}>
                            {t(repairStatusLabelKeys[repairRequest.status])}
                          </Badge>
                        </TableCell>
                        <TableCell className="text-muted-foreground">
                          {toDateInputValue(repairRequest.reported_date)}
                        </TableCell>
                        <TableCell className="text-right">
                          <div className="flex flex-wrap justify-end gap-2">
                            <Button
                              type="button"
                              size="icon"
                              variant="outline"
                              title={t('repairEdit')}
                              aria-label={t('repairEdit')}
                              onClick={() => onEditRepairRequest(repairRequest)}
                            >
                              <Pencil />
                            </Button>
                            <Button
                              type="button"
                              size="icon"
                              variant="destructive"
                              title={t('repairDelete')}
                              aria-label={t('repairDelete')}
                              onClick={() => onDeleteRepairRequest(repairRequest)}
                              disabled={deletingRepairRequestId === repairRequest.id}
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

            {filteredRepairRequests.length > 0 && (
              <div className="flex items-center justify-between gap-3">
                <p className="text-sm text-muted-foreground">
                  {t('rolePermissionsShowingLabel')} {rangeStart}-{rangeEnd}{' '}
                  {t('rolePermissionsOfLabel')} {filteredRepairRequests.length} {t('rolePermissionsResultsLabel')}
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
                      {REPAIR_REQUEST_PAGE_SIZE_OPTIONS.map((size) => (
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
