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

/**
 * Stream chat response using Server-Sent Events.
 * Returns a promise that resolves when the stream ends.
 */
export async function streamChat(
  token: string,
  payload: { question: string; id?: string },
  callbacks: {
    onMessage?: (chunk: string) => void
    onDone?: () => void
    onError?: (error: string) => void
  }
): Promise<void> {
  const response = await fetch('/api/chat/chat_stream', {
    method: 'POST',
    headers: {
      'Content-Type': 'application/json',
      'Authorization': `Bearer ${token}`,
    },
    body: JSON.stringify(payload),
  })

  if (!response.ok) {
    const error = await response.text()
    callbacks.onError?.(`HTTP ${response.status}: ${error}`)
    return
  }

  const reader = response.body?.getReader()
  const decoder = new TextDecoder()

  if (!reader) {
    callbacks.onError?.('Response body is not readable')
    return
  }

  let buffer = ''

  try {
    while (true) {
      const { done, value } = await reader.read()

      if (done) {
        callbacks.onDone?.()
        break
      }

      // Decode and process SSE data
      buffer += decoder.decode(value, { stream: true })

      // Split by lines and process complete SSE messages
      const lines = buffer.split('\n')
      buffer = lines.pop() || '' // Keep incomplete line in buffer

      for (const line of lines) {
        const trimmed = line.trim()

        // SSE format: event: xxx, data: yyy, id: zzz
        if (trimmed.startsWith('data:')) {
          const data = trimmed.slice(5).trim()

          // Parse JSON data
          try {
            // The backend sends simple string chunks, not JSON
            if (data) {
              callbacks.onMessage?.(data)
            }
          } catch {
            // If not JSON, just send as-is
            callbacks.onMessage?.(data)
          }
        }
      }
    }
  } catch (error: any) {
    callbacks.onError?.(error.message)
  } finally {
    reader.releaseLock()
  }
}

export async function aiOps(token: string) {
  return request<any>(`/api/ai_ops`, {
    method: 'POST',
    headers: { Authorization: `Bearer ${token}` },
    body: JSON.stringify({}),
  })
}

// Playbook APIs
export async function getPlaybooks(token: string) {
  return request<{ success: boolean; playbooks: any[]; count: number }>(`/api/ops/playbooks`, {
    method: 'GET',
    headers: { Authorization: `Bearer ${token}` },
  })
}

export async function getPlaybook(token: string, id: string) {
  return request<{ success: boolean; playbook: any }>(`/api/ops/playbooks/${encodeURIComponent(id)}`, {
    method: 'GET',
    headers: { Authorization: `Bearer ${token}` },
  })
}

export async function executePlaybook(
  token: string,
  playbookId: string,
  payload: { playbook_id: string; parameters?: Record<string, any>; reason: string; dry_run?: boolean }
) {
  return request<{ success: boolean; execution: any }>(`/api/ops/playbooks/${encodeURIComponent(playbookId)}/execute`, {
    method: 'POST',
    headers: { Authorization: `Bearer ${token}` },
    body: JSON.stringify(payload),
  })
}

export async function getExecution(token: string, executionId: string) {
  return request<{ success: boolean; execution: any }>(`/api/ops/executions/${encodeURIComponent(executionId)}`, {
    method: 'GET',
    headers: { Authorization: `Bearer ${token}` },
  })
}

export async function getAuditLog(token: string) {
  return request<{ success: boolean; audit_log: any[]; count: number }>(`/api/ops/audit/log`, {
    method: 'GET',
    headers: { Authorization: `Bearer ${token}` },
  })
}
