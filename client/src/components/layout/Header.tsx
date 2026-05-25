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
import { useAuth } from '@/context/AuthContext'

interface HeaderProps {
  onMenuClick?: () => void
}

export function Header({ onMenuClick }: HeaderProps) {
  const { t, i18n } = useTranslation()
  const { theme, toggleTheme } = useTheme()
  const { logout } = useAuth()
  const navigate = useNavigate()

  async function handleSignOut() {
    await logout()
    navigate('/login', { replace: true })
  }

  const toggleLang = (lang: string) => {
    i18n.changeLanguage(lang)
    localStorage.setItem('lang', lang)
  }

  const notifs = [
    { msg: t('header.notif1'), time: '5m ago' },
    { msg: t('header.notif2'), time: '2h ago' },
    { msg: t('header.notif3'), time: '1d ago' },
  ]

  return (
    <header className="h-16 border-b bg-background flex items-center px-4 gap-4 shrink-0">
      {/* Mobile menu button */}
      <Button variant="ghost" size="icon" className="md:hidden" onClick={onMenuClick}>
        <Menu className="w-6 h-6" />
      </Button>

      {/* Search */}
      <div className="flex-1 max-w-md relative hidden sm:block">
        <Search className="absolute left-3 top-1/2 -translate-y-1/2 w-4 h-4 text-muted-foreground" />
        <Input
          placeholder={t('header.searchPlaceholder')}
          className="pl-9 bg-muted/40 border-0 focus-visible:ring-1"
        />
      </div>

      <div className="flex-1 md:hidden" />

      <div className="flex items-center gap-2 ml-auto">
        {/* Dark mode toggle */}
        <Button variant="ghost" size="icon" onClick={toggleTheme} aria-label="Toggle theme">
          {theme === 'dark' ? <Sun className="w-6 h-6" /> : <Moon className="w-6 h-6" />}
        </Button>

        {/* Language switcher */}
        <DropdownMenu>
          <DropdownMenuTrigger asChild>
            <Button variant="ghost" size="icon" className="relative">
              <Globe className="w-6 h-6" />
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
            <Button variant="ghost" size="icon" className="relative">
              <Bell className="w-6 h-6" />
              <Badge className="absolute -top-1 -right-1 h-4 w-4 p-0 flex items-center justify-center text-[10px]">
                3
              </Badge>
            </Button>
          </DropdownMenuTrigger>
          <DropdownMenuContent align="end" className="w-72 sm:w-80">
            <DropdownMenuLabel>{t('header.notifications')}</DropdownMenuLabel>
            <DropdownMenuSeparator />
            {notifs.map((n, i) => (
              <DropdownMenuItem key={i} className="flex flex-col items-start gap-1 py-3">
                <span className="text-sm">{n.msg}</span>
                <span className="text-xs text-muted-foreground">{n.time}</span>
              </DropdownMenuItem>
            ))}
          </DropdownMenuContent>
        </DropdownMenu>

        {/* User menu */}
        <DropdownMenu>
          <DropdownMenuTrigger asChild>
            <Button variant="ghost" size="icon">
              <UserCircle className="w-6 h-6" />
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
