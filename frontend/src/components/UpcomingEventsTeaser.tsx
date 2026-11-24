import { useEffect, useState } from 'react'
import { Link } from 'react-router-dom'
import { EventDateBadge } from './EventDateBadge'
import { loadEvents, relativeDay, type WardEvent } from '../lib/events'

const MAX = 3

/**
 * Small text-only "Upcoming Events" block for the home page. Renders nothing while loading,
 * on error, or when there are no upcoming events, so the home page never shows an empty box.
 */
export function UpcomingEventsTeaser() {
  const [events, setEvents] = useState<WardEvent[]>([])
  const [today, setToday] = useState('')

  useEffect(() => {
    let alive = true
    loadEvents('upcoming')
      .then((d) => {
        if (!alive) return
        setEvents((d.events ?? []).slice(0, MAX))
        setToday(d.today)
      })
      .catch(() => {})
    return () => {
      alive = false
    }
  }, [])

  if (events.length === 0) return null

  return (
    <section className="content-block home-events" aria-labelledby="sec-events">
      <div className="section-heading">
        <h2 id="sec-events">Upcoming Events</h2>
        <p className="section-sub">Baptisms, ward activities, temple trips, and service projects. Bring a friend!</p>
      </div>
      <ul className="home-events__list" role="list">
        {events.map((e) => {
          const rel = relativeDay(today, e.date)
          return (
            <li key={e.id}>
              <Link to={`/events/${e.id}`} className="home-events__item">
                <EventDateBadge date={e.date} />
                <span className="home-events__text">
                  <span className="home-events__title">
                    {rel ? <span className="home-events__soon">{rel}</span> : null}
                    {e.title}
                  </span>
                  <span className="home-events__meta">
                    {e.timeLabel}
                    {e.location ? ` · ${e.location}` : ''}
                  </span>
                </span>
              </Link>
            </li>
          )
        })}
      </ul>
      <p className="cta-row">
        <Link className="btn btn-secondary" to="/events">
          See All Events
        </Link>
      </p>
    </section>
  )
}
