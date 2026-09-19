// A section is null when the signed-in role can't read the menu it comes from.
export type ApiDashboardSummary = {
  // The calendar month the "this month" figures refer to, as the server sees it.
  period: {
    year: number
    month: number
  }
  rooms: {
    total: number
    available: number
    occupied: number
    maintenance: number
  } | null
  contracts: {
    total: number
    active: number
    // Active contracts ending within the next 30 days.
    expiring_soon: number
    // Contracts still active although their end date has passed.
    past_end: number
  } | null
  invoices: {
    outstanding_count: number
    outstanding_amount: number
    // The part of the outstanding total that is past its due date.
    overdue_count: number
    overdue_amount: number
  } | null
  repairs: {
    pending: number
    in_progress: number
  } | null
  payments: {
    received_this_month: number
  } | null
  expenses: {
    this_month: number
  } | null
  // Occupied rooms with no reading yet this month.
  electricity_readings: {
    missing: number
  } | null
  water_readings: {
    missing: number
  } | null
}
