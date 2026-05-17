package main

import (
	"context"
	"crypto/sha256"
	"fmt"
	"io/fs"
	"log"
	"net"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/gorilla/csrf"
	"github.com/joho/godotenv"

	"github.com/albertson/albertsonmissionplan/internal/auth"
	"github.com/albertson/albertsonmissionplan/internal/handlers"
	"github.com/albertson/albertsonmissionplan/internal/mail"
	"github.com/albertson/albertsonmissionplan/internal/spa"
	"github.com/albertson/albertsonmissionplan/internal/store"
	"github.com/albertson/albertsonmissionplan/web"
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

	dbURL := strings.TrimSpace(os.Getenv("DATABASE_URL"))
	if dbURL == "" {
		log.Fatal("DATABASE_URL must be set (PostgreSQL connection string, e.g. from Render Postgres)")
	}
	db, err := store.Open(context.Background(), dbURL)
	if err != nil {
		log.Fatalf("database: %v", err)
	}
	defer db.Close()

	sessionSecret := strings.TrimSpace(os.Getenv("SESSION_SECRET"))
	if len(sessionSecret) < 16 {
		log.Fatal("SESSION_SECRET must be set to a random string (at least 16 characters)")
	}
	csrfKey := sha256.Sum256([]byte(sessionSecret))

	cookieStore := auth.CookieStoreFromSecret(sessionSecret)
	cookieSecure := strings.EqualFold(os.Getenv("COOKIE_SECURE"), "true") ||
		strings.HasPrefix(strings.TrimSpace(os.Getenv("PUBLIC_BASE_URL")), "https://")
	auth.SetCookieSecure(cookieStore, cookieSecure)

	ac := &auth.Config{Store: cookieStore}

	mc, err := mail.FromEnv()
	if err != nil {
		log.Fatalf("mail: %v", err)
	}
	if mc != nil && strings.TrimSpace(os.Getenv("PUBLIC_BASE_URL")) == "" {
		log.Printf("warning: MAIL_ENABLED=true but PUBLIC_BASE_URL is empty — outbound emails will omit usable links")
	}
	pubURL := strings.TrimRight(strings.TrimSpace(os.Getenv("PUBLIC_BASE_URL")), "/")

	srv := &handlers.Server{DB: db, Auth: ac, Mail: mc, PublicBaseURL: pubURL}

	mux := http.NewServeMux()

	mux.HandleFunc("GET /api/bootstrap", srv.APIBootstrap)
	mux.HandleFunc("GET /api/home/experiences", srv.APIHomeExperiences)
	mux.HandleFunc("GET /api/experiences", srv.APIExperiences)
	mux.HandleFunc("POST /api/experiences", srv.APIPostExperience)
	mux.HandleFunc("GET /api/ward-plan", srv.APIWardPlan)

	mux.HandleFunc("GET /api/admin/pending", srv.APIAdminPending)
	mux.HandleFunc("POST /api/admin/moderate", srv.APIAdminModerate)
	mux.HandleFunc("GET /api/admin/published", srv.APIAdminListPublished)
	mux.HandleFunc("POST /api/admin/published/edit", srv.APIAdminEditPublished)
	mux.HandleFunc("POST /api/admin/published/delete", srv.APIAdminDeletePublished)
	mux.HandleFunc("GET /api/admin/pending-accounts", srv.APIAdminPendingAccounts)
	mux.HandleFunc("POST /api/admin/decide-account", srv.APIAdminDecidePendingAccount)
	mux.HandleFunc("GET /api/admin/approved-leaders", srv.APIAdminApprovedLeaders)
	mux.HandleFunc("POST /api/admin/remove-approved-leader", srv.APIAdminRemoveApprovedLeader)
	mux.HandleFunc("GET /api/admin/ward-plan", srv.APIAdminWardPlan)
	mux.HandleFunc("POST /api/admin/ward-plan", srv.APIAdminWardPlanSave)

	mux.HandleFunc("POST /auth/register", srv.APIAuthRegister)
	mux.HandleFunc("POST /auth/login", srv.APIAuthLogin)
	mux.HandleFunc("POST /auth/forgot-password", srv.APIAuthForgotPassword)
	mux.HandleFunc("POST /auth/reset-password", srv.APIAuthResetPassword)
	mux.HandleFunc("POST /auth/logout", srv.Logout)
	mux.HandleFunc("GET /auth/logout", srv.Logout)

	indexHTML, err := fs.ReadFile(web.Dist, "dist/index.html")
	if err != nil {
		log.Fatal(`web UI missing: run "cd frontend && npm install && npm run build" (outputs to backend/web/dist), then retry`)
	}
	distFS, err := fs.Sub(web.Dist, "dist")
	if err != nil {
		log.Fatal(err)
	}
	mux.Handle("/", spa.FileServer(distFS, indexHTML))

	csrfFailLogger := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		reason := csrf.FailureReason(r)
		log.Printf("csrf reject %s %s: %v (origin=%q referer=%q cookie_present=%v token_header=%q)",
			r.Method, r.URL.Path, reason,
			r.Header.Get("Origin"), r.Header.Get("Referer"),
			func() bool { _, err := r.Cookie("_gorilla_csrf"); return err == nil }(),
			r.Header.Get("X-CSRF-Token"),
		)
		http.Error(w, fmt.Sprintf("Forbidden — %v", reason), http.StatusForbidden)
	})

	csrfOpts := []csrf.Option{csrf.Secure(cookieSecure), csrf.ErrorHandler(csrfFailLogger)}
	// Vite dev server runs on a different origin than the API, so by default gorilla/csrf
	// rejects its POSTs as cross-origin (returning 403 before the handler runs). Trust the
	// dev origin only when explicitly enabled with DEV_TRUST_VITE=true. Vite normally uses
	// 5173 but falls back to the next free port (5174, 5175, …) if strictPort is off, so we
	// trust a small range. When running `npm run dev:host`, the phone hits Vite at the
	// LAN IP, so we also enumerate every local IPv4 interface and trust those on the same
	// port range. For Windows-host LAN IPs that WSL2 can't enumerate, the user can list them
	// via DEV_EXTRA_TRUSTED_ORIGINS=192.168.1.100:5173,192.168.1.100:5174 etc.
	// Production builds serve the SPA from this Go server (same-origin) and need none of this.
	if strings.EqualFold(strings.TrimSpace(os.Getenv("DEV_TRUST_VITE")), "true") {
		origins := buildDevTrustedOrigins(os.Getenv("DEV_EXTRA_TRUSTED_ORIGINS"))
		csrfOpts = append(csrfOpts, csrf.TrustedOrigins(origins))
		log.Printf("DEV_TRUST_VITE=true — trusting %d Vite dev origins for CSRF (dev only)", len(origins))
		log.Printf("  trusted origins: %v", origins)
	}
	csrfProtected := csrf.Protect(csrfKey[:], csrfOpts...)(mux)

	// Endpoints that authenticate the user (login/register/logout) are intentionally exempt
	// from CSRF: they verify credentials directly, which is the CSRF-equivalent check, and
	// requiring a CSRF token here just creates brittle dev-time chicken-and-egg failures
	// (e.g. stale _gorilla_csrf cookies from prior dev sessions). Every other state-changing
	// endpoint (admin/moderate, ward plan save, post experience, decide account, etc.) still
	// requires a valid CSRF token because they act on the established session.
	csrfExempt := map[string]bool{
		"/auth/login":           true,
		"/auth/register":        true,
		"/auth/forgot-password": true,
		"/auth/reset-password":  true,
		"/auth/logout":          true,
	}
	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if csrfExempt[r.URL.Path] {
			mux.ServeHTTP(w, r)
			return
		}
		csrfProtected.ServeHTTP(w, r)
	})

	addr := ":8080"
	if p := os.Getenv("PORT"); p != "" {
		addr = ":" + strings.TrimPrefix(strings.TrimSpace(p), ":")
	}

	server := &http.Server{
		Addr:              addr,
		Handler:           logRequest(handler),
		ReadHeaderTimeout: 10 * time.Second,
	}

	fmt.Fprintf(os.Stderr, "ward mission SPA + API on http://localhost%s (PostgreSQL)\n", addr)
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

// buildDevTrustedOrigins returns the host:port strings gorilla/csrf should accept on POST.
// Includes: localhost / 127.0.0.1 on Vite's common port range, every IPv4 of every local
// non-loopback interface on the same range, plus anything the operator passed in
// DEV_EXTRA_TRUSTED_ORIGINS (comma-separated). WSL2 cannot see Windows-host LAN IPs from
// inside the VM, so the env var is the escape hatch for `npm run dev:host` setups where
// the phone reaches Vite via the Windows host's LAN IP.
func buildDevTrustedOrigins(extra string) []string {
	seen := make(map[string]struct{})
	add := func(host string) {
		host = strings.TrimSpace(host)
		if host == "" {
			return
		}
		for port := 5173; port <= 5180; port++ {
			origin := fmt.Sprintf("%s:%d", host, port)
			if _, ok := seen[origin]; !ok {
				seen[origin] = struct{}{}
			}
		}
	}

	add("localhost")
	add("127.0.0.1")

	if ifs, err := net.InterfaceAddrs(); err == nil {
		for _, a := range ifs {
			ipnet, ok := a.(*net.IPNet)
			if !ok || ipnet.IP.IsLoopback() {
				continue
			}
			if ip4 := ipnet.IP.To4(); ip4 != nil {
				add(ip4.String())
			}
		}
	}

	for _, raw := range strings.Split(extra, ",") {
		raw = strings.TrimSpace(raw)
		if raw == "" {
			continue
		}
		// Accept either bare host (we tack on the port range) or host:port (literal).
		if _, _, err := net.SplitHostPort(raw); err == nil {
			seen[raw] = struct{}{}
		} else {
			add(raw)
		}
	}

	out := make([]string, 0, len(seen))
	for o := range seen {
		out = append(out, o)
	}
	return out
}
