package retriever

import (
	"github.com/WyRainBow/ops-portal/internal/ai/embedder"
	"github.com/WyRainBow/ops-portal/internal/ai/reranker"
	"github.com/WyRainBow/ops-portal/utility/client"
	"github.com/WyRainBow/ops-portal/utility/common"
	"context"

	"github.com/cloudwego/eino-ext/components/retriever/milvus"
	"github.com/cloudwego/eino/components/retriever"
)

// NewMilvusRetriever creates a basic Milvus retriever without reranking.
func NewMilvusRetriever(ctx context.Context) (rtr retriever.Retriever, err error) {
	cli, err := client.NewMilvusClient(ctx)
	if err != nil {
		return nil, err
	}
	eb, err := embedder.DoubaoEmbedding(ctx)
	if err != nil {
		return nil, err
	}
	r, err := milvus.NewRetriever(ctx, &milvus.RetrieverConfig{
		Client:      cli,
		Collection:  common.MilvusCollectionName,
		VectorField: "vector",
		OutputFields: []string{
			"id",
			"content",
			"metadata",
		},
		TopK:      1,
		Embedding: eb,
	})
	if err != nil {
		return nil, err
	}
	return r, nil
}

// NewRerankingRetriever creates a retriever with cross-encoder reranking.
// It retrieves more documents from Milvus and reranks them for better relevance.
func NewRerankingRetriever(ctx context.Context) (retriever.Retriever, error) {
	// Create base retriever with higher TopK (for reranking)
	cli, err := client.NewMilvusClient(ctx)
	if err != nil {
		return nil, err
	}
	eb, err := embedder.DoubaoEmbedding(ctx)
	if err != nil {
		return nil, err
	}

	// Retrieve 3x more documents for reranking
	baseRetriever, err := milvus.NewRetriever(ctx, &milvus.RetrieverConfig{
		Client:      cli,
		Collection:  common.MilvusCollectionName,
		VectorField: "vector",
		OutputFields: []string{
			"id",
			"content",
			"metadata",
		},
		TopK:      5, // Retrieve more for reranking
		Embedding: eb,
	})
	if err != nil {
		return nil, err
	}

	// Create reranker
	r := reranker.NewCrossEncoderReranker(eb, 3) // Return top 3

	// Wrap retriever with reranker
	return reranker.NewRerankingRetriever(baseRetriever, r), nil
}
