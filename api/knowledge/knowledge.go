package knowledge

import "github.com/WyRainBow/ops-portal/api/knowledge"

// IKnowledgeV1 is the interface for knowledge v1 controller.
type IKnowledgeV1 interface {
	// Document management
	ListDocuments(ctx context.Context, req *knowledge.v1.ListDocumentsReq) (res *knowledge.v1.ListDocumentsRes, err error)
	DeleteDocument(ctx context.Context, req *knowledge.v1.DeleteDocumentReq) (res *knowledge.v1.DeleteDocumentRes, err error)

	// Search
	SearchDocuments(ctx context.Context, req *knowledge.v1.SearchDocumentsReq) (res *knowledge.v1.SearchDocumentsRes, err error)

	// Status
	UploadStatus(ctx context.Context, req *struct{}) (res *struct {
		FileDir string `json:"file_dir"`
		Exists  bool   `json:"exists"`
	}, err error)
}
