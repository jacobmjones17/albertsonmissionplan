package spa

import (
	"io/fs"
	"net/http"
	"path"
	"strings"
)

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
			serveIndexHTML(w, indexHTML)
			return
		}
		rel := strings.TrimPrefix(upath, "/")
		if rel == "" || rel == "." {
			serveIndexHTML(w, indexHTML)
			return
		}
		f, err := dist.Open(rel)
		if err == nil {
			_ = f.Close()
			fileServer.ServeHTTP(w, r)
			return
		}
		serveIndexHTML(w, indexHTML)
	})
}

func serveIndexHTML(w http.ResponseWriter, indexHTML []byte) {
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	// Always revalidate so clients pick up new hashed JS/CSS after `npm run build` + Go rebuild.
	w.Header().Set("Cache-Control", "no-cache, must-revalidate")
	w.WriteHeader(http.StatusOK)
	_, _ = w.Write(indexHTML)
}
