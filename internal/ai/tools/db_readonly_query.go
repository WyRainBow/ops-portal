package tools

import (
	"context"
	"encoding/json"
	"regexp"
	"strings"

	"github.com/WyRainBow/ops-portal/internal/store"
	"github.com/cloudwego/eino/components/tool"
	"github.com/cloudwego/eino/components/tool/utils"
)

type DBReadonlyQueryInput struct {
	SQL string `json:"sql" jsonschema:"description=Readonly SQL query (SELECT/WITH only). Avoid selecting large tables without LIMIT."`
}

func NewDBReadonlyQueryTool() tool.InvokableTool {
	t, err := utils.InferOptionableTool(
		"db_readonly_query",
		"Run a readonly SQL query (SELECT/WITH) against the PostgreSQL database. Use this tool to inspect users, members, logs, traces. Writes are forbidden.",
		func(ctx context.Context, input *DBReadonlyQueryInput, opts ...tool.Option) (output string, err error) {
			sql := strings.TrimSpace(input.SQL)
			if sql == "" {
				return `{"success":false,"error":"sql is empty"}`, nil
			}
			if err := validateReadonlySQL(sql); err != nil {
				return `{"success":false,"error":"` + escapeJSON(err.Error()) + `"}`, nil
			}

			db, err := store.DB(ctx)
			if err != nil {
				return `{"success":false,"error":"db init failed"}`, nil
			}

			rows := make([]map[string]any, 0)
			// hard cap: if user didn't provide LIMIT, add a conservative one.
			if !hasLimit(sql) {
				sql = sql + " LIMIT 200"
			}
			if err := db.WithContext(ctx).Raw(sql).Scan(&rows).Error; err != nil {
				return `{"success":false,"error":"` + escapeJSON(err.Error()) + `"}`, nil
			}
			out := map[string]any{
				"success": true,
				"count":   len(rows),
				"rows":    rows,
			}
			b, _ := json.MarshalIndent(out, "", "  ")
			return string(b), nil
		},
	)
	if err != nil {
		panic(err)
	}
	return t
}

var (
	denyRE  = regexp.MustCompile(`(?i)\\b(insert|update|delete|drop|alter|truncate|create|grant|revoke)\\b`)
	semiRE  = regexp.MustCompile(`;`)
	allowTbl = []string{"users", "members", "api_request_logs", "api_error_logs", "api_trace_spans", "permission_audit_logs"}
)

func validateReadonlySQL(sql string) error {
	if semiRE.MatchString(sql) {
		return errf("semicolon is not allowed")
	}
	low := strings.ToLower(strings.TrimSpace(sql))
	if !(strings.HasPrefix(low, "select ") || strings.HasPrefix(low, "with ")) {
		return errf("only SELECT/WITH is allowed")
	}
	if denyRE.MatchString(sql) {
		return errf("write operations are forbidden")
	}
	ok := false
	for _, t := range allowTbl {
		if strings.Contains(low, t) {
			ok = true
			break
		}
	}
	if !ok {
		return errf("query must target an allowlisted table")
	}
	return nil
}

func hasLimit(sql string) bool {
	return regexp.MustCompile(`(?i)\\blimit\\b`).MatchString(sql)
}

type simpleErr string

func (e simpleErr) Error() string { return string(e) }

func errf(msg string) error { return simpleErr(msg) }

