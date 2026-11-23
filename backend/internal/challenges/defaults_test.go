package challenges

import (
	"os"
	"path/filepath"
	"regexp"
	"strconv"
	"strings"
	"testing"
)

// The database seed (Defaults) and the frontend's offline fallback (monthlyData.ts) must hold
// the same original content, so the site looks identical whichever one is showing.
func TestDefaultsMatchFrontend(t *testing.T) {
	path := filepath.Join("..", "..", "..", "frontend", "src", "monthlyData.ts")
	raw, err := os.ReadFile(path)
	if err != nil {
		t.Skipf("frontend source not available: %v", err)
	}
	src := string(raw)
	if n := len(regexp.MustCompile(`id: 'g\d+'`).FindAllString(src, -1)); n != len(Defaults) {
		t.Fatalf("monthlyData.ts has %d goals, Defaults has %d", n, len(Defaults))
	}
	for _, d := range Defaults {
		goalStart := strings.Index(src, "id: 'g"+itoa(d.Month)+"'")
		if goalStart < 0 {
			t.Fatalf("month %d: goal id g%d missing from monthlyData.ts", d.Month, d.Month)
		}
		if !strings.Contains(src, "g"+itoa(d.Month)+": '"+d.Heading+"'") {
			t.Errorf("month %d heading %q not in CHALLENGE_CARD_HEADING", d.Month, d.Heading)
		}
		goal := src[goalStart:]
		if end := strings.Index(goal[1:], "id: 'g"); end >= 0 {
			goal = goal[:end+1]
		}
		if !strings.Contains(goal, "title: '"+d.Goal+"'") {
			t.Errorf("month %d goal %q differs from monthlyData.ts", d.Month, d.Goal)
		}
		for _, b := range d.Bullets {
			if !strings.Contains(goal, "who: '"+b.Who+"'") || !strings.Contains(goal, "text: '"+b.Text+"'") {
				t.Errorf("month %d idea %q: %q differs from monthlyData.ts", d.Month, b.Who, b.Text)
			}
		}
	}
}

func itoa(n int) string { return strconv.Itoa(n) }

func TestDefaultArtPresent(t *testing.T) {
	for _, d := range Defaults {
		art, err := DefaultArt(d.Month)
		if err != nil {
			t.Fatal(err)
		}
		if art.ContentType != "image/webp" || len(art.Image) < 1000 || len(art.Thumb) < 1000 {
			t.Fatalf("month %d art looks wrong: %s %d/%d bytes", d.Month, art.ContentType, len(art.Image), len(art.Thumb))
		}
	}
	if _, err := DefaultArt(13); err == nil {
		t.Fatal("month 13 should fail")
	}
}
