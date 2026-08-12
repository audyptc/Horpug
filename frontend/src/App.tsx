import { useMemo, useState, type ReactNode } from 'react'
import {
  Bell,
  ChartNoAxesColumn,
  CirclePlay,
  Menu,
  PanelLeftClose,
  PanelLeftOpen,
  Search,
  Upload,
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
  label: string
  to: string
  icon: LucideIcon
  badge: string
}

const menuItems: MenuItem[] = [
  {
    label: 'Dashboard',
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
  return (
    <>
      <p className="sidebar-label">Main Menu</p>
      <nav className="menu-list" aria-label="Admin menu">
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
              <span className="menu-text">{item.label}</span>
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
  const metrics = useMemo(
    () => [
      { title: 'Total Views', value: '1,284,920', detail: '+12.4% this month' },
      { title: 'Subscribers', value: '98,432', detail: '+1,203 new' },
      { title: 'Watch Time', value: '3,920 hrs', detail: '+8.7% this week' },
    ],
    []
  )

  const recentUploads = useMemo(
    () => [
      { title: 'Campus Tour 2026', status: 'Published', reach: '24.8K', updated: '2 hours ago' },
      {
        title: 'Student Interview #12',
        status: 'Processing',
        reach: 'Pending',
        updated: '12 minutes ago',
      },
      {
        title: 'Welcome Freshmen Highlights',
        status: 'Draft',
        reach: 'Not Published',
        updated: 'Yesterday',
      },
    ],
    []
  )

  return (
    <main className="content">
      <section className="welcome">
        <h1>Dashboard</h1>
        <p>Overview of your channel performance and latest content actions.</p>
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
          <CardTitle>Recent Upload Queue</CardTitle>
          <CardDescription>Video processing status from your latest uploads.</CardDescription>
        </CardHeader>
        <CardContent>
          <div className="table-wrap">
            <Table>
              <TableHeader>
                <TableRow>
                  <TableHead>Video</TableHead>
                  <TableHead>Status</TableHead>
                  <TableHead>Reach</TableHead>
                  <TableHead className="text-right">Last Update</TableHead>
                </TableRow>
              </TableHeader>
              <TableBody>
                {recentUploads.map((upload) => (
                  <TableRow key={upload.title}>
                    <TableCell className="font-semibold">{upload.title}</TableCell>
                    <TableCell>
                      <Badge
                        variant={
                          upload.status === 'Published'
                            ? 'default'
                            : upload.status === 'Processing'
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

  return (
    <div className="app-shell">
      <header className="topbar">
        <div className="brand-area">
          <Sheet open={isMobileMenuOpen} onOpenChange={setIsMobileMenuOpen}>
            <SheetTrigger asChild>
              <Button size="icon" variant="ghost" className="mobile-only" aria-label="Open menu">
                <Menu size={18} />
              </Button>
            </SheetTrigger>
            <SheetContent side="left" className="mobile-sheet">
              <SheetHeader>
                <SheetTitle>Menu</SheetTitle>
                <SheetDescription>Navigate to your admin sections</SheetDescription>
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
            aria-label="Collapse sidebar"
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
              <p className="brand-subtitle">Studio Dashboard</p>
            </div>
          </div>
        </div>

        <div className="topbar-actions">
          <label className="search-box" htmlFor="search-dashboard">
            <Search size={16} aria-hidden="true" />
            <input
              id="search-dashboard"
              type="search"
              placeholder="Search videos, analytics..."
            />
          </label>
          <Button size="icon" variant="ghost" aria-label="Notifications">
            <Bell size={18} />
          </Button>
          <Button className="upload-btn">
            <Upload size={16} />
            Upload
          </Button>
          <DropdownMenu>
            <DropdownMenuTrigger asChild>
              <Button variant="outline" className="user-trigger">
                <UserCircle2 size={16} />
                Creator
              </Button>
            </DropdownMenuTrigger>
            <DropdownMenuContent align="end" className="w-44">
              <DropdownMenuLabel>My Account</DropdownMenuLabel>
              <DropdownMenuSeparator />
              <DropdownMenuItem>Profile</DropdownMenuItem>
              <DropdownMenuItem>Channel settings</DropdownMenuItem>
              <DropdownMenuSeparator />
              <DropdownMenuItem>Sign out</DropdownMenuItem>
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
