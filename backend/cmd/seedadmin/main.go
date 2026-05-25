// seedadmin creates (or resets) an approved leader account directly in the database,
// without going through the registration UI. Uses PostgreSQL when DATABASE_URL is set;
// otherwise SQLite (same defaults as the main server).
//
// Example:
//
//	go run ./cmd/seedadmin -email you@example.org -password 'changeme'
//
// Or with Postgres:
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

	ctx, cancel := context.WithTimeout(context.Background(), 20*time.Second)
	defer cancel()

	db, err := store.Open(ctx)
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
		if err := store.UpsertLeaderCredentialCLI(ctx, db, hash, fn, ln, "cli-seed", em); err != nil {
			log.Fatalf("reset existing credential: %v", err)
		}
		fmt.Printf("reset password for existing leader account: %s (already approved or now approved)\n", em)
	default:
		log.Fatalf("insert credential: %v", err)
	}
}
