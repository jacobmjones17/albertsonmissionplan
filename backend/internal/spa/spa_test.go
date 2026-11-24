package spa

import (
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"regexp"
	"sort"
	"strings"
	"testing"
	"testing/fstest"
)

func TestFileServerStatuses(t *testing.T) {
	dist := fstest.MapFS{
		"index.html":      {Data: []byte("<html>app</html>")},
		"assets/app.js":   {Data: []byte("console.log(1)")},
		"site/header.jpg": {Data: []byte("jpg")},
	}
	h := FileServer(dist, []byte("<html>app</html>"))

	cases := []struct {
		path     string
		status   int
		wantHTML bool
	}{
		{"/", 200, true},
		{"/ward-plan", 200, true},
		{"/ward-plan/", 200, true},
		{"/Monthly-Challenges", 200, true},
		{"/monthly-challenges/print", 200, true},
		{"/ward-plan/print", 200, true},
		{"/admin/reset-password", 200, true},
		{"/admin/challenges", 200, true},
		{"/assets/app.js", 200, false},
		{"/site/header.jpg", 200, false},
		{"/does-not-exist", 404, true},
		{"/ward-plann", 404, true},
		{"/admin/nope", 404, true},
		{"/site/missing.jpg", 404, false},
		{"/api/nope", 404, false},
		{"/auth/nope", 404, false},
		// Directories are never listed; they get the not-found page with a 404 status.
		{"/site", 404, true},
		{"/site/", 404, true},
		{"/assets", 404, true},
		{"/SITE", 404, true},
		// Per-event share URLs are client routes with a dynamic id.
		{"/events/12", 200, true},
		{"/events/12/", 200, true},
		{"/Events/12", 200, true},
		{"/events/abc", 404, true},
		{"/events/0", 404, true},
		{"/events/12/extra", 404, true},
	}
	for _, c := range cases {
		rec := httptest.NewRecorder()
		h.ServeHTTP(rec, httptest.NewRequest(http.MethodGet, c.path, nil))
		if rec.Code != c.status {
			t.Errorf("%s: status %d, want %d", c.path, rec.Code, c.status)
		}
		isHTML := strings.Contains(rec.Body.String(), "<html>app</html>")
		if isHTML != c.wantHTML {
			t.Errorf("%s: served app shell = %v, want %v", c.path, isHTML, c.wantHTML)
		}
	}
}

// TestClientRoutesMatchApp guards against the Go route list drifting from the React router:
// every key of PAGE_TITLES in frontend/src/App.tsx must be a client route here and vice versa.
func TestClientRoutesMatchApp(t *testing.T) {
	appPath := filepath.Join("..", "..", "..", "frontend", "src", "App.tsx")
	src, err := os.ReadFile(appPath)
	if err != nil {
		t.Skipf("frontend source not available (%v)", err)
	}
	block := regexp.MustCompile(`(?s)const PAGE_TITLES[^{]*\{(.*?)\n\}`).FindSubmatch(src)
	if block == nil {
		t.Fatalf("could not find PAGE_TITLES in %s", appPath)
	}
	var fromApp []string
	for _, m := range regexp.MustCompile(`'(/[^']*)'\s*:`).FindAllSubmatch(block[1], -1) {
		fromApp = append(fromApp, string(m[1]))
	}
	var fromGo []string
	for r := range clientRoutes {
		fromGo = append(fromGo, r)
	}
	sort.Strings(fromApp)
	sort.Strings(fromGo)
	if strings.Join(fromApp, ",") != strings.Join(fromGo, ",") {
		t.Fatalf("route lists differ\nApp.tsx PAGE_TITLES: %v\nspa.clientRoutes:     %v", fromApp, fromGo)
	}
}

func TestEventSharePath(t *testing.T) {
	if id, ok := EventSharePath("/events/42"); !ok || id != 42 {
		t.Fatalf("got %d %v", id, ok)
	}
	if id, ok := EventSharePath("/Events/42/"); !ok || id != 42 {
		t.Fatalf("slash/case: %d %v", id, ok)
	}
	for _, p := range []string{"/events", "/events/", "/events/0", "/events/-3", "/events/abc", "/events/12/extra", "/ward-plan/12"} {
		if id, ok := EventSharePath(p); ok {
			t.Errorf("%s: unexpected id %d", p, id)
		}
	}
	if !IsClientRoute("/events/42") || IsClientRoute("/events/abc") {
		t.Fatal("IsClientRoute share path")
	}
}

func TestApplyOpenGraph(t *testing.T) {
	in := []byte(`<!doctype html><html><head>
<title>Albertson Ward Plan</title>
<meta name="description" content="Albertson Ward mission plan and resources." />
<meta property="og:title" content="Albertson Ward" />
<meta property="og:description" content="Albertson Ward mission plan and resources." />
<meta property="og:url" content="https://albertsonwardplan.com/" />
<meta property="og:image" content="https://albertsonwardplan.com/site/og-share.jpg" />
<meta property="og:image:width" content="1200" />
<meta property="og:image:height" content="1600" />
<meta property="og:image:alt" content="Albertson Ward mission plan" />
<meta name="twitter:title" content="Albertson Ward" />
<meta name="twitter:description" content="Albertson Ward mission plan and resources." />
<meta name="twitter:image" content="https://albertsonwardplan.com/site/og-share.jpg" />
</head></html>`)
	out := string(ApplyOpenGraph(in, OpenGraph{
		Title:       `Ward Christmas Party | Albertson Ward`,
		Description: `Saturday, December 12, 2026 · 6:00 – 8:30 PM · Cultural Hall`,
		URL:         `https://albertsonwardplan.com/events/9`,
		Image:       `https://albertsonwardplan.com/api/events/9/flyer/image?v=abc`,
		ImageAlt:    `Flyer for Ward Christmas Party`,
		ImageWidth:  800,
		ImageHeight: 1000,
	}))
	wants := []string{
		`<title>Ward Christmas Party | Albertson Ward</title>`,
		`content="Ward Christmas Party | Albertson Ward"`,
		`content="Saturday, December 12, 2026 · 6:00 – 8:30 PM · Cultural Hall"`,
		`content="https://albertsonwardplan.com/events/9"`,
		`content="https://albertsonwardplan.com/api/events/9/flyer/image?v=abc"`,
		`content="Flyer for Ward Christmas Party"`,
		`content="800"`,
		`content="1000"`,
	}
	for _, w := range wants {
		if !strings.Contains(out, w) {
			t.Errorf("missing %q in:\n%s", w, out)
		}
	}
	if strings.Contains(out, "Albertson Ward mission plan and resources.") || strings.Contains(out, "og-share.jpg") {
		t.Errorf("stale default tags remain:\n%s", out)
	}
}
