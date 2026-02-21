'use client'

import Link from 'next/link'
import { useEffect, useState } from 'react'
import { usePathname, useRouter } from 'next/navigation'
import { clearToken, getToken, parseJwtRole } from '../lib/auth'

const nav = [
  { href: '/overview', label: '概览', icon: IconRadar },
  { href: '/runtime', label: '运行状态', icon: IconPulse },
  { href: '/observability', label: '可观测', icon: IconTelescope },
  { href: '/interfaces', label: '接口清单', icon: IconApi },
  { href: '/logs/requests', label: '请求日志', icon: IconRoute },
  { href: '/logs/errors', label: '错误日志', icon: IconBug },
  { href: '/traces', label: '链路追踪', icon: IconTrace },
  { href: '/alerts', label: '告警中心', icon: IconAlert },
  { href: '/users', label: '用户', icon: IconUsers },
  { href: '/members', label: '成员', icon: IconId },
  { href: '/permissions', label: '权限审计', icon: IconShield },
  { href: '/ops', label: 'Playbook', icon: IconPlaybook },
  { href: '/assistant/chat', label: '助手对话', icon: IconChat },
  { href: '/assistant/aiops', label: '告警分析', icon: IconAIOps },
  { href: '/settings', label: '设置', icon: IconGear },
]

export function Shell({ children }: { children: React.ReactNode }) {
  const pathname = usePathname()
  const router = useRouter()
  // 仅客户端挂载后读取 role，避免 SSR 与客户端 hydration 不一致（getToken 在服务端为 null）
  const [role, setRole] = useState<string>('—')
  useEffect(() => {
    setRole(parseJwtRole(getToken()))
  }, [])

  const logout = () => {
    clearToken()
    router.replace('/login')
  }

  return (
    <div className="min-h-screen text-slate-900">
      <div className="ops-grid">
        <div className="mx-auto flex max-w-[1320px] gap-6 px-4 py-6">
          <aside className="w-[280px] shrink-0">
            <div className="sticky top-6 space-y-4">
              <div className="ops-card rounded-3xl p-4">
                <div className="flex items-start justify-between gap-3">
                  <div className="min-w-0">
                    <div className="text-xs uppercase tracking-[0.24em] text-[color:var(--muted)]">Ops Portal</div>
                    <div className="ops-display mt-1 text-lg font-semibold">可观测与运维助手</div>
                    <div className="mt-2 text-xs text-[color:var(--muted)]">role: {role}</div>
                  </div>
                  <div className="rounded-2xl border border-slate-200 bg-white p-2">
                    <IconChip />
                  </div>
                </div>

                <button
                  className="mt-4 w-full rounded-2xl border border-slate-300 bg-white px-3 py-2 text-sm hover:bg-slate-100 focus-visible:outline-none focus-visible:ring-4 focus-visible:ring-slate-200"
                  onClick={logout}
                  type="button"
                >
                  退出登录
                </button>
              </div>

              <nav className="ops-card rounded-3xl p-2">
                <div className="px-3 py-2 text-[11px] uppercase tracking-[0.28em] text-slate-200/60">Console</div>
                <div className="space-y-1">
                  {nav.slice(0, 8).map((n) => (
                    <NavItem key={n.href} href={n.href} label={n.label} icon={n.icon} active={isActive(pathname, n.href)} />
                  ))}
                </div>
                <div className="my-2 h-px bg-slate-200" />
                <div className="px-3 py-2 text-[11px] uppercase tracking-[0.28em] text-slate-200/60">Admin</div>
                <div className="space-y-1">
                  {nav.slice(8, 12).map((n) => (
                    <NavItem key={n.href} href={n.href} label={n.label} icon={n.icon} active={isActive(pathname, n.href)} />
                  ))}
                </div>
                <div className="my-2 h-px bg-slate-200" />
                <div className="px-3 py-2 text-[11px] uppercase tracking-[0.28em] text-slate-200/60">Assistant</div>
                <div className="space-y-1">
                  {nav.slice(12).map((n) => (
                    <NavItem key={n.href} href={n.href} label={n.label} icon={n.icon} active={isActive(pathname, n.href)} />
                  ))}
                </div>
              </nav>

              <div className="rounded-3xl border border-slate-200 bg-white p-4">
                <div className="text-xs text-slate-200/60">Shortcut</div>
                <div className="mt-2 font-mono text-xs text-[color:var(--accent2)]">
                  {`{job="resume-backend", stream="error"}`}
                </div>
              </div>
            </div>
          </aside>

          <main className="min-w-0 flex-1">
            <div className="ops-enter">{children}</div>
          </main>
        </div>
      </div>
    </div>
  )
}

function isActive(pathname: string | null, href: string) {
  if (!pathname) return false
  return pathname === href || (href !== '/overview' && pathname.startsWith(href))
}

function NavItem(props: {
  href: string
  label: string
  icon: (p: { active: boolean }) => React.ReactNode
  active: boolean
}) {
  return (
    <Link
      href={props.href}
      className={[
        'group flex items-center gap-3 rounded-2xl px-3 py-2 text-sm transition',
        props.active ? 'bg-slate-100 text-slate-900' : 'text-slate-600 hover:bg-slate-50 hover:text-slate-900',
      ].join(' ')}
    >
      <span
        className={[
          'inline-flex h-8 w-8 items-center justify-center rounded-2xl border',
          props.active ? 'border-slate-300 bg-white' : 'border-slate-200 bg-white group-hover:border-slate-300',
        ].join(' ')}
      >
        {props.icon({ active: props.active })}
      </span>
      <span className="min-w-0 truncate">{props.label}</span>
      <span className="ml-auto h-2 w-2 rounded-full bg-slate-900 opacity-0 transition group-hover:opacity-30" />
      {props.active ? <span className="h-2 w-2 rounded-full bg-slate-900 opacity-80" /> : null}
    </Link>
  )
}

function IconBase(props: { children: React.ReactNode; active?: boolean }) {
  return (
    <svg width="18" height="18" viewBox="0 0 24 24" fill="none" aria-hidden="true">
      <g
        stroke={props.active ? '#111111' : '#334155'}
        strokeWidth="1.6"
        strokeLinecap="round"
        strokeLinejoin="round"
      >
        {props.children}
      </g>
    </svg>
  )
}

function IconRadar({ active }: { active: boolean }) {
  return (
    <IconBase active={active}>
      <path d="M12 21a9 9 0 1 0-9-9" />
      <path d="M12 12l6-6" />
      <path d="M12 12a1 1 0 1 0 0.001 0" />
      <path d="M3 12h3" />
    </IconBase>
  )
}

function IconPulse({ active }: { active: boolean }) {
  return (
    <IconBase active={active}>
      <path d="M3 12h4l2-5 3 10 2-5h7" />
      <path d="M3 19h18" />
    </IconBase>
  )
}

function IconTelescope({ active }: { active: boolean }) {
  return (
    <IconBase active={active}>
      <path d="M6 14l-2 2" />
      <path d="M10 12l4 8" />
      <path d="M14 20h-4" />
      <path d="M4 10l16-6 1 4-16 6-1-4z" />
    </IconBase>
  )
}

function IconApi({ active }: { active: boolean }) {
  return (
    <IconBase active={active}>
      <path d="M7 7h10" />
      <path d="M7 12h10" />
      <path d="M7 17h10" />
      <path d="M5 7a1 1 0 1 0 0.001 0" />
      <path d="M19 17a1 1 0 1 0 0.001 0" />
    </IconBase>
  )
}

function IconRoute({ active }: { active: boolean }) {
  return (
    <IconBase active={active}>
      <path d="M6 5h4" />
      <path d="M14 19h4" />
      <path d="M10 5c6 0 0 14 6 14" />
      <path d="M6 5a1 1 0 1 0 0.001 0" />
      <path d="M18 19a1 1 0 1 0 0.001 0" />
    </IconBase>
  )
}

function IconBug({ active }: { active: boolean }) {
  return (
    <IconBase active={active}>
      <path d="M9 9h6" />
      <path d="M10 5l-1-2" />
      <path d="M14 5l1-2" />
      <path d="M8 12h8v5a4 4 0 0 1-8 0v-5z" />
      <path d="M6 13l2 1" />
      <path d="M18 13l-2 1" />
    </IconBase>
  )
}

function IconTrace({ active }: { active: boolean }) {
  return (
    <IconBase active={active}>
      <path d="M7 7h10" />
      <path d="M7 12h6" />
      <path d="M7 17h10" />
      <path d="M15 12l2-2" />
      <path d="M15 12l2 2" />
    </IconBase>
  )
}

function IconUsers({ active }: { active: boolean }) {
  return (
    <IconBase active={active}>
      <path d="M16 11a4 4 0 1 0-8 0" />
      <path d="M4 21a8 8 0 0 1 16 0" />
    </IconBase>
  )
}

function IconId({ active }: { active: boolean }) {
  return (
    <IconBase active={active}>
      <path d="M7 7h10" />
      <path d="M7 12h7" />
      <path d="M7 17h10" />
      <path d="M4 5v14a2 2 0 0 0 2 2h12a2 2 0 0 0 2-2V5a2 2 0 0 0-2-2H6a2 2 0 0 0-2 2z" />
    </IconBase>
  )
}

function IconShield({ active }: { active: boolean }) {
  return (
    <IconBase active={active}>
      <path d="M12 3l8 4v6c0 5-3 8-8 9-5-1-8-4-8-9V7l8-4z" />
      <path d="M9 12l2 2 4-4" />
    </IconBase>
  )
}

function IconChat({ active }: { active: boolean }) {
  return (
    <IconBase active={active}>
      <path d="M5 6h14v10H8l-3 3V6z" />
      <path d="M8 9h8" />
      <path d="M8 12h6" />
    </IconBase>
  )
}

function IconAIOps({ active }: { active: boolean }) {
  return (
    <IconBase active={active}>
      <path d="M12 2v4" />
      <path d="M12 18v4" />
      <path d="M4 10h4" />
      <path d="M16 10h4" />
      <path d="M7 7l2 2" />
      <path d="M17 7l-2 2" />
      <path d="M7 13l2-2" />
      <path d="M17 13l-2-2" />
      <path d="M12 7a3 3 0 1 0 0 6a3 3 0 0 0 0-6z" />
    </IconBase>
  )
}

function IconPlaybook({ active }: { active: boolean }) {
  return (
    <IconBase active={active}>
      <path d="M14 2H6a2 2 0 0 0-2 2v16a2 2 0 0 0 2 2h12a2 2 0 0 0 2-2V8z" />
      <path d="M14 2v6h6" />
      <path d="M16 13H8" />
      <path d="M16 17H8" />
      <path d="M10 9H8" />
    </IconBase>
  )
}

function IconAlert({ active }: { active: boolean }) {
  return (
    <IconBase active={active}>
      <path d="M12 9v4" />
      <path d="M12 17h.01" />
      <path d="M5.6 5.6l12.8 12.8" />
      <path d="M18.4 5.6L5.6 18.4" />
      <circle cx="12" cy="12" r="10" />
    </IconBase>
  )
}

function IconGear({ active }: { active: boolean }) {
  return (
    <IconBase active={active}>
      <path d="M12 15a3 3 0 1 0 0-6 3 3 0 0 0 0 6z" />
      <path d="M19.4 15a1.65 1.65 0 0 0 .33 1.82l.06.06a2 2 0 0 1 0 2.83 2 2 0 0 1-2.83 0l-.06-.06a1.65 1.65 0 0 0-1.82-.33 1.65 1.65 0 0 0-1 1.51V21a2 2 0 0 1-2 2 2 2 0 0 1-2-2v-.09A1.65 1.65 0 0 0 9 19.4a1.65 1.65 0 0 0-1.82.33l-.06.06a2 2 0 0 1-2.83 0 2 2 0 0 1 0-2.83l.06-.06a1.65 1.65 0 0 0 .33-1.82 1.65 1.65 0 0 0-1.51-1H3a2 2 0 0 1-2-2 2 2 0 0 1 2-2h.09A1.65 1.65 0 0 0 4.6 9a1.65 1.65 0 0 0-.33-1.82l-.06-.06a2 2 0 0 1 0-2.83 2 2 0 0 1 2.83 0l.06.06a1.65 1.65 0 0 0 1.82.33H9a1.65 1.65 0 0 0 1-1.51V3a2 2 0 0 1 2-2 2 2 0 0 1 2 2v.09a1.65 1.65 0 0 0 1 1.51 1.65 1.65 0 0 0 1.82-.33l.06-.06a2 2 0 0 1 2.83 0 2 2 0 0 1 0 2.83l-.06.06a1.65 1.65 0 0 0-.33 1.82V9a1.65 1.65 0 0 0 1.51 1H21a2 2 0 0 1 2 2 2 2 0 0 1-2 2h-.09a1.65 1.65 0 0 0-1.51 1z" />
    </IconBase>
  )
}

function IconChip() {
  return (
    <svg width="18" height="18" viewBox="0 0 24 24" fill="none" aria-hidden="true">
      <g stroke="#111111" strokeWidth="1.6" strokeLinecap="round" strokeLinejoin="round">
        <path d="M9 9h6v6H9V9z" />
        <path d="M4 10h2M4 14h2M18 10h2M18 14h2" />
        <path d="M10 4v2M14 4v2M10 18v2M14 18v2" />
        <path d="M7 7h10v10H7V7z" opacity="0.45" />
      </g>
    </svg>
  )
}
