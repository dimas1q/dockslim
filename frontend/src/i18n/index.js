import { createI18n } from 'vue-i18n'
import en from './locales/en.json'
import ru from './locales/ru.json'

const SUPPORTED_LOCALES = ['ru', 'en']
const STORAGE_KEY = 'dockslim_locale'

const detectLocale = () => {
  if (typeof window === 'undefined') {
    return 'ru'
  }
  const stored = window.localStorage.getItem(STORAGE_KEY)
  if (stored && SUPPORTED_LOCALES.includes(stored)) {
    return stored
  }
  return 'ru'
}

const setDocumentLang = (locale) => {
  if (typeof document !== 'undefined') {
    document.documentElement.lang = locale
  }
}

const i18n = createI18n({
  legacy: false,
  locale: detectLocale(),
  fallbackLocale: 'ru',
  messages: {
    en,
    ru,
  },
})

setDocumentLang(i18n.global.locale.value)

const setLocale = (locale) => {
  if (!SUPPORTED_LOCALES.includes(locale)) {
    return
  }
  i18n.global.locale.value = locale
  if (typeof window !== 'undefined') {
    window.localStorage.setItem(STORAGE_KEY, locale)
  }
  setDocumentLang(locale)
}

export { SUPPORTED_LOCALES, setLocale }
export default i18n
