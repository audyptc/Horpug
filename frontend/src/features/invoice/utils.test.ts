import { describe, expect, it } from 'vitest'
import { formatPeriod, parsePeriodInputValue, sumMeterAmountForPeriod, toPeriodInputValue } from './utils'

describe('period input', () => {
  it('parses <input type="month"> values', () => {
    expect(parsePeriodInputValue('2026-09')).toEqual({ year: 2026, month: 9 })
    expect(parsePeriodInputValue('2026-9')).toBeNull()
    expect(parsePeriodInputValue('')).toBeNull()
  })

  it('round-trips and formats', () => {
    expect(toPeriodInputValue(2026, 1)).toBe('2026-01')
    expect(formatPeriod(2026, 1)).toBe('01/2026')
  })
})

describe('sumMeterAmountForPeriod', () => {
  // Must match the backend's [periodStart, periodEnd) filter so the preview
  // equals what gets billed.
  const readings = [
    { reading_date: '2026-08-31T00:00:00Z', total_amount: 100 },
    { reading_date: '2026-09-01T00:00:00Z', total_amount: 200 },
    { reading_date: '2026-09-30T00:00:00Z', total_amount: 50.5 },
    { reading_date: '2026-10-01T00:00:00Z', total_amount: 400 },
  ]

  it('sums only readings inside the billing month', () => {
    expect(sumMeterAmountForPeriod(readings, 2026, 9)).toBe(250.5)
  })

  it('returns null, not 0, when the month has no reading', () => {
    expect(sumMeterAmountForPeriod(readings, 2026, 11)).toBeNull()
    expect(sumMeterAmountForPeriod([{ reading_date: '2026-11-02T00:00:00Z', total_amount: 0 }], 2026, 11)).toBe(0)
  })
})
