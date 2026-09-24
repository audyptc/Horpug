import { useState } from 'react'
import LineLinkPage from './LineLinkPage'
import TenantPortalPage from '@/features/tenantportal/TenantPortalPage'
import { isTenantLinkVisit } from './liffParams'

// The LIFF app has a single endpoint URL (set in the LINE console), so both
// LIFF flows live behind it: a tenant's personal linking link carries a
// tenant_id and runs the linking flow; any other visit (e.g. the OA's rich
// menu, https://liff.line.me/{LIFF_ID}) opens the tenant self-service pages.
export default function LiffEntryPage() {
  // Decided once: the linking flow rewrites the URL as LINE redirects.
  const [linking] = useState(isTenantLinkVisit)
  return linking ? <LineLinkPage /> : <TenantPortalPage />
}
