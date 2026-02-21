'use client'

import { useEffect, useMemo, useState } from 'react'
import { getToken } from '../../../lib/auth'
import { getApiRoutes } from '../../../lib/api'
import { Badge, Button, Card, Input, Select } from '../../../components/Ui'

type RouteItem = {
  method: string
  path: string
  summary?: string
  operation_id?: string
  tags?: string[]
  deprecated?: boolean
  source?: string
}

type ProjectTab = 'all' | 'ops-portal' | 'resume-agent'

export default function InterfacesPage() {
  const token = getToken() || ''

  const [q, setQ] = useState('')
  const [method, setMethod] = useState('')
  const [tag, setTag] = useState('')
  const [projectTab, setProjectTab] = useState<ProjectTab>('all')
  const [hideDocs, setHideDocs] = useState(true)

  const [loading, setLoading] = useState(false)
  const [err, setErr] = useState<string | null>(null)
  const [data, setData] = useState<any | null>(null)
  const [copiedKey, setCopiedKey] = useState<string | null>(null)

  const query = useMemo(() => {
    return {
      q: q.trim() || undefined,
      method: method || undefined,
      tag: tag || undefined,
      source: projectTab === 'all' ? undefined : projectTab,
      hide_docs: hideDocs,
    }
  }, [q, method, tag, projectTab, hideDocs])

  const refresh = async () => {
    setLoading(true)
    setErr(null)
    try {
      const res = await getApiRoutes(token, query)
      setData(res)
    } catch (e: any) {
      setErr(e?.message || String(e))
    } finally {
      setLoading(false)
    }
  }

  useEffect(() => {
    refresh()
    // eslint-disable-next-line react-hooks/exhaustive-deps
  }, [JSON.stringify(query)])

  const items = (data?.items || []) as RouteItem[]
  const total = Number(data?.total || 0)
  const methods = (data?.methods || {}) as Record<string, number>
  const tags = (data?.tags || {}) as Record<string, number>
  const sources = (data?.sources || {}) as Record<string, number>

  const tagOptions = Object.keys(tags)
    .filter((t) => t && t !== '_')
    .sort((a, b) => a.localeCompare(b))

  const methodOptions = Object.keys(methods).sort((a, b) => a.localeCompare(b))

  const countOps = Number(sources['ops-portal'] || 0)
  const countResume = Number(sources['resume-agent'] || 0)

  return (
    <div className="space-y-6">
      <div className="flex items-end justify-between gap-4">
        <div>
          <div className="text-xs uppercase tracking-[0.22em] text-slate-300/70">Inventory</div>
          <div className="mt-1 text-2xl font-semibold">接口清单</div>
          <div className="mt-2 text-sm text-slate-200/70">从 ops-portal + resume-agent OpenAPI 聚合接口列表，做成可搜索的 Swagger 视图。</div>
        </div>
        <div className="flex flex-wrap items-center gap-2">
          <Button
            tone="ghost"
            onClick={() => {
              const origin = window.location.origin || 'http://127.0.0.1:18080'
              window.open(`${origin}/swagger`, '_blank', 'noreferrer')
            }}
            type="button"
          >
            打开 Swagger
          </Button>
          <Button tone="primary" onClick={refresh} disabled={loading} type="button">
            {loading ? '刷新中…' : '刷新'}
          </Button>
        </div>
      </div>

      <div className="grid grid-cols-1 gap-3 md:grid-cols-4">
        <div className="ops-card rounded-3xl p-4">
          <div className="text-xs uppercase tracking-[0.24em] text-slate-200/60">Total</div>
          <div className="ops-display mt-2 text-3xl font-semibold">{total}</div>
          <div className="mt-2 text-xs text-slate-200/65">registered endpoints</div>
        </div>
        <div className="ops-card rounded-3xl p-4">
          <div className="text-xs uppercase tracking-[0.24em] text-slate-200/60">Methods</div>
          <div className="mt-3 flex flex-wrap gap-2">
            {Object.entries(methods)
              .sort(([a], [b]) => a.localeCompare(b))
              .map(([m, n]) => (
                <Badge key={m} tone={methodTone(m)}>
                  {m} · {n}
                </Badge>
              ))}
            {Object.keys(methods).length === 0 ? <div className="text-xs text-slate-200/60">-</div> : null}
          </div>
        </div>
        <div className="ops-card rounded-3xl p-4">
          <div className="text-xs uppercase tracking-[0.24em] text-slate-200/60">Top tags</div>
          <div className="mt-3 flex flex-wrap gap-2">
            {Object.entries(tags)
              .filter(([t]) => t && t !== '_')
              .sort((a, b) => b[1] - a[1])
              .slice(0, 8)
              .map(([t, n]) => (
                <Badge key={t} tone="neutral">
                  {t} · {n}
                </Badge>
              ))}
            {Object.keys(tags).filter((t) => t && t !== '_').length === 0 ? (
              <div className="text-xs text-slate-200/60">-</div>
            ) : null}
          </div>
        </div>
        <div className="ops-card rounded-3xl p-4">
          <div className="text-xs uppercase tracking-[0.24em] text-slate-200/60">Projects</div>
          <div className="mt-3 flex flex-wrap gap-2">
            <Badge tone="neutral">ops-portal · {countOps}</Badge>
            <Badge tone="neutral">resume-agent · {countResume}</Badge>
          </div>
        </div>
      </div>

      <Card
        title="筛选"
        subtitle="支持 path/summary/tags 模糊匹配；默认隐藏 /swagger 与 /api.json。"
        right={err ? <Badge tone="bad">{err}</Badge> : <Badge tone="neutral">items: {items.length}</Badge>}
      >
        {!countResume ? (
          <div className="mb-3 rounded-xl border border-amber-300/30 bg-amber-400/10 px-3 py-2 text-xs text-amber-100/90">
            未拉取到 resume-agent 接口。请确认后端可访问 `OPS_PORTAL_RESUME_OPENAPI_URL`（默认 `http://127.0.0.1:9000/openapi.json`）。
          </div>
        ) : null}

        <div className="mb-4 inline-flex rounded-xl border border-white/10 bg-black/15 p-1">
          <button
            type="button"
            onClick={() => setProjectTab('all')}
            className={`rounded-lg px-3 py-1.5 text-xs ${projectTab === 'all' ? 'bg-white text-black' : 'text-slate-200/80 hover:bg-white/10'}`}
          >
            全部 ({total})
          </button>
          <button
            type="button"
            onClick={() => setProjectTab('ops-portal')}
            className={`rounded-lg px-3 py-1.5 text-xs ${projectTab === 'ops-portal' ? 'bg-white text-black' : 'text-slate-200/80 hover:bg-white/10'}`}
          >
            ops-portal ({countOps})
          </button>
          <button
            type="button"
            onClick={() => setProjectTab('resume-agent')}
            className={`rounded-lg px-3 py-1.5 text-xs ${projectTab === 'resume-agent' ? 'bg-white text-black' : 'text-slate-200/80 hover:bg-white/10'}`}
          >
            resume-agent ({countResume})
          </button>
        </div>

        <div className="grid grid-cols-1 gap-3 md:grid-cols-4">
          <div className="md:col-span-2">
            <div className="mb-1 text-xs text-slate-200/70">search</div>
            <Input value={q} onChange={(e) => setQ(e.target.value)} placeholder="例如: /api/admin  或  login  或  loki" />
          </div>
          <div>
            <div className="mb-1 text-xs text-slate-200/70">method</div>
            <Select value={method} onChange={(e) => setMethod(e.target.value)}>
              <option value="">All</option>
              {methodOptions.map((m) => (
                <option key={m} value={m}>
                  {m}
                </option>
              ))}
            </Select>
          </div>
          <div>
            <div className="mb-1 text-xs text-slate-200/70">tag</div>
            <Select value={tag} onChange={(e) => setTag(e.target.value)}>
              <option value="">All</option>
              {tagOptions.map((t) => (
                <option key={t} value={t}>
                  {t}
                </option>
              ))}
            </Select>
          </div>
        </div>

        <div className="mt-4 flex flex-wrap items-center gap-3">
          <label className="inline-flex select-none items-center gap-2 text-xs text-slate-200/70">
            <input
              type="checkbox"
              checked={hideDocs}
              onChange={(e) => setHideDocs(e.target.checked)}
              className="h-4 w-4 rounded border-white/15 bg-black/40"
            />
            隐藏文档路由
          </label>
          <Button
            tone="ghost"
            onClick={() => {
              setQ('')
              setMethod('')
              setTag('')
              setProjectTab('all')
              setHideDocs(true)
            }}
            disabled={loading}
            type="button"
          >
            清空
          </Button>
        </div>
      </Card>

      <Card title="接口列表" subtitle="点击行右侧按钮可复制 curl（使用 127.0.0.1，适合 SSH tunnel / 服务器本机）。">
        <div className="overflow-auto rounded-2xl border border-white/10">
          <table className="w-full min-w-[980px] text-left text-sm">
            <thead className="bg-white/5 text-xs text-slate-200/70">
              <tr>
                <th className="px-3 py-2">method</th>
                <th className="px-3 py-2">path</th>
                <th className="px-3 py-2">summary</th>
                <th className="px-3 py-2">tag</th>
                <th className="px-3 py-2">source</th>
                <th className="sticky right-0 min-w-[100px] shrink-0 bg-white/5 px-3 py-2 text-right">action</th>
              </tr>
            </thead>
            <tbody className="divide-y divide-white/5">
              {items.map((it, index) => {
                const tag1 = (it.tags && it.tags[0]) || ''
                const rowKey = `${it.source ?? 'unknown'}-${it.method}-${it.path}-${index}`
                return (
                  <tr key={rowKey} className="group hover:bg-white/5">
                    <td className="px-3 py-2">
                      <Badge tone={methodTone(it.method)}>{it.method}</Badge>
                    </td>
                    <td className="px-3 py-2">
                      <div className="flex items-center gap-2">
                        <span className="font-mono text-xs text-[color:var(--accent2)]">{it.path}</span>
                        {it.deprecated ? <Badge tone="warn">deprecated</Badge> : null}
                      </div>
                    </td>
                    <td className="px-3 py-2 text-slate-100">{it.summary || <span className="text-slate-200/50">-</span>}</td>
                    <td className="px-3 py-2 text-slate-200/75">{tag1 || <span className="text-slate-200/50">_</span>}</td>
                    <td className="px-3 py-2 text-slate-200/75">{it.source || <span className="text-slate-200/50">_</span>}</td>
                    <td className="sticky right-0 min-w-[100px] shrink-0 bg-transparent px-3 py-2 text-right group-hover:bg-white/5">
                      <Button
                        tone={copiedKey === rowKey ? 'primary' : 'ghost'}
                        type="button"
                        onClick={async () => {
                          const curl = curlFor(it.method, it.path)
                          try {
                            await navigator.clipboard.writeText(curl)
                            setCopiedKey(rowKey)
                            setTimeout(() => setCopiedKey(null), 1500)
                          } catch {
                            setCopiedKey(null)
                          }
                        }}
                        title="复制 curl"
                      >
                        {copiedKey === rowKey ? '已复制' : '复制 curl'}
                      </Button>
                    </td>
                  </tr>
                )
              })}
              {items.length === 0 ? (
                <tr>
                  <td className="px-3 py-6 text-center text-slate-200/60" colSpan={6}>
                    无数据
                  </td>
                </tr>
              ) : null}
            </tbody>
          </table>
        </div>
      </Card>
    </div>
  )
}

function methodTone(m: string): 'ok' | 'warn' | 'bad' | 'neutral' {
  const u = String(m || '').toUpperCase()
  if (u === 'GET' || u === 'HEAD' || u === 'OPTIONS') return 'ok'
  if (u === 'POST' || u === 'PUT' || u === 'PATCH') return 'warn'
  if (u === 'DELETE') return 'bad'
  return 'neutral'
}

function curlFor(method: string, path: string) {
  const m = String(method || 'GET').toUpperCase()
  const base = typeof window !== 'undefined' && window.location?.origin ? window.location.origin : 'http://127.0.0.1:18080'
  const p = path.startsWith('/api/') ? path : `/api${path.startsWith('/') ? path : `/${path}`}`
  if (m === 'GET') return `curl -sS -H 'Authorization: Bearer <TOKEN>' '${base}${p}'`
  return `curl -sS -X ${m} -H 'content-type: application/json' -H 'Authorization: Bearer <TOKEN>' '${base}${p}' -d '{}'`
}
