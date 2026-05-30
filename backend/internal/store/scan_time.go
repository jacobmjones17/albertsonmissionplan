package store

import (
	"database/sql"
	"fmt"
	"strings"
	"time"
)

var sqliteTimeLayouts = []string{
	time.RFC3339Nano,
	time.RFC3339,
	"2006-01-02 15:04:05.999999999",
	"2006-01-02 15:04:05",
}

func parseTimeString(s string) (time.Time, error) {
	s = strings.TrimSpace(s)
	if s == "" {
		return time.Time{}, fmt.Errorf("empty time string")
	}
	for _, layout := range sqliteTimeLayouts {
		if t, err := time.Parse(layout, s); err == nil {
			return t, nil
		}
	}
	return time.Time{}, fmt.Errorf("cannot parse time %q", s)
}

func scanTimeValue(src any) (time.Time, error) {
	if src == nil {
		return time.Time{}, fmt.Errorf("null time")
	}
	switch v := src.(type) {
	case time.Time:
		return v, nil
	case string:
		return parseTimeString(v)
	case []byte:
		return parseTimeString(string(v))
	default:
		return time.Time{}, fmt.Errorf("unsupported time type %T", src)
	}
}

// nullTime scans TIMESTAMPTZ (Postgres) or TEXT datetimes (SQLite).
type nullTime struct {
	sql.NullTime
}

func (n *nullTime) Scan(src any) error {
	if src == nil {
		n.Valid = false
		n.Time = time.Time{}
		return nil
	}
	t, err := scanTimeValue(src)
	if err != nil {
		return err
	}
	n.Time = t
	n.Valid = true
	return nil
}

// wallTime scans a non-null timestamp from either driver.
type wallTime struct {
	time.Time
}

func (t *wallTime) Scan(src any) error {
	parsed, err := scanTimeValue(src)
	if err != nil {
		return err
	}
	t.Time = parsed
	return nil
}
