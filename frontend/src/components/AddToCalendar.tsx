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

/**
 * "Add to Calendar" menu: Apple Calendar and Outlook take the .ics file (iPhone opens it
 * straight into Calendar); Google Calendar opens a pre-filled event.
 */
export function AddToCalendar({ event }: { event: WardEvent }) {
  const ref = useRef<HTMLDetailsElement>(null)

  // Close when tapping elsewhere or pressing Escape.
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

  return (
    <details className="add-cal" ref={ref}>
      <summary className="add-cal__summary">
        <CalendarIcon />
        Add to Calendar
      </summary>
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
    </details>
  )
}
