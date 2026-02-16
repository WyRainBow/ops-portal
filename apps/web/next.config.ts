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
    ];
  },
};

export default nextConfig;
