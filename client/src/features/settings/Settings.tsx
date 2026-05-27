import { useState, useEffect } from 'react'
import { useTranslation } from 'react-i18next'
import { Save, Bell, Palette, Sun, Moon, User2, Camera, KeyRound } from 'lucide-react'
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
import { Avatar, AvatarFallback, AvatarImage } from '@/components/ui/avatar'
import { useAuth } from '@/features/auth/AuthContext'
import { userService } from '@/features/users/userService'
import type { ApiUser } from '@/types/api'
import { cn } from '@/lib/utils'

function getAvatar(name: string) {
  return `https://api.dicebear.com/9.x/avataaars/svg?seed=${encodeURIComponent(name)}`
}

export function Settings() {
  const { t, i18n } = useTranslation()
  const { theme, setTheme, sidebarColor, setSidebarColor } = useTheme()
  const { userId } = useAuth()
  const [saved, setSaved] = useState(false)
  const [activeSection, setActiveSection] = useState('profile')
  const [currentUser, setCurrentUser] = useState<ApiUser | null>(null)
  const [profile, setProfile] = useState({ name: '', email: '' })
  const [notifs, setNotifs] = useState({ newUsers: true, security: true, updates: false })

  useEffect(() => {
    if (!userId) return
    userService.getById(userId).then((user) => {
      setCurrentUser(user)
      setProfile({ name: user.full_name, email: user.email })
    }).catch(() => {})
  }, [userId])

  function handleSave() {
    setSaved(true)
    setTimeout(() => setSaved(false), 2000)
  }

  function scrollTo(id: string) {
    setActiveSection(id)
    document.getElementById(id)?.scrollIntoView({ behavior: 'smooth', block: 'start' })
  }

  const navItems = [
    { id: 'profile', label: t('settings.profile'), icon: User2, color: 'text-violet-500' },
    { id: 'notifications', label: t('settings.notifications'), icon: Bell, color: 'text-blue-500' },
    { id: 'security', label: t('settings.security'), icon: KeyRound, color: 'text-emerald-500' },
    { id: 'appearance', label: t('settings.appearance'), icon: Palette, color: 'text-amber-500' },
  ]

  const notifItems = [
    { key: 'newUsers' as const, label: t('settings.notif.newUsers'), desc: t('settings.notif.newUsersDesc') },
    { key: 'security' as const, label: t('settings.notif.security'), desc: t('settings.notif.securityDesc') },
    { key: 'updates' as const, label: t('settings.notif.updates'), desc: t('settings.notif.updatesDesc') },
  ]

  const avatarName = currentUser?.full_name ?? profile.name ?? '?'

  return (
    <div className="max-w-5xl mx-auto">
      {/* Header */}
      <div className="flex items-start justify-between gap-4 mb-8">
        <div>
          <h1 className="text-2xl font-bold tracking-tight">{t('settings.title')}</h1>
          <p className="text-muted-foreground text-sm mt-1">{t('settings.subtitle')}</p>
        </div>
        <Button onClick={handleSave} className="gap-2 min-w-28 shrink-0">
          <Save className="w-4 h-4" />
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
                <div className="relative">
                  <Avatar className="h-20 w-20 ring-4 ring-background shadow-md">
                    <AvatarImage src={avatarName !== '?' ? getAvatar(avatarName) : undefined} />
                    <AvatarFallback className="text-lg">
                      {avatarName.split(' ').map((n) => n[0]).join('').slice(0, 2)}
                    </AvatarFallback>
                  </Avatar>
                  <button
                    type="button"
                    className="absolute bottom-0 right-0 p-1.5 rounded-full bg-background border shadow-sm hover:bg-muted transition-colors"
                  >
                    <Camera className="w-3 h-3 text-muted-foreground" />
                  </button>
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
            <CardContent className="pt-6">
              <div className="grid grid-cols-1 sm:grid-cols-2 gap-4">
                <div className="space-y-1.5">
                  <Label>{t('settings.fullName')}</Label>
                  <Input
                    value={profile.name}
                    onChange={(e) => setProfile((p) => ({ ...p, name: e.target.value }))}
                  />
                </div>
                <div className="space-y-1.5">
                  <Label>{t('settings.email')}</Label>
                  <Input
                    type="email"
                    value={profile.email}
                    onChange={(e) => setProfile((p) => ({ ...p, email: e.target.value }))}
                  />
                </div>
              </div>
            </CardContent>
          </Card>

          {/* Notifications */}
          <Card id="notifications">
            <CardHeader className="flex flex-row items-center gap-3 pb-4">
              <div className="p-2 rounded-lg bg-blue-500/10 shrink-0">
                <Bell className="w-4 h-4 text-blue-500" />
              </div>
              <div>
                <CardTitle className="text-base">{t('settings.notifications')}</CardTitle>
                <CardDescription>{t('settings.notificationsDesc')}</CardDescription>
              </div>
            </CardHeader>
            <Separator />
            <CardContent className="pt-2">
              {notifItems.map((item, i) => (
                <div key={item.key}>
                  <div className="flex items-center justify-between py-3.5">
                    <div>
                      <p className="text-sm font-medium">{item.label}</p>
                      <p className="text-xs text-muted-foreground mt-0.5">{item.desc}</p>
                    </div>
                    <button
                      type="button"
                      onClick={() => setNotifs((n) => ({ ...n, [item.key]: !n[item.key] }))}
                      className={cn(
                        'relative inline-flex h-5 w-9 items-center rounded-full transition-colors focus:outline-none shrink-0 ml-4',
                        notifs[item.key] ? 'bg-primary' : 'bg-muted-foreground/30'
                      )}
                    >
                      <span
                        className={cn(
                          'inline-block h-3.5 w-3.5 rounded-full bg-white shadow transition-transform',
                          notifs[item.key] ? 'translate-x-5' : 'translate-x-0.5'
                        )}
                      />
                    </button>
                  </div>
                  {i < notifItems.length - 1 && <Separator />}
                </div>
              ))}
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
              <div className="space-y-1.5 max-w-sm">
                <Label>{t('settings.currentPassword')}</Label>
                <Input type="password" placeholder="••••••••" />
              </div>
              <div className="grid grid-cols-1 sm:grid-cols-2 gap-4">
                <div className="space-y-1.5">
                  <Label>{t('settings.newPassword')}</Label>
                  <Input type="password" placeholder="••••••••" />
                </div>
                <div className="space-y-1.5">
                  <Label>{t('settings.confirmPassword')}</Label>
                  <Input type="password" placeholder="••••••••" />
                </div>
              </div>
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

              {/* Language & Timezone */}
              <div className="grid grid-cols-1 sm:grid-cols-2 gap-4">
                <div className="space-y-1.5">
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
                <div className="space-y-1.5">
                  <Label>{t('settings.timezone')}</Label>
                  <Select defaultValue="asia_bangkok">
                    <SelectTrigger>
                      <SelectValue />
                    </SelectTrigger>
                    <SelectContent>
                      <SelectItem value="asia_bangkok">Asia/Bangkok (GMT+7)</SelectItem>
                      <SelectItem value="utc">UTC</SelectItem>
                      <SelectItem value="us_eastern">US/Eastern</SelectItem>
                    </SelectContent>
                  </Select>
                </div>
              </div>
            </CardContent>
          </Card>

          <div className="pb-6" />
        </div>
      </div>
    </div>
  )
}
