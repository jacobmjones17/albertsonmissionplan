package store

import (
	"context"
	"database/sql"
	"regexp"
)

type driverKind int

const (
	driverPostgreSQL driverKind = iota
	driverSQLite
)

// activeDriver is set by Open before any queries run.
var activeDriver driverKind

var pgPlaceholderRe = regexp.MustCompile(`\$[1-9][0-9]*`)

func bindSQL(q string) string {
	if activeDriver != driverSQLite {
		return q
	}
	return pgPlaceholderRe.ReplaceAllString(q, "?")
}

func sxExec(db *sql.DB, ctx context.Context, q string, args ...any) (sql.Result, error) {
	return db.ExecContext(ctx, bindSQL(q), args...)
}

func sxQuery(db *sql.DB, ctx context.Context, q string, args ...any) (*sql.Rows, error) {
	return db.QueryContext(ctx, bindSQL(q), args...)
}

func sxQueryRow(db *sql.DB, ctx context.Context, q string, args ...any) *sql.Row {
	return db.QueryRowContext(ctx, bindSQL(q), args...)
}

func sxTxExec(tx *sql.Tx, ctx context.Context, q string, args ...any) (sql.Result, error) {
	return tx.ExecContext(ctx, bindSQL(q), args...)
}

func sxTxQueryRow(tx *sql.Tx, ctx context.Context, q string, args ...any) *sql.Row {
	return tx.QueryRowContext(ctx, bindSQL(q), args...)
}
