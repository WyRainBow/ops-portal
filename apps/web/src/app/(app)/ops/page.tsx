'use client'

import { useCallback, useEffect, useMemo, useState } from 'react'
import { getToken, parseJwtRole } from '../../../lib/auth'
import { getPlaybooks, executePlaybook } from '../../../lib/api'
import { Badge, Button, Card, Input, TextArea } from '../../../components/Ui'

interface Playbook {
  id: string
  name: string
  description: string
  category: string
  severity: 'low' | 'medium' | 'high' | 'critical'
  command: string
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

export default function OpsPage() {
  const token = getToken() || ''
  const role = parseJwtRole(token)
  const isAdmin = role === 'admin'

  const [playbooks, setPlaybooks] = useState<Playbook[]>([])
  const [loading, setLoading] = useState(false)
  const [err, setErr] = useState<string | null>(null)
  const [executing, setExecuting] = useState<string | null>(null)

  // Execute modal state
  const [selectedPlaybook, setSelectedPlaybook] = useState<Playbook | null>(null)
  const [showExecuteModal, setShowExecuteModal] = useState(false)
  const [parameters, setParameters] = useState<Record<string, any>>({})
  const [reason, setReason] = useState('')
  const [dryRun, setDryRun] = useState(true)
  const [executionResult, setExecutionResult] = useState<any>(null)

  const load = useCallback(async () => {
    if (!token) return
    setLoading(true)
    setErr(null)
    try {
      const data = await getPlaybooks(token)
      setPlaybooks(data.playbooks || [])
    } catch (e: any) {
      setErr(e?.message || '加载失败')
    } finally {
      setLoading(false)
    }
  }, [token])

  useEffect(() => {
    void load()
  }, [load])

  const handleExecute = useCallback(async () => {
    if (!selectedPlaybook || !token) return

    // Validate required parameters
    for (const param of selectedPlaybook.parameters) {
      if (param.required && !parameters[param.name]) {
        alert(`请填写必填参数: ${param.name}`)
        return
      }
    }

    if (!reason.trim()) {
      alert('请填写执行原因（用于审计）')
      return
    }

    setExecuting(selectedPlaybook.id)
    setErr(null)
    try {
      const result = await executePlaybook(token, selectedPlaybook.id, {
        playbook_id: selectedPlaybook.id,
        parameters,
        reason,
        dry_run: dryRun,
      })
      setExecutionResult(result.execution)
      setShowExecuteModal(false)
      setSelectedPlaybook(null)
      setParameters({})
      setReason('')
      setDryRun(true)
    } catch (e: any) {
      setErr(e?.message || '执行失败')
    } finally {
      setExecuting(null)
    }
  }, [selectedPlaybook, token, parameters, reason, dryRun])

  const openExecuteModal = useCallback((pb: Playbook) => {
    setSelectedPlaybook(pb)
    // Initialize parameters with defaults
    const defaults: Record<string, any> = {}
    for (const param of pb.parameters) {
      if (param.default !== undefined) {
        defaults[param.name] = param.default
      }
    }
    setParameters(defaults)
    setReason('')
    setDryRun(true)
    setExecutionResult(null)
    setShowExecuteModal(true)
  }, [])

  const closeModal = useCallback(() => {
    setShowExecuteModal(false)
    setSelectedPlaybook(null)
    setParameters({})
    setReason('')
    setDryRun(true)
  }, [])

  const severityTone = (severity: string): 'ok' | 'warn' | 'bad' | 'neutral' => {
    switch (severity) {
      case 'low': return 'ok'
      case 'medium': return 'warn'
      case 'high': return 'bad'
      case 'critical': return 'bad'
      default: return 'neutral'
    }
  }

  const header = useMemo(() => {
    return (
      <Card title="Playbook 列表" subtitle="预定义的运维操作剧本" right={err ? <Badge tone="bad">{err}</Badge> : <Badge tone="neutral">role: {role}</Badge>}>
        <div className="flex items-center justify-between">
          <div className="text-sm text-slate-200/70">
            共 {playbooks.length} 个可用剧本
          </div>
          <Button tone="ghost" onClick={() => void load()}>
            刷新
          </Button>
        </div>
      </Card>
    )
  }, [playbooks.length, err, role, load])

  return (
    <div className="space-y-6">
      <div>
        <div className="text-xs uppercase tracking-[0.22em] text-slate-300/70">Ops</div>
        <div className="mt-1 text-2xl font-semibold">Playbook 管理</div>
        <div className="mt-2 text-sm text-slate-200/70">预定义的运维操作剧本，支持安全执行和审计</div>
      </div>
      {header}

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
                <div className="font-mono">{executionResult.duration ? `${executionResult.duration}ms` : '-'}</div>
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

      {/* Playbooks Grid */}
      <Card title="可用剧本" right={loading ? <Badge tone="warn">loading</Badge> : null}>
        <div className="grid gap-4 sm:grid-cols-2 lg:grid-cols-3">
          {playbooks.map((pb) => (
            <div
              key={pb.id}
              className="group relative overflow-hidden rounded-2xl border border-white/10 bg-black/20 p-4 transition hover:bg-white/5"
            >
              <div className="flex items-start justify-between">
                <div className="min-w-0 flex-1">
                  <div className="flex items-center gap-2">
                    <h3 className="truncate font-semibold">{pb.name}</h3>
                    <Badge tone={severityTone(pb.severity)}>{pb.severity}</Badge>
                  </div>
                  <p className="mt-1 text-xs text-slate-200/60">{pb.description}</p>
                  <div className="mt-2 flex flex-wrap gap-1">
                    <span className="rounded-full border border-white/10 bg-white/5 px-2 py-0.5 text-[10px] text-slate-200/60">
                      {pb.category}
                    </span>
                    {pb.require_confirm && (
                      <span className="rounded-full border border-amber-500/20 bg-amber-500/10 px-2 py-0.5 text-[10px] text-amber-400">
                        需确认
                      </span>
                    )}
                  </div>
                  {pb.parameters.length > 0 && (
                    <div className="mt-2 text-xs text-slate-200/40">
                      参数: {pb.parameters.map(p => p.name).join(', ')}
                    </div>
                  )}
                </div>
              </div>
              <div className="mt-3 flex gap-2">
                <Button
                  tone="ghost"
                  className="flex-1 text-xs"
                  onClick={() => openExecuteModal(pb)}
                  disabled={executing === pb.id}
                >
                  {executing === pb.id ? '执行中...' : '执行'}
                </Button>
              </div>
            </div>
          ))}
        </div>
        {playbooks.length === 0 && !loading && (
          <div className="py-12 text-center text-slate-200/50">
            暂无可用剧本
          </div>
        )}
      </Card>

      {/* Execute Modal */}
      {showExecuteModal && selectedPlaybook && (
        <div className="fixed inset-0 z-50 flex items-center justify-center bg-black/50 p-4" onClick={closeModal}>
          <div className="w-full max-w-md rounded-3xl border border-white/10 bg-black/80 p-6" onClick={(e) => e.stopPropagation()}>
            <div className="flex items-center justify-between">
              <h3 className="text-lg font-semibold">执行 {selectedPlaybook.name}</h3>
              <button onClick={closeModal} className="text-slate-200/60 hover:text-slate-200">
                ✕
              </button>
            </div>

            <p className="mt-2 text-sm text-slate-200/60">{selectedPlaybook.description}</p>

            {/* Parameters */}
            {selectedPlaybook.parameters.length > 0 && (
              <div className="mt-4 space-y-3">
                <div className="text-sm font-medium">参数</div>
                {selectedPlaybook.parameters.map((param) => (
                  <div key={param.name}>
                    <div className="mb-1 text-xs text-slate-200/60">
                      {param.name}
                      {param.required && <span className="text-rose-400"> *</span>}
                      {param.description && ` - ${param.description}`}
                    </div>
                    {param.type === 'bool' ? (
                      <select
                        className="ops-input w-full rounded-2xl px-3 py-2 text-sm"
                        value={parameters[param.name] || 'false'}
                        onChange={(e) => setParameters({ ...parameters, [param.name]: e.target.value === 'true' })}
                      >
                        <option value="false">false</option>
                        <option value="true">true</option>
                      </select>
                    ) : param.type === 'int' ? (
                      <Input
                        type="number"
                        value={parameters[param.name] || ''}
                        onChange={(e) => setParameters({ ...parameters, [param.name]: parseInt(e.target.value) || 0 })}
                        placeholder={param.default?.toString() || `请输入 ${param.name}`}
                      />
                    ) : (
                      <Input
                        type="text"
                        value={parameters[param.name] || ''}
                        onChange={(e) => setParameters({ ...parameters, [param.name]: e.target.value })}
                        placeholder={param.default?.toString() || `请输入 ${param.name}`}
                      />
                    )}
                  </div>
                ))}
              </div>
            )}

            {/* Reason */}
            <div className="mt-4">
              <div className="mb-1 text-xs text-slate-200/60">执行原因 *</div>
              <TextArea
                value={reason}
                onChange={(e) => setReason(e.target.value)}
                placeholder="请输入执行原因（用于审计）"
                rows={2}
              />
            </div>

            {/* Dry Run */}
            <div className="mt-4 flex items-center gap-2">
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

            {/* Actions */}
            <div className="mt-6 flex gap-3">
              <Button tone="ghost" className="flex-1" onClick={closeModal}>
                取消
              </Button>
              <Button
                tone="primary"
                className="flex-1"
                onClick={handleExecute}
                disabled={executing !== null}
              >
                {executing ? '执行中...' : dryRun ? '预演' : '执行'}
              </Button>
            </div>

            {/* Severity Warning */}
            {selectedPlaybook.severity === 'high' || selectedPlaybook.severity === 'critical' ? (
              <div className="mt-4 rounded-xl bg-rose-500/10 p-3 text-xs text-rose-400">
                ⚠️ 警告：此操作风险等级为 {selectedPlaybook.severity}，请谨慎操作
              </div>
            ) : null}
          </div>
        </div>
      )}
    </div>
  )
}
