package store

import (
	"context"
	"os"
	"path/filepath"
	"testing"
)

func TestMigrateOrgSectionTitles(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "test.db")
	os.Setenv("DATABASE_PATH", path)
	t.Cleanup(func() { os.Unsetenv("DATABASE_PATH") })

	ctx := context.Background()
	db, err := Open(ctx)
	if err != nil {
		t.Fatal(err)
	}
	defer db.Close()

	if _, err := sxExec(db, ctx,
		`UPDATE org_sections SET title = $1 WHERE slug = $2`,
		"Missionary coordination", "missionary-coordination",
	); err != nil {
		t.Fatal(err)
	}
	if _, err := sxExec(db, ctx,
		`UPDATE org_sections SET title = $1 WHERE slug = $2`,
		"Ward council", "ward-council",
	); err != nil {
		t.Fatal(err)
	}

	if err := migrateOrgSectionTitles(ctx, db); err != nil {
		t.Fatal(err)
	}

	orgs, err := ListOrgSections(ctx, db)
	if err != nil {
		t.Fatal(err)
	}
	want := map[string]string{
		"missionary-coordination": "Missionary Coordination",
		"ward-council":            "Young Single Adults",
	}
	for _, o := range orgs {
		if title, ok := want[o.Slug]; ok {
			if o.Title != title {
				t.Fatalf("slug %q: got title %q, want %q", o.Slug, o.Title, title)
			}
			delete(want, o.Slug)
		}
	}
	if len(want) != 0 {
		t.Fatalf("missing orgs: %v", want)
	}
}
