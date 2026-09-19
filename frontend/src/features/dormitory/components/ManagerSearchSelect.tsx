import { api } from '@/shared/api/client'
import { RemoteCombobox } from '@/shared/components/ui/remote-combobox'
import type { ApiUser } from '../types'

const SEARCH_LIMIT = 20

type ManagerSearchSelectProps = {
  // Users already chosen as managers, left out of the results.
  excludeIds: string[]
  onSelectUser: (user: ApiUser) => void
  placeholder: string
  searchPlaceholder: string
  noResultsLabel: string
}

// Searches /users/active server-side as the user types (username or email), so
// users past the API's preload cap can still be added as managers. It never
// holds a selection itself: picking a user hands it to the caller, which lists
// it, so the trigger always shows the placeholder.
export function ManagerSearchSelect({
  excludeIds,
  onSelectUser,
  placeholder,
  searchPlaceholder,
  noResultsLabel,
}: ManagerSearchSelectProps) {
  return (
    <RemoteCombobox<ApiUser>
      fetchItems={(query) =>
        api
          .get<ApiUser[]>('/users/active', { params: { q: query.trim() || undefined, limit: SEARCH_LIMIT } })
          .then(({ data }) => data.filter((user) => !excludeIds.includes(user.id)))
      }
      resetKey={excludeIds.join(',')}
      getKey={(user) => user.id}
      renderItem={(user) => (
        <>
          <span className="truncate font-medium">{user.username}</span>
          <span className="truncate text-muted-foreground">{user.email}</span>
        </>
      )}
      onSelect={onSelectUser}
      selectedLabel=""
      placeholder={placeholder}
      searchPlaceholder={searchPlaceholder}
      emptyText={noResultsLabel}
    />
  )
}
