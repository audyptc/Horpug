import { useState, useEffect } from 'react'
import { useTranslation } from 'react-i18next'
import { Building2 } from 'lucide-react'
import { Button } from '@/components/ui/button'
import { roleService } from '@/features/roles/roleService'
import {
  Dialog,
  DialogContent,
  DialogHeader,
  DialogTitle,
  DialogDescription,
  DialogFooter,
} from '@/components/ui/dialog'
import {
  Select,
  SelectContent,
  SelectItem,
  SelectTrigger,
  SelectValue,
} from '@/components/ui/select'
import { dormitoryService } from '@/features/dormitory/dormitoryService'
import { useToast } from '@/components/ui/toast'
import type { ApiUser, ApiDormitory, ApiDormitoryRoleAssignment, ApiRole } from '@/types'

interface Props {
  user: ApiUser | null
  open: boolean
  onOpenChange: (open: boolean) => void
}

export function AssignDormitoriesDialog({ user, open, onOpenChange }: Props) {
  const { t } = useTranslation()
  const toast = useToast()

  const [allDormitories, setAllDormitories] = useState<ApiDormitory[]>([])
  const [roles, setRoles] = useState<ApiRole[]>([])
  const [selected, setSelected] = useState<Map<string, string>>(new Map())
  const [loading, setLoading] = useState(false)
  const [saving, setSaving] = useState(false)
  const [error, setError] = useState('')

  useEffect(() => {
    if (!open || !user) return
    setLoading(true)
    setError('')
    Promise.all([dormitoryService.list(), dormitoryService.getForUser(user.id), roleService.listActive()])
      .then(([all, assigned, activeRoles]) => {
        setAllDormitories(all)
        setRoles(activeRoles)
        setSelected(new Map(assigned.map((item: ApiDormitoryRoleAssignment) => [item.dormitory_id, item.role_id])))
      })
      .catch(() => setError(t('dormitories.assignLoadError')))
      .finally(() => setLoading(false))
  }, [open, user, t])

  function toggle(id: string) {
    setSelected((prev) => {
      const next = new Map(prev)
      if (next.has(id)) {
        next.delete(id)
      } else if (roles[0]?.id) {
        next.set(id, roles[0].id)
      }
      return next
    })
  }

  function setRole(dormitoryId: string, roleId: string) {
    setSelected((prev) => {
      const next = new Map(prev)
      next.set(dormitoryId, roleId)
      return next
    })
  }

  function toggleAll() {
    setSelected((prev) => {
      if (prev.size === allDormitories.length) {
        return new Map()
      }

      const next = new Map<string, string>()
      const defaultRoleID = roles[0]?.id ?? ''
      allDormitories.forEach((d) => {
        next.set(d.id, prev.get(d.id) ?? defaultRoleID)
      })
      return next
    })
  }

  async function handleSave() {
    if (!user) return
    setSaving(true)
    try {
      await dormitoryService.assignToUser(user.id, {
        items: Array.from(selected.entries()).map(([dormitory_id, role_id]) => ({ dormitory_id, role_id })),
      })
      toast.success(t('dormitories.assignSaveSuccess'))
      onOpenChange(false)
    } catch {
      toast.error(t('dormitories.assignSaveError'))
    } finally {
      setSaving(false)
    }
  }

  const allChecked = allDormitories.length > 0 && selected.size === allDormitories.length
  const saveDisabled = saving || loading || (selected.size > 0 && Array.from(selected.values()).some((roleId) => !roleId))

  return (
    <Dialog open={open} onOpenChange={onOpenChange}>
      <DialogContent className="max-w-md max-h-[85vh] flex flex-col">
        <DialogHeader>
          <DialogTitle className="flex items-center gap-2">
            <Building2 className="w-5 h-5 text-primary" />
            {t('dormitories.assignTitle')} <span className="font-medium">{user?.full_name}</span>
          </DialogTitle>
          <DialogDescription>{t('dormitories.assignDesc')}</DialogDescription>
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
          ) : roles.length === 0 ? (
            <p className="text-sm text-muted-foreground py-8 text-center">{t('dormitories.assignNoRoles')}</p>
          ) : allDormitories.length === 0 ? (
            <p className="text-sm text-muted-foreground py-8 text-center">{t('dormitories.noDormitories')}</p>
          ) : (
            <div className="space-y-1">
              <label className="flex items-center gap-2 px-2 py-2 border-b cursor-pointer text-sm font-medium">
                <input
                  type="checkbox"
                  className="h-4 w-4 cursor-pointer accent-primary"
                  checked={allChecked}
                  disabled={roles.length === 0}
                  onChange={toggleAll}
                />
                {t('dormitories.assignSelectAll')}
              </label>
              {allDormitories.map((d) => (
                <label
                  key={d.id}
                  className="flex items-center gap-3 px-2 py-2 rounded-md hover:bg-muted/30 cursor-pointer text-sm"
                >
                  <input
                    type="checkbox"
                    className="h-4 w-4 cursor-pointer accent-primary"
                    checked={selected.has(d.id)}
                    onChange={() => toggle(d.id)}
                  />
                  <div className="flex-1 min-w-0">
                    <div>{d.name}</div>
                    <div className="text-xs text-muted-foreground truncate">{d.address}</div>
                  </div>
                  <div className="w-40 shrink-0" onClick={(e) => e.preventDefault()}>
                    <Select
                      value={selected.get(d.id) ?? ''}
                      onValueChange={(value) => setRole(d.id, value)}
                      disabled={!selected.has(d.id)}
                    >
                      <SelectTrigger>
                        <SelectValue placeholder={t('dormitories.assignRolePlaceholder')} />
                      </SelectTrigger>
                      <SelectContent>
                        {roles.map((role) => (
                          <SelectItem key={role.id} value={role.id}>
                            {role.name}
                          </SelectItem>
                        ))}
                      </SelectContent>
                    </Select>
                  </div>
                </label>
              ))}
            </div>
          )}
        </div>

        <DialogFooter className="border-t pt-4">
          <Button variant="outline" onClick={() => onOpenChange(false)}>
            {t('common.cancel')}
          </Button>
          <Button onClick={handleSave} disabled={saveDisabled}>
            {saving ? t('common.loading') : t('common.save')}
          </Button>
        </DialogFooter>
      </DialogContent>
    </Dialog>
  )
}
