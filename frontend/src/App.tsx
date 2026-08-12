import { useMemo, useState, type ReactNode } from 'react'
import { useTheme } from '@/lib/use-theme'
import { useLanguage, type TranslationKey } from '@/lib/language'
import {
  Bell,
  ChartNoAxesColumn,
  CirclePlay,
  Languages,
  Menu,
  Moon,
  PanelLeftClose,
  PanelLeftOpen,
  Search,
  Sun,
  UserCircle2,
  type LucideIcon,
} from 'lucide-react'
import { NavLink, Navigate, Route, Routes } from 'react-router-dom'
import { Badge } from '@/components/ui/badge'
import { Button } from '@/components/ui/button'
import {
  Card,
  CardContent,
  CardDescription,
  CardHeader,
  CardTitle,
} from '@/components/ui/card'
import {
  DropdownMenu,
  DropdownMenuContent,
  DropdownMenuItem,
  DropdownMenuLabel,
  DropdownMenuSeparator,
  DropdownMenuTrigger,
} from '@/components/ui/dropdown-menu'
import {
  Sheet,
  SheetContent,
  SheetDescription,
  SheetHeader,
  SheetTitle,
  SheetTrigger,
} from '@/components/ui/sheet'
import {
  Table,
  TableBody,
  TableCell,
  TableHead,
  TableHeader,
  TableRow,
} from '@/components/ui/table'

type MenuItem = {
  labelKey: TranslationKey
  to: string
  icon: LucideIcon
  badge: string
}

const menuItems: MenuItem[] = [
  {
    labelKey: 'dashboard',
    to: '/dashboard',
    icon: ChartNoAxesColumn,
    badge: '1',
  },
]

function SidebarNav({
  collapsed = false,
  onNavigate,
}: {
  collapsed?: boolean
  onNavigate?: () => void
}) {
  const { t } = useLanguage()

  return (
    <>
      <p className="sidebar-label">{t('mainMenu')}</p>
      <nav className="menu-list" aria-label={t('adminMenuLabel')}>
        {menuItems.map((item) => {
          const Icon = item.icon
          return (
            <NavLink
              key={item.to}
              to={item.to}
              className={({ isActive }) =>
                `menu-item ${isActive ? 'active' : ''} ${collapsed ? 'collapsed' : ''}`
              }
              onClick={onNavigate}
            >
              <Icon size={16} />
              <span className="menu-text">{t(item.labelKey)}</span>
              <Badge variant="secondary" className="menu-badge">
                {item.badge}
              </Badge>
            </NavLink>
          )
        })}
      </nav>
    </>
  )
}

function DashboardPage() {
  const { t, language } = useLanguage()

  const metrics = useMemo(
    () => [
      { title: t('totalViews'), value: '1,284,920', detail: t('totalViewsDetail') },
      { title: t('subscribers'), value: '98,432', detail: t('subscribersDetail') },
      { title: t('watchTime'), value: '3,920 hrs', detail: t('watchTimeDetail') },
    ],
    // eslint-disable-next-line react-hooks/exhaustive-deps
    [language]
  )

  const recentUploads = useMemo(
    () => [
      {
        title: 'Campus Tour 2026',
        status: t('statusPublished'),
        reach: '24.8K',
        updated: t('timeHoursAgo'),
      },
      {
        title: 'Student Interview #12',
        status: t('statusProcessing'),
        reach: t('reachPending'),
        updated: t('timeMinutesAgo'),
      },
      {
        title: 'Welcome Freshmen Highlights',
        status: t('statusDraft'),
        reach: t('reachNotPublished'),
        updated: t('timeYesterday'),
      },
    ],
    // eslint-disable-next-line react-hooks/exhaustive-deps
    [language]
  )

  return (
    <main className="content">
      <section className="welcome">
        <h1>{t('dashboard')}</h1>
        <p>{t('dashboardOverview')}</p>
      </section>

      <section className="metric-grid">
        {metrics.map((metric) => (
          <Card key={metric.title}>
            <CardHeader>
              <CardDescription>{metric.title}</CardDescription>
              <CardTitle>{metric.value}</CardTitle>
            </CardHeader>
            <CardContent>
              <p className="metric-detail">{metric.detail}</p>
            </CardContent>
          </Card>
        ))}
      </section>

      <Card>
        <CardHeader>
          <CardTitle>{t('recentUploadQueue')}</CardTitle>
          <CardDescription>{t('recentUploadQueueDescription')}</CardDescription>
        </CardHeader>
        <CardContent>
          <div className="table-wrap">
            <Table>
              <TableHeader>
                <TableRow>
                  <TableHead>{t('tableVideo')}</TableHead>
                  <TableHead>{t('tableStatus')}</TableHead>
                  <TableHead>{t('tableReach')}</TableHead>
                  <TableHead className="text-right">{t('tableLastUpdate')}</TableHead>
                </TableRow>
              </TableHeader>
              <TableBody>
                {recentUploads.map((upload) => (
                  <TableRow key={upload.title}>
                    <TableCell className="font-semibold">{upload.title}</TableCell>
                    <TableCell>
                      <Badge
                        variant={
                          upload.status === t('statusPublished')
                            ? 'default'
                            : upload.status === t('statusProcessing')
                              ? 'secondary'
                              : 'outline'
                        }
                      >
                        {upload.status}
                      </Badge>
                    </TableCell>
                    <TableCell>{upload.reach}</TableCell>
                    <TableCell className="text-right text-muted-foreground">{upload.updated}</TableCell>
                  </TableRow>
                ))}
              </TableBody>
            </Table>
          </div>
        </CardContent>
      </Card>
    </main>
  )
}

function AdminLayout({ children }: { children: ReactNode }) {
  const [isSidebarCollapsed, setIsSidebarCollapsed] = useState(false)
  const [isMobileMenuOpen, setIsMobileMenuOpen] = useState(false)
  const { isDark, toggleTheme } = useTheme()
  const { language, setLanguage, t } = useLanguage()

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
              <CirclePlay size={20} strokeWidth={2.4} />
            </span>
            <div>
              <p className="brand-title">YouTube Admin</p>
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
          <Button size="icon" variant="ghost" className="rounded-full" aria-label={t('notifications')}>
            <Bell size={18} />
          </Button>
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
          <DropdownMenu>
            <DropdownMenuTrigger asChild>
              <Button size="icon" variant="ghost" className="user-trigger" aria-label={t('accountMenu')}>
                <UserCircle2 size={22} />
              </Button>
            </DropdownMenuTrigger>
            <DropdownMenuContent align="end" className="w-44">
              <DropdownMenuLabel>{t('myAccount')}</DropdownMenuLabel>
              <DropdownMenuSeparator />
              <DropdownMenuItem>{t('profile')}</DropdownMenuItem>
              <DropdownMenuItem>{t('channelSettings')}</DropdownMenuItem>
              <DropdownMenuSeparator />
              <DropdownMenuItem>{t('signOut')}</DropdownMenuItem>
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

function App() {
  return (
    <Routes>
      <Route path="/" element={<Navigate to="/dashboard" replace />} />
      <Route
        path="/dashboard"
        element={
          <AdminLayout>
            <DashboardPage />
          </AdminLayout>
        }
      />
      <Route path="*" element={<Navigate to="/dashboard" replace />} />
    </Routes>
  )
}

export default App
