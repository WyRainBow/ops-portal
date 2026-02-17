package parallel

import (
	"context"
	"fmt"
	"sync"
	"time"

	"github.com/WyRainBow/ops-portal/internal/ai/errors"
	"github.com/cloudwego/eino/components/tool"
)

// ToolCall represents a tool invocation.
type ToolCall struct {
	Name   string
	Input  any
	Result string
	Error  error
	Done   bool
}

// Executor executes tools in parallel.
type Executor struct {
	maxConcurrency int
}

// NewExecutor creates a new parallel executor.
func NewExecutor(maxConcurrency int) *Executor {
	if maxConcurrency <= 0 {
		maxConcurrency = 5 // Default max concurrency
	}
	return &Executor{
		maxConcurrency: maxConcurrency,
	}
}

// Execute executes multiple tools in parallel.
// Returns results in the same order as the input tools.
func (e *Executor) Execute(ctx context.Context, calls []ToolCall) []ToolCall {
	if len(calls) == 0 {
		return calls
	}

	// If only one call, execute directly
	if len(calls) == 1 {
		return e.executeSingle(ctx, calls[0])
	}

	// Execute in parallel with semaphore
	return e.executeParallel(ctx, calls)
}

// executeSingle executes a single tool call.
func (e *Executor) executeSingle(ctx context.Context, call ToolCall) []ToolCall {
	start := time.Now()

	// Get the tool implementation
	// This assumes tools are registered in the global registry
	// For now, we'll mark as not implemented
	call.Error = fmt.Errorf("tool execution not implemented")
	call.Done = true

	duration := time.Since(start)
	errors.Info("parallel", fmt.Sprintf("tool %s completed in %v", call.Name, duration))

	return []ToolCall{call}
}

// executeParallel executes multiple tools in parallel.
func (e *Executor) executeParallel(ctx context.Context, calls []ToolCall) []ToolCall {
	sem := make(chan struct{}, e.maxConcurrency)
	var wg sync.WaitGroup
	results := make([]ToolCall, len(calls))

	for i, call := range calls {
		i, call := i, call // Capture loop variables

		wg.Add(1)
		go func(idx int, tc ToolCall) {
			defer wg.Done()

			// Acquire semaphore
			select {
			case sem <- struct{}{}:
				defer func() { <-sem }()
			case <-ctx.Done():
				tc.Error = ctx.Err()
				tc.Done = true
				results[idx] = tc
				return
			}

			// Execute tool
			start := time.Now()
			// TODO: Actually execute the tool
			// For now, simulate execution
			tc.Error = fmt.Errorf("tool execution not implemented")
			tc.Done = true
			duration := time.Since(start)

			errors.Debug("parallel", fmt.Sprintf("tool %s completed in %v", tc.Name, duration))
			results[idx] = tc
		}(i, call)
	}

	// Wait for all to complete
	wg.Wait()

	return results
}

// ExecuteWithDependencies executes tools with dependency resolution.
// Tools can depend on results from other tools.
func (e *Executor) ExecuteWithDependencies(ctx context.Context, calls []ToolCall, dependencies map[string][]string) []ToolCall {
	if len(dependencies) == 0 {
		return e.Execute(ctx, calls)
	}

	// Build dependency graph
	// Execute tools with no dependencies first
	// Then execute tools that depend on completed tools
	// This is a simplified topological sort

	results := make([]ToolCall, len(calls))
	completed := make(map[string]bool)
	var wg sync.WaitGroup
	var mu sync.Mutex

	// Keep executing until all are done
	for len(completed) < len(calls) {
		for i, call := range calls {
			if call.Done {
				continue
			}

			// Check if dependencies are met
			depsMet := true
			for _, dep := range dependencies[call.Name] {
				if !completed[dep] {
					depsMet = false
					break
				}
			}

			if !depsMet {
				continue
			}

			// Execute this tool
			i, call := i, call
			wg.Add(1)
			go func(idx int, tc ToolCall) {
				defer wg.Done()

				// Execute tool
				// TODO: Actual execution
				tc.Error = fmt.Errorf("tool execution not implemented")
				tc.Done = true

				mu.Lock()
				results[idx] = tc
				completed[tc.Name] = true
				mu.Unlock()
			}(i, call)
		}

		// Wait a bit for current batch
		wg.Wait()
	}

	return results
}

// BatchExecute executes tools in batches to avoid overwhelming systems.
func (e *Executor) BatchExecute(ctx context.Context, calls []ToolCall, batchSize int) []ToolCall {
	if batchSize <= 0 {
		batchSize = e.maxConcurrency
	}

	results := make([]ToolCall, 0, len(calls))

	for i := 0; i < len(calls); i += batchSize {
		end := i + batchSize
		if end > len(calls) {
			end = len(calls)
		}

		batch := calls[i:end]
		batchResults := e.Execute(ctx, batch)
		results = append(results, batchResults...)

		// Small delay between batches
		if end < len(calls) {
			select {
			case <-time.After(100 * time.Millisecond):
			case <-ctx.Done():
				break
			}
		}
	}

	return results
}

// ParallelExecutorConfig configures parallel execution.
type ParallelExecutorConfig struct {
	MaxConcurrency  int           // Maximum parallel tool executions
	Timeout         time.Duration // Maximum time to wait for all tools
	ContinueOnError bool          // Whether to continue if one tool fails
}

// DefaultParallelConfig returns default config.
func DefaultParallelConfig() *ParallelExecutorConfig {
	return &ParallelExecutorConfig{
		MaxConcurrency:  5,
		Timeout:         5 * time.Minute,
		ContinueOnError: true,
	}
}

// ExecuteWithConfig executes tools with custom configuration.
func (e *Executor) ExecuteWithConfig(ctx context.Context, calls []ToolCall, config *ParallelExecutorConfig) []ToolCall {
	if config == nil {
		config = DefaultParallelConfig()
	}

	// Create timeout context
	timeoutCtx, cancel := context.WithTimeout(ctx, config.Timeout)
	defer cancel()

	// Execute with max concurrency
	e.maxConcurrency = config.MaxConcurrency
	results := e.Execute(timeoutCtx, calls)

	return results
}

// ToolDependency defines a tool dependency relationship.
type ToolDependency struct {
	Tool      string
	DependsOn []string
}

// DependencyGraph represents tool dependencies.
type DependencyGraph struct {
	nodes map[string]*ToolDependency
}

// NewDependencyGraph creates a new dependency graph.
func NewDependencyGraph() *DependencyGraph {
	return &DependencyGraph{
		nodes: make(map[string]*ToolDependency),
	}
}

// Add adds a tool to the graph.
func (g *DependencyGraph) Add(tool string, dependsOn []string) {
	g.nodes[tool] = &ToolDependency{
		Tool:      tool,
		DependsOn: dependsOn,
	}
}

// Resolve returns the execution order (topological sort).
func (g *DependencyGraph) Resolve() ([]string, error) {
	order := make([]string, 0)
	visited := make(map[string]bool)
	temp := make(map[string]bool)

	var visit func(string) error
	visit = func(tool string) error {
		if temp[tool] {
			return fmt.Errorf("cyclic dependency detected at %s", tool)
		}
		if visited[tool] {
			return nil
		}

		temp[tool] = true

		node, ok := g.nodes[tool]
		if ok {
			for _, dep := range node.DependsOn {
				if err := visit(dep); err != nil {
					return err
				}
			}
		}

		delete(temp, tool)
		visited[tool] = true
		order = append(order, tool)
		return nil
	}

	for tool := range g.nodes {
		if err := visit(tool); err != nil {
			return nil, err
		}
	}

	return order, nil
}
