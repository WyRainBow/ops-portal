'use client'

import { Card } from '../../../components/Ui'
import { useTheme } from '../../../components/ThemeProvider'
import type { Theme } from '../../../lib/theme'

const THEMES: { value: Theme; label: string }[] = [
  { value: 'light', label: '浅色' },
  { value: 'dark', label: '深色' },
  { value: 'system', label: '跟随系统' },
]

export default function SettingsPage() {
  const { theme, setTheme } = useTheme()

  return (
    <div className="space-y-6">
      <div>
        <div className="text-xs uppercase tracking-[0.22em] text-slate-300/70">Settings</div>
        <div className="mt-1 text-2xl font-semibold">设置</div>
        <div className="mt-2 text-sm text-slate-200/70">外观与偏好</div>
      </div>

      <Card title="外观" subtitle="选择浅色、深色或跟随系统（prefers-color-scheme）">
        <div className="flex flex-wrap gap-2">
          {THEMES.map((t) => (
            <button
              key={t.value}
              type="button"
              onClick={() => setTheme(t.value)}
              className={[
                'rounded-2xl border px-4 py-2 text-sm transition',
                theme === t.value
                  ? 'border-slate-900 bg-slate-900 text-white dark:border-slate-100 dark:bg-slate-100 dark:text-slate-900'
                  : 'border-slate-300 bg-white text-slate-700 hover:bg-slate-50 dark:border-slate-600 dark:bg-white/5 dark:text-slate-200 dark:hover:bg-white/10',
              ].join(' ')}
            >
              {t.label}
            </button>
          ))}
        </div>
      </Card>
    </div>
  )
}
