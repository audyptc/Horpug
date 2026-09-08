export const USER_PAGE_SIZE_OPTIONS = [10, 20, 50, 100] as const

// Mirrors the sort fields the users endpoint whitelists; anything else is
// rejected there with a 400.
export type UserSortKey = 'username' | 'email' | 'role' | 'is_active'

export type UserSortDirection = 'asc' | 'desc'
export type UserStatusFilter = 'all' | 'active' | 'inactive'

// Mirrors the columns the endpoint accepts as f[<column>]=value; anything else
// is rejected there with a 400.
export const USER_TEXT_FILTER_KEYS = ['username', 'email', 'role'] as const

export type UserTextFilterKey = (typeof USER_TEXT_FILTER_KEYS)[number]
export type UserColumnFilters = Partial<Record<UserTextFilterKey, string>>

export function isTextFilterKey(key: string): key is UserTextFilterKey {
  return (USER_TEXT_FILTER_KEYS as readonly string[]).includes(key)
}
