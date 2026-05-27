import { NavLink, useLocation } from 'react-router-dom'
import { useState } from 'react'
import { useTranslation } from 'react-i18next'
import {
  LayoutDashboard,
  Users,
  Settings,
  ShieldCheck,
  BarChart3,
  BedDouble,
  UserRound,
  FileText,
  Gauge,
  Receipt,
  Wallet,
  Wrench,
  HandCoins,
  Megaphone,
  FileBarChart2,
  Car,
  ChevronLeft,
  ChevronRight,
  ChevronDown,
  X,
  Zap,
  SlidersHorizontal,
} from 'lucide-react'
import { cn } from '@/lib/utils'
import { ScrollArea } from '@/components/ui/scroll-area'
import { Button } from '@/components/ui/button'

interface SubNavItem {
  label: string
  to: string
  icon: React.ElementType
}

interface NavItem {
  label: string
  icon: React.ElementType
  to: string
  children?: SubNavItem[]
}

interface SidebarProps {
  collapsed: boolean
  onToggle: () => void
  onMobileClose?: () => void
}

export function Sidebar({ collapsed, onToggle, onMobileClose }: SidebarProps) {
  const { t } = useTranslation()
  const location = useLocation()

  const navItems: NavItem[] = [
    { label: t('nav.dashboard'), icon: LayoutDashboard, to: '/' },
    { label: t('nav.rooms'), icon: BedDouble, to: '/rooms' },
    { label: t('nav.tenants'), icon: UserRound, to: '/tenants' },
    { label: t('nav.contracts'), icon: FileText, to: '/contracts' },
    { label: t('nav.meters'), icon: Gauge, to: '/meters' },
    { label: t('nav.bills'), icon: Receipt, to: '/bills' },
    { label: t('nav.expenses'), icon: Wallet, to: '/expenses' },
    { label: t('nav.maintenance'), icon: Wrench, to: '/maintenance' },
    { label: t('nav.payments'), icon: HandCoins, to: '/payments' },
    { label: t('nav.announcements'), icon: Megaphone, to: '/announcements' },
    { label: t('nav.parking'), icon: Car, to: '/parking' },
    { label: t('nav.analytics'), icon: BarChart3, to: '/analytics' },
    { label: t('nav.reports'), icon: FileBarChart2, to: '/reports' },
    {
      label: t('nav.settings'),
      icon: Settings,
      to: '/settings',
      children: [
        { label: t('nav.users'), icon: Users, to: '/settings/users' },
        { label: t('nav.roles'), icon: ShieldCheck, to: '/settings/roles' },
        { label: 'ทั่วไป', icon: SlidersHorizontal, to: '/settings/general' },
      ],
    },
  ]

  const [openGroups, setOpenGroups] = useState<Set<string>>(() => {
    const initial = new Set<string>()
    navItems.forEach((item) => {
      if (item.children && location.pathname.startsWith(item.to)) {
        initial.add(item.to)
      }
    })
    return initial
  })

  function toggleGroup(to: string) {
    setOpenGroups((prev) => {
      const next = new Set(prev)
      next.has(to) ? next.delete(to) : next.add(to)
      return next
    })
  }

  return (
    <aside
      className={cn(
        'flex flex-col h-screen bg-sidebar border-r border-sidebar-border transition-all duration-300 shrink-0',
        collapsed ? 'w-60 md:w-16' : 'w-60'
      )}
    >
      {/* Logo */}
      <div className="flex items-center h-16 px-4 border-b border-sidebar-border shrink-0">
        <div className="flex items-center gap-2 min-w-0 flex-1">
          <div className="flex items-center justify-center w-8 h-8 rounded-lg bg-sidebar-primary shrink-0">
            <Zap className="w-4 h-4 text-sidebar-primary-foreground" />
          </div>
          <span className={cn('font-bold text-sidebar-foreground text-lg tracking-tight truncate', collapsed && 'md:hidden')}>
            Horpug
          </span>
        </div>
        <Button
          variant="ghost"
          size="icon"
          onClick={onMobileClose}
          className="md:hidden text-sidebar-foreground/70 hover:bg-sidebar-accent hover:text-sidebar-accent-foreground shrink-0"
        >
          <X className="w-4 h-4" />
        </Button>
      </div>

      {/* Nav */}
      <ScrollArea className="flex-1 py-4">
        <nav className="px-2 space-y-1">
          {navItems.map(({ label, icon: Icon, to, children }) => {
            const isActive =
              to === '/' ? location.pathname === '/' : location.pathname.startsWith(to)
            const hasChildren = children && children.length > 0

            const isOpen = hasChildren ? openGroups.has(to) : false

            return (
              <div key={to}>
                {hasChildren ? (
                  <button
                    type="button"
                    onClick={() => toggleGroup(to)}
                    className={cn(
                      'w-full flex items-center gap-3 px-3 py-2.5 rounded-lg text-sm font-medium transition-colors',
                      isActive
                        ? 'bg-sidebar-primary text-sidebar-primary-foreground'
                        : 'text-sidebar-foreground/70 hover:bg-sidebar-accent hover:text-sidebar-accent-foreground',
                      collapsed && 'md:justify-center md:px-2'
                    )}
                    title={collapsed ? label : undefined}
                  >
                    <Icon className="w-5 h-5 shrink-0" />
                    <span className={cn('truncate flex-1 text-left', collapsed && 'md:hidden')}>{label}</span>
                    <ChevronDown
                      className={cn(
                        'w-4 h-4 shrink-0 transition-transform md:block hidden',
                        isOpen && 'rotate-180'
                      )}
                    />
                  </button>
                ) : (
                  <NavLink
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
                    <Icon className="w-5 h-5 shrink-0" />
                    <span className={cn('truncate', collapsed && 'md:hidden')}>{label}</span>
                  </NavLink>
                )}

                {/* Submenu */}
                {hasChildren && isOpen && (
                  <div className={cn('mt-1 space-y-0.5', collapsed ? 'md:hidden' : 'ml-3')}>
                    {children.map((child) => {
                      const isChildActive = location.pathname === child.to || location.pathname.startsWith(child.to)
                      const ChildIcon = child.icon
                      return (
                        <NavLink
                          key={child.to}
                          to={child.to}
                          className={cn(
                            'flex items-center gap-3 px-3 py-2 rounded-lg text-sm font-medium transition-colors',
                            isChildActive
                              ? 'bg-sidebar-primary/20 text-sidebar-primary-foreground'
                              : 'text-sidebar-foreground/60 hover:bg-sidebar-accent hover:text-sidebar-accent-foreground'
                          )}
                        >
                          <ChildIcon className="w-4 h-4 shrink-0" />
                          <span className="truncate">{child.label}</span>
                        </NavLink>
                      )
                    })}
                  </div>
                )}
              </div>
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
