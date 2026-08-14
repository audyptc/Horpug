import { ChevronLeft, ChevronRight, Pencil, Trash2 } from 'lucide-react'
import { useLanguage, type TranslationKey } from '@/shared/i18n/language'
import { Badge } from '@/shared/components/ui/badge'
import { Card, CardContent, CardHeader, CardTitle } from '@/shared/components/ui/card'
import { Table, TableBody, TableCell, TableHead, TableHeader, TableRow } from '@/shared/components/ui/table'
import { Button } from '@/shared/components/ui/button'
import type { ApiParcel, ParcelStatus } from '../types'
import { PARCEL_PAGE_SIZE_OPTIONS, toDateInputValue } from '../utils'

const parcelStatusLabelKeys: Record<ParcelStatus, TranslationKey> = {
  pending: 'parcelStatusPending',
  picked_up: 'parcelStatusPickedUp',
  returned: 'parcelStatusReturned',
}

const parcelStatusBadgeVariant: Record<ParcelStatus, 'default' | 'outline' | 'destructive' | 'secondary'> = {
  pending: 'outline',
  picked_up: 'default',
  returned: 'destructive',
}

type ParcelListCardProps = {
  isLoading: boolean
  loadError: string | null
  deleteError: string | null
  parcels: ApiParcel[] | null
  query: string
  onQueryChange: (query: string) => void
  filteredParcels: ApiParcel[]
  paginatedParcels: ApiParcel[]
  currentPage: number
  totalPages: number
  rangeStart: number
  rangeEnd: number
  pageSize: number
  onPageSizeChange: (size: number) => void
  onPrevPage: () => void
  onNextPage: () => void
  deletingParcelId: string | null
  onCreateParcel: () => void
  onEditParcel: (parcel: ApiParcel) => void
  onDeleteParcel: (parcel: ApiParcel) => void
}

export function ParcelListCard({
  isLoading,
  loadError,
  deleteError,
  parcels,
  query,
  onQueryChange,
  filteredParcels,
  paginatedParcels,
  currentPage,
  totalPages,
  rangeStart,
  rangeEnd,
  pageSize,
  onPageSizeChange,
  onPrevPage,
  onNextPage,
  deletingParcelId,
  onCreateParcel,
  onEditParcel,
  onDeleteParcel,
}: ParcelListCardProps) {
  const { t } = useLanguage()

  return (
    <Card>
      <CardHeader className="flex flex-row items-start justify-between gap-4">
        <div>
          <CardTitle>{t('menuParcels')}</CardTitle>
        </div>
        <Button onClick={onCreateParcel} disabled={isLoading}>
          {t('parcelCreate')}
        </Button>
      </CardHeader>
      <CardContent className="flex flex-col gap-4">
        {loadError && <p className="resource-error">{loadError}</p>}
        {deleteError && <p className="resource-error">{deleteError}</p>}

        {!loadError && isLoading && <p className="metric-detail">{t('loading')}</p>}

        {!loadError && !isLoading && parcels && parcels.length === 0 && (
          <p className="metric-detail">{t('parcelNoParcels')}</p>
        )}

        {!loadError && !isLoading && parcels && parcels.length > 0 && (
          <>
            <label className="flex w-full max-w-md flex-col gap-1.5 text-sm font-medium">
              {t('parcelSearchLabel')}
              <input
                type="search"
                className="h-10 rounded-md border border-input bg-transparent px-3 text-sm"
                placeholder={t('parcelSearchPlaceholder')}
                value={query}
                onChange={(event) => onQueryChange(event.target.value)}
              />
            </label>

            {filteredParcels.length === 0 && <p className="metric-detail">{t('parcelNoMatching')}</p>}

            {filteredParcels.length > 0 && (
              <div className="table-wrap">
                <Table>
                  <TableHeader>
                    <TableRow>
                      <TableHead>{t('parcelTenantColumn')}</TableHead>
                      <TableHead>{t('parcelRoomColumn')}</TableHead>
                      <TableHead>{t('parcelCourierColumn')}</TableHead>
                      <TableHead>{t('parcelTrackingNumberColumn')}</TableHead>
                      <TableHead>{t('parcelStatusColumn')}</TableHead>
                      <TableHead>{t('parcelReceivedDateColumn')}</TableHead>
                      <TableHead className="text-right">{t('parcelActionsColumn')}</TableHead>
                    </TableRow>
                  </TableHeader>
                  <TableBody>
                    {paginatedParcels.map((parcel) => (
                      <TableRow key={parcel.id}>
                        <TableCell className="font-semibold">{parcel.tenant_name || '—'}</TableCell>
                        <TableCell className="text-muted-foreground">
                          {parcel.room_number
                            ? `${parcel.room_number}${parcel.dormitory_name ? ` (${parcel.dormitory_name})` : ''}`
                            : '—'}
                        </TableCell>
                        <TableCell className="text-muted-foreground">{parcel.courier || '—'}</TableCell>
                        <TableCell className="text-muted-foreground">{parcel.tracking_number || '—'}</TableCell>
                        <TableCell>
                          <Badge variant={parcelStatusBadgeVariant[parcel.status]}>
                            {t(parcelStatusLabelKeys[parcel.status])}
                          </Badge>
                        </TableCell>
                        <TableCell className="text-muted-foreground">
                          {toDateInputValue(parcel.received_date)}
                        </TableCell>
                        <TableCell className="text-right">
                          <div className="flex flex-wrap justify-end gap-2">
                            <Button
                              type="button"
                              size="icon"
                              variant="outline"
                              title={t('parcelEdit')}
                              aria-label={t('parcelEdit')}
                              onClick={() => onEditParcel(parcel)}
                            >
                              <Pencil />
                            </Button>
                            <Button
                              type="button"
                              size="icon"
                              variant="destructive"
                              title={t('parcelDelete')}
                              aria-label={t('parcelDelete')}
                              onClick={() => onDeleteParcel(parcel)}
                              disabled={deletingParcelId === parcel.id}
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

            {filteredParcels.length > 0 && (
              <div className="flex items-center justify-between gap-3">
                <p className="text-sm text-muted-foreground">
                  {t('rolePermissionsShowingLabel')} {rangeStart}-{rangeEnd}{' '}
                  {t('rolePermissionsOfLabel')} {filteredParcels.length} {t('rolePermissionsResultsLabel')}
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
                      {PARCEL_PAGE_SIZE_OPTIONS.map((size) => (
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
