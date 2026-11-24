import type { WardEvent } from '../lib/events'
import { eventShareText, eventShareUrl } from '../lib/events'

function ShareIcon() {
  return (
    <svg viewBox="0 0 24 24" width="18" height="18" aria-hidden="true" focusable="false">
      <path
        d="M12 3v11"
        fill="none"
        stroke="currentColor"
        strokeWidth="1.8"
        strokeLinecap="round"
      />
      <path
        d="M8.5 6.5 12 3l3.5 3.5"
        fill="none"
        stroke="currentColor"
        strokeWidth="1.8"
        strokeLinecap="round"
        strokeLinejoin="round"
      />
      <path
        d="M6 11.5v7A1.5 1.5 0 0 0 7.5 20h9a1.5 1.5 0 0 0 1.5-1.5v-7"
        fill="none"
        stroke="currentColor"
        strokeWidth="1.8"
        strokeLinecap="round"
      />
    </svg>
  )
}

function isAbort(e: unknown): boolean {
  return e instanceof DOMException ? e.name === 'AbortError' : e instanceof Error && e.name === 'AbortError'
}

async function copyText(text: string): Promise<boolean> {
  try {
    await navigator.clipboard.writeText(text)
    return true
  } catch {
    try {
      const ta = document.createElement('textarea')
      ta.value = text
      ta.setAttribute('readonly', '')
      ta.style.position = 'fixed'
      ta.style.left = '-9999px'
      document.body.appendChild(ta)
      ta.select()
      const ok = document.execCommand('copy')
      ta.remove()
      return ok
    } catch {
      return false
    }
  }
}

/**
 * Opens the device share sheet (Web Share API) with the event title, a short
 * date/time/location blurb, and a link to /events/{id}. The flyer is not attached
 * as a file — iMessage and other apps show it once via the link's Open Graph preview.
 * When navigator.share isn't available, the link is copied and onCopied is called.
 */
export function ShareEvent({ event, onCopied }: { event: WardEvent; onCopied: () => void }) {
  async function share() {
    const url = eventShareUrl(event.id)
    const data: ShareData = { title: event.title, text: eventShareText(event), url }
    if (typeof navigator.share === 'function') {
      try {
        await navigator.share(data)
        return
      } catch (e) {
        if (isAbort(e)) return
      }
    }
    if (await copyText(url)) onCopied()
  }

  return (
    <button type="button" className="event-share" onClick={() => void share()}>
      <ShareIcon />
      Share
    </button>
  )
}
