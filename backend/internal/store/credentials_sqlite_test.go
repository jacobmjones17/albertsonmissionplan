package store

import (
	"context"
	"os"
	"testing"

	"github.com/albertson/albertsonmissionplan/internal/auth"
)

func TestLeaderCredentialHashSQLiteApprovedAt(t *testing.T) {
	path := t.TempDir() + "/test.db"
	os.Setenv("DATABASE_PATH", path)
	os.Unsetenv("DATABASE_URL")
	t.Cleanup(func() {
		os.Unsetenv("DATABASE_PATH")
	})

	ctx := context.Background()
	db, err := Open(ctx)
	if err != nil {
		t.Fatal(err)
	}
	defer db.Close()

	hash, err := auth.HashPassword("admin")
	if err != nil {
		t.Fatal(err)
	}
	if err := InsertLeaderCredential(ctx, db, "leader@example.com", "Ada", "Lovelace", hash, "test"); err != nil {
		t.Fatal(err)
	}

	got, err := LeaderCredentialHash(ctx, db, "leader@example.com")
	if err != nil {
		t.Fatal(err)
	}
	if !auth.CheckPassword("admin", got) {
		t.Fatal("password mismatch")
	}
}
