package errors

import (
	"encoding/json"
	"fmt"
	"log"
	"os"
	"runtime/debug"
	"time"
)

// ToolError represents a structured error from tool execution.
// Instead of panicking or calling log.Fatal, tools should return this
// as a JSON string that the Agent can understand.
type ToolError struct {
	Tool      string    `json:"tool"`
	Message   string    `json:"message"`
	Cause     string    `json:"cause,omitempty"`
	Timestamp time.Time `json:"timestamp"`
}

func (e ToolError) Error() string {
	if e.Cause != "" {
		return fmt.Sprintf("[%s] %s: %s", e.Tool, e.Message, e.Cause)
	}
	return fmt.Sprintf("[%s] %s", e.Tool, e.Message)
}

// ToJSON converts the error to a JSON string suitable for Agent consumption.
func (e ToolError) ToJSON() string {
	out := map[string]any{
		"success": false,
		"error":   e.Message,
		"tool":    e.Tool,
	}
	if e.Cause != "" {
		out["cause"] = e.Cause
	}
	b, _ := json.MarshalIndent(out, "", "  ")
	return string(b)
}

// NewToolError creates a new ToolError.
func NewToolError(tool, message string, cause error) ToolError {
	causeStr := ""
	if cause != nil {
		causeStr = cause.Error()
	}
	return ToolError{
		Tool:      tool,
		Message:   message,
		Cause:     causeStr,
		Timestamp: time.Now(),
	}
}

// ToolInitError is used when a tool fails to initialize.
// This should be logged but not panic the whole process.
type ToolInitError struct {
	Tool  string
	Cause error
}

func (e ToolInitError) Error() string {
	return fmt.Sprintf("tool init failed [%s]: %v", e.Tool, e.Cause)
}

// Logger provides structured logging for AI operations.
type Logger struct {
	prefix string
	logger *log.Logger
}

// Default logger instance.
var defaultLogger = &Logger{
	prefix: "ops-portal",
	logger: log.New(os.Stdout, "", 0),
}

// SetPrefix sets the log prefix.
func SetPrefix(p string) {
	defaultLogger.prefix = p
}

// Error logs an error with context.
func Error(tool, message string, err error) {
	defaultLogger.Error(tool, message, err)
}

// Error logs an error with context.
func (l *Logger) Error(tool, message string, err error) {
	cause := ""
	if err != nil {
		cause = err.Error()
	}
	l.logger.Printf("[ERROR] [%s] %s: %s", tool, message, cause)
}

// Warn logs a warning.
func Warn(tool, message string) {
	defaultLogger.Warn(tool, message)
}

// Warn logs a warning.
func (l *Logger) Warn(tool, message string) {
	l.logger.Printf("[WARN] [%s] %s", tool, message)
}

// Info logs an info message.
func Info(tool, message string) {
	defaultLogger.Info(tool, message)
}

// Info logs an info message.
func (l *Logger) Info(tool, message string) {
	l.logger.Printf("[INFO] [%s] %s", tool, message)
}

// Debug logs a debug message (only when DEBUG is set).
func Debug(tool, message string) {
	defaultLogger.Debug(tool, message)
}

// Debug logs a debug message.
func (l *Logger) Debug(tool, message string) {
	if os.Getenv("DEBUG") != "" {
		l.logger.Printf("[DEBUG] [%s] %s", tool, message)
	}
}

// ToolCall logs a tool invocation.
func ToolCall(tool string, input any) {
	defaultLogger.ToolCall(tool, input)
}

// ToolCall logs a tool invocation.
func (l *Logger) ToolCall(tool string, input any) {
	inputJSON, _ := json.Marshal(input)
	l.logger.Printf("[TOOL] [%s] input=%s", tool, string(inputJSON))
}

// ToolResult logs a tool result.
func ToolResult(tool string, success bool, duration time.Duration) {
	defaultLogger.ToolResult(tool, success, duration)
}

// ToolResult logs a tool result.
func (l *Logger) ToolResult(tool string, success bool, duration time.Duration) {
	status := "SUCCESS"
	if !success {
		status = "ERROR"
	}
	l.logger.Printf("[TOOL] [%s] %s duration=%s", tool, status, duration)
}

// RecoverPanic recovers from panic and logs it with stack trace.
// Returns the panic value as error, or nil if no panic occurred.
func RecoverPanic(tool string) (err error) {
	if r := recover(); r != nil {
		stack := debug.Stack()
		l := defaultLogger
		l.logger.Printf("[PANIC] [%s] recovered: %v\n%s", tool, r, string(stack))
		return fmt.Errorf("panic in %s: %v", tool, r)
	}
	return nil
}

// SafeToolWrapper wraps a tool function with panic recovery.
func SafeToolWrapper(toolName string, fn func() (string, error)) (output string, err error) {
	defer func() {
		if r := recover(); r != nil {
			toolErr := NewToolError(toolName, "panic during execution", fmt.Errorf("%v", r))
			output = toolErr.ToJSON()
			err = toolErr
			Error(toolName, "panic recovered", toolErr)
		}
	}()
	return fn()
}

// MustTool is like Must but logs instead of panicking.
// Use this when a tool MUST exist for the system to work,
// but you want to handle the error gracefully rather than crash.
func MustTool(toolName string, tool any, err error) any {
	if err != nil {
		Error(toolName, "tool init failed, returning nil tool", err)
		// Return nil instead of panicking
		return nil
	}
	return tool
}

// WrapError wraps an error with tool context.
func WrapError(tool, message string, err error) error {
	return NewToolError(tool, message, err)
}

// IsToolError checks if an error is a ToolError.
func IsToolError(err error) bool {
	_, ok := err.(ToolError)
	return ok
}
