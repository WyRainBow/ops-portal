package admin

import (
	"context"
	"time"

	v1 "github.com/WyRainBow/ops-portal/api/admin/v1"
	"github.com/WyRainBow/ops-portal/internal/store"

	"github.com/gogf/gf/v2/errors/gerror"
	"gorm.io/gorm"
)

type traceAggRow struct {
	TraceID      string     `gorm:"column:trace_id"`
	LatestAt     *time.Time `gorm:"column:latest_at"`
	RequestCount int64      `gorm:"column:request_count"`
	ErrorCount   int64      `gorm:"column:error_count"`
	AvgLatencyMs float64    `gorm:"column:avg_latency_ms"`
}

func (c *ControllerV1) Traces(ctx context.Context, req *v1.TracesReq) (res *v1.TracesRes, err error) {
	if _, err := requireAdminOrMember(ctx); err != nil {
		return nil, err
	}
	db, err := store.DB(ctx)
	if err != nil {
		return nil, gerror.Newf("db init failed: %v", err)
	}

	page := req.Page
	if page <= 0 {
		page = 1
	}
	pageSize := req.PageSize
	if pageSize <= 0 {
		pageSize = 20
	}
	if pageSize > 200 {
		pageSize = 200
	}

	whereSQL := ""
	args := []any{}
	if req.TraceID != "" {
		whereSQL = "WHERE trace_id = ?"
		args = append(args, req.TraceID)
	}

	var total int64
	countSQL := "SELECT COUNT(*) FROM (SELECT 1 FROM api_request_logs " + whereSQL + " GROUP BY trace_id) t"
	if err := db.WithContext(ctx).Raw(countSQL, args...).Scan(&total).Error; err != nil {
		return nil, gerror.Newf("db count failed: %v", err)
	}

	q := db.WithContext(ctx).Table("api_request_logs").
		Select(`
			trace_id as trace_id,
			MAX(created_at) as latest_at,
			COUNT(*) as request_count,
			SUM(CASE WHEN status_code >= 500 THEN 1 ELSE 0 END) as error_count,
			COALESCE(AVG(latency_ms), 0) as avg_latency_ms
		`).
		Group("trace_id")
	if req.TraceID != "" {
		q = q.Where("trace_id = ?", req.TraceID)
	}

	var rows []traceAggRow
	if err := q.Order("MAX(created_at) DESC").Offset((page - 1) * pageSize).Limit(pageSize).Scan(&rows).Error; err != nil {
		return nil, gerror.Newf("db query failed: %v", err)
	}

	items := make([]v1.TraceListItem, 0, len(rows))
	for _, r := range rows {
		latest := ""
		if r.LatestAt != nil {
			latest = r.LatestAt.UTC().Format(time.RFC3339Nano)
		}
		items = append(items, v1.TraceListItem{
			TraceID:      r.TraceID,
			LatestAt:     latest,
			RequestCount: r.RequestCount,
			ErrorCount:   r.ErrorCount,
			AvgLatencyMs: r.AvgLatencyMs,
		})
	}

	return &v1.TracesRes{
		Items:    items,
		Total:    total,
		Page:     page,
		PageSize: pageSize,
	}, nil
}

func (c *ControllerV1) TraceDetail(ctx context.Context, req *v1.TraceDetailReq) (res *v1.TraceDetailRes, err error) {
	if _, err := requireAdminOrMember(ctx); err != nil {
		return nil, err
	}
	db, err := store.DB(ctx)
	if err != nil {
		return nil, gerror.Newf("db init failed: %v", err)
	}

	// Prefer stored spans.
	var spans []store.APITraceSpan
	if err := db.WithContext(ctx).Where("trace_id = ?", req.TraceID).Order("start_time ASC").Find(&spans).Error; err != nil {
		return nil, gerror.Newf("db query failed: %v", err)
	}
	if len(spans) > 0 {
		out := make([]v1.TraceSpanItem, 0, len(spans))
		for _, s := range spans {
			parent := ""
			if s.ParentSpanID != nil {
				parent = *s.ParentSpanID
			}
			// tags is jsonb; leave as map if possible.
			tags := map[string]any{}
			if m, ok := s.Tags.(map[string]any); ok {
				tags = m
			}
			out = append(out, v1.TraceSpanItem{
				SpanID:       s.SpanID,
				ParentSpanID: parent,
				SpanName:     s.SpanName,
				StartTime:    s.StartTime.UTC().Format(time.RFC3339Nano),
				EndTime:      s.EndTime.UTC().Format(time.RFC3339Nano),
				DurationMs:   s.DurationMs,
				Status:       s.Status,
				Tags:         tags,
			})
		}
		return &v1.TraceDetailRes{TraceID: req.TraceID, Spans: out}, nil
	}

	// Fallback: synthetic spans from requests.
	var reqRows []store.APIRequestLog
	if err := db.WithContext(ctx).Where("trace_id = ?", req.TraceID).Order("created_at ASC").Find(&reqRows).Error; err != nil {
		return nil, gerror.Newf("db query failed: %v", err)
	}
	if len(reqRows) == 0 {
		return nil, gerror.New("trace 不存在")
	}

	out := make([]v1.TraceSpanItem, 0, len(reqRows))
	for _, r := range reqRows {
		start := time.Now().UTC()
		if r.CreatedAt != nil {
			start = r.CreatedAt.UTC()
		}
		end := start.Add(time.Duration(r.LatencyMs*float64(time.Millisecond)) )
		status := "ok"
		if r.StatusCode >= 500 {
			status = "error"
		}
		tags := map[string]any{
			"status_code": r.StatusCode,
			"ip":          "",
			"user_id":     r.UserID,
		}
		if r.IP != nil {
			tags["ip"] = *r.IP
		}
		out = append(out, v1.TraceSpanItem{
			SpanID:       r.RequestID,
			ParentSpanID: "",
			SpanName:     r.Method + " " + r.Path,
			StartTime:    start.Format(time.RFC3339Nano),
			EndTime:      end.Format(time.RFC3339Nano),
			DurationMs:   r.LatencyMs,
			Status:       status,
			Tags:         tags,
		})
	}

	return &v1.TraceDetailRes{TraceID: req.TraceID, Spans: out}, nil
}

var _ = gorm.ErrRecordNotFound

