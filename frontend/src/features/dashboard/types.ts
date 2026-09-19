// A section is null when the signed-in role can't read the menu it comes from.
export type ApiDashboardSummary = {
  rooms: {
    total: number
    available: number
    occupied: number
    maintenance: number
  } | null
  contracts: {
    total: number
    active: number
  } | null
  invoices: {
    outstanding_count: number
    outstanding_amount: number
  } | null
  repairs: {
    pending: number
    in_progress: number
  } | null
}
