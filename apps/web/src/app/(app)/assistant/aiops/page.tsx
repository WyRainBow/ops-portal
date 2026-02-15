'use client'

import { useState } from 'react'
import { getToken } from '../../../../lib/auth'
import { aiOps } from '../../../../lib/api'
import { Badge, Button, Card, TextArea } from '../../../../components/Ui'

export default function AIOpsPage() {
  const token = getToken() || ''
  const [loading, setLoading] = useState(false)
  const [err, setErr] = useState<string | null>(null)
  const [result, setResult] = useState<string>('')
  const [detail, setDetail] = useState<string[]>([])

  const run = async () => {
    setLoading(true)
    setErr(null)
    setResult('')
    setDetail([])
    try {
      const r = await aiOps(token)
      setResult(r?.result || '')
      setDetail(r?.detail || [])
    } catch (e: any) {
      setErr(e?.message || String(e))
    } finally {
      setLoading(false)
    }
  }

  return (
    <div className="space-y-6">
      <div className="flex items-end justify-between gap-4">
        <div>
          <div className="text-xs uppercase tracking-[0.22em] text-slate-300/70">AIOps</div>
          <div className="mt-1 text-2xl font-semibold">告警分析助手（只读）</div>
          <div className="mt-2 text-sm text-slate-200/70">会查询 Prometheus alerts + Loki 最近错误日志，然后生成一段可读的分析报告。</div>
        </div>
        <div className="flex items-center gap-2">
          {err ? <Badge tone="bad">{err}</Badge> : null}
          <Button tone="primary" onClick={run} disabled={loading} type="button">
            {loading ? '生成中…' : '生成分析报告'}
          </Button>
        </div>
      </div>

      <div className="grid grid-cols-1 gap-6 lg:grid-cols-3">
        <Card title="报告" subtitle="result（面向人类的总结）。" className="lg:col-span-2">
          <TextArea value={result} readOnly rows={16} className="text-sm leading-relaxed" placeholder="点击右上角按钮生成…" />
        </Card>
        <Card title="工具输出" subtitle="detail（工具调用的结构化/半结构化内容）。">
          <TextArea value={(detail || []).join('\n')} readOnly rows={16} className="font-mono text-xs leading-relaxed" />
        </Card>
      </div>
    </div>
  )
}
