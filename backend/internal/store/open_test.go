package store

import (
	"strings"
	"testing"
)

func TestNormalizePostgresURL(t *testing.T) {
	in := "postgres://user:pass@host.example/db"
	got := normalizePostgresURL(in)
	if !strings.Contains(got, "sslmode=require") || !strings.Contains(got, "default_query_exec_mode=simple_protocol") {
		t.Fatalf("missing expected params: %q", got)
	}

	existing := "postgres://user:pass@host.example/db?sslmode=disable&default_query_exec_mode=cache_statement"
	if normalizePostgresURL(existing) != existing {
		t.Fatalf("should not override explicit params")
	}
}
