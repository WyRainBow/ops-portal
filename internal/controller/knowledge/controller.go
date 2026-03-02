package knowledge

import (
	"github.com/WyRainBow/ops-portal/api/knowledge/v1"
	"github.com/WyRainBow/ops-portal/utility/client"
	"github.com/WyRainBow/ops-portal/utility/common"
	"context"
	"fmt"

	"github.com/gogf/gf/v2/errors/gerror"
	"github.com/gogf/gf/v2/frame/g"
	"github.com/gogf/gf/v2/os/gfile"
)

// ControllerV1 handles knowledge base operations.
type ControllerV1 struct{}

// NewV1 creates a new knowledge controller.
func NewV1() *ControllerV1 {
	return &ControllerV1{}
}

// ListDocuments returns the list of indexed documents.
// GET /api/knowledge/documents
func (c *ControllerV1) ListDocuments(ctx context.Context, req *v1.ListDocumentsReq) (res *v1.ListDocumentsRes, err error) {
	cli, err := client.NewMilvusClient(ctx)
	if err != nil {
		return nil, gerror.Wrapf(err, "failed to create Milvus client")
	}

	// Query all documents
	expr := "id != \"\"" // Simple expression to get all
	queryResult, err := cli.Query(ctx, common.MilvusCollectionName, []string{}, expr, []string{"id"})
	if err != nil {
		return nil, gerror.Wrapf(err, "failed to query documents")
	}

	// Extract document info from results
	docs := make([]v1.DocumentInfo, 0)
	if len(queryResult) > 0 {
		// Process results to build document list
		// This is a simplified implementation - in production you'd parse all columns
		for i := 0; i < queryResult[0].Len(); i++ {
			id, _ := queryResult[0].GetAsString(i)
			docs = append(docs, v1.DocumentInfo{
				ID:       id,
				Source:   id, // Simplified - in production extract from metadata column
				Metadata: map[string]any{},
				CreatedAt: "",
			})
		}
	}

	res = &v1.ListDocumentsRes{
		Documents: docs,
		Total:     int64(len(docs)),
		Page:      req.Page,
		Size:      req.Size,
	}
	return res, nil
}

// DeleteDocument deletes a document from the knowledge base.
// DELETE /api/knowledge/documents/:id
func (c *ControllerV1) DeleteDocument(ctx context.Context, req *v1.DeleteDocumentReq) (res *v1.DeleteDocumentRes, err error) {
	cli, err := client.NewMilvusClient(ctx)
	if err != nil {
		return nil, gerror.Wrapf(err, "failed to create Milvus client")
	}

	// Delete by ID
	expr := fmt.Sprintf(`id == "%s"`, req.ID)
	err = cli.Delete(ctx, common.MilvusCollectionName, "", expr)
	if err != nil {
		return nil, gerror.Wrapf(err, "failed to delete document")
	}

	res = &v1.DeleteDocumentRes{
		Success: true,
		Message: fmt.Sprintf("Document %s deleted successfully", req.ID),
	}
	return res, nil
}

// SearchDocuments performs vector search on the knowledge base.
// POST /api/knowledge/search
func (c *ControllerV1) SearchDocuments(ctx context.Context, req *v1.SearchDocumentsReq) (res *v1.SearchDocumentsRes, err error) {
	// This is a simplified implementation
	// In production, you would:
	// 1. Convert query to embedding vector
	// 2. Call Milvus search API
	// 3. Return formatted results

	// For now, return empty results with a note
	res = &v1.SearchDocumentsRes{
		Results: []v1.SearchResult{},
		Count:   0,
	}
	g.Log().Infof(ctx, "Vector search requested for query: %s (top_k=%d)", req.Query, req.TopK)
	g.Log().Warningf(ctx, "Vector search requires embedding model - returning empty results")

	return res, nil
}

// UploadStatus returns the upload and indexing status.
// GET /api/knowledge/status
func (c *ControllerV1) UploadStatus(ctx context.Context, req *struct{}) (res *struct {
	FileDir string `json:"file_dir"`
	Exists  bool   `json:"exists"`
}, err error) {
	fileDir := common.FileDir
	exists := gfile.Exists(fileDir)

	res = &struct {
		FileDir string `json:"file_dir"`
		Exists  bool   `json:"exists"`
	}{
		FileDir: fileDir,
		Exists:  exists,
	}
	return res, nil
}
