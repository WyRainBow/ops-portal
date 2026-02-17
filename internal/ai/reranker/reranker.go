package reranker

import (
	"context"
	"fmt"
	"sort"

	"github.com/WyRainBow/ops-portal/internal/ai/errors"
	"github.com/cloudwego/eino/components/embedding"
	"github.com/cloudwego/eino/components/retriever"
	"github.com/cloudwego/eino/schema"
)

// Reranker is an interface for reranking retrieved documents.
type Reranker interface {
	Rerank(ctx context.Context, query string, docs []*schema.Document) ([]*schema.Document, error)
}

// Score is a document with its relevance score.
type Score struct {
	Document *schema.Document
	Score    float64
}

// CrossEncoderReranker uses cross-encoder scoring for reranking.
// It scores each (query, document) pair and reorders by relevance.
type CrossEncoderReranker struct {
	embedder embedding.Embedder
	topK     int
}

// NewCrossEncoderReranker creates a new cross-encoder reranker.
// For v1, we use a simple embedding-based similarity score as proxy.
// A true cross-encoder would require a specialized model.
func NewCrossEncoderReranker(embedder embedding.Embedder, topK int) *CrossEncoderReranker {
	if topK <= 0 {
		topK = 5 // Default topK
	}
	return &CrossEncoderReranker{
		embedder: embedder,
		topK:     topK,
	}
}

// Rerank rescores and reorders documents based on query relevance.
func (r *CrossEncoderReranker) Rerank(ctx context.Context, query string, docs []*schema.Document) ([]*schema.Document, error) {
	if len(docs) == 0 {
		return docs, nil
	}

	errors.Debug("reranker", fmt.Sprintf("reranking %d documents for query: %s", len(docs), query))

	// Get query embedding
	embeddings, err := r.embedder.EmbedStrings(ctx, []string{query})
	if err != nil {
		errors.Error("reranker", "failed to embed query", err)
		// Return original docs on error
		return docs, nil
	}
	if len(embeddings) == 0 {
		return docs, nil
	}
	queryEmbedding := embeddings[0]

	// Score each document
	scores := make([]Score, 0, len(docs))
	for _, doc := range docs {
		score := r.computeScore(queryEmbedding, doc)
		scores = append(scores, Score{
			Document: doc,
			Score:    score,
		})
	}

	// Sort by score descending
	sort.SliceStable(scores, func(i, j int) bool {
		return scores[i].Score > scores[j].Score
	})

	// Return topK
	result := make([]*schema.Document, 0, r.topK)
	for i := 0; i < len(scores) && i < r.topK; i++ {
		doc := scores[i].Document
		// Add score to metadata for debugging
		if doc.MetaData == nil {
			doc.MetaData = make(map[string]any)
		}
		doc.MetaData["rerank_score"] = scores[i].Score
		result = append(result, doc)
	}

	errors.Info("reranker", fmt.Sprintf("reranked: returned %d of %d documents", len(result), len(docs)))

	return result, nil
}

// computeScore computes similarity between query embedding and document.
// For v1, we use cosine similarity as a proxy.
// A true cross-encoder would use a dedicated model.
func (r *CrossEncoderReranker) computeScore(queryEmbedding []float64, doc *schema.Document) float64 {
	// Try to get document embedding from metadata using DenseVector method
	docEmbedding := doc.DenseVector()
	if docEmbedding == nil {
		// Try getting from MetaData directly
		if vec, ok := doc.MetaData["vector"].([]float64); ok {
			docEmbedding = vec
		}
	}

	if docEmbedding == nil {
		// Fallback: if no embedding stored, give neutral score
		return 0.5
	}

	return cosineSimilarity(queryEmbedding, docEmbedding)
}

// cosineSimilarity computes cosine similarity between two vectors.
func cosineSimilarity(a, b []float64) float64 {
	if len(a) != len(b) {
		return 0
	}

	var dotProduct, normA, normB float64
	for i := range a {
		dotProduct += a[i] * b[i]
		normA += a[i] * a[i]
		normB += b[i] * b[i]
	}

	if normA == 0 || normB == 0 {
		return 0
	}

	return dotProduct / (sqrt(normA) * sqrt(normB))
}

// sqrt computes square root.
func sqrt(x float64) float64 {
	// Simple Newton's method
	if x == 0 {
		return 0
	}
	z := x
	for i := 0; i < 10; i++ {
		z = (z + x/z) / 2
	}
	return z
}

// RerankingRetriever wraps a retriever with a reranker.
type RerankingRetriever struct {
	retriever retriever.Retriever
	reranker  Reranker
}

// NewRerankingRetriever creates a retriever that reranks results.
func NewRerankingRetriever(baseRetriever retriever.Retriever, reranker Reranker) retriever.Retriever {
	return &RerankingRetriever{
		retriever: baseRetriever,
		reranker:  reranker,
	}
}

// Retrieve retrieves documents and reranks them.
func (r *RerankingRetriever) Retrieve(ctx context.Context, query string, opts ...retriever.Option) ([]*schema.Document, error) {
	// Parse options to get original TopK
	originalOpts := retriever.GetCommonOptions(nil, opts...)

	// Determine original TopK
	originalTopK := 5
	if originalOpts.TopK != nil {
		originalTopK = *originalOpts.TopK
	}

	rerankTopK := originalTopK * 3 // Retrieve 3x for reranking

	// Create new options with increased TopK
	rerankOpts := []retriever.Option{
		retriever.WithTopK(rerankTopK),
	}

	// Retrieve documents
	docs, err := r.retriever.Retrieve(ctx, query, rerankOpts...)
	if err != nil {
		return nil, err
	}

	// Rerank
	reranked, err := r.reranker.Rerank(ctx, query, docs)
	if err != nil {
		return nil, err
	}

	// Return requested TopK
	if len(reranked) > originalTopK {
		reranked = reranked[:originalTopK]
	}

	return reranked, nil
}
