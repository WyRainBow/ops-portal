package admin

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"sort"
	"strings"
	"time"

	"github.com/WyRainBow/ops-portal/api/admin/v1"

	"github.com/gogf/gf/v2/errors/gerror"
	"github.com/gogf/gf/v2/frame/g"
)

type openAPISpec struct {
	Paths map[string]map[string]openAPIOperation `json:"paths"`
}

type openAPIOperation struct {
	Summary     string   `json:"summary"`
	OperationID string   `json:"operationId"`
	Tags        []string `json:"tags"`
	Deprecated  bool     `json:"deprecated"`
}

func (c *ControllerV1) ApiRoutes(ctx context.Context, req *v1.ApiRoutesReq) (res *v1.ApiRoutesRes, err error) {
	_, err = requireAdminOrMember(ctx)
	if err != nil {
		return nil, err
	}

	hideDocs := req.HideDocs
	if !req.HideDocs && req.Query == "" && req.Tag == "" && req.Method == "" {
		// Backwards-compatible default when query params are omitted.
		hideDocs = true
	}

	var items []v1.ApiRouteItem
	localSpec, err := fetchLocalOpenAPI(ctx)
	if err != nil {
		return nil, err
	}
	items = append(items, specToItems(localSpec, "ops-portal", hideDocs)...)

	if resumeSpec, err := fetchResumeOpenAPI(ctx); err == nil {
		items = append(items, specToItems(resumeSpec, "resume-agent", hideDocs)...)
	} else {
		g.Log().Warning(ctx, "resume-agent openapi unavailable:", err)
	}

	items = filterApiRoutes(items, req)

	sort.Slice(items, func(i, j int) bool {
		ai := primaryTag(items[i].Tags)
		aj := primaryTag(items[j].Tags)
		if ai != aj {
			return ai < aj
		}
		if items[i].Path != items[j].Path {
			return items[i].Path < items[j].Path
		}
		return items[i].Method < items[j].Method
	})

	methods := map[string]int64{}
	tags := map[string]int64{}
	sources := map[string]int64{}
	for _, it := range items {
		methods[it.Method]++
		tags[primaryTag(it.Tags)]++
		sources[it.Source]++
	}

	return &v1.ApiRoutesRes{
		Total:   int64(len(items)),
		Methods: methods,
		Tags:    tags,
		Sources: sources,
		Items:   items,
	}, nil
}

func filterApiRoutes(items []v1.ApiRouteItem, req *v1.ApiRoutesReq) []v1.ApiRouteItem {
	q := strings.ToLower(strings.TrimSpace(req.Query))
	tag := strings.ToLower(strings.TrimSpace(req.Tag))
	method := strings.ToUpper(strings.TrimSpace(req.Method))
	source := strings.ToLower(strings.TrimSpace(req.Source))

	if q == "" && tag == "" && method == "" && source == "" {
		return items
	}
	out := make([]v1.ApiRouteItem, 0, len(items))
	for _, it := range items {
		if method != "" && it.Method != method {
			continue
		}
		if source != "" && strings.ToLower(strings.TrimSpace(it.Source)) != source {
			continue
		}
		if tag != "" {
			ok := false
			for _, t := range it.Tags {
				if strings.ToLower(t) == tag {
					ok = true
					break
				}
			}
			if !ok {
				continue
			}
		}
		if q != "" {
			hay := strings.ToLower(it.Method + " " + it.Path + " " + it.Summary + " " + strings.Join(it.Tags, " "))
			if !strings.Contains(hay, q) {
				continue
			}
		}
		out = append(out, it)
	}
	return out
}

func primaryTag(tags []string) string {
	if len(tags) > 0 && strings.TrimSpace(tags[0]) != "" {
		return tags[0]
	}
	return "_"
}

func isDocsPath(p string) bool {
	if p == "" {
		return false
	}
	if p == "/api.json" || p == "/swagger" || p == "/openapi.json" || p == "/docs" || p == "/redoc" {
		return true
	}
	return strings.HasPrefix(p, "/swagger/") || strings.HasPrefix(p, "/docs")
}

func fetchLocalOpenAPI(ctx context.Context) (*openAPISpec, error) {
	// We intentionally fetch from the local OpenAPI endpoint to avoid coupling
	// to internal GoFrame server APIs.
	port := 18081
	if v := strings.TrimSpace(os.Getenv("OPS_PORTAL_API_PORT")); v != "" {
		if n, err := parsePort(v); err == nil {
			port = n
		}
	}
	url := fmt.Sprintf("http://127.0.0.1:%d/api.json", port)
	return fetchOpenAPIFromURL(ctx, url)
}

func fetchResumeOpenAPI(ctx context.Context) (*openAPISpec, error) {
	url := strings.TrimSpace(os.Getenv("OPS_PORTAL_RESUME_OPENAPI_URL"))
	if url == "" {
		url = "http://127.0.0.1:9000/openapi.json"
	}
	return fetchOpenAPIFromURL(ctx, url)
}

func fetchOpenAPIFromURL(ctx context.Context, url string) (*openAPISpec, error) {
	client := &http.Client{Timeout: 4 * time.Second}
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		return nil, gerror.Wrap(err, "create request failed")
	}
	resp, err := client.Do(req)
	if err != nil {
		return nil, gerror.Wrap(err, "fetch openapi failed")
	}
	defer resp.Body.Close()
	if resp.StatusCode/100 != 2 {
		b, _ := io.ReadAll(io.LimitReader(resp.Body, 64*1024))
		return nil, gerror.Newf("fetch openapi failed: HTTP %d: %s", resp.StatusCode, strings.TrimSpace(string(b)))
	}

	body, err := io.ReadAll(io.LimitReader(resp.Body, 4*1024*1024))
	if err != nil {
		return nil, gerror.Wrap(err, "read openapi body failed")
	}
	var spec openAPISpec
	if err := json.Unmarshal(body, &spec); err != nil {
		// g.Dump can help debugging in dev, but don't leak full body in prod responses.
		g.Log().Debug(ctx, "openapi json unmarshal failed", err)
		return nil, gerror.Wrap(err, "parse openapi failed")
	}
	if spec.Paths == nil {
		return nil, gerror.New("openapi spec has no paths")
	}
	return &spec, nil
}

func specToItems(spec *openAPISpec, source string, hideDocs bool) []v1.ApiRouteItem {
	items := make([]v1.ApiRouteItem, 0, len(spec.Paths)*2)
	for p, methods := range spec.Paths {
		if hideDocs && isDocsPath(p) {
			continue
		}
		for m, op := range methods {
			method := strings.ToUpper(strings.TrimSpace(m))
			if method == "" || strings.HasPrefix(strings.ToLower(m), "x-") {
				continue
			}
			if method == "PARAMETERS" {
				continue
			}
			items = append(items, v1.ApiRouteItem{
				Method:      method,
				Path:        p,
				Summary:     strings.TrimSpace(op.Summary),
				OperationID: strings.TrimSpace(op.OperationID),
				Tags:        op.Tags,
				Deprecated:  op.Deprecated,
				Source:      source,
			})
		}
	}
	return items
}

func parsePort(s string) (int, error) {
	n := 0
	for _, ch := range s {
		if ch < '0' || ch > '9' {
			return 0, gerror.New("invalid port")
		}
		n = n*10 + int(ch-'0')
		if n > 65535 {
			return 0, gerror.New("invalid port")
		}
	}
	if n <= 0 {
		return 0, gerror.New("invalid port")
	}
	return n, nil
}
