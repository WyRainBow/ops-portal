'use client'

import { useEffect, useMemo, useState } from 'react'
import { getToken } from '../../../lib/auth'
import { getRuntimeLogs, getRuntimeStatus, lokiQueryRange } from '../../../lib/api'
import { clampInt } from '../../../lib/format'
import { Badge, Button, Card, Input, Select, TextArea } from '../../../components/Ui'

export default function RuntimePage() {
  const token = getToken() || ''
  const [service, setService] = useState('resume-backend')

  const [statusLoading, setStatusLoading] = useState(false)
  const [statusErr, setStatusErr] = useState<string | null>(null)
  const [status, setStatus] = useState<any | null>(null)

  const [logStream, setLogStream] = useState<'error' | 'out'>('error')
  const [logLines, setLogLines] = useState(200)
  const [logsLoading, setLogsLoading] = useState(false)
  const [logsErr, setLogsErr] = useState<string | null>(null)
  const [logs, setLogs] = useState<any | null>(null)

  const [logql, setLogql] = useState('{job="resume-backend", stream="error"}')
  const [lokiLoading, setLokiLoading] = useState(false)
  const [lokiErr, setLokiErr] = useState<string | null>(null)
  const [lokiText, setLokiText] = useState('')

  const refreshStatus = async () => {
    setStatusLoading(true)
    setStatusErr(null)
    try {
      const s = await getRuntimeStatus(token, { service })
      setStatus(s)
    } catch (e: any) {
      setStatusErr(e?.message || String(e))
    } finally {
      setStatusLoading(false)
    }
  }

  const refreshLogs = async () => {
    setLogsLoading(true)
    setLogsErr(null)
    try {
      const r = await getRuntimeLogs(token, { service, stream: logStream, lines: logLines })
      setLogs(r)
    } catch (e: any) {
      setLogsErr(e?.message || String(e))
    } finally {
      setLogsLoading(false)
    }
  }

  const refreshLoki = async () => {
    setLokiLoading(true)
    setLokiErr(null)
    try {
      const end = Date.now()
      const start = end - 60 * 60 * 1000
      const r = await lokiQueryRange(token, { query: logql, start_ms: start, end_ms: end, limit: 200 })
      // r.lines: [{ts,line,labels}]
      const lines = (r?.lines || []) as any[]
      const text = lines
        .map((x) => {
          const ts = x.ts ? new Date(x.ts).toISOString() : ''
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
    refreshStatus()
    refreshLogs()
    // eslint-disable-next-line react-hooks/exhaustive-deps
  }, [service])

  useEffect(() => {
    refreshLogs()
    // eslint-disable-next-line react-hooks/exhaustive-deps
  }, [logStream, logLines])

  const svc = status?.service || {}
  const pm2ok = Boolean(status?.pm2?.ok)
  const dbok = Boolean(status?.database?.ok)

  const topBadges = useMemo(() => {
    const b: { label: string; tone: any }[] = []
    b.push({ label: pm2ok ? 'PM2 OK' : 'PM2 ?' , tone: pm2ok ? 'ok' : 'warn' })
    b.push({ label: dbok ? 'DB OK' : 'DB FAIL', tone: dbok ? 'ok' : 'bad' })
    const st = String(svc?.status || '')
    b.push({ label: st || 'service ?', tone: st === 'online' ? 'ok' : st ? 'warn' : 'neutral' })
    return b
  }, [pm2ok, dbok, svc?.status])

  return (
    <div className="space-y-6">
      <div className="flex items-end justify-between gap-4">
        <div>
          <div className="text-xs uppercase tracking-[0.22em] text-slate-300/70">Runtime</div>
          <div className="mt-1 text-2xl font-semibold">运行状态（同机）</div>
          <div className="mt-2 text-sm text-slate-200/70">读取 PM2 / Git / 系统资源 / 日志文件。额外提供 Loki 快捷查询（最近 1h）。</div>
        </div>
        <div className="flex items-center gap-2">
          {topBadges.map((x) => (
            <Badge key={x.label} tone={x.tone}>
              {x.label}
            </Badge>
          ))}
          <Button tone="primary" onClick={() => { refreshStatus(); refreshLogs(); }} disabled={statusLoading || logsLoading} type="button">
            {(statusLoading || logsLoading) ? '刷新中…' : '刷新'}
          </Button>
        </div>
      </div>

      <Card
        title="目标服务"
        subtitle="默认 resume-backend（与你的 PM2 名称一致）。"
        right={statusErr ? <Badge tone="bad">{statusErr}</Badge> : null}
      >
        <div className="flex flex-wrap items-end gap-3">
          <div className="w-[260px]">
            <div className="mb-1 text-xs text-slate-200/70">service</div>
            <Input value={service} onChange={(e) => setService(e.target.value)} placeholder="resume-backend" />
          </div>
          <Button tone="ghost" onClick={refreshStatus} disabled={statusLoading} type="button">
            {statusLoading ? '加载中…' : '刷新状态'}
          </Button>
        </div>
      </Card>

      <div className="grid grid-cols-1 gap-6 lg:grid-cols-2">
        <Card title="状态快照" subtitle="结构化 JSON（v1 先展示，后续可做成更细的 KPI 面板）。">
          <pre className="max-h-[520px] overflow-auto rounded-2xl border border-white/10 bg-black/30 p-4 text-xs leading-relaxed text-slate-100">
            {JSON.stringify(status || {}, null, 2)}
          </pre>
        </Card>

        <Card
          title="PM2 日志 Tail"
          subtitle="读 /root/.pm2/logs/* 文件（后端按 service+stream 映射）。"
          right={logsErr ? <Badge tone="bad">{logsErr}</Badge> : null}
        >
          <div className="grid grid-cols-1 gap-3 md:grid-cols-3">
            <div>
              <div className="mb-1 text-xs text-slate-200/70">stream</div>
              <Select value={logStream} onChange={(e) => setLogStream(e.target.value as any)}>
                <option value="error">error</option>
                <option value="out">out</option>
              </Select>
            </div>
            <div>
              <div className="mb-1 text-xs text-slate-200/70">lines</div>
              <Select value={String(logLines)} onChange={(e) => setLogLines(clampInt(Number(e.target.value), 10, 2000))}>
                {[50, 100, 200, 500, 1000, 2000].map((n) => (
                  <option key={n} value={String(n)}>
                    {n}
                  </option>
                ))}
              </Select>
            </div>
            <div className="flex items-end">
              <Button tone="ghost" onClick={refreshLogs} disabled={logsLoading} type="button" className="w-full">
                {logsLoading ? '加载中…' : '刷新日志'}
              </Button>
            </div>
          </div>

          <div className="mt-4">
            <div className="mb-1 text-xs text-slate-200/70">path: {logs?.path || ''}</div>
            <TextArea value={logs?.content || ''} readOnly rows={18} className="font-mono text-xs leading-relaxed" />
          </div>
        </Card>
      </div>

      <Card
        title="Loki 快捷查询（最近 1 小时）"
        subtitle="如果你已经接入 promtail -> loki，这里能直接查到同一台机器上的结构化日志。"
        right={lokiErr ? <Badge tone="bad">{lokiErr}</Badge> : null}
      >
        <div className="grid grid-cols-1 gap-3 md:grid-cols-3">
          <div className="md:col-span-2">
            <div className="mb-1 text-xs text-slate-200/70">LogQL</div>
            <Input value={logql} onChange={(e) => setLogql(e.target.value)} />
          </div>
          <div className="flex items-end">
            <Button tone="primary" onClick={refreshLoki} disabled={lokiLoading} type="button" className="w-full">
              {lokiLoading ? '查询中…' : '查询'}
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
