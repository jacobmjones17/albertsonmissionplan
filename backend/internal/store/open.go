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
		db, err := sql.Open("pgx", normalizePostgresURL(u))
		if err != nil {
			return nil, err
		}
		// Render Postgres closes idle connections; keep the pool fresh on free tiers.
		db.SetMaxOpenConns(10)
		db.SetMaxIdleConns(4)
		db.SetConnMaxLifetime(5 * time.Minute)
		db.SetConnMaxIdleTime(2 * time.Minute)
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
		// Ward Gallery: drawings stored in the database (Render's disk is ephemeral). Image
		// bytes live in the row; list queries never select the two BYTEA columns.
		`CREATE TABLE IF NOT EXISTS gallery_items (
			id BIGSERIAL PRIMARY KEY,
			width INTEGER NOT NULL,
			height INTEGER NOT NULL,
			thumb_width INTEGER NOT NULL,
			thumb_height INTEGER NOT NULL,
			image_hash TEXT NOT NULL,
			image_data BYTEA NOT NULL,
			thumb_data BYTEA NOT NULL,
			hidden BOOLEAN NOT NULL DEFAULT FALSE,
			created_at TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP,
			created_by TEXT NOT NULL DEFAULT '',
			updated_at TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP
		);`,
		// Ward events calendar. Dates and times are ward-local wall-clock values
		// ("2026-12-05", "18:30") interpreted in MAIL_TIME_ZONE, so they never shift with DST.
		`CREATE TABLE IF NOT EXISTS ward_events (
			id BIGSERIAL PRIMARY KEY,
			title TEXT NOT NULL,
			event_date TEXT NOT NULL,
			all_day BOOLEAN NOT NULL DEFAULT FALSE,
			start_time TEXT NOT NULL DEFAULT '',
			end_time TEXT NOT NULL DEFAULT '',
			location TEXT NOT NULL DEFAULT '',
			description TEXT NOT NULL DEFAULT '',
			link_url TEXT NOT NULL DEFAULT '',
			created_at TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP,
			created_by TEXT NOT NULL DEFAULT '',
			updated_at TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP
		);`,
		`CREATE INDEX IF NOT EXISTS idx_ward_events_date ON ward_events(event_date);`,
		// Monthly Challenges: one row per month, editable by leaders. Seeded from the original
		// static content and artwork on first run.
		`CREATE TABLE IF NOT EXISTS monthly_challenges (
			month SMALLINT PRIMARY KEY CHECK (month BETWEEN 1 AND 12),
			heading TEXT NOT NULL,
			goal TEXT NOT NULL,
			bullets TEXT NOT NULL,
			hidden BOOLEAN NOT NULL DEFAULT FALSE,
			image_type TEXT NOT NULL,
			image_hash TEXT NOT NULL,
			image_data BYTEA NOT NULL,
			thumb_data BYTEA NOT NULL,
			image_custom BOOLEAN NOT NULL DEFAULT FALSE,
			updated_at TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP,
			updated_by TEXT NOT NULL DEFAULT ''
		);`,
	}
	for _, s := range stmts {
		if _, err := db.ExecContext(ctx, s); err != nil {
			return err
		}
	}
	if err := migrateSchemaFixes(ctx, db); err != nil {
		return err
	}
	if err := migrateOrgSectionTitles(ctx, db); err != nil {
		return err
	}
	if err := seedIfEmpty(ctx, db); err != nil {
		return err
	}
	return seedMonthlyChallenges(ctx, db)
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
		`CREATE TABLE IF NOT EXISTS gallery_items (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			width INTEGER NOT NULL,
			height INTEGER NOT NULL,
			thumb_width INTEGER NOT NULL,
			thumb_height INTEGER NOT NULL,
			image_hash TEXT NOT NULL,
			image_data BLOB NOT NULL,
			thumb_data BLOB NOT NULL,
			hidden INTEGER NOT NULL DEFAULT 0 CHECK (hidden IN (0, 1)),
			created_at TEXT NOT NULL DEFAULT (datetime('now')),
			created_by TEXT NOT NULL DEFAULT '',
			updated_at TEXT NOT NULL DEFAULT (datetime('now'))
		);`,
		`CREATE TABLE IF NOT EXISTS ward_events (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			title TEXT NOT NULL,
			event_date TEXT NOT NULL,
			all_day INTEGER NOT NULL DEFAULT 0 CHECK (all_day IN (0, 1)),
			start_time TEXT NOT NULL DEFAULT '',
			end_time TEXT NOT NULL DEFAULT '',
			location TEXT NOT NULL DEFAULT '',
			description TEXT NOT NULL DEFAULT '',
			link_url TEXT NOT NULL DEFAULT '',
			created_at TEXT NOT NULL DEFAULT (datetime('now')),
			created_by TEXT NOT NULL DEFAULT '',
			updated_at TEXT NOT NULL DEFAULT (datetime('now'))
		);`,
		`CREATE INDEX IF NOT EXISTS idx_ward_events_date ON ward_events(event_date);`,
		`CREATE TABLE IF NOT EXISTS monthly_challenges (
			month INTEGER PRIMARY KEY CHECK (month BETWEEN 1 AND 12),
			heading TEXT NOT NULL,
			goal TEXT NOT NULL,
			bullets TEXT NOT NULL,
			hidden INTEGER NOT NULL DEFAULT 0 CHECK (hidden IN (0, 1)),
			image_type TEXT NOT NULL,
			image_hash TEXT NOT NULL,
			image_data BLOB NOT NULL,
			thumb_data BLOB NOT NULL,
			image_custom INTEGER NOT NULL DEFAULT 0 CHECK (image_custom IN (0, 1)),
			updated_at TEXT NOT NULL DEFAULT (datetime('now')),
			updated_by TEXT NOT NULL DEFAULT ''
		);`,
	}
	for _, s := range stmts {
		if _, err := db.ExecContext(ctx, s); err != nil {
			return err
		}
	}
	if err := migrateSchemaFixes(ctx, db); err != nil {
		return err
	}
	if err := seedIfEmptySQLiteAddons(ctx, db); err != nil {
		return err
	}
	if err := migrateOrgSectionTitles(ctx, db); err != nil {
		return err
	}
	return seedMonthlyChallenges(ctx, db)
}

// normalizePostgresURL tunes the DSN for managed Postgres (e.g. Render): require TLS and
// avoid server-side prepared statements that break behind transaction-pooling proxies.
func normalizePostgresURL(raw string) string {
	raw = strings.TrimSpace(raw)
	if raw == "" {
		return raw
	}
	sep := "?"
	if strings.Contains(raw, "?") {
		sep = "&"
	}
	if !strings.Contains(raw, "sslmode=") {
		raw += sep + "sslmode=require"
		sep = "&"
	}
	if !strings.Contains(raw, "default_query_exec_mode=") {
		raw += sep + "default_query_exec_mode=simple_protocol"
	}
	return raw
}

func migrateSchemaFixes(ctx context.Context, db *sql.DB) error {
	if activeDriver == driverPostgreSQL {
		fixes := []string{
			`ALTER TABLE org_sections ADD COLUMN IF NOT EXISTS sort_order INTEGER NOT NULL DEFAULT 0`,
			`ALTER TABLE leader_credentials ADD COLUMN IF NOT EXISTS first_name TEXT NOT NULL DEFAULT ''`,
			`ALTER TABLE leader_credentials ADD COLUMN IF NOT EXISTS last_name TEXT NOT NULL DEFAULT ''`,
			`ALTER TABLE leader_credentials ADD COLUMN IF NOT EXISTS approved_at TIMESTAMPTZ`,
			`ALTER TABLE leader_credentials ADD COLUMN IF NOT EXISTS approved_by TEXT`,
		}
		for _, s := range fixes {
			if _, err := db.ExecContext(ctx, s); err != nil {
				return err
			}
		}
		return nil
	}
	if _, err := db.ExecContext(ctx,
		`ALTER TABLE org_sections ADD COLUMN sort_order INTEGER NOT NULL DEFAULT 0`,
	); err != nil && !strings.Contains(strings.ToLower(err.Error()), "duplicate column name") {
		return err
	}
	return nil
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
