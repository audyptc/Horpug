export const ROOM_TYPE_PAGE_SIZE_OPTIONS = [10, 20, 50, 100] as const

// Mirrors the sort fields the room-types endpoint whitelists; anything else
// is rejected there with a 400.
export type RoomTypeSortKey = 'name' | 'dormitory' | 'price' | 'is_active'

export type RoomTypeSortDirection = 'asc' | 'desc'
export type RoomTypeStatusFilter = 'all' | 'active' | 'inactive'

// Mirrors the columns the endpoint accepts as f[<column>]=value; anything else
// is rejected there with a 400.
export const ROOM_TYPE_TEXT_FILTER_KEYS = ['name', 'dormitory'] as const

export type RoomTypeTextFilterKey = (typeof ROOM_TYPE_TEXT_FILTER_KEYS)[number]
export type RoomTypeColumnFilters = Partial<Record<RoomTypeTextFilterKey, string>>

export function isTextFilterKey(key: string): key is RoomTypeTextFilterKey {
  return (ROOM_TYPE_TEXT_FILTER_KEYS as readonly string[]).includes(key)
}
