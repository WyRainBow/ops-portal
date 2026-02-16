'use client'

import { useEffect, useMemo, useState } from 'react'
import { getToken } from '../../../lib/auth'
import { getObsHealth, lokiQueryRange } from '../../../lib/api'
import { Badge, Button, Card, Input, TextArea } from '../../../components/Ui'
import { formatLokiNsToISO } from '../../../lib/format'

export default function ObservabilityPage() {
  const token = getToken() || ''

  const [loading, setLoading] = useState(false)
  const [err, setErr] = useState<string | null>(null)
  const [health, setHealth] = useState<any | null>(null)

  const [logql, setLogql] = useState('{job="resume-backend"} |~ "(?i)error|traceback|exception"')
  const [lokiLoading, setLokiLoading] = useState(false)
  const [lokiErr, setLokiErr] = useState<string | null>(null)
  const [lokiText, setLokiText] = useState('')

  const refresh = async () => {
    setLoading(true)
    setErr(null)
    try {
      const h = await getObsHealth(token)
      setHealth(h)
    } catch (e: any) {
      setErr(e?.message || String(e))
    } finally {
      setLoading(false)
    }
  }

  const queryLoki = async () => {
    setLokiLoading(true)
    setLokiErr(null)
    try {
      const end = Date.now()
      const start = end - 60 * 60 * 1000
      const r = await lokiQueryRange(token, { query: logql, start_ms: start, end_ms: end, limit: 200 })
      const lines = (r?.lines || []) as any[]
      const text = lines
        .map((x) => {
          const ts = formatLokiNsToISO(x.ts)
          return `${ts} ${x.line || ''}`.trimEnd()
        })
        .join('\n')
      setLokiText(text)
    } catch (e: any) {
      setLokiErr(e?.message || String(e))
    } finally {
      setLokiLoading(false)
    }
  }

  useEffect(() => {
    refresh()
    // eslint-disable-next-line react-hooks/exhaustive-deps
  }, [])

  const items = useMemo(() => {
    const c = health?.components || {}
    const out: { name: string; ok: boolean; latency_ms?: number; status_code?: number; detail?: any }[] = []
    ;['grafana', 'loki', 'prometheus', 'node_exporter'].forEach((k) => {
      if (!c[k]) return
      out.push({ name: k, ...c[k] })
    })
    return out
  }, [health])

  const tunnel = health?.tunnels || {}

  return (
    <div className="space-y-6">
      <div className="flex items-end justify-between gap-4">
        <div>
          <div className="text-xs uppercase tracking-[0.22em] text-slate-300/70">Observability</div>
          <div className="mt-1 text-2xl font-semibold">可观测健康检查</div>
          <div className="mt-2 text-sm text-slate-200/70">探测本机 Grafana/Loki/Prometheus/node-exporter 是否 ready，并提供 SSH Tunnel 命令。</div>
        </div>
        <div className="flex items-center gap-2">
          {err ? <Badge tone="bad">{err}</Badge> : null}
          <Button tone="primary" onClick={refresh} disabled={loading} type="button">
            {loading ? '探测中…' : '刷新'}
          </Button>
        </div>
      </div>

      <Card title="Health" subtitle="全部绿色代表可观测组件可用（端口默认都只监听 127.0.0.1）。">
        <div className="grid grid-cols-1 gap-3 md:grid-cols-2 lg:grid-cols-4">
          {items.map((x) => (
            <div key={x.name} className="rounded-2xl border border-white/10 bg-black/20 p-4">
              <div className="text-xs uppercase tracking-[0.22em] text-slate-200/60">{x.name}</div>
              <div className="mt-2 flex items-center gap-2">
                <Badge tone={x.ok ? 'ok' : 'bad'}>{x.ok ? 'OK' : 'FAIL'}</Badge>
                <div className="text-xs text-slate-200/70">{x.latency_ms != null ? `${x.latency_ms}ms` : ''}</div>
              </div>
              <div className="mt-2 text-xs text-slate-200/60">status: {x.status_code ?? ''}</div>
            </div>
          ))}
        </div>
      </Card>

      <div className="grid grid-cols-1 gap-6 lg:grid-cols-2">
        <Card title="SSH Tunnel（Portal）" subtitle="把 ops-portal 暴露到本机 18080（仅用于你本机访问）。">
          <TextArea value={tunnel?.portal || ''} readOnly rows={4} className="font-mono text-xs leading-relaxed" />
        </Card>
        <Card title="SSH Tunnel（Grafana）" subtitle="把 Grafana 暴露到本机 3000（Explore/Dashboard）。">
          <TextArea value={tunnel?.grafana || ''} readOnly rows={4} className="font-mono text-xs leading-relaxed" />
        </Card>
      </div>

      <Card
        title="Loki 快捷查询"
        subtitle={'默认给了一个“抓错误”的 LogQL，你可以直接改成 {job="resume-backend", stream="error"}。'}
        right={lokiErr ? <Badge tone="bad">{lokiErr}</Badge> : null}
      >
        <div className="grid grid-cols-1 gap-3 md:grid-cols-3">
          <div className="md:col-span-2">
            <div className="mb-1 text-xs text-slate-200/70">LogQL</div>
            <Input value={logql} onChange={(e) => setLogql(e.target.value)} />
          </div>
          <div className="flex items-end">
            <Button tone="primary" onClick={queryLoki} disabled={lokiLoading} type="button" className="w-full">
              {lokiLoading ? '查询中…' : '查询最近 1h'}
            </Button>
          </div>
        </div>
        <div className="mt-4">
          <TextArea value={lokiText} readOnly rows={14} className="font-mono text-xs leading-relaxed" placeholder="点击查询…" />
        </div>
      </Card>
    </div>
  )
}
