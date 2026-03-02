package alerting

import (
	"context"
	"fmt"
	"sync"
	"time"

	"github.com/WyRainBow/ops-portal/internal/ai/agent/plan_execute_replan"
)

// DiagnosisResult represents the result of an AI diagnosis
type DiagnosisResult struct {
	IncidentID string
	Result     string
	Detail     []string
	Error      error
	CreatedAt  time.Time
}

// DiagnosisService handles async AI diagnosis for alerts
type DiagnosisService struct {
	mu            sync.RWMutex
	pending       map[string]*Incident
	results       map[string]*DiagnosisResult
	maxConcurrent int
	sem           chan struct{}
}

// Global diagnosis service instance
var globalDiagnosis *DiagnosisService
var once sync.Once

// GlobalDiagnosis returns the global diagnosis service
func GlobalDiagnosis() *DiagnosisService {
	once.Do(func() {
		globalDiagnosis = &DiagnosisService{
			pending:       make(map[string]*Incident),
			results:       make(map[string]*DiagnosisResult),
			maxConcurrent: 3, // Max 3 concurrent diagnoses
			sem:           make(chan struct{}, 3),
		}
	})
	return globalDiagnosis
}

// TriggerDiagnosis asynchronously triggers AI diagnosis for an incident
func (s *DiagnosisService) TriggerDiagnosis(incident *Incident) {
	s.mu.Lock()
	s.pending[incident.ID] = incident
	s.mu.Unlock()

	// Start async diagnosis
	go s.diagnose(incident)
}

// diagnose performs the actual AI diagnosis
func (s *DiagnosisService) diagnose(incident *Incident) {
	// Acquire semaphore
	s.sem <- struct{}{}
	defer func() { <-s.sem }()

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Minute)
	defer cancel()

	startTime := time.Now()

	// Build diagnosis query
	query := fmt.Sprintf(`
"1. 你是一个智能的服务告警分析助手。"
"2. 告警名称：%s"
"3. 告警级别：%s"
"4. 告警描述：%s"
"5. 请调用工具query_internal_docs获取该告警的处理方案。"
"6. 涉及到时间的参数都需要先通过工具get_current_time获取当前时间。"
"7. 涉及到日志的查询,使用工具query_loki_logs从 Loki 查询日志。"
"8. 最后生成告警运维分析报告，格式如下：
告警分析报告
---
# 告警处理详情
## 告警信息
## 根因分析
## 处理建议
"
`, incident.AlertName, incident.Severity, incident.Description)

	result, detail, err := plan_execute_replan.BuildPlanAgent(ctx, query)

	// Store result
	s.mu.Lock()
	defer s.mu.Unlock()

	delete(s.pending, incident.ID)

	diagnosisResult := &DiagnosisResult{
		IncidentID: incident.ID,
		Result:     result,
		Detail:     detail,
		Error:      err,
		CreatedAt:  time.Now(),
	}
	s.results[incident.ID] = diagnosisResult

	duration := time.Since(startTime)
	if err != nil {
		fmt.Printf("[ERROR] AI diagnosis failed for %s: %v (took %v)\n", incident.ID, err, duration)
	} else {
		fmt.Printf("[INFO] AI diagnosis completed for %s (took %v)\n", incident.ID, duration)
	}
}

// GetResult retrieves the diagnosis result for an incident
func (s *DiagnosisService) GetResult(incidentID string) (*DiagnosisResult, bool) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	result, ok := s.results[incidentID]
	return result, ok
}

// IsPending checks if diagnosis is pending for an incident
func (s *DiagnosisService) IsPending(incidentID string) bool {
	s.mu.RLock()
	defer s.mu.RUnlock()

	_, pending := s.pending[incidentID]
	return pending
}

// GetPendingCount returns the number of pending diagnoses
func (s *DiagnosisService) GetPendingCount() int {
	s.mu.RLock()
	defer s.mu.RUnlock()

	return len(s.pending)
}

// GetResultsCount returns the number of completed diagnosis results
func (s *DiagnosisService) GetResultsCount() int {
	s.mu.RLock()
	defer s.mu.RUnlock()

	return len(s.results)
}
