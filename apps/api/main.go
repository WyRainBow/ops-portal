package main

import (
	"os"
	"strconv"
	"strings"

	"github.com/WyRainBow/ops-portal/internal/controller/admin"
	"github.com/WyRainBow/ops-portal/internal/controller/auth"
	"github.com/WyRainBow/ops-portal/internal/controller/chat"
	"github.com/WyRainBow/ops-portal/internal/controller/observability"
	"github.com/WyRainBow/ops-portal/utility/middleware"

	"github.com/gogf/gf/v2/frame/g"
	"github.com/gogf/gf/v2/net/ghttp"
	"github.com/gogf/gf/v2/os/glog"
)

func main() {
	g.Log().SetLevel(glog.LEVEL_WARN)

	s := g.Server()

	// Reduce noisy startup/access output for local ops use.
	s.SetDumpRouterMap(false)
	s.SetAccessLogEnabled(false)
	s.SetLogLevel("warning")

	s.Group("/api", func(group *ghttp.RouterGroup) {
		group.Middleware(middleware.CORSMiddleware)
		group.Middleware(middleware.ResponseMiddleware)

		group.Bind(chat.NewV1())
		group.Bind(auth.NewV1())
		group.Bind(admin.NewV1())
		group.Bind(observability.NewV1())
	})

	port := 18081
	if v := os.Getenv("OPS_PORTAL_API_PORT"); v != "" {
		if p, err := strconv.Atoi(v); err == nil && p > 0 {
			port = p
		}
	}

	addr := os.Getenv("OPS_PORTAL_API_ADDR")
	if strings.TrimSpace(addr) == "" {
		addr = "127.0.0.1"
	}
	listen := addr
	if !strings.Contains(addr, ":") {
		listen = addr + ":" + strconv.Itoa(port)
	}
	s.SetAddr(listen)
	s.Run()
}
