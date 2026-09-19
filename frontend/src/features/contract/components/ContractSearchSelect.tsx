import { api, type ApiPage } from '@/shared/api/client'
import { RemoteCombobox } from '@/shared/components/ui/remote-combobox'
import type { ApiContract } from '../types'
import { formatContractLabel } from '../utils'

const SEARCH_LIMIT = 20

type ContractSearchSelectProps = {
  selectedLabel: string
  onSelectContract: (contract: ApiContract) => void
  placeholder: string
  searchPlaceholder: string
  noResultsLabel: string
  disabled?: boolean
}

// Searches active contracts server-side as the user types (tenant, room or
// dormitory), so a contract past the list endpoint's page cap stays pickable.
export function ContractSearchSelect({
  selectedLabel,
  onSelectContract,
  placeholder,
  searchPlaceholder,
  noResultsLabel,
  disabled = false,
}: ContractSearchSelectProps) {
  return (
    <RemoteCombobox<ApiContract>
      fetchItems={(query) =>
        api
          .get<ApiPage<ApiContract[]>>('/contracts', {
            params: { status: 'active', q: query.trim() || undefined, per_page: SEARCH_LIMIT },
          })
          .then(({ data }) => data.data)
      }
      getKey={(contract) => contract.id}
      renderItem={(contract) => <span className="truncate">{formatContractLabel(contract)}</span>}
      onSelect={onSelectContract}
      selectedLabel={selectedLabel}
      placeholder={placeholder}
      searchPlaceholder={searchPlaceholder}
      emptyText={noResultsLabel}
      disabled={disabled}
    />
  )
}
