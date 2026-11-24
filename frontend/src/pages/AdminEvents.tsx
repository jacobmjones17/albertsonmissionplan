import { useCallback, useEffect, useRef, useState, type FormEvent } from 'react'
import { Link } from 'react-router-dom'
import { apiJson, apiUpload } from '../api'
import { EventDateBadge } from '../components/EventDateBadge'
import { PageHero } from '../components/PageHero'
import { Toast, type ToastPayload } from '../components/Toast'
import { zoneDisplayName, type EventsResponse, type WardEvent } from '../lib/events'
import { useUnsavedChangesWarning } from '../lib/useUnsavedChangesWarning'

type Draft = {
  title: string
  date: string
  allDay: boolean
  startTime: string
  endTime: string
  location: string
  description: string
  link: string
}

const EMPTY: Draft = { title: '', date: '', allDay: false, startTime: '', endTime: '', location: '', description: '', link: '' }

const MAX_DESCRIPTION = 2000
const MAX_FLYER_BYTES = 15 * 1024 * 1024

function precheckFlyer(f: File): string | undefined {
  if (/image\/hei[cf]/i.test(f.type) || /\.(heic|heif)$/i.test(f.name)) {
    return 'HEIC photos can’t be read yet. On iPhone choose the photo from the Photos picker (it converts to JPEG) or set Camera → Formats → Most Compatible.'
  }
  if (f.type && !/^image\/(jpeg|png|webp)$/i.test(f.type)) return 'Please use a JPEG, PNG, or WebP picture.'
  if (f.size > MAX_FLYER_BYTES) return 'This picture is larger than 15 MB. Try a smaller photo.'
  return undefined
}


function draftFrom(e: WardEvent): Draft {
  return {
    title: e.title,
    date: e.date,
    allDay: e.allDay,
    startTime: e.startTime,
    endTime: e.endTime,
    location: e.location,
    description: e.description,
    link: e.link,
  }
}

function sameDraft(a: Draft, b: Draft): boolean {
  return (Object.keys(a) as (keyof Draft)[]).every((k) => a[k] === b[k])
}

/** Same checks as the server, so most mistakes are caught before saving. */
function checkDraft(d: Draft): { field: keyof Draft; message: string } | null {
  if (!d.title.trim()) return { field: 'title', message: 'Give the event a title.' }
  if (!d.date) return { field: 'date', message: 'Choose the date of the event.' }
  if (!d.allDay && !d.startTime) return { field: 'startTime', message: 'Add a start time, or check “All day”.' }
  if (!d.allDay && d.endTime && d.endTime <= d.startTime)
    return { field: 'endTime', message: 'The end time needs to be after the start time.' }
  return null
}

function EventForm({
  initial,
  editing,
  zoneName,
  onSaved,
  onCancel,
  onDirtyChange,
}: {
  initial: Draft
  editing: WardEvent | null
  zoneName: string
  onSaved: (e: WardEvent, created: boolean, warning?: string) => void
  onCancel: () => void
  onDirtyChange: (dirty: boolean) => void
}) {
  const [d, setD] = useState<Draft>(initial)
  const [busy, setBusy] = useState(false)
  const [error, setError] = useState<{ field?: keyof Draft; message: string } | null>(null)
  const [pendingFile, setPendingFile] = useState<File | null>(null)
  const [pendingUrl, setPendingUrl] = useState<string | null>(null)
  const [removeFlyer, setRemoveFlyer] = useState(false)
  const formRef = useRef<HTMLFormElement>(null)
  const fileRef = useRef<HTMLInputElement>(null)
  const pendingUrlRef = useRef<string | null>(null)
  const idp = editing ? `ev-${editing.id}` : 'ev-new'

  const dirty = !sameDraft(d, initial) || pendingFile !== null || removeFlyer
  useEffect(() => onDirtyChange(dirty), [dirty, onDirtyChange])
  useEffect(() => () => onDirtyChange(false), [onDirtyChange])
  useEffect(() => () => {
    if (pendingUrlRef.current) URL.revokeObjectURL(pendingUrlRef.current)
  }, [])

  useEffect(() => {
    formRef.current?.querySelector<HTMLInputElement>('input')?.focus()
  }, [])

  function set<K extends keyof Draft>(k: K, v: Draft[K]) {
    setD((prev) => ({ ...prev, [k]: v }))
    if (error?.field === k) setError(null)
  }

  function focusField(field?: keyof Draft) {
    if (!field) return
    window.requestAnimationFrame(() => document.getElementById(`${idp}-${field}`)?.focus())
  }

  function setPendingPreview(url: string | null) {
    if (pendingUrlRef.current) URL.revokeObjectURL(pendingUrlRef.current)
    pendingUrlRef.current = url
    setPendingUrl(url)
  }

  function chooseFlyer(f?: File) {
    if (!f) return
    const problem = precheckFlyer(f)
    if (problem) {
      setError({ message: problem })
      return
    }
    setPendingPreview(URL.createObjectURL(f))
    setPendingFile(f)
    setRemoveFlyer(false)
    setError(null)
  }

  function clearFlyer() {
    setPendingPreview(null)
    setPendingFile(null)
    if (editing?.hasFlyer) setRemoveFlyer(true)
    setError(null)
  }

  async function submit(ev: FormEvent) {
    ev.preventDefault()
    const bad = checkDraft(d)
    if (bad) {
      setError(bad)
      focusField(bad.field)
      return
    }
    setBusy(true)
    setError(null)
    try {
      const payload = { ...d, startTime: d.allDay ? '' : d.startTime, endTime: d.allDay ? '' : d.endTime }
      const res = await apiJson<{ ok: boolean; event?: WardEvent }>(
        editing ? `/api/admin/events/${editing.id}` : '/api/admin/events',
        { method: 'POST', json: payload },
      )
      let saved = res.event
      const id = saved?.id ?? editing?.id
      let warning: string | undefined
      if (pendingFile && id) {
        try {
          const form = new FormData()
          form.append('image', pendingFile)
          const up = await apiUpload<{ event?: WardEvent }>(`/api/admin/events/${id}/flyer`, form)
          if (up.event) saved = up.event
        } catch (fe) {
          warning = fe instanceof Error ? fe.message : 'Could not upload the flyer.'
        }
      } else if (removeFlyer && id) {
        try {
          const del = await apiJson<{ event?: WardEvent }>(`/api/admin/events/${id}/flyer/delete`, {
            method: 'POST',
            json: {},
          })
          if (del.event) saved = del.event
        } catch (fe) {
          warning = fe instanceof Error ? fe.message : 'Could not remove the flyer.'
        }
      }
      onDirtyChange(false)
      if (saved) onSaved(saved, !editing, warning)
    } catch (e) {
      setError({ message: e instanceof Error ? e.message : 'Could not save the event.' })
    } finally {
      setBusy(false)
    }
  }

  const flyerPreview = pendingUrl ?? (!removeFlyer && editing?.hasFlyer ? editing.flyerThumbUrl || editing.flyerUrl : undefined)

  const invalid = (f: keyof Draft) => (error?.field === f ? true : undefined)
  const errId = `${idp}-error`

  return (
    <form ref={formRef} className="event-form form-stack" onSubmit={(e) => void submit(e)} noValidate>
      <label>
        Event title
        <input
          id={`${idp}-title`}
          type="text"
          value={d.title}
          maxLength={120}
          placeholder="Ward Christmas Party"
          onChange={(e) => set('title', e.target.value)}
          aria-invalid={invalid('title')}
          aria-describedby={invalid('title') ? errId : undefined}
          required
        />
      </label>

      <div className="event-form__when">
        <label className="event-form__date">
          Date
          <input
            id={`${idp}-date`}
            type="date"
            value={d.date}
            onChange={(e) => set('date', e.target.value)}
            aria-invalid={invalid('date')}
            aria-describedby={invalid('date') ? errId : undefined}
            required
          />
        </label>
        <label className="event-form__allday">
          <input type="checkbox" checked={d.allDay} onChange={(e) => set('allDay', e.target.checked)} />
          All day
        </label>
      </div>

      {!d.allDay ? (
        <div className="form-row form-row--two-up">
          <label>
            Start time
            <input
              id={`${idp}-startTime`}
              type="time"
              value={d.startTime}
              step={300}
              onChange={(e) => set('startTime', e.target.value)}
              aria-invalid={invalid('startTime')}
              aria-describedby={invalid('startTime') ? errId : undefined}
            />
          </label>
          <label>
            <span>
              End time <span className="event-form__optional">(optional)</span>
            </span>
            <input
              id={`${idp}-endTime`}
              type="time"
              value={d.endTime}
              step={300}
              onChange={(e) => set('endTime', e.target.value)}
              aria-invalid={invalid('endTime')}
              aria-describedby={invalid('endTime') ? errId : undefined}
            />
          </label>
        </div>
      ) : null}
      {!d.allDay ? <p className="event-form__hint">Times are {zoneName}.</p> : null}

      <label>
        <span>
          Location <span className="event-form__optional">(optional)</span>
        </span>
        <input
          id={`${idp}-location`}
          type="text"
          value={d.location}
          maxLength={200}
          placeholder="Albertson Ward Building, Cultural Hall"
          onChange={(e) => set('location', e.target.value)}
          aria-invalid={invalid('location')}
        />
        <span className="event-form__hint">A building name or street address. The public page adds a map link.</span>
      </label>

      <label>
        <span>
          Description <span className="event-form__optional">(optional)</span>
        </span>
        <textarea
          id={`${idp}-description`}
          value={d.description}
          maxLength={MAX_DESCRIPTION}
          rows={4}
          placeholder={"Breakfast\nMusic\nShort shadow nativity"}
          onChange={(e) => set('description', e.target.value)}
          aria-invalid={invalid('description')}
          aria-describedby={`${idp}-desc-hint ${idp}-privacy`}
        />
        <span id={`${idp}-desc-hint`} className="event-form__hint">
          Put each item on its own line and it will show as a gold bullet on the events page. A single
          paragraph stays a paragraph.
        </span>
      </label>
      <p id={`${idp}-privacy`} className="gallery-privacy-note event-form__privacy">
        <strong>Privacy:</strong> this page is public. Please don&apos;t put children&apos;s full
        names in event titles or descriptions. “Primary Baptism” or “Baptism” is plenty. Leave out
        home addresses and phone numbers too.
      </p>

      <fieldset className="event-form__flyer">
        <legend>
          Flyer <span className="event-form__optional">(optional)</span>
        </legend>
        {flyerPreview ? (
          <div className="event-form__flyer-preview">
            <img src={flyerPreview} alt="Flyer preview" />
          </div>
        ) : null}
        <div className="event-form__flyer-actions">
          <button type="button" className="btn-secondary" disabled={busy} onClick={() => fileRef.current?.click()}>
            {flyerPreview ? 'Replace Flyer' : 'Upload a Flyer'}
          </button>
          {flyerPreview ? (
            <button type="button" className="event-form__text-btn" disabled={busy} onClick={clearFlyer}>
              Remove Flyer
            </button>
          ) : null}
        </div>
        <input
          ref={fileRef}
          id={`${idp}-flyer`}
          type="file"
          accept="image/jpeg,image/png,image/webp,image/*"
          hidden
          onChange={(e) => {
            chooseFlyer(e.target.files?.[0])
            e.target.value = ''
          }}
        />
        <p className="event-form__hint">
          JPEG, PNG, or WebP, up to 15 MB. On iPhone, pick a photo from your library.
        </p>
      </fieldset>

      <label>
        <span>
          Or paste a link <span className="event-form__optional">(optional)</span>
        </span>
        <input
          id={`${idp}-link`}
          type="url"
          inputMode="url"
          value={d.link}
          maxLength={500}
          placeholder="https://…"
          onChange={(e) => set('link', e.target.value)}
          aria-invalid={invalid('link')}
          aria-describedby={invalid('link') ? errId : undefined}
        />
        <span className="event-form__hint">
          For a sign-up sheet or broadcast page. If you also upload a flyer, the photo is what people see on the events page.
        </span>
      </label>

      {error ? (
        <p id={errId} className="event-form__error" role="alert">
          {error.message}
        </p>
      ) : null}

      <div className="event-form__actions">
        <button type="submit" className="btn-primary" disabled={busy}>
          {busy ? 'Saving…' : editing ? 'Save Changes' : 'Add Event'}
        </button>
        <button
          type="button"
          className="btn-secondary"
          disabled={busy}
          onClick={() => {
            if (dirty && !window.confirm('Discard your changes to this event?')) return
            onCancel()
          }}
        >
          Cancel
        </button>
      </div>
    </form>
  )
}

function AdminEventRow({
  event,
  busy,
  onEdit,
  onDelete,
}: {
  event: WardEvent
  busy: boolean
  onEdit: () => void
  onDelete: () => void
}) {
  return (
    <div className={`admin-event${event.upcoming ? '' : ' is-past'}`}>
      <EventDateBadge date={event.date} muted={!event.upcoming} />
      <div className="admin-event__body">
        <p className="admin-event__title">{event.title}</p>
        <p className="admin-event__meta">
          {event.timeLabel}
          {event.location ? ` · ${event.location}` : ''}
          {event.hasFlyer ? ' · Flyer' : event.link ? ' · Link' : ''}
        </p>
      </div>
      <div className="admin-event__actions" role="group" aria-label={`Actions for ${event.title}`}>
        <button type="button" className="btn-secondary" disabled={busy} onClick={onEdit}>
          Edit
        </button>
        <button type="button" className="btn-danger" disabled={busy} onClick={onDelete}>
          Delete
        </button>
      </div>
    </div>
  )
}

export function AdminEvents() {
  const [events, setEvents] = useState<WardEvent[] | null>(null)
  const [timeZone, setTimeZone] = useState('America/Denver')
  const [err, setErr] = useState<string | null>(null)
  const [toast, setToast] = useState<ToastPayload | null>(null)
  // null = closed, 'new' = add form, number = editing that event
  const [editing, setEditing] = useState<'new' | number | null>(null)
  const [dirty, setDirty] = useState(false)
  const [busyId, setBusyId] = useState<number | null>(null)
  const [showPast, setShowPast] = useState(false)

  useUnsavedChangesWarning(dirty, 'You have an event that hasn’t been saved. Leave anyway?')

  const load = useCallback(async () => {
    try {
      const d = await apiJson<EventsResponse>('/api/admin/events')
      setEvents(Array.isArray(d.events) ? d.events : [])
      if (d.timeZone) setTimeZone(d.timeZone)
      setErr(null)
    } catch (e) {
      setErr(e instanceof Error ? e.message : 'Could not load events.')
      setEvents((prev) => prev ?? [])
    }
  }, [])

  useEffect(() => {
    void load()
  }, [load])

  const onDirtyChange = useCallback((v: boolean) => setDirty(v), [])

  function openEditor(next: 'new' | number) {
    if (dirty && editing !== next && !window.confirm('Discard the changes you haven’t saved?')) return
    setDirty(false)
    setEditing(next)
  }

  async function onSaved(e: WardEvent, created: boolean, warning?: string) {
    setEditing(null)
    setDirty(false)
    const base = created
      ? `Added “${e.title}”. ${e.upcoming ? 'It’s on the public events page now.' : 'It’s in the past, so it’s under Past Events.'}`
      : `Saved “${e.title}”.`
    setToast({
      message: warning ? `${base} The flyer could not be saved: ${warning}` : base,
      variant: warning ? 'error' : 'success',
    })
    await load()
    window.requestAnimationFrame(() =>
      document.getElementById(`admin-event-${e.id}`)?.scrollIntoView({ block: 'nearest', behavior: 'smooth' }),
    )
  }

  async function remove(e: WardEvent) {
    if (!window.confirm(`Delete “${e.title}” on ${e.dateLabel}? This can’t be undone.`)) return
    setBusyId(e.id)
    try {
      await apiJson(`/api/admin/events/${e.id}/delete`, { method: 'POST' })
      if (editing === e.id) setEditing(null)
      setToast({ message: `Deleted “${e.title}”.`, variant: 'success' })
      await load()
    } catch (er) {
      setToast({ message: er instanceof Error ? er.message : 'Could not delete.', variant: 'error' })
    } finally {
      setBusyId(null)
    }
  }

  const zoneName = zoneDisplayName(timeZone)
  const upcoming = (events ?? []).filter((e) => e.upcoming)
  // Most recent first for past events.
  const past = (events ?? []).filter((e) => !e.upcoming).reverse()

  function renderList(list: WardEvent[]) {
    return (
      <ul className="admin-event-list" role="list">
        {list.map((e) => (
          <li key={e.id} id={`admin-event-${e.id}`}>
            {editing === e.id ? (
              <div className="admin-event-editor">
                <h3 className="admin-event-editor__title">Edit Event</h3>
                <EventForm
                  initial={draftFrom(e)}
                  editing={e}
                  zoneName={zoneName}
                  onSaved={(saved) => void onSaved(saved, false)}
                  onCancel={() => setEditing(null)}
                  onDirtyChange={onDirtyChange}
                />
              </div>
            ) : (
              <AdminEventRow event={e} busy={busyId === e.id} onEdit={() => openEditor(e.id)} onDelete={() => void remove(e)} />
            )}
          </li>
        ))}
      </ul>
    )
  }

  return (
    <main id="main">
      <PageHero
        titleId="admin-events-title"
        eyebrow="Leader Tools"
        title="Ward Events"
        lede="Add baptisms, ward activities, temple trips, and service projects to the public calendar."
        compact
      />
      <div className="wrap admin-events-page">
        <Toast toast={toast} onDismiss={() => setToast(null)} durationMs={6000} />
        {err ? (
          <p className="panel-sites panel-sites--warn" role="alert">
            {err}
          </p>
        ) : null}

        {editing === 'new' ? (
          <section className="admin-event-editor admin-event-editor--new" aria-labelledby="new-event-h">
            <h2 id="new-event-h" className="moderate-section-title">
              Add an Event
            </h2>
            <EventForm
              initial={EMPTY}
              editing={null}
              zoneName={zoneName}
              onSaved={(saved) => void onSaved(saved, true)}
              onCancel={() => setEditing(null)}
              onDirtyChange={onDirtyChange}
            />
          </section>
        ) : (
          <div className="admin-events-top">
            <button type="button" className="btn-primary admin-events-add" onClick={() => openEditor('new')}>
              + Add Event
            </button>
            <Link to="/events" className="admin-events-view">
              View the public events page
            </Link>
          </div>
        )}

        <section aria-labelledby="upcoming-events-h">
          <h2 id="upcoming-events-h" className="moderate-section-title moderate-section-title--spaced">
            Upcoming{events && upcoming.length ? ` (${upcoming.length})` : ''}
          </h2>
          {events === null ? (
            <p className="loading-note" role="status">
              <span className="loading-note__spinner" aria-hidden="true" />
              Loading…
            </p>
          ) : upcoming.length === 0 ? (
            <div className="panel-sites">
              <p className="lead" style={{ margin: 0 }}>
                No upcoming events. Use <strong>Add Event</strong> to put the next baptism, ward activity,
                temple trip, or service project on the calendar.
              </p>
            </div>
          ) : (
            renderList(upcoming)
          )}
        </section>

        {past.length ? (
          <section aria-labelledby="past-events-h" className="admin-events-past">
            <h2 id="past-events-h" className="moderate-section-title moderate-section-title--spaced">
              <button
                type="button"
                className="admin-events-past__toggle"
                aria-expanded={showPast}
                onClick={() => setShowPast((v) => !v)}
              >
                Past Events ({past.length}) <span aria-hidden="true">{showPast ? '▾' : '▸'}</span>
              </button>
            </h2>
            {showPast ? (
              <>
                <p className="panel-sites__note" style={{ marginTop: 0 }}>
                  The public page shows the last 12 months of past events when someone taps “Show Past Events”.
                </p>
                {renderList(past)}
              </>
            ) : null}
          </section>
        ) : null}
      </div>
    </main>
  )
}
