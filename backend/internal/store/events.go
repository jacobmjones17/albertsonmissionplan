package store

import (
	"context"
	"database/sql"
	"errors"
	"time"
)

// WardEvent is one entry on the public events calendar. Date and times are ward-local
// wall-clock strings: Date "2006-01-02", StartTime/EndTime "15:04" (empty when AllDay).
type WardEvent struct {
	ID          int64
	Title       string
	Date        string
	AllDay      bool
	StartTime   string
	EndTime     string
	Location    string
	Description string
	LinkURL     string
	CreatedAt   time.Time
	UpdatedAt   time.Time
}

var ErrEventNotFound = errors.New("event not found")

const eventCols = `id, title, event_date, all_day, start_time, end_time, location, description, link_url, created_at, updated_at`

// Chronological: by date, all-day events first that day, then by start time.
const eventOrderAsc = ` ORDER BY event_date ASC, all_day DESC, start_time ASC, id ASC`
const eventOrderDesc = ` ORDER BY event_date DESC, all_day ASC, start_time DESC, id DESC`

func scanEvents(rows *sql.Rows) ([]WardEvent, error) {
	defer rows.Close()
	var out []WardEvent
	for rows.Next() {
		var e WardEvent
		var created, updated nullTime
		if err := rows.Scan(&e.ID, &e.Title, &e.Date, &e.AllDay, &e.StartTime, &e.EndTime,
			&e.Location, &e.Description, &e.LinkURL, &created, &updated); err != nil {
			return nil, err
		}
		if created.Valid {
			e.CreatedAt = created.Time
		}
		if updated.Valid {
			e.UpdatedAt = updated.Time
		}
		out = append(out, e)
	}
	return out, rows.Err()
}

// ListUpcomingEvents returns events on or after fromDate ("2006-01-02"), soonest first.
func ListUpcomingEvents(ctx context.Context, db *sql.DB, fromDate string) ([]WardEvent, error) {
	rows, err := sxQuery(db, ctx, `SELECT `+eventCols+` FROM ward_events WHERE event_date >= $1`+eventOrderAsc, fromDate)
	if err != nil {
		return nil, err
	}
	return scanEvents(rows)
}

// ListPastEvents returns events before beforeDate and on or after sinceDate, most recent first.
func ListPastEvents(ctx context.Context, db *sql.DB, sinceDate, beforeDate string) ([]WardEvent, error) {
	rows, err := sxQuery(db, ctx, `SELECT `+eventCols+` FROM ward_events WHERE event_date < $1 AND event_date >= $2`+eventOrderDesc,
		beforeDate, sinceDate)
	if err != nil {
		return nil, err
	}
	return scanEvents(rows)
}

// ListAllEvents returns every event in chronological order (leader manager).
func ListAllEvents(ctx context.Context, db *sql.DB) ([]WardEvent, error) {
	rows, err := sxQuery(db, ctx, `SELECT `+eventCols+` FROM ward_events`+eventOrderAsc)
	if err != nil {
		return nil, err
	}
	return scanEvents(rows)
}

// GetEvent returns one event.
func GetEvent(ctx context.Context, db *sql.DB, id int64) (*WardEvent, error) {
	rows, err := sxQuery(db, ctx, `SELECT `+eventCols+` FROM ward_events WHERE id = $1`, id)
	if err != nil {
		return nil, err
	}
	list, err := scanEvents(rows)
	if err != nil {
		return nil, err
	}
	if len(list) == 0 {
		return nil, ErrEventNotFound
	}
	return &list[0], nil
}

// InsertEvent stores a new event (already validated) and returns its id.
func InsertEvent(ctx context.Context, db *sql.DB, e WardEvent, createdBy string) (int64, error) {
	args := []any{e.Title, e.Date, e.AllDay, e.StartTime, e.EndTime, e.Location, e.Description, e.LinkURL, createdBy}
	q := `INSERT INTO ward_events (title, event_date, all_day, start_time, end_time, location, description,
		link_url, created_by, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, CURRENT_TIMESTAMP, CURRENT_TIMESTAMP)`
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

// UpdateEvent replaces an event's fields.
func UpdateEvent(ctx context.Context, db *sql.DB, e WardEvent) error {
	res, err := sxExec(db, ctx, `UPDATE ward_events SET title = $1, event_date = $2, all_day = $3, start_time = $4,
		end_time = $5, location = $6, description = $7, link_url = $8, updated_at = CURRENT_TIMESTAMP WHERE id = $9`,
		e.Title, e.Date, e.AllDay, e.StartTime, e.EndTime, e.Location, e.Description, e.LinkURL, e.ID)
	return eventAffected(res, err)
}

// DeleteEvent permanently removes an event.
func DeleteEvent(ctx context.Context, db *sql.DB, id int64) error {
	res, err := sxExec(db, ctx, `DELETE FROM ward_events WHERE id = $1`, id)
	return eventAffected(res, err)
}

func eventAffected(res sql.Result, err error) error {
	if err != nil {
		return err
	}
	n, err := res.RowsAffected()
	if err != nil {
		return err
	}
	if n == 0 {
		return ErrEventNotFound
	}
	return nil
}
