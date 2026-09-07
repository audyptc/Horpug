export const TENANT_PAGE_SIZE_OPTIONS = [10, 20, 50, 100] as const

// Mirrors the sort fields the tenants endpoint whitelists; anything else is
// rejected there with a 400.
export type TenantSortKey =
  | 'first_name'
  | 'last_name'
  | 'phone'
  | 'line_id'
  | 'id_card'
  | 'email'
  | 'is_active'

export type TenantSortDirection = 'asc' | 'desc'
export type TenantStatusFilter = 'all' | 'active' | 'inactive'
export type TenantLineFilter = 'all' | 'linked' | 'unlinked'

// Mirrors the columns the endpoint accepts as f[<column>]=value; anything else
// is rejected there with a 400.
export const TENANT_TEXT_FILTER_KEYS = [
  'first_name',
  'last_name',
  'phone',
  'line_id',
  'id_card',
  'email',
] as const

export type TenantTextFilterKey = (typeof TENANT_TEXT_FILTER_KEYS)[number]
export type TenantColumnFilters = Partial<Record<TenantTextFilterKey, string>>

export function isTextFilterKey(key: string): key is TenantTextFilterKey {
  return (TENANT_TEXT_FILTER_KEYS as readonly string[]).includes(key)
}
