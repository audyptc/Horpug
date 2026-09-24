import { describe, expect, it } from 'vitest'
import { sumPaymentItems } from './utils'

describe('sumPaymentItems', () => {
  it('rounds to the satang', () => {
    expect(sumPaymentItems([{ amount: '0.1' }, { amount: '0.2' }])).toBe(0.3)
    expect(sumPaymentItems([{ amount: '1500.55' }, { amount: '2499.45' }])).toBe(4000)
  })

  it('treats blank or partly typed amounts as zero', () => {
    expect(sumPaymentItems([{ amount: '' }, { amount: '-' }, { amount: '100' }])).toBe(100)
    expect(sumPaymentItems([])).toBe(0)
  })
})
