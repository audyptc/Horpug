import { afterEach, describe, expect, it, vi } from 'vitest'
import { readTenantIdFromLineRedirect } from './liffParams'

function visit(href: string) {
  vi.stubGlobal('window', { location: { href } })
}

describe('readTenantIdFromLineRedirect', () => {
  afterEach(() => {
    vi.unstubAllGlobals()
  })

  it('reads the tenant id from the query', () => {
    visit('https://horpug.example.com/liff/link-tenant?tenant_id=t-1')
    expect(readTenantIdFromLineRedirect()).toBe('t-1')
  })

  it('reads it from liff.state after the LINE login round trip', () => {
    visit(`https://horpug.example.com/liff/link-tenant?liff.state=${encodeURIComponent('?tenant_id=t-2')}`)
    expect(readTenantIdFromLineRedirect()).toBe('t-2')
  })

  it('reads it from the hash', () => {
    visit('https://horpug.example.com/liff/link-tenant#tenant_id=t-3')
    expect(readTenantIdFromLineRedirect()).toBe('t-3')
  })

  it('returns null for a rich-menu visit and survives malformed state', () => {
    visit('https://horpug.example.com/liff/link-tenant')
    expect(readTenantIdFromLineRedirect()).toBeNull()
    visit('https://horpug.example.com/liff/link-tenant?liff.state=%E0%A4%A')
    expect(readTenantIdFromLineRedirect()).toBeNull()
  })
})
