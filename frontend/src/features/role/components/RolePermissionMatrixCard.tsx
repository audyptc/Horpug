import { useLanguage } from '@/shared/i18n/language'
import { Badge } from '@/shared/components/ui/badge'
import { Card, CardContent, CardDescription, CardHeader, CardTitle } from '@/shared/components/ui/card'
import { Table, TableBody, TableCell, TableHead, TableHeader, TableRow } from '@/shared/components/ui/table'
import { Button } from '@/shared/components/ui/button'
import { Switch } from '@/shared/components/ui/switch'
import type { ApiMenu } from '@/features/menu/menus'
import type { ApiPermission, ApiRole } from '../types'
import { actionLabelKeys, menuLabel } from '../utils'

type RolePermissionMatrixCardProps = {
  selectedRole: ApiRole
  menuQuery: string
  onMenuQueryChange: (query: string) => void
  filteredMenus: ApiMenu[]
  sortedPermissions: ApiPermission[]
  matrix: Record<string, Set<string>>
  onToggleCell: (menuId: string, permissionId: string) => void
  onSetMenuPermissions: (menuId: string, permissionIds: string[]) => void
  onSetVisiblePermissions: (grantAll: boolean) => void
  onBack: () => void
  onEditRole: () => void
  onSave: () => void
  saving: boolean
  saveError: string | null
  saveSuccess: boolean
  hasUnsavedChanges: boolean
}

export function RolePermissionMatrixCard({
  selectedRole,
  menuQuery,
  onMenuQueryChange,
  filteredMenus,
  sortedPermissions,
  matrix,
  onToggleCell,
  onSetMenuPermissions,
  onSetVisiblePermissions,
  onBack,
  onEditRole,
  onSave,
  saving,
  saveError,
  saveSuccess,
  hasUnsavedChanges,
}: RolePermissionMatrixCardProps) {
  const { t } = useLanguage()

  const isRowFullySelected = (menuId: string) =>
    sortedPermissions.length > 0 && (matrix[menuId]?.size ?? 0) === sortedPermissions.length

  const allVisibleSelected =
    filteredMenus.length > 0 && sortedPermissions.length > 0 && filteredMenus.every((menu) => isRowFullySelected(menu.id))

  return (
    <Card>
      <CardHeader className="flex flex-row items-start justify-between gap-4">
        <div>
          <Button type="button" variant="ghost" size="sm" className="-ml-3 mb-2" onClick={onBack}>
            {t('rolePermissionsBackToRoles')}
          </Button>
          <CardTitle className="flex items-center gap-2">
            {selectedRole.name}
            <Badge variant={selectedRole.is_active ? 'default' : 'outline'}>
              {selectedRole.is_active ? t('statusActive') : t('statusInactive')}
            </Badge>
          </CardTitle>
          <CardDescription>
            {selectedRole.description || t('rolePermissionsDescriptionEmpty')}
          </CardDescription>
        </div>
        <Button type="button" variant="outline" onClick={onEditRole}>
          {t('rolePermissionsEditRole')}
        </Button>
      </CardHeader>
      <CardContent className="flex flex-col gap-4">
        <div className="flex flex-col gap-3 rounded-lg border border-border bg-muted/30 p-3 lg:flex-row lg:items-end lg:justify-between">
          <label className="flex w-full max-w-md flex-col gap-1.5 text-sm font-medium">
            {t('rolePermissionsSearchLabel')}
            <input
              type="search"
              className="h-10 rounded-md border border-input bg-transparent px-3 text-sm"
              placeholder={t('rolePermissionsSearchPlaceholder')}
              value={menuQuery}
              onChange={(event) => onMenuQueryChange(event.target.value)}
            />
          </label>

          <label className="flex items-center gap-2 text-sm font-medium">
            <Switch
              checked={allVisibleSelected}
              onCheckedChange={onSetVisiblePermissions}
              disabled={filteredMenus.length === 0 || sortedPermissions.length === 0}
            />
            {allVisibleSelected ? t('rolePermissionsClearVisible') : t('rolePermissionsSelectAllVisible')}
          </label>
        </div>

        <div className="table-wrap">
          <Table>
            <TableHeader>
              <TableRow>
                <TableHead>{t('rolePermissionsMenuColumn')}</TableHead>
                {sortedPermissions.map((permission) => (
                  <TableHead key={permission.id} className="text-center">
                    {actionLabelKeys[permission.name] ? t(actionLabelKeys[permission.name]) : permission.name}
                  </TableHead>
                ))}
              </TableRow>
            </TableHeader>
            <TableBody>
              {filteredMenus.map((menu) => (
                <TableRow key={menu.id}>
                  <TableCell className="align-top">
                    <div className="flex min-w-48 flex-col gap-2">
                      <div>
                        <p className="font-semibold">{menuLabel(menu, t)}</p>
                        <p className="text-xs text-muted-foreground">{menu.path}</p>
                      </div>
                      <label className="flex items-center gap-2 text-xs font-medium">
                        <Switch
                          checked={isRowFullySelected(menu.id)}
                          onCheckedChange={(checked) =>
                            onSetMenuPermissions(
                              menu.id,
                              checked ? sortedPermissions.map((permission) => permission.id) : []
                            )
                          }
                          disabled={sortedPermissions.length === 0}
                        />
                        {isRowFullySelected(menu.id) ? t('rolePermissionsClearRow') : t('rolePermissionsSelectAllRow')}
                      </label>
                    </div>
                  </TableCell>
                  {sortedPermissions.map((permission) => (
                    <TableCell key={permission.id} className="text-center">
                      <input
                        type="checkbox"
                        aria-label={`${menuLabel(menu, t)} - ${permission.name}`}
                        checked={matrix[menu.id]?.has(permission.id) ?? false}
                        onChange={() => onToggleCell(menu.id, permission.id)}
                        className="h-4 w-4 accent-primary"
                      />
                    </TableCell>
                  ))}
                </TableRow>
              ))}
              {filteredMenus.length === 0 && (
                <TableRow>
                  <TableCell colSpan={sortedPermissions.length + 1} className="py-6 text-center text-muted-foreground">
                    {t('rolePermissionsNoVisibleMenus')}
                  </TableCell>
                </TableRow>
              )}
            </TableBody>
          </Table>
        </div>

        <div className="flex items-center gap-3">
          <Button onClick={onSave} disabled={saving || !hasUnsavedChanges}>
            {saving ? t('rolePermissionsSaving') : t('rolePermissionsSave')}
          </Button>
          {saveSuccess && <p className="metric-detail">{t('rolePermissionsSaved')}</p>}
          {!saveSuccess && hasUnsavedChanges && <p className="metric-detail">{t('rolePermissionsUnsaved')}</p>}
          {saveError && <p className="resource-error">{saveError}</p>}
        </div>
      </CardContent>
    </Card>
  )
}
