import { useState, type FormEvent } from 'react'
import { Navigate, useLocation, useNavigate, type Location } from 'react-router-dom'
import { ArrowRight, Building2, CheckCircle2, ShieldCheck, Sparkles } from 'lucide-react'
import { Button } from '@/shared/components/ui/button'
import { useAuth } from '@/features/auth/AuthProvider'
import { useLanguage } from '@/shared/i18n/language'

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
      <div className="login-shell">
        <section className="login-hero" aria-label={t('brandSubtitle')}>
          <p className="login-hero-kicker">
            <Sparkles size={14} aria-hidden="true" />
            Horpug Admin
          </p>
          <h2>จัดการทุกงานของหอพักในที่เดียว</h2>
          <p>
            เข้าถึงข้อมูลห้อง ผู้เช่า ใบแจ้งหนี้ และงานซ่อมได้จากแดชบอร์ดเดียว พร้อมโครงสร้างที่อ่านง่ายและ
            ดูเป็นมืออาชีพมากขึ้น
          </p>

          <div className="login-highlights">
            <div className="login-highlight">
              <span className="login-highlight-icon" aria-hidden="true">
                <ShieldCheck size={16} />
              </span>
              <div>
                <strong>ปลอดภัยและพร้อมใช้งาน</strong>
                <span>ล็อกอินครั้งเดียวแล้วจัดการข้อมูลต่อได้ทันที</span>
              </div>
            </div>
            <div className="login-highlight">
              <span className="login-highlight-icon" aria-hidden="true">
                <CheckCircle2 size={16} />
              </span>
              <div>
                <strong>ภาพรวมชัดเจน</strong>
                <span>สรุปห้อง ผู้เช่า และงานค้างในหน้าเดียว</span>
              </div>
            </div>
          </div>

          <div className="login-hero-footer" aria-hidden="true">
            <span className="login-hero-pill">Rooms</span>
            <span className="login-hero-pill">Billing</span>
            <span className="login-hero-pill">Repairs</span>
            <span className="login-hero-pill">Reports</span>
          </div>
        </section>

        <form className="login-card login-card--form" onSubmit={handleSubmit}>
          <div className="login-brand">
            <span className="brand-mark" aria-hidden="true">
              <Building2 size={26} strokeWidth={2.4} />
            </span>
            <div>
              <p className="brand-title">Horpug</p>
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
            <span>{isSubmitting ? t('signingIn') : t('signIn')}</span>
            <ArrowRight size={16} aria-hidden="true" />
          </Button>
        </form>
      </div>
    </div>
  )
}
