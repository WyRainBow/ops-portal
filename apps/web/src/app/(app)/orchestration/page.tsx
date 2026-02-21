'use client'

import { useCallback, useEffect, useState } from 'react'
import { getToken } from '../../../lib/auth'
import { aiOps } from '../../../lib/api'
import { Badge, Button, Card, Input, TextArea } from '../../../components/Ui'

interface Task {
  id: string
  description: string
  status: 'pending' | 'planning' | 'executing' | 'completed' | 'failed'
  plan?: Step[]
  result?: string
  created_at: string
  completed_at?: string
}

interface Step {
  id: string
  name: string
  type: 'diagnose' | 'query' | 'execute' | 'verify'
  status: 'pending' | 'running' | 'completed' | 'failed'
  result?: any
}

export default function OrchestrationPage() {
  const token = getToken() || ''

  const [tasks, setTasks] = useState<Task[]>([])
  const [loading, setLoading] = useState(false)
  const [processing, setProcessing] = useState(false)
  const [err, setErr] = useState<string | null>(null)

  // Create task form
  const [description, setDescription] = useState('')
  const [currentTask, setCurrentTask] = useState<Task | null>(null)

  const loadTasks = useCallback(() => {
    // In production, this would fetch from API
    setTasks([])
  }, [])

  useEffect(() => {
    void loadTasks()
  }, [loadTasks])

  const handleCreateTask = useCallback(async () => {
    if (!description.trim() || !token) return

    setProcessing(true)
    setErr(null)

    // Create a temporary task for demo
    const newTask: Task = {
      id: `TASK-${Date.now()}`,
      description: description,
      status: 'planning',
      created_at: new Date().toISOString(),
    }

    setCurrentTask(newTask)
    setDescription('')

    try {
      // Call AI Ops to generate plan
      const result = await aiOps(token)

      // Simulate plan generation
      setCurrentTask({
        ...newTask,
        status: 'completed',
        plan: [
          { id: '1', name: '查询活跃告警', type: 'diagnose', status: 'completed' },
          { id: '2', name: '查询相关日志', type: 'query', status: 'completed' },
          { id: '3', name: '分析根因', type: 'diagnose', status: 'completed' },
          { id: '4', name: '生成处理建议', type: 'diagnose', status: 'completed' },
        ],
        result: result.result || JSON.stringify(result.detail, null, 2),
        completed_at: new Date().toISOString(),
      })
    } catch (e: any) {
      setErr(e?.message || '创建任务失败')
      setCurrentTask(null)
    } finally {
      setProcessing(false)
    }
  }, [description, token])

  const statusTone = (status: string): 'ok' | 'warn' | 'bad' | 'neutral' => {
    switch (status) {
      case 'completed': return 'ok'
      case 'failed': return 'bad'
      case 'executing':
      case 'planning': return 'warn'
      default: return 'neutral'
    }
  }

  const stepTypeLabel = (type: string) => {
    switch (type) {
      case 'diagnose': return '诊断'
      case 'query': return '查询'
      case 'execute': return '执行'
      case 'verify': return '验证'
      default: return type
    }
  }

  return (
    <div className="space-y-6">
      <div>
        <div className="text-xs uppercase tracking-[0.22em] text-slate-300/70">Ops</div>
        <div className="mt-1 text-2xl font-semibold">运维编排</div>
        <div className="mt-2 text-sm text-slate-200/70">Plan-Execute-Replan 智能运维任务编排</div>
      </div>

      {/* Create Task */}
      <Card title="创建运维任务" subtitle="描述问题，AI 将自动生成执行计划" right={err ? <Badge tone="bad">{err}</Badge> : null}>
        <div className="space-y-4">
          <TextArea
            value={description}
            onChange={(e) => setDescription(e.target.value)}
            placeholder="描述需要处理的运维问题，例如：&quot;服务器 CPU 使用率过高，请帮我诊断并处理&quot;"
            rows={3}
          />
          <div className="flex justify-end">
            <Button
              tone="primary"
              onClick={handleCreateTask}
              disabled={!description.trim() || processing}
            >
              {processing ? '分析中...' : '生成执行计划'}
            </Button>
          </div>
          <div className="text-xs text-slate-200/50">
            AI 将自动：1. 分析问题 → 2. 查询告警和日志 → 3. 诊断根因 → 4. 生成处理建议
          </div>
        </div>
      </Card>

      {/* Current Task Result */}
      {currentTask && (
        <Card title="执行结果" subtitle={currentTask.id} right={<Badge tone={statusTone(currentTask.status)}>{currentTask.status}</Badge>}>
          <div className="space-y-4">
            {/* Plan Steps */}
            {currentTask.plan && (
              <div>
                <div className="mb-3 text-sm font-medium">执行步骤</div>
                <div className="space-y-2">
                  {currentTask.plan.map((step, idx) => (
                    <div key={step.id} className="flex items-center gap-3 rounded-xl border border-white/10 bg-white/5 p-3">
                      <div className="flex h-6 w-6 shrink-0 items-center justify-center rounded-full bg-white/10 text-xs">
                        {idx + 1}
                      </div>
                      <div className="min-w-0 flex-1">
                        <div className="text-sm">{step.name}</div>
                        <div className="mt-1 text-xs text-slate-200/60">
                          <Badge tone="neutral">{stepTypeLabel(step.type)}</Badge>
                          <span className="ml-2">{step.status}</span>
                        </div>
                      </div>
                      <Badge tone={statusTone(step.status)}>{step.status}</Badge>
                    </div>
                  ))}
                </div>
              </div>
            )}

            {/* Result */}
            {currentTask.result && (
              <div>
                <div className="mb-2 text-sm font-medium">诊断结果</div>
                <div className="overflow-x-auto rounded-xl bg-black/30 p-4 font-mono text-xs text-slate-300">
                  <pre className="whitespace-pre-wrap">{currentTask.result}</pre>
                </div>
              </div>
            )}

            {/* Actions */}
            {currentTask.status === 'completed' && (
              <div className="flex gap-3">
                <Button tone="ghost" onClick={() => setCurrentTask(null)}>
                  关闭
                </Button>
                <Button tone="primary" onClick={() => {
                  // Navigate to Playbook execution
                  window.location.href = '/ops'
                }}>
                  执行修复操作
                </Button>
              </div>
            )}
          </div>
        </Card>
      )}

      {/* Task History */}
      <Card title="任务历史" right={<Badge tone="neutral">{tasks.length} 条</Badge>}>
        <div className="text-center py-8 text-slate-200/50">
          暂无历史任务
        </div>
      </Card>

      {/* Info Card */}
      <Card title="编排流程说明" right={<Badge tone="neutral">PER</Badge>}>
        <div className="space-y-3 text-sm text-slate-200/70">
          <div className="flex items-center gap-4">
            <div className="flex h-8 w-8 shrink-0 items-center justify-center rounded-full bg-white/10 text-xs font-semibold">
              1
            </div>
            <div>
              <div className="font-medium text-slate-200">Plan（规划）</div>
              <div className="text-xs text-slate-200/50">AI 分析问题，生成执行计划</div>
            </div>
          </div>
          <div className="flex items-center gap-4">
            <div className="flex h-8 w-8 shrink-0 items-center justify-center rounded-full bg-white/10 text-xs font-semibold">
              2
            </div>
            <div>
              <div className="font-medium text-slate-200">Execute（执行）</div>
              <div className="text-xs text-slate-200/50">按计划执行诊断、查询、修复操作</div>
            </div>
          </div>
          <div className="flex items-center gap-4">
            <div className="flex h-8 w-8 shrink-0 items-center justify-center rounded-full bg-white/10 text-xs font-semibold">
              3
            </div>
            <div>
              <div className="font-medium text-slate-200">Replan（重新规划）</div>
              <div className="text-xs text-slate-200/50">如执行失败，AI 重新规划并尝试</div>
            </div>
          </div>
        </div>
      </Card>
    </div>
  )
}
