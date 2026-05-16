package privacyhint_test

import (
	"strings"
	"testing"

	"github.com/albertson/albertsonmissionplan/internal/privacyhint"
)

func TestSuggestions_BrotherName(t *testing.T) {
	h := privacyhint.Suggestions("Brother Anderson asked me to pray with him.")
	if len(h) == 0 || !strings.Contains(strings.ToLower(h[0]), "brother") {
		t.Fatalf("expected brother/sister hint, got %v", h)
	}
}

func TestSuggestions_AllowsCommonChurchPhrases(t *testing.T) {
	cases := []string{
		"Our Young Women class had a great lesson today.",
		"Relief Society was uplifting this Sunday.",
		"We studied the Book of Mormon as a family.",
	}
	for _, text := range cases {
		for _, s := range privacyhint.Suggestions(text) {
			if strings.Contains(s, "Two capitalized words in a row") {
				t.Fatalf("unexpected pair hint for %q: %s", text, s)
			}
		}
	}
}

func TestSuggestions_FlagsJaneDoe(t *testing.T) {
	h := privacyhint.Suggestions("Jane Doe came to church with us.")
	found := false
	for _, s := range h {
		if strings.Contains(s, "Jane Doe") {
			found = true
			break
		}
	}
	if !found {
		t.Fatalf("expected name pair hint for Jane Doe, got %v", h)
	}
}

func TestSuggestions_ShortFirstNamePair(t *testing.T) {
	h := privacyhint.Suggestions("We met Bre Jones at the chapel.")
	if len(h) == 0 || !strings.Contains(h[0], "Bre Jones") {
		t.Fatalf("expected Bre Jones pair hint, got %v", h)
	}
}

func TestSuggestions_LoneAfterAnd(t *testing.T) {
	h := privacyhint.Suggestions("Other elders and Tommy helped with the move.")
	found := false
	for _, s := range h {
		if strings.Contains(s, "Tommy") && strings.Contains(s, "and") {
			found = true
			break
		}
	}
	if !found {
		t.Fatalf("expected lone first-name hint for Tommy, got %v", h)
	}
}

func TestSuggestions_MultipleDistinctNames(t *testing.T) {
	text := "visit Bre Jones and Jacob Jones for dinner, and Tommy joined us."
	h := privacyhint.Suggestions(text)
	if len(h) < 3 {
		t.Fatalf("expected at least 3 hints for three name-like snippets, got %d: %v", len(h), h)
	}
}
