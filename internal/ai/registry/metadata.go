package registry

import (
	"strings"
)

// ToolCategory represents a category of tools.
type ToolCategory string

const (
	CategoryObservability ToolCategory = "observability"
	CategoryDatabase      ToolCategory = "database"
	CategoryKnowledge     ToolCategory = "knowledge"
	CategoryUtility       ToolCategory = "utility"
	CategoryMCP           ToolCategory = "mcp"
	CategoryCustom        ToolCategory = "custom"
)

// AgentType represents the type of agent.
type AgentType string

const (
	AgentTypeChat        AgentType = "chat"
	AgentTypePlanExecute AgentType = "plan_execute"
	AgentTypeKnowledge   AgentType = "knowledge"
	AgentTypeAll         AgentType = "all"
)

// ToolConfig contains configuration for registering a tool.
type ToolConfig struct {
	Name        string
	Description string
	Category    ToolCategory
	AgentTypes  []AgentType
	Enabled     bool
}

// DefaultToolConfig returns a default tool config.
func DefaultToolConfig(name string) ToolConfig {
	return ToolConfig{
		Name:       name,
		Category:   CategoryUtility,
		AgentTypes: []AgentType{AgentTypeAll},
		Enabled:    true,
	}
}

// Validate validates the tool config.
func (c ToolConfig) Validate() error {
	if strings.TrimSpace(c.Name) == "" {
		return ErrInvalidToolName
	}
	return nil
}

// Errors
var (
	ErrInvalidToolName = &RegistryError{Code: "INVALID_NAME", Message: "tool name cannot be empty"}
	ErrToolNotFound    = &RegistryError{Code: "NOT_FOUND", Message: "tool not found"}
	ErrToolDisabled    = &RegistryError{Code: "DISABLED", Message: "tool is disabled"}
)

// RegistryError represents an error in the registry.
type RegistryError struct {
	Code    string
	Message string
}

func (e *RegistryError) Error() string {
	return e.Message
}

// IsNotFound checks if an error is a "not found" error.
func IsNotFound(err error) bool {
	return err == ErrToolNotFound
}

// IsDisabled checks if an error is a "disabled" error.
func IsDisabled(err error) bool {
	return err == ErrToolDisabled
}
