package observability

import (
	"context"

	"github.com/WyRainBow/ops-portal/internal/ai/alerting"
	"github.com/gogf/gf/v2/frame/g"
	"github.com/gogf/gf/v2/net/ghttp"
)

// AlertWebhookController handles Prometheus Alertmanager webhooks.
type AlertWebhookController struct {
	queue *alerting.Queue
}

// NewAlertWebhookController creates a new alert webhook controller.
func NewAlertWebhookController() *AlertWebhookController {
	return &AlertWebhookController{
		queue: alerting.NewQueue(100), // Queue size 100
	}
}

// Webhook handles incoming Alertmanager webhooks.
// POST /api/observability/alerts/webhook
func (c *AlertWebhookController) Webhook(ctx context.Context, req *ghttp.Request) {
	// Read request body
	data := req.GetBody()
	if len(data) == 0 {
		req.Response.WriteJson(g.Map{
			"success": false,
			"error":   "Request body is empty",
		})
		req.Response.WriteStatus(400)
		return
	}

	// Process webhook
	incident, err := c.queue.ProcessWebhook(ctx, data)
	if err != nil {
		g.Log().Errorf(ctx, "Failed to process alert webhook: %v", err)
		req.Response.WriteJson(g.Map{
			"success": false,
			"error":   err.Error(),
		})
		req.Response.WriteStatus(500)
		return
	}

	// Check if diagnosis should be triggered
	ingester := alerting.NewIngester()
	if ingester.ShouldTriggerDiagnosis(incident) {
		g.Log().Infof(ctx, "Alert %s qualifies for AI diagnosis", incident.ID)
		// TODO: Trigger AI diagnosis asynchronously
		// This could be done by:
		// 1. Publishing to a message queue (Redis, RabbitMQ)
		// 2. Sending to a background worker
		// 3. Calling the Plan-Execute-Replan agent
	}

	// Return success
	req.Response.WriteJson(g.Map{
		"success": true,
		"incident_id": incident.ID,
		"alert_name": incident.AlertName,
		"severity": incident.Severity,
		"message": "Alert received and queued",
	})
}

// Status returns the current status of the alert queue.
// GET /api/observability/alerts/status
func (c *AlertWebhookController) Status(ctx context.Context, req *ghttp.Request) {
	req.Response.WriteJson(g.Map{
		"success": true,
		"queue_size": c.queue.Size(),
		"queue_capacity": 100,
	})
}

// RegisterAlertWebhookRoutes registers alert webhook routes.
// This should be called from the router setup.
func RegisterAlertWebhookRoutes(group *ghttp.RouterGroup) {
	controller := NewAlertWebhookController()

	group.Group("/alerts", func(alertGroup *ghttp.RouterGroup) {
		alertGroup.POST("/webhook", controller.Webhook)
		alertGroup.GET("/status", controller.Status)
	})
}
