import { useState } from 'react'
import { useTranslation } from 'react-i18next'
import { Save, Globe, Bell, Shield, Palette, Sun, Moon } from 'lucide-react'
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
import { mockUsers } from '@/data/mockUsers'
import { Avatar, AvatarFallback, AvatarImage } from '@/components/ui/avatar'

const currentUser = mockUsers[0]

export function Settings() {
  const { t, i18n } = useTranslation()
  const { theme, setTheme, sidebarColor, setSidebarColor } = useTheme()
  const [saved, setSaved] = useState(false)
  const [profile, setProfile] = useState({
    name: currentUser.name,
    email: currentUser.email,
    phone: currentUser.phone ?? '',
    department: currentUser.department,
  })
  const [notifs, setNotifs] = useState({ newUsers: true, security: true, updates: false })

  function handleSave() {
    setSaved(true)
    setTimeout(() => setSaved(false), 2000)
  }

  const notifItems = [
    {
      key: 'newUsers' as const,
      label: t('settings.notif.newUsers'),
      desc: t('settings.notif.newUsersDesc'),
    },
    {
      key: 'security' as const,
      label: t('settings.notif.security'),
      desc: t('settings.notif.securityDesc'),
    },
    {
      key: 'updates' as const,
      label: t('settings.notif.updates'),
      desc: t('settings.notif.updatesDesc'),
    },
  ]

  return (
    <div className="space-y-6 max-w-3xl">
      <div>
        <h1 className="text-2xl font-bold tracking-tight">{t('settings.title')}</h1>
        <p className="text-muted-foreground text-sm mt-1">{t('settings.subtitle')}</p>
      </div>

      {/* Profile */}
      <Card>
        <CardHeader className="flex flex-row items-center gap-3 pb-4">
          <div className="p-2 rounded-lg bg-violet-500/10">
            <Globe className="w-4 h-4 text-violet-500" />
          </div>
          <div>
            <CardTitle className="text-base">{t('settings.profile')}</CardTitle>
            <CardDescription>{t('settings.profileDesc')}</CardDescription>
          </div>
        </CardHeader>
        <CardContent className="space-y-6">
          <div className="flex items-center gap-4">
            <Avatar className="h-16 w-16">
              <AvatarImage src={currentUser.avatar} />
              <AvatarFallback>
                {currentUser.name.split(' ').map((n) => n[0]).join('')}
              </AvatarFallback>
            </Avatar>
            <div>
              <p className="font-medium">{currentUser.name}</p>
              <Badge variant="secondary" className="capitalize mt-1">
                {t(`users.roles.${currentUser.role}`)}
              </Badge>
            </div>
          </div>
          <Separator />
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
            <div className="space-y-1.5">
              <Label>{t('settings.phone')}</Label>
              <Input
                value={profile.phone}
                onChange={(e) => setProfile((p) => ({ ...p, phone: e.target.value }))}
              />
            </div>
            <div className="space-y-1.5">
              <Label>{t('settings.department')}</Label>
              <Input
                value={profile.department}
                onChange={(e) => setProfile((p) => ({ ...p, department: e.target.value }))}
              />
            </div>
          </div>
        </CardContent>
      </Card>

      {/* Notifications */}
      <Card>
        <CardHeader className="flex flex-row items-center gap-3 pb-4">
          <div className="p-2 rounded-lg bg-blue-500/10">
            <Bell className="w-4 h-4 text-blue-500" />
          </div>
          <div>
            <CardTitle className="text-base">{t('settings.notifications')}</CardTitle>
            <CardDescription>{t('settings.notificationsDesc')}</CardDescription>
          </div>
        </CardHeader>
        <CardContent className="space-y-4">
          {notifItems.map((item) => (
            <div key={item.key} className="flex items-center justify-between py-1">
              <div>
                <p className="text-sm font-medium">{item.label}</p>
                <p className="text-xs text-muted-foreground">{item.desc}</p>
              </div>
              <button
                type="button"
                onClick={() => setNotifs((n) => ({ ...n, [item.key]: !n[item.key] }))}
                className={`relative inline-flex h-5 w-9 items-center rounded-full transition-colors focus:outline-none ${
                  notifs[item.key] ? 'bg-primary' : 'bg-muted-foreground/30'
                }`}
              >
                <span
                  className={`inline-block h-3.5 w-3.5 rounded-full bg-white shadow transition-transform ${
                    notifs[item.key] ? 'translate-x-5' : 'translate-x-0.5'
                  }`}
                />
              </button>
            </div>
          ))}
        </CardContent>
      </Card>

      {/* Security */}
      <Card>
        <CardHeader className="flex flex-row items-center gap-3 pb-4">
          <div className="p-2 rounded-lg bg-emerald-500/10">
            <Shield className="w-4 h-4 text-emerald-500" />
          </div>
          <div>
            <CardTitle className="text-base">{t('settings.security')}</CardTitle>
            <CardDescription>{t('settings.securityDesc')}</CardDescription>
          </div>
        </CardHeader>
        <CardContent className="space-y-4">
          <div className="space-y-1.5">
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
      <Card>
        <CardHeader className="flex flex-row items-center gap-3 pb-4">
          <div className="p-2 rounded-lg bg-amber-500/10">
            <Palette className="w-4 h-4 text-amber-500" />
          </div>
          <div>
            <CardTitle className="text-base">{t('settings.appearance')}</CardTitle>
            <CardDescription>{t('settings.appearanceDesc')}</CardDescription>
          </div>
        </CardHeader>
        <CardContent className="space-y-4">
          <div className="space-y-1.5">
            <Label>{t('settings.theme')}</Label>
            <div className="flex gap-2">
              <button
                type="button"
                onClick={() => setTheme('light')}
                className={`flex items-center gap-2 px-4 py-2 rounded-lg border text-sm font-medium transition-colors ${
                  theme === 'light'
                    ? 'border-primary bg-primary/10 text-primary'
                    : 'border-border hover:bg-muted'
                }`}
              >
                <Sun className="w-4 h-4" />
                {t('settings.themeLight')}
              </button>
              <button
                type="button"
                onClick={() => setTheme('dark')}
                className={`flex items-center gap-2 px-4 py-2 rounded-lg border text-sm font-medium transition-colors ${
                  theme === 'dark'
                    ? 'border-primary bg-primary/10 text-primary'
                    : 'border-border hover:bg-muted'
                }`}
              >
                <Moon className="w-4 h-4" />
                {t('settings.themeDark')}
              </button>
            </div>
          </div>
          <div className="space-y-1.5">
            <Label>{t('settings.sidebarColor')}</Label>
            <div className="flex gap-2 flex-wrap">
              {SIDEBAR_COLORS.map((c) => (
                <button
                  key={c.value}
                  type="button"
                  title={c.label}
                  onClick={() => setSidebarColor(c.value)}
                  className={`w-8 h-8 rounded-full border-2 transition-all ${
                    sidebarColor === c.value
                      ? 'border-primary scale-110 shadow-md'
                      : 'border-muted-foreground/20 hover:border-muted-foreground/50'
                  }`}
                  style={{ backgroundColor: c.bg }}
                />
              ))}
            </div>
          </div>

          <div className="space-y-1.5">
            <Label>{t('settings.language')}</Label>
            <Select
              value={i18n.language}
              onValueChange={(v) => {
                i18n.changeLanguage(v)
                localStorage.setItem('lang', v)
              }}
            >
              <SelectTrigger className="w-48">
                <SelectValue />
              </SelectTrigger>
              <SelectContent>
                <SelectItem value="en">🇬🇧 English</SelectItem>
                <SelectItem value="th">🇹🇭 ภาษาไทย</SelectItem>
              </SelectContent>
            </Select>
          </div>
          <div className="space-y-1.5">
            <Label>{t('settings.timezone')}</Label>
            <Select defaultValue="asia_bangkok">
              <SelectTrigger className="w-48">
                <SelectValue />
              </SelectTrigger>
              <SelectContent>
                <SelectItem value="asia_bangkok">Asia/Bangkok (GMT+7)</SelectItem>
                <SelectItem value="utc">UTC</SelectItem>
                <SelectItem value="us_eastern">US/Eastern</SelectItem>
              </SelectContent>
            </Select>
          </div>
        </CardContent>
      </Card>

      {/* Save */}
      <div className="flex justify-end">
        <Button onClick={handleSave} className="gap-2 min-w-32">
          <Save className="w-4 h-4" />
          {saved ? t('common.saved') : t('common.save')}
        </Button>
      </div>
    </div>
  )
}
