# ops-portal

`ops-portal` 是为 `Resume-Agent` 提供运维可观测与智能诊断能力的统一门户。

它聚合了三类能力：

1. 可观测入口：统一查询 Prometheus / Loki / Grafana，查看运行状态、错误日志、告警与趋势。
2. 管理门户：面向运维与管理员提供接口清单、运行状态、日志与追踪视图。
3. LLM 运维助手：基于 CloudWeGo Eino 的对话与 AIOps 流程，输出诊断结论与处置建议。

## 面向 Resume-Agent 的能力

### 1) 可观测

- 接入 `resume-backend` 的日志流（Loki）
- 接入宿主机与服务指标（Prometheus + node-exporter）
- 集成 Grafana Dashboard（运行状态、错误趋势、资源监控）
- 在门户内可快速定位接口报错、请求链路与服务异常

### 2) Eino 运维助手

- 保留并演进 Go 侧 Agent 能力（Chat / SSE / AIOps）
- 支持“告警分析 -> 日志检索 -> 指标对照 -> 报告输出”的诊断链路
- 默认只读运维策略，避免高风险自动写操作

### 3) 接口治理

- 展示 `ops-portal` 与 `resume-agent` 的 OpenAPI 接口清单
- 按项目来源、方法、标签和关键词筛选
- 用于联调、回归和发布前自查

## 技术栈

- Backend: GoFrame + CloudWeGo Eino
- Web: Next.js
- Observability: Prometheus / Loki / Grafana / node-exporter

## 目标定位

`ops-portal` 不承载 Resume 业务功能本身，而是专注于：

- 运行可见
- 异常可查
- 告警可分析
- 决策可追溯

并为后续飞书告警推送、Agent 协同修复提供基础。
