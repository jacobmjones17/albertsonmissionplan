package sanitizehtml

import (
	"strings"

	"github.com/microcosm-cc/bluemonday"
)

var wardLinePolicy *bluemonday.Policy

func init() {
	wardLinePolicy = bluemonday.UGCPolicy()
}

// WardBullets splits on newlines, trims each line, HTML-sanitizes non-empty lines,
// and rejoins with newlines for storage.
func WardBullets(raw string) string {
	raw = strings.ReplaceAll(raw, "\r\n", "\n")
	lines := strings.Split(raw, "\n")
	var out []string
	for _, ln := range lines {
		ln = strings.TrimSpace(ln)
		if ln == "" {
			continue
		}
		out = append(out, wardLinePolicy.Sanitize(ln))
	}
	return strings.Join(out, "\n")
}
