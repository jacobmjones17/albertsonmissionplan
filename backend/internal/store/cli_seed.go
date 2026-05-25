package store

import (
	"context"
	"database/sql"
	"strings"
)

// UpsertLeaderCredentialCLI is used by cmd/seedadmin when InsertLeaderCredential returns
// ErrLeaderAlreadyRegistered: reset bcrypt hash and ensure the row stays approved.
func UpsertLeaderCredentialCLI(ctx context.Context, db *sql.DB, hash, fn, ln, approvedByLabel, email string) error {
	_, err := sxExec(db, ctx,
		`UPDATE leader_credentials
			   SET password_hash = $1,
			       first_name    = CASE WHEN $2='' THEN COALESCE(first_name, '') ELSE $3 END,
			       last_name     = CASE WHEN $4='' THEN COALESCE(last_name, '') ELSE $5 END,
			       approved_at   = COALESCE(approved_at, CURRENT_TIMESTAMP),
			       approved_by   = COALESCE(approved_by, $6)
			 WHERE email = $7`,
		hash, fn, fn, ln, ln, strings.TrimSpace(approvedByLabel), normLeaderEmail(email),
	)
	return err
}
