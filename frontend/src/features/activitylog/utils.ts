export const ACTIVITY_LOG_PAGE_SIZE_OPTIONS = [10, 20, 50, 100] as const

export function activityLogActionVariant(
  action: string
): 'default' | 'secondary' | 'destructive' | 'outline' {
  const normalized = action.trim().toLowerCase()
  if (normalized.includes('delete')) return 'destructive'
  if (normalized.includes('update') || normalized.includes('edit')) return 'secondary'
  if (normalized.includes('create') || normalized.includes('login')) return 'default'
  return 'outline'
}
