import { useEffect, useRef, useState, type KeyboardEvent, type ReactNode } from 'react'
import { ChevronsUpDown, Search } from 'lucide-react'

import { useLanguage } from '@/shared/i18n/language'
import { cn } from '@/shared/lib/utils'

const DEBOUNCE_MS = 300

type RemoteComboboxProps<T> = {
  // Loads the items matching what has been typed. Called as soon as the panel
  // opens (empty query) and, debounced, on every keystroke after that.
  fetchItems: (query: string) => Promise<T[]>
  // Changing this refetches while the panel is open — pass whatever the search
  // depends on besides the query (a dormitory filter, say). fetchItems itself
  // isn't tracked, so callers needn't memoise it.
  resetKey?: string
  getKey: (item: T) => string
  renderItem: (item: T) => ReactNode
  onSelect: (item: T) => void
  // Owned by the caller: the picked item usually isn't in the current results
  // (or the page is showing a saved record), so the label can't be derived
  // from them.
  selectedLabel: string
  placeholder: string
  searchPlaceholder: string
  emptyText: string
  // When set, a row at the top lets the caller clear an optional selection.
  clearLabel?: string
  onClear?: () => void
  disabled?: boolean
  className?: string
}

// Combobox that searches server-side as the user types, rather than filtering
// a preloaded list. A preloaded list is capped by the API's page size, so past
// that cap the missing entries simply can't be picked.
export function RemoteCombobox<T>({
  fetchItems,
  resetKey,
  getKey,
  renderItem,
  onSelect,
  selectedLabel,
  placeholder,
  searchPlaceholder,
  emptyText,
  clearLabel,
  onClear,
  disabled = false,
  className,
}: RemoteComboboxProps<T>) {
  const { t } = useLanguage()
  const [open, setOpen] = useState(false)
  const [query, setQuery] = useState('')
  const [results, setResults] = useState<T[]>([])
  const [loading, setLoading] = useState(false)
  const [highlightedIndex, setHighlightedIndex] = useState(0)
  const containerRef = useRef<HTMLDivElement>(null)
  const inputRef = useRef<HTMLInputElement>(null)
  const fetchItemsRef = useRef(fetchItems)
  // Tracks whether the panel just opened, so that first fetch (empty query)
  // skips the debounce instead of forcing a wait before anything appears —
  // typed input still debounces normally.
  const isInitialFetchRef = useRef(true)

  const hasClearRow = Boolean(onClear && clearLabel)
  // The clear row sits at index 0 when present, shifting the results down.
  const rowOffset = hasClearRow ? 1 : 0
  const rowCount = results.length + rowOffset

  useEffect(() => {
    fetchItemsRef.current = fetchItems
  })

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
      fetchItemsRef
        .current(query)
        .then((items) => {
          if (!cancelled) setResults(items)
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
  }, [open, query, resetKey])

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
    if (open) requestAnimationFrame(() => inputRef.current?.focus())
  }, [open])

  function togglePanel() {
    if (!open) {
      setQuery('')
      setHighlightedIndex(0)
    }
    setOpen(!open)
  }

  function changeQuery(value: string) {
    setQuery(value)
    setHighlightedIndex(0)
  }

  function selectItem(item: T) {
    onSelect(item)
    setOpen(false)
  }

  function clearSelection() {
    onClear?.()
    setOpen(false)
  }

  function selectHighlighted() {
    if (hasClearRow && highlightedIndex === 0) {
      clearSelection()
      return
    }
    const item = results[highlightedIndex - rowOffset]
    if (item) selectItem(item)
  }

  function handleKeyDown(event: KeyboardEvent<HTMLInputElement>) {
    if (event.key === 'ArrowDown') {
      event.preventDefault()
      setHighlightedIndex((index) => Math.min(index + 1, rowCount - 1))
    } else if (event.key === 'ArrowUp') {
      event.preventDefault()
      setHighlightedIndex((index) => Math.max(index - 1, 0))
    } else if (event.key === 'Enter') {
      event.preventDefault()
      selectHighlighted()
    } else if (event.key === 'Escape') {
      event.preventDefault()
      setOpen(false)
    }
  }

  const rowClass = (index: number) =>
    cn(
      'flex w-full cursor-default items-center justify-between gap-2 rounded-sm px-2 py-1.5 text-left text-sm outline-none transition-colors',
      index === highlightedIndex && 'bg-accent text-accent-foreground'
    )

  return (
    <div ref={containerRef} className={cn('relative', className)}>
      <button
        type="button"
        className={cn(
          'flex h-10 w-full items-center justify-between rounded-md border border-input bg-transparent px-3 text-sm disabled:cursor-not-allowed disabled:opacity-50',
          !selectedLabel && 'text-muted-foreground'
        )}
        onClick={togglePanel}
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
              onChange={(event) => changeQuery(event.target.value)}
              onKeyDown={handleKeyDown}
              placeholder={searchPlaceholder}
              className="h-10 w-full bg-transparent text-sm outline-none placeholder:text-muted-foreground"
            />
          </div>

          <div className="max-h-60 overflow-y-auto p-1">
            {hasClearRow && (
              <button
                type="button"
                className={cn(rowClass(0), 'text-muted-foreground')}
                onMouseEnter={() => setHighlightedIndex(0)}
                onClick={clearSelection}
              >
                {clearLabel}
              </button>
            )}
            {results.length === 0 && (
              <p className="px-2 py-4 text-center text-sm text-muted-foreground">
                {loading ? t('loading') : emptyText}
              </p>
            )}
            {results.map((item, index) => (
              <button
                type="button"
                key={getKey(item)}
                className={rowClass(index + rowOffset)}
                onMouseEnter={() => setHighlightedIndex(index + rowOffset)}
                onClick={() => selectItem(item)}
              >
                {renderItem(item)}
              </button>
            ))}
          </div>
        </div>
      )}
    </div>
  )
}
