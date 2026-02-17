export type Theme = 'light' | 'dark' | 'system'

const KEY = 'ops_portal_theme'

export function getTheme(): Theme {
  if (typeof window === 'undefined') return 'light'
  const v = window.localStorage.getItem(KEY)
  if (v === 'light' || v === 'dark' || v === 'system') return v
  return 'light'
}

export function setTheme(theme: Theme) {
  if (typeof window === 'undefined') return
  window.localStorage.setItem(KEY, theme)
  applyTheme(theme)
  window.dispatchEvent(new Event('ops-portal-theme-change'))
}

function resolveDark(theme: Theme): boolean {
  if (theme === 'dark') return true
  if (theme === 'light') return false
  return typeof window !== 'undefined' && window.matchMedia('(prefers-color-scheme: dark)').matches
}

export function applyTheme(theme: Theme) {
  if (typeof document === 'undefined') return
  const dark = resolveDark(theme)
  document.documentElement.setAttribute('data-theme', dark ? 'dark' : 'light')
}
