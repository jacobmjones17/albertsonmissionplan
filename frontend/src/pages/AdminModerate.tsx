import { createPortal } from 'react-dom'
import { useCallback, useEffect, useRef, useState } from 'react'
import type { FormEvent } from 'react'
import { Link } from 'react-router-dom'
import { apiJson } from '../api'
import { useBootstrap } from '../BootstrapContext'
import { PageHero } from '../components/PageHero'
import { PrivacyHighlightedBodyField } from '../components/PrivacyHighlightedBodyField'
import { Toast, type ToastPayload } from '../components/Toast'

type Pending = {
  id: number
  body: string
  author: string
  when: string
}

function previewSnippet(body: string, maxLen: number): string {
  const t = body.trim().replace(/\s+/g, ' ')
  if (t.length <= maxLen) {
    return t
  }
  return `${t.slice(0, maxLen).trimEnd()}…`
}

export function AdminModerate() {
  const { refresh } = useBootstrap()
  const dialogRef = useRef<HTMLDialogElement>(null)
  const [pending, setPending] = useState<Pending[]>([])
  const [selected, setSelected] = useState<Pending | null>(null)
  const [toast, setToast] = useState<ToastPayload | null>(null)
  const [err, setErr] = useState<string | null>(null)

  const load = useCallback(async () => {
    setErr(null)
    try {
      const d = await apiJson<{ pending: Pending[] | null }>('/api/admin/pending')
      setPending(Array.isArray(d.pending) ? d.pending : [])
    } catch (e) {
      setErr(e instanceof Error ? e.message : 'Could not load queue.')
    }
  }, [])

  useEffect(() => {
    void load()
    void refresh()
  }, [load, refresh])

  useEffect(() => {
    const dlg = dialogRef.current
    if (!dlg) return
    if (selected) {
      if (!dlg.open) dlg.showModal()
    } else {
      dlg.close()
    }
  }, [selected])

  async function approve(e: FormEvent, id: number, body: string, author: string) {
    e.preventDefault()
    setToast(null)
    try {
      await apiJson<{ ok: boolean; message?: string }>('/api/admin/moderate', {
        method: 'POST',
        json: { id, action: 'approve', body, author },
      })
      setToast({ message: 'Published.', variant: 'success' })
      dialogRef.current?.close()
      await load()
    } catch (ex) {
      setToast({
        message: ex instanceof Error ? ex.message : 'Could not publish.',
        variant: 'error',
      })
    }
  }

  async function reject(e: FormEvent, id: number, note: string) {
    e.preventDefault()
    setToast(null)
    try {
      await apiJson('/api/admin/moderate', {
        method: 'POST',
        json: { id, action: 'reject', note },
      })
      setToast({ message: 'Rejected.', variant: 'success' })
      dialogRef.current?.close()
      await load()
    } catch (ex) {
      setToast({
        message: ex instanceof Error ? ex.message : 'Could not reject.',
        variant: 'error',
      })
    }
  }

  return (
    <>
      <PageHero
        titleId="mod-title"
        eyebrow="Moderation"
        title="Review mission experiences"
        lede="Edit text to remove full names or identifying details before publishing. What you publish is exactly what the ward will see."
        compact
      />
      <main id="main" className="wrap">
        <p className="admin-crumb">
          <Link to="/admin">&larr; Back to leader tools</Link>
        </p>
        <p className="privacy-hint-note">
          In <strong>Review</strong>, name-like phrases are listed first when found, with the same text
          highlighted in yellow in the editor. Use <strong>Ignore</strong> for harmless matches or edit
          before publishing—guesses only.
        </p>
        <Toast toast={toast} onDismiss={() => setToast(null)} />
        {err ? <p className="panel-sites panel-sites--warn">{err}</p> : null}
        {pending.length === 0 ? (
          <div className="panel-sites">
            <p className="lead" style={{ margin: 0 }}>
              No submissions waiting. Check back later.
            </p>
          </div>
        ) : (
          <ul className="moderate-queue">
            {pending.map((p) => (
              <li key={p.id} className="moderate-queue-item">
                <div className="moderate-queue-main">
                  <p className="moderate-queue-meta">
                    Submitted {p.when} · #{p.id}
                  </p>
                  <p className="moderate-queue-preview">{previewSnippet(p.body, 140)}</p>
                  {p.author ? <p className="moderate-queue-byline">From: {p.author}</p> : null}
                </div>
                <button
                  type="button"
                  className="btn-secondary moderate-queue-review"
                  onClick={() => setSelected(p)}
                >
                  Review
                </button>
              </li>
            ))}
          </ul>
        )}
        {createPortal(
          <dialog
            ref={dialogRef}
            className="moderate-review-dialog"
            aria-labelledby={selected ? 'moderate-review-title' : undefined}
            onClose={() => setSelected(null)}
          >
            {selected ? (
              <ModerateReviewPanel
                key={selected.id}
                p={selected}
                onClose={() => dialogRef.current?.close()}
                onApprove={approve}
                onReject={reject}
              />
            ) : null}
          </dialog>,
          document.body,
        )}
      </main>
    </>
  )
}

function ModerateReviewPanel({
  p,
  onClose,
  onApprove,
  onReject,
}: {
  p: Pending
  onClose: () => void
  onApprove: (e: FormEvent, id: number, body: string, author: string) => void
  onReject: (e: FormEvent, id: number, note: string) => void
}) {
  const [body, setBody] = useState(p.body)
  const [author, setAuthor] = useState(p.author)
  const [note, setNote] = useState('')

  return (
    <div className="moderate-review-panel">
      <div className="moderate-review-panel-head">
        <h2 id="moderate-review-title" className="moderate-review-panel-title">
          Review submission #{p.id}
        </h2>
        <button type="button" className="btn-secondary moderate-review-panel-close" onClick={onClose}>
          Close
        </button>
      </div>
      <div className="moderate-review-panel-body">
        <p className="moderate-meta">
          Submitted {p.when}
          {p.author ? ` · From: ${p.author}` : ''}
        </p>
        <form
          className="form-stack moderate-approve-form"
          onSubmit={(e) => onApprove(e, p.id, body, author)}
        >
          <PrivacyHighlightedBodyField
            label="Text to publish (edit to protect privacy)"
            value={body}
            onChange={setBody}
            maxLength={8000}
            textareaProps={{ name: 'body', rows: 10, required: true }}
          />
          <label>
            Display name on site (optional)
            <input
              type="text"
              name="author"
              maxLength={120}
              placeholder="e.g. Anonymous, or a first name only"
              value={author}
              onChange={(e) => setAuthor(e.target.value)}
            />
          </label>
          <button type="submit" className="btn-primary">
            Publish
          </button>
        </form>
        <form className="inline-form reject-form" onSubmit={(e) => onReject(e, p.id, note)}>
          <label>
            Reason for rejecting (optional)
            <input
              type="text"
              name="note"
              maxLength={500}
              placeholder="e.g. Needs less identifying detail—saved for leaders only"
              value={note}
              onChange={(e) => setNote(e.target.value)}
            />
          </label>
          <button type="submit" className="btn-danger">
            Reject
          </button>
        </form>
        <p className="moderate-reject-hint">
          This note is seen only by leaders reviewing the ward tool. Submitters never see it. It helps
          you or other leaders remember why something was declined.
        </p>
      </div>
    </div>
  )
}
