import { useState, useEffect } from 'react'
import { useTranslation } from 'react-i18next'
import { Save, Palette, Sun, Moon, User2, KeyRound, Loader2, UserCircle } from 'lucide-react'
import { useTheme, SIDEBAR_COLORS } from '@/lib/theme'
import { Card, CardContent, CardHeader, CardTitle, CardDescription } from '@/components/ui/card'
import { Button } from '@/components/ui/button'
import { Input } from '@/components/ui/input'
import { Label } from '@/components/ui/label'
import { Badge } from '@/components/ui/badge'
import { Separator } from '@/components/ui/separator'
import {
  Select,
  SelectContent,
  SelectItem,
  SelectTrigger,
  SelectValue,
} from '@/components/ui/select'
import { useAuth } from '@/features/auth/AuthContext'
import { userService } from '@/features/users/userService'
import type { ApiUser } from '@/types'
import { cn } from '@/lib/utils'


export function Settings() {
  const { t, i18n } = useTranslation()
  const { theme, setTheme, sidebarColor, setSidebarColor } = useTheme()
  const { userId } = useAuth()
  const [activeSection, setActiveSection] = useState('profile')
  const [currentUser, setCurrentUser] = useState<ApiUser | null>(null)

  const [profile, setProfile] = useState({ name: '', email: '' })
  const [saving, setSaving] = useState(false)
  const [saved, setSaved] = useState(false)
  const [saveError, setSaveError] = useState('')

  const [password, setPassword] = useState({ newPass: '', confirm: '' })
  const [passwordSaving, setPasswordSaving] = useState(false)
  const [passwordSaved, setPasswordSaved] = useState(false)
  const [passwordError, setPasswordError] = useState('')

  useEffect(() => {
    if (!userId) return
    userService.getById(userId).then((user) => {
      setCurrentUser(user)
      setProfile({ name: user.full_name, email: user.email })
    }).catch(() => {})
  }, [userId])

  async function handleSave() {
    if (!userId) return
    setSaving(true)
    setSaveError('')
    try {
      await userService.update(userId, { full_name: profile.name })
      setCurrentUser((u) => u ? { ...u, full_name: profile.name } : u)
      setSaved(true)
      setTimeout(() => setSaved(false), 2000)
    } catch {
      setSaveError(t('common.errorRetry', { defaultValue: 'เกิดข้อผิดพลาด ลองใหม่อีกครั้ง' }))
    } finally {
      setSaving(false)
    }
  }

  async function handlePasswordChange() {
    setPasswordError('')
    if (!password.newPass) {
      setPasswordError(t('settings.passwordRequired', { defaultValue: 'กรุณากรอกรหัสผ่านใหม่' }))
      return
    }
    if (password.newPass.length < 6) {
      setPasswordError(t('settings.passwordTooShort', { defaultValue: 'รหัสผ่านต้องมีอย่างน้อย 6 ตัวอักษร' }))
      return
    }
    if (password.newPass !== password.confirm) {
      setPasswordError(t('settings.passwordMismatch', { defaultValue: 'รหัสผ่านใหม่ไม่ตรงกัน' }))
      return
    }
    if (!userId) return
    setPasswordSaving(true)
    try {
      await userService.update(userId, { password: password.newPass })
      setPassword({ newPass: '', confirm: '' })
      setPasswordSaved(true)
      setTimeout(() => setPasswordSaved(false), 2000)
    } catch {
      setPasswordError(t('common.errorRetry', { defaultValue: 'เกิดข้อผิดพลาด ลองใหม่อีกครั้ง' }))
    } finally {
      setPasswordSaving(false)
    }
  }

  function scrollTo(id: string) {
    setActiveSection(id)
    document.getElementById(id)?.scrollIntoView({ behavior: 'smooth', block: 'start' })
  }

  const navItems = [
    { id: 'profile', label: t('settings.profile'), icon: User2, color: 'text-violet-500' },
    { id: 'security', label: t('settings.security'), icon: KeyRound, color: 'text-emerald-500' },
    { id: 'appearance', label: t('settings.appearance'), icon: Palette, color: 'text-amber-500' },
  ]

  return (
    <div className="max-w-5xl mx-auto">
      {/* Header */}
      <div className="flex items-start justify-between gap-4 mb-8">
        <div>
          <h1 className="text-2xl font-bold tracking-tight">{t('settings.title')}</h1>
          <p className="text-muted-foreground text-sm mt-1">{t('settings.subtitle')}</p>
        </div>
        <Button onClick={handleSave} disabled={saving} className="gap-2 min-w-28 shrink-0">
          {saving ? <Loader2 className="w-4 h-4 animate-spin" /> : <Save className="w-4 h-4" />}
          {saved ? t('common.saved') : t('common.save')}
        </Button>
      </div>

      <div className="flex flex-col lg:flex-row gap-8 items-start">
        {/* Left sticky nav */}
        <aside className="w-full lg:w-52 shrink-0">
          <nav className="lg:sticky lg:top-6">
            <div className="flex lg:flex-col gap-1 flex-wrap">
              {navItems.map(({ id, label, icon: Icon, color }) => (
                <button
                  key={id}
                  type="button"
                  onClick={() => scrollTo(id)}
                  className={cn(
                    'flex items-center gap-2.5 px-3 py-2 rounded-lg text-sm font-medium transition-colors text-left w-full',
                    activeSection === id
                      ? 'bg-muted text-foreground'
                      : 'text-muted-foreground hover:text-foreground hover:bg-muted/60'
                  )}
                >
                  <Icon className={cn('w-4 h-4 shrink-0', activeSection === id ? color : '')} />
                  {label}
                </button>
              ))}
            </div>
          </nav>
        </aside>

        {/* Content */}
        <div className="flex-1 min-w-0 space-y-6">
          {/* Profile */}
          <Card id="profile" className="overflow-hidden">
            <div className="h-24 bg-linear-to-r from-violet-500/20 via-violet-400/10 to-transparent" />
            <CardHeader className="-mt-10 pb-4">
              <div className="flex items-end justify-between">
                <div className="h-20 w-20 flex items-center justify-center rounded-full ring-4 ring-background shadow-md bg-muted">
                  <UserCircle className="w-12 h-12 text-muted-foreground" />
                </div>
                {currentUser?.role && (
                  <Badge variant="secondary" className="capitalize mb-1">
                    {t(`users.roles.${currentUser.role.name.toLowerCase()}`, { defaultValue: currentUser.role.name })}
                  </Badge>
                )}
              </div>
              <div className="mt-3">
                <CardTitle className="text-base">{t('settings.profile')}</CardTitle>
                <CardDescription>{t('settings.profileDesc')}</CardDescription>
              </div>
            </CardHeader>
            <Separator />
            <CardContent className="pt-6 space-y-4">
              <div className="grid grid-cols-1 sm:grid-cols-2 gap-4">
                <div className="space-y-1.5">
                  <Label>{t('settings.fullName')}</Label>
                  <Input
                    value={profile.name}
                    onChange={(e) => setProfile((p) => ({ ...p, name: e.target.value }))}
                  />
                </div>
                <div className="space-y-1.5">
                  <Label className="flex items-center gap-1.5">
                    {t('settings.email')}
                    <span className="text-xs font-normal text-muted-foreground">
                      ({t('settings.emailReadOnly', { defaultValue: 'แก้ไขไม่ได้' })})
                    </span>
                  </Label>
                  <Input
                    type="email"
                    value={profile.email}
                    disabled
                    className="disabled:opacity-60 disabled:cursor-not-allowed"
                  />
                </div>
              </div>
              {saveError && (
                <p className="text-sm text-destructive">{saveError}</p>
              )}
            </CardContent>
          </Card>

          {/* Security */}
          <Card id="security">
            <CardHeader className="flex flex-row items-center gap-3 pb-4">
              <div className="p-2 rounded-lg bg-emerald-500/10 shrink-0">
                <KeyRound className="w-4 h-4 text-emerald-500" />
              </div>
              <div>
                <CardTitle className="text-base">{t('settings.security')}</CardTitle>
                <CardDescription>{t('settings.securityDesc')}</CardDescription>
              </div>
            </CardHeader>
            <Separator />
            <CardContent className="pt-6 space-y-4">
              <div className="grid grid-cols-1 sm:grid-cols-2 gap-4 max-w-sm sm:max-w-none">
                <div className="space-y-1.5">
                  <Label>{t('settings.newPassword')}</Label>
                  <Input
                    type="password"
                    placeholder="••••••••"
                    value={password.newPass}
                    onChange={(e) => setPassword((p) => ({ ...p, newPass: e.target.value }))}
                  />
                </div>
                <div className="space-y-1.5">
                  <Label>{t('settings.confirmPassword')}</Label>
                  <Input
                    type="password"
                    placeholder="••••••••"
                    value={password.confirm}
                    onChange={(e) => setPassword((p) => ({ ...p, confirm: e.target.value }))}
                  />
                </div>
              </div>
              {passwordError && (
                <p className="text-sm text-destructive">{passwordError}</p>
              )}
              <Button
                onClick={handlePasswordChange}
                disabled={passwordSaving || (!password.newPass && !password.confirm)}
                variant="outline"
                className="gap-2"
              >
                {passwordSaving
                  ? <Loader2 className="w-4 h-4 animate-spin" />
                  : <KeyRound className="w-4 h-4" />
                }
                {passwordSaved
                  ? t('settings.passwordChanged', { defaultValue: 'เปลี่ยนรหัสผ่านแล้ว' })
                  : t('settings.changePassword', { defaultValue: 'เปลี่ยนรหัสผ่าน' })
                }
              </Button>
            </CardContent>
          </Card>

          {/* Appearance */}
          <Card id="appearance">
            <CardHeader className="flex flex-row items-center gap-3 pb-4">
              <div className="p-2 rounded-lg bg-amber-500/10 shrink-0">
                <Palette className="w-4 h-4 text-amber-500" />
              </div>
              <div>
                <CardTitle className="text-base">{t('settings.appearance')}</CardTitle>
                <CardDescription>{t('settings.appearanceDesc')}</CardDescription>
              </div>
            </CardHeader>
            <Separator />
            <CardContent className="pt-6 space-y-6">
              {/* Theme */}
              <div className="space-y-2">
                <Label>{t('settings.theme')}</Label>
                <div className="grid grid-cols-2 gap-3 max-w-xs">
                  <button
                    type="button"
                    onClick={() => setTheme('light')}
                    className={cn(
                      'flex flex-col items-center gap-2 p-4 rounded-xl border-2 text-sm font-medium transition-all',
                      theme === 'light'
                        ? 'border-primary bg-primary/5 text-primary'
                        : 'border-border hover:border-muted-foreground/40 hover:bg-muted/50'
                    )}
                  >
                    <Sun className="w-5 h-5" />
                    {t('settings.themeLight')}
                  </button>
                  <button
                    type="button"
                    onClick={() => setTheme('dark')}
                    className={cn(
                      'flex flex-col items-center gap-2 p-4 rounded-xl border-2 text-sm font-medium transition-all',
                      theme === 'dark'
                        ? 'border-primary bg-primary/5 text-primary'
                        : 'border-border hover:border-muted-foreground/40 hover:bg-muted/50'
                    )}
                  >
                    <Moon className="w-5 h-5" />
                    {t('settings.themeDark')}
                  </button>
                </div>
              </div>

              <Separator />

              {/* Sidebar color */}
              <div className="space-y-2">
                <Label>{t('settings.sidebarColor')}</Label>
                <div className="flex gap-2.5 flex-wrap">
                  {SIDEBAR_COLORS.map((c) => (
                    <button
                      key={c.value}
                      type="button"
                      title={c.label}
                      onClick={() => setSidebarColor(c.value)}
                      className={cn(
                        'w-9 h-9 rounded-full border-2 transition-all',
                        sidebarColor === c.value
                          ? 'border-primary scale-110 shadow-md ring-2 ring-primary/30 ring-offset-1'
                          : 'border-muted-foreground/20 hover:scale-105 hover:border-muted-foreground/50'
                      )}
                      style={{ backgroundColor: c.bg }}
                    />
                  ))}
                </div>
              </div>

              <Separator />

              {/* Language */}
              <div className="space-y-1.5 max-w-xs">
                <Label>{t('settings.language')}</Label>
                <Select
                  value={i18n.language}
                  onValueChange={(v) => {
                    i18n.changeLanguage(v)
                    localStorage.setItem('lang', v)
                  }}
                >
                  <SelectTrigger>
                    <SelectValue />
                  </SelectTrigger>
                  <SelectContent>
                    <SelectItem value="en">🇺🇸 English</SelectItem>
                    <SelectItem value="th">🇹🇭 ภาษาไทย</SelectItem>
                  </SelectContent>
                </Select>
              </div>
            </CardContent>
          </Card>

          <div className="pb-6" />
        </div>
      </div>
    </div>
  )
}
