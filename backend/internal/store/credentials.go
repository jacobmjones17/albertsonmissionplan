package store

import (
	"context"
	"database/sql"
	"errors"
	"strings"
	"time"
)

var ErrLeaderAlreadyRegistered = errors.New("leader email already registered")
var ErrLeaderAccountPending = errors.New("leader account awaiting admin approval")
var ErrCannotRemoveLastLeader = errors.New("cannot remove the last approved leader account")

func normLeaderEmail(email string) string {
	return strings.ToLower(strings.TrimSpace(email))
}

func isUniqueViolation(err error) bool {
	if err == nil {
		return false
	}
	msg := strings.ToLower(err.Error())
	return strings.Contains(msg, "duplicate key") ||
		strings.Contains(msg, "unique constraint failed") ||
		strings.Contains(msg, "unique constraint")
}

func HasApprovedLeaderAccount(ctx context.Context, db *sql.DB) (bool, error) {
	var n int
	if err := sxQueryRow(db, ctx, `SELECT COUNT(*) FROM leader_credentials WHERE approved_at IS NOT NULL`).Scan(&n); err != nil {
		return false, err
	}
	return n > 0, nil
}

func IsApprovedLeaderAccount(ctx context.Context, db *sql.DB, email string) (bool, error) {
	em := normLeaderEmail(email)
	if em == "" {
		return false, nil
	}
	var n int
	if err := sxQueryRow(db, ctx,
		`SELECT COUNT(*) FROM leader_credentials WHERE email = $1 AND approved_at IS NOT NULL`, em,
	).Scan(&n); err != nil {
		return false, err
	}
	return n > 0, nil
}

func InsertLeaderCredential(ctx context.Context, db *sql.DB, email, firstName, lastName, passwordHash, autoApprovedBy string) error {
	em := normLeaderEmail(email)
	if em == "" {
		return errors.New("empty email")
	}
	fn := strings.TrimSpace(firstName)
	ln := strings.TrimSpace(lastName)
	var err error
	if strings.TrimSpace(autoApprovedBy) != "" {
		_, err = sxExec(db, ctx,
			`INSERT INTO leader_credentials (email, password_hash, first_name, last_name, created_at, approved_at, approved_by)
			 VALUES ($1, $2, $3, $4, CURRENT_TIMESTAMP, CURRENT_TIMESTAMP, $5)`,
			em, passwordHash, fn, ln, strings.TrimSpace(autoApprovedBy),
		)
	} else {
		_, err = sxExec(db, ctx,
			`INSERT INTO leader_credentials (email, password_hash, first_name, last_name, created_at)
			 VALUES ($1, $2, $3, $4, CURRENT_TIMESTAMP)`,
			em, passwordHash, fn, ln,
		)
	}
	if err != nil {
		if isUniqueViolation(err) {
			return ErrLeaderAlreadyRegistered
		}
		return err
	}
	return nil
}

func LeaderName(ctx context.Context, db *sql.DB, email string) (firstName, lastName string, err error) {
	em := normLeaderEmail(email)
	if em == "" {
		return "", "", sql.ErrNoRows
	}
	var fn, ln sql.NullString
	err = sxQueryRow(db, ctx,
		`SELECT first_name, last_name FROM leader_credentials WHERE email = $1`, em,
	).Scan(&fn, &ln)
	if err != nil {
		return "", "", err
	}
	return strings.TrimSpace(fn.String), strings.TrimSpace(ln.String), nil
}

func LeaderCredentialHash(ctx context.Context, db *sql.DB, email string) (string, error) {
	em := normLeaderEmail(email)
	var h string
	var approvedAt nullTime
	err := sxQueryRow(db, ctx,
		`SELECT password_hash, approved_at FROM leader_credentials WHERE email = $1`, em,
	).Scan(&h, &approvedAt)
	if err != nil {
		return "", err
	}
	if !approvedAt.Valid {
		return "", ErrLeaderAccountPending
	}
	return h, nil
}

type PendingLeaderAccount struct {
	Email     string
	FirstName string
	LastName  string
	CreatedAt time.Time
}

func ListPendingLeaderAccounts(ctx context.Context, db *sql.DB) ([]PendingLeaderAccount, error) {
	rows, err := sxQuery(db, ctx,
		`SELECT email, COALESCE(first_name, ''), COALESCE(last_name, ''), created_at
		   FROM leader_credentials WHERE approved_at IS NULL ORDER BY created_at ASC`,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var out []PendingLeaderAccount
	for rows.Next() {
		var p PendingLeaderAccount
		var created wallTime
		if err := rows.Scan(&p.Email, &p.FirstName, &p.LastName, &created); err != nil {
			return nil, err
		}
		p.FirstName = strings.TrimSpace(p.FirstName)
		p.LastName = strings.TrimSpace(p.LastName)
		p.CreatedAt = created.Time
		out = append(out, p)
	}
	return out, rows.Err()
}

func ListApprovedLeaderEmails(ctx context.Context, db *sql.DB) ([]string, error) {
	rows, err := sxQuery(db, ctx,
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

func ApproveLeaderAccount(ctx context.Context, db *sql.DB, email, approvedBy string) error {
	em := normLeaderEmail(email)
	if em == "" {
		return errors.New("empty email")
	}
	res, err := sxExec(db, ctx,
		`UPDATE leader_credentials SET approved_at = CURRENT_TIMESTAMP, approved_by = $1 WHERE email = $2 AND approved_at IS NULL`,
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

func DeletePendingLeaderAccount(ctx context.Context, db *sql.DB, email string) error {
	em := normLeaderEmail(email)
	if em == "" {
		return errors.New("empty email")
	}
	res, err := sxExec(db, ctx,
		`DELETE FROM leader_credentials WHERE email = $1 AND approved_at IS NULL`,
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

type ApprovedLeaderAccount struct {
	Email      string
	FirstName  string
	LastName   string
	ApprovedAt time.Time
}

func ListApprovedLeaderAccounts(ctx context.Context, db *sql.DB) ([]ApprovedLeaderAccount, error) {
	rows, err := sxQuery(db, ctx,
		`SELECT email, COALESCE(first_name, ''), COALESCE(last_name, ''), approved_at
		   FROM leader_credentials WHERE approved_at IS NOT NULL ORDER BY approved_at ASC`,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var out []ApprovedLeaderAccount
	for rows.Next() {
		var a ApprovedLeaderAccount
		var approved wallTime
		if err := rows.Scan(&a.Email, &a.FirstName, &a.LastName, &approved); err != nil {
			return nil, err
		}
		a.FirstName = strings.TrimSpace(a.FirstName)
		a.LastName = strings.TrimSpace(a.LastName)
		a.ApprovedAt = approved.Time
		out = append(out, a)
	}
	return out, rows.Err()
}

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
	if err := sxTxQueryRow(tx, ctx,
		`SELECT COUNT(*) FROM leader_credentials WHERE approved_at IS NOT NULL`,
	).Scan(&total); err != nil {
		return err
	}
	if total <= 1 {
		return ErrCannotRemoveLastLeader
	}

	res, err := sxTxExec(tx, ctx,
		`DELETE FROM leader_credentials WHERE email = $1 AND approved_at IS NOT NULL`,
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

	if _, err := sxTxExec(tx, ctx, `DELETE FROM password_reset_tokens WHERE email = $1`, em); err != nil {
		return err
	}

	return tx.Commit()
}
