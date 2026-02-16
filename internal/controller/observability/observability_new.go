package observability

import "github.com/WyRainBow/ops-portal/api/observability"

type ControllerV1 struct{}

func NewV1() observability.IObservabilityV1 {
	return &ControllerV1{}
}

