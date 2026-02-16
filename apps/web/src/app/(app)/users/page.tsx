'use client'

import { useCallback, useEffect, useMemo, useState } from 'react'
import { getToken, parseJwtRole } from '../../../lib/auth'
import { getUsers, updateUserQuota, updateUserRole } from '../../../lib/api'
import { Badge, Button, Card, Input } from '../../../components/Ui'

export default function UsersPage() {
  const token = getToken() || ''
  const role = parseJwtRole(token)
  const isAdmin = role === 'admin'

  const [keyword, setKeyword] = useState('')
  const [userRole, setUserRole] = useState('')
  const [ip, setIp] = useState('')
  const [items, setItems] = useState<any[]>([])
  const [page, setPage] = useState(1)
  const [pageSize] = useState(20)
  const [loading, setLoading] = useState(false)
  const [err, setErr] = useState<string | null>(null)

  const load = useCallback(async () => {
    if (!token) return
    setLoading(true)
    setErr(null)
    try {
      const data = await getUsers(token, { page, page_size: pageSize, keyword, role: userRole, ip, with_total: true })
      setItems(data.items || [])
    } catch (e: any) {
      setErr(e?.message || '加载失败')
    } finally {
      setLoading(false)
    }
  }, [ip, keyword, page, pageSize, token, userRole])

  useEffect(() => {
    void load()
  }, [load])

  const onUpdateRole = async (id: number) => {
    const r = prompt('输入新角色：admin/member/user', 'member')
    if (!r || !token) return
    try {
      await updateUserRole(token, id, r)
      await load()
    } catch (e: any) {
      alert(e?.message || '更新失败')
    }
  }

  const onUpdateQuota = async (id: number) => {
    const v = prompt('输入额度（空表示不限制）', '')
    if (!token) return
    const quota = v && v.trim() !== '' ? Number(v) : null
    try {
      await updateUserQuota(token, id, quota)
      await load()
    } catch (e: any) {
      alert(e?.message || '更新失败')
    }
  }

  const header = useMemo(() => {
    return (
      <Card title="筛选" subtitle="keyword/role/ip 支持模糊筛选（role 需精确）。" right={err ? <Badge tone="bad">{err}</Badge> : <Badge tone="neutral">role: {role}</Badge>}>
        <div className="flex flex-wrap items-end gap-3">
        <div>
          <div className="mb-1 text-xs text-slate-200/70">关键词</div>
          <Input value={keyword} onChange={(e) => setKeyword(e.target.value)} />
        </div>
        <div>
          <div className="mb-1 text-xs text-slate-200/70">角色</div>
          <Input value={userRole} onChange={(e) => setUserRole(e.target.value)} placeholder="admin/member/user" />
        </div>
        <div>
          <div className="mb-1 text-xs text-slate-200/70">IP</div>
          <Input value={ip} onChange={(e) => setIp(e.target.value)} />
        </div>
        <Button
          type="button"
          onClick={() => {
            setPage(1)
            void load()
          }}
          tone="primary"
        >
          查询
        </Button>
      </div>
      </Card>
    )
  }, [keyword, userRole, ip, err, role, load])

  return (
    <div className="space-y-6">
      <div>
        <div className="text-xs uppercase tracking-[0.22em] text-slate-300/70">Admin</div>
        <div className="mt-1 text-2xl font-semibold">用户</div>
        <div className="mt-2 text-sm text-slate-200/70">admin 可改 role/quota；member 只读。</div>
      </div>
      {header}
      <Card title="列表" subtitle={`page=${page} size=${pageSize}`} right={loading ? <Badge tone="warn">loading</Badge> : null}>
      <div className="overflow-hidden rounded-2xl border border-white/10 bg-black/20">
        <table className="w-full text-left text-sm">
          <thead className="bg-white/5 text-xs font-semibold uppercase tracking-wide text-slate-200/70">
            <tr>
              <th className="px-4 py-3">ID</th>
              <th className="px-4 py-3">用户名</th>
              <th className="px-4 py-3">角色</th>
              <th className="px-4 py-3">额度</th>
              <th className="px-4 py-3">最近 IP</th>
              <th className="px-4 py-3">操作</th>
            </tr>
          </thead>
          <tbody>
            {items.map((u) => (
              <tr key={u.id} className="border-t border-white/5 hover:bg-white/5">
                <td className="px-4 py-3 font-mono">{u.id}</td>
                <td className="px-4 py-3">{u.username}</td>
                <td className="px-4 py-3">
                  <Badge tone={u.role === 'admin' ? 'warn' : u.role === 'member' ? 'ok' : 'neutral'}>{u.role}</Badge>
                </td>
                <td className="px-4 py-3 font-mono">{u.api_quota ?? '-'}</td>
                <td className="px-4 py-3 font-mono">{u.last_login_ip || '-'}</td>
                <td className="px-4 py-3">
                  {isAdmin ? (
                    <div className="flex gap-2">
                      <Button tone="ghost" onClick={() => onUpdateRole(u.id)} type="button">
                        改角色
                      </Button>
                      <Button tone="ghost" onClick={() => onUpdateQuota(u.id)} type="button">
                        改额度
                      </Button>
                    </div>
                  ) : (
                    <span className="text-xs text-slate-200/40">只读</span>
                  )}
                </td>
              </tr>
            ))}
            {items.length === 0 && !loading ? (
              <tr>
                <td className="px-4 py-8 text-center text-slate-200/60" colSpan={6}>
                  无数据
                </td>
              </tr>
            ) : null}
          </tbody>
        </table>
      </div>
      </Card>
    </div>
  )
}
