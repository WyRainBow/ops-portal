'use client'

import { useCallback, useEffect, useMemo, useState } from 'react'
import { getToken } from '../../../lib/auth'
import { lokiQueryRange } from '../../../lib/api'
import { Badge, Button, Card } from '../../../components/Ui'

interface Alert {
  status: string
  labels: Record<string, string>
  annotations: Record<string, string>
  startsAt: string
  endsAt?: string
  generatorURL: string
  fingerprint: string
}

interface PrometheusAlertResponse {
  data: {
    result: Array<{
      metric: Record<string, string>
      value: [number, string]
    }>
  }
}

export default function AlertsPage() {
  const token = getToken() || ''

  const [alerts, setAlerts] = useState<Alert[]>([])
  const [loading, setLoading] = useState(false)
  const [err, setErr] = useState<string | null>(null)
  const [showFiringOnly, setShowFiringOnly] = useState(true)

  const loadAlerts = useCallback(async () => {
    if (!token) return
    setLoading(true)
    setErr(null)

    try {
      // Query Prometheus for alerts
      const endTime = Date.now()
      const startTime = endTime - 5 * 60 * 1000 // Last 5 minutes

      const result = await lokiQueryRange(token, {
        query: 'ALERTS{}',
        start_ms: startTime,
        end_ms: endTime,
        limit: 100,
      }) as PrometheusAlertResponse

      // Parse alerts from Prometheus response
      const parsedAlerts: Alert[] = []
      for (const item of result.data?.result || []) {
        const labels = item.metric || {}
        const value = item.value?.[1] || '{}'

        try {
          const alertData = JSON.parse(value)
          if (Array.isArray(alertData)) {
            parsedAlerts.push(...alertData)
          } else {
            parsedAlerts.push(alertData)
          }
        } catch {
          // Skip invalid alerts
        }
      }

      setAlerts(parsedAlerts)
    } catch (e: any) {
      setErr(e?.message || '加载失败')
    } finally {
      setLoading(false)
    }
  }, [token])

  useEffect(() => {
    void loadAlerts()
    // Auto-refresh every 30 seconds
    const interval = setInterval(() => void loadAlerts(), 30000)
    return () => clearInterval(interval)
  }, [loadAlerts])

  const severityTone = (severity: string): 'ok' | 'warn' | 'bad' | 'neutral' => {
    switch (severity?.toLowerCase()) {
      case 'critical': return 'bad'
      case 'warning': return 'warn'
      case 'info': return 'ok'
      default: return 'neutral'
    }
  }

  const statusTone = (status: string): 'ok' | 'warn' | 'bad' | 'neutral' => {
    switch (status?.toLowerCase()) {
      case 'firing': return 'bad'
      case 'resolved': return 'ok'
      default: return 'neutral'
    }
  }

  const filteredAlerts = showFiringOnly
    ? alerts.filter(a => a.status === 'firing')
    : alerts

  const groupedAlerts = useMemo(() => {
    const groups: Record<string, Alert[]> = {}
    for (const alert of filteredAlerts) {
      const key = alert.labels?.alertname || 'unknown'
      if (!groups[key]) groups[key] = []
      groups[key].push(alert)
    }
    return groups
  }, [filteredAlerts])

  const header = useMemo(() => {
    const firingCount = alerts.filter(a => a.status === 'firing').length
    return (
      <Card title="告警列表" subtitle="来自 Prometheus Alertmanager" right={err ? <Badge tone="bad">{err}</Badge> : <Badge tone="neutral">{firingCount} firing</Badge>}>
        <div className="flex items-center justify-between">
          <div className="flex items-center gap-4">
            <label className="flex items-center gap-2 text-sm">
              <input
                type="checkbox"
                checked={showFiringOnly}
                onChange={(e) => setShowFiringOnly(e.target.checked)}
                className="h-4 w-4 rounded"
              />
              仅显示活跃告警
            </label>
          </div>
          <Button tone="ghost" onClick={() => void loadAlerts()}>
            刷新
          </Button>
        </div>
      </Card>
    )
  }, [alerts.length, err, showFiringOnly, loadAlerts])

  return (
    <div className="space-y-6">
      <div>
        <div className="text-xs uppercase tracking-[0.22em] text-slate-300/70">Observability</div>
        <div className="mt-1 text-2xl font-semibold">告警中心</div>
        <div className="mt-2 text-sm text-slate-200/70">实时监控告警状态，支持一键触发 AI 诊断</div>
      </div>
      {header}

      {/* Alerts by Group */}
      <Card title="告警分组" right={loading ? <Badge tone="warn">loading</Badge> : null}>
        <div className="space-y-4">
          {Object.entries(groupedAlerts).map(([alertName, groupAlerts]) => (
            <div key={alertName} className="overflow-hidden rounded-2xl border border-white/10 bg-black/20">
              <div className="flex items-center justify-between bg-white/5 px-4 py-3">
                <div className="flex items-center gap-3">
                  <h3 className="font-semibold">{alertName}</h3>
                  <Badge tone="neutral">{groupAlerts.length} 条</Badge>
                </div>
                <div className="flex items-center gap-2">
                  {groupAlerts.some(a => a.status === 'firing') && (
                    <Badge tone="bad">firing</Badge>
                  )}
                </div>
              </div>
              <div className="divide-y divide-white/5">
                {groupAlerts.map((alert, idx) => (
                  <div key={`${alert.fingerprint}-${idx}`} className="px-4 py-3">
                    <div className="flex items-start justify-between">
                      <div className="min-w-0 flex-1">
                        <div className="mb-2 flex flex-wrap items-center gap-2">
                          <Badge tone={statusTone(alert.status)}>{alert.status}</Badge>
                          {alert.labels?.severity && (
                            <Badge tone={severityTone(alert.labels.severity)}>{alert.labels.severity}</Badge>
                          )}
                          {alert.labels.instance && (
                            <span className="text-xs font-mono text-slate-200/60">{alert.labels.instance}</span>
                          )}
                        </div>
                        {alert.annotations?.summary && (
                          <p className="text-sm text-slate-200/80">{alert.annotations.summary}</p>
                        )}
                        {alert.annotations?.description && (
                          <p className="mt-1 text-xs text-slate-200/60">{alert.annotations.description}</p>
                        )}
                        <div className="mt-2 text-xs text-slate-200/40">
                          开始时间: {new Date(alert.startsAt).toLocaleString('zh-CN')}
                        </div>
                      </div>
                      <Button
                        tone="ghost"
                        className="ml-4 text-xs"
                        onClick={() => {
                          // Navigate to AI Ops page with alert context
                          window.location.href = `/assistant/aiops?alert=${encodeURIComponent(alert.labels.alertname || '')}`
                        }}
                      >
                        AI 诊断
                      </Button>
                    </div>
                  </div>
                ))}
              </div>
            </div>
          ))}
        </div>

        {Object.keys(groupedAlerts).length === 0 && !loading && (
          <div className="py-12 text-center text-slate-200/50">
            {showFiringOnly ? '暂无活跃告警' : '暂无告警'}
          </div>
        )}
      </Card>
    </div>
  )
}
