package v1

import "github.com/gogf/gf/v2/frame/g"

// ListDocumentsReq - 获取已索引文档列表请求
type ListDocumentsReq struct {
	g.Meta `path:"/documents" method:"get" summary:"获取已索引文档列表"`
	Page   int `json:"page" d:"1"`           // 页码
	Size   int `json:"size" d:"20"`          // 每页数量
}

// ListDocumentsRes - 获取已索引文档列表响应
type ListDocumentsRes struct {
	Documents []DocumentInfo `json:"documents"`
	Total     int64          `json:"total"`
	Page      int            `json:"page"`
	Size      int            `json:"size"`
}

// DocumentInfo - 文档信息
type DocumentInfo struct {
	ID       string                 `json:"id"`
	Source   string                 `json:"source"`
	Metadata map[string]any         `json:"metadata"`
	CreatedAt string                 `json:"created_at"`
}

// DeleteDocumentReq - 删除文档请求
type DeleteDocumentReq struct {
	g.Meta `path:"/documents/:id" method:"delete" summary:"删除文档"`
	ID     string `json:"id" v:"required#文档ID必填"`
}

// DeleteDocumentRes - 删除文档响应
type DeleteDocumentRes struct {
	Success bool   `json:"success"`
	Message string `json:"message"`
}

// SearchDocumentsReq - 向量搜索请求
type SearchDocumentsReq struct {
	g.Meta   `path:"/search" method:"post" summary:"向量搜索"`
	Query    string `json:"query" v:"required#搜索内容必填"`
	TopK     int    `json:"top_k" d:"5"`         // 返回结果数量
}

// SearchDocumentsRes - 向量搜索响应
type SearchDocumentsRes struct {
	Results []SearchResult `json:"results"`
	Count   int             `json:"count"`
}

// SearchResult - 搜索结果
type SearchResult struct {
	Content   string  `json:"content"`
	Score     float64 `json:"score"`
	Source    string  `json:"source"`
	Metadata  any     `json:"metadata"`
}
