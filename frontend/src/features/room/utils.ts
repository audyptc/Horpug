import type { RoomStatus } from './types'

export const ROOM_PAGE_SIZE_OPTIONS = [10, 20, 50, 100] as const

export const ROOM_STATUSES: RoomStatus[] = ['available', 'occupied', 'maintenance']

// Mirrors the sort fields the rooms endpoint whitelists; anything else is
// rejected there with a 400.
export type RoomSortKey = 'room_number' | 'dormitory' | 'room_type' | 'floor' | 'status' | 'is_active'

export type RoomSortDirection = 'asc' | 'desc'
export type RoomActiveFilter = 'all' | 'active' | 'inactive'
export type RoomStatusFilter = 'all' | RoomStatus

// Mirrors the columns the endpoint accepts as f[<column>]=value; anything else
// is rejected there with a 400.
export const ROOM_TEXT_FILTER_KEYS = ['room_number', 'dormitory', 'room_type'] as const

export type RoomTextFilterKey = (typeof ROOM_TEXT_FILTER_KEYS)[number]
export type RoomColumnFilters = Partial<Record<RoomTextFilterKey, string>>

export function isTextFilterKey(key: string): key is RoomTextFilterKey {
  return (ROOM_TEXT_FILTER_KEYS as readonly string[]).includes(key)
}
