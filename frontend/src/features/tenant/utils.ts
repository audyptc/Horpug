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
