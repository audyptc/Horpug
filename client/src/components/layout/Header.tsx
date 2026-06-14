import { useState, useEffect } from 'react'
import { Bell, Search, Menu, Globe, Sun, Moon, UserCircle } from 'lucide-react'
import { useTranslation } from 'react-i18next'
import { useNavigate } from 'react-router-dom'
import { Button } from '@/components/ui/button'
import { Input } from '@/components/ui/input'
import { Badge } from '@/components/ui/badge'
import {
  DropdownMenu,
  DropdownMenuContent,
  DropdownMenuItem,
  DropdownMenuLabel,
  DropdownMenuSeparator,
  DropdownMenuTrigger,
} from '@/components/ui/dropdown-menu'
import { useTheme } from '@/lib/theme'
import { useAuth } from '@/features/auth/AuthContext'
import { useNotifications } from '@/hooks/useNotifications'
import { SearchDialog } from '@/components/shared/search/SearchDialog'

interface HeaderProps {
  onMenuClick?: () => void
}

export function Header({ onMenuClick }: HeaderProps) {
  const { t, i18n } = useTranslation()
  const { theme, toggleTheme } = useTheme()
  const { logout } = useAuth()
  const navigate = useNavigate()
  const notifications = useNotifications()
  const [searchOpen, setSearchOpen] = useState(false)

  useEffect(() => {
    function handleKeyDown(e: KeyboardEvent) {
      if (e.key === 'k' && (e.ctrlKey || e.metaKey)) {
        e.preventDefault()
        setSearchOpen(true)
      }
    }
    document.addEventListener('keydown', handleKeyDown)
    return () => document.removeEventListener('keydown', handleKeyDown)
  }, [])

  async function handleSignOut() {
    await logout()
    navigate('/login', { replace: true })
  }

  const toggleLang = (lang: string) => {
    i18n.changeLanguage(lang)
    localStorage.setItem('lang', lang)
  }

  return (
    <header className="h-16 border-b bg-background flex items-center px-4 gap-4 shrink-0">
      {/* Mobile menu button */}
      <Button variant="ghost" size="icon" className="md:hidden" onClick={onMenuClick}>
        <Menu className="w-10 h-10" />
      </Button>

      {/* Search */}
      <div className="flex-1 max-w-md relative hidden sm:block">
        <Search className="absolute left-3 top-1/2 -translate-y-1/2 w-4 h-4 text-muted-foreground" />
        <Input
          placeholder={t('header.searchPlaceholder')}
          className="pl-9 bg-muted/40 border-0 focus-visible:ring-1 cursor-pointer"
          readOnly
          onClick={() => setSearchOpen(true)}
        />
        <kbd className="absolute right-3 top-1/2 -translate-y-1/2 hidden lg:inline-flex h-5 items-center gap-0.5 rounded border bg-muted px-1.5 text-[10px] text-muted-foreground pointer-events-none">
          Ctrl K
        </kbd>
      </div>

      <SearchDialog open={searchOpen} onOpenChange={setSearchOpen} />

      <div className="flex-1 md:hidden" />

      <div className="flex items-center gap-2 ml-auto">
        {/* Dark mode toggle */}
        <Button variant="ghost" size="icon" onClick={toggleTheme} aria-label="Toggle theme" className="h-11 w-11 [&_svg]:size-6">
          {theme === 'dark' ? <Sun /> : <Moon />}
        </Button>

        {/* Language switcher */}
        <DropdownMenu>
          <DropdownMenuTrigger asChild>
            <Button variant="ghost" size="icon" className="relative h-11 w-11 [&_svg]:size-6">
              <Globe />
              <span className="absolute -bottom-0.5 -right-0.5 text-[9px] font-bold bg-primary text-primary-foreground rounded px-0.5 leading-tight">
                {i18n.language === 'th' ? 'TH' : 'EN'}
              </span>
            </Button>
          </DropdownMenuTrigger>
          <DropdownMenuContent align="end">
            <DropdownMenuLabel>
              <Globe className="w-3.5 h-3.5 inline mr-1.5" />
              Language
            </DropdownMenuLabel>
            <DropdownMenuSeparator />
            <DropdownMenuItem
              onClick={() => toggleLang('en')}
              className={i18n.language === 'en' ? 'bg-accent' : ''}
            >
              <span className="mr-2 text-base">🇬🇧</span> English
            </DropdownMenuItem>
            <DropdownMenuItem
              onClick={() => toggleLang('th')}
              className={i18n.language === 'th' ? 'bg-accent' : ''}
            >
              <span className="mr-2 text-base">🇹🇭</span> ภาษาไทย
            </DropdownMenuItem>
          </DropdownMenuContent>
        </DropdownMenu>

        {/* Notifications */}
        <DropdownMenu>
          <DropdownMenuTrigger asChild>
            <Button variant="ghost" size="icon" className="relative h-11 w-11 [&_svg]:size-6">
              <Bell />
              {notifications.length > 0 && (
                <Badge className="absolute -top-1 -right-1 h-4 w-4 p-0 flex items-center justify-center text-[10px]">
                  {notifications.length}
                </Badge>
              )}
            </Button>
          </DropdownMenuTrigger>
          <DropdownMenuContent align="end" className="w-72 sm:w-80">
            <DropdownMenuLabel>{t('header.notifications')}</DropdownMenuLabel>
            <DropdownMenuSeparator />
            {notifications.length === 0 ? (
              <DropdownMenuItem className="py-3 text-sm text-muted-foreground justify-center">
                {t('header.notifNoAlerts')}
              </DropdownMenuItem>
            ) : (
              notifications.map((n) => {
                let msg: string
                let sub: string
                if (n.type === 'overdue_bill') {
                  msg = t('header.notifOverdue', {
                    room: n.room_number,
                    name: n.tenant_name,
                    days: n.days_overdue,
                    amount: (n.total_amount ?? 0).toLocaleString(),
                  })
                  sub = `${n.days_overdue}d overdue`
                } else if (n.type === 'expiring_contract') {
                  msg = t('header.notifExpiring', {
                    room: n.room_number,
                    name: n.tenant_name,
                    days: n.days_remaining,
                  })
                  sub = `${n.days_remaining}d remaining`
                } else {
                  msg = t('header.notifMaintenance', {
                    room: n.room_number,
                    title: n.title,
                  })
                  sub = n.priority ?? ''
                }
                return (
                  <DropdownMenuItem key={n.id} className="flex flex-col items-start gap-1 py-3">
                    <span className="text-sm">{msg}</span>
                    <span className="text-xs text-muted-foreground">{sub}</span>
                  </DropdownMenuItem>
                )
              })
            )}
          </DropdownMenuContent>
        </DropdownMenu>

        {/* User menu */}
        <DropdownMenu>
          <DropdownMenuTrigger asChild>
            <Button variant="ghost" size="icon" className="h-11 w-11 [&_svg]:size-6">
              <UserCircle />
            </Button>
          </DropdownMenuTrigger>
          <DropdownMenuContent align="end" className="w-48">
            <DropdownMenuLabel>{t('header.myAccount')}</DropdownMenuLabel>
            <DropdownMenuSeparator />
            <DropdownMenuItem>{t('header.profile')}</DropdownMenuItem>
            <DropdownMenuItem>{t('header.settings')}</DropdownMenuItem>
            <DropdownMenuSeparator />
            <DropdownMenuItem className="text-destructive" onClick={handleSignOut}>
              {t('header.signOut')}
            </DropdownMenuItem>
          </DropdownMenuContent>
        </DropdownMenu>
      </div>
    </header>
  )
}
