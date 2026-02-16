package v1

import "github.com/gogf/gf/v2/frame/g"

type HealthReq struct {
	g.Meta `path:"/observability/health" method:"get" summary:"可观测组件健康检查"`
}

type ComponentHealth struct {
	OK        bool    `json:"ok"`
	Status    int     `json:"status_code"`
	LatencyMs float64 `json:"latency_ms"`
	Error     string  `json:"error,omitempty"`
	URL       string  `json:"url,omitempty"`
}

type HealthRes struct {
	ServerTimeUTC string                    `json:"server_time_utc"`
	Components    map[string]ComponentHealth `json:"components"`
	Tunnels       map[string]string          `json:"tunnels"`
}

type LokiQueryRangeReq struct {
	g.Meta `path:"/observability/loki/query_range" method:"post" summary:"查询 Loki 日志（range）"`
	Query  string `json:"query"`
	// Prefer milliseconds to avoid JS number precision issues; server will convert to ns.
	StartMs int64 `json:"start_ms,omitempty"` // unix ms; optional
	EndMs   int64 `json:"end_ms,omitempty"`   // unix ms; optional
	// Legacy: unix ns; optional. Keep for compatibility with tools/curl.
	Start int64 `json:"start,omitempty"`
	End   int64 `json:"end,omitempty"`
	Limit  int    `json:"limit,omitempty"`
}

type LokiLine struct {
	Ts     string         `json:"ts"`
	Line   string         `json:"line"`
	Labels map[string]any `json:"labels,omitempty"`
}

type LokiQueryRangeRes struct {
	Query string     `json:"query"`
	Lines []LokiLine `json:"lines"`
}

type PromQueryReq struct {
	g.Meta `path:"/observability/prom/query" method:"post" summary:"Prometheus instant query"`
	Query  string `json:"query"`
	Time   int64  `json:"time,omitempty"` // unix seconds; optional
}

type PromQueryRes struct {
	Query  string         `json:"query"`
	Result map[string]any `json:"result"`
}
