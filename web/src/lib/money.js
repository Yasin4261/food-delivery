// Money formatting — the single place the app renders an amount.
//
// The platform charges whatever the API's CURRENCY is configured to (TRY by
// default); nothing in the UI may hardcode a symbol, or customers get shown one
// currency and charged another (#125). Keep VITE_CURRENCY in sync with the
// backend's CURRENCY.

import { i18n } from '@/i18n'

// CURRENCY is the ISO-4217 code every amount is denominated in.
export const CURRENCY = import.meta.env.VITE_CURRENCY || 'TRY'

// formatMoney renders an amount in the platform currency, localised to the
// active UI language (₺45,00 in Turkish; ₺45.00 in English). Nullish amounts
// render as zero rather than "NaN".
export function formatMoney(amount, locale) {
  const value = Number(amount ?? 0)
  const lang = locale || i18n.global.locale.value || 'tr'
  try {
    return new Intl.NumberFormat(lang, { style: 'currency', currency: CURRENCY }).format(value)
  } catch {
    // Unknown currency code configured — fall back to an unambiguous form
    // rather than throwing inside a render.
    return `${value.toFixed(2)} ${CURRENCY}`
  }
}
