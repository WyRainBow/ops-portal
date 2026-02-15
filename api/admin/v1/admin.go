package v1

import "github.com/gogf/gf/v2/frame/g"

// =================
// Overview
// =================

type OverviewReq struct {
	g.Meta `path:"/admin/overview" method:"get" summary:"概览指标"`
}

type OverviewRes struct {
	TotalUsers      int64   `json:"total_users"`
	TotalMembers    int64   `json:"total_members"`
	Requests24H     int64   `json:"requests_24h"`
	Errors24H       int64   `json:"errors_24h"`
	ErrorRate24H    float64 `json:"error_rate_24h"`
	AvgLatencyMs24H float64 `json:"avg_latency_ms_24h"`
}

// =================
// API Catalog
// =================

type ApiRoutesReq struct {
	g.Meta `path:"/admin/api/routes" method:"get" summary:"接口清单（类似 Swagger）"`

	// Free text query (matches path/summary/tags).
	Query  string `json:"q" in:"query"`
	Tag    string `json:"tag" in:"query"`
	Method string `json:"method" in:"query"`

	// Default true: hide docs endpoints like /swagger, /api.json.
	HideDocs bool `json:"hide_docs" in:"query"`
}

type ApiRouteItem struct {
	Method      string   `json:"method"`
	Path        string   `json:"path"`
	Summary     string   `json:"summary,omitempty"`
	OperationID string   `json:"operation_id,omitempty"`
	Tags        []string `json:"tags,omitempty"`
	Deprecated  bool     `json:"deprecated,omitempty"`
}

type ApiRoutesRes struct {
	Total   int64            `json:"total"`
	Methods map[string]int64 `json:"methods"`
	Tags    map[string]int64 `json:"tags"`
	Items   []ApiRouteItem   `json:"items"`
}

// =================
// Users
// =================

type UsersReq struct {
	g.Meta   `path:"/admin/users" method:"get" summary:"用户列表"`
	Page     int `json:"page" in:"query"`
	PageSize int `json:"page_size" in:"query"`

	Keyword   string `json:"keyword" in:"query"`
	Role      string `json:"role" in:"query"`
	IP        string `json:"ip" in:"query"`
	WithTotal bool   `json:"with_total" in:"query"`
}

type UserItem struct {
	ID          int64  `json:"id"`
	Username    string `json:"username"`
	Email       string `json:"email,omitempty"`
	Role        string `json:"role"`
	LastLoginIP string `json:"last_login_ip,omitempty"`
	APIQuota    *int64 `json:"api_quota,omitempty"`
	CreatedAt   string `json:"created_at,omitempty"`
	UpdatedAt   string `json:"updated_at,omitempty"`
}

type UsersRes struct {
	Items    []UserItem `json:"items"`
	Total    int64      `json:"total"`
	Page     int        `json:"page"`
	PageSize int        `json:"page_size"`
}

type UpdateUserRoleReq struct {
	g.Meta `path:"/admin/users/{userId}/role" method:"patch" summary:"修改用户角色"`
	UserID int64  `json:"userId" in:"path"`
	Role   string `json:"role"`
}

type UpdateUserQuotaReq struct {
	g.Meta   `path:"/admin/users/{userId}/quota" method:"patch" summary:"修改用户额度"`
	UserID   int64  `json:"userId" in:"path"`
	APIQuota *int64 `json:"api_quota"`
}

// GoFrame expects response structs to end with "Res".
// We keep the response JSON shape by aliasing to the item structs.
type UpdateUserRoleRes = UserItem
type UpdateUserQuotaRes = UserItem

// =================
// Members
// =================

type MembersReq struct {
	g.Meta   `path:"/admin/members" method:"get" summary:"成员列表"`
	Page     int    `json:"page" in:"query"`
	PageSize int    `json:"page_size" in:"query"`
	Keyword  string `json:"keyword" in:"query"`
}

type MemberItem struct {
	ID       int64  `json:"id"`
	Name     string `json:"name"`
	Username string `json:"username,omitempty"`
	Position string `json:"position,omitempty"`
	Team     string `json:"team,omitempty"`
	Status   string `json:"status"`
	UserID   *int64 `json:"user_id,omitempty"`
	UserRole string `json:"user_role,omitempty"`
	CreatedAt string `json:"created_at,omitempty"`
	UpdatedAt string `json:"updated_at,omitempty"`
}

type MembersRes struct {
	Items    []MemberItem `json:"items"`
	Total    int64        `json:"total"`
	Page     int          `json:"page"`
	PageSize int          `json:"page_size"`
}

type CreateMemberReq struct {
	g.Meta    `path:"/admin/members" method:"post" summary:"创建成员"`
	UserID    int64  `json:"user_id"`
	Position  string `json:"position,omitempty"`
	Team      string `json:"team,omitempty"`
	Status    string `json:"status,omitempty"`
	UserRole  string `json:"user_role,omitempty"`
}

type CreateMemberRes = MemberItem

type UpdateMemberReq struct {
	g.Meta    `path:"/admin/members/{memberId}" method:"patch" summary:"更新成员"`
	MemberID  int64  `json:"memberId" in:"path"`
	UserID    int64  `json:"user_id"`
	Position  string `json:"position,omitempty"`
	Team      string `json:"team,omitempty"`
	Status    string `json:"status,omitempty"`
	UserRole  string `json:"user_role,omitempty"`
}

type UpdateMemberRes = MemberItem

type DeleteMemberReq struct {
	g.Meta   `path:"/admin/members/{memberId}" method:"delete" summary:"删除成员"`
	MemberID int64 `json:"memberId" in:"path"`
}

type DeleteMemberRes struct {
	Success bool `json:"success"`
}

// =================
// Permissions
// =================

type PermissionRolesReq struct {
	g.Meta `path:"/admin/permissions/roles" method:"get" summary:"角色权限矩阵"`
}

type PermissionRolesRes struct {
	Roles map[string]any `json:"roles"`
}

type PermissionAuditsReq struct {
	g.Meta   `path:"/admin/permissions/audits" method:"get" summary:"权限审计"`
	Page     int `json:"page" in:"query"`
	PageSize int `json:"page_size" in:"query"`
}

type PermissionAuditItem struct {
	ID              int64  `json:"id"`
	OperatorUserID  *int64 `json:"operator_user_id,omitempty"`
	TargetUserID    *int64 `json:"target_user_id,omitempty"`
	OperatorUsername string `json:"operator_username,omitempty"`
	TargetUsername   string `json:"target_username,omitempty"`
	FromRole        string `json:"from_role,omitempty"`
	ToRole          string `json:"to_role,omitempty"`
	Action          string `json:"action"`
	CreatedAt       string `json:"created_at,omitempty"`
}

type PermissionAuditsRes struct {
	Items    []PermissionAuditItem `json:"items"`
	Total    int64                 `json:"total"`
	Page     int                   `json:"page"`
	PageSize int                   `json:"page_size"`
}

// =================
// Logs
// =================

type RequestLogsReq struct {
	g.Meta     `path:"/admin/logs/requests" method:"get" summary:"请求日志"`
	Page       int    `json:"page" in:"query"`
	PageSize   int    `json:"page_size" in:"query"`
	TraceID    string `json:"trace_id" in:"query"`
	Path       string `json:"path" in:"query"`
	StatusCode *int64 `json:"status_code" in:"query"`
	MinStatusCode *int64 `json:"min_status_code" in:"query"`
	MaxStatusCode *int64 `json:"max_status_code" in:"query"`
}

type RequestLogItem struct {
	ID        int64   `json:"id"`
	TraceID   string  `json:"trace_id"`
	RequestID string  `json:"request_id"`
	Method    string  `json:"method"`
	Path      string  `json:"path"`
	StatusCode int64  `json:"status_code"`
	LatencyMs float64 `json:"latency_ms"`
	UserID    *int64  `json:"user_id,omitempty"`
	IP        string  `json:"ip,omitempty"`
	CreatedAt string  `json:"created_at,omitempty"`
}

type RequestLogsRes struct {
	Items    []RequestLogItem `json:"items"`
	Total    int64            `json:"total"`
	Page     int              `json:"page"`
	PageSize int              `json:"page_size"`
}

type ErrorLogsReq struct {
	g.Meta   `path:"/admin/logs/errors" method:"get" summary:"错误日志"`
	Page     int    `json:"page" in:"query"`
	PageSize int    `json:"page_size" in:"query"`
	TraceID  string `json:"trace_id" in:"query"`
	Keyword  string `json:"keyword" in:"query"`
}

type ErrorLogItem struct {
	ID           int64  `json:"id"`
	RequestLogID *int64 `json:"request_log_id,omitempty"`
	TraceID      string `json:"trace_id"`
	ErrorType    string `json:"error_type,omitempty"`
	ErrorMessage string `json:"error_message"`
	Service      string `json:"service,omitempty"`
	CreatedAt    string `json:"created_at,omitempty"`
}

type ErrorLogsRes struct {
	Items    []ErrorLogItem `json:"items"`
	Total    int64          `json:"total"`
	Page     int            `json:"page"`
	PageSize int            `json:"page_size"`
}

// =================
// Traces
// =================

type TracesReq struct {
	g.Meta   `path:"/admin/traces" method:"get" summary:"Trace 列表"`
	Page     int    `json:"page" in:"query"`
	PageSize int    `json:"page_size" in:"query"`
	TraceID  string `json:"trace_id" in:"query"`
}

type TraceListItem struct {
	TraceID      string  `json:"trace_id"`
	LatestAt     string  `json:"latest_at,omitempty"`
	RequestCount int64   `json:"request_count"`
	ErrorCount   int64   `json:"error_count"`
	AvgLatencyMs float64 `json:"avg_latency_ms"`
}

type TracesRes struct {
	Items    []TraceListItem `json:"items"`
	Total    int64           `json:"total"`
	Page     int             `json:"page"`
	PageSize int             `json:"page_size"`
}

type TraceDetailReq struct {
	g.Meta   `path:"/admin/traces/{traceId}" method:"get" summary:"Trace 详情"`
	TraceID  string `json:"traceId" in:"path"`
}

type TraceSpanItem struct {
	SpanID       string         `json:"span_id"`
	ParentSpanID string         `json:"parent_span_id,omitempty"`
	SpanName     string         `json:"span_name"`
	StartTime    string         `json:"start_time"`
	EndTime      string         `json:"end_time"`
	DurationMs   float64        `json:"duration_ms"`
	Status       string         `json:"status"`
	Tags         map[string]any `json:"tags,omitempty"`
}

type TraceDetailRes struct {
	TraceID string          `json:"trace_id"`
	Spans   []TraceSpanItem `json:"spans"`
}

// =================
// Runtime (read-only)
// =================

type RuntimeStatusReq struct {
	g.Meta  `path:"/admin/runtime/status" method:"get" summary:"运行状态"`
	Service string `json:"service" in:"query"`
}

type RuntimeStatusRes struct {
	ServerTimeUTC string         `json:"server_time_utc"`
	Database      map[string]any `json:"database"`
	Git           map[string]any `json:"git"`
	System        map[string]any `json:"system"`
	PM2           map[string]any `json:"pm2"`
	Service       map[string]any `json:"service"`
	Logs          map[string]any `json:"logs"`
}

type RuntimeLogsReq struct {
	g.Meta  `path:"/admin/runtime/logs" method:"get" summary:"运行日志"`
	Service string `json:"service" in:"query"`
	Stream  string `json:"stream" in:"query"` // out|error
	Lines   int    `json:"lines" in:"query"`
}

type RuntimeLogsRes struct {
	Service string `json:"service"`
	Stream  string `json:"stream"`
	Lines   int    `json:"lines"`
	Path    string `json:"path"`
	Content string `json:"content"`
}
