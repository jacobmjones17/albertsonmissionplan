package store

import (
	"context"
	"database/sql"
	"errors"
	"strings"
	"time"

	_ "modernc.org/sqlite"
)

var (
	// ErrApproveBodyTooShort is returned when the leader's edited text is too short to publish.
	ErrApproveBodyTooShort = errors.New("edited experience is too short to publish")
	// ErrApproveBodyTooLong is returned when the leader's edited text exceeds the limit.
	ErrApproveBodyTooLong = errors.New("edited experience is too long")
)

// OrgSection is one auxiliary / quorum block on the ward plan.
type OrgSection struct {
	Slug    string
	Title   string
	Bullets []string
}

// Testimonial is a member-submitted experience (may be pending).
type Testimonial struct {
	ID            int64
	Body          string
	AuthorLabel   string
	Status        string
	CreatedAt     time.Time
	ReviewedAt    sql.NullTime
	ReviewedBy    sql.NullString
	ModeratorNote sql.NullString
}

// Open opens the SQLite database and runs migrations.
func Open(ctx context.Context, path string) (*sql.DB, error) {
	db, err := sql.Open("sqlite", "file:"+path+"?_foreign_keys=on")
	if err != nil {
		return nil, err
	}
	db.SetMaxOpenConns(1)
	if err := db.PingContext(ctx); err != nil {
		_ = db.Close()
		return nil, err
	}
	if err := migrate(ctx, db); err != nil {
		_ = db.Close()
		return nil, err
	}
	return db, nil
}

func migrate(ctx context.Context, db *sql.DB) error {
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
			created_at TEXT NOT NULL,
			reviewed_at TEXT,
			reviewed_by TEXT,
			moderator_note TEXT
		);`,
		`CREATE INDEX IF NOT EXISTS idx_testimonials_status ON testimonials(status);`,
		`CREATE TABLE IF NOT EXISTS leader_roles (
			email TEXT PRIMARY KEY NOT NULL,
			is_admin INTEGER NOT NULL DEFAULT 0 CHECK (is_admin IN (0, 1)),
			created_at TEXT NOT NULL DEFAULT '',
			updated_at TEXT NOT NULL DEFAULT ''
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
	}
	for _, s := range stmts {
		if _, err := db.ExecContext(ctx, s); err != nil {
			return err
		}
	}
	// Additive migrations for older databases. We run these unconditionally and ignore
	// "duplicate column" errors so a fresh CREATE TABLE above stays valid too.
	additive := []string{
		`ALTER TABLE leader_credentials ADD COLUMN approved_at TEXT;`,
		`ALTER TABLE leader_credentials ADD COLUMN approved_by TEXT;`,
		`ALTER TABLE leader_credentials ADD COLUMN first_name TEXT NOT NULL DEFAULT '';`,
		`ALTER TABLE leader_credentials ADD COLUMN last_name TEXT NOT NULL DEFAULT '';`,
	}
	for _, s := range additive {
		if _, err := db.ExecContext(ctx, s); err != nil && !strings.Contains(strings.ToLower(err.Error()), "duplicate column name") {
			return err
		}
	}
	tableCreates := []string{
		`CREATE TABLE IF NOT EXISTS password_reset_tokens (
			token_hash TEXT PRIMARY KEY NOT NULL,
			email TEXT NOT NULL,
			expires_at TEXT NOT NULL,
			created_at TEXT NOT NULL DEFAULT (datetime('now'))
		);`,
		`CREATE INDEX IF NOT EXISTS idx_password_reset_email ON password_reset_tokens(email);`,
	}
	for _, s := range tableCreates {
		if _, err := db.ExecContext(ctx, s); err != nil {
			return err
		}
	}
	return seedIfEmpty(ctx, db)
}

func seedIfEmpty(ctx context.Context, db *sql.DB) error {
	var n int
	if err := db.QueryRowContext(ctx, `SELECT COUNT(*) FROM org_sections`).Scan(&n); err != nil {
		return err
	}
	if n > 0 {
		return seedWardGoalsIfEmpty(ctx, db)
	}

	type seed struct {
		slug    string
		title   string
		sort    int
		bullets string
	}
	seeds := []seed{
		{"elders-quorum", "Elders Quorum", 10, strings.Join([]string{
			"Monitor the new convert list and help them in their progress wherever possible.",
			"Help new converts visit the temple within one month of their baptism.",
			"Conduct an activity each quarter in which non-members and less-active members can attend.",
			"Coordinate and work with the full-time missionaries at least once a month.",
			"Participate in ministering in all forms.",
			"Invite all to take monthly challenges.",
		}, "\n")},
		{"relief-society", "Relief Society", 20, strings.Join([]string{
			"Invite non-members and less-active members to attend activities.",
			"Plan an activity each quarter in which non-members and less-active members can attend.",
			"Coordinate and work with the full-time missionaries at least once a month.",
			"Participate in ministering in all forms.",
			"Invite all to take monthly challenges.",
		}, "\n")},
		{"aaronic-priesthood", "Aaronic Priesthood (Young Men)", 30, strings.Join([]string{
			"Invite at least one non-member or less-active member to attend activities.",
			"Accompany and fellowship a less-active member, non-member, or people being taught (ages 11–17).",
			"Participate in ministering in all forms.",
			"Participate in indexing.",
			"Submit at least two names to the temple.",
			"Combine with another youth group in the community to conduct a service project at least once a year.",
			"Exercise and perform Aaronic Priesthood ordinances.",
		}, "\n")},
		{"young-women", "Young Women", 40, strings.Join([]string{
			"Invite at least one non-member or less-active member to attend activities.",
			"Participate in ministering in all forms.",
			"Accompany and fellowship a less-active member, non-member, or people being taught by the missionaries in their age group.",
			"Participate in indexing.",
			"Submit at least two names to the temple.",
			"Combine with another youth group in the community to conduct a service project at least once a year.",
		}, "\n")},
		{"primary", "Primary", 50, strings.Join([]string{
			"Motivate adult members or adult people being taught by the missionaries to have their children attend Primary classes and activities.",
			"Invite all to take monthly challenges.",
		}, "\n")},
		{"sunday-school", "Sunday School", 60, strings.Join([]string{
			"Introduce and welcome returning members and non-members at the beginning of each class.",
			"Invite all to participate in monthly activities.",
		}, "\n")},
		{"bishopric", "Bishopric", 70, strings.Join([]string{
			"Invite all those speaking in sacrament meeting to invite non-members to hear their talk.",
			"Ensure ward activities are planned to appropriately invite non-member friends.",
			"Encourage all families to join the Love, Share, and Invite mission focus.",
			"Encourage all families to participate in the ward’s monthly challenge.",
		}, "\n")},
		{"missionary-coordination", "Missionary coordination", 80, strings.Join([]string{
			"Conduct consistent weekly coordination meetings.",
			"Conduct consistent ward council meetings to discuss progress and the covenant path of recent converts.",
			"Work on the prospective elders list, as assigned by the bishop.",
			"Visit all families that are new or recently moved into ward boundaries.",
			"Help members learn and execute Love, Share, and Invite.",
			"Invite all to take monthly challenges.",
		}, "\n")},
		{"ward-council", "Ward council", 90, strings.Join([]string{
			"Lead by example—take on and complete the monthly challenge as a ward council.",
			"Invite and encourage members of your quorums and organizations to participate in the ward monthly challenge.",
			"Ensure ward activities focus on inviting non-member and less-active friends.",
			"Invite all to participate in monthly activities.",
		}, "\n")},
	}
	tx, err := db.BeginTx(ctx, nil)
	if err != nil {
		return err
	}
	for _, r := range seeds {
		if _, err := tx.ExecContext(ctx,
			`INSERT INTO org_sections (slug, title, bullets, sort_order) VALUES (?, ?, ?, ?)`,
			r.slug, r.title, r.bullets, r.sort,
		); err != nil {
			_ = tx.Rollback()
			return err
		}
	}
	if err := tx.Commit(); err != nil {
		return err
	}
	return seedWardGoalsIfEmpty(ctx, db)
}

func seedWardGoalsIfEmpty(ctx context.Context, db *sql.DB) error {
	var n int
	if err := db.QueryRowContext(ctx, `SELECT COUNT(*) FROM ward_goals`).Scan(&n); err != nil {
		return err
	}
	if n > 0 {
		return nil
	}
	bullets := strings.Join([]string{
		"Strengthen **ministering efforts** with consistency.",
		"Focus missionary and member efforts on **baptism** and helping new members prepare for the **temple**.",
		"Support **strong sacrament meeting attendance** and a welcoming experience for members and visitors.",
		"**Reach out in love** to less-active families and help them progress on the covenant path toward the **temple**.",
		"**Reach out** to each less-active member with pastoral care and fellowship on the covenant path.",
		"Implementing the principles of [Love, Share, and Invite](/love-share-invite).",
		"Encouraging the entire ward family to **participate** in [monthly challenges](/monthly-challenges).",
		"Encouraging each active household to create their own **family mission plan**.",
	}, "\n")
	_, err := db.ExecContext(ctx,
		`INSERT INTO ward_goals (id, bullets, updated_at, updated_by) VALUES (1, ?, '', '')`,
		bullets,
	)
	return err
}

// ListOrgSections returns organizations in display order.
func ListOrgSections(ctx context.Context, db *sql.DB) ([]OrgSection, error) {
	rows, err := db.QueryContext(ctx,
		`SELECT slug, title, bullets FROM org_sections ORDER BY sort_order ASC, slug ASC`,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var out []OrgSection
	for rows.Next() {
		var slug, title, raw string
		if err := rows.Scan(&slug, &title, &raw); err != nil {
			return nil, err
		}
		out = append(out, OrgSection{
			Slug:    slug,
			Title:   title,
			Bullets: splitBullets(raw),
		})
	}
	return out, rows.Err()
}

// UpdateOrgBullets updates bullet list for one org (leaders).
func UpdateOrgBullets(ctx context.Context, db *sql.DB, slug, bullets string) error {
	res, err := db.ExecContext(ctx,
		`UPDATE org_sections SET bullets = ? WHERE slug = ?`,
		strings.TrimSpace(bullets), slug,
	)
	if err != nil {
		return err
	}
	n, err := res.RowsAffected()
	if err != nil {
		return err
	}
	if n == 0 {
		return errors.New("unknown organization")
	}
	return nil
}

func splitBullets(s string) []string {
	lines := strings.Split(strings.ReplaceAll(s, "\r\n", "\n"), "\n")
	var out []string
	for _, ln := range lines {
		ln = strings.TrimSpace(ln)
		if ln != "" {
			out = append(out, ln)
		}
	}
	return out
}

// GetWardGoals returns ward-wide goal lines.
func GetWardGoals(ctx context.Context, db *sql.DB) ([]string, error) {
	var raw string
	err := db.QueryRowContext(ctx, `SELECT bullets FROM ward_goals WHERE id = 1`).Scan(&raw)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	return splitBullets(raw), nil
}

// SetWardGoals replaces ward-wide bullets (one per line in storage).
func SetWardGoals(ctx context.Context, db *sql.DB, bullets string, editor string) error {
	res, err := db.ExecContext(ctx,
		`UPDATE ward_goals SET bullets = ?, updated_at = datetime('now'), updated_by = ? WHERE id = 1`,
		strings.TrimSpace(bullets), editor,
	)
	if err != nil {
		return err
	}
	n, err := res.RowsAffected()
	if err != nil {
		return err
	}
	if n > 0 {
		return nil
	}
	_, err = db.ExecContext(ctx,
		`INSERT INTO ward_goals (id, bullets, updated_at, updated_by) VALUES (1, ?, datetime('now'), ?)`,
		strings.TrimSpace(bullets), editor,
	)
	return err
}

// InsertTestimonial creates a pending testimonial.
func InsertTestimonial(ctx context.Context, db *sql.DB, body, authorLabel string) (int64, error) {
	res, err := db.ExecContext(ctx,
		`INSERT INTO testimonials (body, author_label, status, created_at) VALUES (?, ?, 'pending', datetime('now'))`,
		body, strings.TrimSpace(authorLabel),
	)
	if err != nil {
		return 0, err
	}
	return res.LastInsertId()
}

// ListApprovedTestimonials returns recent approved posts for public display.
func ListApprovedTestimonials(ctx context.Context, db *sql.DB, limit int) ([]Testimonial, error) {
	if limit <= 0 {
		limit = 50
	}
	rows, err := db.QueryContext(ctx,
		`SELECT id, body, author_label, status, created_at, reviewed_at, reviewed_by, moderator_note
		 FROM testimonials WHERE status = 'approved' ORDER BY datetime(created_at) DESC LIMIT ?`,
		limit,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	return scanTestimonials(rows)
}

// ListPendingTestimonials returns testimonials awaiting moderation.
func ListPendingTestimonials(ctx context.Context, db *sql.DB) ([]Testimonial, error) {
	rows, err := db.QueryContext(ctx,
		`SELECT id, body, author_label, status, created_at, reviewed_at, reviewed_by, moderator_note
		 FROM testimonials WHERE status = 'pending' ORDER BY datetime(created_at) ASC`,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	return scanTestimonials(rows)
}

// SetTestimonialStatus sets rejected (no body change) for pending rows.
func SetTestimonialStatus(ctx context.Context, db *sql.DB, id int64, status string, reviewer string, note string) error {
	if status != "rejected" {
		return errors.New("invalid status for SetTestimonialStatus")
	}
	res, err := db.ExecContext(ctx,
		`UPDATE testimonials SET status = ?, reviewed_at = datetime('now'), reviewed_by = ?, moderator_note = ?
		 WHERE id = ? AND status = 'pending'`,
		status, reviewer, strings.TrimSpace(note), id,
	)
	if err != nil {
		return err
	}
	n, err := res.RowsAffected()
	if err != nil {
		return err
	}
	if n == 0 {
		return errors.New("submission not found or already moderated")
	}
	return nil
}

// ApproveTestimonial publishes a pending submission, optionally replacing body and display name (e.g. after redacting names).
func ApproveTestimonial(ctx context.Context, db *sql.DB, id int64, body, authorLabel, reviewer string) error {
	body = strings.TrimSpace(body)
	if len(body) < 10 {
		return ErrApproveBodyTooShort
	}
	if len(body) > 8000 {
		return ErrApproveBodyTooLong
	}
	res, err := db.ExecContext(ctx,
		`UPDATE testimonials SET body = ?, author_label = ?, status = 'approved',
			reviewed_at = datetime('now'), reviewed_by = ?, moderator_note = ''
		 WHERE id = ? AND status = 'pending'`,
		body, strings.TrimSpace(authorLabel), reviewer, id,
	)
	if err != nil {
		return err
	}
	n, err := res.RowsAffected()
	if err != nil {
		return err
	}
	if n == 0 {
		return errors.New("submission not found or already moderated")
	}
	return nil
}

// UpdateApprovedTestimonial edits an already-published experience. Used by admins to fix typos,
// redact identifying details that slipped through review, or update the author label. Re-stamps
// reviewed_at/reviewed_by so the audit trail records who last touched the live row. Refuses to
// operate on pending/rejected rows (those go through ApproveTestimonial / SetTestimonialStatus).
func UpdateApprovedTestimonial(ctx context.Context, db *sql.DB, id int64, body, authorLabel, editor string) error {
	body = strings.TrimSpace(body)
	if len(body) < 10 {
		return ErrApproveBodyTooShort
	}
	if len(body) > 8000 {
		return ErrApproveBodyTooLong
	}
	res, err := db.ExecContext(ctx,
		`UPDATE testimonials SET body = ?, author_label = ?,
			reviewed_at = datetime('now'), reviewed_by = ?
		 WHERE id = ? AND status = 'approved'`,
		body, strings.TrimSpace(authorLabel), editor, id,
	)
	if err != nil {
		return err
	}
	n, err := res.RowsAffected()
	if err != nil {
		return err
	}
	if n == 0 {
		return errors.New("published experience not found")
	}
	return nil
}

// DeleteTestimonial permanently removes a row regardless of status. Used by admins to take a
// published experience down. No soft-delete: once removed it cannot be restored, which is what
// you want when the goal is "get this off the public site immediately."
func DeleteTestimonial(ctx context.Context, db *sql.DB, id int64) error {
	res, err := db.ExecContext(ctx, `DELETE FROM testimonials WHERE id = ?`, id)
	if err != nil {
		return err
	}
	n, err := res.RowsAffected()
	if err != nil {
		return err
	}
	if n == 0 {
		return errors.New("experience not found")
	}
	return nil
}

func scanTestimonials(rows *sql.Rows) ([]Testimonial, error) {
	var out []Testimonial
	for rows.Next() {
		var t Testimonial
		var created string
		var reviewed sql.NullString
		var reviewedBy, modNote sql.NullString
		if err := rows.Scan(&t.ID, &t.Body, &t.AuthorLabel, &t.Status, &created, &reviewed, &reviewedBy, &modNote); err != nil {
			return nil, err
		}
		t.CreatedAt = parseSQLiteTime(created)
		if reviewed.Valid {
			if rt := parseSQLiteTime(reviewed.String); !rt.IsZero() {
				t.ReviewedAt = sql.NullTime{Time: rt, Valid: true}
			}
		}
		t.ReviewedBy = reviewedBy
		t.ModeratorNote = modNote
		out = append(out, t)
	}
	return out, rows.Err()
}

func parseSQLiteTime(s string) time.Time {
	s = strings.TrimSpace(s)
	if s == "" {
		return time.Time{}
	}
	layouts := []string{
		"2006-01-02 15:04:05",
		time.RFC3339,
		"2006-01-02T15:04:05Z07:00",
	}
	for _, l := range layouts {
		if t, err := time.ParseInLocation(l, s, time.Local); err == nil {
			return t
		}
		if t, err := time.Parse(l, s); err == nil {
			return t
		}
	}
	return time.Time{}
}
