import {
  useCallback,
  useEffect,
  useId,
  useMemo,
  useState,
  type ChangeEvent,
  type FormEvent,
} from 'react'
import { Link, useSearchParams } from 'react-router-dom'
import { apiJson } from '../api'
import { useBootstrap } from '../BootstrapContext'
import { PageHero } from '../components/PageHero'
import { Toast, type ToastPayload } from '../components/Toast'
import { useLeaderCounts } from '../lib/leaderCounts'
import { PASSWORD_GUIDANCE, PASSWORD_MAX_LENGTH } from '../lib/passwordRules'

type AuthTab = 'login' | 'register'

type PasswordInputProps = {
  value: string
  onChange: (next: string) => void
  autoComplete: 'current-password' | 'new-password'
  required?: boolean
  maxLength?: number
}

/** Password input with an eye-icon toggle that reveals/hides the typed value. */
function PasswordInput({
  value,
  onChange,
  autoComplete,
  required,
  maxLength,
}: PasswordInputProps) {
  const [show, setShow] = useState(false)
  const inputId = useId()
  return (
    <span className="password-field">
      <input
        id={inputId}
        type={show ? 'text' : 'password'}
        autoComplete={autoComplete}
        required={required}
        maxLength={maxLength}
        value={value}
        onChange={(e: ChangeEvent<HTMLInputElement>) => onChange(e.target.value)}
      />
      <button
        type="button"
        className="password-field__toggle"
        aria-label={show ? 'Hide password' : 'Show password'}
        aria-pressed={show}
        aria-controls={inputId}
        onClick={() => setShow((s) => !s)}
      >
        {show ? <EyeOffIcon /> : <EyeIcon />}
      </button>
    </span>
  )
}

function EyeIcon() {
  return (
    <svg
      width="20"
      height="20"
      viewBox="0 0 24 24"
      fill="none"
      stroke="currentColor"
      strokeWidth="1.75"
      strokeLinecap="round"
      strokeLinejoin="round"
      aria-hidden="true"
      focusable="false"
    >
      <path d="M2.25 12s3.75-7.5 9.75-7.5 9.75 7.5 9.75 7.5-3.75 7.5-9.75 7.5S2.25 12 2.25 12z" />
      <circle cx="12" cy="12" r="3" />
    </svg>
  )
}

function EyeOffIcon() {
  return (
    <svg
      width="20"
      height="20"
      viewBox="0 0 24 24"
      fill="none"
      stroke="currentColor"
      strokeWidth="1.75"
      strokeLinecap="round"
      strokeLinejoin="round"
      aria-hidden="true"
      focusable="false"
    >
      <path d="M3 3l18 18" />
      <path d="M10.58 10.58a3 3 0 0 0 4.24 4.24" />
      <path d="M9.88 5.09A10.94 10.94 0 0 1 12 4.5c6 0 9.75 7.5 9.75 7.5a17.43 17.43 0 0 1-3.07 4.05" />
      <path d="M6.1 6.1A17.5 17.5 0 0 0 2.25 12S6 19.5 12 19.5a10.94 10.94 0 0 0 5.18-1.32" />
    </svg>
  )
}

function InfoIcon() {
  return (
    <svg
      width="22"
      height="22"
      viewBox="0 0 24 24"
      fill="none"
      stroke="currentColor"
      strokeWidth="1.75"
      strokeLinecap="round"
      strokeLinejoin="round"
      aria-hidden="true"
      focusable="false"
    >
      <circle cx="12" cy="12" r="9" />
      <line x1="12" y1="11" x2="12" y2="16" />
      <line x1="12" y1="7.5" x2="12" y2="7.6" />
    </svg>
  )
}

type RegisterResponse = {
  ok: boolean
  status?: 'pending' | 'approved'
  message?: string
}

function CountStatus({ count, waiting, clear }: { count: number | null; waiting: (n: number) => string; clear: string }) {
  if (count === null) return null
  return count > 0 ? (
    <span className="admin-links__status admin-links__status--waiting">{waiting(count)}</span>
  ) : (
    <span className="admin-links__status admin-links__status--clear">{clear}</span>
  )
}


export function AdminHome() {
  const { user, refresh } = useBootstrap()
  const [params] = useSearchParams()
  const msg = useMemo(() => params.get('msg'), [params])
  const signInEmail = useMemo(() => params.get('email'), [params])

  const [tab, setTab] = useState<AuthTab>('login')
  const [email, setEmail] = useState('')
  const [firstName, setFirstName] = useState('')
  const [lastName, setLastName] = useState('')
  const [password, setPassword] = useState('')
  const [password2, setPassword2] = useState('')
  const [formErr, setFormErr] = useState<string | null>(null)
  const [busy, setBusy] = useState(false)
  const [toast, setToast] = useState<ToastPayload | null>(null)

  // At-a-glance counts shared with the leader nav (see lib/leaderCounts).
  const { pendingExperiences, pendingAccounts, error: countsErr } = useLeaderCounts()

  useEffect(() => {
    void refresh()
  }, [refresh])

  const [galleryCount, setGalleryCount] = useState<{ total: number; hidden: number } | null>(null)
  useEffect(() => {
    if (!user?.isLeader) {
      setGalleryCount(null)
      return
    }
    apiJson<{ items: { hidden?: boolean }[] | null }>('/api/admin/gallery')
      .then((d) => {
        const list = Array.isArray(d.items) ? d.items : []
        setGalleryCount({ total: list.length, hidden: list.filter((i) => i.hidden).length })
      })
      .catch(() => setGalleryCount(null))
  }, [user?.isLeader])

  const [upcomingEvents, setUpcomingEvents] = useState<number | null>(null)
  useEffect(() => {
    if (!user?.isLeader) {
      setUpcomingEvents(null)
      return
    }
    apiJson<{ events: { upcoming?: boolean }[] | null }>('/api/admin/events')
      .then((d) => setUpcomingEvents((Array.isArray(d.events) ? d.events : []).filter((e) => e.upcoming).length))
      .catch(() => setUpcomingEvents(null))
  }, [user?.isLeader])

  useEffect(() => {
    if (signInEmail) setEmail(signInEmail)
  }, [signInEmail])

  /**
   * Reset the auth form when switching between sign-in and create-account tabs so half-typed
   * passwords / mismatched-password errors don't leak from one form to the other.
   */
  const switchTab = useCallback((next: AuthTab) => {
    setTab(next)
    setEmail('')
    setFirstName('')
    setLastName('')
    setPassword('')
    setPassword2('')
    setFormErr(null)
  }, [])

  async function submitLogin(ev: FormEvent) {
    ev.preventDefault()
    setFormErr(null)
    setBusy(true)
    try {
      await apiJson<{ ok: boolean }>('/auth/login', {
        method: 'POST',
        json: { email: email.trim(), password },
      })
      window.history.replaceState({}, '', '/admin')
      setPassword('')
      await refresh()
    } catch (e) {
      setFormErr(e instanceof Error ? e.message : 'Could not sign in.')
    } finally {
      setBusy(false)
    }
  }

  async function submitRegister(ev: FormEvent) {
    ev.preventDefault()
    setFormErr(null)
    if (password !== password2) {
      setFormErr('Passwords do not match.')
      return
    }
    setBusy(true)
    try {
      const res = await apiJson<RegisterResponse>('/auth/register', {
        method: 'POST',
        json: {
          email: email.trim(),
          firstName: firstName.trim(),
          lastName: lastName.trim(),
          password,
        },
      })
      setPassword('')
      setPassword2('')
      if (res.status === 'approved') {
        setToast({
          message: 'Account created — signed in as the first leader.',
          variant: 'success',
        })
        window.history.replaceState({}, '', '/admin')
        await refresh()
      } else {
        setToast({
          message:
            res.message ??
            'Account request submitted! An existing leader must approve it before you can sign in.',
          variant: 'success',
        })
        // Clean slate on the sign-in tab so the user doesn't see their own pending email
        // sitting in the field if they try to sign in immediately.
        switchTab('login')
      }
    } catch (e) {
      setFormErr(e instanceof Error ? e.message : 'Could not create account.')
    } finally {
      setBusy(false)
    }
  }

  const displayName = (user?.displayName ?? '').trim()

  return (
    <main id="main">
      <PageHero
        titleId="admin-title"
        eyebrow="Albertson Ward"
        title={user?.isLeader ? 'Leader Tools' : 'Leader Sign-In'}
        lede={
          user?.isLeader
            ? 'Review experiences, approve leaders, edit the ward plan, post events, and share drawings.'
            : 'For ward leaders: sign in to review experiences and edit the ward plan.'
        }
        compact
      />
      <div className="wrap">
        <Toast toast={toast} onDismiss={() => setToast(null)} durationMs={7000} />
        {msg ? <p className="flash">{msg}</p> : null}
        {user?.email && !user.isLeader ?
          <p className="panel-sites panel-sites--warn">That account is not authorized for leader tools.</p>
        : null}
        {user?.isLeader ?
          <>
            <p className="lead admin-signed-in">
              Signed in as: <strong>{displayName || user.email}</strong>
            </p>
            <ul className="admin-links" role="list">
              <li>
                <Link to="/admin/moderate">
                  <span className="admin-links__text">
                    <span className="admin-links__label">Review Mission Experiences</span>
                    <span className="admin-links__desc">Approve, edit, or remove member stories</span>
                    <CountStatus
                      count={pendingExperiences}
                      waiting={(n) => `${n} waiting for review`}
                      clear="Nothing waiting"
                    />
                  </span>
                  {pendingExperiences ?
                    <span className="admin-link-badge" aria-hidden="true">
                      {pendingExperiences}
                    </span>
                  : null}
                </Link>
              </li>
              <li>
                <Link to="/admin/approvals">
                  <span className="admin-links__text">
                    <span className="admin-links__label">Leader Accounts</span>
                    <span className="admin-links__desc">Approve new leaders and manage access</span>
                    <CountStatus
                      count={pendingAccounts}
                      waiting={(n) => `${n} ${n === 1 ? 'request' : 'requests'} waiting`}
                      clear="No requests waiting"
                    />
                  </span>
                  {pendingAccounts ?
                    <span className="admin-link-badge" aria-hidden="true">
                      {pendingAccounts}
                    </span>
                  : null}
                </Link>
              </li>
              <li>
                <Link to="/admin/events">
                  <span className="admin-links__text">
                    <span className="admin-links__label">Ward Events</span>
                    <span className="admin-links__desc">Add baptisms, activities, open houses, and firesides to the calendar</span>
                    {upcomingEvents !== null ? (
                      <span className="admin-links__status">
                        {upcomingEvents === 0
                          ? 'No upcoming events'
                          : `${upcomingEvents} upcoming ${upcomingEvents === 1 ? 'event' : 'events'}`}
                      </span>
                    ) : null}
                  </span>
                </Link>
              </li>
              <li>
                <Link to="/admin/gallery">
                  <span className="admin-links__text">
                    <span className="admin-links__label">Ward Gallery</span>
                    <span className="admin-links__desc">Upload, hide, or remove drawings from children and members</span>
                    {galleryCount !== null ? (
                      <span className="admin-links__status">
                        {galleryCount.total === 0
                          ? 'No drawings yet'
                          : `${galleryCount.total} ${galleryCount.total === 1 ? 'drawing' : 'drawings'}${galleryCount.hidden ? ` · ${galleryCount.hidden} hidden` : ''}`}
                      </span>
                    ) : null}
                  </span>
                </Link>
              </li>
              <li>
                <Link to="/admin/ward-plan">
                  <span className="admin-links__text">
                    <span className="admin-links__label">Edit Ward Plan Goals</span>
                    <span className="admin-links__desc">Ward-wide and organization goals</span>
                  </span>
                </Link>
              </li>
            </ul>
            {countsErr ?
              <p className="panel-sites panel-sites--warn" style={{ marginTop: '1rem' }}>
                {countsErr}
              </p>
            : null}
          </>
        : (
          <>
            <p className="lead">
              Register with your name, email, and a password, or sign in if you already have an
              account. New accounts must be approved by an existing leader before they can sign in.
            </p>
            <p style={{ marginTop: '1.25rem' }}>
              <button
                type="button"
                className={tab === 'login' ? 'btn btn-primary' : 'btn btn-secondary'}
                onClick={() => switchTab('login')}
              >
                Sign in
              </button>{' '}
              <button
                type="button"
                className={tab === 'register' ? 'btn btn-primary' : 'btn btn-secondary'}
                onClick={() => switchTab('register')}
              >
                Create account
              </button>
            </p>
            {formErr ? <p className="flash">{formErr}</p> : null}
            {tab === 'login' ?
              <form
                className="form-stack"
                style={{ marginTop: '1rem', maxWidth: '28rem' }}
                onSubmit={submitLogin}
              >
                <label>
                  Email
                  <input
                    type="email"
                    autoComplete="email"
                    required
                    value={email}
                    onChange={(e) => setEmail(e.target.value)}
                  />
                </label>
                <label>
                  Password
                  <PasswordInput
                    autoComplete="current-password"
                    required
                    value={password}
                    onChange={setPassword}
                  />
                </label>
                <p className="panel-sites__note" style={{ margin: 0 }}>
                  <Link to="/admin/forgot-password">Forgot password?</Link>
                </p>
                <p>
                  <button type="submit" className="btn btn-primary" disabled={busy}>
                    {busy ? 'Signing in…' : 'Sign in'}
                  </button>
                </p>
              </form>
            : (
              <form
                className="form-stack"
                style={{ marginTop: '1rem', maxWidth: '28rem' }}
                onSubmit={submitRegister}
              >
                <div className="toast-banner toast-banner--info" role="status" aria-live="polite">
                  <span className="toast-banner__icon" aria-hidden="true">
                    <InfoIcon />
                  </span>
                  <div className="toast-banner__body">
                    <strong className="toast-banner__title">Approval required</strong>
                    <p className="toast-banner__text">
                      After you submit, your account is queued for review. An existing leader has to
                      approve it before you can sign in. Approved leaders may get an email alert when
                      the server is configured for outbound mail.
                    </p>
                  </div>
                </div>
                <div className="form-row form-row--two-up">
                  <label>
                    First name
                    <input
                      type="text"
                      autoComplete="given-name"
                      required
                      maxLength={80}
                      value={firstName}
                      onChange={(e) => setFirstName(e.target.value)}
                    />
                  </label>
                  <label>
                    Last name
                    <input
                      type="text"
                      autoComplete="family-name"
                      required
                      maxLength={80}
                      value={lastName}
                      onChange={(e) => setLastName(e.target.value)}
                    />
                  </label>
                </div>
                <label>
                  Email
                  <input
                    type="email"
                    autoComplete="email"
                    required
                    value={email}
                    onChange={(e) => setEmail(e.target.value)}
                  />
                </label>
                <label>
                  Password
                  <PasswordInput
                    autoComplete="new-password"
                    required
                    maxLength={PASSWORD_MAX_LENGTH}
                    value={password}
                    onChange={setPassword}
                  />
                </label>
                <label>
                  Confirm password
                  <PasswordInput
                    autoComplete="new-password"
                    required
                    maxLength={PASSWORD_MAX_LENGTH}
                    value={password2}
                    onChange={setPassword2}
                  />
                </label>
                <p className="panel-sites__note" style={{ margin: 0 }}>
                  {PASSWORD_GUIDANCE}
                </p>
                <p>
                  <button type="submit" className="btn btn-primary" disabled={busy}>
                    {busy ? 'Submitting…' : 'Request account'}
                  </button>
                </p>
              </form>
            )}
          </>
        )}
      </div>
    </main>
  )
}
