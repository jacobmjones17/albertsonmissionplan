package spa

import (
	"net/http"
	"net/http/httptest"
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
