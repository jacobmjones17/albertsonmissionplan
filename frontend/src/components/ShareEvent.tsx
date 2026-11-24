import { useEffect, useState } from 'react'
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

function canShareFiles(file: File): boolean {
  try {
    return typeof navigator.canShare === 'function' && navigator.canShare({ files: [file] })
  } catch {
    return false
  }
}

function flyerFileName(title: string): string {
  const slug = title
    .toLowerCase()
    .replace(/[^a-z0-9]+/g, '-')
    .replace(/^-+|-+$/g, '')
    .slice(0, 60)
  return `${slug || 'event-flyer'}.jpg`
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
 * Opens the device share sheet (Web Share API) with the event title, a short date/time/location
 * blurb, and a link to /events/{id}. If a flyer was uploaded and the device can attach files,
 * the flyer image is included — it is prefetched so iOS doesn't drop the user-activation
 * gesture. When navigator.share isn't available, the link is copied and onCopied is called.
 */
export function ShareEvent({ event, onCopied }: { event: WardEvent; onCopied: () => void }) {
  const [flyerFile, setFlyerFile] = useState<File | null>(null)

  useEffect(() => {
    if (!event.hasFlyer || !event.flyerUrl) {
      setFlyerFile(null)
      return
    }
    const url = event.flyerUrl
    const ac = new AbortController()
    let cancelled = false
    fetch(url, { signal: ac.signal })
      .then((r) => {
        if (!r.ok) throw new Error('flyer')
        return r.blob()
      })
      .then((blob) => {
        if (cancelled) return
        setFlyerFile(new File([blob], flyerFileName(event.title), { type: blob.type || 'image/jpeg' }))
      })
      .catch(() => {
        if (!cancelled) setFlyerFile(null)
      })
    return () => {
      cancelled = true
      ac.abort()
    }
  }, [event.hasFlyer, event.flyerUrl, event.title])

  async function share() {
    const url = eventShareUrl(event.id)
    const data: ShareData = { title: event.title, text: eventShareText(event), url }

    // Only attach a prefetched flyer — never await a fetch here, or iOS rejects the share.
    if (flyerFile && canShareFiles(flyerFile) && typeof navigator.share === 'function') {
      try {
        await navigator.share({ ...data, files: [flyerFile] })
        return
      } catch (e) {
        if (isAbort(e)) return
      }
    }
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
