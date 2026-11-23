package store

import (
	"context"
	"errors"
	"os"
	"reflect"
	"testing"

	"github.com/albertson/albertsonmissionplan/internal/challenges"
)

func testMonthlyChallengesStore(t *testing.T) {
	ctx := context.Background()
	db, err := Open(ctx)
	if err != nil {
		t.Fatal(err)
	}
	defer db.Close()
	// Start from the seeded state even when reusing a Postgres test database.
	_, _ = db.ExecContext(ctx, `DELETE FROM monthly_challenges`)
	if err := seedMonthlyChallenges(ctx, db); err != nil {
		t.Fatal(err)
	}

	// Seeding: every month present, identical to the original static content and artwork.
	all, err := ListMonthlyChallenges(ctx, db, true)
	if err != nil || len(all) != 12 {
		t.Fatalf("seeded list = %d items, %v", len(all), err)
	}
	for i, c := range all {
		d := challenges.Defaults[i]
		if c.Month != i+1 || c.Heading != d.Heading || c.Goal != d.Goal || !reflect.DeepEqual(c.Bullets, d.Bullets) || c.Hidden || c.ImageCustom {
			t.Fatalf("month %d seeded as %+v, want %+v", i+1, c, d)
		}
		art, err := challenges.DefaultArt(c.Month)
		if err != nil {
			t.Fatal(err)
		}
		img, ctype, hash, _, err := MonthlyChallengeImage(ctx, db, c.Month, false)
		if err != nil || ctype != "image/webp" || string(img) != string(art.Image) || !IsDefaultArt(c.Month, hash) {
			t.Fatalf("month %d image type=%q len=%d err=%v", c.Month, ctype, len(img), err)
		}
		thumb, _, _, _, _ := MonthlyChallengeImage(ctx, db, c.Month, true)
		if string(thumb) != string(art.Thumb) {
			t.Fatalf("month %d thumb differs from the original 480px art", c.Month)
		}
	}

	// Edits survive a restart (re-running migrations + seeding never overwrites rows).
	edit := ChallengeText{Heading: "Kindness Challenge", Goal: "Be kind", Bullets: []challenges.Bullet{{Who: "Everyone", Text: "Smile **today**."}}, Hidden: true}
	if err := UpdateMonthlyChallengeText(ctx, db, 2, edit, "lead@example.org"); err != nil {
		t.Fatal(err)
	}
	if err := seedMonthlyChallenges(ctx, db); err != nil {
		t.Fatal(err)
	}
	got, err := GetMonthlyChallenge(ctx, db, 2)
	if err != nil || got.Heading != "Kindness Challenge" || !got.Hidden || got.UpdatedBy != "lead@example.org" || len(got.Bullets) != 1 {
		t.Fatalf("after reseed = %+v, %v", got, err)
	}
	pub, _ := ListMonthlyChallenges(ctx, db, false)
	if len(pub) != 11 || pub[1].Month != 3 {
		t.Fatalf("public list should skip hidden February: %d items", len(pub))
	}

	// A missing month is re-seeded on its own.
	_, _ = db.ExecContext(ctx, `DELETE FROM monthly_challenges WHERE month = 5`)
	if err := seedMonthlyChallenges(ctx, db); err != nil {
		t.Fatal(err)
	}
	if c, err := GetMonthlyChallenge(ctx, db, 5); err != nil || c.Heading != "Health & Wellness Challenge" {
		t.Fatalf("re-seeded May = %+v, %v", c, err)
	}

	// Replace a picture, swap months, then reset.
	custom := challenges.Art{ContentType: "image/jpeg", Image: []byte{0xFF, 0xD8, 1}, Thumb: []byte{0xFF, 0xD8, 2}}
	if err := SetMonthlyChallengeImage(ctx, db, 2, custom, true, "lead@example.org"); err != nil {
		t.Fatal(err)
	}
	if err := SwapMonthlyChallenges(ctx, db, 2, 7, "lead@example.org"); err != nil {
		t.Fatal(err)
	}
	feb, _ := GetMonthlyChallenge(ctx, db, 2)
	jul, _ := GetMonthlyChallenge(ctx, db, 7)
	if feb.Heading != "Gratitude Challenge" || feb.Hidden || !feb.ImageCustom || jul.Heading != "Kindness Challenge" || !jul.Hidden || jul.ImageType != "image/jpeg" {
		t.Fatalf("swap: feb=%+v jul=%+v", feb, jul)
	}
	if err := SwapMonthlyChallenges(ctx, db, 2, 7, "lead@example.org"); err != nil { // swap back
		t.Fatal(err)
	}
	if feb, _ = GetMonthlyChallenge(ctx, db, 7); feb.ImageCustom || !IsDefaultArt(7, feb.ImageHash) {
		t.Fatalf("July's own art should count as original again: %+v", feb)
	}
	if err := ResetMonthlyChallenge(ctx, db, 2, "lead@example.org"); err != nil {
		t.Fatal(err)
	}
	feb, _ = GetMonthlyChallenge(ctx, db, 2)
	d := challenges.Defaults[1]
	if feb.Heading != d.Heading || feb.Goal != d.Goal || !reflect.DeepEqual(feb.Bullets, d.Bullets) || feb.Hidden || feb.ImageCustom || feb.ImageType != "image/webp" {
		t.Fatalf("reset feb = %+v", feb)
	}
	if err := UpdateMonthlyChallengeText(ctx, db, 13, edit, ""); !errors.Is(err, ErrChallengeNotFound) {
		t.Fatalf("month 13 err = %v", err)
	}
	if _, err := GetMonthlyChallenge(ctx, db, 0); !errors.Is(err, ErrChallengeNotFound) {
		t.Fatalf("month 0 err = %v", err)
	}
}

func TestMonthlyChallengesStoreSQLite(t *testing.T) {
	t.Setenv("DATABASE_URL", "")
	t.Setenv("DATABASE_PATH", t.TempDir()+"/c.db")
	testMonthlyChallengesStore(t)
}

// Set TEST_POSTGRES_URL (a throwaway database) to run the same checks against PostgreSQL.
func TestMonthlyChallengesStorePostgres(t *testing.T) {
	u := os.Getenv("TEST_POSTGRES_URL")
	if u == "" {
		t.Skip("TEST_POSTGRES_URL not set")
	}
	t.Setenv("DATABASE_URL", u)
	testMonthlyChallengesStore(t)
}
