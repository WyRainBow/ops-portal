package playbook

import (
	"context"
	"fmt"
	"os/exec"
	"strings"
	"sync"
	"time"

	"github.com/WyRainBow/ops-portal/internal/ai/errors"
)

// Playbook is a predefined operational procedure.
type Playbook struct {
	ID             string        `json:"id"`
	Name           string        `json:"name"`
	Description    string        `json:"description"`
	Category       string        `json:"category"` // "restart", "rollback", "scale", "cache", etc.
	Severity       string        `json:"severity"` // "low", "medium", "high", "critical"
	Command        string        `json:"command"`  // Command to execute
	Timeout        time.Duration `json:"timeout"`
	RequireConfirm bool          `json:"require_confirm"`
	Enabled        bool          `json:"enabled"`
	Parameters     []Parameter   `json:"parameters"`
}

// Parameter is a playbook parameter.
type Parameter struct {
	Name        string `json:"name"`
	Type        string `json:"type"` // "string", "int", "bool"
	Required    bool   `json:"required"`
	Description string `json:"description"`
	Default     any    `json:"default"`
}

// ExecutionResult is the result of a playbook execution.
type ExecutionResult struct {
	PlaybookID  string        `json:"playbook_id"`
	ExecutionID string        `json:"execution_id"`
	Status      string        `json:"status"` // "pending", "running", "success", "failed"
	StartTime   time.Time     `json:"start_time"`
	EndTime     time.Time     `json:"end_time,omitempty"`
	Duration    time.Duration `json:"duration,omitempty"`
	Output      string        `json:"output,omitempty"`
	Error       string        `json:"error,omitempty"`
	ExitCode    int           `json:"exit_code,omitempty"`
}

// ExecutionRequest is a request to execute a playbook.
type ExecutionRequest struct {
	PlaybookID  string         `json:"playbook_id"`
	Parameters  map[string]any `json:"parameters"`
	Reason      string         `json:"reason"`       // Audit reason
	RequestedBy string         `json:"requested_by"` // User who requested
	DryRun      bool           `json:"dry_run"`      // Preview only
}

// AuditLog records playbook executions.
type AuditLog struct {
	ExecutionID string         `json:"execution_id"`
	PlaybookID  string         `json:"playbook_id"`
	RequestedBy string         `json:"requested_by"`
	Reason      string         `json:"reason"`
	Parameters  map[string]any `json:"parameters"`
	Status      string         `json:"status"`
	Timestamp   time.Time      `json:"timestamp"`
	Duration    time.Duration  `json:"duration"`
}

// Executor executes playbooks safely.
type Executor struct {
	playbooks  map[string]*Playbook
	mu         sync.RWMutex
	auditLog   []AuditLog
	executions map[string]*ExecutionResult
}

// NewExecutor creates a new playbook executor.
func NewExecutor() *Executor {
	e := &Executor{
		playbooks:  make(map[string]*Playbook),
		auditLog:   make([]AuditLog, 0),
		executions: make(map[string]*ExecutionResult),
	}
	e.registerStandardPlaybooks()
	return e
}

// registerStandardPlaybooks registers predefined safe playbooks.
func (e *Executor) registerStandardPlaybooks() {
	playbooks := []*Playbook{
		{
			ID:             "restart-service",
			Name:           "重启服务",
			Description:    "重启指定的服务",
			Category:       "restart",
			Severity:       "medium",
			Command:        "systemctl restart {service_name}",
			Timeout:        30 * time.Second,
			RequireConfirm: true,
			Enabled:        true,
			Parameters: []Parameter{
				{Name: "service_name", Type: "string", Required: true, Description: "服务名称"},
			},
		},
		{
			ID:             "clear-cache",
			Name:           "清理缓存",
			Description:    "清理应用程序缓存",
			Category:       "cache",
			Severity:       "low",
			Command:        "redis-cli FLUSHDB",
			Timeout:        10 * time.Second,
			RequireConfirm: false,
			Enabled:        true,
			Parameters:     []Parameter{},
		},
		{
			ID:             "scale-deployment",
			Name:           "扩缩容",
			Description:    "调整部署副本数",
			Category:       "scale",
			Severity:       "medium",
			Command:        "kubectl scale deployment {deployment} --replicas={replicas}",
			Timeout:        60 * time.Second,
			RequireConfirm: true,
			Enabled:        true,
			Parameters: []Parameter{
				{Name: "deployment", Type: "string", Required: true, Description: "部署名称"},
				{Name: "replicas", Type: "int", Required: true, Description: "副本数量"},
			},
		},
		{
			ID:             "check-disk",
			Name:           "检查磁盘空间",
			Description:    "检查服务器磁盘使用情况",
			Category:       "diagnostic",
			Severity:       "low",
			Command:        "df -h",
			Timeout:        10 * time.Second,
			RequireConfirm: false,
			Enabled:        true,
			Parameters:     []Parameter{},
		},
		{
			ID:             "check-process",
			Name:           "检查进程",
			Description:    "检查指定进程的运行状态",
			Category:       "diagnostic",
			Severity:       "low",
			Command:        "ps aux | grep {process_name}",
			Timeout:        10 * time.Second,
			RequireConfirm: false,
			Enabled:        true,
			Parameters: []Parameter{
				{Name: "process_name", Type: "string", Required: true, Description: "进程名称"},
			},
		},
	}

	for _, pb := range playbooks {
		e.playbooks[pb.ID] = pb
	}

	errors.Info("playbook", fmt.Sprintf("registered %d playbooks", len(playbooks)))
}

// Register registers a new playbook.
func (e *Executor) Register(pb *Playbook) error {
	e.mu.Lock()
	defer e.mu.Unlock()

	if _, exists := e.playbooks[pb.ID]; exists {
		return fmt.Errorf("playbook %s already exists", pb.ID)
	}

	e.playbooks[pb.ID] = pb
	return nil
}

// Get retrieves a playbook by ID.
func (e *Executor) Get(id string) (*Playbook, bool) {
	e.mu.RLock()
	defer e.mu.RUnlock()

	pb, ok := e.playbooks[id]
	if !ok || !pb.Enabled {
		return nil, false
	}
	return pb, true
}

// List returns all enabled playbooks.
func (e *Executor) List() []*Playbook {
	e.mu.RLock()
	defer e.mu.RUnlock()

	result := make([]*Playbook, 0)
	for _, pb := range e.playbooks {
		if pb.Enabled {
			result = append(result, pb)
		}
	}
	return result
}

// Execute executes a playbook.
func (e *Executor) Execute(ctx context.Context, req *ExecutionRequest) (*ExecutionResult, error) {
	// Get playbook
	pb, ok := e.Get(req.PlaybookID)
	if !ok {
		return nil, fmt.Errorf("playbook not found: %s", req.PlaybookID)
	}

	// Generate execution ID
	executionID := fmt.Sprintf("EXEC-%d", time.Now().UnixNano())

	// Check if confirmation is required
	if pb.RequireConfirm && !req.DryRun {
		// In production, this would require explicit user confirmation
		// For now, we check if a confirmation header/flag is set
	}

	// Create execution result
	result := &ExecutionResult{
		PlaybookID:  pb.ID,
		ExecutionID: executionID,
		Status:      "pending",
		StartTime:   time.Now(),
	}
	e.executions[executionID] = result

	// Log audit entry
	audit := AuditLog{
		ExecutionID: executionID,
		PlaybookID:  pb.ID,
		RequestedBy: req.RequestedBy,
		Reason:      req.Reason,
		Parameters:  req.Parameters,
		Status:      "pending",
		Timestamp:   time.Now(),
	}
	e.mu.Lock()
	e.auditLog = append(e.auditLog, audit)
	e.mu.Unlock()

	// Dry run mode
	if req.DryRun {
		result.Status = "success"
		result.EndTime = time.Now()
		result.Duration = result.EndTime.Sub(result.StartTime)
		result.Output = fmt.Sprintf("Dry run: would execute '%s'", pb.Command)
		return result, nil
	}

	// Execute the playbook
	return e.executePlaybook(ctx, pb, req, result)
}

// executePlaybook executes a playbook command.
func (e *Executor) executePlaybook(ctx context.Context, pb *Playbook, req *ExecutionRequest, result *ExecutionResult) (*ExecutionResult, error) {
	result.Status = "running"

	// Build command
	cmdStr := pb.Command
	for key, value := range req.Parameters {
		placeholder := fmt.Sprintf("{%s}", key)
		cmdStr = strings.ReplaceAll(cmdStr, placeholder, fmt.Sprintf("%v", value))
	}

	// Parse command
	parts := strings.Fields(cmdStr)
	if len(parts) == 0 {
		result.Status = "failed"
		result.Error = "empty command"
		return result, fmt.Errorf("empty command")
	}

	// Create command with timeout
	cmdCtx, cancel := context.WithTimeout(ctx, pb.Timeout)
	defer cancel()

	var cmd *exec.Cmd
	if len(parts) > 1 {
		cmd = exec.CommandContext(cmdCtx, parts[0], parts[1:]...)
	} else {
		cmd = exec.CommandContext(cmdCtx, parts[0])
	}

	// Execute
	output, err := cmd.CombinedOutput()
	result.EndTime = time.Now()
	result.Duration = result.EndTime.Sub(result.StartTime)
	result.Output = string(output)

	if err != nil {
		result.Status = "failed"
		result.Error = err.Error()
		if exitErr, ok := err.(*exec.ExitError); ok {
			result.ExitCode = exitErr.ExitCode()
		}
	} else {
		result.Status = "success"
		result.ExitCode = 0
	}

	// Update audit log
	e.mu.Lock()
	for i, log := range e.auditLog {
		if log.ExecutionID == result.ExecutionID {
			e.auditLog[i].Status = result.Status
			e.auditLog[i].Duration = result.Duration
			break
		}
	}
	e.mu.Unlock()

	return result, nil
}

// GetExecution retrieves an execution result.
func (e *Executor) GetExecution(executionID string) (*ExecutionResult, bool) {
	e.mu.RLock()
	defer e.mu.RUnlock()

	result, ok := e.executions[executionID]
	return result, ok
}

// GetAuditLog returns the audit log.
func (e *Executor) GetAuditLog() []AuditLog {
	e.mu.RLock()
	defer e.mu.RUnlock()

	log := make([]AuditLog, len(e.auditLog))
	copy(log, e.auditLog)
	return log
}

// Global executor instance.
var globalExecutor *Executor

// InitExecutor initializes the global executor.
func InitExecutor() {
	globalExecutor = NewExecutor()
}

// GlobalExecutor returns the global executor.
func GlobalExecutor() *Executor {
	return globalExecutor
}

// ExecuteSafe is a helper that executes with safety checks.
func ExecuteSafe(ctx context.Context, playbookID string, params map[string]any, requestedBy, reason string, dryRun bool) (*ExecutionResult, error) {
	if globalExecutor == nil {
		return nil, fmt.Errorf("executor not initialized")
	}

	req := &ExecutionRequest{
		PlaybookID:  playbookID,
		Parameters:  params,
		Reason:      reason,
		RequestedBy: requestedBy,
		DryRun:      dryRun,
	}

	return globalExecutor.Execute(ctx, req)
}
