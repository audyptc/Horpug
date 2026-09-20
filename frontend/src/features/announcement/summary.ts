import { useEffect, useSyncExternalStore } from 'react'
import { api } from '@/shared/api/client'
import type { ApiAnnouncementSummary } from './types'

// One shared copy of the summary, so the sidebar badge, the dashboard card and
// the list page all agree without each fetching (or prop-drilling) its own.
// A module-level store rather than a provider: it has no dependency on where
// in the tree it is used.
let state: ApiAnnouncementSummary | null = null
const listeners = new Set<() => void>()
let inFlight: Promise<void> | null = null

function emit() {
  listeners.forEach((listener) => listener())
}

// Re-reads the summary from the server. Errors are swallowed on purpose: a role
// without access to announcements gets a 403 here, and that should just mean
// "no badge", not a failure the user has to see.
export function refreshAnnouncementSummary(): Promise<void> {
  if (inFlight) return inFlight
  inFlight = api
    .get<ApiAnnouncementSummary>('/announcements/summary')
    .then(({ data }) => {
      state = data
      emit()
    })
    .catch(() => {})
    .finally(() => {
      inFlight = null
    })
  return inFlight
}

// Clears the copy on sign-out so the next user never briefly sees the last
// user's count.
export function resetAnnouncementSummary() {
  state = null
  emit()
}

// Lets a page reflect one announcement being opened without waiting on a
// round trip; the refresh that follows makes it authoritative.
export function markOneRead() {
  if (state && state.unread_count > 0) {
    state = { ...state, unread_count: state.unread_count - 1 }
    emit()
  }
}

function subscribe(listener: () => void) {
  listeners.add(listener)
  return () => listeners.delete(listener)
}

const REFRESH_INTERVAL_MS = 5 * 60 * 1000

// `enabled` should be false when the user's role has no announcements menu, so
// those users don't make a request that can only be refused.
export function useAnnouncementSummary(enabled = true): ApiAnnouncementSummary | null {
  const summary = useSyncExternalStore(subscribe, () => state)

  useEffect(() => {
    if (!enabled) return

    void refreshAnnouncementSummary()
    const timer = window.setInterval(() => void refreshAnnouncementSummary(), REFRESH_INTERVAL_MS)
    const onFocus = () => void refreshAnnouncementSummary()
    window.addEventListener('focus', onFocus)
    return () => {
      window.clearInterval(timer)
      window.removeEventListener('focus', onFocus)
    }
  }, [enabled])

  return enabled ? summary : null
}
