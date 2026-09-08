export const ACTIVITY_LOG_PAGE_SIZE_OPTIONS = [10, 20, 50, 100] as const

// Mirrors the sort fields the activity-logs endpoint whitelists; anything
// else is rejected there with a 400.
export type ActivityLogSortKey =
  | 'created_at'
  | 'username'
  | 'action'
  | 'entity_type'
  | 'description'
  | 'ip_address'

export type ActivityLogSortDirection = 'asc' | 'desc'

// Mirrors the columns the endpoint accepts as f[<column>]=value; anything else
// is rejected there with a 400.
export const ACTIVITY_LOG_TEXT_FILTER_KEYS = [
  'username',
  'action',
  'entity_type',
  'description',
  'ip_address',
] as const

export type ActivityLogTextFilterKey = (typeof ACTIVITY_LOG_TEXT_FILTER_KEYS)[number]
export type ActivityLogColumnFilters = Partial<Record<ActivityLogTextFilterKey, string>>

export function isTextFilterKey(key: string): key is ActivityLogTextFilterKey {
  return (ACTIVITY_LOG_TEXT_FILTER_KEYS as readonly string[]).includes(key)
}

export function activityLogActionVariant(
  action: string
): 'default' | 'secondary' | 'destructive' | 'outline' {
  const normalized = action.trim().toLowerCase()
  if (normalized.includes('delete')) return 'destructive'
  if (normalized.includes('update') || normalized.includes('edit')) return 'secondary'
  if (normalized.includes('create') || normalized.includes('login')) return 'default'
  return 'outline'
}
