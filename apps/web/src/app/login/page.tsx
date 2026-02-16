'use client'

import { useState } from 'react'
import { useRouter } from 'next/navigation'
import { login } from '../../lib/api'
import { setToken } from '../../lib/auth'
import { Badge, Button, Card, Input } from '../../components/Ui'

export default function LoginPage() {
  const router = useRouter()
  const [username, setUsername] = useState('')
  const [password, setPassword] = useState('')
  const [loading, setLoading] = useState(false)
  const [err, setErr] = useState<string | null>(null)

  const submit = async (e: React.FormEvent) => {
    e.preventDefault()
    setErr(null)
    setLoading(true)
    try {
      const data = await login(username, password)
      setToken(data.access_token)
      router.replace('/overview')
    } catch (e: any) {
      setErr(e?.message || '登录失败')
    } finally {
      setLoading(false)
    }
  }

  return (
    <div className="min-h-screen">
      <div className="ops-grid">
        <div className="mx-auto flex max-w-[980px] flex-col gap-6 px-4 py-16">
          <div className="flex items-start justify-between gap-4">
            <div>
              <div className="text-xs uppercase tracking-[0.22em] text-slate-300/70">Ops Portal</div>
              <div className="mt-2 text-3xl font-semibold">登录</div>
              <div className="mt-2 text-sm text-slate-200/70">建议：仅通过 SSH 隧道访问，并在 Nginx 额外启用 Basic Auth。</div>
            </div>
            {err ? <Badge tone="bad">{err}</Badge> : null}
          </div>

          <div className="grid grid-cols-1 gap-6 lg:grid-cols-2">
            <Card
              title="凭据"
              subtitle="JWT 登录（admin/member）。"
              right={
                <Badge tone="neutral" >
                  /api/auth/login
                </Badge>
              }
            >
              <form onSubmit={submit} className="space-y-3">
                <div>
                  <div className="mb-1 text-xs text-slate-200/70">用户名</div>
                  <Input value={username} onChange={(e) => setUsername(e.target.value)} placeholder="cocoyu" autoComplete="username" />
                </div>
                <div>
                  <div className="mb-1 text-xs text-slate-200/70">密码</div>
                  <Input
                    value={password}
                    onChange={(e) => setPassword(e.target.value)}
                    type="password"
                    placeholder="0000"
                    autoComplete="current-password"
                  />
                </div>
                <Button tone="primary" type="submit" disabled={loading} className="w-full">
                  {loading ? '登录中…' : '登录'}
                </Button>
              </form>
            </Card>

            <Card title="访问方式" subtitle="单机部署默认仅监听 127.0.0.1。">
              <pre className="rounded-2xl border border-white/10 bg-black/30 p-4 text-xs leading-relaxed text-slate-100">
{`# Portal
ssh -i ~/.ssh/id_rsa -p 2222 -L 18080:127.0.0.1:18080 root@106.53.113.137

# Grafana
ssh -i ~/.ssh/id_rsa -p 2222 -L 3000:127.0.0.1:3000 root@106.53.113.137`}
              </pre>
            </Card>
          </div>
        </div>
      </div>
    </div>
  )
}
