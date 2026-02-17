package tools

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"os"
	"sort"
	"strconv"
	"time"

	"github.com/cloudwego/eino/components/tool"
	"github.com/cloudwego/eino/components/tool/utils"
)

// Keep the old function name used by plan_execute_replan.NewExecutor.
// Previously this returned Tencent Cloud MCP tools. For v1 we switch to Loki HTTP API.
func GetLogMcpTool() ([]tool.BaseTool, error) {
	return []tool.BaseTool{NewLokiQueryRangeTool()}, nil
}

type LokiQueryInput struct {
	Query string `json:"query" jsonschema:"description=LogQL query, e.g. {job=\"resume-backend\", stream=\"error\"} |= \"ERROR\""`
	// Unix timestamp in nanoseconds. If omitted, defaults to last 1 hour.
	Start int64 `json:"start,omitempty" jsonschema:"description=Start time (unix ns). Optional."`
	End   int64 `json:"end,omitempty" jsonschema:"description=End time (unix ns). Optional."`
	Limit int   `json:"limit,omitempty" jsonschema:"description=Max lines. Default 200, max 2000."`
}

type LokiLine struct {
	Ts   string         `json:"ts"`
	Line string         `json:"line"`
	Meta map[string]any `json:"meta,omitempty"`
}

// NewLokiQueryRangeTool queries Loki directly via HTTP.
func NewLokiQueryRangeTool() tool.InvokableTool {
	t, err := utils.InferOptionableTool(
		"query_loki_logs",
		"Query logs from Loki using LogQL and a time range. Use this tool to fetch error logs, tracebacks, or specific service logs.",
		func(ctx context.Context, input *LokiQueryInput, opts ...tool.Option) (output string, err error) {
			q := input.Query
			if q == "" {
				return `{"success":false,"error":"query is empty"}`, nil
			}
			limit := input.Limit
			if limit <= 0 {
				limit = 200
			}
			if limit > 2000 {
				limit = 2000
			}

			now := time.Now().UTC()
			start := input.Start
			end := input.End
			if end <= 0 {
				end = now.UnixNano()
			}
			if start <= 0 {
				start = now.Add(-1 * time.Hour).UnixNano()
			}
			// Be resilient to callers passing seconds/ms/us instead of ns.
			// 2026 unix ns ~ 1.7e18. We normalize smaller magnitudes up to ns.
			start = normalizeUnixToNs(start)
			end = normalizeUnixToNs(end)
			if time.Duration(end-start) > 24*time.Hour {
				start = time.Unix(0, end).Add(-24 * time.Hour).UnixNano()
			}

			base := os.Getenv("OBS_LOKI_URL")
			if base == "" {
				base = "http://127.0.0.1:3100"
			}
			u, _ := url.Parse(base + "/loki/api/v1/query_range")
			qq := u.Query()
			qq.Set("query", q)
			qq.Set("start", strconv.FormatInt(start, 10))
			qq.Set("end", strconv.FormatInt(end, 10))
			qq.Set("limit", strconv.Itoa(limit))
			u.RawQuery = qq.Encode()

			cctx, cancel := context.WithTimeout(ctx, 12*time.Second)
			defer cancel()
			req, _ := http.NewRequestWithContext(cctx, http.MethodGet, u.String(), nil)
			resp, err := http.DefaultClient.Do(req)
			if err != nil {
				return fmt.Sprintf(`{"success":false,"error":"%s"}`, escapeJSON(err.Error())), nil
			}
			defer resp.Body.Close()
			body, _ := io.ReadAll(resp.Body)
			if resp.StatusCode/100 != 2 {
				return fmt.Sprintf(`{"success":false,"status":%d,"error":"%s"}`, resp.StatusCode, escapeJSON(string(body))), nil
			}

			var raw map[string]any
			if err := json.Unmarshal(body, &raw); err != nil {
				return fmt.Sprintf(`{"success":false,"error":"parse json failed: %s"}`, escapeJSON(err.Error())), nil
			}

			lines := extractLines(raw)
			// Sort by ts asc for readability.
			sort.SliceStable(lines, func(i, j int) bool { return lines[i].Ts < lines[j].Ts })

			out := map[string]any{
				"success": true,
				"query":   q,
				"count":   len(lines),
				"lines":   lines,
			}
			b, _ := json.MarshalIndent(out, "", "  ")
			return string(b), nil
		},
	)
	if err != nil {
		// Instead of panic, return an error tool
		return createErrorLokiTool(err)
	}
	return t
}

// createErrorLokiTool returns a tool that always returns an error
func createErrorLokiTool(createErr error) tool.InvokableTool {
	t, _ := utils.InferOptionableTool(
		"query_loki_logs",
		"Error tool - Loki query tool failed to initialize",
		func(ctx context.Context, input any, opts ...tool.Option) (output string, err error) {
			return fmt.Sprintf(`{"success":false,"error":"Tool initialization failed: %s"}`, escapeJSON(createErr.Error())), nil
		},
	)
	return t
}

func normalizeUnixToNs(v int64) int64 {
	if v <= 0 {
		return v
	}
	// seconds
	if v < 1e11 {
		return v * 1e9
	}
	// milliseconds
	if v < 1e14 {
		return v * 1e6
	}
	// microseconds
	if v < 1e17 {
		return v * 1e3
	}
	// nanoseconds
	return v
}

func extractLines(raw map[string]any) []LokiLine {
	out := make([]LokiLine, 0)
	data, _ := raw["data"].(map[string]any)
	result, _ := data["result"].([]any)
	for _, it := range result {
		obj, _ := it.(map[string]any)
		stream, _ := obj["stream"].(map[string]any)
		values, _ := obj["values"].([]any)
		for _, vv := range values {
			pair, _ := vv.([]any)
			if len(pair) < 2 {
				continue
			}
			ts, _ := pair[0].(string)
			line, _ := pair[1].(string)
			out = append(out, LokiLine{Ts: ts, Line: line, Meta: stream})
		}
	}
	return out
}

func escapeJSON(s string) string {
	b, _ := json.Marshal(s)
	// b is quoted string; strip quotes
	if len(b) >= 2 {
		return string(b[1 : len(b)-1])
	}
	return s
}
