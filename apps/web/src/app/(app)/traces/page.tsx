'use client'

import Link from 'next/link'
import { useEffect, useMemo, useState } from 'react'
import { getToken } from '../../../lib/auth'
import { getTraces } from '../../../lib/api'
import { clampInt, formatRFC3339 } from '../../../lib/format'
import { Badge, Button, Card, Input, Select } from '../../../components/Ui'

export default function TracesPage() {
  const token = getToken() || ''
  const [page, setPage] = useState(1)
  const [pageSize, setPageSize] = useState(50)
  const [traceId, setTraceId] = useState('')

  const [loading, setLoading] = useState(false)
  const [err, setErr] = useState<string | null>(null)
  const [data, setData] = useState<any | null>(null)

  const query = useMemo(() => {
    const q: Record<string, any> = { page, page_size: pageSize }
    if (traceId.trim()) q.trace_id = traceId.trim()
    return q
  }, [page, pageSize, traceId])

  const refresh = async () => {
    setLoading(true)
    setErr(null)
    try {
      const res = await getTraces(token, query)
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
          <div className="text-xs uppercase tracking-[0.22em] text-slate-300/70">Tracing</div>
          <div className="mt-1 text-2xl font-semibold">链路追踪</div>
          <div className="mt-2 text-sm text-slate-200/70">按 trace_id 聚合展示请求/错误/平均耗时；点进去看 spans。</div>
        </div>
        <Button tone="primary" onClick={refresh} disabled={loading} type="button">
          {loading ? '刷新中…' : '刷新'}
        </Button>
      </div>

      <Card title="筛选" subtitle="trace_id 支持精确匹配（等值）。" right={err ? <Badge tone="bad">{err}</Badge> : <Badge tone="neutral">total: {total}</Badge>}>
        <div className="grid grid-cols-1 gap-3 md:grid-cols-3">
          <div className="md:col-span-2">
            <div className="mb-1 text-xs text-slate-200/70">trace_id</div>
            <Input value={traceId} onChange={(e) => setTraceId(e.target.value)} placeholder="例如: 2f7a..." />
          </div>
          <div>
            <div className="mb-1 text-xs text-slate-200/70">page size</div>
            <Select value={String(pageSize)} onChange={(e) => setPageSize(clampInt(Number(e.target.value), 5, 200))}>
              {[20, 50, 100, 200].map((n) => (
                <option key={n} value={String(n)}>
                  {n} / page
                </option>
              ))}
            </Select>
          </div>
        </div>
        <div className="mt-4 flex flex-wrap items-center gap-3">
          <div className="text-xs text-slate-200/70">
            Page {page}/{totalPages}
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
              setPage(1)
            }}
            disabled={loading}
            type="button"
          >
            清空
          </Button>
        </div>
      </Card>

      <Card title="结果" subtitle="点 trace_id 进入详情。">
        <div className="overflow-auto rounded-2xl border border-white/10">
          <table className="w-full min-w-[920px] text-left text-sm">
            <thead className="bg-white/5 text-xs text-slate-200/70">
              <tr>
                <th className="px-3 py-2">latest</th>
                <th className="px-3 py-2">trace_id</th>
                <th className="px-3 py-2">requests</th>
                <th className="px-3 py-2">errors</th>
                <th className="px-3 py-2">avg latency(ms)</th>
              </tr>
            </thead>
            <tbody className="divide-y divide-white/5">
              {items.map((it) => (
                <tr key={String(it.trace_id)} className="hover:bg-white/5">
                  <td className="px-3 py-2 text-slate-200/80">{formatRFC3339(it.latest_at)}</td>
                  <td className="px-3 py-2">
                    <Link className="font-mono text-xs text-[color:var(--accent)] hover:underline" href={`/traces/${encodeURIComponent(it.trace_id)}`}>
                      {it.trace_id}
                    </Link>
                  </td>
                  <td className="px-3 py-2 text-slate-200/80">{it.request_count}</td>
                  <td className="px-3 py-2">{Number(it.error_count) > 0 ? <Badge tone="bad">{it.error_count}</Badge> : <Badge tone="ok">0</Badge>}</td>
                  <td className="px-3 py-2 text-slate-200/80">{Number(it.avg_latency_ms || 0).toFixed(1)}</td>
                </tr>
              ))}
              {items.length === 0 ? (
                <tr>
                  <td className="px-3 py-6 text-center text-slate-200/60" colSpan={5}>
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
