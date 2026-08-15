import { useTheme } from '@/shared/hooks/use-theme'
import { useLanguage } from '@/shared/i18n/language'
import { Button } from '@/shared/components/ui/button'
import { Bell, Building2, Languages, Moon, PanelLeftClose, PanelLeftOpen, Search, Sun } from 'lucide-react'
import { MobileNavSheet } from './MobileNavSheet'
import { UserMenu } from './UserMenu'

export function TopBar({
  isSidebarCollapsed,
  onToggleSidebar,
  isMobileMenuOpen,
  onMobileMenuOpenChange,
}: {
  isSidebarCollapsed: boolean
  onToggleSidebar: () => void
  isMobileMenuOpen: boolean
  onMobileMenuOpenChange: (open: boolean) => void
}) {
  const { isDark, toggleTheme } = useTheme()
  const { language, setLanguage, t } = useLanguage()

  return (
    <header className="topbar">
      <div className="brand-area">
        <MobileNavSheet open={isMobileMenuOpen} onOpenChange={onMobileMenuOpenChange} />

        <Button
          size="icon"
          variant="ghost"
          className="desktop-only"
          aria-label={isSidebarCollapsed ? t('expandSidebar') : t('collapseSidebar')}
          onClick={onToggleSidebar}
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
          <input id="search-dashboard" type="search" placeholder={t('searchPlaceholder')} />
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
        <UserMenu />
      </div>
    </header>
  )
}
