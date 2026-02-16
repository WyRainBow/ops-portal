'use client'

import { useMemo, useRef, useState } from 'react'
import { getToken } from '../../../../lib/auth'
import { chat } from '../../../../lib/api'
import { Badge, Button, Card, TextArea } from '../../../../components/Ui'

type Msg = { role: 'user' | 'assistant'; content: string; ts: number }

export default function AssistantChatPage() {
  const token = getToken() || ''
  const [qid, setQid] = useState(() => String(Date.now()))
  const [input, setInput] = useState('')
  const [loading, setLoading] = useState(false)
  const [err, setErr] = useState<string | null>(null)
  const [msgs, setMsgs] = useState<Msg[]>([])
  const boxRef = useRef<HTMLDivElement | null>(null)

  const transcript = useMemo(
    () =>
      msgs
        .map((m) => `${m.role === 'user' ? 'YOU' : 'OPS'}: ${m.content}`)
        .join('\n\n')
        .trim(),
    [msgs],
  )

  const send = async () => {
    const q = input.trim()
    if (!q) return
    setInput('')
    setErr(null)
    setLoading(true)
    const now = Date.now()
    setMsgs((m) => [...m, { role: 'user', content: q, ts: now }])
    try {
      const r = await chat(token, { id: qid, question: q })
      setMsgs((m) => [...m, { role: 'assistant', content: r?.answer || '', ts: Date.now() }])
      setTimeout(() => {
        boxRef.current?.scrollTo({ top: boxRef.current.scrollHeight, behavior: 'smooth' })
      }, 50)
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
          <div className="text-xs uppercase tracking-[0.22em] text-slate-300/70">Assistant</div>
          <div className="mt-1 text-2xl font-semibold">运维助手（只读）</div>
          <div className="mt-2 text-sm text-slate-200/70">v1：先走非流式对话（POST /api/chat）。后续可加 SSE（/api/chat_stream）。</div>
        </div>
        <div className="flex items-center gap-2">
          {err ? <Badge tone="bad">{err}</Badge> : null}
          <Button
            tone="ghost"
            type="button"
            onClick={() => {
              setMsgs([])
              setQid(String(Date.now()))
              setErr(null)
            }}
            disabled={loading}
          >
            新会话
          </Button>
        </div>
      </div>

      <div className="grid grid-cols-1 gap-6 lg:grid-cols-3">
        <Card title="对话" subtitle="建议直接问：最近 1 小时 resume-backend 有什么错误？">
          <div ref={boxRef} className="max-h-[520px] overflow-auto rounded-2xl border border-white/10 bg-black/25 p-4">
            {msgs.length === 0 ? (
              <div className="text-sm text-slate-200/60">还没有消息。输入问题并发送。</div>
            ) : (
              <div className="space-y-4">
                {msgs.map((m, idx) => (
                  <div key={idx} className="space-y-1">
                    <div className="flex items-center gap-2">
                      <Badge tone={m.role === 'user' ? 'neutral' : 'ok'}>{m.role === 'user' ? 'YOU' : 'OPS'}</Badge>
                      <div className="text-xs text-slate-200/60">{new Date(m.ts).toLocaleTimeString()}</div>
                    </div>
                    <div className="whitespace-pre-wrap text-sm leading-relaxed text-slate-100">{m.content}</div>
                  </div>
                ))}
              </div>
            )}
          </div>

          <div className="mt-4 space-y-2">
            <TextArea
              value={input}
              onChange={(e) => setInput(e.target.value)}
              rows={4}
              placeholder="输入你的问题…"
              className="text-sm leading-relaxed"
            />
            <div className="flex items-center justify-between gap-3">
              <div className="text-xs text-slate-200/60">session id: {qid}</div>
              <Button tone="primary" onClick={send} disabled={loading} type="button">
                {loading ? '思考中…' : '发送'}
              </Button>
            </div>
          </div>
        </Card>

        <div className="lg:col-span-2 space-y-6">
          <Card title="Transcript" subtitle="便于复制粘贴（也可用于后续接入“导出为工单”）。" right={
            <Button
              tone="ghost"
              type="button"
              onClick={async () => {
                await navigator.clipboard.writeText(transcript)
              }}
              disabled={!transcript}
            >
              复制
            </Button>
          }>
            <TextArea value={transcript} readOnly rows={22} className="font-mono text-xs leading-relaxed" />
          </Card>
        </div>
      </div>
    </div>
  )
}
