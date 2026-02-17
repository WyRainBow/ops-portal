package tools

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"
	"time"

	"github.com/WyRainBow/ops-portal/internal/ai/errors"
	"github.com/cloudwego/eino/components/tool"
	"github.com/cloudwego/eino/components/tool/utils"
	"gorm.io/driver/mysql"
	"gorm.io/gorm"
)

// MySQLAutomatedInput is the input for automated MySQL operations.
// This tool is designed for AI automation - no interactive prompts.
type MySQLAutomatedInput struct {
	DSN         string `json:"dsn" jsonschema:"description=The Data Source Name for connecting to the MySQL database (user:pass@tcp(host:port)/dbname)"`
	SQL         string `json:"sql" jsonschema:"description=The SQL query to execute. For writes, must match whitelist tables."`
	OperateType string `json:"operate_type" jsonschema:"description=Operation type: query (SELECT), insert, update, or delete"`
	DryRun      bool   `json:"dry_run,omitempty" jsonschema:"description=If true, preview the operation without executing (for write operations). Default false."`
	Reason      string `json:"reason,omitempty" jsonschema:"description=Reason for this operation (for audit logging). Required for write operations."`
}

// MySQLResult is the result of a MySQL operation.
type MySQLResult struct {
	Success   bool   `json:"success"`
	Operation string `json:"operation,omitempty"`
	SQL       string `json:"sql,omitempty"`
	Rows      int64  `json:"rows_affected,omitempty"`
	Data      any    `json:"data,omitempty"`
	Error     string `json:"error,omitempty"`
	DryRun    bool   `json:"dry_run,omitempty"`
	Message   string `json:"message,omitempty"`
}

// Whitelist configuration for write operations.
// Only these tables can be modified through this tool.
var writeWhitelist = map[string][]string{
	"insert": {"ops_incidents", "ops_maintenance_windows", "ops_alert_annotations"},
	"update": {"ops_incidents", "ops_maintenance_windows", "ops_alert_annotations"},
	"delete": {"ops_temp_cache", "ops_alert_annotations"},
}

// readOnlyQueries are query prefixes that are always allowed.
var readOnlyQueries = []string{"select", "show", "describe", "explain", "with"}

// NewMysqlCrudTool creates a MySQL tool suitable for AI automation.
// Key differences from the old version:
// - No interactive stdin prompts
// - Dry-run mode for previewing write operations
// - Write whitelist for security
// - Structured error handling (no panics/log.Fatal)
// - Audit logging for all operations
func NewMysqlCrudTool() tool.InvokableTool {
	t, err := utils.InferOptionableTool(
		"mysql_crud",
		"Execute SQL queries against MySQL database. Supports SELECT queries and whitelisted INSERT/UPDATE/DELETE operations. Write operations require a 'reason' field for audit. Use dry_run=true to preview writes before executing.",
		func(ctx context.Context, input *MySQLAutomatedInput, opts ...tool.Option) (output string, err error) {
			start := time.Now()
			toolName := "mysql_crud"

			defer func() {
				duration := time.Since(start)
				success := err == nil && !strings.Contains(output, `"success":false`)
				errors.ToolResult(toolName, success, duration)
			}()

			errors.ToolCall(toolName, input)

			// Validate input
			if input.DSN == "" {
				result := MySQLResult{
					Success: false,
					Error:   "DSN is required",
				}
				return toJSON(result), nil
			}

			if input.SQL == "" {
				result := MySQLResult{
					Success: false,
					Error:   "SQL is required",
				}
				return toJSON(result), nil
			}

			input.OperateType = strings.ToLower(strings.TrimSpace(input.OperateType))
			if input.OperateType == "" {
				// Auto-detect operation type from SQL
				sqlLower := strings.ToLower(strings.TrimSpace(input.SQL))
				for _, prefix := range readOnlyQueries {
					if strings.HasPrefix(sqlLower, prefix+" ") || sqlLower == prefix {
						input.OperateType = "query"
						break
					}
				}
				if input.OperateType == "" {
					input.OperateType = "query" // Default to query
				}
			}

			// For write operations, require reason
			if isWriteOperation(input.OperateType) && input.Reason == "" && !input.DryRun {
				result := MySQLResult{
					Success: false,
					Error:   "Reason is required for write operations. Please provide a 'reason' field explaining why this operation is needed.",
				}
				return toJSON(result), nil
			}

			// Check whitelist for write operations
			if err := checkWriteWhitelist(input.OperateType, input.SQL); err != nil && !input.DryRun {
				result := MySQLResult{
					Success: false,
					Error:   err.Error(),
				}
				return toJSON(result), nil
			}

			// Connect to database
			db, err := gorm.Open(mysql.Open(input.DSN), &gorm.Config{
				// Disable foreign key constraints for automated operations
				DisableForeignKeyConstraintWhenMigrating: true,
			})
			if err != nil {
				errors.Error(toolName, "database connection failed", err)
				result := MySQLResult{
					Success: false,
					Error:   fmt.Sprintf("Database connection failed: %v", err),
				}
				return toJSON(result), nil
			}

			sqlDB, err := db.DB()
			if err != nil {
				errors.Error(toolName, "get sql db failed", err)
				result := MySQLResult{
					Success: false,
					Error:   fmt.Sprintf("Get underlying DB failed: %v", err),
				}
				return toJSON(result), nil
			}
			// Set reasonable timeouts
			sqlDB.SetMaxIdleConns(1)
			sqlDB.SetMaxOpenConns(1)
			sqlDB.SetConnMaxLifetime(5 * time.Minute)

			// Execute operation
			switch input.OperateType {
			case "query":
				return executeQuery(db, input, toolName)
			case "insert", "update", "delete":
				return executeWrite(db, input, toolName)
			default:
				result := MySQLResult{
					Success: false,
					Error:   fmt.Sprintf("Unsupported operation type: %s. Use: query, insert, update, or delete", input.OperateType),
				}
				return toJSON(result), nil
			}
		},
	)

	if err != nil {
		// Instead of panic, we wrap the creation with a safe wrapper
		// The error will be handled by the SafeToolWrapper in the factory
		errors.Error("mysql_crud", "tool creation failed", err)
		// Return a tool that always returns an error message
		return createErrorTool("mysql_crud", err)
	}

	return t
}

// createErrorTool returns a tool that always returns an error
func createErrorTool(name string, createErr error) tool.InvokableTool {
	t, _ := utils.InferOptionableTool(
		name,
		fmt.Sprintf("Error tool for %s - this tool failed to initialize", name),
		func(ctx context.Context, input any, opts ...tool.Option) (output string, err error) {
			result := MySQLResult{
				Success: false,
				Error:   fmt.Sprintf("Tool initialization failed: %v", createErr),
			}
			return toJSON(result), nil
		},
	)
	return t
}

// executeQuery executes a SELECT query and returns results
func executeQuery(db *gorm.DB, input *MySQLAutomatedInput, toolName string) (string, error) {
	// Add limit if not present for safety
	sql := input.SQL
	if !strings.Contains(strings.ToLower(sql), "limit") {
		sql += " LIMIT 500"
	}

	// Execute query
	rows := make([]map[string]any, 0)
	if err := db.Raw(sql).Scan(&rows).Error; err != nil {
		errors.Error(toolName, "query execution failed", err)
		result := MySQLResult{
			Success: false,
			SQL:     input.SQL,
			Error:   fmt.Sprintf("Query failed: %v", err),
		}
		return toJSON(result), nil
	}

	result := MySQLResult{
		Success:   true,
		Operation: "query",
		SQL:       input.SQL,
		Data:      rows,
		Message:   fmt.Sprintf("Query returned %d rows", len(rows)),
	}

	errors.Info(toolName, fmt.Sprintf("query returned %d rows", len(rows)))
	return toJSON(result), nil
}

// executeWrite executes INSERT/UPDATE/DELETE operations
func executeWrite(db *gorm.DB, input *MySQLAutomatedInput, toolName string) (string, error) {
	// Extract table name for audit
	tableName := extractTableName(input.SQL)

	// Dry run mode - just return what would happen
	if input.DryRun {
		result := MySQLResult{
			Success:   true,
			Operation: input.OperateType,
			SQL:       input.SQL,
			DryRun:    true,
			Message:   fmt.Sprintf("Dry run: would execute %s on table '%s'. Remove dry_run=true to execute.", input.OperateType, tableName),
		}
		return toJSON(result), nil
	}

	// Execute the write operation
	result := db.Exec(input.SQL)
	if result.Error != nil {
		errors.Error(toolName, fmt.Sprintf("%s execution failed", input.OperateType), result.Error)
		out := MySQLResult{
			Success:   false,
			Operation: input.OperateType,
			SQL:       input.SQL,
			Error:     fmt.Sprintf("Execution failed: %v", result.Error),
		}
		return toJSON(out), nil
	}

	// Log the audit trail
	errors.Info(toolName, fmt.Sprintf("%s executed on %s: %d rows affected. Reason: %s",
		input.OperateType, tableName, result.RowsAffected, input.Reason))

	out := MySQLResult{
		Success:   true,
		Operation: input.OperateType,
		SQL:       input.SQL,
		Rows:      result.RowsAffected,
		Message:   fmt.Sprintf("%s on '%s' affected %d rows", strings.ToUpper(input.OperateType), tableName, result.RowsAffected),
	}

	return toJSON(out), nil
}

// checkWriteWhitelist validates that write operations target allowed tables
func checkWriteWhitelist(opType, sql string) error {
	whitelist, ok := writeWhitelist[opType]
	if !ok || len(whitelist) == 0 {
		return fmt.Errorf("operation type '%s' is not whitelisted for writes", opType)
	}

	tableName := extractTableName(sql)
	if tableName == "" {
		return fmt.Errorf("could not extract table name from SQL")
	}

	for _, allowed := range whitelist {
		if tableName == allowed {
			return nil
		}
	}

	return fmt.Errorf("table '%s' is not whitelisted for %s operations. Allowed tables: %v",
		tableName, opType, whitelist)
}

// extractTableName extracts the table name from a SQL statement
func extractTableName(sql string) string {
	sql = strings.TrimSpace(sql)
	sqlLower := strings.ToLower(sql)

	// Remove leading keywords
	prefixes := []string{"insert into ", "update ", "delete from ", "select from ", "from "}
	for _, prefix := range prefixes {
		if strings.HasPrefix(sqlLower, prefix) {
			rest := strings.TrimSpace(sql[len(prefix):])
			// Extract table name (before space, comma, or semicolon)
			for i, ch := range rest {
				if ch == ' ' || ch == ',' || ch == ';' {
					return rest[:i]
				}
			}
			return rest
		}
	}

	return ""
}

// isWriteOperation checks if the operation type is a write operation
func isWriteOperation(opType string) bool {
	return opType == "insert" || opType == "update" || opType == "delete"
}

// toJSON converts a result to JSON string
func toJSON(result MySQLResult) string {
	b, _ := json.MarshalIndent(result, "", "  ")
	return string(b)
}
