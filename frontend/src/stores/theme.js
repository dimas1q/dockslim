import { computed, ref } from 'vue'

const STORAGE_KEY = 'dockslim_theme'
const SUPPORTED_THEMES = ['light', 'dark']

const theme = ref('light')

const applyTheme = (value, { persist = true } = {}) => {
  const next = SUPPORTED_THEMES.includes(value) ? value : 'light'
  theme.value = next
  if (typeof document !== 'undefined') {
    document.documentElement.dataset.theme = next
    document.documentElement.classList.toggle('dark', next === 'dark')
  }
  if (persist && typeof window !== 'undefined') {
    window.localStorage.setItem(STORAGE_KEY, next)
  }
}

const initTheme = () => {
  if (typeof window === 'undefined') {
    return
  }
  const stored = window.localStorage.getItem(STORAGE_KEY)
  if (stored && SUPPORTED_THEMES.includes(stored)) {
    applyTheme(stored, { persist: false })
    return
  }
  const prefersDark =
    window.matchMedia && window.matchMedia('(prefers-color-scheme: dark)').matches
  applyTheme(prefersDark ? 'dark' : 'light', { persist: false })
}

const toggleTheme = () => {
  applyTheme(theme.value === 'dark' ? 'light' : 'dark')
}

const isDark = computed(() => theme.value === 'dark')

export { SUPPORTED_THEMES, theme, isDark, applyTheme, initTheme, toggleTheme }
