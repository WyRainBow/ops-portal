package alerting

import (
	"context"
	"encoding/json"
	"fmt"
	"sync"
	"time"

	"github.com/WyRainBow/ops-portal/internal/ai/errors"
)

// Store stores incidents in memory.
// In production, this should be replaced with a database.
type Store struct {
	mu         sync.RWMutex
	incidents  map[string]*Incident
	byFingerprint map[string]*Incident // Track alerts by fingerprint for deduplication
}

// NewStore creates a new incident store.
func NewStore() *Store {
	return &Store{
		incidents:  make(map[string]*Incident),
		byFingerprint: make(map[string]*Incident),
	}
}

// Add adds an incident to the store.
func (s *Store) Add(incident *Incident) {
	s.mu.Lock()
	defer s.mu.Unlock()

	s.incidents[incident.ID] = incident

	// Track by fingerprint for resolved alerts
	if fp := incident.Labels["fingerprint"]; fp != "" {
		s.byFingerprint[fp] = incident
	}
}

// Get retrieves an incident by ID.
func (s *Store) Get(id string) (*Incident, bool) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	incident, ok := s.incidents[id]
	return incident, ok
}

// List returns all incidents.
func (s *Store) List() []*Incident {
	s.mu.RLock()
	defer s.mu.RUnlock()

	result := make([]*Incident, 0, len(s.incidents))
	for _, inc := range s.incidents {
		result = append(result, inc)
	}
	return result
}

// ListFiring returns only firing incidents.
func (s *Store) ListFiring() []*Incident {
	s.mu.RLock()
	defer s.mu.RUnlock()

	result := make([]*Incident, 0)
	for _, inc := range s.incidents {
		if inc.Status == "firing" {
			result = append(result, inc)
		}
	}
	return result
}

// UpdateStatus updates an incident's status.
func (s *Store) UpdateStatus(id, status string) bool {
	s.mu.Lock()
	defer s.mu.Unlock()

	incident, ok := s.incidents[id]
	if !ok {
		return false
	}
	incident.Status = status
	return true
}

// Delete removes an incident from the store.
func (s *Store) Delete(id string) bool {
	s.mu.Lock()
	defer s.mu.Unlock()

	incident, ok := s.incidents[id]
	if !ok {
		return false
	}

	delete(s.incidents, id)
	if fp := incident.Labels["fingerprint"]; fp != "" {
		delete(s.byFingerprint, fp)
	}
	return true
}

// Global store instance.
var globalStore *Store

// InitStore initializes the global store.
func InitStore() {
	globalStore = NewStore()
}

// GlobalStore returns the global store.
func GlobalStore() *Store {
	if globalStore == nil {
		InitStore()
	}
	return globalStore
}

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
