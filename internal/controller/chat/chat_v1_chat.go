package chat

import (
	"github.com/WyRainBow/ops-portal/api/chat/v1"
	"github.com/WyRainBow/ops-portal/internal/ai/agent/plan_execute_replan"
	"context"
)

func (c *ControllerV1) Chat(ctx context.Context, req *v1.ChatReq) (res *v1.ChatRes, err error) {
	msg := req.Question
	// Use the Plan/Execute/Replan agent so the assistant can call observability tools.
	// This endpoint is intentionally read-only: tools do not execute commands or write DB.
	query := `
"你是一个只读的智能运维助手。"
"你必须在回答前按需调用工具获取事实（例如 Loki 日志、Prometheus 告警、只读数据库查询），避免凭空猜测。"
"你不能执行任何修复动作，只能输出诊断结论与建议的下一步命令（供人工执行）。"
"用户问题如下："
` + msg

	out, _, err := plan_execute_replan.BuildPlanAgent(ctx, query)
	if err != nil {
		return nil, err
	}
	res = &v1.ChatRes{
		Answer: out,
	}
	return res, nil
}
