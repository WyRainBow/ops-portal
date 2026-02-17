package registry

import (
	"context"
	"fmt"
	"strings"
	"sync"

	"github.com/WyRainBow/ops-portal/internal/ai/errors"
	"github.com/WyRainBow/ops-portal/internal/ai/tools"
	"github.com/cloudwego/eino/components/tool"
)

// ToolMetadata holds metadata about a registered tool.
type ToolMetadata struct {
	Name        string
	Description string
	Category    string   // e.g., "observability", "database", "utility"
	Enabled     bool     // Whether the tool is currently enabled
	AgentTypes  []string // Which agent types can use this tool: "chat", "plan_execute", "all"
}

// ToolWrapper wraps a tool with its metadata.
type ToolWrapper struct {
	Tool     tool.BaseTool
	Metadata ToolMetadata
}

// Registry manages tool registration and lookup.
type Registry struct {
	mu    sync.RWMutex
	tools map[string]*ToolWrapper
}

// Global registry instance.
var globalRegistry = &Registry{
	tools: make(map[string]*ToolWrapper),
}

// Global returns the global tool registry.
func Global() *Registry {
	return globalRegistry
}

// Register registers a tool with the given metadata.
// If a tool with the same name already exists, it will be replaced.
func (r *Registry) Register(t tool.BaseTool, metadata ToolMetadata) {
	r.mu.Lock()
	defer r.mu.Unlock()

	// Infer name from tool if not provided
	if metadata.Name == "" {
		if inv, ok := t.(tool.InvokableTool); ok {
			// Try to get info from the tool
			metadata.Name = fmt.Sprintf("%T", inv)
		} else {
			metadata.Name = fmt.Sprintf("%T", t)
		}
	}

	r.tools[metadata.Name] = &ToolWrapper{
		Tool:     t,
		Metadata: metadata,
	}

	errors.Info("registry", fmt.Sprintf("registered tool: %s (category=%s, enabled=%v)", metadata.Name, metadata.Category, metadata.Enabled))
}

// Unregister removes a tool from the registry.
func (r *Registry) Unregister(name string) {
	r.mu.Lock()
	defer r.mu.Unlock()

	if _, ok := r.tools[name]; ok {
		delete(r.tools, name)
		errors.Info("registry", fmt.Sprintf("unregistered tool: %s", name))
	}
}

// Get retrieves a tool by name.
// Returns nil if the tool is not found or is disabled.
func (r *Registry) Get(name string) tool.BaseTool {
	r.mu.RLock()
	defer r.mu.RUnlock()

	if wrapper, ok := r.tools[name]; ok {
		if wrapper.Metadata.Enabled {
			return wrapper.Tool
		}
	}
	return nil
}

// GetAll returns all enabled tools for a given agent type.
// If agentType is "all" or empty, returns all enabled tools.
func (r *Registry) GetAll(agentType string) []tool.BaseTool {
	r.mu.RLock()
	defer r.mu.RUnlock()

	var result []tool.BaseTool
	for _, wrapper := range r.tools {
		if !wrapper.Metadata.Enabled {
			continue
		}
		// Check if tool is compatible with the agent type
		if agentType == "" || agentType == "all" {
			result = append(result, wrapper.Tool)
			continue
		}
		for _, allowedType := range wrapper.Metadata.AgentTypes {
			if allowedType == "all" || allowedType == agentType {
				result = append(result, wrapper.Tool)
				break
			}
		}
	}
	return result
}

// GetMetadata returns the metadata for a tool.
func (r *Registry) GetMetadata(name string) (ToolMetadata, bool) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	if wrapper, ok := r.tools[name]; ok {
		return wrapper.Metadata, true
	}
	return ToolMetadata{}, false
}

// ListMetadata returns all tool metadata.
func (r *Registry) ListMetadata() []ToolMetadata {
	r.mu.RLock()
	defer r.mu.RUnlock()

	result := make([]ToolMetadata, 0, len(r.tools))
	for _, wrapper := range r.tools {
		result = append(result, wrapper.Metadata)
	}
	return result
}

// Enable enables a tool by name.
func (r *Registry) Enable(name string) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	if wrapper, ok := r.tools[name]; ok {
		wrapper.Metadata.Enabled = true
		errors.Info("registry", fmt.Sprintf("enabled tool: %s", name))
		return nil
	}
	return fmt.Errorf("tool not found: %s", name)
}

// Disable disables a tool by name.
func (r *Registry) Disable(name string) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	if wrapper, ok := r.tools[name]; ok {
		wrapper.Metadata.Enabled = false
		errors.Info("registry", fmt.Sprintf("disabled tool: %s", name))
		return nil
	}
	return fmt.Errorf("tool not found: %s", name)
}

// FilterByCategory returns all enabled tools in a category.
func (r *Registry) FilterByCategory(category string) []tool.BaseTool {
	r.mu.RLock()
	defer r.mu.RUnlock()

	var result []tool.BaseTool
	for _, wrapper := range r.tools {
		if wrapper.Metadata.Enabled && wrapper.Metadata.Category == category {
			result = append(result, wrapper.Tool)
		}
	}
	return result
}

// RegisterCategory registers multiple tools at once with the same category.
func (r *Registry) RegisterCategory(category string, tools map[string]tool.BaseTool, agentTypes []string) {
	for name, t := range tools {
		r.Register(t, ToolMetadata{
			Name:       name,
			Category:   category,
			Enabled:    true,
			AgentTypes: agentTypes,
		})
	}
}

// GetByPrefix returns all tools whose name starts with the given prefix.
func (r *Registry) GetByPrefix(prefix string) []tool.BaseTool {
	r.mu.RLock()
	defer r.mu.RUnlock()

	var result []tool.BaseTool
	for name, wrapper := range r.tools {
		if wrapper.Metadata.Enabled && strings.HasPrefix(name, prefix) {
			result = append(result, wrapper.Tool)
		}
	}
	return result
}

// Count returns the number of registered tools (including disabled).
func (r *Registry) Count() int {
	r.mu.RLock()
	defer r.mu.RUnlock()

	return len(r.tools)
}

// CountEnabled returns the number of enabled tools.
func (r *Registry) CountEnabled() int {
	r.mu.RLock()
	defer r.mu.RUnlock()

	count := 0
	for _, wrapper := range r.tools {
		if wrapper.Metadata.Enabled {
			count++
		}
	}
	return count
}

// RegisterStandardTools registers all standard ops-portal tools.
// This is a convenience function to register all built-in tools.
func RegisterStandardTools(ctx context.Context) error {
	// Register observability tools
	registry := Global()

	// Log query tool
	logTool := tools.NewLokiQueryRangeTool()
	registry.Register(logTool, ToolMetadata{
		Name:       "query_loki_logs",
		Category:   "observability",
		Enabled:    true,
		AgentTypes: []string{"chat", "plan_execute", "all"},
	})

	// Prometheus alerts tool
	promTool := tools.NewPrometheusAlertsQueryTool()
	registry.Register(promTool, ToolMetadata{
		Name:       "query_prometheus_alerts",
		Category:   "observability",
		Enabled:    true,
		AgentTypes: []string{"chat", "plan_execute", "all"},
	})

	// Database tools
	dbTool := tools.NewDBReadonlyQueryTool()
	registry.Register(dbTool, ToolMetadata{
		Name:       "db_readonly_query",
		Category:   "database",
		Enabled:    true,
		AgentTypes: []string{"chat", "all"},
	})

	// MySQL CRUD tool (use with caution)
	mysqlTool := tools.NewMysqlCrudTool()
	registry.Register(mysqlTool, ToolMetadata{
		Name:       "mysql_crud",
		Category:   "database",
		Enabled:    false, // Disabled by default for safety
		AgentTypes: []string{"plan_execute", "all"},
	})

	// Documentation tool
	docsTool := tools.NewQueryInternalDocsTool()
	registry.Register(docsTool, ToolMetadata{
		Name:       "query_internal_docs",
		Category:   "knowledge",
		Enabled:    true,
		AgentTypes: []string{"chat", "all"},
	})

	// Time tool
	timeTool := tools.NewGetCurrentTimeTool()
	registry.Register(timeTool, ToolMetadata{
		Name:       "get_current_time",
		Category:   "utility",
		Enabled:    true,
		AgentTypes: []string{"chat", "plan_execute", "all"},
	})

	errors.Info("registry", fmt.Sprintf("registered standard tools: total=%d, enabled=%d",
		registry.Count(), registry.CountEnabled()))

	return nil
}
