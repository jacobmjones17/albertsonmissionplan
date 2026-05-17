package store

import (
	"context"
	"crypto/sha256"
	"database/sql"
	"encoding/hex"
	"errors"
	"fmt"
	"strings"
)

// ErrPasswordResetInvalidToken means the token is unknown or expired.
var ErrPasswordResetInvalidToken = errors.New("invalid or expired reset token")

// PasswordResetTokenHashHex is the hex-encoded SHA-256 of the raw secret bytes emailed to the user.
func PasswordResetTokenHashHex(rawSecret []byte) string {
	sum := sha256.Sum256(rawSecret)
	return hex.EncodeToString(sum[:])
}

// DeletePasswordResetTokensForEmail removes outstanding reset links for one account.
func DeletePasswordResetTokensForEmail(ctx context.Context, db *sql.DB, email string) error {
	em := normLeaderEmail(email)
	if em == "" {
		return errors.New("empty email")
	}
	_, err := db.ExecContext(ctx, `DELETE FROM password_reset_tokens WHERE email = ?`, em)
	return err
}

// InsertPasswordResetToken stores token_hash (SHA-256 hex of secret). expiresUnix is Unix seconds.
func InsertPasswordResetToken(ctx context.Context, db *sql.DB, email, tokenHashHex string, expiresUnix int64) error {
	em := normLeaderEmail(email)
	if em == "" || strings.TrimSpace(tokenHashHex) == "" || expiresUnix <= 0 {
		return errors.New("invalid password reset token insert")
	}
	_, err := db.ExecContext(ctx,
		`INSERT INTO password_reset_tokens (token_hash, email, expires_at, created_at)
		 VALUES (?, ?, ?, datetime('now'))`,
		strings.TrimSpace(tokenHashHex), em, fmt.Sprintf("%d", expiresUnix),
	)
	return err
}

// ResetLeaderPasswordWithToken validates token, deletes the row, and updates password_hash for an approved leader.
func ResetLeaderPasswordWithToken(ctx context.Context, db *sql.DB, rawTokenHex, newPasswordHash string) error {
	rawTokenHex = strings.TrimSpace(strings.ToLower(rawTokenHex))
	if len(rawTokenHex) != 64 {
		return ErrPasswordResetInvalidToken
	}
	rawBytes, err := hex.DecodeString(rawTokenHex)
	if err != nil || len(rawBytes) != 32 {
		return ErrPasswordResetInvalidToken
	}

	tokenHashHex := PasswordResetTokenHashHex(rawBytes)

	tx, err := db.BeginTx(ctx, nil)
	if err != nil {
		return err
	}
	defer func() { _ = tx.Rollback() }()

	var email string
	err = tx.QueryRowContext(ctx,
		`SELECT email FROM password_reset_tokens WHERE token_hash = ?
		   AND CAST(expires_at AS INTEGER) > CAST(strftime('%s','now') AS INTEGER)`,
		tokenHashHex,
	).Scan(&email)
	if errors.Is(err, sql.ErrNoRows) {
		return ErrPasswordResetInvalidToken
	}
	if err != nil {
		return err
	}

	if _, err := tx.ExecContext(ctx, `DELETE FROM password_reset_tokens WHERE token_hash = ?`, tokenHashHex); err != nil {
		return err
	}

	res, err := tx.ExecContext(ctx,
		`UPDATE leader_credentials SET password_hash = ? WHERE email = ? AND approved_at IS NOT NULL`,
		newPasswordHash, email,
	)
	if err != nil {
		return err
	}
	n, err := res.RowsAffected()
	if err != nil {
		return err
	}
	if n == 0 {
		return ErrPasswordResetInvalidToken
	}

	return tx.Commit()
}
