import { api } from '@/shared/api/client'
import { RemoteCombobox } from '@/shared/components/ui/remote-combobox'
import type { ApiRole } from '../types'

const SEARCH_LIMIT = 20

type RoleSearchSelectProps = {
  // Owned by the caller: the picked role usually isn't in the current results,
  // so the label can't be derived from them.
  selectedLabel: string
  onSelectRole: (role: ApiRole) => void
  placeholder: string
  searchPlaceholder: string
  noResultsLabel: string
  disabled?: boolean
}

// Searches /roles/active server-side as the user types (by name), so roles
// past the API's preload cap can still be picked.
export function RoleSearchSelect({
  selectedLabel,
  onSelectRole,
  placeholder,
  searchPlaceholder,
  noResultsLabel,
  disabled,
}: RoleSearchSelectProps) {
  return (
    <RemoteCombobox<ApiRole>
      fetchItems={(query) =>
        api
          .get<ApiRole[]>('/roles/active', { params: { q: query.trim() || undefined, limit: SEARCH_LIMIT } })
          .then(({ data }) => data)
      }
      getKey={(role) => role.id}
      renderItem={(role) => <span className="truncate font-medium">{role.name}</span>}
      onSelect={onSelectRole}
      selectedLabel={selectedLabel}
      placeholder={placeholder}
      searchPlaceholder={searchPlaceholder}
      emptyText={noResultsLabel}
      disabled={disabled}
    />
  )
}
