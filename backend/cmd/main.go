package main

import (
	"context"
	"crypto/sha256"
	"fmt"
	"io/fs"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/gorilla/csrf"
	"github.com/joho/godotenv"

	"github.com/albertson/albertsonmissionplan/internal/auth"
	"github.com/albertson/albertsonmissionplan/internal/handlers"
	"github.com/albertson/albertsonmissionplan/internal/spa"
	"github.com/albertson/albertsonmissionplan/internal/store"
	"github.com/albertson/albertsonmissionplan/web"
	"golang.org/x/oauth2"
)

// loadDotenv applies .env files from the current working directory up toward the filesystem root.
// Later files override earlier ones (so a repo-root .env wins over backend/.env).
// Uses Overload (not Load) so values in .env replace variables already set in the shell/IDE.
func loadDotenv() {
	wd, err := os.Getwd()
	if err != nil {
		return
	}
	var paths []string
	for d := wd; d != filepath.Dir(d); d = filepath.Dir(d) {
		p := filepath.Join(d, ".env")
		if st, err := os.Stat(p); err == nil && !st.IsDir() {
			paths = append(paths, p)
		}
	}
	for _, p := range paths {
		if err := godotenv.Overload(p); err != nil {
			log.Printf("warning: could not load %s: %v", p, err)
		}
	}
}

func main() {
	loadDotenv()
	ctx := context.Background()

	dbPath := strings.TrimSpace(os.Getenv("DATABASE_PATH"))
	if dbPath == "" {
		dbPath = "wardmission.db"
	}
	db, err := store.Open(ctx, dbPath)
	if err != nil {
		log.Fatalf("database %s: %v", dbPath, err)
	}
	defer db.Close()

	sessionSecret := strings.TrimSpace(os.Getenv("SESSION_SECRET"))
	if len(sessionSecret) < 16 {
		log.Fatal("SESSION_SECRET must be set to a random string (at least 16 characters)")
	}
	csrfKey := sha256.Sum256([]byte(sessionSecret))

	skipOAuth := strings.EqualFold(strings.TrimSpace(os.Getenv("DEV_SKIP_OAUTH")), "true")
	var oauthCfg *oauth2.Config
	if skipOAuth {
		log.Println("warning: DEV_SKIP_OAUTH=true — Google sign-in is disabled (public site and APIs work; leader routes will not authenticate)")
	} else {
		googleID := strings.TrimSpace(os.Getenv("GOOGLE_CLIENT_ID"))
		googleSecret := strings.TrimSpace(os.Getenv("GOOGLE_CLIENT_SECRET"))
		redirect := strings.TrimSpace(os.Getenv("OAUTH_REDIRECT_URL"))
		if googleID == "" || googleSecret == "" || redirect == "" {
			log.Fatal(`Set GOOGLE_CLIENT_ID, GOOGLE_CLIENT_SECRET, and OAUTH_REDIRECT_URL, or use DEV_SKIP_OAUTH=true for local dev without leader sign-in. See Google Cloud Console → APIs & Services → Credentials (OAuth 2.0 Web client).`)
		}
		oauthCfg = auth.NewGoogleOAuth(googleID, googleSecret, redirect)
		log.Println("Google OAuth enabled for leader sign-in")
	}

	cookieStore := auth.CookieStoreFromSecret(sessionSecret)
	cookieSecure := strings.EqualFold(os.Getenv("COOKIE_SECURE"), "true") ||
		strings.HasPrefix(strings.TrimSpace(os.Getenv("PUBLIC_BASE_URL")), "https://")
	auth.SetCookieSecure(cookieStore, cookieSecure)

	allow := auth.ParseAllowlist(os.Getenv("ALLOWED_LEADER_EMAILS"))
	if len(allow) == 0 {
		log.Println("warning: ALLOWED_LEADER_EMAILS is empty — no one can use leader tools until you add comma-separated Google emails")
	}

	ac := &auth.Config{
		OAuth:   oauthCfg,
		Store:   cookieStore,
		Allowed: allow,
	}

	srv := &handlers.Server{DB: db, Auth: ac}

	mux := http.NewServeMux()

	mux.HandleFunc("GET /api/bootstrap", srv.APIBootstrap)
	mux.HandleFunc("GET /api/home/experiences", srv.APIHomeExperiences)
	mux.HandleFunc("GET /api/experiences", srv.APIExperiences)
	mux.HandleFunc("POST /api/experiences", srv.APIPostExperience)
	mux.HandleFunc("GET /api/ward-plan", srv.APIWardPlan)

	mux.HandleFunc("GET /api/admin/pending", srv.APIAdminPending)
	mux.HandleFunc("POST /api/admin/moderate", srv.APIAdminModerate)
	mux.HandleFunc("GET /api/admin/ward-plan", srv.APIAdminWardPlan)
	mux.HandleFunc("POST /api/admin/ward-plan", srv.APIAdminWardPlanSave)

	mux.HandleFunc("GET /auth/google", srv.GoogleStart)
	mux.HandleFunc("GET /auth/callback", srv.GoogleCallback)
	mux.HandleFunc("POST /auth/logout", srv.Logout)

	indexHTML, err := fs.ReadFile(web.Dist, "dist/index.html")
	if err != nil {
		log.Fatal(`web UI missing: run "cd frontend && npm install && npm run build" (outputs to backend/web/dist), then retry`)
	}
	distFS, err := fs.Sub(web.Dist, "dist")
	if err != nil {
		log.Fatal(err)
	}
	mux.Handle("/", spa.FileServer(distFS, indexHTML))

	handler := csrf.Protect(csrfKey[:], csrf.Secure(cookieSecure))(mux)

	addr := ":8080"
	if p := os.Getenv("PORT"); p != "" {
		addr = ":" + strings.TrimPrefix(strings.TrimSpace(p), ":")
	}

	server := &http.Server{
		Addr:              addr,
		Handler:           logRequest(handler),
		ReadHeaderTimeout: 10 * time.Second,
	}

	fmt.Fprintf(os.Stderr, "ward mission SPA + API on http://localhost%s (db %s)\n", addr, dbPath)
	fmt.Fprintf(os.Stderr, "dev: copy .env.example to .env at repo root; run frontend with \"cd frontend && npm run dev\"\n")
	if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
		log.Fatal(err)
	}
}

func logRequest(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()
		next.ServeHTTP(w, r)
		log.Printf("%s %s %s", r.Method, r.URL.Path, time.Since(start).Round(time.Millisecond))
	})
}
