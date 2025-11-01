package http

import (
	"net/http"
	"os"
	"path/filepath"
	"strings"
)

// StaticHandler serves built SPA assets from a directory with sensible caching.
// - Files under /assets/ are considered hashed and served with long cache headers.
// - Requests that don't match a file fall back to index.html with no-store.
type StaticHandler struct {
	rootDir string
}

func NewStaticHandler(rootDir string) StaticHandler {
	return StaticHandler{rootDir: rootDir}
}

func (h StaticHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet && r.Method != http.MethodHead {
		http.NotFound(w, r)
		return
	}

	reqPath := r.URL.Path
	if reqPath == "/" {
		reqPath = "/index.html"
	}

	absRoot, _ := filepath.Abs(h.rootDir)
	relPath := strings.TrimPrefix(filepath.Clean(reqPath), string(filepath.Separator))
	absPath := filepath.Join(absRoot, relPath)

	// Prevent directory traversal
	if !strings.HasPrefix(absPath, absRoot) {
		http.NotFound(w, r)
		return
	}

	// Serve file if it exists, otherwise fall back to index.html
	if fileExists(absPath) {
		if filepath.Base(absPath) == "index.html" {
			setNoStore(w)
		} else {
			setCacheHeaders(w, reqPath)
		}
		http.ServeFile(w, r, absPath)
		return
	}

	indexPath := filepath.Join(absRoot, "index.html")
	if !fileExists(indexPath) {
		http.NotFound(w, r)
		return
	}
	setNoStore(w)
	http.ServeFile(w, r, indexPath)
}

func fileExists(path string) bool {
	info, err := os.Stat(path)
	return err == nil && !info.IsDir()
}

func setCacheHeaders(w http.ResponseWriter, reqPath string) {
	if isHashedAsset(reqPath) {
		w.Header().Set("Cache-Control", "public, max-age=31536000, immutable")
		return
	}
	// Default: no explicit cache header for other static files
}

func setNoStore(w http.ResponseWriter) {
	w.Header().Set("Cache-Control", "no-store")
}

func isHashedAsset(p string) bool {
	// Simple rule: anything under /assets/ gets long cache
	// Vite outputs hashed assets under this directory
	if p == "" {
		return false
	}
	return strings.HasPrefix(p, "/assets/") || strings.HasPrefix(p, "assets/")
}
