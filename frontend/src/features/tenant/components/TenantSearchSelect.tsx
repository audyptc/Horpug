import { api } from '@/shared/api/client'
import { RemoteCombobox } from '@/shared/components/ui/remote-combobox'
import type { ApiTenant } from '../types'

const SEARCH_LIMIT = 20

type TenantSearchSelectProps = {
  selectedLabel: string
  onSelectTenant: (tenant: ApiTenant) => void
  // Set both to let an optional tenant be cleared again after it's been picked.
  clearLabel?: string
  onClear?: () => void
  placeholder: string
  searchPlaceholder: string
  noResultsLabel: string
  disabled?: boolean
}

// Searches /tenants/active server-side as the user types (name, phone or id
// card), so tenants past the API's preload cap stay reachable. As with
// RoomSearchSelect, the selected label is owned by the caller.
export function TenantSearchSelect({
  selectedLabel,
  onSelectTenant,
  clearLabel,
  onClear,
  placeholder,
  searchPlaceholder,
  noResultsLabel,
  disabled = false,
}: TenantSearchSelectProps) {
  return (
    <RemoteCombobox<ApiTenant>
      fetchItems={(query) =>
        api
          .get<ApiTenant[]>('/tenants/active', {
            params: { q: query.trim() || undefined, limit: SEARCH_LIMIT },
          })
          .then(({ data }) => data)
      }
      getKey={(tenant) => tenant.id}
      renderItem={(tenant) => (
        <>
          <span className="truncate font-medium">
            {tenant.first_name} {tenant.last_name}
          </span>
          {tenant.phone && <span className="shrink-0 text-muted-foreground">{tenant.phone}</span>}
        </>
      )}
      onSelect={onSelectTenant}
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
