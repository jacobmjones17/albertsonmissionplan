package privacyhint

import (
	"regexp"
	"strings"
)

var (
	reBrotherSister = regexp.MustCompile(`(?i)\b(?:brother|sister)\s+[A-Z][a-z]{2,}\b`)
	// First word ≥3 letters (e.g. "Bre"), second ≥3 — catches short first names + surnames.
	reTwoCap = regexp.MustCompile(`\b([A-Z][a-z]{2,})\s+([A-Z][a-z]{2,})\b`)
	// Capitalized given name right after "and"/"or" (e.g. "... and Tommy").
	reAfterAndOr = regexp.MustCompile(`\b(?:and|or)\s+([A-Z][a-z]{2,})\b`)
)

// allowedPairs are two-word phrases that often appear capitalized in Latter-day Saint writing
// and are not typically private personal names.
var allowedPairs = map[string]struct{}{
	"book mormon": {}, "jesus christ": {}, "heavenly father": {}, "the lord": {},
	"holy ghost": {}, "latter day": {}, "new testament": {}, "old testament": {},
	"our young": {},
	"young women": {}, "young men": {}, "elders quorum": {}, "relief society": {},
	"sunday school": {}, "ward council": {}, "stake conference": {}, "general conference": {},
	"sacrament meeting": {}, "family history": {}, "temple recommend": {}, "temple trip": {},
	"family home": {}, "home evening": {}, "full time": {}, "part time": {},
	"ministering brothers": {}, "ministering sisters": {}, "mission leaders": {}, "ward mission": {},
	"primary program": {}, "primary children": {}, "aaronic priesthood": {}, "melchizedek priesthood": {},
	"first presidency": {}, "quorum twelve": {}, "salt lake": {}, "north america": {}, "south america": {},
	"united states": {}, "new year": {}, "new members": {}, "new convert": {},
	"great blessing": {}, "great experience": {}, "other ward": {}, "other stake": {},
	"my family": {}, "our family": {}, "our ward": {}, "our stake": {}, "our missionaries": {},
	"our neighbors": {}, "our friends": {}, "bishopric youth": {}, "stake president": {},
	"mission president": {}, "area authority": {}, "relief presidency": {}, "elders president": {},
	"elder quorum": {}, // rare typo
	"high council": {}, "patriarch blessing": {},
	"eternal families": {}, "eternal family": {},
	"president nelson": {}, "president eyring": {}, "president cook": {},
	"temple square": {},
}

// skipAfterConj — unlikely to be a person’s name right after "and"/"or".
var skipAfterConj = map[string]struct{}{
	"then": {}, "the": {}, "also": {}, "finally": {}, "now": {}, "again": {}, "yet": {}, "so": {},
	"when": {}, "why": {}, "how": {}, "every": {}, "each": {}, "all": {}, "some": {}, "most": {},
	"such": {}, "own": {}, "same": {}, "other": {}, "more": {}, "less": {},
}

const maxSuggestions = 6

// Suggestions returns non-blocking reminders for leaders reviewing a submission.
// Heuristics miss some names and flag some innocent phrases—treat as a nudge only.
func Suggestions(text string) []string {
	text = strings.TrimSpace(text)
	if text == "" {
		return []string{}
	}

	out := make([]string, 0)

	if s := reBrotherSister.FindString(text); s != "" {
		out = append(out,
			`We noticed "`+s+`" — titles like Brother or Sister followed by a capitalized word are often someone's name. Consider editing to something like "a counselor in our elders quorum" if needed.`)
	}

	if len(out) >= maxSuggestions {
		return out[:maxSuggestions]
	}

	seenPair := make(map[string]struct{})
	for _, m := range reTwoCap.FindAllStringSubmatch(text, -1) {
		if len(m) < 3 {
			continue
		}
		pair := strings.ToLower(m[1] + " " + m[2])
		if _, ok := allowedPairs[pair]; ok {
			continue
		}
		full := m[0]
		if _, ok := seenPair[full]; ok {
			continue
		}
		seenPair[full] = struct{}{}

		out = append(out,
			`Two capitalized words in a row ("`+full+`") sometimes name a specific person. Glance twice before publishing — this check is automatic and not perfect.`)

		if len(out) >= maxSuggestions {
			return out[:maxSuggestions]
		}
	}

	seenLone := make(map[string]struct{})
	for _, m := range reAfterAndOr.FindAllStringSubmatch(text, -1) {
		if len(m) < 2 {
			continue
		}
		name := m[1]
		if _, ok := skipAfterConj[strings.ToLower(name)]; ok {
			continue
		}
		key := name
		if _, ok := seenLone[key]; ok {
			continue
		}
		seenLone[key] = struct{}{}

		out = append(out,
			`Capitalized word after "and" / "or" ("`+name+`") is sometimes a first name. Glance twice — this check is automatic and not perfect.`)

		if len(out) >= maxSuggestions {
			break
		}
	}

	return out
}
