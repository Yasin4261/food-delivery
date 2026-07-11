import { createI18n } from 'vue-i18n'
import en from './en'
import tr from './tr'

// Saved choice wins; otherwise Turkish browsers get Turkish, everyone else
// English.
function initialLocale() {
  const saved = localStorage.getItem('locale')
  if (saved === 'tr' || saved === 'en') return saved
  return (navigator.language || '').toLowerCase().startsWith('tr') ? 'tr' : 'en'
}

export const i18n = createI18n({
  legacy: false,
  globalInjection: true, // $t available in every template
  locale: initialLocale(),
  fallbackLocale: 'en',
  messages: { en, tr },
})

// setLocale switches the UI language and persists the choice.
export function setLocale(locale) {
  i18n.global.locale.value = locale
  localStorage.setItem('locale', locale)
  document.documentElement.lang = locale
}
