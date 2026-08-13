import { useState, type ReactNode } from 'react'
import { useTheme } from '@/shared/hooks/use-theme'
import { useLanguage } from '@/shared/i18n/language'
import { useAuth } from '@/features/auth/AuthProvider'
import { SidebarNav } from '@/features/menu/SidebarNav'
import {
  Bell,
  Building2,
  Languages,
  Menu,
  Moon,
  PanelLeftClose,
  PanelLeftOpen,
  Search,
  Sun,
  UserCircle2,
} from 'lucide-react'
import { useNavigate } from 'react-router-dom'
import { Button } from '@/shared/components/ui/button'
import {
  DropdownMenu,
  DropdownMenuContent,
  DropdownMenuItem,
  DropdownMenuLabel,
  DropdownMenuSeparator,
  DropdownMenuTrigger,
} from '@/shared/components/ui/dropdown-menu'
import {
  Sheet,
  SheetContent,
  SheetDescription,
  SheetHeader,
  SheetTitle,
  SheetTrigger,
} from '@/shared/components/ui/sheet'

export function AdminLayout({ children }: { children: ReactNode }) {
  const [isSidebarCollapsed, setIsSidebarCollapsed] = useState(false)
  const [isMobileMenuOpen, setIsMobileMenuOpen] = useState(false)
  const { isDark, toggleTheme } = useTheme()
  const { language, setLanguage, t } = useLanguage()
  const { session, logout } = useAuth()
  const navigate = useNavigate()

  const handleSignOut = async () => {
    await logout()
    navigate('/login', { replace: true })
  }

  return (
    <div className="app-shell">
      <header className="topbar">
        <div className="brand-area">
          <Sheet open={isMobileMenuOpen} onOpenChange={setIsMobileMenuOpen}>
            <SheetTrigger asChild>
              <Button size="icon" variant="ghost" className="mobile-only" aria-label={t('openMenu')}>
                <Menu size={18} />
              </Button>
            </SheetTrigger>
            <SheetContent side="left" className="mobile-sheet">
              <SheetHeader>
                <SheetTitle>{t('mobileMenuTitle')}</SheetTitle>
                <SheetDescription>{t('mobileMenuDescription')}</SheetDescription>
              </SheetHeader>
              <div className="mobile-sheet-nav">
                <SidebarNav onNavigate={() => setIsMobileMenuOpen(false)} />
              </div>
            </SheetContent>
          </Sheet>

          <Button
            size="icon"
            variant="ghost"
            className="desktop-only"
            aria-label={isSidebarCollapsed ? t('expandSidebar') : t('collapseSidebar')}
            onClick={() => setIsSidebarCollapsed((value) => !value)}
          >
            {isSidebarCollapsed ? <PanelLeftOpen size={18} /> : <PanelLeftClose size={18} />}
          </Button>

          <div className="brand">
            <span className="brand-mark" aria-hidden="true">
              <Building2 size={20} strokeWidth={2.4} />
            </span>
            <div>
              <p className="brand-title">Horpug</p>
              <p className="brand-subtitle">{t('brandSubtitle')}</p>
            </div>
          </div>
        </div>

        <div className="topbar-center">
          <label className="search-box" htmlFor="search-dashboard">
            <Search size={16} aria-hidden="true" />
            <input
              id="search-dashboard"
              type="search"
              placeholder={t('searchPlaceholder')}
            />
          </label>
        </div>

        <div className="topbar-actions">

          <Button
            size="icon"
            variant="ghost"
            className="rounded-full"
            aria-label={isDark ? t('switchToLight') : t('switchToDark')}
            onClick={toggleTheme}
          >
            {isDark ? <Sun size={18} /> : <Moon size={18} />}
          </Button>
          <Button
            variant="outline"
            className="lang-btn"
            aria-label={t('language')}
            onClick={() => setLanguage(language === 'en' ? 'th' : 'en')}
          >
            <Languages size={16} />
            {language === 'en' ? 'EN' : 'ไทย'}
          </Button>
          <Button size="icon" variant="ghost" className="rounded-full" aria-label={t('notifications')}>
            <Bell size={18} />
          </Button>
          <DropdownMenu>
            <DropdownMenuTrigger asChild>
              <Button size="icon" variant="ghost" className="user-trigger" aria-label={t('accountMenu')}>
                <UserCircle2 size={22} />
              </Button>
            </DropdownMenuTrigger>
            <DropdownMenuContent align="end" className="w-56">
              <DropdownMenuLabel>
                <p className="account-name">{session?.user.username}</p>
                <p className="account-email">{session?.user.email}</p>
                {session?.user.role?.name && (
                  <p className="account-role">{session.user.role.name}</p>
                )}
              </DropdownMenuLabel>
              <DropdownMenuSeparator />
              <DropdownMenuItem>{t('profile')}</DropdownMenuItem>
              <DropdownMenuItem>{t('channelSettings')}</DropdownMenuItem>
              <DropdownMenuSeparator />
              <DropdownMenuItem onSelect={handleSignOut}>{t('signOut')}</DropdownMenuItem>
            </DropdownMenuContent>
          </DropdownMenu>
        </div>
      </header>

      <div className={`admin-grid ${isSidebarCollapsed ? 'sidebar-collapsed' : ''}`}>
        <aside className="sidebar desktop-sidebar">
          <SidebarNav collapsed={isSidebarCollapsed} />
        </aside>
        {children}
      </div>
    </div>
  )
}
