'use client'

import { useEffect, useMemo, useState } from 'react'
import { getToken } from '../../../../lib/auth'
import { getErrorLogs } from '../../../../lib/api'
import { clampInt, formatRFC3339 } from '../../../../lib/format'
import { Badge, Button, Card, Input, Select, TextArea } from '../../../../components/Ui'

export default function ErrorLogsPage() {
  const token = getToken() || ''
  const [page, setPage] = useState(1)
  const [pageSize, setPageSize] = useState(50)
  const [traceId, setTraceId] = useState('')
  const [keyword, setKeyword] = useState('')

  const [loading, setLoading] = useState(false)
  const [err, setErr] = useState<string | null>(null)
  const [data, setData] = useState<any | null>(null)
  const [selected, setSelected] = useState<any | null>(null)

  const query = useMemo(() => {
    const q: Record<string, any> = { page, page_size: pageSize }
    if (traceId.trim()) q.trace_id = traceId.trim()
    if (keyword.trim()) q.keyword = keyword.trim()
    return q
  }, [page, pageSize, traceId, keyword])

  const refresh = async () => {
    setLoading(true)
    setErr(null)
    try {
      const res = await getErrorLogs(token, query)
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
          <div className="mt-1 text-2xl font-semibold">错误日志</div>
          <div className="mt-2 text-sm text-slate-200/70">来自 PostgreSQL 的错误记录（error_message 可关键词搜索）。</div>
        </div>
        <Button tone="primary" onClick={refresh} disabled={loading} type="button">
          {loading ? '刷新中…' : '刷新'}
        </Button>
      </div>

      <div className="grid grid-cols-1 gap-6 lg:grid-cols-3">
        <div className="lg:col-span-2 space-y-6">
          <Card
            title="筛选"
            subtitle="建议：先按 keyword 定位问题类型，再用 trace_id 去链路追踪查看上下文。"
            right={err ? <Badge tone="bad">{err}</Badge> : <Badge tone="neutral">total: {total}</Badge>}
          >
            <div className="grid grid-cols-1 gap-3 md:grid-cols-2">
              <div>
                <div className="mb-1 text-xs text-slate-200/70">trace_id</div>
                <Input value={traceId} onChange={(e) => setTraceId(e.target.value)} placeholder="例如: 2f7a..." />
              </div>
              <div>
                <div className="mb-1 text-xs text-slate-200/70">keyword</div>
                <Input value={keyword} onChange={(e) => setKeyword(e.target.value)} placeholder="exception / timeout / auth..." />
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
                  setKeyword('')
                  setPage(1)
                  setSelected(null)
                }}
                disabled={loading}
                type="button"
              >
                清空
              </Button>
            </div>
          </Card>

          <Card title="结果" subtitle="点击一条记录，可在右侧查看完整 error_message。">
            <div className="overflow-auto rounded-2xl border border-white/10">
              <table className="w-full min-w-[940px] text-left text-sm">
                <thead className="bg-white/5 text-xs text-slate-200/70">
                  <tr>
                    <th className="px-3 py-2">时间</th>
                    <th className="px-3 py-2">service</th>
                    <th className="px-3 py-2">type</th>
                    <th className="px-3 py-2">trace_id</th>
                    <th className="px-3 py-2">message</th>
                  </tr>
                </thead>
                <tbody className="divide-y divide-white/5">
                  {items.map((it) => (
                    <tr
                      key={String(it.id)}
                      className={['cursor-pointer hover:bg-white/5', selected?.id === it.id ? 'bg-white/5' : ''].join(' ')}
                      onClick={() => setSelected(it)}
                    >
                      <td className="px-3 py-2 text-slate-200/80">{formatRFC3339(it.created_at)}</td>
                      <td className="px-3 py-2 text-slate-200/80">{it.service || ''}</td>
                      <td className="px-3 py-2">
                        <Badge tone="bad">{it.error_type || 'error'}</Badge>
                      </td>
                      <td className="px-3 py-2 font-mono text-xs text-[color:var(--accent)]">{it.trace_id || ''}</td>
                      <td className="px-3 py-2 text-slate-100">
                        {(it.error_message || '').slice(0, 120)}
                        {(it.error_message || '').length > 120 ? '…' : ''}
                      </td>
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

        <Card
          title="详情"
          subtitle={selected?.trace_id ? `trace_id: ${selected.trace_id}` : '选择一条错误记录查看完整信息'}
          right={
            selected?.trace_id ? (
              <Button
                tone="ghost"
                type="button"
                onClick={async () => {
                  await navigator.clipboard.writeText(selected.trace_id || '')
                }}
              >
                复制 trace_id
              </Button>
            ) : null
          }
        >
          <div className="space-y-3">
            <div className="text-xs text-slate-200/70">error_message</div>
            <TextArea
              value={selected?.error_message || ''}
              readOnly
              rows={18}
              className="font-mono text-xs leading-relaxed"
              placeholder="点击左侧表格中的一条记录…"
            />
          </div>
        </Card>
      </div>
    </div>
  )
}
