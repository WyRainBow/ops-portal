# Ops Portal Access (Localhost-Only)

## 端口规划（建议）
- Portal (nginx): `127.0.0.1:18080`
- Web (Next): `127.0.0.1:18082`
- API (GoFrame): `127.0.0.1:18081`
- Grafana: `127.0.0.1:3000`
- Loki: `127.0.0.1:3100`
- Prometheus: `127.0.0.1:9090`
- node-exporter: `127.0.0.1:9100`

## SSH 隧道（从本机访问服务器 localhost 服务）
Portal:
```bash
ssh -i ~/.ssh/id_rsa -p 2222 -L 18080:127.0.0.1:18080 root@106.53.113.137
```

Grafana:
```bash
ssh -i ~/.ssh/id_rsa -p 2222 -L 3000:127.0.0.1:3000 root@106.53.113.137
```

## 运行（服务器）
1. 复制 `deploy/.env.example` 为 `deploy/.env` 并填好 `OPS_PORTAL_DB_DSN`、`OPS_PORTAL_JWT_SECRET`。
2. 生成 BasicAuth:
```bash
apk add --no-cache apache2-utils || true
htpasswd -bc deploy/htpasswd ops change-me
```
3. 启动:
```bash
cd deploy
docker compose up -d --build
```

