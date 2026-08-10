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
  Droplets,
  Receipt,
  Wallet,
  Wrench,
  HandCoins,
  Megaphone,
  FileBarChart2,
  Car,
  Package,
  FolderOpen,
  ChevronLeft,
  ChevronRight,
  ChevronDown,
  X,
  Zap,
  SlidersHorizontal,
  Building2,
  Banknote,
  Bell,
  TrendingUp,
  History,
  LayoutList,
} from 'lucide-react'
import { cn } from '@/lib/utils'
import { ScrollArea } from '@/components/ui/scroll-area'
import { Button } from '@/components/ui/button'
import { useAuth } from '@/features/auth/AuthContext'

interface NavItem {
  label: string
  icon: React.ElementType
  to: string
}

interface NavGroup {
  key: string
  label?: string
  icon?: React.ElementType
  items: NavItem[]
}

interface SidebarProps {
  collapsed: boolean
  onToggle: () => void
  onMobileClose?: () => void
}

export function Sidebar({ collapsed, onToggle, onMobileClose }: SidebarProps) {
  const { t } = useTranslation()
  const location = useLocation()
  const { allowedPaths } = useAuth()

  function isAllowed(to: string): boolean {
    return allowedPaths === null || allowedPaths.has(to)
  }

  const navGroups: NavGroup[] = [
    {
      key: 'main',
      items: [
        { label: t('nav.dashboard'), icon: LayoutDashboard, to: '/' },
      ],
    },
    {
      key: 'rooms',
      label: t('nav.groups.rooms'),
      icon: Building2,
      items: [
        { label: t('nav.rooms'), icon: BedDouble, to: '/rooms' },
        { label: t('nav.tenants'), icon: UserRound, to: '/tenants' },
        { label: t('nav.contracts'), icon: FileText, to: '/contracts' },
      ],
    },
    {
      key: 'finance',
      label: t('nav.groups.finance'),
      icon: Banknote,
      items: [
        { label: t('nav.electricMeters'), icon: Zap, to: '/electric-meters' },
        { label: t('nav.waterMeters'), icon: Droplets, to: '/water-meters' },
        { label: t('nav.bills'), icon: Receipt, to: '/bills' },       
        { label: t('nav.payments'), icon: HandCoins, to: '/payments' },
        { label: t('nav.expenses'), icon: Wallet, to: '/expenses' },
      ],
    },
    {
      key: 'services',
      label: t('nav.groups.services'),
      icon: Bell,
      items: [
        { label: t('nav.maintenance'), icon: Wrench, to: '/maintenance' },
        { label: t('nav.announcements'), icon: Megaphone, to: '/announcements' },
        { label: t('nav.parking'), icon: Car, to: '/parking' },
        { label: t('nav.parcels'), icon: Package, to: '/parcels' },
        { label: t('nav.documents'), icon: FolderOpen, to: '/documents' },
      ],
    },
    {
      key: 'reports',
      label: t('nav.groups.reports'),
      icon: TrendingUp,
      items: [
        { label: t('nav.analytics'), icon: BarChart3, to: '/analytics' },
        { label: t('nav.reports'), icon: FileBarChart2, to: '/reports' },
        { label: t('nav.activityLogs'), icon: History, to: '/activity-logs' },
      ],
    },
    {
      key: 'settings',
      label: t('nav.groups.settings'),
      icon: Settings,
      items: [
        { label: t('nav.users'), icon: Users, to: '/settings/users' },
        { label: t('nav.roles'), icon: ShieldCheck, to: '/settings/roles' },
        { label: t('nav.roomTypes'), icon: LayoutList, to: '/settings/room-types' },
        { label: t('nav.dormitories'), icon: Building2, to: '/settings/dormitories' },
        { label: 'ทั่วไป', icon: SlidersHorizontal, to: '/settings/general' },
      ],
    },
  ]

  const [openGroups, setOpenGroups] = useState<Set<string>>(() => {
    const initial = new Set<string>()
    navGroups.forEach((group) => {
      if (!group.label) {
        initial.add(group.key)
        return
      }
      const hasActiveItem = group.items.some((item) =>
        item.to === '/' ? location.pathname === '/' : location.pathname.startsWith(item.to)
      )
      if (hasActiveItem) initial.add(group.key)
    })
    return initial
  })

  function toggleGroup(key: string) {
    setOpenGroups((prev) => {
      const next = new Set(prev)
      if (next.has(key)) {
        next.delete(key)
      } else {
        next.add(key)
      }
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
            <Building2 className="w-4 h-4 text-sidebar-primary-foreground" />
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
      <ScrollArea className="flex-1 py-3">
        <nav className="px-2 space-y-1">
          {navGroups.map((group) => {
            const visibleItems = group.items.filter((item) => isAllowed(item.to))
            if (visibleItems.length === 0) return null

            const GroupIcon = group.icon
            const isGroupOpen = openGroups.has(group.key)
            const hasGroupLabel = !!group.label

            const isGroupActive = visibleItems.some((item) =>
              item.to === '/' ? location.pathname === '/' : location.pathname.startsWith(item.to)
            )

            return (
              <div key={group.key}>
                {/* Group header button */}
                {hasGroupLabel && GroupIcon && (
                  <button
                    type="button"
                    onClick={() => toggleGroup(group.key)}
                    className={cn(
                      'w-full flex items-center gap-3 px-3 py-2.5 rounded-lg text-sm font-medium transition-colors',
                      isGroupActive && !isGroupOpen
                        ? 'bg-sidebar-primary text-sidebar-primary-foreground'
                        : 'text-sidebar-foreground/70 hover:bg-sidebar-accent hover:text-sidebar-accent-foreground',
                      collapsed && 'md:justify-center md:px-2'
                    )}
                    title={collapsed ? group.label : undefined}
                  >
                    <GroupIcon className="w-5 h-5 shrink-0" />
                    <span className={cn('truncate flex-1 text-left', collapsed && 'md:hidden')}>
                      {group.label}
                    </span>
                    <ChevronDown
                      className={cn(
                        'w-4 h-4 shrink-0 transition-transform duration-200 hidden md:block',
                        !isGroupOpen && '-rotate-90'
                      )}
                    />
                  </button>
                )}

                {/* Group items */}
                {(isGroupOpen || !hasGroupLabel || collapsed) && (
                  <div
                    className={cn(
                      'space-y-0.5',
                      hasGroupLabel && !collapsed && 'ml-3 mt-0.5'
                    )}
                  >
                    {visibleItems.map(({ label, icon: Icon, to }) => {
                      const isActive =
                        to === '/' ? location.pathname === '/' : location.pathname.startsWith(to)

                      return (
                        <NavLink
                          key={to}
                          to={to}
                          className={cn(
                            'flex items-center gap-3 rounded-lg text-sm font-medium transition-colors',
                            hasGroupLabel && !collapsed
                              ? 'px-3 py-2'
                              : 'px-3 py-2.5',
                            isActive
                              ? hasGroupLabel && !collapsed
                                ? 'bg-sidebar-primary/20 text-sidebar-primary-foreground'
                                : 'bg-sidebar-primary text-sidebar-primary-foreground'
                              : 'text-sidebar-foreground/60 hover:bg-sidebar-accent hover:text-sidebar-accent-foreground',
                            collapsed && 'md:justify-center md:px-2'
                          )}
                          title={collapsed ? label : undefined}
                        >
                          <Icon
                            className={cn(
                              'shrink-0',
                              hasGroupLabel && !collapsed ? 'w-4 h-4' : 'w-5 h-5'
                            )}
                          />
                          <span className={cn('truncate', collapsed && 'md:hidden')}>{label}</span>
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
