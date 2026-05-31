package auth

import (
	"os"
	"strconv"
	"strings"
	"time"
)

// SessionPolicy controls leader session lifetime (signed cookie, not OAuth).
type SessionPolicy struct {
	// IdleTimeout: sign out after this long without a validated/touched request.
	IdleTimeout time.Duration
	// MaxLifetime: absolute cap from initial sign-in, even if still active.
	MaxLifetime time.Duration
}

// DefaultSessionPolicy is used when env vars are unset or invalid.
func DefaultSessionPolicy() SessionPolicy {
	return SessionPolicy{
		IdleTimeout: 2 * time.Hour,
		MaxLifetime: 24 * time.Hour,
	}
}

// SessionPolicyFromEnv reads SESSION_IDLE_MINUTES and SESSION_MAX_HOURS.
func SessionPolicyFromEnv() SessionPolicy {
	p := DefaultSessionPolicy()
	if v := strings.TrimSpace(os.Getenv("SESSION_IDLE_MINUTES")); v != "" {
		if n, err := strconv.Atoi(v); err == nil && n > 0 {
			p.IdleTimeout = time.Duration(n) * time.Minute
		}
	}
	if v := strings.TrimSpace(os.Getenv("SESSION_MAX_HOURS")); v != "" {
		if n, err := strconv.Atoi(v); err == nil && n > 0 {
			p.MaxLifetime = time.Duration(n) * time.Hour
		}
	}
	if p.MaxLifetime < p.IdleTimeout {
		p.MaxLifetime = p.IdleTimeout
	}
	return p
}
