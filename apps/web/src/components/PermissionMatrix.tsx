'use client'

import { Badge } from './Ui'

interface Permission {
  name: string
  description?: string
}

interface RolePermissions {
  [resource: string]: string[] // e.g. "logs": ["read"], "users": ["read", "write"]
}

interface RolesData {
  [role: string]: RolePermissions
}

interface PermissionMatrixProps {
  data: RolesData
}

const resourceLabels: Record<string, string> = {
  logs: '日志',
  members: '成员',
  traces: '追踪',
  users: '用户',
  playbooks: 'Playbook',
  alerts: '告警',
  knowledge: '知识库',
  observability: '可观测性',
}

const permissionLabels: Record<string, { label: string; color: 'ok' | 'warn' | 'bad' | 'neutral' }> = {
  read: { label: '读', color: 'neutral' },
  write: { label: '写', color: 'ok' },
  grant_admin: { label: '授权管理员', color: 'ok' },
  write_non_admin: { label: '写(非管理员)', color: 'neutral' },
  execute: { label: '执行', color: 'ok' },
  delete: { label: '删除', color: 'bad' },
}

const roleLabels: Record<string, string> = {
  admin: '管理员',
  member: '成员',
  user: '普通用户',
}

const roleOrder: Record<string, number> = {
  admin: 0,
  member: 1,
  user: 2,
}

export function PermissionMatrix({ data }: PermissionMatrixProps) {
  const roles = Object.keys(data).sort((a, b) => {
    return (roleOrder[a] ?? 99) - (roleOrder[b] ?? 99)
  })

  // Collect all unique resources across all roles
  const allResources = Array.from(
    new Set(
      roles.flatMap((role) => Object.keys(data[role] || {}))
    )
  ).sort()

  if (roles.length === 0 || allResources.length === 0) {
    return (
      <div className="flex min-h-[200px] items-center justify-center rounded-2xl border border-white/10 bg-white/5">
        <p className="text-slate-200/60">暂无权限数据</p>
      </div>
    )
  }

  return (
    <div className="overflow-auto rounded-2xl border border-white/10">
      <table className="w-full min-w-[600px] text-left text-sm">
        <thead className="bg-white/5 text-xs text-slate-200/70">
          <tr>
            <th className="px-4 py-3 font-medium">资源</th>
            {roles.map((role) => (
              <th key={role} className="px-4 py-3 text-center font-medium">
                {roleLabels[role] || role}
              </th>
            ))}
          </tr>
        </thead>
        <tbody className="divide-y divide-white/5">
          {allResources.map((resource) => (
            <tr key={resource} className="hover:bg-white/5">
              <td className="px-4 py-3 font-medium text-slate-200">
                {resourceLabels[resource] || resource}
              </td>
              {roles.map((role) => {
                const permissions = data[role]?.[resource] || []
                return (
                  <td key={`${role}-${resource}`} className="px-4 py-3">
                    <div className="flex flex-wrap justify-center gap-1">
                      {permissions.length === 0 ? (
                        <span className="text-slate-200/30">-</span>
                      ) : (
                        permissions.map((perm) => {
                          const permInfo = permissionLabels[perm] || {
                            label: perm,
                            color: 'neutral' as const,
                          }
                          return (
                            <Badge key={perm} tone={permInfo.color}>
                              {permInfo.label}
                            </Badge>
                          )
                        })
                      )}
                    </div>
                  </td>
                )
              })}
            </tr>
          ))}
        </tbody>
      </table>
    </div>
  )
}

// Compact view for smaller spaces
interface PermissionMatrixCompactProps {
  data: RolesData
}

export function PermissionMatrixCompact({ data }: PermissionMatrixCompactProps) {
  const roles = Object.keys(data).sort((a, b) => {
    return (roleOrder[a] ?? 99) - (roleOrder[b] ?? 99)
  })

  return (
    <div className="grid gap-3 md:grid-cols-2 lg:grid-cols-3">
      {roles.map((role) => {
        const permissions = data[role] || {}
        const resourceCount = Object.keys(permissions).length

        return (
          <div
            key={role}
            className="rounded-2xl border border-white/10 bg-white/5 p-4"
          >
            <div className="mb-3 flex items-center justify-between">
              <span className="text-sm font-semibold">
                {roleLabels[role] || role}
              </span>
              <Badge tone="neutral">
                {resourceCount} 资源
              </Badge>
            </div>
            <div className="space-y-2 text-xs">
              {Object.entries(permissions).map(([resource, perms]) => (
                <div
                  key={resource}
                  className="flex items-center justify-between"
                >
                  <span className="text-slate-200/70">
                    {resourceLabels[resource] || resource}
                  </span>
                  <div className="flex gap-1">
                    {perms.map((perm) => {
                      const permInfo = permissionLabels[perm] || {
                        label: perm,
                        color: 'neutral' as const,
                      }
                      return (
                        <Badge key={perm} tone={permInfo.color}>
                          {permInfo.label}
                        </Badge>
                      )
                    })}
                  </div>
                </div>
              ))}
            </div>
          </div>
        )
      })}
    </div>
  )
}
