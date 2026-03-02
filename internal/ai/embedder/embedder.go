package embedder

import (
	"context"
	"log"
	"os"
	"strings"

	"github.com/cloudwego/eino-ext/components/embedding/dashscope"
	"github.com/cloudwego/eino/components/embedding"
	"github.com/gogf/gf/v2/frame/g"
)

func DoubaoEmbedding(ctx context.Context) (eb embedding.Embedder, err error) {
	// Try environment variable first, then config file
	model := strings.TrimSpace(os.Getenv("DOUBAO_EMBEDDING_MODEL"))
	if model == "" {
		v, err := g.Cfg().Get(ctx, "doubao_embedding_model.model")
		if err != nil {
			return nil, err
		}
		model = strings.TrimSpace(v.String())
	}

	apiKey := strings.TrimSpace(os.Getenv("DOUBAO_EMBEDDING_API_KEY"))
	if apiKey == "" {
		v, err := g.Cfg().Get(ctx, "doubao_embedding_model.api_key")
		if err != nil {
			return nil, err
		}
		apiKey = strings.TrimSpace(v.String())
	}

	dim := 2048
	embedder, err := dashscope.NewEmbedder(ctx, &dashscope.EmbeddingConfig{
		Model:      model,
		APIKey:     apiKey,
		Dimensions: &dim,
	})
	if err != nil {
		log.Printf("new embedder error: %v\n", err)
		return nil, err
	}
	return embedder, nil
}
