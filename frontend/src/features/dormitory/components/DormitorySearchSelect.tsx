import { api } from '@/shared/api/client'
import { RemoteCombobox } from '@/shared/components/ui/remote-combobox'
import type { ApiDormitory } from '../types'

const SEARCH_LIMIT = 20

type DormitorySearchSelectProps = {
  // Owned by the caller: the picked dormitory usually isn't in the current
  // results, so the label can't be derived from them.
  selectedLabel: string
  onSelectDormitory: (dormitory: ApiDormitory) => void
  placeholder: string
  searchPlaceholder: string
  noResultsLabel: string
  disabled?: boolean
}

// Searches /dormitories/active server-side as the user types (by name), so
// dormitories past the API's preload cap can still be picked.
export function DormitorySearchSelect({
  selectedLabel,
  onSelectDormitory,
  placeholder,
  searchPlaceholder,
  noResultsLabel,
  disabled,
}: DormitorySearchSelectProps) {
  return (
    <RemoteCombobox<ApiDormitory>
      fetchItems={(query) =>
        api
          .get<ApiDormitory[]>('/dormitories/active', {
            params: { q: query.trim() || undefined, limit: SEARCH_LIMIT },
          })
          .then(({ data }) => data)
      }
      getKey={(dormitory) => dormitory.id}
      renderItem={(dormitory) => <span className="truncate font-medium">{dormitory.name}</span>}
      onSelect={onSelectDormitory}
      selectedLabel={selectedLabel}
      placeholder={placeholder}
      searchPlaceholder={searchPlaceholder}
      emptyText={noResultsLabel}
      disabled={disabled}
    />
  )
}
