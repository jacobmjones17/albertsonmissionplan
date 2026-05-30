import { createPortal } from 'react-dom'
import { useCallback, useEffect, useRef, useState } from 'react'
import type { FormEvent } from 'react'
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
  const [published, setPublished] = useState<Published[]>([])
  const [dialog, setDialog] = useState<DialogState>(null)
  const [toast, setToast] = useState<ToastPayload | null>(null)
  const [err, setErr] = useState<string | null>(null)

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
  }, [])

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
    <>
      <PageHero
        titleId="mod-title"
        eyebrow="Moderation"
        title="Review mission experiences"
        lede="Edit text to remove full names or identifying details before publishing. What you publish is exactly what the ward will see."
        compact
      />
      <main id="main" className="wrap">
        <p className="privacy-hint-note">
          In <strong>Review</strong>, name-like phrases are listed first when found, with the same text
          highlighted in yellow in the editor. Use <strong>Ignore</strong> for harmless matches or edit
          before publishing—guesses only.
        </p>
        <Toast toast={toast} onDismiss={() => setToast(null)} />
        {err ? <p className="panel-sites panel-sites--warn">{err}</p> : null}

        <h2 className="moderate-section-title">Pending submissions</h2>
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
                  onClick={() => setDialog({ kind: 'review', item: p })}
                >
                  Review
                </button>
              </li>
            ))}
          </ul>
        )}

        <h2 className="moderate-section-title moderate-section-title--spaced">
          Published experiences
        </h2>
        <p className="panel-sites__note" style={{ marginTop: 0 }}>
          These are live on the public Mission Experiences page. Editing changes what visitors
          see immediately; deleting removes it permanently.
        </p>
        {published.length === 0 ? (
          <div className="panel-sites">
            <p className="lead" style={{ margin: 0 }}>
              Nothing published yet — approved submissions appear here.
            </p>
          </div>
        ) : (
          <ul className="moderate-queue">
            {published.map((p) => (
              <li key={p.id} className="moderate-queue-item">
                <div className="moderate-queue-main">
                  <p className="moderate-queue-meta">
                    Published {p.when} · #{p.id}
                  </p>
                  <p className="moderate-queue-preview">{previewSnippet(p.body, 140)}</p>
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
  onReject: (id: number) => void
}) {
  const [body, setBody] = useState(p.body)
  const [author, setAuthor] = useState(p.author)

  return (
    <div className="moderate-review-panel">
      <div className="moderate-review-panel-head">
        <h2 id="moderate-dialog-title" className="moderate-review-panel-title">
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
          Edit published experience #{p.id}
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
          Published {p.when} · #{p.id} · Shown as {p.author || 'A ward member'}
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
