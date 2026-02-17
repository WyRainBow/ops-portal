'use client'

import { createContext, useCallback, useContext, useEffect, useState } from 'react'
import { applyTheme, getTheme, setTheme as persistTheme, type Theme } from '../lib/theme'

const ThemeContext = createContext<{ theme: Theme; setTheme: (t: Theme) => void } | null>(null)

export function ThemeProvider({ children }: { children: React.ReactNode }) {
  const [theme, setThemeState] = useState<Theme>('light')

  useEffect(() => {
    const t = getTheme()
    setThemeState(t)
    applyTheme(t)
    const onStorage = () => {
      setThemeState(getTheme())
      applyTheme(getTheme())
    }
    window.addEventListener('ops-portal-theme-change', onStorage)
    const mq = window.matchMedia('(prefers-color-scheme: dark)')
    const onPrefersChange = () => applyTheme(getTheme())
    mq.addEventListener('change', onPrefersChange)
    return () => {
      window.removeEventListener('ops-portal-theme-change', onStorage)
      mq.removeEventListener('change', onPrefersChange)
    }
  }, [])

  const setTheme = useCallback((t: Theme) => {
    setThemeState(t)
    persistTheme(t)
  }, [])

  return (
    <ThemeContext.Provider value={{ theme, setTheme }}>
      {children}
    </ThemeContext.Provider>
  )
}

export function useTheme() {
  const ctx = useContext(ThemeContext)
  if (!ctx) return { theme: 'light' as Theme, setTheme: (_: Theme) => {} }
  return ctx
}
