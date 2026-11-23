package store

import (
	"context"
	"crypto/sha256"
	"database/sql"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"time"

	"github.com/albertson/albertsonmissionplan/internal/challenges"
)

// MonthlyChallenge is one month's challenge (image bytes are fetched separately). Month (1–12)
// is both the key and the display order.
type MonthlyChallenge struct {
	Month       int
	Heading     string
	Goal        string
	Bullets     []challenges.Bullet
	Hidden      bool
	ImageType   string
	ImageHash   string
	ImageCustom bool // a leader uploaded a replacement picture
	UpdatedAt   time.Time
	UpdatedBy   string
}

// ChallengeText is the editable text of one month.
type ChallengeText struct {
	Heading string
	Goal    string
	Bullets []challenges.Bullet
	Hidden  bool
}

var ErrChallengeNotFound = errors.New("monthly challenge not found")

const challengeCols = `month, heading, goal, bullets, hidden, image_type, image_hash, image_custom, updated_at, updated_by`

func artHash(b []byte) string {
	sum := sha256.Sum256(b)
	return hex.EncodeToString(sum[:6])
}

func encodeBullets(b []challenges.Bullet) (string, error) {
	if b == nil {
		b = []challenges.Bullet{}
	}
	out, err := json.Marshal(b)
	return string(out), err
}

// seedMonthlyChallenges inserts the original challenge (text and artwork) for every month that
// has no row yet, so a fresh or upgraded database shows exactly what the site showed before
// challenges became editable. Existing rows are never touched.
func seedMonthlyChallenges(ctx context.Context, db *sql.DB) error {
	have := map[int]bool{}
	rows, err := db.QueryContext(ctx, `SELECT month FROM monthly_challenges`)
	if err != nil {
		return err
	}
	for rows.Next() {
		var m int
		if err := rows.Scan(&m); err != nil {
			rows.Close()
			return err
		}
		have[m] = true
	}
	rows.Close()
	if err := rows.Err(); err != nil {
		return err
	}
	for _, d := range challenges.Defaults {
		if have[d.Month] {
			continue
		}
		art, err := challenges.DefaultArt(d.Month)
		if err != nil {
			return err
		}
		bullets, err := encodeBullets(d.Bullets)
		if err != nil {
			return err
		}
		if _, err := sxExec(db, ctx, `INSERT INTO monthly_challenges (month, heading, goal, bullets, hidden,
			image_type, image_hash, image_data, thumb_data, image_custom, updated_at, updated_by)
			VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, CURRENT_TIMESTAMP, '')
			ON CONFLICT (month) DO NOTHING`,
			d.Month, d.Heading, d.Goal, bullets, false, art.ContentType, artHash(art.Image), art.Image, art.Thumb, false,
		); err != nil {
			return fmt.Errorf("seed monthly challenge %d: %w", d.Month, err)
		}
	}
	return nil
}

func scanChallenges(rows *sql.Rows) ([]MonthlyChallenge, error) {
	defer rows.Close()
	var out []MonthlyChallenge
	for rows.Next() {
		var c MonthlyChallenge
		var bullets string
		var updated nullTime
		var by sql.NullString
		if err := rows.Scan(&c.Month, &c.Heading, &c.Goal, &bullets, &c.Hidden, &c.ImageType, &c.ImageHash,
			&c.ImageCustom, &updated, &by); err != nil {
			return nil, err
		}
		if err := json.Unmarshal([]byte(bullets), &c.Bullets); err != nil {
			return nil, fmt.Errorf("monthly challenge %d bullets: %w", c.Month, err)
		}
		if updated.Valid {
			c.UpdatedAt = updated.Time
		}
		c.UpdatedBy = by.String
		out = append(out, c)
	}
	return out, rows.Err()
}

// ListMonthlyChallenges returns the challenges January → December; includeHidden is for leaders.
func ListMonthlyChallenges(ctx context.Context, db *sql.DB, includeHidden bool) ([]MonthlyChallenge, error) {
	q := `SELECT ` + challengeCols + ` FROM monthly_challenges`
	var rows *sql.Rows
	var err error
	if includeHidden {
		rows, err = sxQuery(db, ctx, q+` ORDER BY month`)
	} else {
		rows, err = sxQuery(db, ctx, q+` WHERE hidden = $1 ORDER BY month`, false)
	}
	if err != nil {
		return nil, err
	}
	return scanChallenges(rows)
}

// GetMonthlyChallenge returns one month's challenge.
func GetMonthlyChallenge(ctx context.Context, db *sql.DB, month int) (*MonthlyChallenge, error) {
	rows, err := sxQuery(db, ctx, `SELECT `+challengeCols+` FROM monthly_challenges WHERE month = $1`, month)
	if err != nil {
		return nil, err
	}
	list, err := scanChallenges(rows)
	if err != nil {
		return nil, err
	}
	if len(list) == 0 {
		return nil, ErrChallengeNotFound
	}
	return &list[0], nil
}

func challengeAffected(res sql.Result, err error) error {
	if err != nil {
		return err
	}
	n, err := res.RowsAffected()
	if err != nil {
		return err
	}
	if n == 0 {
		return ErrChallengeNotFound
	}
	return nil
}

// UpdateMonthlyChallengeText saves a month's heading, goal, ideas, and visibility.
func UpdateMonthlyChallengeText(ctx context.Context, db *sql.DB, month int, t ChallengeText, editor string) error {
	bullets, err := encodeBullets(t.Bullets)
	if err != nil {
		return err
	}
	res, err := sxExec(db, ctx, `UPDATE monthly_challenges SET heading = $1, goal = $2, bullets = $3, hidden = $4,
		updated_at = CURRENT_TIMESTAMP, updated_by = $5 WHERE month = $6`,
		t.Heading, t.Goal, bullets, t.Hidden, editor, month)
	return challengeAffected(res, err)
}

// SetMonthlyChallengeImage replaces a month's picture (custom=false when restoring the original).
func SetMonthlyChallengeImage(ctx context.Context, db *sql.DB, month int, art challenges.Art, custom bool, editor string) error {
	res, err := sxExec(db, ctx, `UPDATE monthly_challenges SET image_type = $1, image_hash = $2, image_data = $3,
		thumb_data = $4, image_custom = $5, updated_at = CURRENT_TIMESTAMP, updated_by = $6 WHERE month = $7`,
		art.ContentType, artHash(art.Image), art.Image, art.Thumb, custom, editor, month)
	return challengeAffected(res, err)
}

// ResetMonthlyChallenge restores a month's original text, picture, and visibility.
func ResetMonthlyChallenge(ctx context.Context, db *sql.DB, month int, editor string) error {
	d, ok := challenges.ForMonth(month)
	if !ok {
		return ErrChallengeNotFound
	}
	art, err := challenges.DefaultArt(month)
	if err != nil {
		return err
	}
	if err := UpdateMonthlyChallengeText(ctx, db, month, ChallengeText{Heading: d.Heading, Goal: d.Goal, Bullets: d.Bullets}, editor); err != nil {
		return err
	}
	return SetMonthlyChallengeImage(ctx, db, month, art, false, editor)
}

type challengeRow struct {
	heading, goal, bullets, imageType, imageHash string
	hidden, custom                               bool
	image, thumb                                 []byte
}

// SwapMonthlyChallenges moves month a's challenge (text, picture, visibility) to month b and
// vice versa, e.g. to run the Gratitude challenge in November instead of July.
func SwapMonthlyChallenges(ctx context.Context, db *sql.DB, a, b int, editor string) error {
	if a == b {
		return nil
	}
	tx, err := db.BeginTx(ctx, nil)
	if err != nil {
		return err
	}
	defer func() { _ = tx.Rollback() }()
	read := func(m int) (challengeRow, error) {
		var r challengeRow
		err := sxTxQueryRow(tx, ctx, `SELECT heading, goal, bullets, hidden, image_type, image_hash, image_custom,
			image_data, thumb_data FROM monthly_challenges WHERE month = $1`, m).
			Scan(&r.heading, &r.goal, &r.bullets, &r.hidden, &r.imageType, &r.imageHash, &r.custom, &r.image, &r.thumb)
		if errors.Is(err, sql.ErrNoRows) {
			return r, ErrChallengeNotFound
		}
		return r, err
	}
	write := func(m int, r challengeRow) error {
		_, err := sxTxExec(tx, ctx, `UPDATE monthly_challenges SET heading = $1, goal = $2, bullets = $3, hidden = $4,
			image_type = $5, image_hash = $6, image_custom = $7, image_data = $8, thumb_data = $9,
			updated_at = CURRENT_TIMESTAMP, updated_by = $10 WHERE month = $11`,
			r.heading, r.goal, r.bullets, r.hidden, r.imageType, r.imageHash, r.custom, r.image, r.thumb, editor, m)
		return err
	}
	ra, err := read(a)
	if err != nil {
		return err
	}
	rb, err := read(b)
	if err != nil {
		return err
	}
	// A picture counts as "original" only in the month it shipped with.
	ra.custom = !isDefaultArtFor(b, ra.imageHash)
	rb.custom = !isDefaultArtFor(a, rb.imageHash)
	if err := write(a, rb); err != nil {
		return err
	}
	if err := write(b, ra); err != nil {
		return err
	}
	return tx.Commit()
}

func isDefaultArtFor(month int, hash string) bool {
	art, err := challenges.DefaultArt(month)
	return err == nil && artHash(art.Image) == hash
}

// MonthlyChallengeImage returns a month's display image (thumb=false) or thumbnail, its content
// type, hash, and whether the month is hidden.
func MonthlyChallengeImage(ctx context.Context, db *sql.DB, month int, thumb bool) (data []byte, contentType, hash string, hidden bool, err error) {
	col := "image_data"
	if thumb {
		col = "thumb_data"
	}
	err = sxQueryRow(db, ctx, `SELECT `+col+`, image_type, image_hash, hidden FROM monthly_challenges WHERE month = $1`, month).
		Scan(&data, &contentType, &hash, &hidden)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, "", "", false, ErrChallengeNotFound
	}
	return data, contentType, hash, hidden, err
}

// IsDefaultArt reports whether hash is the original artwork for month.
func IsDefaultArt(month int, hash string) bool { return isDefaultArtFor(month, hash) }
