import { describe, it, expect } from 'vitest'
import { formatMoney, CURRENCY } from './money'

// The platform charges the API's configured currency; the UI must never render
// a hardcoded symbol (#125 — it used to show "$" while charging TRY).
describe('formatMoney', () => {
  it('defaults to the platform currency', () => {
    expect(CURRENCY).toBe('TRY')
  })

  it('never renders a dollar sign for the default currency', () => {
    expect(formatMoney(45, 'tr')).not.toContain('$')
    expect(formatMoney(45, 'en')).not.toContain('$')
  })

  it('renders the currency per locale convention', () => {
    // Turkish uses the ₺ symbol; English spells the ISO code (Intl's own
    // convention — both unambiguous, neither is a dollar).
    expect(formatMoney(45, 'tr')).toContain('₺')
    expect(formatMoney(45, 'en')).toContain('TRY')
  })

  it('localises the decimal separator', () => {
    // Turkish uses a comma, English a dot.
    expect(formatMoney(45.5, 'tr')).toContain('45,50')
    expect(formatMoney(45.5, 'en')).toContain('45.50')
  })

  it('always shows two decimals', () => {
    expect(formatMoney(45, 'en')).toContain('45.00')
  })

  it('treats null/undefined as zero rather than NaN', () => {
    expect(formatMoney(null, 'en')).toContain('0.00')
    expect(formatMoney(undefined, 'en')).toContain('0.00')
    expect(formatMoney(null, 'en')).not.toContain('NaN')
  })
})
