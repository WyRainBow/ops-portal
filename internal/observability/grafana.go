package obsgrafana

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"os"
	"strings"
	"time"
)

// Client is a Grafana client.
type Client struct {
	baseURL    string
	apiKey     string
	httpClient *http.Client
}

// NewClient creates a new Grafana client.
func NewClient(baseURL, apiKey string) *Client {
	return &Client{
		baseURL: strings.TrimSuffix(baseURL, "/"),
		apiKey:  apiKey,
		httpClient: &http.Client{
			Timeout: 30 * time.Second,
		},
	}
}

// DefaultClient returns the default Grafana client from environment.
func DefaultClient() *Client {
	baseURL := os.Getenv("OBS_GRAFANA_URL")
	if baseURL == "" {
		baseURL = "http://127.0.0.1:3000"
	}
	apiKey := os.Getenv("GRAFANA_API_KEY")
	return NewClient(baseURL, apiKey)
}

// Dashboard represents a Grafana dashboard.
type Dashboard struct {
	ID     int    `json:"id"`
	UID    string `json:"uid"`
	Title  string `json:"title"`
	Tags   []string `json:"tags"`
	URL    string `json:"url"`
}

// SearchDashboards searches for dashboards.
func (c *Client) SearchDashboards(ctx context.Context, query string) ([]Dashboard, error) {
	u, _ := url.Parse(c.baseURL + "/api/search")
	q := u.Query()
	q.Set("type", "dash-db")
	if query != "" {
		q.Set("query", query)
	}
	u.RawQuery = q.Encode()

	req, _ := http.NewRequestWithContext(ctx, "GET", u.String(), nil)
	if c.apiKey != "" {
		req.Header.Set("Authorization", "Bearer "+c.apiKey)
	}

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode/100 != 2 {
		return nil, fmt.Errorf("grafana search failed: %d", resp.StatusCode)
	}

	var dashboards []Dashboard
	if err := json.NewDecoder(resp.Body).Decode(&dashboards); err != nil {
		return nil, err
	}

	return dashboards, nil
}

// DashboardSnapshot is a snapshot of a dashboard.
type DashboardSnapshot struct {
	Key       string    `json:"key"`
	DeleteKey string    `json:"deleteKey"`
	URL       string    `json:"url"`
	ExpiresAt time.Time `json:"expires"`
}

// CreateSnapshot creates a dashboard snapshot.
func (c *Client) CreateSnapshot(ctx context.Context, dashboardUID string, timeRange *TimeRange) (*DashboardSnapshot, error) {
	// Build snapshot URL
	snapshotURL := fmt.Sprintf("%s/snapshot/%s", c.baseURL, dashboardUID)
	if timeRange != nil {
		snapshotURL += fmt.Sprintf("?from=%d&to=%d", timeRange.From/1000, timeRange.To/1000)
	}

	return &DashboardSnapshot{
		Key: snapshotURL,
		URL: snapshotURL,
		// Grafana snapshots are created via the UI or API
		// This is a simplified implementation
	}, nil
}

// TimeRange is a time range for queries.
type TimeRange struct {
	From int64 `json:"from"` // Unix timestamp in milliseconds
	To   int64 `json:"to"`   // Unix timestamp in milliseconds
}

// NewTimeRange creates a new time range.
func NewTimeRange(from, to time.Time) *TimeRange {
	return &TimeRange{
		From: from.UnixMilli(),
		To:   to.UnixMilli(),
	}
}

// LastHours creates a time range for the last N hours.
func LastHours(hours int) *TimeRange {
	now := time.Now()
	from := now.Add(-time.Duration(hours) * time.Hour)
	return NewTimeRange(from, now)
}

// LastMinutes creates a time range for the last N minutes.
func LastMinutes(minutes int) *TimeRange {
	now := time.Now()
	from := now.Add(-time.Duration(minutes) * time.Minute)
	return NewTimeRange(from, now)
}

// ExploreQuery represents an Explore query.
type ExploreQuery struct {
	Query  string   `json:"query"`
	Range  *TimeRange `json:"range"`
	Source string   `json:"source"` // "prometheus", "loki", etc.
}

// BuildExploreURL builds a Grafana Explore URL for a query.
func (c *Client) BuildExploreURL(query ExploreQuery) string {
	baseURL := c.baseURL + "/explore"

	// Encode query parameters
	params := make([]string, 0)
	params = append(params, "left="+url.QueryEscape(query.Query))

	if query.Range != nil {
		params = append(params, fmt.Sprintf("from=%d", query.Range.From))
		params = append(params, fmt.Sprintf("to=%d", query.Range.To))
	}

	if query.Source != "" {
		params = append(params, "source="+query.Source)
	}

	return baseURL + "?" + strings.Join(params, "&")
}

// DashboardURL builds a URL to a dashboard.
func (c *Client) DashboardURL(uid string, timeRange *TimeRange) string {
	baseURL := c.baseURL + "/d/" + uid

	if timeRange != nil {
		return fmt.Sprintf("%s?from=%d&to=%d", baseURL, timeRange.From, timeRange.To)
	}

	return baseURL
}

// LokiQueryURL builds a Loki Explore URL.
func LokiQueryURL(query string, timeRange *TimeRange) string {
	baseURL := os.Getenv("OBS_GRAFANA_URL")
	if baseURL == "" {
		baseURL = "http://127.0.0.1:3000"
	}
	baseURL = strings.TrimSuffix(baseURL, "/")

	// URL encode the query
	encodedQuery := url.QueryEscape(query)

	var timeParams string
	if timeRange != nil {
		timeParams = fmt.Sprintf("&from=%d&to=%d", timeRange.From, timeRange.To)
	}

	return fmt.Sprintf("%s/explore?left=%s%s", baseURL, encodedQuery, timeParams)
}

// PrometheusQueryURL builds a Prometheus Explore URL.
func PrometheusQueryURL(query string, timeRange *TimeRange) string {
	baseURL := os.Getenv("OBS_GRAFANA_URL")
	if baseURL == "" {
		baseURL = "http://127.0.0.1:3000"
	}
	baseURL = strings.TrimSuffix(baseURL, "/")

	// URL encode the query
	encodedQuery := url.QueryEscape(query)

	var timeParams string
	if timeRange != nil {
		timeParams = fmt.Sprintf("&from=%d&to=%d", timeRange.From, timeRange.To)
	}

	return fmt.Sprintf("%s/explore?prometheus=%s%s", baseURL, encodedQuery, timeParams)
}

// Annotation represents a Grafana annotation.
type Annotation struct {
	Text   string            `json:"text"`
	Time   int64             `json:"time"`   // Unix timestamp in milliseconds
	Tags   []string          `json:"tags"`
	Data   map[string]string `json:"data,omitempty"`
}

// CreateAnnotation creates a Grafana annotation.
func (c *Client) CreateAnnotation(ctx context.Context, ann *Annotation) error {
	u := c.baseURL + "/api/annotations"

	body, err := json.Marshal(ann)
	if err != nil {
		return err
	}

	req, err := http.NewRequestWithContext(ctx, "POST", u, strings.NewReader(string(body)))
	if err != nil {
		return err
	}

	if c.apiKey != "" {
		req.Header.Set("Authorization", "Bearer "+c.apiKey)
	}
	req.Header.Set("Content-Type", "application/json")

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode/100 != 2 {
		respBody, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("create annotation failed: %d - %s", resp.StatusCode, string(respBody))
	}

	return nil
}

// AlertingLink generates links for alert-related resources.
type AlertingLink struct {
	AlertName    string
	LabelFilters map[string]string
	TimeRange    *TimeRange
}

// GenerateLinks generates relevant links for an alert.
func GenerateLinks(alert *AlertingLink) map[string]string {
	links := make(map[string]string)

	// Loki query based on labels
	if alert.AlertName != "" {
		lokiQuery := fmt.Sprintf(`{alertname="%s"}`, alert.AlertName)
		links["loki"] = LokiQueryURL(lokiQuery, alert.TimeRange)
	}

	// Prometheus query based on alert
	if alert.AlertName != "" {
		promQuery := fmt.Sprintf(`ALERTS{alertname="%s"}`, alert.AlertName)
		links["prometheus"] = PrometheusQueryURL(promQuery, alert.TimeRange)
	}

	return links
}

// AlertContext provides context for an alert.
type AlertContext struct {
	AlertName   string
	Severity    string
	StartsAt    time.Time
	EndsAt      time.Time
	Labels      map[string]string
	Annotations map[string]string
}

// BuildAlertDashboard builds or finds a dashboard for an alert.
func (c *Client) BuildAlertDashboard(ctx context.Context, alertCtx *AlertContext) (string, error) {
	// Search for existing dashboard
	query := fmt.Sprintf("%s %s", alertCtx.Severity, alertCtx.AlertName)
	dashboards, err := c.SearchDashboards(ctx, query)
	if err != nil {
		return "", err
	}

	if len(dashboards) > 0 {
		return c.DashboardURL(dashboards[0].UID, NewTimeRange(alertCtx.StartsAt, alertCtx.EndsAt)), nil
	}

	// No existing dashboard, return generic Explore URL
	queryStr := fmt.Sprintf(`{alertname="%s"}`, alertCtx.AlertName)
	tr := NewTimeRange(alertCtx.StartsAt, alertCtx.EndsAt)
	return LokiQueryURL(queryStr, tr), nil
}
