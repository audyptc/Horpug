import { ChartNoAxesColumn } from 'lucide-react'
import { NavLink } from 'react-router-dom'
import { useLanguage } from '@/shared/i18n/language'
import { menuMeta, useMenus } from './menus'

export function SidebarNav({
  collapsed = false,
  onNavigate,
}: {
  collapsed?: boolean
  onNavigate?: () => void
}) {
  const { t } = useLanguage()
  const { menus } = useMenus()

  return (
    <>
      <p className="sidebar-label">{t('mainMenu')}</p>
      <nav className="menu-list" aria-label={t('adminMenuLabel')}>
        <NavLink
          to="/dashboard"
          className={({ isActive }) =>
            `menu-item ${isActive ? 'active' : ''} ${collapsed ? 'collapsed' : ''}`
          }
          onClick={onNavigate}
        >
          <ChartNoAxesColumn size={16} />
          <span className="menu-text">{t('dashboard')}</span>
        </NavLink>
        {menus.map((menu) => {
          const meta = menuMeta[menu.path]
          if (!meta) return null
          const Icon = meta.icon
          return (
            <NavLink
              key={menu.id}
              to={menu.path}
              className={({ isActive }) =>
                `menu-item ${isActive ? 'active' : ''} ${collapsed ? 'collapsed' : ''}`
              }
              onClick={onNavigate}
            >
              <Icon size={16} />
              <span className="menu-text">{t(meta.labelKey)}</span>
            </NavLink>
          )
        })}
      </nav>
    </>
  )
}
