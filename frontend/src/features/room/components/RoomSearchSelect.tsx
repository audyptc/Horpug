import { useEffect, useRef, useState } from 'react'
import { api } from '@/shared/api/client'
import { useLanguage } from '@/shared/i18n/language'
import type { ApiRoom, RoomStatus } from '../types'

const SEARCH_LIMIT = 20
const DEBOUNCE_MS = 300

type RoomSearchSelectProps = {
  selectedLabel: string
  onSelectRoom: (room: ApiRoom) => void
  onClearSelection: () => void
  statusFilter?: RoomStatus
  dormitoryId?: string
  placeholder: string
  changeLabel: string
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
  onClearSelection,
  statusFilter,
  dormitoryId,
  placeholder,
  changeLabel,
  noResultsLabel,
  disabled,
}: RoomSearchSelectProps) {
  const { t } = useLanguage()
  const [query, setQuery] = useState('')
  const [open, setOpen] = useState(false)
  const [results, setResults] = useState<ApiRoom[]>([])
  const [loading, setLoading] = useState(false)
  const containerRef = useRef<HTMLDivElement>(null)

  useEffect(() => {
    if (!open || selectedLabel) return

    let cancelled = false

    const timer = setTimeout(() => {
      setLoading(true)
      api
        .get<ApiRoom[]>('/rooms/active', {
          params: {
            q: query.trim() || undefined,
            dormitory_id: dormitoryId || undefined,
            limit: SEARCH_LIMIT,
          },
        })
        .then(({ data }) => {
          if (cancelled) return
          setResults(statusFilter ? data.filter((room) => room.status === statusFilter) : data)
        })
        .catch(() => {
          if (!cancelled) setResults([])
        })
        .finally(() => {
          if (!cancelled) setLoading(false)
        })
    }, DEBOUNCE_MS)

    return () => {
      cancelled = true
      clearTimeout(timer)
    }
  }, [open, query, dormitoryId, statusFilter, selectedLabel])

  useEffect(() => {
    function handleClickOutside(event: MouseEvent) {
      if (containerRef.current && !containerRef.current.contains(event.target as Node)) {
        setOpen(false)
      }
    }
    document.addEventListener('mousedown', handleClickOutside)
    return () => document.removeEventListener('mousedown', handleClickOutside)
  }, [])

  if (selectedLabel) {
    return (
      <div className="flex items-center justify-between gap-2 rounded-md border border-input px-3 py-2 text-sm">
        <span>{selectedLabel}</span>
        {!disabled && (
          <button
            type="button"
            className="shrink-0 text-xs font-medium text-primary underline-offset-2 hover:underline"
            onClick={() => {
              setQuery('')
              setResults([])
              onClearSelection()
            }}
          >
            {changeLabel}
          </button>
        )}
      </div>
    )
  }

  return (
    <div ref={containerRef} className="relative flex flex-col gap-1.5">
      <input
        type="text"
        className="h-10 rounded-md border border-input bg-transparent px-3 text-sm"
        placeholder={placeholder}
        value={query}
        disabled={disabled}
        onFocus={() => setOpen(true)}
        onChange={(event) => {
          setQuery(event.target.value)
          setOpen(true)
        }}
        onKeyDown={(event) => {
          if (event.key === 'Escape') setOpen(false)
        }}
      />
      {open && (
        <div className="max-h-56 overflow-y-auto rounded-md border border-input bg-popover text-popover-foreground shadow-md">
          {loading && <p className="px-3 py-2 text-sm text-muted-foreground">{t('loading')}</p>}
          {!loading && results.length === 0 && (
            <p className="px-3 py-2 text-sm text-muted-foreground">{noResultsLabel}</p>
          )}
          {!loading &&
            results.map((room) => (
              <button
                key={room.id}
                type="button"
                className="flex w-full items-center justify-between gap-2 px-3 py-2 text-left text-sm hover:bg-accent hover:text-accent-foreground"
                onClick={() => {
                  onSelectRoom(room)
                  setOpen(false)
                }}
              >
                <span className="font-medium">{room.room_number}</span>
                {room.dormitory_name && (
                  <span className="text-muted-foreground">{room.dormitory_name}</span>
                )}
              </button>
            ))}
        </div>
      )}
    </div>
  )
}
