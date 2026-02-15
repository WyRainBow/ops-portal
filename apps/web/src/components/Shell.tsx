'use client'

import Link from 'next/link'
import { usePathname, useRouter } from 'next/navigation'
import { clearToken, getToken, parseJwtRole } from '../lib/auth'

const nav = [
  { href: '/overview', label: '概览' },
  { href: '/users', label: '用户' },
  { href: '/members', label: '成员' },
  { href: '/logs/requests', label: '请求日志' },
  { href: '/logs/errors', label: '错误日志' },
  { href: '/traces', label: '链路追踪' },
  { href: '/permissions', label: '权限审计' },
  { href: '/runtime', label: '运行状态' },
  { href: '/observability', label: '可观测' },
  { href: '/assistant/chat', label: '助手对话' },
  { href: '/assistant/aiops', label: '告警分析' },
]

export function Shell({ children }: { children: React.ReactNode }) {
  const pathname = usePathname()
  const router = useRouter()
  const token = getToken()
  const role = parseJwtRole(token)

  const logout = () => {
    clearToken()
    router.replace('/login')
  }

  return (
    <div className="min-h-screen text-slate-100">
      <div className="ops-grid">
        <div className="mx-auto flex max-w-[1320px] gap-6 px-4 py-6">
          <aside className="w-[260px] shrink-0">
            <div className="ops-card rounded-2xl p-4">
              <div className="text-xs uppercase tracking-[0.24em] text-[color:var(--muted)]">Ops Portal</div>
              <div className="mt-1 text-base font-semibold">可观测与运维助手</div>
              <div className="mt-2 text-xs text-[color:var(--muted)]">role: {role}</div>
            <button
              className="mt-4 w-full rounded-xl border border-[color:var(--stroke)] bg-black/20 px-3 py-2 text-sm hover:bg-black/30"
              onClick={logout}
              type="button"
            >
              退出登录
            </button>
            </div>
            <nav className="mt-4 space-y-2">
              {nav.map((n) => {
                const active = pathname === n.href || (n.href !== '/overview' && pathname?.startsWith(n.href))
                return (
                  <Link
                    key={n.href}
                    href={n.href}
                    className={[
                      'block rounded-2xl border px-4 py-2 text-sm transition',
                      active
                        ? 'border-white/20 bg-white/10 font-semibold'
                        : 'border-transparent text-slate-200/80 hover:border-white/10 hover:bg-white/5',
                    ].join(' ')}
                  >
                    {n.label}
                  </Link>
                )
              })}
            </nav>
          </aside>
          <main className="min-w-0 flex-1">{children}</main>
        </div>
      </div>
    </div>
  )
}
