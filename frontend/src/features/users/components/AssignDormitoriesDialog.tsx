import { useState, useEffect } from 'react'
import { useTranslation } from 'react-i18next'
import { Building2 } from 'lucide-react'
import { Button } from '@/components/ui/button'
import {
  Dialog,
  DialogContent,
  DialogHeader,
  DialogTitle,
  DialogDescription,
  DialogFooter,
} from '@/components/ui/dialog'
import { dormitoryService } from '@/features/dormitory/dormitoryService'
import { useToast } from '@/components/ui/toast'
import type { ApiUser, ApiDormitory } from '@/types'

interface Props {
  user: ApiUser | null
  open: boolean
  onOpenChange: (open: boolean) => void
}

export function AssignDormitoriesDialog({ user, open, onOpenChange }: Props) {
  const { t } = useTranslation()
  const toast = useToast()

  const [allDormitories, setAllDormitories] = useState<ApiDormitory[]>([])
  const [selected, setSelected] = useState<Set<string>>(new Set())
  const [loading, setLoading] = useState(false)
  const [saving, setSaving] = useState(false)
  const [error, setError] = useState('')

  useEffect(() => {
    if (!open || !user) return
    setLoading(true)
    setError('')
    Promise.all([dormitoryService.list(), dormitoryService.getForUser(user.id)])
      .then(([all, assigned]) => {
        setAllDormitories(all)
        setSelected(new Set(assigned.map((d) => d.id)))
      })
      .catch(() => setError(t('dormitories.assignLoadError')))
      .finally(() => setLoading(false))
  }, [open, user, t])

  function toggle(id: string) {
    setSelected((prev) => {
      const next = new Set(prev)
      if (next.has(id)) next.delete(id)
      else next.add(id)
      return next
    })
  }

  function toggleAll() {
    setSelected((prev) =>
      prev.size === allDormitories.length ? new Set() : new Set(allDormitories.map((d) => d.id))
    )
  }

  async function handleSave() {
    if (!user) return
    setSaving(true)
    try {
      await dormitoryService.assignToUser(user.id, { dormitory_ids: Array.from(selected) })
      toast.success(t('dormitories.assignSaveSuccess'))
      onOpenChange(false)
    } catch {
      toast.error(t('dormitories.assignSaveError'))
    } finally {
      setSaving(false)
    }
  }

  const allChecked = allDormitories.length > 0 && selected.size === allDormitories.length

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
          ) : allDormitories.length === 0 ? (
            <p className="text-sm text-muted-foreground py-8 text-center">{t('dormitories.noDormitories')}</p>
          ) : (
            <div className="space-y-1">
              <label className="flex items-center gap-2 px-2 py-2 border-b cursor-pointer text-sm font-medium">
                <input
                  type="checkbox"
                  className="h-4 w-4 cursor-pointer accent-primary"
                  checked={allChecked}
                  onChange={toggleAll}
                />
                {t('dormitories.assignSelectAll')}
              </label>
              {allDormitories.map((d) => (
                <label
                  key={d.id}
                  className="flex items-center gap-2 px-2 py-2 rounded-md hover:bg-muted/30 cursor-pointer text-sm"
                >
                  <input
                    type="checkbox"
                    className="h-4 w-4 cursor-pointer accent-primary"
                    checked={selected.has(d.id)}
                    onChange={() => toggle(d.id)}
                  />
                  {d.name}
                </label>
              ))}
            </div>
          )}
        </div>

        <DialogFooter className="border-t pt-4">
          <Button variant="outline" onClick={() => onOpenChange(false)}>
            {t('common.cancel')}
          </Button>
          <Button onClick={handleSave} disabled={saving || loading}>
            {saving ? t('common.loading') : t('common.save')}
          </Button>
        </DialogFooter>
      </DialogContent>
    </Dialog>
  )
}
