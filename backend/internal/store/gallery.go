package store

import (
	"context"
	"database/sql"
	"errors"
	"time"
)

// GalleryItem is one drawing's metadata (image bytes are fetched separately).
type GalleryItem struct {
	ID          int64
	Title       string
	Caption     string
	Artist      string
	AgeGroup    string
	Width       int
	Height      int
	ThumbWidth  int
	ThumbHeight int
	ImageHash   string
	SortOrder   int
	Featured    bool
	Hidden      bool
	CreatedAt   time.Time
}

// GalleryItemInput is the editable text + flags of a drawing.
type GalleryItemInput struct {
	Title    string
	Caption  string
	Artist   string
	AgeGroup string
	Featured bool
	Hidden   bool
}

// GalleryImages holds the processed bytes for a new drawing.
type GalleryImages struct {
	Width, Height           int
	ThumbWidth, ThumbHeight int
	Hash                    string
	Image, Thumb            []byte
}

var ErrGalleryNotFound = errors.New("gallery item not found")

const galleryCols = `id, title, caption, artist, age_group, width, height, thumb_width, thumb_height,
	image_hash, sort_order, featured, hidden, created_at`

// Featured drawings first, then the leaders' order (new uploads get the lowest sort_order so
// they appear at the top until someone reorders).
const galleryOrder = ` ORDER BY featured DESC, sort_order ASC, id DESC`

func scanGalleryItems(rows *sql.Rows) ([]GalleryItem, error) {
	defer rows.Close()
	var out []GalleryItem
	for rows.Next() {
		var g GalleryItem
		var created nullTime
		if err := rows.Scan(&g.ID, &g.Title, &g.Caption, &g.Artist, &g.AgeGroup, &g.Width, &g.Height,
			&g.ThumbWidth, &g.ThumbHeight, &g.ImageHash, &g.SortOrder, &g.Featured, &g.Hidden, &created); err != nil {
			return nil, err
		}
		if created.Valid {
			g.CreatedAt = created.Time
		}
		out = append(out, g)
	}
	return out, rows.Err()
}

// ListGalleryItems returns drawings in display order; includeHidden is for leaders.
func ListGalleryItems(ctx context.Context, db *sql.DB, includeHidden bool) ([]GalleryItem, error) {
	q := `SELECT ` + galleryCols + ` FROM gallery_items`
	if !includeHidden {
		q += ` WHERE hidden = $1`
		rows, err := sxQuery(db, ctx, q+galleryOrder, false)
		if err != nil {
			return nil, err
		}
		return scanGalleryItems(rows)
	}
	rows, err := sxQuery(db, ctx, q+galleryOrder)
	if err != nil {
		return nil, err
	}
	return scanGalleryItems(rows)
}

// GetGalleryItem returns one drawing's metadata.
func GetGalleryItem(ctx context.Context, db *sql.DB, id int64) (*GalleryItem, error) {
	rows, err := sxQuery(db, ctx, `SELECT `+galleryCols+` FROM gallery_items WHERE id = $1`, id)
	if err != nil {
		return nil, err
	}
	items, err := scanGalleryItems(rows)
	if err != nil {
		return nil, err
	}
	if len(items) == 0 {
		return nil, ErrGalleryNotFound
	}
	return &items[0], nil
}

// InsertGalleryItem stores a new drawing at the top of the order.
func InsertGalleryItem(ctx context.Context, db *sql.DB, in GalleryItemInput, img GalleryImages, createdBy string) (int64, error) {
	var minOrder sql.NullInt64
	if err := sxQueryRow(db, ctx, `SELECT MIN(sort_order) FROM gallery_items`).Scan(&minOrder); err != nil {
		return 0, err
	}
	order := 0
	if minOrder.Valid {
		order = int(minOrder.Int64) - 1
	}
	args := []any{in.Title, in.Caption, in.Artist, in.AgeGroup, img.Width, img.Height, img.ThumbWidth,
		img.ThumbHeight, img.Hash, img.Image, img.Thumb, order, in.Featured, in.Hidden, createdBy}
	q := `INSERT INTO gallery_items (title, caption, artist, age_group, width, height, thumb_width,
		thumb_height, image_hash, image_data, thumb_data, sort_order, featured, hidden, created_by,
		created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, $14, $15, CURRENT_TIMESTAMP, CURRENT_TIMESTAMP)`
	if activeDriver == driverSQLite {
		res, err := sxExec(db, ctx, q, args...)
		if err != nil {
			return 0, err
		}
		return res.LastInsertId()
	}
	var id int64
	err := sxQueryRow(db, ctx, q+` RETURNING id`, args...).Scan(&id)
	return id, err
}

// UpdateGalleryItem changes a drawing's text and flags.
func UpdateGalleryItem(ctx context.Context, db *sql.DB, id int64, in GalleryItemInput) error {
	res, err := sxExec(db, ctx, `UPDATE gallery_items SET title = $1, caption = $2, artist = $3,
		age_group = $4, featured = $5, hidden = $6, updated_at = CURRENT_TIMESTAMP WHERE id = $7`,
		in.Title, in.Caption, in.Artist, in.AgeGroup, in.Featured, in.Hidden, id)
	return affectedOrNotFound(res, err)
}

// DeleteGalleryItem removes a drawing and its image bytes.
func DeleteGalleryItem(ctx context.Context, db *sql.DB, id int64) error {
	res, err := sxExec(db, ctx, `DELETE FROM gallery_items WHERE id = $1`, id)
	return affectedOrNotFound(res, err)
}

// ReorderGalleryItems sets sort_order to each id's position in ids. Ids not listed keep their
// order (after the listed ones).
func ReorderGalleryItems(ctx context.Context, db *sql.DB, ids []int64) error {
	tx, err := db.BeginTx(ctx, nil)
	if err != nil {
		return err
	}
	defer func() { _ = tx.Rollback() }()
	for i, id := range ids {
		if _, err := sxTxExec(tx, ctx, `UPDATE gallery_items SET sort_order = $1 WHERE id = $2`, i, id); err != nil {
			return err
		}
	}
	return tx.Commit()
}

// GalleryImage returns the stored display image (thumb=false) or thumbnail, plus the item's
// hidden flag and content hash.
func GalleryImage(ctx context.Context, db *sql.DB, id int64, thumb bool) (data []byte, hash string, hidden bool, err error) {
	col := "image_data"
	if thumb {
		col = "thumb_data"
	}
	err = sxQueryRow(db, ctx, `SELECT `+col+`, image_hash, hidden FROM gallery_items WHERE id = $1`, id).
		Scan(&data, &hash, &hidden)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, "", false, ErrGalleryNotFound
	}
	return data, hash, hidden, err
}

// GalleryCounts returns total and hidden drawing counts.
func GalleryCounts(ctx context.Context, db *sql.DB) (total, hidden int, err error) {
	err = sxQueryRow(db, ctx, `SELECT COUNT(*), COALESCE(SUM(CASE WHEN hidden THEN 1 ELSE 0 END), 0) FROM gallery_items`).
		Scan(&total, &hidden)
	return total, hidden, err
}

func affectedOrNotFound(res sql.Result, err error) error {
	if err != nil {
		return err
	}
	n, err := res.RowsAffected()
	if err != nil {
		return err
	}
	if n == 0 {
		return ErrGalleryNotFound
	}
	return nil
}
