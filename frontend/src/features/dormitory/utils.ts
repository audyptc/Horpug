export const DORMITORY_PAGE_SIZE_OPTIONS = [10, 20, 50, 100] as const

// Mirrors the sort fields the dormitories endpoint whitelists; anything else
// is rejected there with a 400.
export type DormitorySortKey = 'name' | 'address' | 'phone' | 'is_active'

export type DormitorySortDirection = 'asc' | 'desc'
export type DormitoryStatusFilter = 'all' | 'active' | 'inactive'

// Mirrors the columns the endpoint accepts as f[<column>]=value; anything else
// is rejected there with a 400.
export const DORMITORY_TEXT_FILTER_KEYS = ['name', 'address', 'phone'] as const

export type DormitoryTextFilterKey = (typeof DORMITORY_TEXT_FILTER_KEYS)[number]
export type DormitoryColumnFilters = Partial<Record<DormitoryTextFilterKey, string>>

export function isTextFilterKey(key: string): key is DormitoryTextFilterKey {
  return (DORMITORY_TEXT_FILTER_KEYS as readonly string[]).includes(key)
}
