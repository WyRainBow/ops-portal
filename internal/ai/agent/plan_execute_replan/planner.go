package plan_execute_replan

import (
	"github.com/WyRainBow/ops-portal/internal/ai/models"
	"context"

	"github.com/cloudwego/eino/adk"
	"github.com/cloudwego/eino/adk/prebuilt/planexecute"
)

func NewPlanner(ctx context.Context) (adk.Agent, error) {
	// Try DashScope Qwen first (Aliyun), fallback to DeepSeek
	planModel, err := models.OpenAIForDashScopeQwen(ctx)
	if err != nil {
		planModel, err = models.OpenAIForDeepSeekV31Think(ctx)
		if err != nil {
			return nil, err
		}
	}
	return planexecute.NewPlanner(ctx, &planexecute.PlannerConfig{
		ToolCallingChatModel: planModel,
	})
}
