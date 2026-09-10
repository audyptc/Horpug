import type { VehicleType } from './types'

export const PARKING_PAGE_SIZE_OPTIONS = [10, 20, 50, 100] as const

export const VEHICLE_TYPES: VehicleType[] = ['car', 'motorcycle', 'other']

// Mirrors the sort fields the parking endpoint whitelists; anything else is
// rejected there with a 400.
export type ParkingSortKey = 'tenant_name' | 'room_number' | 'vehicle_type' | 'license_plate' | 'parking_spot'
export type ParkingSortDirection = 'asc' | 'desc'
export type VehicleTypeFilter = 'all' | VehicleType

// Mirrors the columns the endpoint accepts as f[<column>]=value; anything
// else is rejected there with a 400.
export const PARKING_TEXT_FILTER_KEYS = ['tenant_name', 'room_number', 'license_plate', 'parking_spot'] as const

export type ParkingTextFilterKey = (typeof PARKING_TEXT_FILTER_KEYS)[number]
export type ParkingColumnFilters = Partial<Record<ParkingTextFilterKey, string>>

export function isParkingTextFilterKey(key: string): key is ParkingTextFilterKey {
  return (PARKING_TEXT_FILTER_KEYS as readonly string[]).includes(key)
}
