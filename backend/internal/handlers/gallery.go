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
	"unicode/utf8"

	"github.com/albertson/albertsonmissionplan/internal/gallery"
	"github.com/albertson/albertsonmissionplan/internal/store"
)

const (
	galleryTitleMax    = 120
	galleryCaptionMax  = 600
	galleryArtistMax   = 40
	galleryAgeGroupMax = 40
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
	mux.HandleFunc("POST /api/admin/gallery/reorder", s.APIAdminGalleryReorder)
	mux.HandleFunc("POST /api/admin/gallery/{id}/edit", s.APIAdminGalleryEdit)
	mux.HandleFunc("POST /api/admin/gallery/{id}/delete", s.APIAdminGalleryDelete)
}

type galleryItemDTO struct {
	ID          int64  `json:"id"`
	Title       string `json:"title"`
	Caption     string `json:"caption"`
	Artist      string `json:"artist"`
	AgeGroup    string `json:"ageGroup"`
	Width       int    `json:"width"`
	Height      int    `json:"height"`
	ThumbWidth  int    `json:"thumbWidth"`
	ThumbHeight int    `json:"thumbHeight"`
	ImageURL    string `json:"imageUrl"`
	ThumbURL    string `json:"thumbUrl"`
	Alt         string `json:"alt"`
	Featured    bool   `json:"featured"`
	Hidden      bool   `json:"hidden,omitempty"`
	When        string `json:"when"`
}

// galleryAlt builds descriptive alt text from what leaders typed.
func galleryAlt(g store.GalleryItem) string {
	var b strings.Builder
	b.WriteString("Drawing")
	if g.Title != "" {
		fmt.Fprintf(&b, ": %s", g.Title)
	}
	if g.Artist != "" {
		fmt.Fprintf(&b, ", by %s", g.Artist)
		if g.AgeGroup != "" {
			fmt.Fprintf(&b, " (%s)", g.AgeGroup)
		}
	} else if g.AgeGroup != "" {
		fmt.Fprintf(&b, ", by a member of %s", g.AgeGroup)
	}
	if g.Caption != "" {
		fmt.Fprintf(&b, ". %s", g.Caption)
	}
	return b.String()
}

func toGalleryDTO(g store.GalleryItem) galleryItemDTO {
	return galleryItemDTO{
		ID: g.ID, Title: g.Title, Caption: g.Caption, Artist: g.Artist, AgeGroup: g.AgeGroup,
		Width: g.Width, Height: g.Height, ThumbWidth: g.ThumbWidth, ThumbHeight: g.ThumbHeight,
		ImageURL: fmt.Sprintf("/api/gallery/%d/image?v=%s", g.ID, g.ImageHash),
		ThumbURL: fmt.Sprintf("/api/gallery/%d/thumb?v=%s", g.ID, g.ImageHash),
		Alt:      galleryAlt(g),
		Featured: g.Featured, Hidden: g.Hidden,
		When: formatWhen(g.CreatedAt),
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

func cleanField(s string, max int) string {
	s = strings.Join(strings.Fields(s), " ")
	if utf8.RuneCountInString(s) > max {
		s = string([]rune(s)[:max])
	}
	return strings.TrimSpace(s)
}

func cleanCaption(s string) string {
	s = strings.TrimSpace(strings.ReplaceAll(s, "\r\n", "\n"))
	if utf8.RuneCountInString(s) > galleryCaptionMax {
		s = strings.TrimSpace(string([]rune(s)[:galleryCaptionMax]))
	}
	return s
}

func cleanGalleryInput(in store.GalleryItemInput) store.GalleryItemInput {
	in.Title = cleanField(in.Title, galleryTitleMax)
	in.Caption = cleanCaption(in.Caption)
	in.Artist = cleanField(in.Artist, galleryArtistMax)
	in.AgeGroup = cleanField(in.AgeGroup, galleryAgeGroupMax)
	return in
}

// APIAdminGalleryUpload accepts one drawing per request (multipart field "image" plus optional
// title, caption, artist, ageGroup). The manager page uploads a batch one file at a time so each
// file gets its own progress and error.
func (s *Server) APIAdminGalleryUpload(w http.ResponseWriter, r *http.Request) {
	email, ok := s.requireAdmin(w, r)
	if !ok {
		return
	}
	// Allow the 15 MB image plus form fields and multipart overhead.
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

	in := cleanGalleryInput(store.GalleryItemInput{
		Title:    r.FormValue("title"),
		Caption:  r.FormValue("caption"),
		Artist:   r.FormValue("artist"),
		AgeGroup: r.FormValue("ageGroup"),
	})
	id, err := store.InsertGalleryItem(r.Context(), s.DB, in, store.GalleryImages{
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

type galleryEditRequest struct {
	Title    string `json:"title"`
	Caption  string `json:"caption"`
	Artist   string `json:"artist"`
	AgeGroup string `json:"ageGroup"`
	Featured bool   `json:"featured"`
	Hidden   bool   `json:"hidden"`
}

func galleryIDFromPath(w http.ResponseWriter, r *http.Request) (int64, bool) {
	id, err := strconv.ParseInt(r.PathValue("id"), 10, 64)
	if err != nil || id <= 0 {
		writeJSON(w, http.StatusNotFound, map[string]string{"error": "Drawing not found."})
		return 0, false
	}
	return id, true
}

// APIAdminGalleryEdit updates a drawing's text, featured, and hidden flags.
func (s *Server) APIAdminGalleryEdit(w http.ResponseWriter, r *http.Request) {
	if _, ok := s.requireAdmin(w, r); !ok {
		return
	}
	id, ok := galleryIDFromPath(w, r)
	if !ok {
		return
	}
	var req galleryEditRequest
	if err := readJSON(r, &req); err != nil {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "Invalid request."})
		return
	}
	in := cleanGalleryInput(store.GalleryItemInput(req))
	if err := store.UpdateGalleryItem(r.Context(), s.DB, id, in); err != nil {
		if errors.Is(err, store.ErrGalleryNotFound) {
			writeJSON(w, http.StatusNotFound, map[string]string{"error": "Drawing not found."})
			return
		}
		log.Printf("gallery edit %d: %v", id, err)
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

// APIAdminGalleryReorder saves the manager's drag/arrow order: {"ids": [3, 1, 2]}.
func (s *Server) APIAdminGalleryReorder(w http.ResponseWriter, r *http.Request) {
	if _, ok := s.requireAdmin(w, r); !ok {
		return
	}
	var req struct {
		IDs []int64 `json:"ids"`
	}
	if err := readJSON(r, &req); err != nil || len(req.IDs) > 5000 {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "Invalid request."})
		return
	}
	if err := store.ReorderGalleryItems(r.Context(), s.DB, req.IDs); err != nil {
		log.Printf("gallery reorder: %v", err)
		writeJSON(w, http.StatusInternalServerError, map[string]string{"error": "Could not save the order."})
		return
	}
	writeJSON(w, http.StatusOK, map[string]any{"ok": true})
}
