// seedadmin is a tiny CLI that creates (or resets) an approved leader account
// directly in the SQLite database, without going through the registration UI.
// Useful for bootstrapping the first admin or recovering access after a wipe.
//
// Example:
//
//	go run ./cmd/seedadmin -email you@example.org -password 'changeme'
//
// If the email already has a row, its password hash is replaced and the row is
// re-marked as approved. The account is immediately usable for sign-in — there
// is no separate email allowlist to maintain.
package main

import (
	"context"
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

	dbPath := strings.TrimSpace(os.Getenv("DATABASE_PATH"))
	if dbPath == "" {
		dbPath = "wardmission.db"
	}

	ctx, cancel := context.WithTimeout(context.Background(), 20*time.Second)
	defer cancel()

	db, err := store.Open(ctx, dbPath)
	if err != nil {
		log.Fatalf("open db %s: %v", dbPath, err)
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
	case err.Error() == store.ErrLeaderAlreadyRegistered.Error():
		// Row exists. Reset password + force-approve via direct SQL since the store API
		// does not expose an "upsert" path (intentionally — UI flow never resets passwords).
		// When -first/-last are provided they replace whatever is on file; empty flags keep
		// the existing name untouched (so re-running with just -password is non-destructive).
		if _, err := db.ExecContext(ctx,
			`UPDATE leader_credentials
			   SET password_hash = ?,
			       first_name    = CASE WHEN ?='' THEN COALESCE(first_name, '') ELSE ? END,
			       last_name     = CASE WHEN ?='' THEN COALESCE(last_name,  '') ELSE ? END,
			       approved_at   = COALESCE(approved_at, datetime('now')),
			       approved_by   = COALESCE(approved_by, ?)
			 WHERE email = ?`,
			hash, fn, fn, ln, ln, "cli-seed", em,
		); err != nil {
			log.Fatalf("reset existing credential: %v", err)
		}
		fmt.Printf("reset password for existing leader account: %s (already approved or now approved)\n", em)
	default:
		log.Fatalf("insert credential: %v", err)
	}
}
