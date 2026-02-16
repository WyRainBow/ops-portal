'use client'

import { useEffect, useState } from 'react'
import { getToken, parseJwtRole } from '../../../lib/auth'
import { getPermissionAudits, getPermissionRoles } from '../../../lib/api'
import { clampInt, formatRFC3339 } from '../../../lib/format'
import { Badge, Button, Card, Select } from '../../../components/Ui'

export default function PermissionsPage() {
  const token = getToken() || ''
  const role = parseJwtRole(token)

  const [loading, setLoading] = useState(false)
  const [err, setErr] = useState<string | null>(null)
  const [roles, setRoles] = useState<any | null>(null)

  const [auditsLoading, setAuditsLoading] = useState(false)
  const [auditsErr, setAuditsErr] = useState<string | null>(null)
  const [audits, setAudits] = useState<any | null>(null)
  const [page, setPage] = useState(1)
  const [pageSize, setPageSize] = useState(50)

  const refreshRoles = async () => {
    setLoading(true)
    setErr(null)
    try {
      const r = await getPermissionRoles(token)
      setRoles(r)
    } catch (e: any) {
      setErr(e?.message || String(e))
    } finally {
      setLoading(false)
    }
  }

  const refreshAudits = async () => {
    setAuditsLoading(true)
    setAuditsErr(null)
    try {
      const a = await getPermissionAudits(token, { page, page_size: pageSize })
      setAudits(a)
    } catch (e: any) {
      setAuditsErr(e?.message || String(e))
    } finally {
      setAuditsLoading(false)
    }
  }

  useEffect(() => {
    refreshRoles()
    // eslint-disable-next-line react-hooks/exhaustive-deps
  }, [])

  useEffect(() => {
    refreshAudits()
    // eslint-disable-next-line react-hooks/exhaustive-deps
  }, [page, pageSize])

  const auditItems = (audits?.items || []) as any[]
  const total = Number(audits?.total || 0)
  const totalPages = Math.max(1, Math.ceil(total / pageSize))

  return (
    <div className="space-y-6">
      <div className="flex items-end justify-between gap-4">
        <div>
          <div className="text-xs uppercase tracking-[0.22em] text-slate-300/70">RBAC</div>
          <div className="mt-1 text-2xl font-semibold">权限矩阵与审计</div>
          <div className="mt-2 text-sm text-slate-200/70">当前登录角色：{role}（member 只读；admin 可在用户页更改角色）。</div>
        </div>
        <div className="flex items-center gap-2">
          {err ? <Badge tone="bad">{err}</Badge> : null}
          <Button tone="primary" onClick={refreshRoles} disabled={loading} type="button">
            {loading ? '刷新中…' : '刷新矩阵'}
          </Button>
        </div>
      </div>

      <Card title="角色权限矩阵" subtitle="后端返回为结构化 JSON（v1 先做展示，后续可做成可视化矩阵）。">
        <pre className="max-h-[380px] overflow-auto rounded-2xl border border-white/10 bg-black/30 p-4 text-xs leading-relaxed text-slate-100">
          {JSON.stringify(roles?.roles || roles || {}, null, 2)}
        </pre>
      </Card>

      <Card
        title="权限审计"
        subtitle="谁在什么时候把谁的角色从 A 改成 B（或其他动作）。"
        right={
          <div className="flex items-center gap-2">
            {auditsErr ? <Badge tone="bad">{auditsErr}</Badge> : <Badge tone="neutral">total: {total}</Badge>}
            <Button tone="ghost" onClick={refreshAudits} disabled={auditsLoading} type="button">
              {auditsLoading ? '刷新中…' : '刷新'}
            </Button>
          </div>
        }
      >
        <div className="mb-3 flex flex-wrap items-center gap-3">
          <div className="text-xs text-slate-200/70">
            Page {page}/{totalPages}
          </div>
          <div className="w-[120px]">
            <Select value={String(pageSize)} onChange={(e) => setPageSize(clampInt(Number(e.target.value), 5, 200))}>
              {[20, 50, 100, 200].map((n) => (
                <option key={n} value={String(n)}>
                  {n} / page
                </option>
              ))}
            </Select>
          </div>
          <Button tone="ghost" onClick={() => setPage((p) => Math.max(1, p - 1))} disabled={page <= 1 || auditsLoading} type="button">
            上一页
          </Button>
          <Button tone="ghost" onClick={() => setPage((p) => Math.min(totalPages, p + 1))} disabled={page >= totalPages || auditsLoading} type="button">
            下一页
          </Button>
        </div>

        <div className="overflow-auto rounded-2xl border border-white/10">
          <table className="w-full min-w-[980px] text-left text-sm">
            <thead className="bg-white/5 text-xs text-slate-200/70">
              <tr>
                <th className="px-3 py-2">time</th>
                <th className="px-3 py-2">operator</th>
                <th className="px-3 py-2">target</th>
                <th className="px-3 py-2">action</th>
                <th className="px-3 py-2">from</th>
                <th className="px-3 py-2">to</th>
              </tr>
            </thead>
            <tbody className="divide-y divide-white/5">
              {auditItems.map((it) => (
                <tr key={String(it.id)} className="hover:bg-white/5">
                  <td className="px-3 py-2 text-slate-200/80">{formatRFC3339(it.created_at)}</td>
                  <td className="px-3 py-2 text-slate-200/80">{it.operator_username || it.operator_user_id || ''}</td>
                  <td className="px-3 py-2 text-slate-200/80">{it.target_username || it.target_user_id || ''}</td>
                  <td className="px-3 py-2">{it.action ? <Badge tone="neutral">{it.action}</Badge> : ''}</td>
                  <td className="px-3 py-2 text-slate-200/80">{it.from_role || ''}</td>
                  <td className="px-3 py-2 text-slate-200/80">{it.to_role || ''}</td>
                </tr>
              ))}
              {auditItems.length === 0 ? (
                <tr>
                  <td className="px-3 py-6 text-center text-slate-200/60" colSpan={6}>
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
