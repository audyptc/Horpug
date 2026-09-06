import { useEffect, useState } from 'react'
import { Building2 } from 'lucide-react'
import { api, extractErrorMessage } from '@/shared/api/client'
import { useLanguage } from '@/shared/i18n/language'

type Status = 'loading' | 'linking' | 'success' | 'error'

function readTenantIdFromLineRedirect(): string | null {
  const url = new URL(window.location.href)
  const directTenantId = url.searchParams.get('tenant_id')
  if (directTenantId) return directTenantId

  const encodedState = url.searchParams.get('liff.state')
  if (encodedState) {
    try {
      const decoded = decodeURIComponent(encodedState)
      const stateParams = new URLSearchParams(decoded)
      const stateTenantId = stateParams.get('tenant_id')
      if (stateTenantId) return stateTenantId
    } catch {
      // Ignore malformed state and fall through to the generic error below.
    }
  }

  const hashState = new URLSearchParams(url.hash.replace(/^#/, ''))
  const hashTenantId = hashState.get('tenant_id')
  if (hashTenantId) return hashTenantId

  return null
}

// Public page a tenant opens from inside the LINE app (via their personal
// linking link, e.g. https://liff.line.me/{LIFF_ID}?tenant_id=...). It logs
// them into the OA's LIFF app, reads the LINE userId LINE issues for that
// login, and sends it to the backend to attach to their tenant record — no
// login of their own required, since the tenant_id in the URL is the shared
// secret that authorizes the link (see tenant handler.LinkLine).
export default function LineLinkPage() {
  const { t } = useLanguage()
  const [status, setStatus] = useState<Status>('loading')
  const [error, setError] = useState<string | null>(null)

  useEffect(() => {
    let cancelled = false

    async function run() {
      const liffId = import.meta.env.VITE_LIFF_ID as string | undefined
      if (!liffId) {
        setStatus('error')
        setError(t('lineLinkNotConfigured'))
        return
      }

      try {
        const liff = (await import('@line/liff')).default
        await liff.init({ liffId })

        // liff.init() restores the real URL — LINE briefly encodes the
        // original query string (our tenant_id) into a `liff.state` param
        // while redirecting through liff.line.me, so tenant_id must be read
        // from the live URL after init, not from what react-router saw on
        // the very first render. Some users also land on a redirect that keeps
        // the state in the hash or encoded query string, so we try both.
        const tenantId = readTenantIdFromLineRedirect()
        if (!tenantId) {
          setStatus('error')
          setError(t('lineLinkMissingTenant'))
          return
        }

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
  }, [])

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
