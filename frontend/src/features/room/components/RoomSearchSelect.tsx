import { api } from '@/shared/api/client'
import { RemoteCombobox } from '@/shared/components/ui/remote-combobox'
import type { ApiRoom, RoomStatus } from '../types'

const SEARCH_LIMIT = 20

type RoomSearchSelectProps = {
  selectedLabel: string
  onSelectRoom: (room: ApiRoom) => void
  // Set both to let an optional room be cleared again after it's been picked.
  clearLabel?: string
  onClear?: () => void
  statusFilter?: RoomStatus
  dormitoryId?: string
  placeholder: string
  searchPlaceholder: string
  noResultsLabel: string
  disabled?: boolean
}

// Searches /rooms/active server-side as the user types, instead of preloading
// a capped list of rooms up front — so it keeps working once the caller
// manages enough dormitories/rooms to exceed that cap, and lets typing a
// dormitory name disambiguate rooms that share a number across dormitories.
// Selection state (selectedLabel) is owned by the caller so there's a single
// source of truth for "is a room picked" instead of syncing an input value
// with a separately tracked id.
export function RoomSearchSelect({
  selectedLabel,
  onSelectRoom,
  clearLabel,
  onClear,
  statusFilter,
  dormitoryId,
  placeholder,
  searchPlaceholder,
  noResultsLabel,
  disabled = false,
}: RoomSearchSelectProps) {
  return (
    <RemoteCombobox<ApiRoom>
      fetchItems={(query) =>
        api
          .get<ApiRoom[]>('/rooms/active', {
            params: {
              q: query.trim() || undefined,
              dormitory_id: dormitoryId || undefined,
              status: statusFilter || undefined,
              limit: SEARCH_LIMIT,
            },
          })
          .then(({ data }) => data)
      }
      resetKey={`${dormitoryId ?? ''}|${statusFilter ?? ''}`}
      getKey={(room) => room.id}
      renderItem={(room) => (
        <>
          <span className="font-medium">{room.room_number}</span>
          {room.dormitory_name && <span className="truncate text-muted-foreground">{room.dormitory_name}</span>}
        </>
      )}
      onSelect={onSelectRoom}
      selectedLabel={selectedLabel}
      placeholder={placeholder}
      searchPlaceholder={searchPlaceholder}
      emptyText={noResultsLabel}
      clearLabel={clearLabel}
      onClear={onClear}
      disabled={disabled}
    />
  )
}
