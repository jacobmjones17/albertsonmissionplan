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
		{"/admin/reset-password", 200, true},
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
