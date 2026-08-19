import { ChevronLeft, ChevronRight, ChevronsLeft, ChevronsRight, Pencil, Trash2 } from 'lucide-react'
import { useLanguage, type TranslationKey } from '@/shared/i18n/language'
import { Badge } from '@/shared/components/ui/badge'
import { Card, CardContent, CardHeader, CardTitle } from '@/shared/components/ui/card'
import { Table, TableBody, TableCell, TableHead, TableHeader, TableRow } from '@/shared/components/ui/table'
import { Button } from '@/shared/components/ui/button'
import type { ApiParking, VehicleType } from '../types'
import { PARKING_PAGE_SIZE_OPTIONS } from '../utils'

const vehicleTypeLabelKeys: Record<VehicleType, TranslationKey> = {
  car: 'parkingVehicleTypeCar',
  motorcycle: 'parkingVehicleTypeMotorcycle',
  other: 'parkingVehicleTypeOther',
}

type ParkingListCardProps = {
  isLoading: boolean
  loadError: string | null
  deleteError: string | null
  parkings: ApiParking[] | null
  query: string
  onQueryChange: (query: string) => void
  filteredParkings: ApiParking[]
  paginatedParkings: ApiParking[]
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
  deletingParkingId: string | null
  onCreateParking: () => void
  onEditParking: (parking: ApiParking) => void
  onDeleteParking: (parking: ApiParking) => void
}

export function ParkingListCard({
  isLoading,
  loadError,
  deleteError,
  parkings,
  query,
  onQueryChange,
  filteredParkings,
  paginatedParkings,
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
  deletingParkingId,
  onCreateParking,
  onEditParking,
  onDeleteParking,
}: ParkingListCardProps) {
  const { t } = useLanguage()

  return (
    <Card>
      <CardHeader className="flex flex-row items-start justify-between gap-4">
        <div>
          <CardTitle>{t('menuParking')}</CardTitle>
        </div>
        <Button onClick={onCreateParking} disabled={isLoading}>
          {t('parkingCreate')}
        </Button>
      </CardHeader>
      <CardContent className="flex flex-col gap-4">
        {loadError && <p className="resource-error">{loadError}</p>}
        {deleteError && <p className="resource-error">{deleteError}</p>}

        {!loadError && isLoading && <p className="metric-detail">{t('loading')}</p>}

        {!loadError && !isLoading && parkings && parkings.length === 0 && (
          <p className="metric-detail">{t('parkingNoRegistrations')}</p>
        )}

        {!loadError && !isLoading && parkings && parkings.length > 0 && (
          <>
            <label className="flex w-full max-w-md flex-col gap-1.5 text-sm font-medium">
              {t('parkingSearchLabel')}
              <input
                type="search"
                className="h-10 rounded-md border border-input bg-transparent px-3 text-sm"
                placeholder={t('parkingSearchPlaceholder')}
                value={query}
                onChange={(event) => onQueryChange(event.target.value)}
              />
            </label>

            {filteredParkings.length === 0 && <p className="metric-detail">{t('parkingNoMatching')}</p>}

            {filteredParkings.length > 0 && (
              <div className="table-wrap parking-table-wrap">
                <Table>
                  <TableHeader>
                    <TableRow>
                      <TableHead>{t('parkingTenantColumn')}</TableHead>
                      <TableHead>{t('parkingRoomColumn')}</TableHead>
                      <TableHead>{t('parkingVehicleTypeColumn')}</TableHead>
                      <TableHead>{t('parkingLicensePlateColumn')}</TableHead>
                      <TableHead>{t('parkingParkingSpotColumn')}</TableHead>
                      <TableHead className="text-right">{t('parkingActionsColumn')}</TableHead>
                    </TableRow>
                  </TableHeader>
                  <TableBody>
                    {paginatedParkings.map((parking) => (
                      <TableRow key={parking.id}>
                        <TableCell className="font-semibold">{parking.tenant_name || '—'}</TableCell>
                        <TableCell className="text-muted-foreground">
                          {parking.room_number
                            ? `${parking.room_number}${parking.dormitory_name ? ` (${parking.dormitory_name})` : ''}`
                            : '—'}
                        </TableCell>
                        <TableCell>
                          <Badge variant="secondary">{t(vehicleTypeLabelKeys[parking.vehicle_type])}</Badge>
                        </TableCell>
                        <TableCell className="text-muted-foreground">{parking.license_plate || '—'}</TableCell>
                        <TableCell className="text-muted-foreground">{parking.parking_spot || '—'}</TableCell>
                        <TableCell className="text-right">
                          <div className="flex flex-wrap justify-end gap-2">
                            <Button
                              type="button"
                              size="icon"
                              variant="outline"
                              title={t('parkingEdit')}
                              aria-label={t('parkingEdit')}
                              onClick={() => onEditParking(parking)}
                            >
                              <Pencil />
                            </Button>
                            <Button
                              type="button"
                              size="icon"
                              variant="destructive"
                              title={t('parkingDelete')}
                              aria-label={t('parkingDelete')}
                              onClick={() => onDeleteParking(parking)}
                              disabled={deletingParkingId === parking.id}
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

            {filteredParkings.length > 0 && (
              <div className="flex flex-col gap-3 sm:flex-row sm:items-center sm:justify-between">
                <p className="text-sm text-muted-foreground">
                  {t('rolePermissionsShowingLabel')} {rangeStart}-{rangeEnd}{' '}
                  {t('rolePermissionsOfLabel')} {filteredParkings.length} {t('rolePermissionsResultsLabel')}
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
                      {PARKING_PAGE_SIZE_OPTIONS.map((size) => (
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
