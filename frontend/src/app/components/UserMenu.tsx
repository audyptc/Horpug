import { useState } from 'react'
import { useAuth } from '@/features/auth/AuthProvider'
import { ChangePasswordSheet } from '@/features/auth/ChangePasswordSheet'
import { InformationDialog } from '@/shared/components/information-dialog'
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
  const [changePasswordOpen, setChangePasswordOpen] = useState(false)
  const [changedNoticeOpen, setChangedNoticeOpen] = useState(false)
  const [protectedNoticeOpen, setProtectedNoticeOpen] = useState(false)
  // The seeded admin's password is re-applied from ADMIN_PASSWORD on every
  // startup, so it is changed there, not here.
  const isProtected = session?.user.is_protected === true

  const handleSignOut = async () => {
    await logout()
    navigate('/login', { replace: true })
  }

  return (
    <>
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
          <DropdownMenuItem onSelect={() => (isProtected ? setProtectedNoticeOpen(true) : setChangePasswordOpen(true))}>
            {t('changePasswordMenu')}
          </DropdownMenuItem>
          <DropdownMenuItem onSelect={handleSignOut}>{t('signOut')}</DropdownMenuItem>
        </DropdownMenuContent>
      </DropdownMenu>
      <ChangePasswordSheet
        open={changePasswordOpen}
        onOpenChange={setChangePasswordOpen}
        onChanged={() => setChangedNoticeOpen(true)}
      />
      <InformationDialog
        open={protectedNoticeOpen}
        onOpenChange={setProtectedNoticeOpen}
        title={t('changePasswordProtectedTitle')}
        description={t('changePasswordProtectedDescription')}
        actionLabel={t('changePasswordDoneAction')}
      />
      <InformationDialog
        open={changedNoticeOpen}
        onOpenChange={setChangedNoticeOpen}
        title={t('changePasswordDoneTitle')}
        description={t('changePasswordDoneDescription')}
        actionLabel={t('changePasswordDoneAction')}
      />
    </>
  )
}
