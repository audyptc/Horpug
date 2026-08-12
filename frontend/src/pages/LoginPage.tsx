import { useState, type FormEvent } from 'react'
import { Navigate, useLocation, useNavigate, type Location } from 'react-router-dom'
import { CirclePlay } from 'lucide-react'
import { Button } from '@/components/ui/button'
import { useAuth } from '@/lib/auth'
import { useLanguage } from '@/lib/language'

export default function LoginPage() {
  const { isAuthenticated, isLoading, login } = useAuth()
  const { t } = useLanguage()
  const navigate = useNavigate()
  const location = useLocation()
  const [username, setUsername] = useState('')
  const [password, setPassword] = useState('')
  const [error, setError] = useState<string | null>(null)
  const [isSubmitting, setIsSubmitting] = useState(false)

  // Wait for the boot-time silent refresh (see auth.tsx) so a visitor who
  // still has a valid refresh cookie doesn't see the form flash before being
  // redirected straight to the dashboard.
  if (isLoading) {
    return <div className="route-loading" aria-hidden="true" />
  }

  if (isAuthenticated) {
    const redirectTo = (location.state as { from?: Location } | null)?.from?.pathname ?? '/dashboard'
    return <Navigate to={redirectTo} replace />
  }

  const handleSubmit = async (event: FormEvent) => {
    event.preventDefault()
    setError(null)
    setIsSubmitting(true)
    try {
      await login(username, password)
      navigate('/dashboard', { replace: true })
    } catch (err) {
      setError(err instanceof Error ? err.message : t('loginError'))
    } finally {
      setIsSubmitting(false)
    }
  }

  return (
    <div className="login-page">
      <form className="login-card" onSubmit={handleSubmit}>
        <div className="login-brand">
          <span className="brand-mark" aria-hidden="true">
            <CirclePlay size={22} strokeWidth={2.4} />
          </span>
          <div>
            <p className="brand-title">YouTube Admin</p>
            <p className="brand-subtitle">{t('brandSubtitle')}</p>
          </div>
        </div>

        <h1>{t('loginTitle')}</h1>
        <p className="login-subtitle">{t('loginSubtitle')}</p>

        {error && <p className="login-error">{error}</p>}

        <label className="login-field">
          <span>{t('username')}</span>
          <input
            value={username}
            onChange={(event) => setUsername(event.target.value)}
            autoComplete="username"
            required
          />
        </label>

        <label className="login-field">
          <span>{t('password')}</span>
          <input
            type="password"
            value={password}
            onChange={(event) => setPassword(event.target.value)}
            autoComplete="current-password"
            required
          />
        </label>

        <Button type="submit" disabled={isSubmitting} className="login-submit">
          {isSubmitting ? t('signingIn') : t('signIn')}
        </Button>
      </form>
    </div>
  )
}
