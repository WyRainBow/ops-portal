package main

import (
	"context"
	"fmt"
	"os"
	"strings"

	"github.com/WyRainBow/ops-portal/internal/ai/models"
	"github.com/cloudwego/eino/schema"
)

func main() {
	ctx := context.Background()

	// Smoke test: call the configured DeepSeek model via Eino OpenAI-compatible client.
	// Configure via env:
	// - OPS_PORTAL_DS_QUICK_API_KEY
	// - OPS_PORTAL_DS_QUICK_BASE_URL
	// - OPS_PORTAL_DS_QUICK_MODEL
	if strings.TrimSpace(os.Getenv("OPS_PORTAL_DS_QUICK_API_KEY")) == "" {
		fmt.Println("Missing OPS_PORTAL_DS_QUICK_API_KEY. Example:")
		fmt.Println(`  export OPS_PORTAL_DS_QUICK_API_KEY="..."; export OPS_PORTAL_DS_QUICK_BASE_URL="https://ark.cn-beijing.volces.com/api/v3"; export OPS_PORTAL_DS_QUICK_MODEL="deepseek-v3-1-terminus"`)
		os.Exit(2)
	}
	chatModel, err := models.OpenAIForDeepSeekV3Quick(ctx)
	if err != nil {
		panic(err)
	}

	resp, err := chatModel.Generate(ctx, []*schema.Message{
		{Role: schema.System, Content: "You are a helpful SRE assistant. Reply in Chinese."},
		{Role: schema.User, Content: "用两句话解释什么是 Plan-Execute-Replan，并举一个告警排查例子。"},
	})
	if err != nil {
		panic(err)
	}
	// 输出结果
	fmt.Println(resp.Content)
}
