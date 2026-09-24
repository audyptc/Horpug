import { describe, expect, it } from 'vitest'
import { isCompletePreset } from './utils'

describe('isCompletePreset', () => {
  it('needs a name and a positive amount', () => {
    expect(isCompletePreset({ key: 1, name: 'ทำความสะอาด', amount: '500' })).toBe(true)
    expect(isCompletePreset({ key: 1, name: '  ', amount: '500' })).toBe(false)
    expect(isCompletePreset({ key: 1, name: 'ค่าซ่อม', amount: '0' })).toBe(false)
    expect(isCompletePreset({ key: 1, name: 'ค่าซ่อม', amount: '' })).toBe(false)
    expect(isCompletePreset({ key: 1, name: 'ค่าซ่อม', amount: 'abc' })).toBe(false)
  })
})
