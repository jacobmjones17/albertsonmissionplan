// seedadmin creates (or resets) an approved leader account directly in PostgreSQL,
// without going through the registration UI. Useful for bootstrapping or recovering access.
//
// Example:
//
//	DATABASE_URL=postgres://... go run ./cmd/seedadmin -email you@example.org -password 'changeme'
//
// If the email already has a row, its password hash is replaced and the row is
// re-marked as approved. There is no separate email allowlist to maintain.
package main

import (
	"context"
	"errors"
	"flag"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/joho/godotenv"

	"github.com/albertson/albertsonmissionplan/internal/auth"
	"github.com/albertson/albertsonmissionplan/internal/store"
)

func loadDotenv() {
	wd, err := os.Getwd()
	if err != nil {
		return
	}
	for d := wd; d != filepath.Dir(d); d = filepath.Dir(d) {
		p := filepath.Join(d, ".env")
		if st, err := os.Stat(p); err == nil && !st.IsDir() {
			_ = godotenv.Overload(p)
		}
	}
}

func main() {
	loadDotenv()

	email := flag.String("email", "", "leader email to seed (required)")
	password := flag.String("password", "", "plaintext password to set (required; will be bcrypt-hashed)")
	first := flag.String("first", "", "first name to display in admin UI (optional)")
	last := flag.String("last", "", "last name to display in admin UI (optional)")
	flag.Parse()

	if strings.TrimSpace(*email) == "" || strings.TrimSpace(*password) == "" {
		fmt.Fprintln(os.Stderr, "usage: seedadmin -email <email> -password <password> [-first <name>] [-last <name>]")
		os.Exit(2)
	}

	dbURL := strings.TrimSpace(os.Getenv("DATABASE_URL"))
	if dbURL == "" {
		log.Fatal("DATABASE_URL must be set (PostgreSQL connection string)")
	}

	ctx, cancel := context.WithTimeout(context.Background(), 20*time.Second)
	defer cancel()

	db, err := store.Open(ctx, dbURL)
	if err != nil {
		log.Fatalf("open db: %v", err)
	}
	defer db.Close()

	hash, err := auth.HashPassword(*password)
	if err != nil {
		log.Fatalf("hash password: %v", err)
	}

	em := strings.ToLower(strings.TrimSpace(*email))
	fn := strings.TrimSpace(*first)
	ln := strings.TrimSpace(*last)
	err = store.InsertLeaderCredential(ctx, db, em, fn, ln, hash, "cli-seed")
	switch {
	case err == nil:
		fmt.Printf("seeded new approved leader account: %s\n", em)
	case errors.Is(err, store.ErrLeaderAlreadyRegistered):
		if _, err := db.ExecContext(ctx,
			`UPDATE leader_credentials
			   SET password_hash = $1,
			       first_name    = CASE WHEN $2='' THEN COALESCE(first_name, '') ELSE $3 END,
			       last_name     = CASE WHEN $4='' THEN COALESCE(last_name, '') ELSE $5 END,
			       approved_at   = COALESCE(approved_at, NOW()),
			       approved_by   = COALESCE(approved_by, $6)
			 WHERE email = $7`,
			hash, fn, fn, ln, ln, "cli-seed", em,
		); err != nil {
			log.Fatalf("reset existing credential: %v", err)
		}
		fmt.Printf("reset password for existing leader account: %s (already approved or now approved)\n", em)
	default:
		log.Fatalf("insert credential: %v", err)
	}
}
