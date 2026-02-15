package observability

import (
	"context"
	"encoding/json"
	"io"
	"net/http"
	"net/url"
	"os"
	"strconv"
	"time"

	v1 "github.com/WyRainBow/ops-portal/api/observability/v1"
	"github.com/gogf/gf/v2/errors/gerror"
)

func (c *ControllerV1) LokiQueryRange(ctx context.Context, req *v1.LokiQueryRangeReq) (res *v1.LokiQueryRangeRes, err error) {
	if err := requireAdminOrMember(ctx); err != nil {
		return nil, err
	}
	if req.Query == "" {
		return nil, gerror.New("query 不能为空")
	}

	limit := req.Limit
	if limit <= 0 {
		limit = 200
	}
	if limit > 2000 {
		limit = 2000
	}

	now := time.Now().UTC()
	start := req.Start
	end := req.End
	// Prefer ms fields (safer for JS/JSON), convert to ns.
	if req.StartMs > 0 {
		start = req.StartMs * int64(time.Millisecond)
	}
	if req.EndMs > 0 {
		end = req.EndMs * int64(time.Millisecond)
	}
	if end <= 0 {
		end = now.UnixNano()
	}
	if start <= 0 {
		start = now.Add(-1 * time.Hour).UnixNano()
	}
	// cap to 24h
	if time.Duration(end-start) > 24*time.Hour {
		start = time.Unix(0, end).Add(-24 * time.Hour).UnixNano()
	}

	base := os.Getenv("OBS_LOKI_URL")
	if base == "" {
		base = "http://127.0.0.1:3100"
	}
	u, _ := url.Parse(base + "/loki/api/v1/query_range")
	q := u.Query()
	q.Set("query", req.Query)
	q.Set("start", strconv.FormatInt(start, 10))
	q.Set("end", strconv.FormatInt(end, 10))
	q.Set("limit", strconv.Itoa(limit))
	u.RawQuery = q.Encode()

	cctx, cancel := context.WithTimeout(ctx, 10*time.Second)
	defer cancel()
	httpReq, _ := http.NewRequestWithContext(cctx, http.MethodGet, u.String(), nil)
	resp, err := http.DefaultClient.Do(httpReq)
	if err != nil {
		return nil, gerror.Newf("loki request failed: %v", err)
	}
	defer resp.Body.Close()
	body, _ := io.ReadAll(resp.Body)
	if resp.StatusCode/100 != 2 {
		return nil, gerror.Newf("loki status=%d body=%s", resp.StatusCode, string(body))
	}

	// Parse Loki response (streams -> values).
	var raw map[string]any
	if err := json.Unmarshal(body, &raw); err != nil {
		return nil, gerror.Newf("loki json parse failed: %v", err)
	}
	lines := extractLokiLines(raw)
	return &v1.LokiQueryRangeRes{
		Query: req.Query,
		Lines: lines,
	}, nil
}

func extractLokiLines(raw map[string]any) []v1.LokiLine {
	out := make([]v1.LokiLine, 0)
	data, _ := raw["data"].(map[string]any)
	result, _ := data["result"].([]any)
	for _, it := range result {
		streamObj, _ := it.(map[string]any)
		labels, _ := streamObj["stream"].(map[string]any)
		values, _ := streamObj["values"].([]any)
		for _, vv := range values {
			pair, _ := vv.([]any)
			if len(pair) < 2 {
				continue
			}
			ts, _ := pair[0].(string)
			line, _ := pair[1].(string)
			out = append(out, v1.LokiLine{
				Ts:     ts,
				Line:   line,
				Labels: labels,
			})
		}
	}
	return out
}
