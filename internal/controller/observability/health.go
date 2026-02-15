package observability

import (
	"context"
	"fmt"
	"io"
	"net/http"
	"os"
	"time"

	v1 "github.com/WyRainBow/ops-portal/api/observability/v1"
)

func (c *ControllerV1) Health(ctx context.Context, req *v1.HealthReq) (res *v1.HealthRes, err error) {
	if err := requireAdminOrMember(ctx); err != nil {
		return nil, err
	}

	grafana := getenv("OBS_GRAFANA_URL", "http://127.0.0.1:3000")
	loki := getenv("OBS_LOKI_URL", "http://127.0.0.1:3100")
	prom := getenv("OBS_PROM_URL", "http://127.0.0.1:9090")
	node := getenv("OBS_NODE_EXPORTER_URL", "http://127.0.0.1:9100")

	components := map[string]v1.ComponentHealth{
		"grafana":      probe(ctx, grafana+"/api/health"),
		"loki":         probe(ctx, loki+"/ready"),
		"prometheus":   probe(ctx, prom+"/-/ready"),
		"node_exporter": probe(ctx, node+"/metrics"),
	}

	host := getenv("OPS_PORTAL_SSH_HOST", "106.53.113.137")
	user := getenv("OPS_PORTAL_SSH_USER", "root")
	port := getenv("OPS_PORTAL_SSH_PORT", "2222")
	key := getenv("OPS_PORTAL_SSH_KEY", "~/.ssh/id_rsa")
	tunnels := map[string]string{
		"portal":  fmt.Sprintf("ssh -i %s -p %s -L 18080:127.0.0.1:18080 %s@%s", key, port, user, host),
		"grafana": fmt.Sprintf("ssh -i %s -p %s -L 3000:127.0.0.1:3000 %s@%s", key, port, user, host),
		"loki":    fmt.Sprintf("ssh -i %s -p %s -L 3100:127.0.0.1:3100 %s@%s", key, port, user, host),
		"prom":    fmt.Sprintf("ssh -i %s -p %s -L 9090:127.0.0.1:9090 %s@%s", key, port, user, host),
	}

	return &v1.HealthRes{
		ServerTimeUTC: time.Now().UTC().Format(time.RFC3339Nano),
		Components:    components,
		Tunnels:       tunnels,
	}, nil
}

func probe(ctx context.Context, u string) v1.ComponentHealth {
	h := v1.ComponentHealth{URL: u}
	t0 := time.Now()
	cctx, cancel := context.WithTimeout(ctx, 6*time.Second)
	defer cancel()
	req, err := http.NewRequestWithContext(cctx, http.MethodGet, u, nil)
	if err != nil {
		h.Error = err.Error()
		return h
	}
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		h.Error = err.Error()
		return h
	}
	defer resp.Body.Close()
	_, _ = io.Copy(io.Discard, resp.Body)
	h.Status = resp.StatusCode
	h.LatencyMs = float64(time.Since(t0).Milliseconds())
	h.OK = resp.StatusCode >= 200 && resp.StatusCode < 300
	return h
}

func getenv(k, def string) string {
	if v := os.Getenv(k); v != "" {
		return v
	}
	return def
}
