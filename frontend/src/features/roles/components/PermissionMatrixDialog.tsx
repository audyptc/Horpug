import { useState, useEffect } from 'react'
import { useTranslation } from 'react-i18next'
import { ShieldCheck } from 'lucide-react'
import { Button } from '@/components/ui/button'
import {
  Dialog,
  DialogContent,
  DialogHeader,
  DialogTitle,
  DialogDescription,
  DialogFooter,
} from '@/components/ui/dialog'
import { roleService } from '@/features/roles/roleService'
import { useToast } from '@/components/ui/toast'
import type { ApiRole, ApiMenu, ApiPermission } from '@/types'
import { cn } from '@/lib/utils'

type PermMatrix = Record<string, Set<string>>

const roleColorMap: Record<string, string> = {
  admin: 'text-violet-600 bg-violet-500/10',
  manager: 'text-blue-600 bg-blue-500/10',
  editor: 'text-amber-600 bg-amber-500/10',
  viewer: 'text-slate-600 bg-slate-500/10',
}

function getRoleColor(roleName: string) {
  return roleColorMap[roleName.toLowerCase()] ?? 'text-slate-600 bg-slate-500/10'
}

interface Props {
  role: ApiRole | null
  open: boolean
  onOpenChange: (open: boolean) => void
}

export function PermissionMatrixDialog({ role, open, onOpenChange }: Props) {
  const { t } = useTranslation()
  const toast = useToast()

  const [allMenus, setAllMenus] = useState<ApiMenu[]>([])
  const [allPermissions, setAllPermissions] = useState<ApiPermission[]>([])
  const [matrix, setMatrix] = useState<PermMatrix>({})
  const [loading, setLoading] = useState(false)
  const [saving, setSaving] = useState(false)
  const [error, setError] = useState('')

  useEffect(() => {
    if (!open || !role) return
    setLoading(true)
    setError('')
    Promise.all([
      roleService.listMenus(),
      roleService.listPermissions(),
      roleService.getById(role.id),
    ])
      .then(([menus, permissions, roleDetail]) => {
        setAllMenus(menus)
        setAllPermissions(permissions)
        const m: PermMatrix = {}
        for (const menu of menus) {
          m[menu.id] = new Set<string>()
        }
        for (const rmp of roleDetail.menu_permissions ?? []) {
          if (!m[rmp.menu_id]) m[rmp.menu_id] = new Set()
          for (const p of rmp.permissions) {
            m[rmp.menu_id].add(p.id)
          }
        }
        setMatrix(m)
      })
      .catch(() => setError(t('roles.permLoadError')))
      .finally(() => setLoading(false))
  }, [open, role, t])

  function togglePermission(menuId: string, permId: string) {
    setMatrix((prev) => {
      const next = { ...prev }
      const set = new Set(next[menuId] ?? [])
      if (set.has(permId)) {
        set.delete(permId)
      } else {
        set.add(permId)
        const readPerm = allPermissions.find((p) => p.name.toLowerCase() === 'read')
        if (readPerm && permId !== readPerm.id) set.add(readPerm.id)
      }
      next[menuId] = set
      return next
    })
  }

  function toggleRowAll(menuId: string) {
    setMatrix((prev) => {
      const next = { ...prev }
      const currentSet = next[menuId] ?? new Set()
      const allChecked = allPermissions.every((p) => currentSet.has(p.id))
      next[menuId] = allChecked ? new Set() : new Set(allPermissions.map((p) => p.id))
      return next
    })
  }

  function toggleColAll(permId: string) {
    setMatrix((prev) => {
      const next = { ...prev }
      const allChecked = allMenus.every((m) => (next[m.id] ?? new Set()).has(permId))
      const readPerm = allPermissions.find((p) => p.name.toLowerCase() === 'read')
      for (const menu of allMenus) {
        const set = new Set(next[menu.id] ?? [])
        if (allChecked) {
          set.delete(permId)
        } else {
          set.add(permId)
          if (readPerm && permId !== readPerm.id) set.add(readPerm.id)
        }
        next[menu.id] = set
      }
      return next
    })
  }

  async function handleSave() {
    if (!role) return
    setSaving(true)
    try {
      const items = Object.entries(matrix)
        .filter(([, permSet]) => permSet.size > 0)
        .map(([menuId, permSet]) => ({
          menu_id: menuId,
          permission_ids: Array.from(permSet),
        }))
      await roleService.assignPermissions(role.id, { items })
      toast.success(t('roles.permSaveSuccess'))
      onOpenChange(false)
    } catch {
      toast.error(t('roles.permSaveError'))
    } finally {
      setSaving(false)
    }
  }

  return (
    <Dialog open={open} onOpenChange={onOpenChange}>
      <DialogContent className="max-w-3xl max-h-[85vh] flex flex-col">
        <DialogHeader>
          <DialogTitle className="flex items-center gap-2">
            <ShieldCheck className="w-5 h-5 text-primary" />
            {t('roles.permissionsFor')}{' '}
            <span
              className={cn(
                'px-2 py-0.5 rounded-md text-sm font-medium',
                role ? getRoleColor(role.name) : ''
              )}
            >
              {role?.name}
            </span>
          </DialogTitle>
          <DialogDescription>{t('roles.permissionsDesc')}</DialogDescription>
        </DialogHeader>

        <div className="flex-1 overflow-auto min-h-0">
          {loading ? (
            <div className="flex items-center justify-center py-16 text-sm text-muted-foreground">
              {t('common.loading')}
            </div>
          ) : error ? (
            <div className="rounded-md border border-destructive/30 bg-destructive/10 px-4 py-3 text-sm text-destructive">
              {error}
            </div>
          ) : allMenus.length === 0 ? (
            <p className="text-sm text-muted-foreground py-8 text-center">{t('roles.noMenus')}</p>
          ) : (
            <table className="w-full text-sm border-collapse">
              <thead className="sticky top-0 bg-background z-10">
                <tr className="border-b bg-muted/40">
                  <th className="text-left px-4 py-3 font-medium text-muted-foreground min-w-36">เมนู</th>
                  {allPermissions.map((perm) => (
                    <th key={perm.id} className="px-3 py-3 font-medium text-muted-foreground text-center min-w-20">
                      <div className="flex flex-col items-center gap-1">
                        <span className="capitalize">{perm.name}</span>
                        <input
                          type="checkbox"
                          title={`เลือกทั้งหมด: ${perm.name}`}
                          className="h-4 w-4 cursor-pointer accent-primary"
                          checked={allMenus.length > 0 && allMenus.every((m) => (matrix[m.id] ?? new Set()).has(perm.id))}
                          onChange={() => toggleColAll(perm.id)}
                        />
                      </div>
                    </th>
                  ))}
                  <th className="px-3 py-3 font-medium text-muted-foreground text-center min-w-20">
                    <div className="flex flex-col items-center gap-1">
                      <span>{t('roles.selectAll')}</span>
                      <span className="h-4 w-4 block" />
                    </div>
                  </th>
                </tr>
              </thead>
              <tbody>
                {allMenus.map((menu, i) => {
                  const menuPerms = matrix[menu.id] ?? new Set()
                  const allChecked = allPermissions.every((p) => menuPerms.has(p.id))
                  const someChecked = allPermissions.some((p) => menuPerms.has(p.id))
                  return (
                    <tr
                      key={menu.id}
                      className={cn(
                        'border-b transition-colors hover:bg-muted/20',
                        i === allMenus.length - 1 && 'border-0'
                      )}
                    >
                      <td className="px-4 py-3 font-medium">{menu.name}</td>
                      {allPermissions.map((perm) => (
                        <td key={perm.id} className="px-3 py-3 text-center">
                          <input
                            type="checkbox"
                            className="h-4 w-4 cursor-pointer accent-primary"
                            checked={menuPerms.has(perm.id)}
                            onChange={() => togglePermission(menu.id, perm.id)}
                          />
                        </td>
                      ))}
                      <td className="px-3 py-3 text-center">
                        <input
                          type="checkbox"
                          title={t('roles.selectAll')}
                          className="h-4 w-4 cursor-pointer accent-primary"
                          checked={allChecked}
                          ref={(el) => {
                            if (el) el.indeterminate = !allChecked && someChecked
                          }}
                          onChange={() => toggleRowAll(menu.id)}
                        />
                      </td>
                    </tr>
                  )
                })}
              </tbody>
            </table>
          )}
        </div>

        <DialogFooter className="border-t pt-4">
          <Button variant="outline" onClick={() => onOpenChange(false)}>
            {t('common.cancel')}
          </Button>
          <Button onClick={handleSave} disabled={saving || loading}>
            {saving ? t('common.loading') : t('roles.saveChanges')}
          </Button>
        </DialogFooter>
      </DialogContent>
    </Dialog>
  )
}
