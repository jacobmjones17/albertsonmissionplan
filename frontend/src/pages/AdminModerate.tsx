import { createPortal } from 'react-dom'
import { useCallback, useEffect, useMemo, useRef, useState } from 'react'
import type { FormEvent } from 'react'
import { apiJson } from '../api'
import { useBootstrap } from '../BootstrapContext'
import { PageHero } from '../components/PageHero'
import { PrivacyHighlightedBodyField } from '../components/PrivacyHighlightedBodyField'
import { Toast, type ToastPayload } from '../components/Toast'
import { useLeaderCounts } from '../lib/leaderCounts'

type Pending = {
  id: number
  body: string
  author: string
  when: string
}

type Published = {
  id: number
  body: string
  author: string
  when: string
}

/**
 * One of the three dialog modes the admin can open: review a pending submission, edit a
 * published one, or confirm a delete. Keeping all three as a single discriminated union lets
 * us drive a single shared <dialog> element rather than juggling refs for each.
 */
type DialogState =
  | { kind: 'review'; item: Pending }
  | { kind: 'edit'; item: Published }
  | { kind: 'delete'; item: Published }
  | null

const PREVIEW_LEN = 140

type PublishedSort = 'newest' | 'oldest' | 'author'

/** Queue preview that can expand to the full story inline (no dialog needed just to read it). */
function ExpandableStory({ body }: { body: string }) {
  const [open, setOpen] = useState(false)
  const long = body.trim().replace(/\s+/g, ' ').length > PREVIEW_LEN
  return (
    <>
      {open ? (
        <p className="moderate-queue-preview moderate-queue-preview--full">{body.trim()}</p>
      ) : (
        <p className="moderate-queue-preview">{previewSnippet(body, PREVIEW_LEN)}</p>
      )}
      {long ? (
        <button
          type="button"
          className="moderate-expand-btn"
          aria-expanded={open}
          onClick={() => setOpen((o) => !o)}
        >
          {open ? 'Show less' : 'Read full story'}
        </button>
      ) : null}
    </>
  )
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
  const { refresh: refreshCounts } = useLeaderCounts()
  const dialogRef = useRef<HTMLDialogElement>(null)
  const [pending, setPending] = useState<Pending[]>([])
  const [published, setPublished] = useState<Published[]>([])
  const [dialog, setDialog] = useState<DialogState>(null)
  const [toast, setToast] = useState<ToastPayload | null>(null)
  const [err, setErr] = useState<string | null>(null)
  const [query, setQuery] = useState('')
  const [sort, setSort] = useState<PublishedSort>('newest')

  const load = useCallback(async () => {
    setErr(null)
    try {
      const [pendingRes, publishedRes] = await Promise.all([
        apiJson<{ pending: Pending[] | null }>('/api/admin/pending'),
        apiJson<{ published: Published[] | null }>('/api/admin/published'),
      ])
      setPending(Array.isArray(pendingRes.pending) ? pendingRes.pending : [])
      setPublished(Array.isArray(publishedRes.published) ? publishedRes.published : [])
    } catch (e) {
      setErr(e instanceof Error ? e.message : 'Could not load queue.')
    }
    void refreshCounts()
  }, [refreshCounts])

  const shownPublished = useMemo(() => {
    const q = query.trim().toLowerCase()
    const list = q
      ? published.filter(
          (p) => p.body.toLowerCase().includes(q) || (p.author || '').toLowerCase().includes(q),
        )
      : [...published]
    // Ids increase as submissions arrive, so they sort newest/oldest reliably.
    if (sort === 'newest') list.sort((a, b) => b.id - a.id)
    else if (sort === 'oldest') list.sort((a, b) => a.id - b.id)
    else
      list.sort((a, b) =>
        (a.author || 'A ward member').localeCompare(b.author || 'A ward member', undefined, {
          sensitivity: 'base',
        }),
      )
    return list
  }, [published, query, sort])

  useEffect(() => {
    void load()
    void refresh()
  }, [load, refresh])

  useEffect(() => {
    const dlg = dialogRef.current
    if (!dlg) return
    if (dialog) {
      if (!dlg.open) dlg.showModal()
    } else {
      dlg.close()
    }
  }, [dialog])

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

  async function reject(id: number) {
    if (
      !window.confirm(
        'Reject this submission? It will not be published and will be removed from the review list.',
      )
    ) {
      return
    }
    setToast(null)
    try {
      await apiJson('/api/admin/moderate', {
        method: 'POST',
        json: { id, action: 'reject' },
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

  async function saveEdit(e: FormEvent, id: number, body: string, author: string) {
    e.preventDefault()
    setToast(null)
    try {
      await apiJson('/api/admin/published/edit', {
        method: 'POST',
        json: { id, body, author },
      })
      setToast({ message: 'Saved.', variant: 'success' })
      dialogRef.current?.close()
      await load()
    } catch (ex) {
      setToast({
        message: ex instanceof Error ? ex.message : 'Could not save changes.',
        variant: 'error',
      })
    }
  }

  async function confirmDelete(e: FormEvent, id: number) {
    e.preventDefault()
    setToast(null)
    try {
      await apiJson('/api/admin/published/delete', {
        method: 'POST',
        json: { id },
      })
      setToast({ message: 'Deleted.', variant: 'success' })
      dialogRef.current?.close()
      await load()
    } catch (ex) {
      setToast({
        message: ex instanceof Error ? ex.message : 'Could not delete.',
        variant: 'error',
      })
    }
  }

  return (
    <main id="main">
      <PageHero
        titleId="mod-title"
        eyebrow="Leader Tools"
        title="Review Mission Experiences"
        lede="Read each story, remove full names or private details, then publish. What you publish is exactly what the ward will see."
        compact
      />
      <div className="wrap">
        <Toast toast={toast} onDismiss={() => setToast(null)} />
        {err ? <p className="panel-sites panel-sites--warn">{err}</p> : null}

        <h2 className="moderate-section-title">
          Waiting for Review{pending.length ? ` (${pending.length})` : ''}
        </h2>
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
                    Submitted {p.when}
                  </p>
                  <ExpandableStory body={p.body} />
                  {p.author ? <p className="moderate-queue-byline">From: {p.author}</p> : null}
                </div>
                <button
                  type="button"
                  className="btn-secondary moderate-queue-review"
                  onClick={() => setDialog({ kind: 'review', item: p })}
                >
                  Review
                </button>
              </li>
            ))}
          </ul>
        )}

        <h2 className="moderate-section-title moderate-section-title--spaced">
          Published Experiences{published.length ? ` (${published.length})` : ''}
        </h2>
        <p className="panel-sites__note" style={{ marginTop: 0 }}>
          These are live on the public Mission Experiences page. Editing changes what visitors
          see immediately; deleting removes it permanently.
        </p>
        {published.length > 0 ? (
          <div className="moderate-filters" role="search">
            <label className="moderate-filters__search">
              <span>Search</span>
              <input
                type="search"
                placeholder="Words in the story or the name shown"
                value={query}
                onChange={(e) => setQuery(e.target.value)}
              />
            </label>
            <label className="moderate-filters__sort">
              <span>Sort</span>
              <select value={sort} onChange={(e) => setSort(e.target.value as PublishedSort)}>
                <option value="newest">Newest first</option>
                <option value="oldest">Oldest first</option>
                <option value="author">Name shown (A–Z)</option>
              </select>
            </label>
          </div>
        ) : null}
        {published.length === 0 ? (
          <div className="panel-sites">
            <p className="lead" style={{ margin: 0 }}>
              Nothing published yet — approved submissions appear here.
            </p>
          </div>
        ) : shownPublished.length === 0 ? (
          <div className="panel-sites">
            <p className="lead" style={{ margin: 0 }} role="status">
              No published experiences match “{query.trim()}”.
            </p>
          </div>
        ) : (
          <ul className="moderate-queue">
            {shownPublished.map((p) => (
              <li key={p.id} className="moderate-queue-item">
                <div className="moderate-queue-main">
                  <p className="moderate-queue-meta">
                    Published {p.when}
                  </p>
                  <ExpandableStory body={p.body} />
                  <p className="moderate-queue-byline">
                    Shown as: {p.author ? p.author : 'A ward member'}
                  </p>
                </div>
                <div className="moderate-queue-actions">
                  <button
                    type="button"
                    className="btn-secondary"
                    onClick={() => setDialog({ kind: 'edit', item: p })}
                  >
                    Edit
                  </button>
                  <button
                    type="button"
                    className="btn-danger"
                    onClick={() => setDialog({ kind: 'delete', item: p })}
                  >
                    Delete
                  </button>
                </div>
              </li>
            ))}
          </ul>
        )}

        {createPortal(
          <dialog
            ref={dialogRef}
            className="moderate-review-dialog"
            aria-labelledby={dialog ? 'moderate-dialog-title' : undefined}
            onClose={() => setDialog(null)}
          >
            {dialog?.kind === 'review' ? (
              <ModerateReviewPanel
                key={`review-${dialog.item.id}`}
                p={dialog.item}
                onClose={() => dialogRef.current?.close()}
                onApprove={approve}
                onReject={reject}
              />
            ) : null}
            {dialog?.kind === 'edit' ? (
              <EditPublishedPanel
                key={`edit-${dialog.item.id}`}
                p={dialog.item}
                onClose={() => dialogRef.current?.close()}
                onSave={saveEdit}
              />
            ) : null}
            {dialog?.kind === 'delete' ? (
              <DeleteConfirmPanel
                key={`delete-${dialog.item.id}`}
                p={dialog.item}
                onClose={() => dialogRef.current?.close()}
                onConfirm={confirmDelete}
              />
            ) : null}
          </dialog>,
          document.body,
        )}
      </div>
    </main>
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
  onReject: (id: number) => void
}) {
  const [body, setBody] = useState(p.body)
  const [author, setAuthor] = useState(p.author)

  return (
    <div className="moderate-review-panel">
      <div className="moderate-review-panel-head">
        <h2 id="moderate-dialog-title" className="moderate-review-panel-title">
          Review Submission
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
          <div
            className="moderate-review-actions-row"
            role="group"
            aria-label="Publish or reject this submission"
          >
            <button type="submit" className="btn-primary moderate-review-actions-row__publish">
              Publish
            </button>
            <button
              type="button"
              className="btn-danger moderate-review-actions-row__reject"
              onClick={() => void onReject(p.id)}
            >
              Reject
            </button>
          </div>
        </form>
      </div>
    </div>
  )
}

function EditPublishedPanel({
  p,
  onClose,
  onSave,
}: {
  p: Published
  onClose: () => void
  onSave: (e: FormEvent, id: number, body: string, author: string) => void
}) {
  const [body, setBody] = useState(p.body)
  const [author, setAuthor] = useState(p.author)

  return (
    <div className="moderate-review-panel">
      <div className="moderate-review-panel-head">
        <h2 id="moderate-dialog-title" className="moderate-review-panel-title">
          Edit Published Experience
        </h2>
        <button type="button" className="btn-secondary moderate-review-panel-close" onClick={onClose}>
          Close
        </button>
      </div>
      <div className="moderate-review-panel-body">
        <p className="moderate-meta">
          Originally published {p.when}. Changes go live immediately.
        </p>
        <form
          className="form-stack moderate-approve-form"
          onSubmit={(e) => onSave(e, p.id, body, author)}
        >
          <PrivacyHighlightedBodyField
            label="Published text (edits replace the live version)"
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
            Save changes
          </button>
        </form>
        <p className="moderate-reject-hint">
          To take this experience down entirely, close this dialog and click <strong>Delete</strong>.
        </p>
      </div>
    </div>
  )
}

function DeleteConfirmPanel({
  p,
  onClose,
  onConfirm,
}: {
  p: Published
  onClose: () => void
  onConfirm: (e: FormEvent, id: number) => void
}) {
  return (
    <div className="moderate-review-panel">
      <div className="moderate-review-panel-head">
        <h2 id="moderate-dialog-title" className="moderate-review-panel-title">
          Delete published experience?
        </h2>
        <button type="button" className="btn-secondary moderate-review-panel-close" onClick={onClose}>
          Close
        </button>
      </div>
      <div className="moderate-review-panel-body">
        <p className="moderate-meta">
          Published {p.when} · Shown as {p.author || 'A ward member'}
        </p>
        <blockquote className="moderate-queue-preview" style={{ marginTop: '0.5rem' }}>
          {previewSnippet(p.body, 400)}
        </blockquote>
        <p className="panel-sites panel-sites--warn" style={{ marginTop: '1rem' }}>
          This permanently removes the experience from the public site and the database. There
          is no undo.
        </p>
        <form className="inline-form reject-form" onSubmit={(e) => onConfirm(e, p.id)}>
          <button type="submit" className="btn-danger">
            Yes, delete permanently
          </button>
          <button type="button" className="btn-secondary" onClick={onClose} style={{ marginLeft: '0.5rem' }}>
            Cancel
          </button>
        </form>
      </div>
    </div>
  )
}
