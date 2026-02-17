import type { NextConfig } from "next";

const apiBase = process.env.OPS_PORTAL_API_BASE || "http://127.0.0.1:18081";

const nextConfig: NextConfig = {
  output: "standalone",
  async rewrites() {
    // In dev (or standalone Next), proxy /api/* to Go API to avoid CORS.
    // In prod behind nginx, /api can also be routed to Go directly; this stays harmless.
    return [
      {
        source: "/api/:path*",
        destination: `${apiBase}/api/:path*`,
      },
      // 代理 /swagger 到后端 Swagger UI，使「打开 Swagger」按钮可用
      { source: "/swagger", destination: `${apiBase}/swagger` },
      { source: "/swagger/", destination: `${apiBase}/swagger/` },
      { source: "/swagger/:path*", destination: `${apiBase}/swagger/:path*` },
      // 代理 OpenAPI 规范，ReDoc/Swagger 会从当前域名请求 /api.json
      { source: "/api.json", destination: `${apiBase}/api.json` },
    ];
  },
};

export default nextConfig;
