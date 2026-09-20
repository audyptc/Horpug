import type { TranslationKey } from '@/shared/i18n/language'
import type { AnnouncementCategory } from './types'

export const ANNOUNCEMENT_PAGE_SIZE_OPTIONS = [10, 20, 50, 100] as const

// Mirrors the sort fields the announcements endpoint whitelists; anything
// else is rejected there with a 400.
export type AnnouncementSortKey = 'dormitory_name' | 'title' | 'category' | 'is_published' | 'published_date'
export type AnnouncementSortDirection = 'asc' | 'desc'
export type AnnouncementStatusFilter = 'all' | 'published' | 'draft'
export type AnnouncementCategoryFilter = 'all' | AnnouncementCategory

export const ANNOUNCEMENT_CATEGORIES: AnnouncementCategory[] = [
  'general',
  'urgent',
  'billing',
  'maintenance',
  'event',
]

export const announcementCategoryLabelKeys: Record<AnnouncementCategory, TranslationKey> = {
  general: 'announcementCategoryGeneral',
  urgent: 'announcementCategoryUrgent',
  billing: 'announcementCategoryBilling',
  maintenance: 'announcementCategoryMaintenance',
  event: 'announcementCategoryEvent',
}

// Urgent is the one that must stand out; the rest are told apart by label.
export function announcementCategoryVariant(category: AnnouncementCategory) {
  switch (category) {
    case 'urgent':
      return 'destructive' as const
    case 'billing':
      return 'warning' as const
    case 'maintenance':
      return 'secondary' as const
    case 'event':
      return 'success' as const
    default:
      return 'outline' as const
  }
}

// Mirrors the columns the endpoint accepts as f[<column>]=value; anything
// else is rejected there with a 400.
export const ANNOUNCEMENT_TEXT_FILTER_KEYS = ['dormitory_name', 'title'] as const

export type AnnouncementTextFilterKey = (typeof ANNOUNCEMENT_TEXT_FILTER_KEYS)[number]
export type AnnouncementColumnFilters = Partial<Record<AnnouncementTextFilterKey, string>>

export function isAnnouncementTextFilterKey(key: string): key is AnnouncementTextFilterKey {
  return (ANNOUNCEMENT_TEXT_FILTER_KEYS as readonly string[]).includes(key)
}

export function toDateInputValue(value?: string): string {
  return value ? value.slice(0, 10) : ''
}

export function toApiDate(value: string): string {
  return `${value}T00:00:00Z`
}
