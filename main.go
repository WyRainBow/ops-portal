package main

import (
	"github.com/WyRainBow/ops-portal/internal/ai/alerting"
	"github.com/WyRainBow/ops-portal/internal/ai/registry"
	"github.com/WyRainBow/ops-portal/internal/cache"
	"github.com/WyRainBow/ops-portal/internal/config"
	"github.com/WyRainBow/ops-portal/internal/controller/admin"
	"github.com/WyRainBow/ops-portal/internal/controller/auth"
	"github.com/WyRainBow/ops-portal/internal/controller/chat"
	"github.com/WyRainBow/ops-portal/internal/controller/observability"
	"github.com/WyRainBow/ops-portal/internal/controller/ops"
	"github.com/WyRainBow/ops-portal/internal/metrics"
	"github.com/WyRainBow/ops-portal/internal/ops/playbook"
	"github.com/WyRainBow/ops-portal/utility/common"
	"github.com/WyRainBow/ops-portal/utility/middleware"

	"github.com/gogf/gf/v2/frame/g"
	"github.com/gogf/gf/v2/net/ghttp"
	"github.com/gogf/gf/v2/os/gctx"
)

func main() {
	ctx := gctx.New()

	// Validate critical configuration
	if err := config.ValidateCritical(ctx); err != nil {
		g.Log().Fatalf(ctx, "Configuration validation failed: %v", err)
	}

	// Initialize metrics
	metrics.InitializeStandardMetrics()

	// Initialize cache (in-memory by default)
	// TODO: Add Redis support when configured
	_ = cache.Global()

	// Initialize alert store
	alerting.InitStore()

	// Initialize playbook executor
	playbook.InitExecutor()

	// Initialize tool registry
	// This must be done before any agent that uses tools
	if err := registry.RegisterStandardTools(ctx); err != nil {
		g.Log().Errorf(ctx, "Failed to register standard tools: %v", err)
		// Continue anyway - some tools may still be available
	}

	fileDir, err := g.Cfg().Get(ctx, "file_dir")
	if err != nil {
		panic(err)
	}
	common.FileDir = fileDir.String()
	s := g.Server()

	// Global middleware for all /api routes
	s.Group("/api", func(group *ghttp.RouterGroup) {
		group.Middleware(middleware.CORSMiddleware)

		// Auth endpoints - skip JWT auth
		group.Group("/auth", func(authGroup *ghttp.RouterGroup) {
			authGroup.Middleware(middleware.ResponseMiddleware)
			authGroup.Bind(
				auth.NewV1(),
			)
		})

		// Chat endpoints - require JWT auth
		group.Group("/chat", func(chatGroup *ghttp.RouterGroup) {
			chatGroup.Middleware(middleware.JWTAuth(nil))
			chatGroup.Middleware(middleware.ResponseMiddleware)
			chatGroup.Bind(chat.NewV1())
		})

		// Admin endpoints - require admin role
		group.Group("/admin", func(adminGroup *ghttp.RouterGroup) {
			adminGroup.Middleware(middleware.AdminAuth())
			adminGroup.Middleware(middleware.ResponseMiddleware)
			adminGroup.Bind(admin.NewV1())
		})

		// Observability endpoints - require admin or member role
		group.Group("/observability", func(obsGroup *ghttp.RouterGroup) {
			obsGroup.Middleware(middleware.JWTAuth(nil))
			// Role checking is done in handlers using RequireAnyRole
			obsGroup.Middleware(middleware.ResponseMiddleware)
			obsGroup.Bind(observability.NewV1())

			// Register alert webhook routes
			observability.RegisterAlertWebhookRoutes(obsGroup)
		})

		// Ops endpoints - require admin role for playbook execution
		group.Group("/ops", func(opsGroup *ghttp.RouterGroup) {
			opsGroup.Middleware(middleware.AdminAuth())
			opsGroup.Middleware(middleware.ResponseMiddleware)
			ops.RegisterOpsRoutes(opsGroup)
		})
	})

	s.SetPort(6872)
	s.Run()
}
