import { useState, useEffect } from 'react'
import { useTranslation } from 'react-i18next'
import { Button } from '@/components/ui/button'
import { Input } from '@/components/ui/input'
import { Label } from '@/components/ui/label'
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
import { userService } from '@/features/users/userService'
import { useToast } from '@/components/ui/toast'
import type { ApiUser, ApiRole } from '@/types'

const emptyForm = {
  full_name: '',
  email: '',
  password: '',
  role_id: '',
  is_active: true,
}

interface Props {
  open: boolean
  onOpenChange: (open: boolean) => void
  editingUser: ApiUser | null
  roles: ApiRole[]
  onSaved: () => void
}

export function UserFormDialog({ open, onOpenChange, editingUser, roles, onSaved }: Props) {
  const { t } = useTranslation()
  const toast = useToast()
  const [form, setForm] = useState(emptyForm)
  const [saving, setSaving] = useState(false)
  const [saveError, setSaveError] = useState('')

  useEffect(() => {
    if (!open) return
    setSaveError('')
    if (editingUser) {
      setForm({
        full_name: editingUser.full_name,
        email: editingUser.email,
        password: '',
        role_id: editingUser.role?.id ?? '',
        is_active: editingUser.is_active,
      })
    } else {
      setForm(emptyForm)
    }
  }, [open, editingUser])

  async function handleSave() {
    if (!form.full_name || !form.email) return
    setSaving(true)
    setSaveError('')
    try {
      if (editingUser) {
        const payload: Record<string, unknown> = {
          full_name: form.full_name,
          is_active: form.is_active,
        }
        if (form.password) payload.password = form.password
        await userService.update(editingUser.id, payload)
        if (form.role_id && form.role_id !== editingUser.role?.id) {
          await userService.assignRole(editingUser.id, { role_id: form.role_id })
        }
        toast.success(t('users.editSuccess'))
      } else {
        if (!form.password) return
        const created = await userService.create({
          full_name: form.full_name,
          email: form.email,
          password: form.password,
        })
        if (form.role_id) {
          await userService.assignRole(created.id, { role_id: form.role_id })
        }
        toast.success(t('users.createSuccess'))
      }
      onOpenChange(false)
      onSaved()
    } catch (err: unknown) {
      const msg =
        (err as { response?: { data?: { message?: string } } })?.response?.data?.message ??
        t('users.saveError')
      setSaveError(msg)
    } finally {
      setSaving(false)
    }
  }

  return (
    <Dialog open={open} onOpenChange={onOpenChange}>
      <DialogContent className="max-w-md">
        <DialogHeader>
          <DialogTitle>{editingUser ? t('users.editUser') : t('users.createUser')}</DialogTitle>
          <DialogDescription>
            {editingUser ? t('users.editDesc') : t('users.createDesc')}
          </DialogDescription>
        </DialogHeader>
        <div className="grid gap-4 py-4">
          <div className="grid grid-cols-2 gap-4">
            <div className="col-span-2 space-y-1.5">
              <Label>{t('users.fullName')} *</Label>
              <Input
                placeholder="John Doe"
                value={form.full_name}
                onChange={(e) => setForm((f) => ({ ...f, full_name: e.target.value }))}
              />
            </div>
            {!editingUser && (
              <div className="col-span-2 space-y-1.5">
                <Label>{t('users.email')} *</Label>
                <Input
                  type="email"
                  placeholder="john@example.com"
                  value={form.email}
                  onChange={(e) => setForm((f) => ({ ...f, email: e.target.value }))}
                />
              </div>
            )}
            <div className="col-span-2 space-y-1.5">
              <Label>
                {t('users.password')}
                {!editingUser && ' *'}
                {editingUser && (
                  <span className="ml-1 text-xs text-muted-foreground">
                    ({t('users.passwordOptional')})
                  </span>
                )}
              </Label>
              <Input
                type="password"
                placeholder="••••••••"
                value={form.password}
                onChange={(e) => setForm((f) => ({ ...f, password: e.target.value }))}
              />
            </div>
            <div className="space-y-1.5">
              <Label>{t('users.role')}</Label>
              <Select
                value={form.role_id}
                onValueChange={(v) => setForm((f) => ({ ...f, role_id: v }))}
              >
                <SelectTrigger>
                  <SelectValue placeholder={t('users.selectRole')} />
                </SelectTrigger>
                <SelectContent>
                  {roles.map((r) => (
                    <SelectItem key={r.id} value={r.id}>
                      {r.name}
                    </SelectItem>
                  ))}
                </SelectContent>
              </Select>
            </div>
            <div className="space-y-1.5">
              <Label>{t('users.status')}</Label>
              <Select
                value={form.is_active ? 'active' : 'inactive'}
                onValueChange={(v) => setForm((f) => ({ ...f, is_active: v === 'active' }))}
              >
                <SelectTrigger>
                  <SelectValue />
                </SelectTrigger>
                <SelectContent>
                  <SelectItem value="active">{t('users.statuses.active')}</SelectItem>
                  <SelectItem value="inactive">{t('users.statuses.inactive')}</SelectItem>
                </SelectContent>
              </Select>
            </div>
          </div>
        </div>
        {saveError && (
          <p className="text-sm text-destructive px-1 pb-2">{saveError}</p>
        )}
        <DialogFooter>
          <Button variant="outline" onClick={() => onOpenChange(false)}>
            {t('common.cancel')}
          </Button>
          <Button
            onClick={handleSave}
            disabled={saving || !form.full_name || (!editingUser && (!form.email || !form.password))}
          >
            {saving ? t('common.loading') : editingUser ? t('users.saveChanges') : t('users.createUser')}
          </Button>
        </DialogFooter>
      </DialogContent>
    </Dialog>
  )
}
