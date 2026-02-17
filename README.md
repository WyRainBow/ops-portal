# ops-portal

`ops-portal` 是 `Resume-Agent` 的运维可观测与智能诊断门户。

它不是简历业务本身，而是用于回答三个问题：

1. 现在是否稳定（状态与指标）
2. 为什么报错（日志与链路）
3. 应该怎么处理（LLM 诊断报告与操作建议）

---

## 项目定位

基于 `.技术文档` 的设计，`ops-portal` 落地为三类 Agent 能力的一体化门户：

- 知识库 Agent：沉淀 runbook / 文档知识，支撑 RAG 检索。
- 对话 Agent：面向值班/排障的自然语言问答（ReAct + Tool）。
- 运维 Agent：面向告警处理的 Plan-Execute-Replan 诊断闭环。

当前目标是服务 `Resume-Agent`：统一可观测、可排障、可追踪。

---

## 面向 Resume-Agent 的核心能力

### 1) 可观测门户

- 统一接入 `resume-backend` 的日志、指标与仪表盘。
- 可查看运行状态、错误日志、接口日志、追踪链路。
- 支持从门户直接跳转 Grafana / Loki / Prometheus 查询路径。

### 2) LLM 运维诊断（Eino）

- 基于 CloudWeGo Eino 的 Chat / SSE / AIOps 能力。
- 运维 Agent 采用 Plan-Execute-Replan：
  - Planner：生成结构化排查计划
  - Executor：调用日志/指标工具执行步骤
  - Replanner：根据执行结果继续、调整或终止
- 输出可读“诊断报告”，用于值班决策和群内协作。

### 3) 接口治理

- 展示 `ops-portal` 与 `resume-agent` 两个项目的 OpenAPI 接口清单。
- 支持按项目、方法、标签、关键词筛选。
- 用于发布前联调、故障定位与接口资产盘点。

---

## 架构边界（生产建议）

`ops-portal` 按分层思路演进：

- L1（数据面）：可观测查询与诊断（只读）
- L2（通知面）：告警摘要推送（飞书等）
- L3（协作面）：群内 Agent 协同（OpenClaw / Claude Board）

默认原则：

- 先只读，再半自动，再自动修复。
- Web 层不直接放高危执行能力。
- 修复动作通过白名单 playbook + 审计 + 人工确认逐步放开。

---

## 技术栈

- Backend: GoFrame + CloudWeGo Eino
- Web: Next.js
- Observability: Prometheus / Loki / Grafana / node-exporter
- Data: PostgreSQL（管理侧读写）

