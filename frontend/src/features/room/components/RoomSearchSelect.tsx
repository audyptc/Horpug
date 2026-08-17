import { useEffect, useRef, useState } from 'react'
import { ChevronsUpDown, Search } from 'lucide-react'
import { api } from '@/shared/api/client'
import { useLanguage } from '@/shared/i18n/language'
import { cn } from '@/shared/lib/utils'
import type { ApiRoom, RoomStatus } from '../types'

const SEARCH_LIMIT = 20
const DEBOUNCE_MS = 300

type RoomSearchSelectProps = {
  selectedLabel: string
  onSelectRoom: (room: ApiRoom) => void
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
  statusFilter,
  dormitoryId,
  placeholder,
  searchPlaceholder,
  noResultsLabel,
  disabled = false,
}: RoomSearchSelectProps) {
  const { t } = useLanguage()
  const [open, setOpen] = useState(false)
  const [query, setQuery] = useState('')
  const [results, setResults] = useState<ApiRoom[]>([])
  const [loading, setLoading] = useState(false)
  const [highlightedIndex, setHighlightedIndex] = useState(0)
  const containerRef = useRef<HTMLDivElement>(null)
  const inputRef = useRef<HTMLInputElement>(null)
  // Tracks whether the panel just opened, so that first fetch (still showing
  // the placeholder query) skips the debounce instead of forcing a wait
  // before anything appears — typed input still debounces normally.
  const isInitialFetchRef = useRef(true)

  useEffect(() => {
    if (!open) {
      isInitialFetchRef.current = true
      return
    }

    let cancelled = false
    const delay = isInitialFetchRef.current ? 0 : DEBOUNCE_MS
    isInitialFetchRef.current = false

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
    }, delay)

    return () => {
      cancelled = true
      clearTimeout(timer)
    }
  }, [open, query, dormitoryId, statusFilter])

  useEffect(() => {
    if (!open) return

    function handleClickOutside(event: MouseEvent) {
      if (containerRef.current && !containerRef.current.contains(event.target as Node)) {
        setOpen(false)
      }
    }

    document.addEventListener('mousedown', handleClickOutside)
    return () => document.removeEventListener('mousedown', handleClickOutside)
  }, [open])

  useEffect(() => {
    if (open) {
      setQuery('')
      setHighlightedIndex(0)
      requestAnimationFrame(() => inputRef.current?.focus())
    }
  }, [open])

  useEffect(() => {
    setHighlightedIndex(0)
  }, [query])

  function selectRoom(room: ApiRoom) {
    onSelectRoom(room)
    setOpen(false)
  }

  function handleKeyDown(event: React.KeyboardEvent<HTMLInputElement>) {
    if (event.key === 'ArrowDown') {
      event.preventDefault()
      setHighlightedIndex((index) => Math.min(index + 1, results.length - 1))
    } else if (event.key === 'ArrowUp') {
      event.preventDefault()
      setHighlightedIndex((index) => Math.max(index - 1, 0))
    } else if (event.key === 'Enter') {
      event.preventDefault()
      const room = results[highlightedIndex]
      if (room) selectRoom(room)
    } else if (event.key === 'Escape') {
      event.preventDefault()
      setOpen(false)
    }
  }

  return (
    <div ref={containerRef} className="relative">
      <button
        type="button"
        className={cn(
          'flex h-10 w-full items-center justify-between rounded-md border border-input bg-transparent px-3 text-sm disabled:cursor-not-allowed disabled:opacity-50',
          !selectedLabel && 'text-muted-foreground'
        )}
        onClick={() => setOpen((value) => !value)}
        disabled={disabled}
      >
        <span className="truncate">{selectedLabel || placeholder}</span>
        <ChevronsUpDown className="size-4 shrink-0 opacity-50" />
      </button>

      {open && (
        <div className="absolute z-50 mt-1 w-full overflow-hidden rounded-md border border-border bg-popover text-popover-foreground shadow-md">
          <div className="flex items-center gap-2 border-b border-border px-3">
            <Search className="size-4 shrink-0 opacity-50" />
            <input
              ref={inputRef}
              type="text"
              value={query}
              onChange={(event) => setQuery(event.target.value)}
              onKeyDown={handleKeyDown}
              placeholder={searchPlaceholder}
              className="h-10 w-full bg-transparent text-sm outline-none placeholder:text-muted-foreground"
            />
          </div>

          <div className="max-h-60 overflow-y-auto p-1">
            {results.length === 0 && (
              <p className="px-2 py-4 text-center text-sm text-muted-foreground">
                {loading ? t('loading') : noResultsLabel}
              </p>
            )}
            {results.map((room, index) => (
              <button
                type="button"
                key={room.id}
                className={cn(
                  'flex w-full cursor-default items-center justify-between gap-2 rounded-sm px-2 py-1.5 text-left text-sm outline-none transition-colors',
                  index === highlightedIndex && 'bg-accent text-accent-foreground'
                )}
                onMouseEnter={() => setHighlightedIndex(index)}
                onClick={() => selectRoom(room)}
              >
                <span className="font-medium">{room.room_number}</span>
                {room.dormitory_name && (
                  <span className="truncate text-muted-foreground">{room.dormitory_name}</span>
                )}
              </button>
            ))}
          </div>
        </div>
      )}
    </div>
  )
}
