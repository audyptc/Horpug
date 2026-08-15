import { useAuth } from '@/features/auth/AuthProvider'
import { Button } from '@/shared/components/ui/button'
import {
  DropdownMenu,
  DropdownMenuContent,
  DropdownMenuItem,
  DropdownMenuLabel,
  DropdownMenuSeparator,
  DropdownMenuTrigger,
} from '@/shared/components/ui/dropdown-menu'
import { useLanguage } from '@/shared/i18n/language'
import { UserCircle2 } from 'lucide-react'
import { useNavigate } from 'react-router-dom'

export function UserMenu() {
  const { t } = useLanguage()
  const { session, logout } = useAuth()
  const navigate = useNavigate()

  const handleSignOut = async () => {
    await logout()
    navigate('/login', { replace: true })
  }

  return (
    <DropdownMenu>
      <DropdownMenuTrigger asChild>
        <Button size="icon" variant="ghost" className="user-trigger" aria-label={t('accountMenu')}>
          <UserCircle2 size={22} />
        </Button>
      </DropdownMenuTrigger>
      <DropdownMenuContent align="end" className="w-56">
        <DropdownMenuLabel>
          <p className="account-name">{session?.user.username}</p>
          <p className="account-email">{session?.user.email}</p>
          {session?.user.role?.name && <p className="account-role">{session.user.role.name}</p>}
        </DropdownMenuLabel>
        <DropdownMenuSeparator />
        <DropdownMenuItem>{t('profile')}</DropdownMenuItem>
        <DropdownMenuItem>{t('channelSettings')}</DropdownMenuItem>
        <DropdownMenuSeparator />
        <DropdownMenuItem onSelect={handleSignOut}>{t('signOut')}</DropdownMenuItem>
      </DropdownMenuContent>
    </DropdownMenu>
  )
}
