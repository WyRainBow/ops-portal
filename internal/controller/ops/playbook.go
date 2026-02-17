package ops

import (
	"context"

	"github.com/WyRainBow/ops-portal/internal/ops/playbook"
	"github.com/WyRainBow/ops-portal/utility/middleware"
	"github.com/gogf/gf/v2/frame/g"
	"github.com/gogf/gf/v2/net/ghttp"
)

// PlaybookController handles playbook operations.
type PlaybookController struct{}

// ListPlaybooks lists all available playbooks.
// GET /api/ops/playbooks
func (c *PlaybookController) ListPlaybooks(ctx context.Context, req *ghttp.Request) {
	playbooks := playbook.GlobalExecutor().List()

	req.Response.WriteJson(g.Map{
		"success": true,
		"playbooks": playbooks,
		"count":    len(playbooks),
	})
}

// GetPlaybook retrieves a specific playbook.
// GET /api/ops/playbooks/:id
func (c *PlaybookController) GetPlaybook(ctx context.Context, req *ghttp.Request) {
	id := req.Get("id").String()

	pb, ok := playbook.GlobalExecutor().Get(id)
	if !ok {
		req.Response.WriteJson(g.Map{
			"success": false,
			"error":   "Playbook not found",
		})
		req.Response.WriteStatus(404)
		return
	}

	req.Response.WriteJson(g.Map{
		"success":  true,
		"playbook": pb,
	})
}

// ExecuteRequest is the request body for playbook execution.
type ExecuteRequest struct {
	PlaybookID string                 `json:"playbook_id" v:"required#Playbook ID is required"`
	Parameters map[string]any         `json:"parameters"`
	Reason     string                 `json:"reason" v:"required#Reason is required for audit"`
	DryRun     bool                   `json:"dry_run"`
}

// ExecutePlaybook executes a playbook.
// POST /api/ops/playbooks/:id/execute
func (c *PlaybookController) ExecutePlaybook(ctx context.Context, req *ghttp.Request) {
	var input ExecuteRequest
	if err := req.Parse(&input); err != nil {
		req.Response.WriteJson(g.Map{
			"success": false,
			"error":   err.Error(),
		})
		req.Response.WriteStatus(400)
		return
	}

	// Get user from context
	user := middleware.GetUserContext(ctx)
	if user == nil {
		req.Response.WriteJson(g.Map{
			"success": false,
			"error":   "Unauthorized",
		})
		req.Response.WriteStatus(401)
		return
	}

	// Execute playbook
	result, err := playbook.ExecuteSafe(
		ctx,
		input.PlaybookID,
		input.Parameters,
		user.Username,
		input.Reason,
		input.DryRun,
	)

	if err != nil {
		req.Response.WriteJson(g.Map{
			"success": false,
			"error":   err.Error(),
		})
		req.Response.WriteStatus(500)
		return
	}

	req.Response.WriteJson(g.Map{
		"success":   true,
		"execution": result,
	})
}

// GetExecution retrieves an execution result.
// GET /api/ops/executions/:id
func (c *PlaybookController) GetExecution(ctx context.Context, req *ghttp.Request) {
	id := req.Get("id").String()

	result, ok := playbook.GlobalExecutor().GetExecution(id)
	if !ok {
		req.Response.WriteJson(g.Map{
			"success": false,
			"error":   "Execution not found",
		})
		req.Response.WriteStatus(404)
		return
	}

	req.Response.WriteJson(g.Map{
		"success":   true,
		"execution": result,
	})
}

// GetAuditLog retrieves the audit log.
// GET /api/ops/audit/log
func (c *PlaybookController) GetAuditLog(ctx context.Context, req *ghttp.Request) {
	log := playbook.GlobalExecutor().GetAuditLog()

	req.Response.WriteJson(g.Map{
		"success": true,
		"audit_log": log,
		"count":    len(log),
	})
}

// RegisterOpsRoutes registers ops routes.
func RegisterOpsRoutes(group *ghttp.RouterGroup) {
	controller := &PlaybookController{}

	group.Group("/ops", func(opsGroup *ghttp.RouterGroup) {
		// Require admin role for all ops operations
		opsGroup.Middleware(middleware.AdminAuth())

		// Playbook management
		opsGroup.GET("/playbooks", controller.ListPlaybooks)
		opsGroup.GET("/playbooks/:id", controller.GetPlaybook)
		opsGroup.POST("/playbooks/:id/execute", controller.ExecutePlaybook)

		// Execution tracking
		opsGroup.GET("/executions/:id", controller.GetExecution)

		// Audit log
		opsGroup.GET("/audit/log", controller.GetAuditLog)
	})
}
