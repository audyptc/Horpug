import { useNavigate } from 'react-router-dom'
import { LogOut, UserCircle2 } from 'lucide-react'
import { useAuth } from '@/features/auth/AuthProvider'
import { useLanguage } from '@/shared/i18n/language'
import { Button } from '@/shared/components/ui/button'

export function SidebarFooter({ collapsed = false }: { collapsed?: boolean }) {
  const { t } = useLanguage()
  const { session, logout } = useAuth()
  const navigate = useNavigate()

  const handleSignOut = async () => {
    await logout()
    navigate('/login', { replace: true })
  }

  if (!session) return null

  return (
    <div className={`sidebar-footer ${collapsed ? 'collapsed' : ''}`}>
      <UserCircle2 size={28} className="sidebar-footer-avatar" aria-hidden="true" />
      {!collapsed && (
        <div className="sidebar-footer-info">
          <p className="sidebar-footer-name">{session.user.username}</p>
          <p className="sidebar-footer-email">{session.user.email}</p>
        </div>
      )}
      <Button
        size="icon"
        variant="ghost"
        className="sidebar-footer-signout"
        aria-label={t('signOut')}
        title={t('signOut')}
        onClick={handleSignOut}
      >
        <LogOut size={16} />
      </Button>
    </div>
  )
}
