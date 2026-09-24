export const PENDING_TENANT_ID_KEY = 'liff_pending_tenant_id'

export function readTenantIdFromLineRedirect(): string | null {
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

// isTenantLinkVisit reports whether this LIFF visit is a tenant's personal
// linking link (it carries a tenant_id, or one stashed before a LINE login
// round trip). Any other visit to the LIFF app, e.g. from the OA's rich menu,
// opens the tenant self-service pages.
export function isTenantLinkVisit(): boolean {
  return readTenantIdFromLineRedirect() !== null || sessionStorage.getItem(PENDING_TENANT_ID_KEY) !== null
}
