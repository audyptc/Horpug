import type { InvoiceStatus } from './types'

export const INVOICE_PAGE_SIZE_OPTIONS = [10, 20, 50, 100] as const

export const INVOICE_STATUSES: InvoiceStatus[] = ['unpaid', 'paid', 'overdue', 'cancelled']

export function toDateInputValue(value?: string): string {
  return value ? value.slice(0, 10) : ''
}

export function toApiDate(value: string): string {
  return `${value}T00:00:00Z`
}

// <input type="month"> gives "YYYY-MM"; split it into the separate
// period_year/period_month integers the API expects.
export function parsePeriodInputValue(value: string): { year: number; month: number } | null {
  const match = /^(\d{4})-(\d{2})$/.exec(value)
  if (!match) return null
  return { year: Number(match[1]), month: Number(match[2]) }
}

export function toPeriodInputValue(year: number, month: number): string {
  return `${year}-${String(month).padStart(2, '0')}`
}

export function formatPeriod(year: number, month: number): string {
  return `${String(month).padStart(2, '0')}/${year}`
}

// Sums total_amount for meter readings whose reading_date falls in the given
// billing period, mirroring the backend's own [periodStart, periodEnd) filter
// so the create-invoice preview matches what gets billed on submit. Returns
// null when no reading was recorded for the period, so callers can tell that
// apart from a reading that happens to total zero.
export function sumMeterAmountForPeriod(
  readings: { reading_date: string; total_amount: number }[],
  year: number,
  month: number
): number | null {
  const periodValue = toPeriodInputValue(year, month)
  const matches = readings.filter((reading) => toDateInputValue(reading.reading_date).slice(0, 7) === periodValue)
  if (matches.length === 0) return null
  return matches.reduce((sum, reading) => sum + reading.total_amount, 0)
}
