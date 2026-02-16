'use client'

import { useEffect, useMemo, useState } from 'react'
import { getToken } from '../../../../lib/auth'
import { getRequestLogs } from '../../../../lib/api'
import { clampInt, formatRFC3339 } from '../../../../lib/format'
import { Badge, Button, Card, Input, Select } from '../../../../components/Ui'

export default function RequestLogsPage() {
  const token = getToken() || ''

  const [page, setPage] = useState(1)
  const [pageSize, setPageSize] = useState(50)
  const [traceId, setTraceId] = useState('')
  const [path, setPath] = useState('')
  const [statusMode, setStatusMode] = useState<'all' | '4xx+' | '5xx+' | 'exact'>('all')
  const [statusExact, setStatusExact] = useState('')

  const [loading, setLoading] = useState(false)
  const [err, setErr] = useState<string | null>(null)
  const [data, setData] = useState<any | null>(null)

  const query = useMemo(() => {
    const q: Record<string, any> = { page, page_size: pageSize }
    if (traceId.trim()) q.trace_id = traceId.trim()
    if (path.trim()) q.path = path.trim()
    if (statusMode === '4xx+') q.min_status_code = 400
    if (statusMode === '5xx+') q.min_status_code = 500
    if (statusMode === 'exact' && statusExact.trim()) q.status_code = Number(statusExact.trim())
    return q
  }, [page, pageSize, traceId, path, statusMode, statusExact])

  const refresh = async () => {
    setLoading(true)
    setErr(null)
    try {
      const res = await getRequestLogs(token, query)
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

  const items = (data?.items || []) as any[]
  const total = Number(data?.total || 0)
  const totalPages = Math.max(1, Math.ceil(total / pageSize))

  return (
    <div className="space-y-6">
      <div className="flex items-end justify-between gap-4">
        <div>
          <div className="text-xs uppercase tracking-[0.22em] text-slate-300/70">Logs</div>
          <div className="mt-1 text-2xl font-semibold">请求日志</div>
          <div className="mt-2 text-sm text-slate-200/70">来自 PostgreSQL 的 API 请求日志（可按 trace/path/status 筛选）。</div>
        </div>
        <Button tone="primary" onClick={refresh} disabled={loading} type="button">
          {loading ? '刷新中…' : '刷新'}
        </Button>
      </div>

      <Card
        title="筛选"
        subtitle="建议：先用 5xx+ 找到错误请求，再复制 trace_id 去链路追踪页看详情。"
        right={err ? <Badge tone="bad">{err}</Badge> : <Badge tone="neutral">total: {total}</Badge>}
      >
        <div className="grid grid-cols-1 gap-3 md:grid-cols-4">
          <div>
            <div className="mb-1 text-xs text-slate-200/70">trace_id</div>
            <Input value={traceId} onChange={(e) => setTraceId(e.target.value)} placeholder="例如: 2f7a..." />
          </div>
          <div>
            <div className="mb-1 text-xs text-slate-200/70">path contains</div>
            <Input value={path} onChange={(e) => setPath(e.target.value)} placeholder="/api/auth" />
          </div>
          <div>
            <div className="mb-1 text-xs text-slate-200/70">status</div>
            <Select value={statusMode} onChange={(e) => setStatusMode(e.target.value as any)}>
              <option value="all">All</option>
              <option value="4xx+">4xx+</option>
              <option value="5xx+">5xx+</option>
              <option value="exact">Exact</option>
            </Select>
          </div>
          <div>
            <div className="mb-1 text-xs text-slate-200/70">exact status_code</div>
            <Input
              value={statusExact}
              onChange={(e) => setStatusExact(e.target.value)}
              placeholder="200/401/503..."
              disabled={statusMode !== 'exact'}
            />
          </div>
        </div>

        <div className="mt-4 flex flex-wrap items-center gap-3">
          <div className="text-xs text-slate-200/70">
            Page {page}/{totalPages}
          </div>
          <div className="w-[120px]">
            <Select value={String(pageSize)} onChange={(e) => setPageSize(clampInt(Number(e.target.value), 5, 200))}>
              {[20, 50, 100, 200].map((n) => (
                <option key={n} value={String(n)}>
                  {n} / page
                </option>
              ))}
            </Select>
          </div>
          <Button tone="ghost" onClick={() => setPage((p) => Math.max(1, p - 1))} disabled={page <= 1 || loading} type="button">
            上一页
          </Button>
          <Button
            tone="ghost"
            onClick={() => setPage((p) => Math.min(totalPages, p + 1))}
            disabled={page >= totalPages || loading}
            type="button"
          >
            下一页
          </Button>
          <Button
            tone="ghost"
            onClick={() => {
              setTraceId('')
              setPath('')
              setStatusMode('all')
              setStatusExact('')
              setPage(1)
            }}
            disabled={loading}
            type="button"
          >
            清空
          </Button>
        </div>
      </Card>

      <Card title="结果" subtitle="点击 trace_id 可复制后跳去链路页（你也可以直接复制）。">
        <div className="overflow-auto rounded-2xl border border-white/10">
          <table className="w-full min-w-[980px] text-left text-sm">
            <thead className="bg-white/5 text-xs text-slate-200/70">
              <tr>
                <th className="px-3 py-2">时间</th>
                <th className="px-3 py-2">status</th>
                <th className="px-3 py-2">method</th>
                <th className="px-3 py-2">path</th>
                <th className="px-3 py-2">latency(ms)</th>
                <th className="px-3 py-2">user_id</th>
                <th className="px-3 py-2">ip</th>
                <th className="px-3 py-2">trace_id</th>
              </tr>
            </thead>
            <tbody className="divide-y divide-white/5">
              {items.map((it) => {
                const status = Number(it.status_code)
                const tone = status >= 500 ? 'bad' : status >= 400 ? 'warn' : status >= 300 ? 'neutral' : 'ok'
                return (
                  <tr key={String(it.id)} className="hover:bg-white/5">
                    <td className="px-3 py-2 text-slate-200/80">{formatRFC3339(it.created_at)}</td>
                    <td className="px-3 py-2">
                      <Badge tone={tone as any}>{status}</Badge>
                    </td>
                    <td className="px-3 py-2 font-mono text-xs text-slate-200/80">{it.method}</td>
                    <td className="px-3 py-2 text-slate-100">{it.path}</td>
                    <td className="px-3 py-2 text-slate-200/80">{Number(it.latency_ms || 0).toFixed(1)}</td>
                    <td className="px-3 py-2 text-slate-200/80">{it.user_id ?? ''}</td>
                    <td className="px-3 py-2 text-slate-200/80">{it.ip || ''}</td>
                    <td className="px-3 py-2">
                      <button
                        type="button"
                        className="font-mono text-xs text-[color:var(--accent)] hover:underline"
                        onClick={async () => {
                          await navigator.clipboard.writeText(it.trace_id || '')
                        }}
                        title="点击复制 trace_id"
                      >
                        {it.trace_id || ''}
                      </button>
                    </td>
                  </tr>
                )
              })}
              {items.length === 0 ? (
                <tr>
                  <td className="px-3 py-6 text-center text-slate-200/60" colSpan={8}>
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
