import { useState } from 'react'
import { ChartNoAxesColumn, ChevronDown, Lock } from 'lucide-react'
import { NavLink } from 'react-router-dom'
import { useLanguage } from '@/shared/i18n/language'
import { menuMeta, useMenus, type ApiMenu } from './menus'

export function SidebarNav({
  collapsed = false,
  onNavigate,
}: {
  collapsed?: boolean
  onNavigate?: () => void
}) {
  const { t } = useLanguage()
  const { menus } = useMenus()

  const metaOrder = Object.keys(menuMeta)
  const byMetaOrder = (a: ApiMenu, b: ApiMenu) => metaOrder.indexOf(a.path) - metaOrder.indexOf(b.path)

  const mainMenus = menus.filter((menu) => !menuMeta[menu.path]?.group).sort(byMetaOrder)
  const accessMenus = menus.filter((menu) => menuMeta[menu.path]?.group === 'access').sort(byMetaOrder)

  const [isAccessOpen, setIsAccessOpen] = useState(true)

  const showAccessMenus = collapsed || isAccessOpen

  function renderMenuLink(menu: ApiMenu) {
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
  }

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
        {mainMenus.map(renderMenuLink)}
      </nav>

      {accessMenus.length > 0 && (
        <div className="menu-group-section">
          <button
            type="button"
            className={`sidebar-label menu-group-label menu-group-toggle ${collapsed ? 'collapsed' : ''}`}
            onClick={() => setIsAccessOpen((open) => !open)}
            aria-expanded={showAccessMenus}
            disabled={collapsed}
            title={t('menuGroupAccess')}
          >
            <span className="menu-group-toggle-label">
              <Lock size={14} aria-hidden="true" />
              <span className="menu-text">{t('menuGroupAccess')}</span>
            </span>
            {!collapsed && (
              <ChevronDown
                size={14}
                className={`menu-group-chevron ${isAccessOpen ? 'open' : ''}`}
                aria-hidden="true"
              />
            )}
          </button>
          {showAccessMenus && (
            <nav className="menu-list menu-group" aria-label={t('menuGroupAccess')}>
              {accessMenus.map(renderMenuLink)}
            </nav>
          )}
        </div>
      )}
    </>
  )
}
