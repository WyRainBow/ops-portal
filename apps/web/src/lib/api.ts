type Wrapped<T> = { message?: string; data?: T }

async function request<T>(path: string, init?: RequestInit): Promise<T> {
  let res: Response
  try {
    res = await fetch(path, {
      ...init,
      headers: {
        'content-type': 'application/json',
        ...(init?.headers || {}),
      },
      cache: 'no-store',
    })
  } catch (e) {
    const msg = e instanceof Error ? e.message : String(e)
    throw new Error(`请求失败（后端未启动或网络错误）: ${msg}`)
  }

  const contentType = res.headers.get('content-type') ?? ''
  const raw = await res.text()

  let body: Wrapped<T>
  if (raw === '') {
    body = {} as Wrapped<T>
  } else if (contentType.includes('application/json')) {
    try {
      body = JSON.parse(raw) as Wrapped<T>
    } catch {
      throw new Error(`接口返回非 JSON: ${raw.slice(0, 100)}${raw.length > 100 ? '…' : ''}`)
    }
  } else {
    // 后端返回了 HTML/纯文本（如 500 Internal Server Error）
    const preview = raw.slice(0, 80).replace(/\s+/g, ' ')
    throw new Error(res.ok ? `接口返回非 JSON: ${preview}…` : `HTTP ${res.status}: ${preview}…`)
  }

  // GoFrame middleware wraps responses; errors may still be HTTP 200 with message.
  if (!res.ok) {
    throw new Error(body?.message || `HTTP ${res.status}`)
  }
  if (body && typeof body === 'object' && 'data' in body) {
    // If handler errored, data is usually null and message is not "OK".
    if ((body as any).data == null && body.message && body.message !== 'OK') {
      throw new Error(body.message)
    }
    return (body as any).data as T
  }
  return body as any
}

export async function login(username: string, password: string) {
  return request<{ access_token: string; token_type: string; user: any }>(`/api/auth/login`, {
    method: 'POST',
    body: JSON.stringify({ username, password }),
  })
}

export async function me(token: string) {
  return request<any>(`/api/auth/me`, {
    method: 'GET',
    headers: { Authorization: `Bearer ${token}` },
  })
}

export async function getOverview(token: string) {
  return request<any>(`/api/admin/overview`, {
    method: 'GET',
    headers: { Authorization: `Bearer ${token}` },
  })
}

export async function getApiRoutes(
  token: string,
  params: { q?: string; tag?: string; method?: string; hide_docs?: boolean } = {}
) {
  const qs = new URLSearchParams()
  Object.entries(params).forEach(([k, v]) => {
    if (v === undefined || v === null || v === '') return
    qs.set(k, String(v))
  })
  return request<any>(`/api/admin/api/routes?${qs.toString()}`, {
    method: 'GET',
    headers: { Authorization: `Bearer ${token}` },
  })
}

export async function getUsers(token: string, params: Record<string, any>) {
  const qs = new URLSearchParams()
  Object.entries(params).forEach(([k, v]) => {
    if (v === undefined || v === null || v === '') return
    qs.set(k, String(v))
  })
  return request<any>(`/api/admin/users?${qs.toString()}`, {
    method: 'GET',
    headers: { Authorization: `Bearer ${token}` },
  })
}

export async function updateUserRole(token: string, userId: number, role: string) {
  return request<any>(`/api/admin/users/${userId}/role`, {
    method: 'PATCH',
    headers: { Authorization: `Bearer ${token}` },
    body: JSON.stringify({ role }),
  })
}

export async function updateUserQuota(token: string, userId: number, api_quota: number | null) {
  return request<any>(`/api/admin/users/${userId}/quota`, {
    method: 'PATCH',
    headers: { Authorization: `Bearer ${token}` },
    body: JSON.stringify({ api_quota }),
  })
}

export async function getMembers(token: string, params: Record<string, any>) {
  const qs = new URLSearchParams()
  Object.entries(params).forEach(([k, v]) => {
    if (v === undefined || v === null || v === '') return
    qs.set(k, String(v))
  })
  return request<any>(`/api/admin/members?${qs.toString()}`, {
    method: 'GET',
    headers: { Authorization: `Bearer ${token}` },
  })
}

export async function createMember(token: string, payload: any) {
  return request<any>(`/api/admin/members`, {
    method: 'POST',
    headers: { Authorization: `Bearer ${token}` },
    body: JSON.stringify(payload),
  })
}

export async function updateMember(token: string, memberId: number, payload: any) {
  return request<any>(`/api/admin/members/${memberId}`, {
    method: 'PATCH',
    headers: { Authorization: `Bearer ${token}` },
    body: JSON.stringify(payload),
  })
}

export async function deleteMember(token: string, memberId: number) {
  return request<any>(`/api/admin/members/${memberId}`, {
    method: 'DELETE',
    headers: { Authorization: `Bearer ${token}` },
  })
}

export async function getRequestLogs(token: string, params: Record<string, any>) {
  const qs = new URLSearchParams()
  Object.entries(params).forEach(([k, v]) => {
    if (v === undefined || v === null || v === '') return
    qs.set(k, String(v))
  })
  return request<any>(`/api/admin/logs/requests?${qs.toString()}`, {
    method: 'GET',
    headers: { Authorization: `Bearer ${token}` },
  })
}

export async function getErrorLogs(token: string, params: Record<string, any>) {
  const qs = new URLSearchParams()
  Object.entries(params).forEach(([k, v]) => {
    if (v === undefined || v === null || v === '') return
    qs.set(k, String(v))
  })
  return request<any>(`/api/admin/logs/errors?${qs.toString()}`, {
    method: 'GET',
    headers: { Authorization: `Bearer ${token}` },
  })
}

export async function getTraces(token: string, params: Record<string, any>) {
  const qs = new URLSearchParams()
  Object.entries(params).forEach(([k, v]) => {
    if (v === undefined || v === null || v === '') return
    qs.set(k, String(v))
  })
  return request<any>(`/api/admin/traces?${qs.toString()}`, {
    method: 'GET',
    headers: { Authorization: `Bearer ${token}` },
  })
}

export async function getTraceDetail(token: string, traceId: string) {
  return request<any>(`/api/admin/traces/${encodeURIComponent(traceId)}`, {
    method: 'GET',
    headers: { Authorization: `Bearer ${token}` },
  })
}

export async function getPermissionRoles(token: string) {
  return request<any>(`/api/admin/permissions/roles`, {
    method: 'GET',
    headers: { Authorization: `Bearer ${token}` },
  })
}

export async function getPermissionAudits(token: string, params: Record<string, any>) {
  const qs = new URLSearchParams()
  Object.entries(params).forEach(([k, v]) => {
    if (v === undefined || v === null || v === '') return
    qs.set(k, String(v))
  })
  return request<any>(`/api/admin/permissions/audits?${qs.toString()}`, {
    method: 'GET',
    headers: { Authorization: `Bearer ${token}` },
  })
}

export async function getRuntimeStatus(token: string, params: Record<string, any>) {
  const qs = new URLSearchParams()
  Object.entries(params).forEach(([k, v]) => {
    if (v === undefined || v === null || v === '') return
    qs.set(k, String(v))
  })
  return request<any>(`/api/admin/runtime/status?${qs.toString()}`, {
    method: 'GET',
    headers: { Authorization: `Bearer ${token}` },
  })
}

export async function getRuntimeLogs(token: string, params: Record<string, any>) {
  const qs = new URLSearchParams()
  Object.entries(params).forEach(([k, v]) => {
    if (v === undefined || v === null || v === '') return
    qs.set(k, String(v))
  })
  return request<any>(`/api/admin/runtime/logs?${qs.toString()}`, {
    method: 'GET',
    headers: { Authorization: `Bearer ${token}` },
  })
}

export async function getObsHealth(token: string) {
  return request<any>(`/api/observability/health`, {
    method: 'GET',
    headers: { Authorization: `Bearer ${token}` },
  })
}

export async function lokiQueryRange(token: string, payload: any) {
  return request<any>(`/api/observability/loki/query_range`, {
    method: 'POST',
    headers: { Authorization: `Bearer ${token}` },
    body: JSON.stringify(payload),
  })
}

export async function chat(token: string, payload: any) {
  return request<any>(`/api/chat`, {
    method: 'POST',
    headers: { Authorization: `Bearer ${token}` },
    body: JSON.stringify(payload),
  })
}

export async function aiOps(token: string) {
  return request<any>(`/api/ai_ops`, {
    method: 'POST',
    headers: { Authorization: `Bearer ${token}` },
    body: JSON.stringify({}),
  })
}
