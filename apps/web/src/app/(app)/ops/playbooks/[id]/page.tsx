'use client'

import { useCallback, useEffect, useMemo, useState } from 'react'
import { useParams } from 'next/navigation'
import { getToken } from '../../../../../lib/auth'
import { getPlaybook, executePlaybook } from '../../../../../lib/api'
import { Badge, Button, Card, Input, TextArea } from '../../../../../components/Ui'

interface Playbook {
  id: string
  name: string
  description: string
  category: string
  severity: 'low' | 'medium' | 'high' | 'critical'
  command: string
  timeout: number
  require_confirm: boolean
  enabled: boolean
  parameters: Array<{
    name: string
    type: string
    required: boolean
    description: string
    default?: any
  }>
}

export default function PlaybookDetailPage() {
  const params = useParams()
  const id = params.id as string
  const token = getToken() || ''

  const [playbook, setPlaybook] = useState<Playbook | null>(null)
  const [loading, setLoading] = useState(false)
  const [err, setErr] = useState<string | null>(null)
  const [executing, setExecuting] = useState(false)
  const [executionResult, setExecutionResult] = useState<any>(null)

  // Execute form state
  const [parameters, setParameters] = useState<Record<string, any>>({})
  const [reason, setReason] = useState('')
  const [dryRun, setDryRun] = useState(true)

  const load = useCallback(async () => {
    if (!token || !id) return
    setLoading(true)
    setErr(null)
    try {
      const data = await getPlaybook(token, id)
      setPlaybook(data.playbook)

      // Initialize parameters with defaults
      if (data.playbook?.parameters) {
        const defaults: Record<string, any> = {}
        for (const param of data.playbook.parameters) {
          if (param.default !== undefined) {
            defaults[param.name] = param.default
          }
        }
        setParameters(defaults)
      }
    } catch (e: any) {
      setErr(e?.message || '加载失败')
    } finally {
      setLoading(false)
    }
  }, [token, id])

  useEffect(() => {
    void load()
  }, [load])

  const handleExecute = useCallback(async () => {
    if (!playbook || !token) return

    // Validate required parameters
    for (const param of playbook.parameters) {
      if (param.required && !parameters[param.name]) {
        alert(`请填写必填参数: ${param.name}`)
        return
      }
    }

    if (!reason.trim()) {
      alert('请填写执行原因（用于审计）')
      return
    }

    setExecuting(true)
    setErr(null)
    try {
      const result = await executePlaybook(token, playbook.id, {
        playbook_id: playbook.id,
        parameters,
        reason,
        dry_run: dryRun,
      })
      setExecutionResult(result.execution)
      // Clear form after success
      setReason('')
      setExecutionResult(result.execution)
    } catch (e: any) {
      setErr(e?.message || '执行失败')
    } finally {
      setExecuting(false)
    }
  }, [playbook, token, parameters, reason, dryRun])

  const severityTone = (severity: string): 'ok' | 'warn' | 'bad' | 'neutral' => {
    switch (severity) {
      case 'low': return 'ok'
      case 'medium': return 'warn'
      case 'high': return 'bad'
      case 'critical': return 'bad'
      default: return 'neutral'
    }
  }

  if (loading) {
    return (
      <div className="space-y-6">
        <div>
          <div className="text-xs uppercase tracking-[0.22em] text-slate-300/70">Ops</div>
          <div className="mt-1 text-2xl font-semibold">Playbook 详情</div>
        </div>
        <Card right={<Badge tone="warn">loading</Badge>}>
          <div className="py-8 text-center text-slate-200/60">加载中...</div>
        </Card>
      </div>
    )
  }

  if (err && !playbook) {
    return (
      <div className="space-y-6">
        <div>
          <div className="text-xs uppercase tracking-[0.22em] text-slate-300/70">Ops</div>
          <div className="mt-1 text-2xl font-semibold">Playbook 详情</div>
        </div>
        <Card right={<Badge tone="bad">error</Badge>}>
          <div className="py-8 text-center text-rose-400">{err}</div>
        </Card>
      </div>
    )
  }

  if (!playbook) {
    return null
  }

  return (
    <div className="space-y-6">
      <div>
        <div className="text-xs uppercase tracking-[0.22em] text-slate-300/70">Ops</div>
        <div className="mt-1 text-2xl font-semibold">{playbook.name}</div>
        <div className="mt-2 text-sm text-slate-200/70">{playbook.description}</div>
      </div>

      {/* Playbook Info */}
      <Card title="基本信息">
        <div className="grid gap-4 sm:grid-cols-2">
          <div>
            <div className="text-xs text-slate-200/60">ID</div>
            <div className="font-mono text-sm">{playbook.id}</div>
          </div>
          <div>
            <div className="text-xs text-slate-200/60">分类</div>
            <div className="text-sm">{playbook.category}</div>
          </div>
          <div>
            <div className="text-xs text-slate-200/60">风险等级</div>
            <div><Badge tone={severityTone(playbook.severity)}>{playbook.severity}</Badge></div>
          </div>
          <div>
            <div className="text-xs text-slate-200/60">超时时间</div>
            <div className="text-sm">{playbook.timeout / 1000}s</div>
          </div>
          <div>
            <div className="text-xs text-slate-200/60">需要确认</div>
            <div className="text-sm">{playbook.require_confirm ? '是' : '否'}</div>
          </div>
          <div>
            <div className="text-xs text-slate-200/60">状态</div>
            <div><Badge tone={playbook.enabled ? 'ok' : 'neutral'}>{playbook.enabled ? '已启用' : '已禁用'}</Badge></div>
          </div>
        </div>
      </Card>

      {/* Command */}
      <Card title="执行命令">
        <pre className="overflow-x-auto rounded-xl bg-black/30 p-3 font-mono text-sm text-slate-300">
          {playbook.command}
        </pre>
      </Card>

      {/* Parameters */}
      {playbook.parameters.length > 0 && (
        <Card title={`参数 (${playbook.parameters.length})`}>
          <div className="overflow-hidden rounded-2xl border border-white/10">
            <table className="w-full text-left text-sm">
              <thead className="bg-white/5 text-xs font-semibold uppercase tracking-wide text-slate-200/70">
                <tr>
                  <th className="px-4 py-3">名称</th>
                  <th className="px-4 py-3">类型</th>
                  <th className="px-4 py-3">必填</th>
                  <th className="px-4 py-3">默认值</th>
                  <th className="px-4 py-3">描述</th>
                </tr>
              </thead>
              <tbody>
                {playbook.parameters.map((param) => (
                  <tr key={param.name} className="border-t border-white/5">
                    <td className="px-4 py-3 font-mono">{param.name}</td>
                    <td className="px-4 py-3">{param.type}</td>
                    <td className="px-4 py-3">
                      {param.required ? <Badge tone="bad">是</Badge> : <Badge tone="neutral">否</Badge>}
                    </td>
                    <td className="px-4 py-3 font-mono text-slate-200/60">{param.default?.toString() || '-'}</td>
                    <td className="px-4 py-3 text-slate-200/60">{param.description || '-'}</td>
                  </tr>
                ))}
              </tbody>
            </table>
          </div>
        </Card>
      )}

      {/* Execution Form */}
      <Card title="执行操作">
        <div className="space-y-4">
          {playbook.parameters.map((param) => (
            <div key={param.name}>
              <div className="mb-1 text-xs text-slate-200/60">
                {param.name}
                {param.required && <span className="text-rose-400"> *</span>}
                {param.description && ` - ${param.description}`}
              </div>
              {param.type === 'bool' ? (
                <select
                  className="ops-input w-full rounded-2xl px-3 py-2 text-sm"
                  value={parameters[param.name]?.toString() || 'false'}
                  onChange={(e) => setParameters({ ...parameters, [param.name]: e.target.value === 'true' })}
                >
                  <option value="false">false</option>
                  <option value="true">true</option>
                </select>
              ) : param.type === 'int' ? (
                <Input
                  type="number"
                  value={parameters[param.name]?.toString() || ''}
                  onChange={(e) => setParameters({ ...parameters, [param.name]: parseInt(e.target.value) || 0 })}
                  placeholder={param.default?.toString() || `请输入 ${param.name}`}
                />
              ) : (
                <Input
                  type="text"
                  value={parameters[param.name]?.toString() || ''}
                  onChange={(e) => setParameters({ ...parameters, [param.name]: e.target.value })}
                  placeholder={param.default?.toString() || `请输入 ${param.name}`}
                />
              )}
            </div>
          ))}

          <div>
            <div className="mb-1 text-xs text-slate-200/60">执行原因 *</div>
            <TextArea
              value={reason}
              onChange={(e) => setReason(e.target.value)}
              placeholder="请输入执行原因（用于审计）"
              rows={2}
            />
          </div>

          <div className="flex items-center gap-2">
            <input
              type="checkbox"
              id="dryRun"
              checked={dryRun}
              onChange={(e) => setDryRun(e.target.checked)}
              className="h-4 w-4 rounded"
            />
            <label htmlFor="dryRun" className="text-sm text-slate-200/80">
              预演模式（Dry Run）- 只显示将要执行的命令，不实际执行
            </label>
          </div>

          <Button
            tone={dryRun ? 'ghost' : 'primary'}
            onClick={handleExecute}
            disabled={executing}
            className="w-full"
          >
            {executing ? '执行中...' : dryRun ? '预演执行' : '执行'}
          </Button>

          {(playbook.severity === 'high' || playbook.severity === 'critical') && (
            <div className="rounded-xl bg-rose-500/10 p-3 text-xs text-rose-400">
              ⚠️ 警告：此操作风险等级为 {playbook.severity}，请谨慎操作
            </div>
          )}
        </div>
      </Card>

      {/* Execution Result */}
      {executionResult && (
        <Card title="执行结果" subtitle={executionResult.execution_id} right={<Badge tone={executionResult.status === 'success' ? 'ok' : 'bad'}>{executionResult.status}</Badge>}>
          <div className="space-y-3 text-sm">
            <div className="grid grid-cols-2 gap-4">
              <div>
                <div className="text-xs text-slate-200/60">Playbook ID</div>
                <div className="font-mono">{executionResult.playbook_id}</div>
              </div>
              <div>
                <div className="text-xs text-slate-200/60">耗时</div>
                <div className="font-mono">{executionResult.duration ? `${Math.round(executionResult.duration / 1000000)}ms` : '-'}</div>
              </div>
            </div>
            {executionResult.output && (
              <div>
                <div className="text-xs text-slate-200/60">输出</div>
                <pre className="mt-1 overflow-x-auto rounded-xl bg-black/30 p-3 font-mono text-xs text-slate-300">
                  {executionResult.output}
                </pre>
              </div>
            )}
            {executionResult.error && (
              <div>
                <div className="text-xs text-slate-200/60">错误</div>
                <div className="mt-1 rounded-xl bg-rose-500/10 p-3 font-mono text-xs text-rose-400">
                  {executionResult.error}
                </div>
              </div>
            )}
          </div>
        </Card>
      )}
    </div>
  )
}
