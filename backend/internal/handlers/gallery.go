package handlers

import (
	"errors"
	"fmt"
	"io"
	"log"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/albertson/albertsonmissionplan/internal/gallery"
	"github.com/albertson/albertsonmissionplan/internal/store"
)

// Image processing allocates tens of MB for a big scan; run one at a time so a bulk upload
// can't exhaust memory on a small (512 MB) host.
var galleryProcessSlot = make(chan struct{}, 1)

// RegisterGalleryRoutes adds the public and leader gallery endpoints.
func (s *Server) RegisterGalleryRoutes(mux *http.ServeMux) {
	mux.HandleFunc("GET /api/gallery", s.APIGallery)
	mux.HandleFunc("GET /api/gallery/{id}/{size}", s.APIGalleryImage)
	mux.HandleFunc("GET /api/admin/gallery", s.APIAdminGallery)
	mux.HandleFunc("POST /api/admin/gallery", s.APIAdminGalleryUpload)
	mux.HandleFunc("POST /api/admin/gallery/{id}/visibility", s.APIAdminGalleryVisibility)
	mux.HandleFunc("POST /api/admin/gallery/{id}/delete", s.APIAdminGalleryDelete)
}

type galleryItemDTO struct {
	ID          int64  `json:"id"`
	Width       int    `json:"width"`
	Height      int    `json:"height"`
	ThumbWidth  int    `json:"thumbWidth"`
	ThumbHeight int    `json:"thumbHeight"`
	ImageURL    string `json:"imageUrl"`
	ThumbURL    string `json:"thumbUrl"`
	Alt         string `json:"alt"`
	Hidden      bool   `json:"hidden,omitempty"`
	When        string `json:"when"`
}

// galleryAlt is deliberately generic: the gallery stores no names, titles, or captions so
// nothing identifies the children who drew the pictures.
const galleryAlt = "Drawing by a ward member"

func toGalleryDTO(g store.GalleryItem) galleryItemDTO {
	return galleryItemDTO{
		ID:    g.ID,
		Width: g.Width, Height: g.Height, ThumbWidth: g.ThumbWidth, ThumbHeight: g.ThumbHeight,
		ImageURL: fmt.Sprintf("/api/gallery/%d/image?v=%s", g.ID, g.ImageHash),
		ThumbURL: fmt.Sprintf("/api/gallery/%d/thumb?v=%s", g.ID, g.ImageHash),
		Alt:      galleryAlt,
		Hidden:   g.Hidden,
		When:     formatWhen(g.CreatedAt),
	}
}

func galleryDTOs(items []store.GalleryItem) []galleryItemDTO {
	out := make([]galleryItemDTO, 0, len(items))
	for _, g := range items {
		out = append(out, toGalleryDTO(g))
	}
	return out
}

// APIGallery lists visible drawings for the public Ward Gallery page.
func (s *Server) APIGallery(w http.ResponseWriter, r *http.Request) {
	items, err := store.ListGalleryItems(r.Context(), s.DB, false)
	if err != nil {
		log.Printf("api gallery: %v", err)
		writeJSON(w, http.StatusInternalServerError, map[string]string{"error": "Could not load the gallery."})
		return
	}
	writeJSON(w, http.StatusOK, map[string]any{"items": galleryDTOs(items)})
}

// APIGalleryImage serves a drawing's display image or thumbnail from the database. URLs carry
// the content hash (?v=), so matching requests are cached for a year; hidden drawings are only
// served to leaders and never cached publicly.
func (s *Server) APIGalleryImage(w http.ResponseWriter, r *http.Request) {
	id, err := strconv.ParseInt(r.PathValue("id"), 10, 64)
	size := r.PathValue("size")
	if err != nil || id <= 0 || (size != "image" && size != "thumb") {
		http.NotFound(w, r)
		return
	}
	data, hash, hidden, err := store.GalleryImage(r.Context(), s.DB, id, size == "thumb")
	if errors.Is(err, store.ErrGalleryNotFound) {
		http.NotFound(w, r)
		return
	}
	if err != nil {
		log.Printf("gallery image %d: %v", id, err)
		http.Error(w, "Could not load image", http.StatusInternalServerError)
		return
	}
	if hidden && !s.isModerator(w, r) {
		http.NotFound(w, r)
		return
	}
	etag := `"` + hash + "-" + size + `"`
	w.Header().Set("Content-Type", "image/jpeg")
	w.Header().Set("X-Content-Type-Options", "nosniff")
	w.Header().Set("ETag", etag)
	switch {
	case hidden:
		w.Header().Set("Cache-Control", "private, no-store")
	case r.URL.Query().Get("v") == hash:
		w.Header().Set("Cache-Control", "public, max-age=31536000, immutable")
	default:
		w.Header().Set("Cache-Control", "public, max-age=300")
	}
	if match := r.Header.Get("If-None-Match"); match != "" && match == etag && !hidden {
		w.WriteHeader(http.StatusNotModified)
		return
	}
	w.Header().Set("Content-Length", strconv.Itoa(len(data)))
	w.WriteHeader(http.StatusOK)
	if r.Method != http.MethodHead {
		_, _ = w.Write(data)
	}
}

// isModerator reports whether the request has a signed-in leader session (no error response).
func (s *Server) isModerator(w http.ResponseWriter, r *http.Request) bool {
	email := strings.ToLower(strings.TrimSpace(s.Auth.SessionEmail(w, r)))
	if email == "" {
		return false
	}
	p, err := s.loadLeaderPriv(r.Context(), email)
	return err == nil && p.canModerate()
}

// APIAdminGallery lists every drawing (including hidden) for the manager page.
func (s *Server) APIAdminGallery(w http.ResponseWriter, r *http.Request) {
	if _, ok := s.requireAdmin(w, r); !ok {
		return
	}
	items, err := store.ListGalleryItems(r.Context(), s.DB, true)
	if err != nil {
		log.Printf("api admin gallery: %v", err)
		writeJSON(w, http.StatusInternalServerError, map[string]string{"error": "Could not load the gallery."})
		return
	}
	writeJSON(w, http.StatusOK, map[string]any{"items": galleryDTOs(items)})
}

// APIAdminGalleryUpload accepts one drawing per request (multipart field "image"; no names or
// other text are stored). The manager page uploads a batch one file at a time so each
// file gets its own progress and error.
func (s *Server) APIAdminGalleryUpload(w http.ResponseWriter, r *http.Request) {
	email, ok := s.requireAdmin(w, r)
	if !ok {
		return
	}
	// Allow the 15 MB image plus multipart overhead.
	r.Body = http.MaxBytesReader(w, r.Body, gallery.MaxUploadBytes+(1<<20))
	if err := r.ParseMultipartForm(8 << 20); err != nil {
		var mbe *http.MaxBytesError
		if errors.As(err, &mbe) {
			writeJSON(w, http.StatusRequestEntityTooLarge, map[string]string{"error": gallery.ErrTooLarge.Error()})
			return
		}
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "Could not read the upload."})
		return
	}
	defer func() {
		if r.MultipartForm != nil {
			_ = r.MultipartForm.RemoveAll()
		}
	}()
	f, _, err := r.FormFile("image")
	if err != nil {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "Choose an image to upload."})
		return
	}
	defer f.Close()
	data, err := io.ReadAll(io.LimitReader(f, gallery.MaxUploadBytes+1))
	if err != nil {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "Could not read the upload."})
		return
	}

	galleryProcessSlot <- struct{}{}
	start := time.Now()
	res, err := gallery.Process(data)
	<-galleryProcessSlot
	if err != nil {
		status := http.StatusUnsupportedMediaType
		switch {
		case errors.Is(err, gallery.ErrTooLarge):
			status = http.StatusRequestEntityTooLarge
		case errors.Is(err, gallery.ErrTooManyPx), errors.Is(err, gallery.ErrCorrupt):
			status = http.StatusUnprocessableEntity
		case errors.Is(err, gallery.ErrHEIC), errors.Is(err, gallery.ErrUnsupported):
		default:
			log.Printf("gallery process: %v", err)
			writeJSON(w, http.StatusInternalServerError, map[string]string{"error": "Could not process the image."})
			return
		}
		writeJSON(w, status, map[string]string{"error": err.Error()})
		return
	}

	id, err := store.InsertGalleryItem(r.Context(), s.DB, store.GalleryImages{
		Width: res.DisplayWidth, Height: res.DisplayHeight,
		ThumbWidth: res.ThumbWidth, ThumbHeight: res.ThumbHeight,
		Hash: res.Hash, Image: res.Display, Thumb: res.Thumb,
	}, email)
	if err != nil {
		log.Printf("gallery insert: %v", err)
		writeJSON(w, http.StatusInternalServerError, map[string]string{"error": "Could not save the drawing."})
		return
	}
	log.Printf("gallery upload %d by %s: %d KB in → %dx%d (%d KB) + thumb (%d KB) in %s", id, email,
		len(data)>>10, res.DisplayWidth, res.DisplayHeight, len(res.Display)>>10, len(res.Thumb)>>10,
		time.Since(start).Round(time.Millisecond))
	item, err := store.GetGalleryItem(r.Context(), s.DB, id)
	if err != nil {
		writeJSON(w, http.StatusOK, map[string]any{"ok": true, "id": id})
		return
	}
	writeJSON(w, http.StatusOK, map[string]any{"ok": true, "item": toGalleryDTO(*item)})
}

type galleryVisibilityRequest struct {
	Hidden bool `json:"hidden"`
}

func galleryIDFromPath(w http.ResponseWriter, r *http.Request) (int64, bool) {
	id, err := strconv.ParseInt(r.PathValue("id"), 10, 64)
	if err != nil || id <= 0 {
		writeJSON(w, http.StatusNotFound, map[string]string{"error": "Drawing not found."})
		return 0, false
	}
	return id, true
}

// APIAdminGalleryVisibility hides a drawing from (or shows it on) the public page: {"hidden": true}.
func (s *Server) APIAdminGalleryVisibility(w http.ResponseWriter, r *http.Request) {
	if _, ok := s.requireAdmin(w, r); !ok {
		return
	}
	id, ok := galleryIDFromPath(w, r)
	if !ok {
		return
	}
	var req galleryVisibilityRequest
	if err := readJSON(r, &req); err != nil {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "Invalid request."})
		return
	}
	if err := store.SetGalleryItemHidden(r.Context(), s.DB, id, req.Hidden); err != nil {
		if errors.Is(err, store.ErrGalleryNotFound) {
			writeJSON(w, http.StatusNotFound, map[string]string{"error": "Drawing not found."})
			return
		}
		log.Printf("gallery visibility %d: %v", id, err)
		writeJSON(w, http.StatusInternalServerError, map[string]string{"error": "Could not save."})
		return
	}
	item, err := store.GetGalleryItem(r.Context(), s.DB, id)
	if err != nil {
		writeJSON(w, http.StatusOK, map[string]any{"ok": true})
		return
	}
	writeJSON(w, http.StatusOK, map[string]any{"ok": true, "item": toGalleryDTO(*item)})
}

// APIAdminGalleryDelete permanently removes a drawing.
func (s *Server) APIAdminGalleryDelete(w http.ResponseWriter, r *http.Request) {
	if _, ok := s.requireAdmin(w, r); !ok {
		return
	}
	id, ok := galleryIDFromPath(w, r)
	if !ok {
		return
	}
	if err := store.DeleteGalleryItem(r.Context(), s.DB, id); err != nil {
		if errors.Is(err, store.ErrGalleryNotFound) {
			writeJSON(w, http.StatusNotFound, map[string]string{"error": "Drawing not found."})
			return
		}
		log.Printf("gallery delete %d: %v", id, err)
		writeJSON(w, http.StatusInternalServerError, map[string]string{"error": "Could not delete."})
		return
	}
	writeJSON(w, http.StatusOK, map[string]any{"ok": true})
}
