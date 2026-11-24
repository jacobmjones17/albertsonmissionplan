import { useEffect, useMemo, useState } from 'react'
import { useLocation } from 'react-router-dom'
import { AddToCalendar } from '../components/AddToCalendar'
import { ImageLightbox } from '../components/ImageLightbox'
import { EventDateBadge } from '../components/EventDateBadge'
import { PageHero } from '../components/PageHero'
import {
  WEEKDAY_LABELS,
  addMonths,
  calendarWeeks,
  groupByMonth,
  loadEvents,
  monthTitle,
  relativeDay,
  shortTime,
  zoneDisplayName,
  type WardEvent,
} from '../lib/events'

type View = 'list' | 'calendar'

function PinIcon() {
  return (
    <svg viewBox="0 0 24 24" width="16" height="16" aria-hidden="true" focusable="false">
      <path
        fill="currentColor"
        d="M12 2a7 7 0 0 0-7 7c0 5.25 7 13 7 13s7-7.75 7-13a7 7 0 0 0-7-7Zm0 9.5A2.5 2.5 0 1 1 12 6.5a2.5 2.5 0 0 1 0 5Z"
      />
    </svg>
  )
}

function ClockIcon() {
  return (
    <svg viewBox="0 0 24 24" width="16" height="16" aria-hidden="true" focusable="false">
      <circle cx="12" cy="12" r="8.5" fill="none" stroke="currentColor" strokeWidth="1.9" />
      <path d="M12 7.5V12l3 2" fill="none" stroke="currentColor" strokeWidth="1.9" strokeLinecap="round" />
    </svg>
  )
}

function EventCard({ event, today, past = false }: { event: WardEvent; today: string; past?: boolean }) {
  const rel = past ? '' : relativeDay(today, event.date)
  const titleId = `event-${event.id}-title`
  const [flyerOpen, setFlyerOpen] = useState(false)
  const flyerSrc = event.hasFlyer ? event.flyerUrl || event.flyerThumbUrl : undefined
  return (
    <article className={`event-card${past ? ' event-card--past' : ''}`} id={`event-${event.id}`} aria-labelledby={titleId}>
      <EventDateBadge date={event.date} muted={past} />
      <div className="event-card__body">
        {rel ? <p className="event-card__soon">{rel}</p> : null}
        <h3 id={titleId} className="event-card__title">
          {event.title}
        </h3>
        <p className="event-card__meta">
          <span className="event-card__when">
            <ClockIcon />
            <span className="event-card__when-text">
              <span>{event.dateLabel}</span>
              <span className="event-card__sep" aria-hidden="true">
                {' · '}
              </span>
              <span className="event-card__time">{event.timeLabel}</span>
            </span>
          </span>
          {event.location ? (
            <span className="event-card__where">
              <PinIcon />
              <span>
                {event.location}
                {event.mapUrl ? (
                  <>
                    <span aria-hidden="true"> · </span>
                    <a href={event.mapUrl} target="_blank" rel="noopener noreferrer" className="event-card__map">
                      Map<span className="sr-only"> for {event.location} (opens in a new tab)</span>
                    </a>
                  </>
                ) : null}
              </span>
            </span>
          ) : null}
        </p>
        {event.description ? <p className="event-card__desc">{event.description}</p> : null}
        {flyerSrc ? (
          <>
            <button
              type="button"
              className="event-card__flyer"
              onClick={() => setFlyerOpen(true)}
              aria-label={`View flyer for ${event.title}`}
            >
              <img
                src={event.flyerThumbUrl || flyerSrc}
                alt=""
                width={event.flyerWidth || undefined}
                height={event.flyerHeight || undefined}
              />
              <span className="event-card__flyer-label">View flyer</span>
            </button>
            <ImageLightbox
              open={flyerOpen}
              src={flyerSrc}
              thumbSrc={event.flyerThumbUrl}
              alt={`Flyer for ${event.title}`}
              width={event.flyerWidth}
              height={event.flyerHeight}
              title={event.title}
              onClose={() => setFlyerOpen(false)}
            />
          </>
        ) : null}
        {!past || event.link ? (
          <div className="event-card__actions">
            {!past ? <AddToCalendar event={event} /> : null}
            {event.link ? (
              <a className="event-card__link" href={event.link} target="_blank" rel="noopener noreferrer">
                More Info<span className="sr-only"> about {event.title} (opens in a new tab)</span>
              </a>
            ) : null}
          </div>
        ) : null}
      </div>
    </article>
  )
}

function MonthCalendar({
  month,
  setMonth,
  events,
  today,
  minMonth,
  maxMonth,
}: {
  month: string
  setMonth: (m: string) => void
  events: WardEvent[]
  today: string
  minMonth: string
  maxMonth: string
}) {
  const byDate = useMemo(() => {
    const map = new Map<string, WardEvent[]>()
    for (const e of events) {
      const list = map.get(e.date) ?? []
      list.push(e)
      map.set(e.date, list)
    }
    return map
  }, [events])
  const weeks = calendarWeeks(month)
  const todayMonth = today.slice(0, 7)

  function jump(e: WardEvent) {
    const el = document.getElementById(`event-${e.id}`)
    if (!el) return
    el.scrollIntoView({ behavior: 'smooth', block: 'center' })
    el.classList.remove('is-flash')
    void el.offsetWidth
    el.classList.add('is-flash')
  }

  return (
    <div className="event-cal">
      <div className="event-cal__bar">
        <button
          type="button"
          className="event-cal__nav"
          onClick={() => setMonth(addMonths(month, -1))}
          disabled={month <= minMonth}
          aria-label={`Previous month, ${monthTitle(addMonths(month, -1))}`}
        >
          ‹
        </button>
        <h2 className="event-cal__title" aria-live="polite">
          {monthTitle(month)}
        </h2>
        <button
          type="button"
          className="event-cal__nav"
          onClick={() => setMonth(addMonths(month, 1))}
          disabled={month >= maxMonth}
          aria-label={`Next month, ${monthTitle(addMonths(month, 1))}`}
        >
          ›
        </button>
        {month !== todayMonth ? (
          <button type="button" className="event-cal__today" onClick={() => setMonth(todayMonth)}>
            This Month
          </button>
        ) : null}
      </div>
      <table className="event-cal__grid">
        <thead>
          <tr>
            {WEEKDAY_LABELS.map((d) => (
              <th key={d} scope="col">
                {d}
              </th>
            ))}
          </tr>
        </thead>
        <tbody>
          {weeks.map((week) => (
            <tr key={week[0].date}>
              {week.map((c) => {
                const list = c.inMonth ? byDate.get(c.date) ?? [] : []
                const cls = [
                  'event-cal__cell',
                  c.inMonth ? '' : 'is-out',
                  c.date === today ? 'is-today' : '',
                  c.date < today ? 'is-past' : '',
                  list.length ? 'has-events' : '',
                ]
                  .filter(Boolean)
                  .join(' ')
                return (
                  <td key={c.date} className={cls}>
                    <span className="event-cal__day">
                      {c.day}
                      {c.date === today ? <span className="sr-only"> (today)</span> : null}
                    </span>
                    {list.map((e) => (
                      <button
                        key={e.id}
                        type="button"
                        className="event-cal__chip"
                        onClick={() => jump(e)}
                        title={`${e.title} · ${e.timeLabel}`}
                      >
                        <span className="event-cal__chip-time">{e.allDay ? 'All day' : shortTime(e.startTime)}</span>
                        <span className="event-cal__chip-title">{e.title}</span>
                      </button>
                    ))}
                  </td>
                )
              })}
            </tr>
          ))}
        </tbody>
      </table>
    </div>
  )
}

export function Events() {
  const [upcoming, setUpcoming] = useState<WardEvent[] | null>(null)
  const [past, setPast] = useState<WardEvent[] | null>(null)
  const [today, setToday] = useState('')
  const [timeZone, setTimeZone] = useState('America/Denver')
  const [err, setErr] = useState<string | null>(null)
  const [showPast, setShowPast] = useState(false)
  const [view, setView] = useState<View>('list')
  const [month, setMonth] = useState('')

  useEffect(() => {
    loadEvents('upcoming')
      .then((d) => {
        setUpcoming(d.events ?? [])
        setToday(d.today)
        setTimeZone(d.timeZone || 'America/Denver')
        setMonth(d.events?.[0]?.monthKey ?? d.today.slice(0, 7))
      })
      .catch((e) => {
        setErr(e instanceof Error ? e.message : 'Could not load events.')
        setUpcoming([])
      })
  }, [])

  // Past events load the first time someone asks for them (or opens the calendar).
  const needPast = showPast || view === 'calendar'
  useEffect(() => {
    if (!needPast || past !== null) return
    loadEvents('past')
      .then((d) => setPast(d.events ?? []))
      .catch(() => setPast([]))
  }, [needPast, past])

  // Links like /events#event-12 (from the home page) scroll to that event once it loads.
  const { hash } = useLocation()
  useEffect(() => {
    if (!upcoming || !hash.startsWith('#event-')) return
    const el = document.getElementById(hash.slice(1))
    if (!el) return
    window.requestAnimationFrame(() => {
      el.scrollIntoView({ block: 'center' })
      el.classList.add('is-flash')
    })
  }, [upcoming, hash])

  const groups = useMemo(() => groupByMonth(upcoming ?? []), [upcoming])
  const zoneName = useMemo(() => zoneDisplayName(timeZone), [timeZone])

  const calendarEvents = useMemo(() => [...(past ?? []), ...(upcoming ?? [])], [past, upcoming])
  const monthEvents = useMemo(
    () =>
      calendarEvents
        .filter((e) => e.monthKey === month)
        .sort((a, b) => (a.date + (a.allDay ? '' : a.startTime)).localeCompare(b.date + (b.allDay ? '' : b.startTime))),
    [calendarEvents, month],
  )
  const todayMonth = today.slice(0, 7)
  const lastUpcoming = upcoming?.[upcoming.length - 1]?.monthKey ?? todayMonth
  const maxMonth = lastUpcoming > addMonths(todayMonth, 12) ? lastUpcoming : addMonths(todayMonth, 12)
  const minMonth = addMonths(todayMonth, -12)

  return (
    <main id="main">
      <PageHero
        titleId="events-title"
        eyebrow="Come and see"
        title="Ward Events"
        lede="Baptisms, ward activities, temple trips, and service projects. Everyone is welcome."
      />
      <div className="home-mission-strip home-mission-strip--tight">
        <p>
          Invite a friend! These gatherings are a great way to share the gospel and to get to know
          your ward family. Tap <strong>Add to Calendar</strong> on any event to save it to your phone.
        </p>
      </div>

      <div className="wrap events-page">
        {err ? <p className="panel-sites panel-sites--warn">{err}</p> : null}

        {upcoming === null ? (
          <p className="loading-note" role="status">
            <span className="loading-note__spinner" aria-hidden="true" />
            Loading events…
          </p>
        ) : (
          <>
            <div className="events-toolbar">
              <p className="events-toolbar__zone">All times are {zoneName}.</p>
              <div className="events-toolbar__controls">
                <div className="events-view-toggle" role="group" aria-label="Show events as">
                  <button type="button" aria-pressed={view === 'list'} onClick={() => setView('list')}>
                    List
                  </button>
                  <button type="button" aria-pressed={view === 'calendar'} onClick={() => setView('calendar')}>
                    Calendar
                  </button>
                </div>
              </div>
            </div>

            {view === 'calendar' ? (
              <section className="events-calendar-view" aria-label="Month calendar">
                <MonthCalendar
                  month={month || todayMonth}
                  setMonth={setMonth}
                  events={calendarEvents}
                  today={today}
                  minMonth={minMonth}
                  maxMonth={maxMonth}
                />
                <h2 className="events-month__title events-month__title--sub">
                  {monthEvents.length ? `Events in ${monthTitle(month)}` : `No events in ${monthTitle(month)}`}
                </h2>
                <div className="events-month__list">
                  {monthEvents.map((e) => (
                    <EventCard key={e.id} event={e} today={today} past={!e.upcoming} />
                  ))}
                </div>
              </section>
            ) : null}

            <div className={view === 'calendar' ? 'events-list-view events-list-view--phone-only' : 'events-list-view'}>
              {groups.length === 0 && !err ? (
                <section className="gallery-empty events-empty" aria-labelledby="events-empty-title">
                  <div className="gallery-empty__art" aria-hidden="true">
                    <svg viewBox="0 0 64 64" width="64" height="64">
                      <rect x="8" y="12" width="48" height="44" rx="6" fill="none" stroke="currentColor" strokeWidth="3" />
                      <path d="M8 24h48M20 6v12M44 6v12" fill="none" stroke="currentColor" strokeWidth="3" strokeLinecap="round" />
                      <path d="M32 31l2.6 5.3 5.9.9-4.3 4.1 1 5.8-5.2-2.7-5.2 2.7 1-5.8-4.3-4.1 5.9-.9z" fill="#e0b868" />
                    </svg>
                  </div>
                  <h2 id="events-empty-title">No upcoming events right now</h2>
                  <p>
                    Check back soon. Baptisms, ward activities, temple trips, and service projects will show
                    up here as they&apos;re planned.
                  </p>
                </section>
              ) : (
                groups.map((g) => (
                  <section key={g.key} className="events-month" aria-labelledby={`month-${g.key}`}>
                    <h2 id={`month-${g.key}`} className="events-month__title">
                      {g.label}
                    </h2>
                    <div className="events-month__list">
                      {g.events.map((e) => (
                        <EventCard key={e.id} event={e} today={today} />
                      ))}
                    </div>
                  </section>
                ))
              )}

              <div className="events-past">
                <button
                  type="button"
                  className="btn-secondary events-past__toggle"
                  aria-expanded={showPast}
                  aria-controls="events-past-list"
                  onClick={() => setShowPast((v) => !v)}
                >
                  {showPast ? 'Hide Past Events' : 'Show Past Events'}
                </button>
                {showPast ? (
                  <section id="events-past-list" className="events-month events-month--past" aria-labelledby="events-past-title">
                    <h2 id="events-past-title" className="events-month__title">
                      Past Events
                    </h2>
                    {past === null ? (
                      <p className="loading-note" role="status">
                        <span className="loading-note__spinner" aria-hidden="true" />
                        Loading…
                      </p>
                    ) : past.length === 0 ? (
                      <p className="home-empty-note">No past events from the last year.</p>
                    ) : (
                      <div className="events-month__list">
                        {past.map((e) => (
                          <EventCard key={e.id} event={e} today={today} past />
                        ))}
                      </div>
                    )}
                  </section>
                ) : null}
              </div>
            </div>
          </>
        )}
      </div>
    </main>
  )
}
