package store

import (
	"context"
	"database/sql"
	"errors"
	"strings"
	"time"
)

// ErrLeaderAlreadyRegistered indicates there is already a password for this leader email.
var ErrLeaderAlreadyRegistered = errors.New("leader email already registered")

// ErrLeaderAccountPending indicates the account exists but has not been approved yet.
var ErrLeaderAccountPending = errors.New("leader account awaiting admin approval")

// ErrCannotRemoveLastLeader is returned when deleting would leave zero approved leaders.
var ErrCannotRemoveLastLeader = errors.New("cannot remove the last approved leader account")

func normLeaderEmail(email string) string {
	return strings.ToLower(strings.TrimSpace(email))
}

func isSQLiteUniqueViolation(err error) bool {
	if err == nil {
		return false
	}
	return strings.Contains(strings.ToLower(err.Error()), "unique constraint failed")
}

// HasApprovedLeaderAccount reports whether any leader_credentials row has been approved.
// Used to auto-approve the very first registration so the system can bootstrap without
// an existing admin.
func HasApprovedLeaderAccount(ctx context.Context, db *sql.DB) (bool, error) {
	var n int
	if err := db.QueryRowContext(ctx, `SELECT COUNT(*) FROM leader_credentials WHERE approved_at IS NOT NULL`).Scan(&n); err != nil {
		return false, err
	}
	return n > 0, nil
}

// IsApprovedLeaderAccount reports whether this specific email has an approved row. Used to
// validate that an active session still maps to a real, approved account on every privileged
// request (so deleting/denying an account immediately revokes access on next request).
func IsApprovedLeaderAccount(ctx context.Context, db *sql.DB, email string) (bool, error) {
	em := normLeaderEmail(email)
	if em == "" {
		return false, nil
	}
	var n int
	if err := db.QueryRowContext(ctx,
		`SELECT COUNT(*) FROM leader_credentials WHERE email = ? AND approved_at IS NOT NULL`, em,
	).Scan(&n); err != nil {
		return false, err
	}
	return n > 0, nil
}

// InsertLeaderCredential stores a bcrypt hash for password sign-in.
// When autoApprovedBy is non-empty the row is created already approved (used to bootstrap the
// first admin); otherwise the account is pending until an existing leader approves it.
//
// firstName/lastName are stored verbatim (after trimming) so the admin reviewing a pending
// request can see who is asking for access. They are required at the API layer but tolerated
// as empty here for legacy callers / migrations.
func InsertLeaderCredential(ctx context.Context, db *sql.DB, email, firstName, lastName, passwordHash, autoApprovedBy string) error {
	em := normLeaderEmail(email)
	if em == "" {
		return errors.New("empty email")
	}
	fn := strings.TrimSpace(firstName)
	ln := strings.TrimSpace(lastName)
	var err error
	if strings.TrimSpace(autoApprovedBy) != "" {
		_, err = db.ExecContext(ctx,
			`INSERT INTO leader_credentials (email, password_hash, first_name, last_name, created_at, approved_at, approved_by)
			 VALUES (?, ?, ?, ?, datetime('now'), datetime('now'), ?)`,
			em, passwordHash, fn, ln, strings.TrimSpace(autoApprovedBy),
		)
	} else {
		_, err = db.ExecContext(ctx,
			`INSERT INTO leader_credentials (email, password_hash, first_name, last_name, created_at)
			 VALUES (?, ?, ?, ?, datetime('now'))`,
			em, passwordHash, fn, ln,
		)
	}
	if err != nil {
		if isSQLiteUniqueViolation(err) {
			return ErrLeaderAlreadyRegistered
		}
		return err
	}
	return nil
}

// LeaderName returns the stored first + last name for an account (either pending or approved).
// Empty strings are valid and mean "not provided" (e.g. accounts seeded via the CLI before names
// were collected). Returns sql.ErrNoRows if the account does not exist.
func LeaderName(ctx context.Context, db *sql.DB, email string) (firstName, lastName string, err error) {
	em := normLeaderEmail(email)
	if em == "" {
		return "", "", sql.ErrNoRows
	}
	var fn, ln sql.NullString
	err = db.QueryRowContext(ctx,
		`SELECT first_name, last_name FROM leader_credentials WHERE email = ?`, em,
	).Scan(&fn, &ln)
	if err != nil {
		return "", "", err
	}
	return strings.TrimSpace(fn.String), strings.TrimSpace(ln.String), nil
}

// LeaderCredentialHash returns the stored bcrypt hash if the account exists AND is approved.
// Returns ErrLeaderAccountPending when the row exists but has not been approved yet.
// Returns sql.ErrNoRows when no account exists.
func LeaderCredentialHash(ctx context.Context, db *sql.DB, email string) (string, error) {
	em := normLeaderEmail(email)
	var h string
	var approvedAt sql.NullString
	err := db.QueryRowContext(ctx,
		`SELECT password_hash, approved_at FROM leader_credentials WHERE email = ?`, em,
	).Scan(&h, &approvedAt)
	if err != nil {
		return "", err
	}
	if !approvedAt.Valid || strings.TrimSpace(approvedAt.String) == "" {
		return "", ErrLeaderAccountPending
	}
	return h, nil
}

// PendingLeaderAccount is one row awaiting admin approval.
type PendingLeaderAccount struct {
	Email     string
	FirstName string
	LastName  string
	CreatedAt time.Time
}

// ListPendingLeaderAccounts returns accounts created via /auth/register that are still pending.
func ListPendingLeaderAccounts(ctx context.Context, db *sql.DB) ([]PendingLeaderAccount, error) {
	rows, err := db.QueryContext(ctx,
		`SELECT email, COALESCE(first_name, ''), COALESCE(last_name, ''), created_at
		   FROM leader_credentials WHERE approved_at IS NULL ORDER BY datetime(created_at) ASC`,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var out []PendingLeaderAccount
	for rows.Next() {
		var p PendingLeaderAccount
		var created string
		if err := rows.Scan(&p.Email, &p.FirstName, &p.LastName, &created); err != nil {
			return nil, err
		}
		p.FirstName = strings.TrimSpace(p.FirstName)
		p.LastName = strings.TrimSpace(p.LastName)
		p.CreatedAt = parseSQLiteTime(created)
		out = append(out, p)
	}
	return out, rows.Err()
}

// ListApprovedLeaderEmails returns emails for approved leaders (notify targets for pending signup alerts).
func ListApprovedLeaderEmails(ctx context.Context, db *sql.DB) ([]string, error) {
	rows, err := db.QueryContext(ctx,
		`SELECT email FROM leader_credentials WHERE approved_at IS NOT NULL ORDER BY email ASC`,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var out []string
	for rows.Next() {
		var em string
		if err := rows.Scan(&em); err != nil {
			return nil, err
		}
		em = strings.TrimSpace(em)
		if em != "" {
			out = append(out, em)
		}
	}
	return out, rows.Err()
}

// ApproveLeaderAccount sets approved_at + approved_by for a pending account.
func ApproveLeaderAccount(ctx context.Context, db *sql.DB, email, approvedBy string) error {
	em := normLeaderEmail(email)
	if em == "" {
		return errors.New("empty email")
	}
	res, err := db.ExecContext(ctx,
		`UPDATE leader_credentials SET approved_at = datetime('now'), approved_by = ? WHERE email = ? AND approved_at IS NULL`,
		strings.TrimSpace(approvedBy), em,
	)
	if err != nil {
		return err
	}
	n, err := res.RowsAffected()
	if err != nil {
		return err
	}
	if n == 0 {
		return errors.New("no pending account for that email")
	}
	return nil
}

// DeletePendingLeaderAccount removes a pending account (deny request). Refuses to touch an
// already-approved row so an admin cannot accidentally delete an active account via this path.
func DeletePendingLeaderAccount(ctx context.Context, db *sql.DB, email string) error {
	em := normLeaderEmail(email)
	if em == "" {
		return errors.New("empty email")
	}
	res, err := db.ExecContext(ctx,
		`DELETE FROM leader_credentials WHERE email = ? AND approved_at IS NULL`,
		em,
	)
	if err != nil {
		return err
	}
	n, err := res.RowsAffected()
	if err != nil {
		return err
	}
	if n == 0 {
		return errors.New("no pending account for that email")
	}
	return nil
}

// ApprovedLeaderAccount is one approved leader row (for admin listing).
type ApprovedLeaderAccount struct {
	Email      string
	FirstName  string
	LastName   string
	ApprovedAt time.Time
}

// ListApprovedLeaderAccounts returns approved accounts ordered by approval time (oldest first).
func ListApprovedLeaderAccounts(ctx context.Context, db *sql.DB) ([]ApprovedLeaderAccount, error) {
	rows, err := db.QueryContext(ctx,
		`SELECT email, COALESCE(first_name, ''), COALESCE(last_name, ''), approved_at
		   FROM leader_credentials WHERE approved_at IS NOT NULL ORDER BY datetime(approved_at) ASC`,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var out []ApprovedLeaderAccount
	for rows.Next() {
		var a ApprovedLeaderAccount
		var approved string
		if err := rows.Scan(&a.Email, &a.FirstName, &a.LastName, &approved); err != nil {
			return nil, err
		}
		a.FirstName = strings.TrimSpace(a.FirstName)
		a.LastName = strings.TrimSpace(a.LastName)
		a.ApprovedAt = parseSQLiteTime(approved)
		out = append(out, a)
	}
	return out, rows.Err()
}

// DeleteApprovedLeaderAccount removes an approved leader credential row. Refuses when it would
// leave zero approved leaders. Clears outstanding password-reset tokens for that email.
func DeleteApprovedLeaderAccount(ctx context.Context, db *sql.DB, email string) error {
	em := normLeaderEmail(email)
	if em == "" {
		return errors.New("empty email")
	}
	tx, err := db.BeginTx(ctx, nil)
	if err != nil {
		return err
	}
	defer func() { _ = tx.Rollback() }()

	var total int
	if err := tx.QueryRowContext(ctx,
		`SELECT COUNT(*) FROM leader_credentials WHERE approved_at IS NOT NULL`,
	).Scan(&total); err != nil {
		return err
	}
	if total <= 1 {
		return ErrCannotRemoveLastLeader
	}

	res, err := tx.ExecContext(ctx,
		`DELETE FROM leader_credentials WHERE email = ? AND approved_at IS NOT NULL`,
		em,
	)
	if err != nil {
		return err
	}
	n, err := res.RowsAffected()
	if err != nil {
		return err
	}
	if n == 0 {
		return errors.New("no approved leader account for that email")
	}

	if _, err := tx.ExecContext(ctx, `DELETE FROM password_reset_tokens WHERE email = ?`, em); err != nil {
		return err
	}

	return tx.Commit()
}
