import { NavLink, useLocation } from 'react-router-dom'
import { useTranslation } from 'react-i18next'
import {
  LayoutDashboard,
  Users,
  Settings,
  ShieldCheck,
  BarChart3,
  Bell,
  ChevronLeft,
  ChevronRight,
  Zap,
} from 'lucide-react'
import { cn } from '@/lib/utils'
import { ScrollArea } from '@/components/ui/scroll-area'
import { Button } from '@/components/ui/button'

interface SidebarProps {
  collapsed: boolean
  onToggle: () => void
}

export function Sidebar({ collapsed, onToggle }: SidebarProps) {
  const { t } = useTranslation()
  const location = useLocation()

  const navItems = [
    { label: t('nav.dashboard'), icon: LayoutDashboard, to: '/' },
    { label: t('nav.users'), icon: Users, to: '/users' },
    { label: t('nav.analytics'), icon: BarChart3, to: '/analytics' },
    { label: t('nav.notifications'), icon: Bell, to: '/notifications' },
    { label: t('nav.roles'), icon: ShieldCheck, to: '/roles' },
    { label: t('nav.settings'), icon: Settings, to: '/settings' },
  ]

  return (
    <aside
      className={cn(
        'flex flex-col h-screen bg-sidebar border-r border-sidebar-border transition-all duration-300 shrink-0',
        collapsed ? 'w-60 md:w-16' : 'w-60'
      )}
    >
      {/* Logo */}
      <div className="flex items-center h-16 px-4 border-b border-sidebar-border shrink-0">
        <div className="flex items-center gap-2 min-w-0">
          <div className="flex items-center justify-center w-8 h-8 rounded-lg bg-sidebar-primary shrink-0">
            <Zap className="w-4 h-4 text-sidebar-primary-foreground" />
          </div>
          <span className={cn('font-bold text-sidebar-foreground text-lg tracking-tight truncate', collapsed && 'md:hidden')}>
            Horpug
          </span>
        </div>
      </div>

      {/* Nav */}
      <ScrollArea className="flex-1 py-4">
        <nav className="px-2 space-y-1">
          {navItems.map(({ label, icon: Icon, to }) => {
            const isActive =
              to === '/' ? location.pathname === '/' : location.pathname.startsWith(to)
            return (
              <NavLink
                key={to}
                to={to}
                className={cn(
                  'flex items-center gap-3 px-3 py-2.5 rounded-lg text-sm font-medium transition-colors',
                  isActive
                    ? 'bg-sidebar-primary text-sidebar-primary-foreground'
                    : 'text-sidebar-foreground/70 hover:bg-sidebar-accent hover:text-sidebar-accent-foreground',
                  collapsed && 'md:justify-center md:px-2'
                )}
                title={collapsed ? label : undefined}
              >
                <Icon className="w-4 h-4 shrink-0" />
                <span className={cn('truncate', collapsed && 'md:hidden')}>{label}</span>
              </NavLink>
            )
          })}
        </nav>
      </ScrollArea>

      {/* Collapse toggle — desktop only */}
      <div className="hidden md:block p-2 border-t border-sidebar-border shrink-0">
        <Button
          variant="ghost"
          size="icon"
          onClick={onToggle}
          className={cn(
            'w-full text-sidebar-foreground/70 hover:bg-sidebar-accent hover:text-sidebar-accent-foreground',
            collapsed ? 'px-2' : ''
          )}
        >
          {collapsed ? (
            <ChevronRight className="w-4 h-4" />
          ) : (
            <div className="flex items-center gap-2 w-full">
              <ChevronLeft className="w-4 h-4" />
              <span className="text-sm">{t('nav.collapse')}</span>
            </div>
          )}
        </Button>
      </div>
    </aside>
  )
}
