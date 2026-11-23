package spa

import (
	"io/fs"
	"net/http"
	"path"
	"strings"
)

// clientRoutes lists every path the React router renders a real page for (see
// frontend/src/App.tsx). Unknown paths still get index.html so the SPA can render its
// "page not found" view, but with a 404 status so crawlers and link checkers see the truth.
// Keep this in sync when adding a route in App.tsx.
var clientRoutes = map[string]bool{
	"/":                         true,
	"/ward-plan":                true,
	"/monthly-challenges":       true,
	"/monthly-challenges/print": true,
	"/experiences":              true,
	"/love-share-invite":        true,
	"/admin":                    true,
	"/admin/forgot-password":    true,
	"/admin/reset-password":     true,
	"/admin/moderate":           true,
	"/admin/approvals":          true,
	"/admin/ward-plan":          true,
}

// IsClientRoute reports whether the cleaned request path is a known SPA route. Matching is
// case-insensitive because React Router matches routes case-insensitively by default.
func IsClientRoute(cleanPath string) bool {
	return clientRoutes[strings.ToLower(cleanPath)]
}

// FileServer serves Vite static assets and falls back to index.html for client routing.
func FileServer(dist fs.FS, indexHTML []byte) http.Handler {
	fileServer := http.FileServer(http.FS(dist))
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet && r.Method != http.MethodHead {
			http.Error(w, "Not found", http.StatusNotFound)
			return
		}
		upath := path.Clean("/" + r.URL.Path)
		if upath == "/" || upath == "/." {
			serveIndexHTML(w, indexHTML, http.StatusOK)
			return
		}
		rel := strings.TrimPrefix(upath, "/")
		if rel == "" || rel == "." {
			serveIndexHTML(w, indexHTML, http.StatusOK)
			return
		}
		f, err := dist.Open(rel)
		if err == nil {
			_ = f.Close()
			fileServer.ServeHTTP(w, r)
			return
		}
		// Unknown API/auth endpoints and missing static files (anything with an extension)
		// get a plain 404 instead of the HTML app shell.
		if isAPIPath(upath) || path.Ext(upath) != "" {
			http.Error(w, "Not found", http.StatusNotFound)
			return
		}
		status := http.StatusOK
		if !IsClientRoute(upath) {
			status = http.StatusNotFound
		}
		serveIndexHTML(w, indexHTML, status)
	})
}

func isAPIPath(p string) bool {
	lower := strings.ToLower(p)
	return lower == "/api" || strings.HasPrefix(lower, "/api/") ||
		lower == "/auth" || strings.HasPrefix(lower, "/auth/")
}

func serveIndexHTML(w http.ResponseWriter, indexHTML []byte, status int) {
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	// Always revalidate so clients pick up new hashed JS/CSS after `npm run build` + Go rebuild.
	w.Header().Set("Cache-Control", "no-cache, must-revalidate")
	w.WriteHeader(status)
	_, _ = w.Write(indexHTML)
}
