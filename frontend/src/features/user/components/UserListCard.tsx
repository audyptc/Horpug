import { ChevronLeft, ChevronRight, Pencil, Trash2 } from 'lucide-react'
import { useLanguage } from '@/shared/i18n/language'
import { Badge } from '@/shared/components/ui/badge'
import { Card, CardContent, CardHeader, CardTitle } from '@/shared/components/ui/card'
import { Table, TableBody, TableCell, TableHead, TableHeader, TableRow } from '@/shared/components/ui/table'
import { Button } from '@/shared/components/ui/button'
import type { ApiUser } from '../types'
import { USER_PAGE_SIZE_OPTIONS } from '../utils'

type UserListCardProps = {
  isLoading: boolean
  loadError: string | null
  deleteError: string | null
  users: ApiUser[] | null
  query: string
  onQueryChange: (query: string) => void
  filteredUsers: ApiUser[]
  paginatedUsers: ApiUser[]
  currentPage: number
  totalPages: number
  rangeStart: number
  rangeEnd: number
  pageSize: number
  onPageSizeChange: (size: number) => void
  onPrevPage: () => void
  onNextPage: () => void
  deletingUserId: string | null
  onCreateUser: () => void
  onEditUser: (user: ApiUser) => void
  onDeleteUser: (user: ApiUser) => void
}

export function UserListCard({
  isLoading,
  loadError,
  deleteError,
  users,
  query,
  onQueryChange,
  filteredUsers,
  paginatedUsers,
  currentPage,
  totalPages,
  rangeStart,
  rangeEnd,
  pageSize,
  onPageSizeChange,
  onPrevPage,
  onNextPage,
  deletingUserId,
  onCreateUser,
  onEditUser,
  onDeleteUser,
}: UserListCardProps) {
  const { t } = useLanguage()

  return (
    <Card>
      <CardHeader className="flex flex-row items-start justify-between gap-4">
        <div>
          <CardTitle>{t('menuUsers')}</CardTitle>
        </div>
        <Button onClick={onCreateUser} disabled={isLoading}>
          {t('userCreate')}
        </Button>
      </CardHeader>
      <CardContent className="flex flex-col gap-4">
        {loadError && <p className="resource-error">{loadError}</p>}
        {deleteError && <p className="resource-error">{deleteError}</p>}

        {!loadError && isLoading && <p className="metric-detail">{t('loading')}</p>}

        {!loadError && !isLoading && users && users.length === 0 && (
          <p className="metric-detail">{t('userNoUsers')}</p>
        )}

        {!loadError && !isLoading && users && users.length > 0 && (
          <>
            <label className="flex w-full max-w-md flex-col gap-1.5 text-sm font-medium">
              {t('userSearchLabel')}
              <input
                type="search"
                className="h-10 rounded-md border border-input bg-transparent px-3 text-sm"
                placeholder={t('userSearchPlaceholder')}
                value={query}
                onChange={(event) => onQueryChange(event.target.value)}
              />
            </label>

            {filteredUsers.length === 0 && <p className="metric-detail">{t('userNoMatching')}</p>}

            {filteredUsers.length > 0 && (
              <div className="table-wrap">
                <Table>
                  <TableHeader>
                    <TableRow>
                      <TableHead>{t('userUsernameColumn')}</TableHead>
                      <TableHead>{t('userEmailColumn')}</TableHead>
                      <TableHead>{t('userRoleColumn')}</TableHead>
                      <TableHead>{t('userStatusColumn')}</TableHead>
                      <TableHead className="text-right">{t('userActionsColumn')}</TableHead>
                    </TableRow>
                  </TableHeader>
                  <TableBody>
                    {paginatedUsers.map((user) => (
                      <TableRow key={user.id}>
                        <TableCell className="font-semibold">{user.username}</TableCell>
                        <TableCell className="text-muted-foreground">{user.email}</TableCell>
                        <TableCell className="text-muted-foreground">
                          {user.role?.name ?? '—'}
                        </TableCell>
                        <TableCell>
                          <Badge variant={user.is_active ? 'default' : 'outline'}>
                            {user.is_active ? t('statusActive') : t('statusInactive')}
                          </Badge>
                        </TableCell>
                        <TableCell className="text-right">
                          <div className="flex flex-wrap justify-end gap-2">
                            <Button
                              type="button"
                              size="icon"
                              variant="outline"
                              title={t('userEdit')}
                              aria-label={t('userEdit')}
                              onClick={() => onEditUser(user)}
                            >
                              <Pencil />
                            </Button>
                            <Button
                              type="button"
                              size="icon"
                              variant="destructive"
                              title={t('userDelete')}
                              aria-label={t('userDelete')}
                              onClick={() => onDeleteUser(user)}
                              disabled={deletingUserId === user.id}
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

            {filteredUsers.length > 0 && (
              <div className="flex items-center justify-between gap-3">
                <p className="text-sm text-muted-foreground">
                  {t('rolePermissionsShowingLabel')} {rangeStart}-{rangeEnd}{' '}
                  {t('rolePermissionsOfLabel')} {filteredUsers.length} {t('rolePermissionsResultsLabel')}
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
                      {USER_PAGE_SIZE_OPTIONS.map((size) => (
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
