package store

import (
	"time"
)

type User struct {
	ID          int64      `gorm:"column:id;primaryKey"`
	Username    string     `gorm:"column:username"`
	Email       *string    `gorm:"column:email"`
	PasswordHash string    `gorm:"column:password_hash"`
	Role        string     `gorm:"column:role"`
	LastLoginIP *string    `gorm:"column:last_login_ip"`
	APIQuota    *int64     `gorm:"column:api_quota"`
	CreatedAt   *time.Time `gorm:"column:created_at"`
	UpdatedAt   *time.Time `gorm:"column:updated_at"`
}

func (User) TableName() string { return "users" }

type Member struct {
	ID        int64      `gorm:"column:id;primaryKey"`
	Name      string     `gorm:"column:name"`
	Email     *string    `gorm:"column:email"`
	Position  *string    `gorm:"column:position"`
	Team      *string    `gorm:"column:team"`
	Status    string     `gorm:"column:status"`
	UserID    *int64     `gorm:"column:user_id"`
	CreatedAt *time.Time `gorm:"column:created_at"`
	UpdatedAt *time.Time `gorm:"column:updated_at"`
}

func (Member) TableName() string { return "members" }

type APIRequestLog struct {
	ID        int64      `gorm:"column:id;primaryKey"`
	TraceID   string     `gorm:"column:trace_id"`
	RequestID string     `gorm:"column:request_id"`
	Method    string     `gorm:"column:method"`
	Path      string     `gorm:"column:path"`
	StatusCode int64     `gorm:"column:status_code"`
	LatencyMs float64    `gorm:"column:latency_ms"`
	UserID    *int64     `gorm:"column:user_id"`
	IP        *string    `gorm:"column:ip"`
	CreatedAt *time.Time `gorm:"column:created_at"`
}

func (APIRequestLog) TableName() string { return "api_request_logs" }

type APIErrorLog struct {
	ID          int64      `gorm:"column:id;primaryKey"`
	RequestLogID *int64    `gorm:"column:request_log_id"`
	TraceID      string    `gorm:"column:trace_id"`
	ErrorType    *string   `gorm:"column:error_type"`
	ErrorMessage string    `gorm:"column:error_message"`
	Service      *string   `gorm:"column:service"`
	CreatedAt    *time.Time `gorm:"column:created_at"`
}

func (APIErrorLog) TableName() string { return "api_error_logs" }

type APITraceSpan struct {
	ID          int64       `gorm:"column:id;primaryKey"`
	TraceID     string      `gorm:"column:trace_id"`
	SpanID      string      `gorm:"column:span_id"`
	ParentSpanID *string    `gorm:"column:parent_span_id"`
	SpanName    string      `gorm:"column:span_name"`
	StartTime   time.Time   `gorm:"column:start_time"`
	EndTime     time.Time   `gorm:"column:end_time"`
	DurationMs  float64     `gorm:"column:duration_ms"`
	Status      string      `gorm:"column:status"`
	Tags        any         `gorm:"column:tags"`
	CreatedAt   *time.Time  `gorm:"column:created_at"`
}

func (APITraceSpan) TableName() string { return "api_trace_spans" }

type PermissionAuditLog struct {
	ID             int64      `gorm:"column:id;primaryKey"`
	OperatorUserID *int64     `gorm:"column:operator_user_id"`
	TargetUserID   *int64     `gorm:"column:target_user_id"`
	FromRole       *string    `gorm:"column:from_role"`
	ToRole         *string    `gorm:"column:to_role"`
	Action         string     `gorm:"column:action"`
	CreatedAt      *time.Time `gorm:"column:created_at"`
}

func (PermissionAuditLog) TableName() string { return "permission_audit_logs" }
