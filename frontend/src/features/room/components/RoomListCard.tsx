import { ChevronLeft, ChevronRight, ChevronsLeft, ChevronsRight, Pencil, Trash2 } from 'lucide-react'
import { useLanguage, type TranslationKey } from '@/shared/i18n/language'
import { Badge } from '@/shared/components/ui/badge'
import { Card, CardContent, CardHeader, CardTitle } from '@/shared/components/ui/card'
import { Table, TableBody, TableCell, TableHead, TableHeader, TableRow } from '@/shared/components/ui/table'
import { Button } from '@/shared/components/ui/button'
import type { ApiRoom, RoomStatus } from '../types'
import { ROOM_PAGE_SIZE_OPTIONS } from '../utils'

const roomStatusLabelKeys: Record<RoomStatus, TranslationKey> = {
  available: 'roomStatusAvailable',
  occupied: 'roomStatusOccupied',
  maintenance: 'roomStatusMaintenance',
}

const roomStatusBadgeVariant: Record<RoomStatus, 'default' | 'outline' | 'destructive'> = {
  available: 'default',
  occupied: 'outline',
  maintenance: 'destructive',
}

type RoomListCardProps = {
  isLoading: boolean
  loadError: string | null
  deleteError: string | null
  rooms: ApiRoom[] | null
  query: string
  onQueryChange: (query: string) => void
  filteredRooms: ApiRoom[]
  paginatedRooms: ApiRoom[]
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
  deletingRoomId: string | null
  onCreateRoom: () => void
  onEditRoom: (room: ApiRoom) => void
  onDeleteRoom: (room: ApiRoom) => void
}

export function RoomListCard({
  isLoading,
  loadError,
  deleteError,
  rooms,
  query,
  onQueryChange,
  filteredRooms,
  paginatedRooms,
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
  deletingRoomId,
  onCreateRoom,
  onEditRoom,
  onDeleteRoom,
}: RoomListCardProps) {
  const { t } = useLanguage()

  return (
    <Card>
      <CardHeader className="flex flex-row items-start justify-between gap-4">
        <div>
          <CardTitle>{t('menuRooms')}</CardTitle>
        </div>
        <Button onClick={onCreateRoom} disabled={isLoading}>
          {t('roomCreate')}
        </Button>
      </CardHeader>
      <CardContent className="flex flex-col gap-4">
        {loadError && <p className="resource-error">{loadError}</p>}
        {deleteError && <p className="resource-error">{deleteError}</p>}

        {!loadError && isLoading && <p className="metric-detail">{t('loading')}</p>}

        {!loadError && !isLoading && rooms && rooms.length === 0 && (
          <p className="metric-detail">{t('roomNoRooms')}</p>
        )}

        {!loadError && !isLoading && rooms && rooms.length > 0 && (
          <>
            <label className="flex w-full max-w-md flex-col gap-1.5 text-sm font-medium">
              {t('roomSearchLabel')}
              <input
                type="search"
                className="h-10 rounded-md border border-input bg-transparent px-3 text-sm"
                placeholder={t('roomSearchPlaceholder')}
                value={query}
                onChange={(event) => onQueryChange(event.target.value)}
              />
            </label>

            {filteredRooms.length === 0 && <p className="metric-detail">{t('roomNoMatching')}</p>}

            {filteredRooms.length > 0 && (
              <div className="table-wrap room-table-wrap">
                <Table>
                  <TableHeader>
                    <TableRow>
                      <TableHead>{t('roomNumberColumn')}</TableHead>
                      <TableHead>{t('roomDormitoryColumn')}</TableHead>
                      <TableHead>{t('roomRoomTypeColumn')}</TableHead>
                      <TableHead>{t('roomFloorColumn')}</TableHead>
                      <TableHead>{t('roomStatusColumn')}</TableHead>
                      <TableHead>{t('roomActiveColumn')}</TableHead>
                      <TableHead className="text-right">{t('roomActionsColumn')}</TableHead>
                    </TableRow>
                  </TableHeader>
                  <TableBody>
                    {paginatedRooms.map((room) => (
                      <TableRow key={room.id}>
                        <TableCell className="font-semibold">{room.room_number}</TableCell>
                        <TableCell className="text-muted-foreground">
                          {room.dormitory_name || '—'}
                        </TableCell>
                        <TableCell className="text-muted-foreground">
                          {room.room_type_name || '—'}
                        </TableCell>
                        <TableCell className="text-muted-foreground">{room.floor}</TableCell>
                        <TableCell>
                          <Badge variant={roomStatusBadgeVariant[room.status]}>
                            {t(roomStatusLabelKeys[room.status])}
                          </Badge>
                        </TableCell>
                        <TableCell>
                          <Badge variant={room.is_active ? 'default' : 'outline'}>
                            {room.is_active ? t('statusActive') : t('statusInactive')}
                          </Badge>
                        </TableCell>
                        <TableCell className="text-right">
                          <div className="flex flex-wrap justify-end gap-2">
                            <Button
                              type="button"
                              size="icon"
                              variant="outline"
                              title={t('roomEdit')}
                              aria-label={t('roomEdit')}
                              onClick={() => onEditRoom(room)}
                            >
                              <Pencil />
                            </Button>
                            <Button
                              type="button"
                              size="icon"
                              variant="destructive"
                              title={t('roomDelete')}
                              aria-label={t('roomDelete')}
                              onClick={() => onDeleteRoom(room)}
                              disabled={deletingRoomId === room.id}
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

            {filteredRooms.length > 0 && (
              <div className="flex flex-col gap-3 sm:flex-row sm:items-center sm:justify-between">
                <p className="text-sm text-muted-foreground">
                  {t('rolePermissionsShowingLabel')} {rangeStart}-{rangeEnd}{' '}
                  {t('rolePermissionsOfLabel')} {filteredRooms.length} {t('rolePermissionsResultsLabel')}
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
                      {ROOM_PAGE_SIZE_OPTIONS.map((size) => (
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
