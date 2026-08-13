import type { RoomStatus } from './types'

export const ROOM_PAGE_SIZE_OPTIONS = [10, 20, 50, 100] as const

export const ROOM_STATUSES: RoomStatus[] = ['available', 'occupied', 'maintenance']
