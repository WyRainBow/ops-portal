'use client'

import { useCallback, useEffect, useState } from 'react'
import { useParams } from 'next/navigation'
import { getToken } from '../../../../../lib/auth'
import { getExecution } from '../../../../../lib/api'
import { Badge, Button, Card } from '../../../../../components/Ui'

interface ExecutionResult {
  execution_id: string
  playbook_id: string
  status: string
  start_time: string
  end_time?: string
  duration?: number
  output?: string
  error?: string
  exit_code?: number
}

export default function ExecutionDetailPage() {
  const params = useParams()
  const id = params.id as string
  const token = getToken() || ''

  const [execution, setExecution] = useState<ExecutionResult | null>(null)
  const [loading, setLoading] = useState(false)
  const [err, setErr] = useState<string | null>(null)

  const load = useCallback(async () => {
    if (!token || !id) return
    setLoading(true)
    setErr(null)
    try {
      const data = await getExecution(token, id)
      setExecution(data.execution)
    } catch (e: any) {
      setErr(e?.message || '加载失败')
    } finally {
      setLoading(false)
    }
  }, [token, id])

  useEffect(() => {
    void load()
    // Auto-refresh for running executions
    const interval = setInterval(() => {
      if (execution?.status === 'running' || execution?.status === 'pending') {
        void load()
      }
    }, 2000)
    return () => clearInterval(interval)
  }, [load, execution?.status])

  const statusTone = (status: string): 'ok' | 'warn' | 'bad' | 'neutral' => {
    switch (status) {
      case 'success': return 'ok'
      case 'failed': return 'bad'
      case 'running': return 'warn'
      case 'pending': return 'neutral'
      default: return 'neutral'
    }
  }

  const formatDate = (dateStr: string) => {
    return new Date(dateStr).toLocaleString('zh-CN')
  }

  const formatDuration = (nanos?: number) => {
    if (!nanos) return '-'
    const ms = nanos / 1000000
    if (ms < 1000) return `${Math.round(ms)}ms`
    return `${(ms / 1000).toFixed(2)}s`
  }

  if (loading) {
    return (
      <div className="space-y-6">
        <div>
          <div className="text-xs uppercase tracking-[0.22em] text-slate-300/70">Ops</div>
          <div className="mt-1 text-2xl font-semibold">执行详情</div>
        </div>
        <Card right={<Badge tone="warn">loading</Badge>}>
          <div className="py-8 text-center text-slate-200/60">加载中...</div>
        </Card>
      </div>
    )
  }

  if (err && !execution) {
    return (
      <div className="space-y-6">
        <div>
          <div className="text-xs uppercase tracking-[0.22em] text-slate-300/70">Ops</div>
          <div className="mt-1 text-2xl font-semibold">执行详情</div>
        </div>
        <Card right={<Badge tone="bad">error</Badge>}>
          <div className="py-8 text-center text-rose-400">{err}</div>
        </Card>
      </div>
    )
  }

  if (!execution) {
    return null
  }

  return (
    <div className="space-y-6">
      <div className="flex items-center justify-between">
        <div>
          <div className="text-xs uppercase tracking-[0.22em] text-slate-300/70">Ops</div>
          <div className="mt-1 text-2xl font-semibold">执行详情</div>
        </div>
        <Button tone="ghost" onClick={() => void load()}>
          刷新
        </Button>
      </div>

      {/* Status Card */}
      <Card title="执行状态" right={<Badge tone={statusTone(execution.status)}>{execution.status}</Badge>}>
        <div className="grid gap-4 sm:grid-cols-2">
          <div>
            <div className="text-xs text-slate-200/60">执行 ID</div>
            <div className="font-mono text-sm">{execution.execution_id}</div>
          </div>
          <div>
            <div className="text-xs text-slate-200/60">Playbook ID</div>
            <div className="font-mono text-sm">{execution.playbook_id}</div>
          </div>
          <div>
            <div className="text-xs text-slate-200/60">开始时间</div>
            <div className="text-sm">{formatDate(execution.start_time)}</div>
          </div>
          <div>
            <div className="text-xs text-slate-200/60">结束时间</div>
            <div className="text-sm">{execution.end_time ? formatDate(execution.end_time) : '-'}</div>
          </div>
          <div>
            <div className="text-xs text-slate-200/60">耗时</div>
            <div className="text-sm">{formatDuration(execution.duration)}</div>
          </div>
          <div>
            <div className="text-xs text-slate-200/60">退出码</div>
            <div className="text-sm">{execution.exit_code ?? '-'}</div>
          </div>
        </div>
      </Card>

      {/* Output */}
      {execution.output && (
        <Card title="输出">
          <pre className="overflow-x-auto rounded-xl bg-black/30 p-4 font-mono text-xs text-slate-300">
            {execution.output}
          </pre>
        </Card>
      )}

      {/* Error */}
      {execution.error && (
        <Card title="错误" right={<Badge tone="bad">error</Badge>}>
          <div className="rounded-xl bg-rose-500/10 p-4 font-mono text-sm text-rose-400">
            {execution.error}
          </div>
        </Card>
      )}

      {/* Empty State */}
      {!execution.output && !execution.error && execution.status === 'success' && (
        <Card title="输出">
          <div className="py-8 text-center text-slate-200/50">
            执行成功，无输出
          </div>
        </Card>
      )}

      {/* Status Progress for running */}
      {(execution.status === 'running' || execution.status === 'pending') && (
        <Card right={<Badge tone="warn">running</Badge>}>
          <div className="flex items-center gap-3">
            <div className="h-2 w-2 animate-pulse rounded-full bg-amber-500" />
            <div className="text-sm text-slate-200/70">执行中，自动刷新...</div>
          </div>
        </Card>
      )}
    </div>
  )
}
