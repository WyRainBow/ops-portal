'use client'

import { useEffect, useState } from 'react'
import { getToken } from '../../../lib/auth'
import { getOverview, getApiRoutes } from '../../../lib/api'
import { Badge, Card } from '../../../components/Ui'

export default function OverviewPage() {
  const [data, setData] = useState<any>(null)
  const [interfaceTotal, setInterfaceTotal] = useState<number | null>(null)
  const [err, setErr] = useState<string | null>(null)

  useEffect(() => {
    const run = async () => {
      const token = getToken() || ''
      if (!token) return
      try {
        const d = await getOverview(token)
        setData(d)
      } catch (e: any) {
        setErr(e?.message || '加载失败')
      }
      try {
        const routes = await getApiRoutes(token, {})
        setInterfaceTotal(Number(routes?.total ?? 0))
      } catch {
        setInterfaceTotal(null)
      }
    }
    void run()
  }, [])

  return (
    <div className="space-y-6">
      <div className="flex items-end justify-between gap-4">
        <div>
          <div className="text-xs uppercase tracking-[0.22em] text-slate-300/70">Overview</div>
          <div className="mt-1 text-2xl font-semibold">概览</div>
          <div className="mt-2 text-sm text-slate-200/70">这些统计来自 resume_db（users/members/api_request_logs/api_error_logs）。</div>
        </div>
        {err ? <Badge tone="bad">{err}</Badge> : null}
      </div>

      <div className="grid grid-cols-1 gap-4 md:grid-cols-2 xl:grid-cols-3">
        <Kpi title="接口总数" value={interfaceTotal} />
        <Kpi title="用户总数" value={data?.total_users} />
        <Kpi title="成员总数" value={data?.total_members} />
        <Kpi title="24h 请求数" value={data?.requests_24h} />
        <Kpi title="24h 错误数" value={data?.errors_24h} />
        <Kpi title="24h 错误率" value={data ? `${data.error_rate_24h}%` : '-'} tone={Number(data?.error_rate_24h || 0) >= 1 ? 'warn' : 'ok'} />
        <Kpi title="24h 平均延迟" value={data ? `${data.avg_latency_ms_24h}ms` : '-'} />
      </div>
    </div>
  )
}

function Kpi({ title, value, tone }: { title: string; value: any; tone?: 'ok' | 'warn' | 'bad' | 'neutral' }) {
  return (
    <Card
      title={title}
      right={tone ? <Badge tone={tone}>{tone.toUpperCase()}</Badge> : null}
    >
      <div className="font-mono text-3xl font-semibold tracking-tight">{value ?? '-'}</div>
    </Card>
  )
}
