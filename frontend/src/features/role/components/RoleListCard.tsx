import { ChevronLeft, ChevronRight, ChevronsLeft, ChevronsRight, KeyRound, Pencil, Trash2 } from 'lucide-react'
import { useLanguage } from '@/shared/i18n/language'
import { Badge } from '@/shared/components/ui/badge'
import { Card, CardContent, CardHeader, CardTitle } from '@/shared/components/ui/card'
import { Table, TableBody, TableCell, TableHead, TableHeader, TableRow } from '@/shared/components/ui/table'
import { Button } from '@/shared/components/ui/button'
import type { ApiRole } from '../types'
import { ROLE_PAGE_SIZE_OPTIONS } from '../utils'

type RoleListCardProps = {
  isLoading: boolean
  loadError: string | null
  deleteError: string | null
  roles: ApiRole[] | null
  roleQuery: string
  onRoleQueryChange: (query: string) => void
  filteredRoles: ApiRole[]
  paginatedRoles: ApiRole[]
  currentRolePage: number
  totalRolePages: number
  rolesRangeStart: number
  rolesRangeEnd: number
  rolePageSize: number
  onRolePageSizeChange: (size: number) => void
  onFirstPage: () => void
  onPrevPage: () => void
  onNextPage: () => void
  onLastPage: () => void
  deletingRoleId: string | null
  onCreateRole: () => void
  onManageRole: (role: ApiRole) => void
  onEditRole: (role: ApiRole) => void
  onDeleteRole: (role: ApiRole) => void
}

export function RoleListCard({
  isLoading,
  loadError,
  deleteError,
  roles,
  roleQuery,
  onRoleQueryChange,
  filteredRoles,
  paginatedRoles,
  currentRolePage,
  totalRolePages,
  rolesRangeStart,
  rolesRangeEnd,
  rolePageSize,
  onRolePageSizeChange,
  onFirstPage,
  onPrevPage,
  onNextPage,
  onLastPage,
  deletingRoleId,
  onCreateRole,
  onManageRole,
  onEditRole,
  onDeleteRole,
}: RoleListCardProps) {
  const { t } = useLanguage()

  return (
    <Card>
      <CardHeader className="flex flex-row items-start justify-between gap-4">
        <div>
          <CardTitle>{t('menuRoles')}</CardTitle>
        </div>
        <Button onClick={onCreateRole} disabled={isLoading}>
          {t('rolePermissionsCreateRole')}
        </Button>
      </CardHeader>
      <CardContent className="flex flex-col gap-4">
        {loadError && <p className="resource-error">{loadError}</p>}
        {deleteError && <p className="resource-error">{deleteError}</p>}

        {!loadError && isLoading && <p className="metric-detail">{t('loading')}</p>}

        {!loadError && !isLoading && roles && roles.length === 0 && (
          <p className="metric-detail">{t('rolePermissionsNoRoles')}</p>
        )}

        {!loadError && !isLoading && roles && roles.length > 0 && (
          <>
            <label className="flex w-full max-w-md flex-col gap-1.5 text-sm font-medium">
              {t('rolePermissionsRoleSearchLabel')}
              <input
                type="search"
                className="h-10 rounded-md border border-input bg-transparent px-3 text-sm"
                placeholder={t('rolePermissionsRoleSearchPlaceholder')}
                value={roleQuery}
                onChange={(event) => onRoleQueryChange(event.target.value)}
              />
            </label>

            {filteredRoles.length === 0 && (
              <p className="metric-detail">{t('rolePermissionsNoMatchingRoles')}</p>
            )}

            {filteredRoles.length > 0 && (
              <div className="table-wrap">
                <Table>
                  <TableHeader>
                    <TableRow>
                      <TableHead>{t('rolePermissionsNameColumn')}</TableHead>
                      <TableHead>{t('rolePermissionsDescriptionColumn')}</TableHead>
                      <TableHead>{t('rolePermissionsStatusColumn')}</TableHead>
                      <TableHead className="text-right">{t('rolePermissionsActionsColumn')}</TableHead>
                    </TableRow>
                  </TableHeader>
                  <TableBody>
                    {paginatedRoles.map((role) => (
                      <TableRow key={role.id}>
                        <TableCell className="font-semibold">
                          <div className="flex items-center gap-2">
                            {role.name}
                            {role.is_protected && (
                              <Badge variant="outline">{t('rolePermissionsProtected')}</Badge>
                            )}
                          </div>
                        </TableCell>
                        <TableCell className="text-muted-foreground">
                          {role.description || t('rolePermissionsDescriptionEmpty')}
                        </TableCell>
                        <TableCell>
                          <Badge variant={role.is_active ? 'default' : 'outline'}>
                            {role.is_active ? t('statusActive') : t('statusInactive')}
                          </Badge>
                        </TableCell>
                        <TableCell className="text-right">
                          <div className="flex flex-wrap justify-end gap-2">
                            <Button
                              type="button"
                              size="icon"
                              variant="outline"
                              title={t('rolePermissionsManage')}
                              aria-label={t('rolePermissionsManage')}
                              onClick={() => onManageRole(role)}
                            >
                              <KeyRound />
                            </Button>
                            <Button
                              type="button"
                              size="icon"
                              variant="outline"
                              title={role.is_protected ? t('rolePermissionsProtectedHint') : t('rolePermissionsEditRole')}
                              aria-label={t('rolePermissionsEditRole')}
                              onClick={() => onEditRole(role)}
                              disabled={role.is_protected}
                            >
                              <Pencil />
                            </Button>
                            <Button
                              type="button"
                              size="icon"
                              variant="destructive"
                              title={role.is_protected ? t('rolePermissionsProtectedHint') : t('rolePermissionsDeleteRole')}
                              aria-label={t('rolePermissionsDeleteRole')}
                              onClick={() => onDeleteRole(role)}
                              disabled={deletingRoleId === role.id || role.is_protected}
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

            {filteredRoles.length > 0 && (
              <div className="flex items-center justify-between gap-3">
                <p className="text-sm text-muted-foreground">
                  {t('rolePermissionsShowingLabel')} {rolesRangeStart}-{rolesRangeEnd}{' '}
                  {t('rolePermissionsOfLabel')} {filteredRoles.length} {t('rolePermissionsResultsLabel')}
                  {totalRolePages > 1 && (
                    <>
                      {' '}
                      · {t('rolePermissionsPageLabel')} {currentRolePage} / {totalRolePages}
                    </>
                  )}
                </p>
                <div className="flex items-center gap-3">
                  <label className="flex items-center gap-1.5 text-sm text-muted-foreground">
                    {t('rolePermissionsPageSizeLabel')}
                    <select
                      className="h-9 rounded-md border border-input bg-transparent px-2 text-sm"
                      value={rolePageSize}
                      onChange={(event) => onRolePageSizeChange(Number(event.target.value))}
                    >
                      {ROLE_PAGE_SIZE_OPTIONS.map((size) => (
                        <option key={size} value={size}>
                          {size}
                        </option>
                      ))}
                    </select>
                  </label>

                  {totalRolePages > 1 && (
                    <div className="flex gap-2">
                      <Button
                        type="button"
                        size="icon"
                        variant="outline"
                        title={t('rolePermissionsFirstPage')}
                        aria-label={t('rolePermissionsFirstPage')}
                        disabled={currentRolePage <= 1}
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
                        disabled={currentRolePage <= 1}
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
                        disabled={currentRolePage >= totalRolePages}
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
                        disabled={currentRolePage >= totalRolePages}
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
