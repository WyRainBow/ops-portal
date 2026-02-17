package metrics

import (
	"sync"
	"time"

	"github.com/WyRainBow/ops-portal/internal/ai/errors"
)

// MetricType represents the type of metric.
type MetricType int

const (
	Counter MetricType = iota
	Gauge
	Histogram
	Summary
)

// Metric represents a single metric.
type Metric struct {
	Name        string
	Type        MetricType
	Value       float64
	Labels      map[string]string
	Description string
}

// MetricsRegistry manages all metrics.
type MetricsRegistry struct {
	mu      sync.RWMutex
	metrics map[string]*Metric
}

// Global registry instance.
var globalRegistry = &MetricsRegistry{
	metrics: make(map[string]*Metric),
}

// Global returns the global metrics registry.
func Global() *MetricsRegistry {
	return globalRegistry
}

// Register registers a new metric.
func (r *MetricsRegistry) Register(name string, metricType MetricType, description string) {
	r.mu.Lock()
	defer r.mu.Unlock()

	r.metrics[name] = &Metric{
		Name:        name,
		Type:        metricType,
		Value:       0,
		Labels:      make(map[string]string),
		Description: description,
	}
}

// Increment increments a counter metric.
func (r *MetricsRegistry) Increment(name string, labels map[string]string) {
	r.mu.Lock()
	defer r.mu.Unlock()

	if metric, ok := r.metrics[name]; ok {
		metric.Value++
		if labels != nil {
			metric.Labels = labels
		}
	}
}

// Add adds a value to a counter metric.
func (r *MetricsRegistry) Add(name string, value float64, labels map[string]string) {
	r.mu.Lock()
	defer r.mu.Unlock()

	if metric, ok := r.metrics[name]; ok {
		metric.Value += value
		if labels != nil {
			metric.Labels = labels
		}
	}
}

// Set sets a gauge metric value.
func (r *MetricsRegistry) Set(name string, value float64, labels map[string]string) {
	r.mu.Lock()
	defer r.mu.Unlock()

	if metric, ok := r.metrics[name]; ok {
		metric.Value = value
		if labels != nil {
			metric.Labels = labels
		}
	}
}

// Timing records a timing value in seconds.
func (r *MetricsRegistry) Timing(name string, duration time.Duration, labels map[string]string) {
	r.mu.Lock()
	defer r.mu.Unlock()

	if metric, ok := r.metrics[name]; ok {
		metric.Value = duration.Seconds()
		if labels != nil {
			metric.Labels = labels
		}
	}
}

// Get retrieves a metric value.
func (r *MetricsRegistry) Get(name string) (float64, bool) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	if metric, ok := r.metrics[name]; ok {
		return metric.Value, true
	}
	return 0, false
}

// GetAll returns all metrics.
func (r *MetricsRegistry) GetAll() map[string]*Metric {
	r.mu.RLock()
	defer r.mu.RUnlock()

	result := make(map[string]*Metric, len(r.metrics))
	for k, v := range r.metrics {
		result[k] = v
	}
	return result
}

// InitializeStandardMetrics initializes all standard ops-portal metrics.
func InitializeStandardMetrics() {
	r := Global()

	// Tool metrics
	r.Register("ops_portal_tool_calls_total", Counter, "Total number of tool calls")
	r.Register("ops_portal_tool_duration_seconds", Histogram, "Tool execution duration in seconds")
	r.Register("ops_portal_tool_errors_total", Counter, "Total number of tool errors")

	// Agent metrics
	r.Register("ops_portal_agent_steps_total", Counter, "Total number of agent steps")
	r.Register("ops_portal_agent_duration_seconds", Histogram, "Agent execution duration in seconds")
	r.Register("ops_portal_agent_errors_total", Counter, "Total number of agent errors")

	// LLM metrics
	r.Register("ops_portal_llm_tokens_total", Counter, "Total number of LLM tokens")
	r.Register("ops_portal_llm_requests_total", Counter, "Total number of LLM requests")
	r.Register("ops_portal_llm_duration_seconds", Histogram, "LLM request duration in seconds")

	// RAG metrics
	r.Register("ops_portal_rag_retrievals_total", Counter, "Total number of RAG retrievals")
	r.Register("ops_portal_rag_documents_retrieved", Summary, "Number of documents retrieved per query")
	r.Register("ops_portal_rag_rerank_duration_seconds", Histogram, "Reranking duration in seconds")

	// Alert metrics
	r.Register("ops_portal_alerts_received_total", Counter, "Total number of alerts received")
	r.Register("ops_portal_alerts_processed_total", Counter, "Total number of alerts processed")
	r.Register("ops_portal_incidents_created_total", Counter, "Total number of incidents created")

	errors.Info("metrics", "standard metrics initialized")
}

// Metric names constants for type safety.
const (
	ToolCallsTotal      = "ops_portal_tool_calls_total"
	ToolDurationSeconds = "ops_portal_tool_duration_seconds"
	ToolErrorsTotal     = "ops_portal_tool_errors_total"

	AgentStepsTotal      = "ops_portal_agent_steps_total"
	AgentDurationSeconds = "ops_portal_agent_duration_seconds"
	AgentErrorsTotal     = "ops_portal_agent_errors_total"

	LLMTokensTotal     = "ops_portal_llm_tokens_total"
	LLMRequestsTotal   = "ops_portal_llm_requests_total"
	LLMDurationSeconds = "ops_portal_llm_duration_seconds"

	RAGRetrievalsTotal       = "ops_portal_rag_retrievals_total"
	RAGDocumentsRetrieved    = "ops_portal_rag_documents_retrieved"
	RAGRerankDurationSeconds = "ops_portal_rag_rerank_duration_seconds"

	AlertsReceivedTotal   = "ops_portal_alerts_received_total"
	AlertsProcessedTotal  = "ops_portal_alerts_processed_total"
	IncidentsCreatedTotal = "ops_portal_incidents_created_total"
)

// InstrumentedTool wraps a tool with metrics collection.
type InstrumentedTool struct {
	name string
}

// NewInstrumentedTool creates a new instrumented tool wrapper.
func NewInstrumentedTool(name string) *InstrumentedTool {
	return &InstrumentedTool{name: name}
}

// RecordCall records a tool call.
func (i *InstrumentedTool) RecordCall(labels map[string]string) {
	Global().Increment(ToolCallsTotal, map[string]string{
		"tool": i.name,
	})
	for k, v := range labels {
		Global().metrics[ToolCallsTotal].Labels[k] = v
	}
}

// RecordSuccess records a successful tool execution.
func (i *InstrumentedTool) RecordSuccess(duration time.Duration) {
	Global().Timing(ToolDurationSeconds, duration, map[string]string{
		"tool":   i.name,
		"status": "success",
	})
}

// RecordError records a tool error.
func (i *InstrumentedTool) RecordError(duration time.Duration, errMsg string) {
	Global().Timing(ToolDurationSeconds, duration, map[string]string{
		"tool":   i.name,
		"status": "error",
	})
	Global().Increment(ToolErrorsTotal, map[string]string{
		"tool":  i.name,
		"error": errMsg,
	})
}
