import { useEffect, useRef } from 'react'
import type { WardEvent } from '../lib/events'

function CalendarIcon() {
  return (
    <svg viewBox="0 0 24 24" width="18" height="18" aria-hidden="true" focusable="false">
      <rect x="3.5" y="5" width="17" height="15.5" rx="2.5" fill="none" stroke="currentColor" strokeWidth="1.8" />
      <path d="M3.5 9.5h17M8 3v4M16 3v4" fill="none" stroke="currentColor" strokeWidth="1.8" strokeLinecap="round" />
      <path d="M12 12.5v5M9.5 15h5" fill="none" stroke="currentColor" strokeWidth="1.8" strokeLinecap="round" />
    </svg>
  )
}

/** Preferred calendar target for this device: .ics on Apple, Google on Android, a menu otherwise. */
export function calendarTarget(ua = typeof navigator !== 'undefined' ? navigator.userAgent : '', touchPoints = typeof navigator !== 'undefined' ? navigator.maxTouchPoints : 0): 'apple' | 'google' | 'menu' {
  // iPadOS 13+ may report as Macintosh with a touch screen.
  if (/iPhone|iPad|iPod/i.test(ua) || (/Macintosh/i.test(ua) && touchPoints > 1)) return 'apple'
  if (/Mac OS X|Macintosh/i.test(ua)) return 'apple'
  if (/Android/i.test(ua)) return 'google'
  return 'menu'
}

/**
 * "Add to Calendar": on iPhone/iPad/Mac this downloads the .ics (Apple Calendar);
 * on Android it opens a pre-filled Google Calendar event. Other devices keep a
 * menu of Apple, Google, and Outlook. A small "Other calendar options" link
 * still lets phone users pick a different calendar.
 */
export function AddToCalendar({ event }: { event: WardEvent }) {
  const ref = useRef<HTMLDetailsElement>(null)
  const target = calendarTarget()

  useEffect(() => {
    function onDown(e: PointerEvent) {
      const el = ref.current
      if (el?.open && e.target instanceof Node && !el.contains(e.target)) el.open = false
    }
    function onKey(e: KeyboardEvent) {
      const el = ref.current
      if (e.key === 'Escape' && el?.open) {
        el.open = false
        el.querySelector('summary')?.focus()
      }
    }
    document.addEventListener('pointerdown', onDown)
    document.addEventListener('keydown', onKey)
    return () => {
      document.removeEventListener('pointerdown', onDown)
      document.removeEventListener('keydown', onKey)
    }
  }, [])

  const close = () => {
    if (ref.current) ref.current.open = false
  }

  const menu = (
    <ul className="add-cal__menu" role="list">
      <li>
        <a href={event.icsUrl} download onClick={close}>
          Apple Calendar <span>iPhone, iPad, Mac</span>
        </a>
      </li>
      <li>
        <a href={event.googleCalendarUrl} target="_blank" rel="noopener noreferrer" onClick={close}>
          Google Calendar <span>Android, Gmail</span>
        </a>
      </li>
      <li>
        <a href={event.icsUrl} download onClick={close}>
          Outlook or other <span>Downloads a calendar file</span>
        </a>
      </li>
    </ul>
  )

  if (target === 'menu') {
    return (
      <details className="add-cal" ref={ref}>
        <summary className="add-cal__summary">
          <CalendarIcon />
          Add to Calendar
        </summary>
        {menu}
      </details>
    )
  }

  const href = target === 'google' ? event.googleCalendarUrl : event.icsUrl
  const extra = target === 'google' ? { target: '_blank' as const, rel: 'noopener noreferrer' } : { download: true }

  return (
    <div className="add-cal add-cal--direct">
      <a className="add-cal__summary add-cal__summary--direct" href={href} {...extra}>
        <CalendarIcon />
        Add to Calendar
      </a>
      <details className="add-cal__more" ref={ref}>
        <summary className="add-cal__more-toggle">Other calendar options</summary>
        {menu}
      </details>
    </div>
  )
}
