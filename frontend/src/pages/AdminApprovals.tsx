import { createPortal } from 'react-dom'
import { useCallback, useEffect, useRef, useState } from 'react'
import type { FormEvent } from 'react'
import { Link } from 'react-router-dom'
import { apiJson } from '../api'
import { useBootstrap } from '../BootstrapContext'
import { PageHero } from '../components/PageHero'
import { Toast, type ToastPayload } from '../components/Toast'

type PendingAccount = {
  email: string
  firstName: string
  lastName: string
  fullName: string
  when: string
}

type ApprovedLeader = {
  email: string
  firstName: string
  lastName: string
  fullName: string
  approvedWhen: string
}

type ApprovalsDialog =
  | { kind: 'pending'; account: PendingAccount }
  | { kind: 'removeApproved'; account: ApprovedLeader }
  | null

function displayLabelPending(p: PendingAccount): string {
  const name = (p.fullName || `${p.firstName} ${p.lastName}`).trim()
  return name || p.email
}

function displayLabelApproved(a: ApprovedLeader): string {
  const name = (a.fullName || `${a.firstName} ${a.lastName}`).trim()
  return name || a.email
}

export function AdminApprovals() {
  const { user, refresh } = useBootstrap()
  const dialogRef = useRef<HTMLDialogElement>(null)
  const [pending, setPending] = useState<PendingAccount[]>([])
  const [approved, setApproved] = useState<ApprovedLeader[]>([])
  const [dialog, setDialog] = useState<ApprovalsDialog>(null)
  const [toast, setToast] = useState<ToastPayload | null>(null)
  const [err, setErr] = useState<string | null>(null)

  const load = useCallback(async () => {
    setErr(null)
    try {
      const [pendingRes, approvedRes] = await Promise.all([
        apiJson<{ pending: PendingAccount[] | null }>('/api/admin/pending-accounts'),
        apiJson<{ approved: ApprovedLeader[] | null }>('/api/admin/approved-leaders'),
      ])
      setPending(Array.isArray(pendingRes.pending) ? pendingRes.pending : [])
      setApproved(Array.isArray(approvedRes.approved) ? approvedRes.approved : [])
    } catch (e) {
      const msg = e instanceof Error ? e.message : 'Could not load accounts.'
      setErr(msg)
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

  async function approve(e: FormEvent, p: PendingAccount) {
    e.preventDefault()
    setToast(null)
    try {
      await apiJson<{ ok: boolean; message?: string }>('/api/admin/decide-account', {
        method: 'POST',
        json: { email: p.email, action: 'approve' },
      })
      setToast({
        message: `Approved ${displayLabelPending(p)}. They can sign in now.`,
        variant: 'success',
      })
      dialogRef.current?.close()
      await load()
    } catch (ex) {
      setToast({
        message: ex instanceof Error ? ex.message : 'Could not approve.',
        variant: 'error',
      })
    }
  }

  async function deny(e: FormEvent, p: PendingAccount) {
    e.preventDefault()
    setToast(null)
    try {
      await apiJson('/api/admin/decide-account', {
        method: 'POST',
        json: { email: p.email, action: 'deny' },
      })
      setToast({
        message: `Denied ${displayLabelPending(p)}. They can register again later.`,
        variant: 'success',
      })
      dialogRef.current?.close()
      await load()
    } catch (ex) {
      setToast({
        message: ex instanceof Error ? ex.message : 'Could not deny.',
        variant: 'error',
      })
    }
  }

  async function removeApproved(e: FormEvent, a: ApprovedLeader) {
    e.preventDefault()
    setToast(null)
    try {
      await apiJson('/api/admin/remove-approved-leader', {
        method: 'POST',
        json: { email: a.email },
      })
      setToast({
        message: `Removed leader access for ${displayLabelApproved(a)}.`,
        variant: 'success',
      })
      dialogRef.current?.close()
      await load()
      await refresh()
    } catch (ex) {
      setToast({
        message: ex instanceof Error ? ex.message : 'Could not remove.',
        variant: 'error',
      })
    }
  }

  const selfEmail = (user?.email ?? '').trim().toLowerCase()
  const canRemoveAnyApproved = approved.length > 1

  return (
    <>
      <PageHero
        titleId="approvals-title"
        eyebrow="Approvals"
        title="Leader accounts"
        lede="Approve pending requests and remove leader access when someone should no longer sign in."
        compact
      />
      <main id="main" className="wrap">
        <p className="admin-crumb">
          <Link to="/admin">&larr; Back to leader tools</Link>
        </p>
        <p className="privacy-hint-note">
          Approving an account gives full leader tools (moderation + ward plan). You can remove an
          approved leader below, but one leader must always remain — add another leader before
          removing the last one.
        </p>
        <Toast toast={toast} onDismiss={() => setToast(null)} />
        {err ? <p className="panel-sites panel-sites--warn">{err}</p> : null}

        <h2 className="moderate-section-title">Pending requests</h2>
        {pending.length === 0 ? (
          <div className="panel-sites">
            <p className="lead" style={{ margin: 0 }}>
              No pending requests. New leaders who register at <code>/admin</code> show up here.
            </p>
          </div>
        ) : (
          <ul className="moderate-queue">
            {pending.map((p) => (
              <li key={p.email} className="moderate-queue-item">
                <div className="moderate-queue-main">
                  <p className="moderate-queue-meta">Requested {p.when}</p>
                  <p className="moderate-queue-preview">{displayLabelPending(p)}</p>
                  <p className="moderate-queue-byline">{p.email}</p>
                </div>
                <button
                  type="button"
                  className="btn-secondary moderate-queue-review"
                  onClick={() => setDialog({ kind: 'pending', account: p })}
                >
                  Review
                </button>
              </li>
            ))}
          </ul>
        )}

        <h2 className="moderate-section-title moderate-section-title--spaced">Approved leaders</h2>
        <p className="panel-sites__note" style={{ marginTop: 0 }}>
          Removing access deletes their login — they can register again later if needed. You cannot
          remove the only remaining leader.
        </p>
        {approved.length === 0 ? (
          <div className="panel-sites">
            <p className="lead" style={{ margin: 0 }}>
              {err ? 'Could not load this section.' : 'No approved leaders listed.'}
            </p>
          </div>
        ) : (
          <ul className="moderate-queue">
            {approved.map((a) => (
              <li key={a.email} className="moderate-queue-item">
                <div className="moderate-queue-main">
                  <p className="moderate-queue-meta">
                    Approved {a.approvedWhen}
                    {selfEmail && a.email.toLowerCase() === selfEmail ?
                      <span className="panel-sites__note"> · You</span>
                    : null}
                  </p>
                  <p className="moderate-queue-preview">{displayLabelApproved(a)}</p>
                  <p className="moderate-queue-byline">{a.email}</p>
                </div>
                <div className="moderate-queue-actions">
                  <button
                    type="button"
                    className="btn-danger"
                    disabled={!canRemoveAnyApproved}
                    title={
                      canRemoveAnyApproved ?
                        'Remove leader access'
                      : 'Approve another leader before removing this one'
                    }
                    onClick={() => setDialog({ kind: 'removeApproved', account: a })}
                  >
                    Remove access
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
            aria-labelledby={
              dialog?.kind === 'pending' ? 'approval-review-title'
              : dialog?.kind === 'removeApproved' ?
                'remove-approved-title'
              : undefined
            }
            onClose={() => setDialog(null)}
          >
            {dialog?.kind === 'pending' ?
              <ApprovalReviewPanel
                key={dialog.account.email}
                p={dialog.account}
                onClose={() => dialogRef.current?.close()}
                onApprove={approve}
                onDeny={deny}
              />
            : null}
            {dialog?.kind === 'removeApproved' ?
              <RemoveApprovedPanel
                key={dialog.account.email}
                a={dialog.account}
                onClose={() => dialogRef.current?.close()}
                onConfirm={removeApproved}
              />
            : null}
          </dialog>,
          document.body,
        )}
      </main>
    </>
  )
}

function ApprovalReviewPanel({
  p,
  onClose,
  onApprove,
  onDeny,
}: {
  p: PendingAccount
  onClose: () => void
  onApprove: (e: FormEvent, p: PendingAccount) => void
  onDeny: (e: FormEvent, p: PendingAccount) => void
}) {
  const label = displayLabelPending(p)
  return (
    <div className="moderate-review-panel">
      <div className="moderate-review-panel-head">
        <h2 id="approval-review-title" className="moderate-review-panel-title">
          Review account request
        </h2>
        <button type="button" className="btn-secondary moderate-review-panel-close" onClick={onClose}>
          Close
        </button>
      </div>
      <div className="moderate-review-panel-body">
        <p className="moderate-meta">Requested {p.when}</p>
        <dl className="approval-detail-grid">
          <dt>Name</dt>
          <dd>{label}</dd>
          <dt>Email</dt>
          <dd>{p.email}</dd>
        </dl>
        <p className="panel-sites__note" style={{ marginTop: '1rem' }}>
          Approving gives this person full leader access (moderate experiences, edit the ward plan,
          approve future leaders). Only approve people you know.
        </p>
        <form className="form-stack moderate-approve-form" onSubmit={(e) => onApprove(e, p)}>
          <button type="submit" className="btn-primary">
            Approve
          </button>
        </form>
        <form className="inline-form reject-form" onSubmit={(e) => onDeny(e, p)}>
          <button type="submit" className="btn-danger">
            Deny
          </button>
        </form>
        <p className="moderate-reject-hint">
          Denying removes this request from the queue. The person can register again later if it was a
          mistake.
        </p>
      </div>
    </div>
  )
}

function RemoveApprovedPanel({
  a,
  onClose,
  onConfirm,
}: {
  a: ApprovedLeader
  onClose: () => void
  onConfirm: (e: FormEvent, a: ApprovedLeader) => void
}) {
  const label = displayLabelApproved(a)
  return (
    <div className="moderate-review-panel">
      <div className="moderate-review-panel-head">
        <h2 id="remove-approved-title" className="moderate-review-panel-title">
          Remove leader access?
        </h2>
        <button type="button" className="btn-secondary moderate-review-panel-close" onClick={onClose}>
          Close
        </button>
      </div>
      <div className="moderate-review-panel-body">
        <p className="moderate-meta">Approved {a.approvedWhen}</p>
        <dl className="approval-detail-grid">
          <dt>Name</dt>
          <dd>{label}</dd>
          <dt>Email</dt>
          <dd>{a.email}</dd>
        </dl>
        <p className="panel-sites panel-sites--warn" style={{ marginTop: '1rem' }}>
          They will no longer be able to sign in as a leader. This cannot be undone — they would need
          to submit a new registration request if appropriate.
        </p>
        <form onSubmit={(e) => onConfirm(e, a)}>
          <div className="moderate-actions" style={{ marginTop: '1rem' }}>
            <button type="submit" className="btn-danger">
              Yes, remove access
            </button>
            <button type="button" className="btn-secondary" onClick={onClose}>
              Cancel
            </button>
          </div>
        </form>
      </div>
    </div>
  )
}
