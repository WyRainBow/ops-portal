package admin

import (
	"context"
	"strings"
	"time"

	v1 "github.com/WyRainBow/ops-portal/api/admin/v1"
	"github.com/WyRainBow/ops-portal/internal/store"

	"github.com/gogf/gf/v2/errors/gerror"
)

func (c *ControllerV1) RequestLogs(ctx context.Context, req *v1.RequestLogsReq) (res *v1.RequestLogsRes, err error) {
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

	q := db.WithContext(ctx).Model(&store.APIRequestLog{})
	if strings.TrimSpace(req.TraceID) != "" {
		q = q.Where("trace_id = ?", strings.TrimSpace(req.TraceID))
	}
	if strings.TrimSpace(req.Path) != "" {
		q = q.Where("path ILIKE ?", "%"+strings.TrimSpace(req.Path)+"%")
	}
	if req.StatusCode != nil {
		q = q.Where("status_code = ?", *req.StatusCode)
	}
	if req.MinStatusCode != nil {
		q = q.Where("status_code >= ?", *req.MinStatusCode)
	}
	if req.MaxStatusCode != nil {
		q = q.Where("status_code <= ?", *req.MaxStatusCode)
	}

	var total int64
	if err := q.Count(&total).Error; err != nil {
		return nil, gerror.Newf("db count failed: %v", err)
	}

	var rows []store.APIRequestLog
	if err := q.Order("created_at DESC").Offset((page - 1) * pageSize).Limit(pageSize).Find(&rows).Error; err != nil {
		return nil, gerror.Newf("db query failed: %v", err)
	}

	items := make([]v1.RequestLogItem, 0, len(rows))
	for _, r := range rows {
		ip := ""
		if r.IP != nil {
			ip = *r.IP
		}
		created := ""
		if r.CreatedAt != nil {
			created = r.CreatedAt.UTC().Format(time.RFC3339Nano)
		}
		items = append(items, v1.RequestLogItem{
			ID:         r.ID,
			TraceID:    r.TraceID,
			RequestID:  r.RequestID,
			Method:     r.Method,
			Path:       r.Path,
			StatusCode: r.StatusCode,
			LatencyMs:  r.LatencyMs,
			UserID:     r.UserID,
			IP:         ip,
			CreatedAt:  created,
		})
	}

	return &v1.RequestLogsRes{
		Items:    items,
		Total:    total,
		Page:     page,
		PageSize: pageSize,
	}, nil
}

func (c *ControllerV1) ErrorLogs(ctx context.Context, req *v1.ErrorLogsReq) (res *v1.ErrorLogsRes, err error) {
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

	q := db.WithContext(ctx).Model(&store.APIErrorLog{})
	if strings.TrimSpace(req.TraceID) != "" {
		q = q.Where("trace_id = ?", strings.TrimSpace(req.TraceID))
	}
	if strings.TrimSpace(req.Keyword) != "" {
		q = q.Where("error_message ILIKE ?", "%"+strings.TrimSpace(req.Keyword)+"%")
	}

	var total int64
	if err := q.Count(&total).Error; err != nil {
		return nil, gerror.Newf("db count failed: %v", err)
	}

	var rows []store.APIErrorLog
	if err := q.Order("created_at DESC").Offset((page - 1) * pageSize).Limit(pageSize).Find(&rows).Error; err != nil {
		return nil, gerror.Newf("db query failed: %v", err)
	}

	items := make([]v1.ErrorLogItem, 0, len(rows))
	for _, r := range rows {
		t := ""
		svc := ""
		if r.ErrorType != nil {
			t = *r.ErrorType
		}
		if r.Service != nil {
			svc = *r.Service
		}
		created := ""
		if r.CreatedAt != nil {
			created = r.CreatedAt.UTC().Format(time.RFC3339Nano)
		}
		items = append(items, v1.ErrorLogItem{
			ID:           r.ID,
			RequestLogID: r.RequestLogID,
			TraceID:      r.TraceID,
			ErrorType:    t,
			ErrorMessage: r.ErrorMessage,
			Service:      svc,
			CreatedAt:    created,
		})
	}

	return &v1.ErrorLogsRes{
		Items:    items,
		Total:    total,
		Page:     page,
		PageSize: pageSize,
	}, nil
}
