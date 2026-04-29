package controllers

import (
	"os"
	"path/filepath"
	"strings"

	"purecore/core"
)

// DocsController serves documentation files from the docs/ directory.
// It reads markdown files and returns their raw content for the frontend to render.
type DocsController struct{}

// GetDoc returns the content of a documentation file.
//
// Query parameters:
//   - locale: "en" or "zh" (default: "en")
//   - page: filename without extension, e.g. "README", "API", "CLI"
//
// Example: GET /api/v1/docs?locale=en&page=README
func (dc *DocsController) GetDoc(req *core.Request, res *core.Response) error {
	locale := req.Input("locale", "en")
	page := req.Input("page", "README")

	// Validate locale (only allow en and zh for security)
	if locale != "en" && locale != "zh" {
		return res.Error("Invalid locale. Supported: en, zh", 400)
	}

	// Prevent path traversal: only allow alphanumeric, hyphens, and underscores
	if !isSafePath(page) {
		return res.Error("Invalid page name", 400)
	}

	// Build file path: docs/{locale}/{page}.md
	filePath := filepath.Join("docs", locale, page+".md")

	content, err := os.ReadFile(filePath)
	if err != nil {
		if os.IsNotExist(err) {
			return res.NotFound("Documentation page not found")
		}
		return res.Error("Failed to read documentation file", 500)
	}

	return res.JSON(200, 0, "OK", map[string]interface{}{
		"locale":  locale,
		"page":    page,
		"content": string(content),
	})
}

// ListDocs returns a list of available documentation pages for a given locale.
//
// Query parameters:
//   - locale: "en" or "zh" (default: "en")
//
// Example: GET /api/v1/docs/list?locale=en
func (dc *DocsController) ListDocs(req *core.Request, res *core.Response) error {
	locale := req.Input("locale", "en")

	if locale != "en" && locale != "zh" {
		return res.Error("Invalid locale. Supported: en, zh", 400)
	}

	dirPath := filepath.Join("docs", locale)
	entries, err := os.ReadDir(dirPath)
	if err != nil {
		if os.IsNotExist(err) {
			return res.NotFound("Documentation directory not found")
		}
		return res.Error("Failed to list documentation files", 500)
	}

	pages := make([]map[string]string, 0)
	for _, entry := range entries {
		if !entry.IsDir() && strings.HasSuffix(entry.Name(), ".md") {
			name := strings.TrimSuffix(entry.Name(), ".md")
			pages = append(pages, map[string]string{
				"page":   name,
				"locale": locale,
			})
		}
	}

	return res.Success(pages)
}

// isSafePath checks if the page name contains only safe characters.
func isSafePath(name string) bool {
	for _, r := range name {
		if !((r >= 'a' && r <= 'z') || (r >= 'A' && r <= 'Z') || (r >= '0' && r <= '9') || r == '-' || r == '_') {
			return false
		}
	}
	return len(name) > 0
}
