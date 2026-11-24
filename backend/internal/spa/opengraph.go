package spa

import (
	"html"
	"regexp"
	"strconv"
	"strings"
)

// OpenGraph holds the tags written into the SPA shell for a per-event share URL,
// so iMessage, Facebook, and other crawlers can show a title, description, and image.
type OpenGraph struct {
	Title       string
	Description string
	URL         string
	Image       string
	ImageAlt    string
	ImageWidth  int
	ImageHeight int
}

var (
	reTitle   = regexp.MustCompile(`(?is)<title>.*?</title>`)
	reMeta    = regexp.MustCompile(`(?is)<meta\s+[^>]*>`)
	reContent = regexp.MustCompile(`(?i)content="[^"]*"`)
)

// ApplyOpenGraph rewrites title and common Open Graph / Twitter tags in index HTML.
// Tags that aren't present are left alone; empty fields are not written.
func ApplyOpenGraph(indexHTML []byte, og OpenGraph) []byte {
	s := string(indexHTML)
	if og.Title != "" {
		s = reTitle.ReplaceAllString(s, "<title>"+html.EscapeString(og.Title)+"</title>")
		s = replaceMeta(s, "property", "og:title", og.Title)
		s = replaceMeta(s, "name", "twitter:title", og.Title)
	}
	if og.Description != "" {
		s = replaceMeta(s, "name", "description", og.Description)
		s = replaceMeta(s, "property", "og:description", og.Description)
		s = replaceMeta(s, "name", "twitter:description", og.Description)
	}
	if og.URL != "" {
		s = replaceMeta(s, "property", "og:url", og.URL)
	}
	if og.Image != "" {
		s = replaceMeta(s, "property", "og:image", og.Image)
		s = replaceMeta(s, "name", "twitter:image", og.Image)
	}
	if og.ImageAlt != "" {
		s = replaceMeta(s, "property", "og:image:alt", og.ImageAlt)
	}
	if og.ImageWidth > 0 {
		s = replaceMeta(s, "property", "og:image:width", strconv.Itoa(og.ImageWidth))
	}
	if og.ImageHeight > 0 {
		s = replaceMeta(s, "property", "og:image:height", strconv.Itoa(og.ImageHeight))
	}
	return []byte(s)
}

func replaceMeta(s, attr, key, value string) string {
	want := strings.ToLower(attr) + `="` + strings.ToLower(key) + `"`
	escaped := `content="` + html.EscapeString(value) + `"`
	return reMeta.ReplaceAllStringFunc(s, func(tag string) string {
		if !strings.Contains(strings.ToLower(tag), want) {
			return tag
		}
		if reContent.MatchString(tag) {
			return reContent.ReplaceAllString(tag, escaped)
		}
		trimmed := strings.TrimSpace(tag)
		trimmed = strings.TrimSuffix(trimmed, "/>")
		trimmed = strings.TrimSuffix(trimmed, ">")
		trimmed = strings.TrimSpace(trimmed)
		return trimmed + " " + escaped + " />"
	})
}
