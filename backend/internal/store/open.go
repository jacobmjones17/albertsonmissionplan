package store

import (
	"context"
	"database/sql"
	"os"
	"strings"
	"time"

	_ "github.com/jackc/pgx/v5/stdlib"
	_ "modernc.org/sqlite"
)

// Open connects to Postgres when DATABASE_URL is set; otherwise SQLite at DATABASE_PATH
// (default wardmission.db). Same schema and migrations are applied automatically.
func Open(ctx context.Context) (*sql.DB, error) {
	if u := strings.TrimSpace(os.Getenv("DATABASE_URL")); u != "" {
		activeDriver = driverPostgreSQL
		db, err := sql.Open("pgx", u)
		if err != nil {
			return nil, err
		}
		db.SetMaxOpenConns(25)
		db.SetConnMaxLifetime(time.Hour)
		if err := db.PingContext(ctx); err != nil {
			_ = db.Close()
			return nil, err
		}
		if err := migratePostgres(ctx, db); err != nil {
			_ = db.Close()
			return nil, err
		}
		return db, nil
	}

	activeDriver = driverSQLite
	path := strings.TrimSpace(os.Getenv("DATABASE_PATH"))
	if path == "" {
		path = "wardmission.db"
	}
	db, err := sql.Open("sqlite", "file:"+path+"?_foreign_keys=on")
	if err != nil {
		return nil, err
	}
	db.SetMaxOpenConns(1)
	db.SetConnMaxLifetime(0)
	if err := db.PingContext(ctx); err != nil {
		_ = db.Close()
		return nil, err
	}
	if err := migrateSQLite(ctx, db); err != nil {
		_ = db.Close()
		return nil, err
	}
	return db, nil
}

func migratePostgres(ctx context.Context, db *sql.DB) error {
	stmts := []string{
		`CREATE TABLE IF NOT EXISTS org_sections (
			slug TEXT PRIMARY KEY,
			title TEXT NOT NULL,
			bullets TEXT NOT NULL,
			sort_order INTEGER NOT NULL
		);`,
		`CREATE TABLE IF NOT EXISTS ward_goals (
			id SMALLINT PRIMARY KEY CHECK (id = 1),
			bullets TEXT NOT NULL,
			updated_at TIMESTAMPTZ,
			updated_by TEXT
		);`,
		`CREATE TABLE IF NOT EXISTS testimonials (
			id BIGSERIAL PRIMARY KEY,
			body TEXT NOT NULL,
			author_label TEXT NOT NULL DEFAULT '',
			status TEXT NOT NULL DEFAULT 'pending',
			created_at TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP,
			reviewed_at TIMESTAMPTZ,
			reviewed_by TEXT,
			moderator_note TEXT
		);`,
		`CREATE INDEX IF NOT EXISTS idx_testimonials_status ON testimonials(status);`,
		`CREATE TABLE IF NOT EXISTS leader_roles (
			email TEXT PRIMARY KEY NOT NULL,
			is_admin BOOLEAN NOT NULL DEFAULT FALSE,
			created_at TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP,
			updated_at TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP
		);`,
		`CREATE TABLE IF NOT EXISTS leader_org_scopes (
			email TEXT NOT NULL REFERENCES leader_roles(email) ON DELETE CASCADE,
			slug TEXT NOT NULL REFERENCES org_sections(slug) ON DELETE CASCADE,
			PRIMARY KEY (email, slug)
		);`,
		`CREATE TABLE IF NOT EXISTS leader_credentials (
			email TEXT PRIMARY KEY NOT NULL,
			password_hash TEXT NOT NULL,
			first_name TEXT NOT NULL DEFAULT '',
			last_name TEXT NOT NULL DEFAULT '',
			created_at TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP,
			approved_at TIMESTAMPTZ,
			approved_by TEXT
		);`,
		`CREATE TABLE IF NOT EXISTS password_reset_tokens (
			token_hash TEXT PRIMARY KEY NOT NULL,
			email TEXT NOT NULL,
			expires_at BIGINT NOT NULL,
			created_at TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP
		);`,
		`CREATE INDEX IF NOT EXISTS idx_password_reset_email ON password_reset_tokens(email);`,
	}
	for _, s := range stmts {
		if _, err := db.ExecContext(ctx, s); err != nil {
			return err
		}
	}
	if err := migrateOrgSectionTitles(ctx, db); err != nil {
		return err
	}
	return seedIfEmpty(ctx, db)
}

func migrateSQLite(ctx context.Context, db *sql.DB) error {
	stmts := []string{
		`CREATE TABLE IF NOT EXISTS org_sections (
			slug TEXT PRIMARY KEY,
			title TEXT NOT NULL,
			bullets TEXT NOT NULL,
			sort_order INTEGER NOT NULL
		);`,
		`CREATE TABLE IF NOT EXISTS ward_goals (
			id INTEGER PRIMARY KEY CHECK (id = 1),
			bullets TEXT NOT NULL,
			updated_at TEXT,
			updated_by TEXT
		);`,
		`CREATE TABLE IF NOT EXISTS testimonials (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			body TEXT NOT NULL,
			author_label TEXT NOT NULL DEFAULT '',
			status TEXT NOT NULL DEFAULT 'pending',
			created_at TEXT NOT NULL DEFAULT (datetime('now')),
			reviewed_at TEXT,
			reviewed_by TEXT,
			moderator_note TEXT
		);`,
		`CREATE INDEX IF NOT EXISTS idx_testimonials_status ON testimonials(status);`,
		`CREATE TABLE IF NOT EXISTS leader_roles (
			email TEXT PRIMARY KEY NOT NULL,
			is_admin INTEGER NOT NULL DEFAULT 0 CHECK (is_admin IN (0, 1)),
			created_at TEXT NOT NULL DEFAULT (datetime('now')),
			updated_at TEXT NOT NULL DEFAULT (datetime('now'))
		);`,
		`CREATE TABLE IF NOT EXISTS leader_org_scopes (
			email TEXT NOT NULL REFERENCES leader_roles(email) ON DELETE CASCADE,
			slug TEXT NOT NULL REFERENCES org_sections(slug) ON DELETE CASCADE,
			PRIMARY KEY (email, slug)
		);`,
		`CREATE TABLE IF NOT EXISTS leader_credentials (
			email TEXT PRIMARY KEY NOT NULL,
			password_hash TEXT NOT NULL,
			first_name TEXT NOT NULL DEFAULT '',
			last_name TEXT NOT NULL DEFAULT '',
			created_at TEXT NOT NULL DEFAULT (datetime('now')),
			approved_at TEXT,
			approved_by TEXT
		);`,
		`CREATE TABLE IF NOT EXISTS password_reset_tokens (
			token_hash TEXT PRIMARY KEY NOT NULL,
			email TEXT NOT NULL,
			expires_at INTEGER NOT NULL,
			created_at TEXT NOT NULL DEFAULT (datetime('now'))
		);`,
		`CREATE INDEX IF NOT EXISTS idx_password_reset_email ON password_reset_tokens(email);`,
	}
	for _, s := range stmts {
		if _, err := db.ExecContext(ctx, s); err != nil {
			return err
		}
	}
	if err := seedIfEmptySQLiteAddons(ctx, db); err != nil {
		return err
	}
	return migrateOrgSectionTitles(ctx, db)
}

func seedIfEmptySQLiteAddons(ctx context.Context, db *sql.DB) error {
	// Match historical SQLite columns for upgrades (no-op when columns exist).
	additive := []string{
		`ALTER TABLE leader_credentials ADD COLUMN approved_at TEXT;`,
		`ALTER TABLE leader_credentials ADD COLUMN approved_by TEXT;`,
		`ALTER TABLE leader_credentials ADD COLUMN first_name TEXT NOT NULL DEFAULT '';`,
		`ALTER TABLE leader_credentials ADD COLUMN last_name TEXT NOT NULL DEFAULT '';`,
	}
	for _, s := range additive {
		if _, err := db.ExecContext(ctx, s); err != nil &&
			!strings.Contains(strings.ToLower(err.Error()), "duplicate column name") {
			return err
		}
	}
	return seedIfEmpty(ctx, db)
}
