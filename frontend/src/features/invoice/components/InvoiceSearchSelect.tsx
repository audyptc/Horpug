import { api, type ApiPage } from '@/shared/api/client'
import { RemoteCombobox } from '@/shared/components/ui/remote-combobox'
import type { ApiInvoice } from '../types'
import { formatInvoiceLabel } from '../utils'

const SEARCH_LIMIT = 20

type InvoiceSearchSelectProps = {
  selectedLabel: string
  onSelectInvoice: (invoice: ApiInvoice) => void
  placeholder: string
  searchPlaceholder: string
  noResultsLabel: string
  disabled?: boolean
}

// Searches invoices server-side as the user types (tenant, room or dormitory),
// for picking the invoice a payment is recorded against. Cancelled invoices
// can't take a payment, so they're dropped from the results; that happens after
// the page is fetched, so a search where many of the matches are cancelled can
// show a little fewer than SEARCH_LIMIT — narrowing the search brings the rest.
export function InvoiceSearchSelect({
  selectedLabel,
  onSelectInvoice,
  placeholder,
  searchPlaceholder,
  noResultsLabel,
  disabled = false,
}: InvoiceSearchSelectProps) {
  return (
    <RemoteCombobox<ApiInvoice>
      fetchItems={(query) =>
        api
          .get<ApiPage<ApiInvoice[]>>('/invoices', {
            params: { q: query.trim() || undefined, per_page: SEARCH_LIMIT },
          })
          .then(({ data }) => data.data.filter((invoice) => invoice.status !== 'cancelled'))
      }
      getKey={(invoice) => invoice.id}
      renderItem={(invoice) => <span className="truncate">{formatInvoiceLabel(invoice)}</span>}
      onSelect={onSelectInvoice}
      selectedLabel={selectedLabel}
      placeholder={placeholder}
      searchPlaceholder={searchPlaceholder}
      emptyText={noResultsLabel}
      disabled={disabled}
    />
  )
}
