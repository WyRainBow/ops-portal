'use client'

import { useEffect, useMemo, useState } from 'react'
import { useParams } from 'next/navigation'
import { getToken } from '../../../../lib/auth'
import { getTraceDetail } from '../../../../lib/api'
import { formatRFC3339 } from '../../../../lib/format'
import { Badge, Button, Card, TextArea } from '../../../../components/Ui'

export default function TraceDetailPage() {
  const token = getToken() || ''
  const params = useParams<{ traceId: string }>()
  const traceId = useMemo(() => decodeURIComponent(params?.traceId || ''), [params])

  const [loading, setLoading] = useState(false)
  const [err, setErr] = useState<string | null>(null)
  const [data, setData] = useState<any | null>(null)
  const [selected, setSelected] = useState<any | null>(null)

  const refresh = async () => {
    setLoading(true)
    setErr(null)
    try {
      const res = await getTraceDetail(token, traceId)
      setData(res)
      setSelected(null)
    } catch (e: any) {
      setErr(e?.message || String(e))
    } finally {
      setLoading(false)
    }
  }

  useEffect(() => {
    if (!traceId) return
    refresh()
    // eslint-disable-next-line react-hooks/exhaustive-deps
  }, [traceId])

  const spans = (data?.spans || []) as any[]

  return (
    <div className="space-y-6">
      <div className="flex items-end justify-between gap-4">
        <div>
          <div className="text-xs uppercase tracking-[0.22em] text-slate-300/70">Trace</div>
          <div className="mt-1 text-2xl font-semibold">Trace 详情</div>
          <div className="mt-2 font-mono text-xs text-[color:var(--accent)]">{traceId}</div>
        </div>
        <div className="flex items-center gap-2">
          {err ? <Badge tone="bad">{err}</Badge> : null}
          <Button tone="primary" onClick={refresh} disabled={loading} type="button">
            {loading ? '刷新中…' : '刷新'}
          </Button>
          <Button
            tone="ghost"
            type="button"
            onClick={async () => {
              await navigator.clipboard.writeText(traceId)
            }}
          >
            复制 trace_id
          </Button>
        </div>
      </div>

      <div className="grid grid-cols-1 gap-6 lg:grid-cols-3">
        <Card title="Spans" subtitle="点击 span 可在右侧查看 tags。">
          <div className="overflow-auto rounded-2xl border border-white/10">
            <table className="w-full min-w-[980px] text-left text-sm">
              <thead className="bg-white/5 text-xs text-slate-200/70">
                <tr>
                  <th className="px-3 py-2">status</th>
                  <th className="px-3 py-2">span_name</th>
                  <th className="px-3 py-2">duration(ms)</th>
                  <th className="px-3 py-2">start</th>
                  <th className="px-3 py-2">end</th>
                  <th className="px-3 py-2">span_id</th>
                  <th className="px-3 py-2">parent</th>
                </tr>
              </thead>
              <tbody className="divide-y divide-white/5">
                {spans.map((s) => (
                  <tr
                    key={String(s.span_id)}
                    className={['cursor-pointer hover:bg-white/5', selected?.span_id === s.span_id ? 'bg-white/5' : ''].join(' ')}
                    onClick={() => setSelected(s)}
                  >
                    <td className="px-3 py-2">
                      <Badge tone={(String(s.status || '').toLowerCase().includes('error') ? 'bad' : 'ok') as any}>{s.status || 'OK'}</Badge>
                    </td>
                    <td className="px-3 py-2 text-slate-100">{s.span_name}</td>
                    <td className="px-3 py-2 text-slate-200/80">{Number(s.duration_ms || 0).toFixed(2)}</td>
                    <td className="px-3 py-2 text-slate-200/80">{formatRFC3339(s.start_time)}</td>
                    <td className="px-3 py-2 text-slate-200/80">{formatRFC3339(s.end_time)}</td>
                    <td className="px-3 py-2 font-mono text-xs text-slate-200/70">{s.span_id}</td>
                    <td className="px-3 py-2 font-mono text-xs text-slate-200/60">{s.parent_span_id || ''}</td>
                  </tr>
                ))}
                {spans.length === 0 ? (
                  <tr>
                    <td className="px-3 py-6 text-center text-slate-200/60" colSpan={7}>
                      无数据
                    </td>
                  </tr>
                ) : null}
              </tbody>
            </table>
          </div>
        </Card>

        <div className="lg:col-span-2 space-y-6">
          <Card title="Tags" subtitle={selected?.span_id ? `span_id: ${selected.span_id}` : '选择一个 span 查看 tags'}>
            <TextArea
              value={selected?.tags ? JSON.stringify(selected.tags, null, 2) : ''}
              readOnly
              rows={18}
              className="font-mono text-xs leading-relaxed"
              placeholder="点击左侧一条 span…"
            />
          </Card>
        </div>
      </div>
    </div>
  )
}
