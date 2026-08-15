import { useState, type ReactNode } from 'react'
import { SidebarNav } from '@/features/menu/SidebarNav'
import { TopBar } from './components/TopBar'

export function AdminLayout({ children }: { children: ReactNode }) {
  const [isSidebarCollapsed, setIsSidebarCollapsed] = useState(false)
  const [isMobileMenuOpen, setIsMobileMenuOpen] = useState(false)

  return (
    <div className="app-shell">
      <TopBar
        isSidebarCollapsed={isSidebarCollapsed}
        onToggleSidebar={() => setIsSidebarCollapsed((value) => !value)}
        isMobileMenuOpen={isMobileMenuOpen}
        onMobileMenuOpenChange={setIsMobileMenuOpen}
      />

      <div className={`admin-grid ${isSidebarCollapsed ? 'sidebar-collapsed' : ''}`}>
        <aside className="sidebar desktop-sidebar">
          <SidebarNav collapsed={isSidebarCollapsed} />
        </aside>
        {children}
      </div>
    </div>
  )
}

