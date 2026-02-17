package feishu

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"time"

	"github.com/WyRainBow/ops-portal/internal/ai/errors"
)

// Client is a Feishu OpenAPI client.
type Client struct {
	appID     string
	appSecret string
	baseURL   string
	httpCli   *http.Client
	token     string
	tokenExp  time.Time
}

// NewClient creates a new Feishu client.
func NewClient(appID, appSecret string) *Client {
	baseURL := os.Getenv("FEISHU_BASE_URL")
	if baseURL == "" {
		baseURL = "https://open.feishu.cn"
	}

	return &Client{
		appID:     appID,
		appSecret: appSecret,
		baseURL:   baseURL,
		httpCli: &http.Client{
			Timeout: 30 * time.Second,
		},
	}
}

// getTenantAccessToken fetches a tenant access token.
func (c *Client) getTenantAccessToken(ctx context.Context) (string, error) {
	// Check if cached token is still valid
	if c.token != "" && time.Now().Before(c.tokenExp) {
		return c.token, nil
	}

	url := fmt.Sprintf("%s/open-apis/auth/v3/tenant_access_token/internal", c.baseURL)
	payload := map[string]string{
		"app_id":     c.appID,
		"app_secret": c.appSecret,
	}

	body, err := json.Marshal(payload)
	if err != nil {
		return "", err
	}

	req, err := http.NewRequestWithContext(ctx, "POST", url, bytes.NewReader(body))
	if err != nil {
		return "", err
	}

	req.Header.Set("Content-Type", "application/json")

	resp, err := c.httpCli.Do(req)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", err
	}

	var result struct {
		Code              int    `json:"code"`
		TenantAccessToken string `json:"tenant_access_token"`
		Expire            int    `json:"expire"`
	}

	if err := json.Unmarshal(respBody, &result); err != nil {
		return "", err
	}

	if result.Code != 0 {
		return "", fmt.Errorf("feishu auth failed: code=%d", result.Code)
	}

	// Cache token with buffer (5 minutes before expiration)
	c.token = result.TenantAccessToken
	c.tokenExp = time.Now().Add(time.Duration(result.Expire-300) * time.Second)

	return c.token, nil
}

// Message is a Feishu message.
type Message struct {
	ReceiveID     string `json:"receive_id"`
	MsgType       string `json:"msg_type"`
	Content       string `json:"content"`
	ReceiveIDType string `json:"receive_id_type,omitempty"`
}

// TextMessage is a text message content.
type TextMessage struct {
	Text string `json:"text"`
}

// SendText sends a text message to a chat.
func (c *Client) SendText(ctx context.Context, chatID, text string) error {
	token, err := c.getTenantAccessToken(ctx)
	if err != nil {
		return err
	}

	content, _ := json.Marshal(TextMessage{Text: text})
	msg := Message{
		ReceiveID:     chatID,
		MsgType:       "text",
		Content:       string(content),
		ReceiveIDType: "chat_id",
	}

	return c.sendMessage(ctx, token, msg)
}

// sendMessage sends a message via Feishu OpenAPI.
func (c *Client) sendMessage(ctx context.Context, token string, msg Message) error {
	url := fmt.Sprintf("%s/open-apis/im/v1/messages?receive_id_type=%s",
		c.baseURL, msg.ReceiveIDType)

	payload := map[string]any{
		"receive_id": msg.ReceiveID,
		"msg_type":   msg.MsgType,
		"content":    msg.Content,
	}

	body, err := json.Marshal(payload)
	if err != nil {
		return err
	}

	req, err := http.NewRequestWithContext(ctx, "POST", url, bytes.NewReader(body))
	if err != nil {
		return err
	}

	req.Header.Set("Authorization", "Bearer "+token)
	req.Header.Set("Content-Type", "application/json")

	resp, err := c.httpCli.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	respBody, _ := io.ReadAll(resp.Body)
	if resp.StatusCode/100 != 2 {
		return fmt.Errorf("feishu send failed: %d - %s", resp.StatusCode, string(respBody))
	}

	return nil
}

// CardMessage is an interactive card message.
type CardMessage struct {
	Header   *CardHeader   `json:"header,omitempty"`
	Elements []CardElement `json:"elements"`
}

// CardHeader is the header of a card.
type CardHeader struct {
	Title    string `json:"title"`
	Subtitle string `json:"subtitle,omitempty"`
}

// CardElement is an element in a card.
type CardElement struct {
	Tag  string `json:"tag"`
	Text string `json:"text,omitempty"`
	// Additional fields based on element type
}

// SendCard sends an interactive card message.
func (c *Client) SendCard(ctx context.Context, chatID string, card CardMessage) error {
	token, err := c.getTenantAccessToken(ctx)
	if err != nil {
		return err
	}

	content, _ := json.Marshal(card)
	msg := Message{
		ReceiveID:     chatID,
		MsgType:       "interactive",
		Content:       string(content),
		ReceiveIDType: "chat_id",
	}

	return c.sendMessage(ctx, token, msg)
}

// AlertNotification is an alert notification for Feishu.
type AlertNotification struct {
	AlertName   string
	Severity    string
	Status      string
	Summary     string
	Description string
	Labels      map[string]string
	StartsAt    time.Time
}

// SendAlertNotification sends an alert notification to Feishu.
func (c *Client) SendAlertNotification(ctx context.Context, chatID string, alert *AlertNotification) error {
	// Create a formatted message
	var severityColor string
	switch alert.Severity {
	case "critical", "error":
		severityColor = "🔴"
	case "warning":
		severityColor = "🟡"
	default:
		severityColor = "🟢"
	}

	text := fmt.Sprintf("%s **运维告警通知**\n\n"+
		"**告警名称**: %s\n"+
		"**状态**: %s\n"+
		"**级别**: %s\n"+
		"**摘要**: %s\n"+
		"**描述**: %s\n"+
		"**时间**: %s\n",
		severityColor,
		alert.AlertName,
		alert.Status,
		alert.Severity,
		alert.Summary,
		alert.Description,
		alert.StartsAt.Format("2006-01-02 15:04:05"),
	)

	if len(alert.Labels) > 0 {
		text += "\n**标签**:\n"
		for k, v := range alert.Labels {
			text += fmt.Sprintf("  - %s: %s\n", k, v)
		}
	}

	return c.SendText(ctx, chatID, text)
}

// DiagnosticReportNotification is a diagnostic report notification.
type DiagnosticReportNotification struct {
	IncidentID      string
	AlertName       string
	ExecutionStatus string
	Summary         string
	Recommendations []string
	Actions         []string
	ReportLink      string
}

// SendDiagnosticReport sends a diagnostic report notification to Feishu.
func (c *Client) SendDiagnosticReport(ctx context.Context, chatID string, report *DiagnosticReportNotification) error {
	statusIcon := "✅"
	if report.ExecutionStatus != "success" {
		statusIcon = "❌"
	}

	text := fmt.Sprintf("%s **AI 诊断报告**\n\n"+
		"**事件ID**: %s\n"+
		"**告警**: %s\n"+
		"**状态**: %s\n"+
		"**摘要**: %s\n",
		statusIcon,
		report.IncidentID,
		report.AlertName,
		report.ExecutionStatus,
		report.Summary,
	)

	if len(report.Recommendations) > 0 {
		text += "\n**建议**:\n"
		for i, rec := range report.Recommendations {
			text += fmt.Sprintf("%d. %s\n", i+1, rec)
		}
	}

	if len(report.Actions) > 0 {
		text += "\n**操作**:\n"
		for i, act := range report.Actions {
			text += fmt.Sprintf("%d. %s\n", i+1, act)
		}
	}

	if report.ReportLink != "" {
		text += fmt.Sprintf("\n**完整报告**: %s\n", report.ReportLink)
	}

	return c.SendText(ctx, chatID, text)
}

// Notifier is a Feishu notifier singleton.
type Notifier struct {
	client *Client
	chatID string
}

var globalNotifier *Notifier

// InitNotifier initializes the global Feishu notifier.
func InitNotifier() error {
	appID := os.Getenv("FEISHU_APP_ID")
	appSecret := os.Getenv("FEISHU_APP_SECRET")
	chatID := os.Getenv("FEISHU_ALERT_CHAT_ID")

	if appID == "" || appSecret == "" || chatID == "" {
		// Not configured, return without error
		return nil
	}

	globalNotifier = &Notifier{
		client: NewClient(appID, appSecret),
		chatID: chatID,
	}

	errors.Info("feishu", "notifier initialized")
	return nil
}

// GlobalNotifier returns the global Feishu notifier.
func GlobalNotifier() *Notifier {
	return globalNotifier
}

// SendAlert sends an alert notification.
func (n *Notifier) SendAlert(ctx context.Context, alert *AlertNotification) error {
	if n == nil {
		return nil // Not configured
	}
	return n.client.SendAlertNotification(ctx, n.chatID, alert)
}

// SendReport sends a diagnostic report notification.
func (n *Notifier) SendReport(ctx context.Context, report *DiagnosticReportNotification) error {
	if n == nil {
		return nil // Not configured
	}
	return n.client.SendDiagnosticReport(ctx, n.chatID, report)
}
