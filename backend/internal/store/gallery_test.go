package store

import (
	"context"
	"errors"
	"os"
	"testing"
)

func testGalleryStore(t *testing.T) {
	ctx := context.Background()
	db, err := Open(ctx)
	if err != nil {
		t.Fatal(err)
	}
	defer db.Close()
	_, _ = db.ExecContext(ctx, `DELETE FROM gallery_items`)

	img := GalleryImages{Width: 4, Height: 3, ThumbWidth: 2, ThumbHeight: 1, Hash: "abc123", Image: []byte{0xFF, 0xD8, 1, 2, 0}, Thumb: []byte{9, 0, 8}}
	a, err := InsertGalleryItem(ctx, db, GalleryItemInput{Title: "A", Artist: "Mia"}, img, "lead@example.org")
	if err != nil {
		t.Fatal(err)
	}
	b, err := InsertGalleryItem(ctx, db, GalleryItemInput{Title: "B", Hidden: true}, img, "lead@example.org")
	if err != nil {
		t.Fatal(err)
	}
	pub, err := ListGalleryItems(ctx, db, false)
	if err != nil || len(pub) != 1 || pub[0].ID != a {
		t.Fatalf("public list = %+v, %v", pub, err)
	}
	all, err := ListGalleryItems(ctx, db, true)
	if err != nil || len(all) != 2 || all[0].ID != b || !all[0].Hidden {
		t.Fatalf("all = %+v, %v", all, err)
	}
	data, hash, hidden, err := GalleryImage(ctx, db, a, false)
	if err != nil || string(data) != string(img.Image) || hash != "abc123" || hidden {
		t.Fatalf("image = %v %q %v %v", data, hash, hidden, err)
	}
	thumb, _, hidden, _ := GalleryImage(ctx, db, b, true)
	if string(thumb) != string(img.Thumb) || !hidden {
		t.Fatalf("thumb = %v hidden=%v", thumb, hidden)
	}
	total, nHidden, err := GalleryCounts(ctx, db)
	if err != nil || total != 2 || nHidden != 1 {
		t.Fatalf("counts %d %d %v", total, nHidden, err)
	}
	if err := UpdateGalleryItem(ctx, db, a, GalleryItemInput{Title: "A2", Artist: "Mia"}); err != nil {
		t.Fatal(err)
	}
	g, err := GetGalleryItem(ctx, db, a)
	if err != nil || g.Title != "A2" || g.Artist != "Mia" || g.Hidden || g.CreatedAt.IsZero() {
		t.Fatalf("get = %+v, %v", g, err)
	}
	if err := DeleteGalleryItem(ctx, db, a); err != nil {
		t.Fatal(err)
	}
	if err := DeleteGalleryItem(ctx, db, a); !errors.Is(err, ErrGalleryNotFound) {
		t.Fatalf("second delete err = %v", err)
	}
	if _, _, _, err := GalleryImage(ctx, db, a, false); !errors.Is(err, ErrGalleryNotFound) {
		t.Fatalf("deleted image err = %v", err)
	}
}

func TestGalleryStoreSQLite(t *testing.T) {
	t.Setenv("DATABASE_URL", "")
	t.Setenv("DATABASE_PATH", t.TempDir()+"/g.db")
	testGalleryStore(t)
}

// Set TEST_POSTGRES_URL (a throwaway database) to run the same checks against PostgreSQL.
func TestGalleryStorePostgres(t *testing.T) {
	u := os.Getenv("TEST_POSTGRES_URL")
	if u == "" {
		t.Skip("TEST_POSTGRES_URL not set")
	}
	t.Setenv("DATABASE_URL", u)
	testGalleryStore(t)
}
