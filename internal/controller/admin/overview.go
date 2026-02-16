package admin

import (
	"context"
	"time"

	v1 "github.com/WyRainBow/ops-portal/api/admin/v1"
	"github.com/WyRainBow/ops-portal/internal/store"

	"github.com/gogf/gf/v2/errors/gerror"
)

func (c *ControllerV1) Overview(ctx context.Context, req *v1.OverviewReq) (res *v1.OverviewRes, err error) {
	if _, err := requireAdminOrMember(ctx); err != nil {
		return nil, err
	}

	db, err := store.DB(ctx)
	if err != nil {
		return nil, gerror.Newf("db init failed: %v", err)
	}

	since := time.Now().UTC().Add(-24 * time.Hour)
	var totalUsers int64
	var totalMembers int64
	var requests24h int64
	var errors24h int64

	if err := db.WithContext(ctx).Model(&store.User{}).Count(&totalUsers).Error; err != nil {
		return nil, gerror.Newf("db query failed: %v", err)
	}
	if err := db.WithContext(ctx).Model(&store.Member{}).Count(&totalMembers).Error; err != nil {
		return nil, gerror.Newf("db query failed: %v", err)
	}
	if err := db.WithContext(ctx).Model(&store.APIRequestLog{}).Where("created_at >= ?", since).Count(&requests24h).Error; err != nil {
		return nil, gerror.Newf("db query failed: %v", err)
	}
	if err := db.WithContext(ctx).Model(&store.APIErrorLog{}).Where("created_at >= ?", since).Count(&errors24h).Error; err != nil {
		return nil, gerror.Newf("db query failed: %v", err)
	}

	type avgRow struct{ Avg float64 }
	var avg avgRow
	_ = db.WithContext(ctx).
		Model(&store.APIRequestLog{}).
		Select("COALESCE(AVG(latency_ms), 0) as avg").
		Where("created_at >= ?", since).
		Scan(&avg).Error

	rate := 0.0
	if requests24h > 0 {
		rate = float64(errors24h) / float64(requests24h) * 100.0
	}

	return &v1.OverviewRes{
		TotalUsers:      totalUsers,
		TotalMembers:    totalMembers,
		Requests24H:     requests24h,
		Errors24H:       errors24h,
		ErrorRate24H:    round2(rate),
		AvgLatencyMs24H: round2(avg.Avg),
	}, nil
}

func round2(v float64) float64 {
	// avoid extra deps
	x := float64(int64(v*100+0.5)) / 100
	return x
}

