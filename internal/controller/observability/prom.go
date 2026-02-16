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

func (c *ControllerV1) PromQuery(ctx context.Context, req *v1.PromQueryReq) (res *v1.PromQueryRes, err error) {
	if err := requireAdminOrMember(ctx); err != nil {
		return nil, err
	}
	if req.Query == "" {
		return nil, gerror.New("query 不能为空")
	}

	base := os.Getenv("OBS_PROM_URL")
	if base == "" {
		base = "http://127.0.0.1:9090"
	}
	u, _ := url.Parse(base + "/api/v1/query")
	q := u.Query()
	q.Set("query", req.Query)
	if req.Time > 0 {
		q.Set("time", strconv.FormatInt(req.Time, 10))
	}
	u.RawQuery = q.Encode()

	cctx, cancel := context.WithTimeout(ctx, 10*time.Second)
	defer cancel()
	httpReq, _ := http.NewRequestWithContext(cctx, http.MethodGet, u.String(), nil)
	resp, err := http.DefaultClient.Do(httpReq)
	if err != nil {
		return nil, gerror.Newf("prom request failed: %v", err)
	}
	defer resp.Body.Close()
	body, _ := io.ReadAll(resp.Body)
	if resp.StatusCode/100 != 2 {
		return nil, gerror.Newf("prom status=%d body=%s", resp.StatusCode, string(body))
	}

	var raw map[string]any
	if err := json.Unmarshal(body, &raw); err != nil {
		return nil, gerror.Newf("prom json parse failed: %v", err)
	}
	return &v1.PromQueryRes{
		Query:  req.Query,
		Result: raw,
	}, nil
}

