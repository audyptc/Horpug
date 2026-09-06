import { CheckCircle2, ChevronLeft, ChevronRight, ChevronsLeft, ChevronsRight, Link2, Pencil, Trash2, Unlink } from 'lucide-react'
import { useLanguage } from '@/shared/i18n/language'
import { Badge } from '@/shared/components/ui/badge'
import { Card, CardContent, CardHeader, CardTitle } from '@/shared/components/ui/card'
import { Table, TableBody, TableCell, TableHead, TableHeader, TableRow } from '@/shared/components/ui/table'
import { Button } from '@/shared/components/ui/button'
import type { ApiTenant } from '../types'
import { TENANT_PAGE_SIZE_OPTIONS } from '../utils'

type TenantListCardProps = {
  isLoading: boolean
  loadError: string | null
  deleteError: string | null
  tenants: ApiTenant[] | null
  query: string
  onQueryChange: (query: string) => void
  filteredTenants: ApiTenant[]
  paginatedTenants: ApiTenant[]
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
  deletingTenantId: string | null
  onCreateTenant: () => void
  onEditTenant: (tenant: ApiTenant) => void
  onDeleteTenant: (tenant: ApiTenant) => void
  onCopyLineLink: (tenant: ApiTenant) => void
  onUnlinkLine: (tenant: ApiTenant) => void
}

export function TenantListCard({
  isLoading,
  loadError,
  deleteError,
  tenants,
  query,
  onQueryChange,
  filteredTenants,
  paginatedTenants,
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
  deletingTenantId,
  onCreateTenant,
  onEditTenant,
  onDeleteTenant,
  onCopyLineLink,
  onUnlinkLine,
}: TenantListCardProps) {
  const { t } = useLanguage()

  return (
    <Card>
      <CardHeader className="flex flex-row items-start justify-between gap-4">
        <div>
          <CardTitle>{t('menuTenants')}</CardTitle>
        </div>
        <Button onClick={onCreateTenant} disabled={isLoading}>
          {t('tenantCreate')}
        </Button>
      </CardHeader>
      <CardContent className="flex flex-col gap-4">
        {loadError && <p className="resource-error">{loadError}</p>}
        {deleteError && <p className="resource-error">{deleteError}</p>}

        {!loadError && isLoading && <p className="metric-detail">{t('loading')}</p>}

        {!loadError && !isLoading && tenants && tenants.length === 0 && (
          <p className="metric-detail">{t('tenantNoTenants')}</p>
        )}

        {!loadError && !isLoading && tenants && tenants.length > 0 && (
          <>
            <label className="flex w-full max-w-md flex-col gap-1.5 text-sm font-medium">
              {t('tenantSearchLabel')}
              <input
                type="search"
                className="h-10 rounded-md border border-input bg-transparent px-3 text-sm"
                placeholder={t('tenantSearchPlaceholder')}
                value={query}
                onChange={(event) => onQueryChange(event.target.value)}
              />
            </label>

            {filteredTenants.length === 0 && <p className="metric-detail">{t('tenantNoMatching')}</p>}

            {filteredTenants.length > 0 && (
              <div className="table-wrap tenant-table-wrap">
                <Table>
                  <TableHeader>
                    <TableRow>
                      <TableHead>{t('tenantFirstNameColumn')}</TableHead>
                      <TableHead>{t('tenantLastNameColumn')}</TableHead>
                      <TableHead>{t('tenantPhoneColumn')}</TableHead>
                      <TableHead>{t('tenantLineIdColumn')}</TableHead>
                      <TableHead>{t('tenantIdCardColumn')}</TableHead>
                      <TableHead>{t('tenantEmailColumn')}</TableHead>
                      <TableHead>{t('tenantActiveColumn')}</TableHead>
                      <TableHead className="text-right">{t('tenantActionsColumn')}</TableHead>
                    </TableRow>
                  </TableHeader>
                  <TableBody>
                    {paginatedTenants.map((tenant) => (
                      <TableRow key={tenant.id}>
                        <TableCell className="font-semibold">{tenant.first_name}</TableCell>
                        <TableCell className="font-semibold">{tenant.last_name}</TableCell>
                        <TableCell className="text-muted-foreground">{tenant.phone || '—'}</TableCell>
                        <TableCell className="text-muted-foreground">
                          <div className="flex items-center gap-1.5">
                            {tenant.line_id || '—'}
                            {tenant.line_user_id && (
                              <CheckCircle2
                                className="h-4 w-4 text-emerald-600"
                                aria-label={t('tenantLineLinked')}
                              >
                                <title>{t('tenantLineLinked')}</title>
                              </CheckCircle2>
                            )}
                          </div>
                        </TableCell>
                        <TableCell className="text-muted-foreground">{tenant.id_card || '—'}</TableCell>
                        <TableCell className="text-muted-foreground">{tenant.email || '—'}</TableCell>
                        <TableCell>
                          <Badge variant={tenant.is_active ? 'default' : 'outline'}>
                            {tenant.is_active ? t('statusActive') : t('statusInactive')}
                          </Badge>
                        </TableCell>
                        <TableCell className="text-right">
                          <div className="flex flex-wrap justify-end gap-2">
                            <Button
                              type="button"
                              size="icon"
                              variant="outline"
                              title={t('tenantCopyLineLink')}
                              aria-label={t('tenantCopyLineLink')}
                              onClick={() => onCopyLineLink(tenant)}
                            >
                              <Link2 />
                            </Button>
                            {tenant.line_user_id && (
                              <Button
                                type="button"
                                size="icon"
                                variant="outline"
                                title={t('tenantUnlinkLine')}
                                aria-label={t('tenantUnlinkLine')}
                                onClick={() => onUnlinkLine(tenant)}
                              >
                                <Unlink />
                              </Button>
                            )}
                            <Button
                              type="button"
                              size="icon"
                              variant="outline"
                              title={t('tenantEdit')}
                              aria-label={t('tenantEdit')}
                              onClick={() => onEditTenant(tenant)}
                            >
                              <Pencil />
                            </Button>
                            <Button
                              type="button"
                              size="icon"
                              variant="destructive"
                              title={t('tenantDelete')}
                              aria-label={t('tenantDelete')}
                              onClick={() => onDeleteTenant(tenant)}
                              disabled={deletingTenantId === tenant.id}
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

            {filteredTenants.length > 0 && (
              <div className="flex flex-col gap-3 sm:flex-row sm:items-center sm:justify-between">
                <p className="text-sm text-muted-foreground">
                  {t('rolePermissionsShowingLabel')} {rangeStart}-{rangeEnd}{' '}
                  {t('rolePermissionsOfLabel')} {filteredTenants.length} {t('rolePermissionsResultsLabel')}
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
                      {TENANT_PAGE_SIZE_OPTIONS.map((size) => (
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
