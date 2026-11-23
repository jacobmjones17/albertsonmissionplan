package store

import (
	"context"
	"errors"
	"os"
	"testing"
)

func testEventsStore(t *testing.T) {
	ctx := context.Background()
	db, err := Open(ctx)
	if err != nil {
		t.Fatal(err)
	}
	defer db.Close()
	_, _ = db.ExecContext(ctx, `DELETE FROM ward_events`)

	mk := func(title, date, start string, allDay bool) int64 {
		t.Helper()
		id, err := InsertEvent(ctx, db, WardEvent{Title: title, Date: date, StartTime: start, AllDay: allDay,
			Location: "Ward Building", Description: "Desc", LinkURL: "https://example.org"}, "lead@example.org")
		if err != nil {
			t.Fatal(err)
		}
		return id
	}
	past := mk("Past fireside", "2026-10-04", "19:00", false)
	evening := mk("Evening baptism", "2026-12-05", "18:00", false)
	morning := mk("Morning service", "2026-12-05", "09:00", false)
	allDay := mk("Temple day", "2026-12-05", "", true)
	later := mk("Open house", "2027-01-15", "17:00", false)

	up, err := ListUpcomingEvents(ctx, db, "2026-11-23")
	if err != nil {
		t.Fatal(err)
	}
	got := []int64{}
	for _, e := range up {
		got = append(got, e.ID)
	}
	want := []int64{allDay, morning, evening, later}
	if len(got) != len(want) {
		t.Fatalf("upcoming = %v, want %v", got, want)
	}
	for i := range want {
		if got[i] != want[i] {
			t.Fatalf("upcoming = %v, want %v", got, want)
		}
	}
	if !up[0].AllDay || up[0].Location != "Ward Building" || up[0].CreatedAt.IsZero() {
		t.Fatalf("scanned %+v", up[0])
	}
	pastList, err := ListPastEvents(ctx, db, "2025-11-23", "2026-11-23")
	if err != nil || len(pastList) != 1 || pastList[0].ID != past {
		t.Fatalf("past = %+v, %v", pastList, err)
	}
	all, err := ListAllEvents(ctx, db)
	if err != nil || len(all) != 5 || all[0].ID != past {
		t.Fatalf("all = %d, %v", len(all), err)
	}

	e, err := GetEvent(ctx, db, later)
	if err != nil {
		t.Fatal(err)
	}
	e.Title = "Christmas open house"
	e.Date = "2026-12-20"
	e.AllDay = true
	e.StartTime = ""
	if err := UpdateEvent(ctx, db, *e); err != nil {
		t.Fatal(err)
	}
	e2, _ := GetEvent(ctx, db, later)
	if e2.Title != "Christmas open house" || e2.Date != "2026-12-20" || !e2.AllDay || e2.StartTime != "" {
		t.Fatalf("updated = %+v", e2)
	}
	if err := UpdateEvent(ctx, db, WardEvent{ID: 999999, Title: "x", Date: "2026-12-01"}); !errors.Is(err, ErrEventNotFound) {
		t.Fatalf("update missing err = %v", err)
	}
	if err := DeleteEvent(ctx, db, past); err != nil {
		t.Fatal(err)
	}
	if err := DeleteEvent(ctx, db, past); !errors.Is(err, ErrEventNotFound) {
		t.Fatalf("second delete err = %v", err)
	}
	if _, err := GetEvent(ctx, db, past); !errors.Is(err, ErrEventNotFound) {
		t.Fatalf("get deleted err = %v", err)
	}
}

func TestEventsStoreSQLite(t *testing.T) {
	t.Setenv("DATABASE_URL", "")
	t.Setenv("DATABASE_PATH", t.TempDir()+"/e.db")
	testEventsStore(t)
}

// Set TEST_POSTGRES_URL (a throwaway database) to run the same checks against PostgreSQL.
func TestEventsStorePostgres(t *testing.T) {
	u := os.Getenv("TEST_POSTGRES_URL")
	if u == "" {
		t.Skip("TEST_POSTGRES_URL not set")
	}
	t.Setenv("DATABASE_URL", u)
	testEventsStore(t)
}
