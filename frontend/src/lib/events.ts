import { apiJson } from '../api'

/** One event as returned by GET /api/events and /api/admin/events. */
export type WardEvent = {
  id: number
  title: string
  /** Ward-local date, "2026-12-05". */
  date: string
  allDay: boolean
  /** Ward-local "18:30" (empty when all day). */
  startTime: string
  endTime: string
  location: string
  description: string
  link: string
  /** True when a flyer photo has been uploaded (takes display priority over link). */
  hasFlyer?: boolean
  flyerUrl?: string
  flyerThumbUrl?: string
  flyerWidth?: number
  flyerHeight?: number
  /** "Saturday, December 5, 2026" */
  dateLabel: string
  /** "6:30 – 8:00 PM" or "All day" */
  timeLabel: string
  /** "MST" / "MDT" */
  zone: string
  /** "2026-12" */
  monthKey: string
  /** "December 2026" */
  monthLabel: string
  upcoming: boolean
  icsUrl: string
  googleCalendarUrl: string
  mapUrl: string
}

export type EventsResponse = {
  events: WardEvent[] | null
  /** Today's date in the ward's time zone. */
  today: string
  /** IANA zone, e.g. "America/Denver". */
  timeZone: string
}

export async function loadEvents(view: 'upcoming' | 'past' = 'upcoming'): Promise<EventsResponse> {
  const d = await apiJson<EventsResponse>(view === 'past' ? '/api/events?view=past' : '/api/events')
  return { ...d, events: Array.isArray(d.events) ? d.events : [] }
}

/** "Mountain Time" for America/Denver (falls back to the zone name). */
export function zoneDisplayName(timeZone: string): string {
  try {
    const part = new Intl.DateTimeFormat('en-US', { timeZone, timeZoneName: 'longGeneric' })
      .formatToParts(new Date())
      .find((p) => p.type === 'timeZoneName')
    if (part?.value) return part.value
  } catch {
    // older browsers: fall through
  }
  return timeZone.replace(/_/g, ' ')
}

const MONTHS_SHORT = ['Jan', 'Feb', 'Mar', 'Apr', 'May', 'Jun', 'Jul', 'Aug', 'Sep', 'Oct', 'Nov', 'Dec']
const DAYS_SHORT = ['Sun', 'Mon', 'Tue', 'Wed', 'Thu', 'Fri', 'Sat']

/** Parse "2026-12-05" as a calendar date (no time-zone shifting). */
export function parseDate(date: string): { y: number; m: number; d: number } {
  const [y, m, d] = date.split('-').map((n) => parseInt(n, 10))
  return { y, m, d }
}

export function dateParts(date: string): { month: string; day: number; weekday: string } {
  const { y, m, d } = parseDate(date)
  const wd = new Date(Date.UTC(y, m - 1, d)).getUTCDay()
  return { month: MONTHS_SHORT[m - 1] ?? '', day: d, weekday: DAYS_SHORT[wd] ?? '' }
}

/** Whole days from `today` to `date` (both "YYYY-MM-DD"). */
export function daysBetween(today: string, date: string): number {
  const a = parseDate(today)
  const b = parseDate(date)
  return Math.round((Date.UTC(b.y, b.m - 1, b.d) - Date.UTC(a.y, a.m - 1, a.d)) / 86_400_000)
}

/** "Today" / "Tomorrow" / "" for an upcoming date. */
export function relativeDay(today: string, date: string): string {
  const n = daysBetween(today, date)
  if (n === 0) return 'Today'
  if (n === 1) return 'Tomorrow'
  return ''
}

export type MonthGroup = { key: string; label: string; events: WardEvent[] }

export function groupByMonth(list: WardEvent[]): MonthGroup[] {
  const out: MonthGroup[] = []
  for (const e of list) {
    const last = out[out.length - 1]
    if (last && last.key === e.monthKey) last.events.push(e)
    else out.push({ key: e.monthKey, label: e.monthLabel, events: [e] })
  }
  return out
}

/** "2026-12" → { y: 2026, m: 12 } and back, with month arithmetic. */
export function addMonths(key: string, delta: number): string {
  const [y, m] = key.split('-').map((n) => parseInt(n, 10))
  const idx = y * 12 + (m - 1) + delta
  const ny = Math.floor(idx / 12)
  const nm = (idx % 12) + 1
  return `${ny}-${String(nm).padStart(2, '0')}`
}

export function monthTitle(key: string): string {
  const [y, m] = key.split('-').map((n) => parseInt(n, 10))
  return new Date(Date.UTC(y, m - 1, 1)).toLocaleDateString('en-US', { month: 'long', year: 'numeric', timeZone: 'UTC' })
}

export type CalendarCell = { date: string; day: number; inMonth: boolean }

/** Six-or-fewer Sunday-first weeks covering the month "YYYY-MM". */
export function calendarWeeks(key: string): CalendarCell[][] {
  const [y, m] = key.split('-').map((n) => parseInt(n, 10))
  const first = new Date(Date.UTC(y, m - 1, 1))
  const start = new Date(first)
  start.setUTCDate(1 - first.getUTCDay())
  const weeks: CalendarCell[][] = []
  const cur = new Date(start)
  for (let w = 0; w < 6; w++) {
    const week: CalendarCell[] = []
    for (let i = 0; i < 7; i++) {
      week.push({
        date: cur.toISOString().slice(0, 10),
        day: cur.getUTCDate(),
        inMonth: cur.getUTCMonth() === m - 1,
      })
      cur.setUTCDate(cur.getUTCDate() + 1)
    }
    if (w > 0 && !week.some((c) => c.inMonth)) break
    weeks.push(week)
  }
  return weeks
}

export const WEEKDAY_LABELS = DAYS_SHORT

/** Compact clock for calendar chips: "18:30" → "6:30 PM", "10:00" → "10 AM". */
export function shortTime(hhmm: string): string {
  const [h, m] = hhmm.split(':').map((n) => parseInt(n, 10))
  if (Number.isNaN(h)) return ''
  const suffix = h >= 12 ? 'PM' : 'AM'
  const h12 = h % 12 === 0 ? 12 : h % 12
  return m ? `${h12}:${String(m).padStart(2, '0')} ${suffix}` : `${h12} ${suffix}`
}
