package observability

import (
	"context"

	"github.com/WyRainBow/ops-portal/api/observability/v1"
)

type IObservabilityV1 interface {
	Health(ctx context.Context, req *v1.HealthReq) (res *v1.HealthRes, err error)
	LokiQueryRange(ctx context.Context, req *v1.LokiQueryRangeReq) (res *v1.LokiQueryRangeRes, err error)
	PromQuery(ctx context.Context, req *v1.PromQueryReq) (res *v1.PromQueryRes, err error)
}

