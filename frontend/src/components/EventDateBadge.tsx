import { dateParts } from '../lib/events'

/** Calendar-page style date block: "DEC / 5 / Sat". */
export function EventDateBadge({ date, muted = false }: { date: string; muted?: boolean }) {
  const p = dateParts(date)
  return (
    <div className={`event-date${muted ? ' event-date--muted' : ''}`} aria-hidden="true">
      <span className="event-date__month">{p.month}</span>
      <span className="event-date__day">{p.day}</span>
      <span className="event-date__weekday">{p.weekday}</span>
    </div>
  )
}
