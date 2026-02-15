package tools

import (
	"context"
	"encoding/json"
	"os"
	"path/filepath"
	"sort"
	"strings"

	"github.com/cloudwego/eino/components/tool"
	"github.com/cloudwego/eino/components/tool/utils"
	"github.com/gogf/gf/v2/frame/g"
	"github.com/gogf/gf/v2/os/gfile"
)

type QueryInternalDocsInput struct {
	Query string `json:"query" jsonschema:"description=The query string to search in internal documentation for relevant information and processing steps"`
}

type docHit struct {
	Path    string `json:"path"`
	Score   int    `json:"score"`
	Snippet string `json:"snippet"`
}

// NewQueryInternalDocsTool searches local markdown/runbook files.
// v2 can re-enable Milvus-based retriever; v1 prioritizes local availability and low ops cost.
func NewQueryInternalDocsTool() tool.InvokableTool {
	t, err := utils.InferOptionableTool(
		"query_internal_docs",
		"Search internal documentation/runbooks for relevant information. Returns top matches with snippets. Use this tool before proposing operational actions.",
		func(ctx context.Context, input *QueryInternalDocsInput, opts ...tool.Option) (output string, err error) {
			q := strings.TrimSpace(input.Query)
			if q == "" {
				return `{"success":false,"error":"query is empty"}`, nil
			}

			root := os.Getenv("OPS_PORTAL_DOCS_DIR")
			if root == "" {
				// compatibility: reuse existing file_dir config used by knowledge index pipeline.
				v, _ := g.Cfg().Get(ctx, "file_dir")
				root = strings.TrimSpace(v.String())
			}
			if root == "" {
				root = "docs"
			}
			// If relative, resolve against cwd.
			if !filepath.IsAbs(root) {
				root = filepath.Clean(root)
			}

			hits := searchDocs(root, q, 8)
			out := map[string]any{
				"success": true,
				"query":   q,
				"root":    root,
				"hits":    hits,
			}
			b, _ := json.MarshalIndent(out, "", "  ")
			return string(b), nil
		})
	if err != nil {
		panic(err)
	}
	return t
}

func searchDocs(root string, query string, limit int) []docHit {
	q := strings.ToLower(query)
	paths := gfile.ScanDirFile(root, "*.md", true)
	if len(paths) == 0 {
		// also scan txt
		paths = append(paths, gfile.ScanDirFile(root, "*.txt", true)...)
	}
	hits := make([]docHit, 0)
	for _, p := range paths {
		b := gfile.GetBytes(p)
		if len(b) == 0 {
			continue
		}
		txt := string(b)
		low := strings.ToLower(txt)
		idx := strings.Index(low, q)
		if idx < 0 {
			continue
		}
		// score: more occurrences -> higher
		score := strings.Count(low, q)
		sn := snippet(txt, idx, 240)
		hits = append(hits, docHit{Path: p, Score: score, Snippet: sn})
	}
	sort.SliceStable(hits, func(i, j int) bool { return hits[i].Score > hits[j].Score })
	if len(hits) > limit {
		hits = hits[:limit]
	}
	return hits
}

func snippet(s string, idx int, max int) string {
	start := idx - max/3
	if start < 0 {
		start = 0
	}
	end := start + max
	if end > len(s) {
		end = len(s)
	}
	seg := s[start:end]
	seg = strings.ReplaceAll(seg, "\r\n", "\n")
	seg = strings.TrimSpace(seg)
	return seg
}

