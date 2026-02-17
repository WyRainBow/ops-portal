package alerting

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/WyRainBow/ops-portal/internal/ai/errors"
)

// Alert represents a Prometheus alert.
type Alert struct {
	Status       string            `json:"status"`
	Labels       map[string]string `json:"labels"`
	Annotations  map[string]string `json:"annotations"`
	StartsAt     time.Time         `json:"startsAt"`
	EndsAt       time.Time         `json:"endsAt"`
	GeneratorURL string            `json:"generatorURL"`
	Fingerprint  string            `json:"fingerprint"`
}

// AlertmanagerWebhook represents the webhook payload from Alertmanager.
type AlertmanagerWebhook struct {
	Receiver          string            `json:"receiver"`
	Status            string            `json:"status"`
	Alerts            []Alert           `json:"alerts"`
	GroupLabels       map[string]string `json:"groupLabels"`
	CommonLabels      map[string]string `json:"commonLabels"`
	CommonAnnotations map[string]string `json:"commonAnnotations"`
	ExternalURL       string            `json:"externalURL"`
	Version           string            `json:"version"`
	GroupKey          string            `json:"groupKey"`
	TruncatedAlerts   int               `json:"truncatedAlerts"`
}

// Incident represents an ops incident created from an alert.
type Incident struct {
	ID          string            `json:"id"`
	AlertName   string            `json:"alert_name"`
	Status      string            `json:"status"`
	Severity    string            `json:"severity"`
	Summary     string            `json:"summary"`
	Description string            `json:"description"`
	Labels      map[string]string `json:"labels"`
	StartedAt   time.Time         `json:"started_at"`
	CreatedAt   time.Time         `json:"created_at"`
}

// Ingester handles alert ingestion from Alertmanager.
type Ingester struct {
	// In production, this would include database/storage for incidents
}

// NewIngester creates a new alert ingester.
func NewIngester() *Ingester {
	return &Ingester{}
}

// IngestAlert handles an incoming alert webhook.
func (i *Ingester) IngestAlert(ctx context.Context, webhook *AlertmanagerWebhook) (*Incident, error) {
	errors.Info("alerting", fmt.Sprintf("received webhook: receiver=%s, status=%s, alerts=%d",
		webhook.Receiver, webhook.Status, len(webhook.Alerts)))

	if len(webhook.Alerts) == 0 {
		return nil, fmt.Errorf("no alerts in webhook")
	}

	// For now, process the first alert as the primary incident
	alert := webhook.Alerts[0]

	// Extract incident details
	incident := &Incident{
		ID:        fmt.Sprintf("INC-%d", time.Now().Unix()),
		Status:    webhook.Status,
		Labels:    alert.Labels,
		StartedAt: alert.StartsAt,
		CreatedAt: time.Now(),
	}

	// Extract alert name
	if name, ok := alert.Labels["alertname"]; ok {
		incident.AlertName = name
	} else {
		incident.AlertName = "unknown"
	}

	// Extract severity
	if severity, ok := alert.Labels["severity"]; ok {
		incident.Severity = severity
	} else {
		incident.Severity = "warning"
	}

	// Extract summary and description
	if summary, ok := alert.Annotations["summary"]; ok {
		incident.Summary = summary
	} else {
		incident.Summary = incident.AlertName
	}

	if desc, ok := alert.Annotations["description"]; ok {
		incident.Description = desc
	} else {
		incident.Description = fmt.Sprintf("Alert: %s", incident.AlertName)
	}

	errors.Info("alerting", fmt.Sprintf("created incident: id=%s, alert=%s, severity=%s",
		incident.ID, incident.AlertName, incident.Severity))

	return incident, nil
}

// ShouldTriggerDiagnosis determines if an alert should trigger AI diagnosis.
func (i *Ingester) ShouldTriggerDiagnosis(incident *Incident) bool {
	// Only trigger diagnosis for firing alerts with high severity
	if incident.Status != "firing" {
		return false
	}

	// Trigger for critical and high severity alerts
	switch incident.Severity {
	case "critical", "high", "error":
		return true
	default:
		return false
	}
}

// FormatIncidentForLLM formats an incident for LLM consumption.
func (i *Ingester) FormatIncidentForLLM(incident *Incident) string {
	return fmt.Sprintf(`Incident: %s
Status: %s
Severity: %s
Summary: %s
Description: %s
Started At: %s
Labels: %v`,
		incident.ID,
		incident.Status,
		incident.Severity,
		incident.Summary,
		incident.Description,
		incident.StartedAt.Format(time.RFC3339),
		incident.Labels,
	)
}

// ParseWebhook parses JSON bytes into an AlertmanagerWebhook.
func ParseWebhook(data []byte) (*AlertmanagerWebhook, error) {
	var webhook AlertmanagerWebhook
	if err := json.Unmarshal(data, &webhook); err != nil {
		return nil, fmt.Errorf("failed to parse webhook: %w", err)
	}
	return &webhook, nil
}

// Queue manages alert queue for processing.
type Queue struct {
	incidents chan *Incident
	ingester  *Ingester
}

// NewQueue creates a new alert queue.
func NewQueue(size int) *Queue {
	return &Queue{
		incidents: make(chan *Incident, size),
		ingester:  NewIngester(),
	}
}

// Enqueue adds an incident to the queue.
func (q *Queue) Enqueue(incident *Incident) error {
	select {
	case q.incidents <- incident:
		errors.Info("alerting", fmt.Sprintf("enqueued incident: %s", incident.ID))
		return nil
	default:
		return fmt.Errorf("alert queue is full")
	}
}

// Dequeue removes and returns an incident from the queue.
// This method blocks until an incident is available.
func (q *Queue) Dequeue(ctx context.Context) (*Incident, error) {
	select {
	case incident := <-q.incidents:
		return incident, nil
	case <-ctx.Done():
		return nil, ctx.Err()
	}
}

// Size returns the current queue size.
func (q *Queue) Size() int {
	return len(q.incidents)
}

// ProcessWebhook processes an incoming webhook and queues the incident.
func (q *Queue) ProcessWebhook(ctx context.Context, data []byte) (*Incident, error) {
	webhook, err := ParseWebhook(data)
	if err != nil {
		return nil, err
	}

	incident, err := q.ingester.IngestAlert(ctx, webhook)
	if err != nil {
		return nil, err
	}

	// Enqueue for processing
	if err := q.Enqueue(incident); err != nil {
		return incident, err
	}

	return incident, nil
}
