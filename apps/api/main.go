package main

import (
	"os"
	"strconv"

	"github.com/WyRainBow/ops-portal/internal/controller/admin"
	"github.com/WyRainBow/ops-portal/internal/controller/auth"
	"github.com/WyRainBow/ops-portal/internal/controller/chat"
	"github.com/WyRainBow/ops-portal/internal/controller/observability"
	"github.com/WyRainBow/ops-portal/utility/middleware"

	"github.com/gogf/gf/v2/frame/g"
	"github.com/gogf/gf/v2/net/ghttp"
)

func main() {
	// Keep GoFrame config compatibility. Prefer env overrides in production.
	s := g.Server()

	s.Group("/api", func(group *ghttp.RouterGroup) {
		group.Middleware(middleware.CORSMiddleware)
		group.Middleware(middleware.ResponseMiddleware)

		group.Bind(chat.NewV1())
		group.Bind(auth.NewV1())
		group.Bind(admin.NewV1())
		group.Bind(observability.NewV1())
	})

	// Default to localhost-only in production; bind is handled by reverse proxy.
	port := 18081
	if v := os.Getenv("OPS_PORTAL_API_PORT"); v != "" {
		if p, err := strconv.Atoi(v); err == nil && p > 0 {
			port = p
		}
	}
	s.SetPort(port)
	s.Run()
}

