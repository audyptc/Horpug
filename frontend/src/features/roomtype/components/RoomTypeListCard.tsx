import { ChevronLeft, ChevronRight, Pencil, Trash2 } from 'lucide-react'
import { useLanguage } from '@/shared/i18n/language'
import { Badge } from '@/shared/components/ui/badge'
import { Card, CardContent, CardHeader, CardTitle } from '@/shared/components/ui/card'
import { Table, TableBody, TableCell, TableHead, TableHeader, TableRow } from '@/shared/components/ui/table'
import { Button } from '@/shared/components/ui/button'
import type { ApiRoomType } from '../types'
import { ROOM_TYPE_PAGE_SIZE_OPTIONS } from '../utils'

type RoomTypeListCardProps = {
  isLoading: boolean
  loadError: string | null
  deleteError: string | null
  roomTypes: ApiRoomType[] | null
  query: string
  onQueryChange: (query: string) => void
  filteredRoomTypes: ApiRoomType[]
  paginatedRoomTypes: ApiRoomType[]
  currentPage: number
  totalPages: number
  rangeStart: number
  rangeEnd: number
  pageSize: number
  onPageSizeChange: (size: number) => void
  onPrevPage: () => void
  onNextPage: () => void
  deletingRoomTypeId: string | null
  onCreateRoomType: () => void
  onEditRoomType: (roomType: ApiRoomType) => void
  onDeleteRoomType: (roomType: ApiRoomType) => void
}

export function RoomTypeListCard({
  isLoading,
  loadError,
  deleteError,
  roomTypes,
  query,
  onQueryChange,
  filteredRoomTypes,
  paginatedRoomTypes,
  currentPage,
  totalPages,
  rangeStart,
  rangeEnd,
  pageSize,
  onPageSizeChange,
  onPrevPage,
  onNextPage,
  deletingRoomTypeId,
  onCreateRoomType,
  onEditRoomType,
  onDeleteRoomType,
}: RoomTypeListCardProps) {
  const { t } = useLanguage()

  return (
    <Card>
      <CardHeader className="flex flex-row items-start justify-between gap-4">
        <div>
          <CardTitle>{t('menuRoomTypes')}</CardTitle>
        </div>
        <Button onClick={onCreateRoomType} disabled={isLoading}>
          {t('roomTypeCreate')}
        </Button>
      </CardHeader>
      <CardContent className="flex flex-col gap-4">
        {loadError && <p className="resource-error">{loadError}</p>}
        {deleteError && <p className="resource-error">{deleteError}</p>}

        {!loadError && isLoading && <p className="metric-detail">{t('loading')}</p>}

        {!loadError && !isLoading && roomTypes && roomTypes.length === 0 && (
          <p className="metric-detail">{t('roomTypeNoRoomTypes')}</p>
        )}

        {!loadError && !isLoading && roomTypes && roomTypes.length > 0 && (
          <>
            <label className="flex w-full max-w-md flex-col gap-1.5 text-sm font-medium">
              {t('roomTypeSearchLabel')}
              <input
                type="search"
                className="h-10 rounded-md border border-input bg-transparent px-3 text-sm"
                placeholder={t('roomTypeSearchPlaceholder')}
                value={query}
                onChange={(event) => onQueryChange(event.target.value)}
              />
            </label>

            {filteredRoomTypes.length === 0 && (
              <p className="metric-detail">{t('roomTypeNoMatching')}</p>
            )}

            {filteredRoomTypes.length > 0 && (
              <div className="table-wrap">
                <Table>
                  <TableHeader>
                    <TableRow>
                      <TableHead>{t('roomTypeNameColumn')}</TableHead>
                      <TableHead>{t('roomTypeDormitoryColumn')}</TableHead>
                      <TableHead>{t('roomTypePriceColumn')}</TableHead>
                      <TableHead>{t('roomTypeStatusColumn')}</TableHead>
                      <TableHead className="text-right">{t('roomTypeActionsColumn')}</TableHead>
                    </TableRow>
                  </TableHeader>
                  <TableBody>
                    {paginatedRoomTypes.map((roomType) => (
                      <TableRow key={roomType.id}>
                        <TableCell className="font-semibold">{roomType.name}</TableCell>
                        <TableCell className="text-muted-foreground">
                          {roomType.dormitory_name || '—'}
                        </TableCell>
                        <TableCell className="text-muted-foreground">
                          {roomType.price.toLocaleString()}
                        </TableCell>
                        <TableCell>
                          <Badge variant={roomType.is_active ? 'default' : 'outline'}>
                            {roomType.is_active ? t('statusActive') : t('statusInactive')}
                          </Badge>
                        </TableCell>
                        <TableCell className="text-right">
                          <div className="flex flex-wrap justify-end gap-2">
                            <Button
                              type="button"
                              size="icon"
                              variant="outline"
                              title={t('roomTypeEdit')}
                              aria-label={t('roomTypeEdit')}
                              onClick={() => onEditRoomType(roomType)}
                            >
                              <Pencil />
                            </Button>
                            <Button
                              type="button"
                              size="icon"
                              variant="destructive"
                              title={t('roomTypeDelete')}
                              aria-label={t('roomTypeDelete')}
                              onClick={() => onDeleteRoomType(roomType)}
                              disabled={deletingRoomTypeId === roomType.id}
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

            {filteredRoomTypes.length > 0 && (
              <div className="flex items-center justify-between gap-3">
                <p className="text-sm text-muted-foreground">
                  {t('rolePermissionsShowingLabel')} {rangeStart}-{rangeEnd}{' '}
                  {t('rolePermissionsOfLabel')} {filteredRoomTypes.length} {t('rolePermissionsResultsLabel')}
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
                      {ROOM_TYPE_PAGE_SIZE_OPTIONS.map((size) => (
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
