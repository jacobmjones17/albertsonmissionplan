package events

import (
	"errors"
	"net/url"
	"strings"
	"testing"
	"time"

	"github.com/albertson/albertsonmissionplan/internal/store"
)

func denver(t *testing.T) *time.Location {
	t.Helper()
	loc, err := time.LoadLocation("America/Denver")
	if err != nil {
		t.Fatal(err)
	}
	return loc
}

func newYork(t *testing.T) *time.Location {
	t.Helper()
	loc, err := time.LoadLocation("America/New_York")
	if err != nil {
		t.Fatal(err)
	}
	return loc
}

func TestNormalizeCleansInput(t *testing.T) {
	e, err := Normalize(Input{
		Title: "  Ward   Christmas\nParty ", Date: "2026-12-12", StartTime: "18:00", EndTime: "20:00",
		Location: " Albertson Ward\nBuilding ", Description: "\r\nBring a side dish.\r\nAll are welcome! ", Link: "churchofjesuschrist.org/comeuntochrist",
	})
	if err != nil {
		t.Fatal(err)
	}
	if e.Title != "Ward Christmas Party" || e.Location != "Albertson Ward Building" {
		t.Fatalf("one-line fields: %q / %q", e.Title, e.Location)
	}
	if e.Description != "Bring a side dish.\nAll are welcome!" {
		t.Fatalf("description %q", e.Description)
	}
	if e.LinkURL != "https://churchofjesuschrist.org/comeuntochrist" {
		t.Fatalf("link %q", e.LinkURL)
	}
	if e.StartTime != "18:00" || e.EndTime != "20:00" || e.AllDay {
		t.Fatalf("times %+v", e)
	}
}

func TestNormalizeAllDayClearsTimes(t *testing.T) {
	e, err := Normalize(Input{Title: "Stake Temple Day", Date: "2027-01-09", AllDay: true, StartTime: "09:00", EndTime: "x"})
	if err != nil {
		t.Fatal(err)
	}
	if e.StartTime != "" || e.EndTime != "" || !e.AllDay {
		t.Fatalf("all day kept times: %+v", e)
	}
}

func TestNormalizeRejects(t *testing.T) {
	ok := Input{Title: "Fireside", Date: "2026-12-06", StartTime: "19:00"}
	cases := []struct {
		field string
		mut   func(*Input)
	}{
		{"title", func(in *Input) { in.Title = "  " }},
		{"title", func(in *Input) { in.Title = strings.Repeat("a", MaxTitle+1) }},
		{"date", func(in *Input) { in.Date = "12/06/2026" }},
		{"date", func(in *Input) { in.Date = "1926-12-06" }},
		{"startTime", func(in *Input) { in.StartTime = "" }},
		{"startTime", func(in *Input) { in.StartTime = "7pm" }},
		{"endTime", func(in *Input) { in.EndTime = "18:00" }},
		{"endTime", func(in *Input) { in.EndTime = "19:00" }},
		{"endTime", func(in *Input) { in.EndTime = "25:00" }},
		{"location", func(in *Input) { in.Location = strings.Repeat("b", MaxLocation+1) }},
		{"description", func(in *Input) { in.Description = strings.Repeat("c", MaxDescription+1) }},
		{"link", func(in *Input) { in.Link = "javascript:alert(1)" }},
		{"link", func(in *Input) { in.Link = "ftp://example.org/x" }},
	}
	for _, c := range cases {
		in := ok
		c.mut(&in)
		_, err := Normalize(in)
		var ve *ValidationError
		if !errors.As(err, &ve) || ve.Field != c.field {
			t.Errorf("%+v: err = %v, want field %s", in, err, c.field)
		}
	}
	if _, err := Normalize(ok); err != nil {
		t.Fatalf("valid input rejected: %v", err)
	}
}

func TestLabels(t *testing.T) {
	cases := []struct {
		e    store.WardEvent
		want string
	}{
		{store.WardEvent{AllDay: true}, "All day"},
		{store.WardEvent{StartTime: "18:30"}, "6:30 PM"},
		{store.WardEvent{StartTime: "18:30", EndTime: "20:00"}, "6:30 – 8:00 PM"},
		{store.WardEvent{StartTime: "10:00", EndTime: "13:00"}, "10:00 AM – 1:00 PM"},
	}
	for _, c := range cases {
		if got := TimeLabel(c.e); got != c.want {
			t.Errorf("TimeLabel(%+v) = %q, want %q", c.e, got, c.want)
		}
	}
	e := store.WardEvent{Date: "2026-12-05"}
	if DateLabel(e) != "Saturday, December 5, 2026" || MonthKey(e) != "2026-12" || MonthLabel(e) != "December 2026" {
		t.Fatalf("date labels %q %q %q", DateLabel(e), MonthKey(e), MonthLabel(e))
	}
}

func TestBoundsAndZoneRespectDST(t *testing.T) {
	loc := denver(t)
	winter := store.WardEvent{Date: "2026-12-05", StartTime: "18:00", EndTime: "19:30"}
	s, e := Bounds(winter, loc)
	if s.UTC().Format(utcStamp) != "20261206T010000Z" || e.UTC().Format(utcStamp) != "20261206T023000Z" {
		t.Fatalf("winter bounds %s %s", s.UTC(), e.UTC())
	}
	if ZoneAbbrev(winter, loc) != "MST" {
		t.Fatalf("winter zone %s", ZoneAbbrev(winter, loc))
	}
	summer := store.WardEvent{Date: "2027-07-10", StartTime: "18:00"}
	s, e = Bounds(summer, loc)
	if s.UTC().Format(utcStamp) != "20270711T000000Z" || e.Sub(s) != time.Hour {
		t.Fatalf("summer bounds %s %s", s.UTC(), e.UTC())
	}
	if ZoneAbbrev(summer, loc) != "MDT" {
		t.Fatalf("summer zone %s", ZoneAbbrev(summer, loc))
	}
}

func TestGoogleCalendarURL(t *testing.T) {
	loc := denver(t)
	e := store.WardEvent{ID: 3, Title: "Baptism", Date: "2026-12-05", StartTime: "10:00", EndTime: "11:00",
		Location: "Albertson Ward Building", Description: "All are welcome."}
	u, err := url.Parse(GoogleCalendarURL(e, loc, "https://albertsonwardplan.com/events"))
	if err != nil {
		t.Fatal(err)
	}
	q := u.Query()
	if u.Host != "calendar.google.com" || q.Get("action") != "TEMPLATE" || q.Get("text") != "Baptism" ||
		q.Get("dates") != "20261205T170000Z/20261205T180000Z" || q.Get("ctz") != "America/Denver" ||
		q.Get("location") != "Albertson Ward Building" || !strings.Contains(q.Get("details"), "All are welcome.") {
		t.Fatalf("google url %s", u)
	}
	allDay := store.WardEvent{Title: "Camp", Date: "2026-12-31", AllDay: true}
	u, _ = url.Parse(GoogleCalendarURL(allDay, loc, ""))
	if u.Query().Get("dates") != "20261231/20270101" {
		t.Fatalf("all-day dates %s", u.Query().Get("dates"))
	}
}

func TestICS(t *testing.T) {
	loc := denver(t)
	now := time.Date(2026, 11, 23, 8, 0, 0, 0, time.UTC)
	e := store.WardEvent{ID: 7, Title: "Fireside; music, and testimony", Date: "2026-12-06", StartTime: "19:00",
		Location: "Stake Center, Room 4", Description: "Line one\nLine two " + strings.Repeat("é", 60),
		LinkURL: "https://example.org/fireside"}
	out := string(ICS(e, loc, "albertsonwardplan.com", "https://albertsonwardplan.com/events", now))
	for _, want := range []string{
		"BEGIN:VCALENDAR\r\n", "VERSION:2.0\r\n", "UID:ward-event-7@albertsonwardplan.com\r\n",
		"DTSTAMP:20261123T080000Z\r\n", "BEGIN:VTIMEZONE\r\n", "TZID:America/Denver\r\n",
		"DTSTART;TZID=America/Denver:20261206T190000\r\n", "DTEND;TZID=America/Denver:20261206T200000\r\n",
		`SUMMARY:Fireside\; music\, and testimony` + "\r\n", `LOCATION:Stake Center\, Room 4` + "\r\n",
		"URL:https://example.org/fireside\r\n", "END:VCALENDAR\r\n",
	} {
		if !strings.Contains(out, want) {
			t.Errorf("ics missing %q in:\n%s", want, out)
		}
	}
	for _, l := range strings.Split(out, "\r\n") {
		if len(l) > 75 {
			t.Errorf("unfolded line (%d octets): %q", len(l), l)
		}
	}
	if strings.Contains(strings.ReplaceAll(out, "\r\n", ""), "\n") {
		t.Error("bare LF in ics")
	}
	allDay := string(ICS(store.WardEvent{ID: 8, Title: "Open House", Date: "2026-12-31", AllDay: true}, loc, "", "", now))
	if !strings.Contains(allDay, "DTSTART;VALUE=DATE:20261231\r\n") || !strings.Contains(allDay, "DTEND;VALUE=DATE:20270101\r\n") {
		t.Fatalf("all-day ics:\n%s", allDay)
	}
}

func TestCalendarExportsEasternTime(t *testing.T) {
	loc := newYork(t)
	now := time.Date(2026, 11, 23, 8, 0, 0, 0, time.UTC)
	page := "https://albertsonwardplan.com/events"

	winter := store.WardEvent{ID: 12, Title: "Christmas Brunch", Date: "2026-12-05", StartTime: "10:00", EndTime: "12:00",
		Location: "3554 NC-903, Albertson, NC 28508"}
	s, e := Bounds(winter, loc)
	if s.UTC().Format(utcStamp) != "20261205T150000Z" || e.UTC().Format(utcStamp) != "20261205T170000Z" {
		t.Fatalf("winter (EST) bounds %s %s", s.UTC(), e.UTC())
	}
	if ZoneAbbrev(winter, loc) != "EST" {
		t.Fatalf("winter zone %s", ZoneAbbrev(winter, loc))
	}
	ics := string(ICS(winter, loc, "albertsonwardplan.com", page, now))
	for _, want := range []string{
		"BEGIN:VTIMEZONE\r\n",
		"TZID:America/New_York\r\n",
		"DTSTART;TZID=America/New_York:20261205T100000\r\n",
		"DTEND;TZID=America/New_York:20261205T120000\r\n",
		"RRULE:FREQ=YEARLY;BYMONTH=3;BYDAY=2SU\r\n",
		"RRULE:FREQ=YEARLY;BYMONTH=11;BYDAY=1SU\r\n",
		"Albertson Ward events: https://albertsonwardplan.com/events",
	} {
		if !strings.Contains(ics, want) {
			t.Errorf("winter ics missing %q in:\n%s", want, ics)
		}
	}
	if strings.Contains(ics, "T170000Z") || strings.Contains(ics, "DTSTART:20261205T") {
		t.Errorf("winter ics should use TZID local times, not UTC Z stamps:\n%s", ics)
	}
	u, err := url.Parse(GoogleCalendarURL(winter, loc, page))
	if err != nil {
		t.Fatal(err)
	}
	q := u.Query()
	if q.Get("dates") != "20261205T150000Z/20261205T170000Z" || q.Get("ctz") != "America/New_York" {
		t.Fatalf("winter google dates %s ctz %s", q.Get("dates"), q.Get("ctz"))
	}
	if !strings.Contains(q.Get("details"), "https://albertsonwardplan.com/events") {
		t.Fatalf("winter google details %s", q.Get("details"))
	}

	summer := store.WardEvent{ID: 13, Title: "Ward Picnic", Date: "2026-07-11", StartTime: "10:00", EndTime: "12:00"}
	s, e = Bounds(summer, loc)
	if s.UTC().Format(utcStamp) != "20260711T140000Z" || e.UTC().Format(utcStamp) != "20260711T160000Z" {
		t.Fatalf("summer (EDT) bounds %s %s", s.UTC(), e.UTC())
	}
	if ZoneAbbrev(summer, loc) != "EDT" {
		t.Fatalf("summer zone %s", ZoneAbbrev(summer, loc))
	}
	ics = string(ICS(summer, loc, "albertsonwardplan.com", page, now))
	if !strings.Contains(ics, "DTSTART;TZID=America/New_York:20260711T100000\r\n") ||
		!strings.Contains(ics, "DTEND;TZID=America/New_York:20260711T120000\r\n") {
		t.Fatalf("summer ics:\n%s", ics)
	}
	u, _ = url.Parse(GoogleCalendarURL(summer, loc, page))
	if u.Query().Get("dates") != "20260711T140000Z/20260711T160000Z" || u.Query().Get("ctz") != "America/New_York" {
		t.Fatalf("summer google %s", u)
	}
}

func TestToday(t *testing.T) {
	loc := denver(t)
	// 3:30 AM UTC on Nov 23 is still Nov 22 evening in Denver.
	if got := Today(loc, time.Date(2026, 11, 23, 3, 30, 0, 0, time.UTC)); got != "2026-11-22" {
		t.Fatalf("today %s", got)
	}
}
