// Package events holds the rules for the ward events calendar: validating leader input,
// human-friendly date/time labels, and "add to calendar" exports (.ics and Google Calendar).
//
// Dates and times are stored as ward-local wall-clock values (see store.WardEvent) and only
// converted to absolute instants here, using the ward time zone (MAIL_TIME_ZONE).
package events

import (
	"fmt"
	"net/url"
	"strings"
	"time"
	"unicode/utf8"

	"github.com/albertson/albertsonmissionplan/internal/store"
)

const (
	DateLayout = "2006-01-02"
	TimeLayout = "15:04"

	MaxTitle       = 120
	MaxLocation    = 200
	MaxDescription = 2000
	MaxLink        = 500
)

// Input is what the leader form submits.
type Input struct {
	Title       string `json:"title"`
	Date        string `json:"date"`
	AllDay      bool   `json:"allDay"`
	StartTime   string `json:"startTime"`
	EndTime     string `json:"endTime"`
	Location    string `json:"location"`
	Description string `json:"description"`
	Link        string `json:"link"`
}

// ValidationError names the form field that needs fixing.
type ValidationError struct {
	Field   string
	Message string
}

func (e *ValidationError) Error() string { return e.Message }

func invalid(field, msg string) error { return &ValidationError{Field: field, Message: msg} }

// oneLine trims and collapses all whitespace (including newlines) to single spaces.
func oneLine(s string) string { return strings.Join(strings.Fields(s), " ") }

// Normalize validates leader input and returns the cleaned event (ID unset).
func Normalize(in Input) (store.WardEvent, error) {
	var e store.WardEvent
	e.Title = oneLine(in.Title)
	if e.Title == "" {
		return e, invalid("title", "Give the event a title.")
	}
	if utf8.RuneCountInString(e.Title) > MaxTitle {
		return e, invalid("title", fmt.Sprintf("Keep the title under %d characters.", MaxTitle))
	}

	d, err := time.Parse(DateLayout, strings.TrimSpace(in.Date))
	if err != nil {
		return e, invalid("date", "Choose the date of the event.")
	}
	if d.Year() < 2000 || d.Year() > 2100 {
		return e, invalid("date", "That date doesn't look right. Check the year.")
	}
	e.Date = d.Format(DateLayout)

	e.AllDay = in.AllDay
	if !e.AllDay {
		st, err := time.Parse(TimeLayout, strings.TrimSpace(in.StartTime))
		if err != nil {
			return e, invalid("startTime", "Add a start time, or mark the event as all day.")
		}
		e.StartTime = st.Format(TimeLayout)
		if end := strings.TrimSpace(in.EndTime); end != "" {
			et, err := time.Parse(TimeLayout, end)
			if err != nil {
				return e, invalid("endTime", "That end time doesn't look right.")
			}
			if !et.After(st) {
				return e, invalid("endTime", "The end time needs to be after the start time.")
			}
			e.EndTime = et.Format(TimeLayout)
		}
	}

	e.Location = oneLine(in.Location)
	if utf8.RuneCountInString(e.Location) > MaxLocation {
		return e, invalid("location", fmt.Sprintf("Keep the location under %d characters.", MaxLocation))
	}

	desc := strings.ReplaceAll(strings.ReplaceAll(in.Description, "\r\n", "\n"), "\r", "\n")
	e.Description = strings.TrimSpace(desc)
	if utf8.RuneCountInString(e.Description) > MaxDescription {
		return e, invalid("description", fmt.Sprintf("Keep the description under %d characters.", MaxDescription))
	}

	link, err := normalizeLink(in.Link)
	if err != nil {
		return e, err
	}
	e.LinkURL = link
	return e, nil
}

func normalizeLink(raw string) (string, error) {
	s := strings.TrimSpace(raw)
	if s == "" {
		return "", nil
	}
	if !strings.Contains(s, "://") {
		s = "https://" + s
	}
	u, err := url.Parse(s)
	if err != nil || (u.Scheme != "http" && u.Scheme != "https") || u.Host == "" || strings.ContainsAny(u.Host, " <>\"") {
		return "", invalid("link", "Enter a full web address, like https://www.churchofjesuschrist.org/…")
	}
	out := u.String()
	if len(out) > MaxLink {
		return "", invalid("link", "That link is too long.")
	}
	return out, nil
}

// Today returns the current date in the ward's time zone.
func Today(loc *time.Location, now time.Time) string {
	return now.In(loc).Format(DateLayout)
}

// Bounds returns the event's start and end instants. All-day events span midnight to
// midnight; a timed event with no end time is treated as one hour long for calendar exports.
func Bounds(e store.WardEvent, loc *time.Location) (start, end time.Time) {
	d, _ := time.ParseInLocation(DateLayout, e.Date, loc)
	if e.AllDay || e.StartTime == "" {
		return d, d.AddDate(0, 0, 1)
	}
	st, _ := time.Parse(TimeLayout, e.StartTime)
	start = time.Date(d.Year(), d.Month(), d.Day(), st.Hour(), st.Minute(), 0, 0, loc)
	end = start.Add(time.Hour)
	if e.EndTime != "" {
		et, _ := time.Parse(TimeLayout, e.EndTime)
		end = time.Date(d.Year(), d.Month(), d.Day(), et.Hour(), et.Minute(), 0, 0, loc)
	}
	return start, end
}

// DateLabel is e.g. "Saturday, December 5, 2026".
func DateLabel(e store.WardEvent) string {
	d, err := time.Parse(DateLayout, e.Date)
	if err != nil {
		return e.Date
	}
	return d.Format("Monday, January 2, 2006")
}

// MonthKey/MonthLabel group events: "2026-12" / "December 2026".
func MonthKey(e store.WardEvent) string {
	if len(e.Date) >= 7 {
		return e.Date[:7]
	}
	return e.Date
}

func MonthLabel(e store.WardEvent) string {
	d, err := time.Parse(DateLayout, e.Date)
	if err != nil {
		return MonthKey(e)
	}
	return d.Format("January 2006")
}

func clock(hhmm string) (string, string) {
	t, err := time.Parse(TimeLayout, hhmm)
	if err != nil {
		return hhmm, ""
	}
	return t.Format("3:04"), t.Format("PM")
}

// TimeLabel is e.g. "All day", "6:30 PM", "6:30 – 8:00 PM", "10:00 AM – 1:00 PM".
func TimeLabel(e store.WardEvent) string {
	if e.AllDay || e.StartTime == "" {
		return "All day"
	}
	sc, sp := clock(e.StartTime)
	if e.EndTime == "" {
		return sc + " " + sp
	}
	ec, ep := clock(e.EndTime)
	if sp == ep {
		return sc + " – " + ec + " " + ep
	}
	return sc + " " + sp + " – " + ec + " " + ep
}

// ZoneAbbrev is the ward zone's abbreviation on the event's date, e.g. "EST" or "EDT".
func ZoneAbbrev(e store.WardEvent, loc *time.Location) string {
	start, _ := Bounds(e, loc)
	if e.AllDay {
		start = start.Add(12 * time.Hour)
	}
	name, _ := start.Zone()
	return name
}

// MapURL opens the location in Google Maps (Apple devices hand it to Maps), or "" if none.
func MapURL(location string) string {
	if strings.TrimSpace(location) == "" {
		return ""
	}
	return "https://www.google.com/maps/search/?api=1&query=" + url.QueryEscape(location)
}

const utcStamp = "20060102T150405Z"

// GoogleCalendarURL opens a pre-filled "new event" page in Google Calendar.
func GoogleCalendarURL(e store.WardEvent, loc *time.Location, pageURL string) string {
	start, end := Bounds(e, loc)
	var dates string
	if e.AllDay || e.StartTime == "" {
		dates = start.Format("20060102") + "/" + end.Format("20060102")
	} else {
		dates = start.UTC().Format(utcStamp) + "/" + end.UTC().Format(utcStamp)
	}
	q := url.Values{}
	q.Set("action", "TEMPLATE")
	q.Set("text", e.Title)
	q.Set("dates", dates)
	q.Set("ctz", loc.String())
	if d := calendarDetails(e, pageURL); d != "" {
		q.Set("details", d)
	}
	if e.Location != "" {
		q.Set("location", e.Location)
	}
	return "https://calendar.google.com/calendar/render?" + q.Encode()
}

func calendarDetails(e store.WardEvent, pageURL string) string {
	parts := []string{}
	if e.Description != "" {
		parts = append(parts, e.Description)
	}
	if e.LinkURL != "" {
		parts = append(parts, e.LinkURL)
	}
	if pageURL != "" {
		parts = append(parts, "Albertson Ward events: "+pageURL)
	}
	return strings.Join(parts, "\n\n")
}

// ICS renders a single-event iCalendar file. host makes the UID globally unique.
func ICS(e store.WardEvent, loc *time.Location, host, pageURL string, now time.Time) []byte {
	start, end := Bounds(e, loc)
	if host == "" {
		host = "albertsonwardplan.com"
	}
	var b strings.Builder
	line := func(s string) { b.WriteString(foldICS(s)) }
	line("BEGIN:VCALENDAR")
	line("VERSION:2.0")
	line("PRODID:-//Albertson Ward//Ward Events//EN")
	line("CALSCALE:GREGORIAN")
	line("METHOD:PUBLISH")
	if !(e.AllDay || e.StartTime == "") {
		writeVTimeZone(line, loc, start.In(loc).Year(), end.In(loc).Year())
	}
	line("BEGIN:VEVENT")
	line(fmt.Sprintf("UID:ward-event-%d@%s", e.ID, host))
	line("DTSTAMP:" + now.UTC().Format(utcStamp))
	if e.AllDay || e.StartTime == "" {
		line("DTSTART;VALUE=DATE:" + start.Format("20060102"))
		line("DTEND;VALUE=DATE:" + end.Format("20060102"))
	} else {
		tzid := loc.String()
		line("DTSTART;TZID=" + tzid + ":" + start.In(loc).Format(localStamp))
		line("DTEND;TZID=" + tzid + ":" + end.In(loc).Format(localStamp))
	}
	line("SUMMARY:" + escapeICS(e.Title))
	if e.Location != "" {
		line("LOCATION:" + escapeICS(e.Location))
	}
	if d := calendarDetails(e, pageURL); d != "" {
		line("DESCRIPTION:" + escapeICS(d))
	}
	if e.LinkURL != "" {
		line("URL:" + e.LinkURL)
	}
	line("END:VEVENT")
	line("END:VCALENDAR")
	return []byte(b.String())
}

var icsEscaper = strings.NewReplacer(`\`, `\\`, ";", `\;`, ",", `\,`, "\r\n", `\n`, "\n", `\n`, "\r", `\n`)

func escapeICS(s string) string { return icsEscaper.Replace(s) }

const localStamp = "20060102T150405"

// writeVTimeZone emits a VTIMEZONE block for loc covering the given years, so calendar
// apps can interpret TZID local times (including EST/EDT).
func writeVTimeZone(line func(string), loc *time.Location, years ...int) {
	if loc == nil {
		return
	}
	minY, maxY := years[0], years[0]
	for _, y := range years[1:] {
		if y < minY {
			minY = y
		}
		if y > maxY {
			maxY = y
		}
	}
	line("BEGIN:VTIMEZONE")
	line("TZID:" + loc.String())

	type trans struct {
		at      time.Time
		name    string
		off     int
		prevOff int
	}
	start := time.Date(minY-1, 1, 1, 0, 0, 0, 0, loc)
	end := time.Date(maxY+2, 1, 1, 0, 0, 0, 0, loc)
	prevName, prevOff := start.Zone()
	var changes []trans
	for t := start.Add(time.Hour); t.Before(end); t = t.Add(time.Hour) {
		name, off := t.Zone()
		if off != prevOff || name != prevName {
			changes = append(changes, trans{at: t, name: name, off: off, prevOff: prevOff})
			prevName, prevOff = name, off
		}
	}

	if len(changes) == 0 {
		name, off := start.Zone()
		line("BEGIN:STANDARD")
		line("TZOFFSETFROM:" + icsOffset(off))
		line("TZOFFSETTO:" + icsOffset(off))
		line("TZNAME:" + name)
		line("DTSTART:19700101T000000")
		line("END:STANDARD")
		line("END:VTIMEZONE")
		return
	}

	// Prefer a single STANDARD + DAYLIGHT pair with US DST RRULEs when they match.
	var standard, daylight *trans
	for i := range changes {
		c := &changes[i]
		if c.off > c.prevOff {
			if daylight == nil {
				daylight = c
			}
		} else if standard == nil {
			standard = c
		}
	}
	emit := func(kind string, c *trans) {
		toDaylight := c.off > c.prevOff
		line("BEGIN:" + kind)
		line("TZOFFSETFROM:" + icsOffset(c.prevOff))
		line("TZOFFSETTO:" + icsOffset(c.off))
		line("TZNAME:" + c.name)
		if rule := usDSTRule(c.at, toDaylight); rule != "" {
			// Conventional 2:00 wall-clock DTSTART used with the US DST RRULEs.
			if toDaylight {
				line("DTSTART:19700308T020000")
			} else {
				line("DTSTART:19701101T020000")
			}
			line("RRULE:" + rule)
		} else {
			line("DTSTART:" + c.at.Format(localStamp))
		}
		line("END:" + kind)
	}
	// RFC 5545: DAYLIGHT then STANDARD is the usual order for US zones.
	if daylight != nil {
		emit("DAYLIGHT", daylight)
	}
	if standard != nil {
		emit("STANDARD", standard)
	}
	line("END:VTIMEZONE")
}

// usDSTRule returns a YEARLY RRULE for US rules since 2007 (2nd Sunday in March / 1st Sunday
// in November at 2:00 local), or "" if the transition doesn't match.
func usDSTRule(at time.Time, toDaylight bool) string {
	// at is the first instant already in the new offset (3:00 AM in spring, 1:00 AM in fall).
	if at.Weekday() != time.Sunday {
		return ""
	}
	dom := at.Day()
	if toDaylight && at.Month() == time.March && dom >= 8 && dom <= 14 {
		return "FREQ=YEARLY;BYMONTH=3;BYDAY=2SU"
	}
	if !toDaylight && at.Month() == time.November && dom >= 1 && dom <= 7 {
		return "FREQ=YEARLY;BYMONTH=11;BYDAY=1SU"
	}
	return ""
}

func icsOffset(seconds int) string {
	sign := "+"
	if seconds < 0 {
		sign = "-"
		seconds = -seconds
	}
	return fmt.Sprintf("%s%02d%02d", sign, seconds/3600, (seconds%3600)/60)
}

// foldICS ends the content line with CRLF, folding at 75 octets without splitting a UTF-8
// character (RFC 5545 §3.1).
func foldICS(s string) string {
	var b strings.Builder
	n := 0
	for _, r := range s {
		w := utf8.RuneLen(r)
		if n+w > 75 {
			b.WriteString("\r\n ")
			n = 1
		}
		b.WriteRune(r)
		n += w
	}
	b.WriteString("\r\n")
	return b.String()
}
