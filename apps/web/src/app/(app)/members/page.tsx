'use client'

import { useCallback, useEffect, useState } from 'react'
import { createMember, deleteMember, getMembers, updateMember } from '../../../lib/api'
import { getToken, parseJwtRole } from '../../../lib/auth'
import { Badge, Button, Card, Input } from '../../../components/Ui'

export default function MembersPage() {
  const token = getToken() || ''
  const role = parseJwtRole(token)
  const isAdmin = role === 'admin'

  const [keyword, setKeyword] = useState('')
  const [items, setItems] = useState<any[]>([])
  const [page, setPage] = useState(1)
  const [pageSize] = useState(20)
  const [err, setErr] = useState<string | null>(null)

  const load = useCallback(async () => {
    if (!token) return
    setErr(null)
    try {
      const data = await getMembers(token, { page, page_size: pageSize, keyword })
      setItems(data.items || [])
    } catch (e: any) {
      setErr(e?.message || '加载失败')
    }
  }, [keyword, page, pageSize, token])

  useEffect(() => {
    void load()
  }, [load])

  const onCreate = async () => {
    if (!isAdmin || !token) return
    const userId = prompt('user_id', '')
    if (!userId) return
    const position = prompt('position (可选)', '') || ''
    const team = prompt('team (可选)', '') || ''
    try {
      await createMember(token, { user_id: Number(userId), position, team, status: 'active', user_role: 'member' })
      await load()
    } catch (e: any) {
      alert(e?.message || '创建失败')
    }
  }

  const onEdit = async (m: any) => {
    if (!isAdmin || !token) return
    const userId = prompt('user_id', String(m.user_id || ''))
    if (!userId) return
    const position = prompt('position', m.position || '') || ''
    const team = prompt('team', m.team || '') || ''
    const status = prompt('status', m.status || 'active') || 'active'
    try {
      await updateMember(token, m.id, { user_id: Number(userId), position, team, status, user_role: m.user_role || '' })
      await load()
    } catch (e: any) {
      alert(e?.message || '更新失败')
    }
  }

  const onDelete = async (m: any) => {
    if (!isAdmin || !token) return
    if (!confirm(`确认删除成员 ${m.name} (${m.id}) ?`)) return
    try {
      await deleteMember(token, m.id)
      await load()
    } catch (e: any) {
      alert(e?.message || '删除失败')
    }
  }

  return (
    <div className="space-y-6">
      <div className="flex flex-wrap items-end justify-between gap-4">
        <div>
          <div className="text-xs uppercase tracking-[0.22em] text-slate-300/70">Admin</div>
          <div className="mt-1 text-2xl font-semibold">成员</div>
          <div className="mt-2 text-sm text-slate-200/70">映射到 resume_db.members（admin 可增删改）。</div>
        </div>
        <div className="flex items-center gap-2">
          {err ? <Badge tone="bad">{err}</Badge> : <Badge tone="neutral">role: {role}</Badge>}
          {isAdmin ? (
            <Button onClick={onCreate} tone="primary" type="button">
              新建成员
            </Button>
          ) : null}
        </div>
      </div>

      <Card title="筛选" subtitle="keyword 支持模糊匹配 name/username。">
        <div className="flex flex-wrap items-end gap-3">
          <div className="min-w-[240px]">
            <div className="mb-1 text-xs text-slate-200/70">关键词</div>
            <Input value={keyword} onChange={(e) => setKeyword(e.target.value)} />
          </div>
          <Button
            type="button"
            onClick={() => {
              setPage(1)
              void load()
            }}
            tone="ghost"
          >
            查询
          </Button>
        </div>
      </Card>

      <Card title="列表" subtitle={`page=${page} size=${pageSize}`}>
      <div className="overflow-hidden rounded-2xl border border-white/10 bg-black/20">
        <table className="w-full text-left text-sm">
          <thead className="bg-white/5 text-xs font-semibold uppercase tracking-wide text-slate-200/70">
            <tr>
              <th className="px-4 py-3">ID</th>
              <th className="px-4 py-3">Name</th>
              <th className="px-4 py-3">Username</th>
              <th className="px-4 py-3">Team</th>
              <th className="px-4 py-3">Status</th>
              <th className="px-4 py-3">操作</th>
            </tr>
          </thead>
          <tbody>
            {items.map((m) => (
              <tr key={m.id} className="border-t border-white/5 hover:bg-white/5">
                <td className="px-4 py-3 font-mono">{m.id}</td>
                <td className="px-4 py-3">{m.name}</td>
                <td className="px-4 py-3">{m.username || '-'}</td>
                <td className="px-4 py-3">{m.team || '-'}</td>
                <td className="px-4 py-3">
                  <Badge tone={m.status === 'active' ? 'ok' : 'neutral'}>{m.status}</Badge>
                </td>
                <td className="px-4 py-3">
                  {isAdmin ? (
                    <div className="flex gap-2">
                      <Button tone="ghost" onClick={() => onEdit(m)} type="button">
                        编辑
                      </Button>
                      <Button tone="danger" onClick={() => onDelete(m)} type="button">
                        删除
                      </Button>
                    </div>
                  ) : (
                    <span className="text-xs text-slate-200/40">只读</span>
                  )}
                </td>
              </tr>
            ))}
            {items.length === 0 ? (
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
