import { ChevronLeft, ChevronRight, Pencil, Trash2 } from 'lucide-react'
import { useLanguage } from '@/shared/i18n/language'
import { Badge } from '@/shared/components/ui/badge'
import { Card, CardContent, CardHeader, CardTitle } from '@/shared/components/ui/card'
import { Table, TableBody, TableCell, TableHead, TableHeader, TableRow } from '@/shared/components/ui/table'
import { Button } from '@/shared/components/ui/button'
import type { ApiDormitory } from '../types'
import { DORMITORY_PAGE_SIZE_OPTIONS } from '../utils'

type DormitoryListCardProps = {
  isLoading: boolean
  loadError: string | null
  deleteError: string | null
  dormitories: ApiDormitory[] | null
  query: string
  onQueryChange: (query: string) => void
  filteredDormitories: ApiDormitory[]
  paginatedDormitories: ApiDormitory[]
  currentPage: number
  totalPages: number
  rangeStart: number
  rangeEnd: number
  pageSize: number
  onPageSizeChange: (size: number) => void
  onPrevPage: () => void
  onNextPage: () => void
  deletingDormitoryId: string | null
  onCreateDormitory: () => void
  onEditDormitory: (dormitory: ApiDormitory) => void
  onDeleteDormitory: (dormitory: ApiDormitory) => void
}

export function DormitoryListCard({
  isLoading,
  loadError,
  deleteError,
  dormitories,
  query,
  onQueryChange,
  filteredDormitories,
  paginatedDormitories,
  currentPage,
  totalPages,
  rangeStart,
  rangeEnd,
  pageSize,
  onPageSizeChange,
  onPrevPage,
  onNextPage,
  deletingDormitoryId,
  onCreateDormitory,
  onEditDormitory,
  onDeleteDormitory,
}: DormitoryListCardProps) {
  const { t } = useLanguage()

  return (
    <Card>
      <CardHeader className="flex flex-row items-start justify-between gap-4">
        <div>
          <CardTitle>{t('menuDormitories')}</CardTitle>
        </div>
        <Button onClick={onCreateDormitory} disabled={isLoading}>
          {t('dormitoryCreate')}
        </Button>
      </CardHeader>
      <CardContent className="flex flex-col gap-4">
        {loadError && <p className="resource-error">{loadError}</p>}
        {deleteError && <p className="resource-error">{deleteError}</p>}

        {!loadError && isLoading && <p className="metric-detail">{t('loading')}</p>}

        {!loadError && !isLoading && dormitories && dormitories.length === 0 && (
          <p className="metric-detail">{t('dormitoryNoDormitories')}</p>
        )}

        {!loadError && !isLoading && dormitories && dormitories.length > 0 && (
          <>
            <label className="flex w-full max-w-md flex-col gap-1.5 text-sm font-medium">
              {t('dormitorySearchLabel')}
              <input
                type="search"
                className="h-10 rounded-md border border-input bg-transparent px-3 text-sm"
                placeholder={t('dormitorySearchPlaceholder')}
                value={query}
                onChange={(event) => onQueryChange(event.target.value)}
              />
            </label>

            {filteredDormitories.length === 0 && (
              <p className="metric-detail">{t('dormitoryNoMatching')}</p>
            )}

            {filteredDormitories.length > 0 && (
              <div className="table-wrap">
                <Table>
                  <TableHeader>
                    <TableRow>
                      <TableHead>{t('dormitoryNameColumn')}</TableHead>
                      <TableHead>{t('dormitoryAddressColumn')}</TableHead>
                      <TableHead>{t('dormitoryPhoneColumn')}</TableHead>
                      <TableHead>{t('dormitoryManagersColumn')}</TableHead>
                      <TableHead>{t('dormitoryStatusColumn')}</TableHead>
                      <TableHead className="text-right">{t('dormitoryActionsColumn')}</TableHead>
                    </TableRow>
                  </TableHeader>
                  <TableBody>
                    {paginatedDormitories.map((dormitory) => (
                      <TableRow key={dormitory.id}>
                        <TableCell className="font-semibold">{dormitory.name}</TableCell>
                        <TableCell className="text-muted-foreground">
                          {dormitory.address || t('rolePermissionsDescriptionEmpty')}
                        </TableCell>
                        <TableCell className="text-muted-foreground">{dormitory.phone || '—'}</TableCell>
                        <TableCell className="text-muted-foreground">
                          {dormitory.managers && dormitory.managers.length > 0
                            ? dormitory.managers.map((manager) => manager.username).join(', ')
                            : t('dormitoryManagersEmpty')}
                        </TableCell>
                        <TableCell>
                          <Badge variant={dormitory.is_active ? 'default' : 'outline'}>
                            {dormitory.is_active ? t('statusActive') : t('statusInactive')}
                          </Badge>
                        </TableCell>
                        <TableCell className="text-right">
                          <div className="flex flex-wrap justify-end gap-2">
                            <Button
                              type="button"
                              size="icon"
                              variant="outline"
                              title={t('dormitoryEdit')}
                              aria-label={t('dormitoryEdit')}
                              onClick={() => onEditDormitory(dormitory)}
                            >
                              <Pencil />
                            </Button>
                            <Button
                              type="button"
                              size="icon"
                              variant="destructive"
                              title={t('dormitoryDelete')}
                              aria-label={t('dormitoryDelete')}
                              onClick={() => onDeleteDormitory(dormitory)}
                              disabled={deletingDormitoryId === dormitory.id}
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

            {filteredDormitories.length > 0 && (
              <div className="flex items-center justify-between gap-3">
                <p className="text-sm text-muted-foreground">
                  {t('rolePermissionsShowingLabel')} {rangeStart}-{rangeEnd}{' '}
                  {t('rolePermissionsOfLabel')} {filteredDormitories.length} {t('rolePermissionsResultsLabel')}
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
                      {DORMITORY_PAGE_SIZE_OPTIONS.map((size) => (
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
