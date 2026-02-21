package models

import (
	"context"
	"os"
	"strings"

	"github.com/cloudwego/eino-ext/components/model/openai"
	"github.com/cloudwego/eino/components/model"
	"github.com/gogf/gf/v2/frame/g"
)

func OpenAIForDeepSeekV31Think(ctx context.Context) (cm model.ToolCallingChatModel, err error) {
	m, err := getCfgOrEnv(ctx, "ds_think_chat_model.model", "OPS_PORTAL_DS_THINK_MODEL")
	if err != nil {
		return nil, err
	}
	k, err := getCfgOrEnv(ctx, "ds_think_chat_model.api_key", "OPS_PORTAL_DS_THINK_API_KEY")
	if err != nil {
		return nil, err
	}
	u, err := getCfgOrEnv(ctx, "ds_think_chat_model.base_url", "OPS_PORTAL_DS_THINK_BASE_URL")
	if err != nil {
		return nil, err
	}
	config := &openai.ChatModelConfig{
		Model:   m,
		APIKey:  k,
		BaseURL: u,
	}
	cm, err = openai.NewChatModel(ctx, config)
	if err != nil {
		return nil, err
	}
	return cm, nil
}

func OpenAIForDeepSeekV3Quick(ctx context.Context) (cm model.ToolCallingChatModel, err error) {
	m, err := getCfgOrEnv(ctx, "ds_quick_chat_model.model", "OPS_PORTAL_DS_QUICK_MODEL")
	if err != nil {
		return nil, err
	}
	k, err := getCfgOrEnv(ctx, "ds_quick_chat_model.api_key", "OPS_PORTAL_DS_QUICK_API_KEY")
	if err != nil {
		return nil, err
	}
	u, err := getCfgOrEnv(ctx, "ds_quick_chat_model.base_url", "OPS_PORTAL_DS_QUICK_BASE_URL")
	if err != nil {
		return nil, err
	}
	config := &openai.ChatModelConfig{
		Model:   m,
		APIKey:  k,
		BaseURL: u,
	}
	cm, err = openai.NewChatModel(ctx, config)
	if err != nil {
		return nil, err
	}
	return cm, nil
}

func getCfgOrEnv(ctx context.Context, cfgKey string, envKey string) (string, error) {
	// Env wins so Docker/.env can override without editing config files.
	if envKey != "" {
		if v := strings.TrimSpace(os.Getenv(envKey)); v != "" {
			return v, nil
		}
	}
	v, err := g.Cfg().Get(ctx, cfgKey)
	if err != nil {
		return "", err
	}
	return strings.TrimSpace(v.String()), nil
}

// OpenAIForDashScopeQwen creates a chat model using DashScope (Aliyun) Qwen via OpenAI compatible API
// DashScope provides OpenAI-compatible API at https://dashscope.aliyuncs.com/compatible-mode/v1
func OpenAIForDashScopeQwen(ctx context.Context) (cm model.ToolCallingChatModel, err error) {
	// Default to qwen-max for reasoning capabilities
	m, err := getCfgOrEnv(ctx, "dashscope_chat_model.model", "DASHSCOPE_MODEL")
	if err != nil {
		m = "qwen-max" // Default model
	}
	k, err := getCfgOrEnv(ctx, "dashscope_chat_model.api_key", "DASHSCOPE_API_KEY")
	if err != nil {
		return nil, err
	}
	// DashScope OpenAI compatible endpoint
	u, err := getCfgOrEnv(ctx, "dashscope_chat_model.base_url", "DASHSCOPE_BASE_URL")
	if err != nil {
		u = "https://dashscope.aliyuncs.com/compatible-mode/v1" // Default base URL
	}

	config := &openai.ChatModelConfig{
		Model:   m,
		APIKey:  k,
		BaseURL: u,
	}
	cm, err = openai.NewChatModel(ctx, config)
	if err != nil {
		return nil, err
	}
	return cm, nil
}
