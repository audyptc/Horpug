import { useEffect, useState } from 'react'
import { useSearchParams } from 'react-router-dom'
import { Building2 } from 'lucide-react'
import { api, extractErrorMessage } from '@/shared/api/client'
import { useLanguage } from '@/shared/i18n/language'

type Status = 'loading' | 'linking' | 'success' | 'error'

// Public page a tenant opens from inside the LINE app (via their personal
// linking link, e.g. https://liff.line.me/{LIFF_ID}?tenant_id=...). It logs
// them into the OA's LIFF app, reads the LINE userId LINE issues for that
// login, and sends it to the backend to attach to their tenant record — no
// login of their own required, since the tenant_id in the URL is the shared
// secret that authorizes the link (see tenant handler.LinkLine).
export default function LineLinkPage() {
  const { t } = useLanguage()
  const [searchParams] = useSearchParams()
  const tenantId = searchParams.get('tenant_id')
  const [status, setStatus] = useState<Status>('loading')
  const [error, setError] = useState<string | null>(null)

  useEffect(() => {
    let cancelled = false

    async function run() {
      if (!tenantId) {
        setStatus('error')
        setError(t('lineLinkMissingTenant'))
        return
      }

      const liffId = import.meta.env.VITE_LIFF_ID as string | undefined
      if (!liffId) {
        setStatus('error')
        setError(t('lineLinkNotConfigured'))
        return
      }

      try {
        const liff = (await import('@line/liff')).default
        await liff.init({ liffId })

        if (!liff.isLoggedIn()) {
          liff.login()
          return
        }

        const idToken = liff.getIDToken()
        if (!idToken) {
          throw new Error('missing id token')
        }

        if (cancelled) return
        setStatus('linking')

        await api.post(`/public/tenants/${tenantId}/line/link`, { id_token: idToken })

        if (!cancelled) setStatus('success')
      } catch (err) {
        if (!cancelled) {
          setStatus('error')
          setError(extractErrorMessage(err, t('lineLinkError')))
        }
      }
    }

    run()

    return () => {
      cancelled = true
    }
    // eslint-disable-next-line react-hooks/exhaustive-deps
  }, [tenantId])

  return (
    <div className="login-page">
      <div className="login-card" style={{ textAlign: 'center' }}>
        <div className="login-brand" style={{ justifyContent: 'center' }}>
          <span className="brand-mark" aria-hidden="true">
            <Building2 size={22} strokeWidth={2.4} />
          </span>
          <div>
            <p className="brand-title">Horpug</p>
          </div>
        </div>

        {(status === 'loading' || status === 'linking') && <p>{t('lineLinkInProgress')}</p>}
        {status === 'success' && <p>{t('lineLinkSuccess')}</p>}
        {status === 'error' && <p className="login-error">{error}</p>}
      </div>
    </div>
  )
}
