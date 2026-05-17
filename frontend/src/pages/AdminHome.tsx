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

type PendingAccount = {
  email: string
  firstName: string
  lastName: string
  fullName: string
  when: string
}
type PendingAccountsResponse = { pending: PendingAccount[] | null }

/** Same JSON shape as GET /api/admin/pending (mission experiences queue). */
type PendingExperiencesResponse = { pending: unknown[] | null }

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

  // Badge counters — flows live at /admin/moderate and /admin/approvals.
  const [pendingAccountCount, setPendingAccountCount] = useState(0)
  const [pendingExperienceCount, setPendingExperienceCount] = useState(0)
  const [pendingAccountsErr, setPendingAccountsErr] = useState<string | null>(null)
  const [pendingExperiencesErr, setPendingExperiencesErr] = useState<string | null>(null)

  useEffect(() => {
    void refresh()
  }, [refresh])

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

  const loadPendingAccountCount = useCallback(async () => {
    setPendingAccountsErr(null)
    try {
      const d = await apiJson<PendingAccountsResponse>('/api/admin/pending-accounts')
      setPendingAccountCount(Array.isArray(d.pending) ? d.pending.length : 0)
    } catch (e) {
      setPendingAccountsErr(e instanceof Error ? e.message : 'Could not load pending accounts.')
    }
  }, [])

  const loadPendingExperienceCount = useCallback(async () => {
    setPendingExperiencesErr(null)
    try {
      const d = await apiJson<PendingExperiencesResponse>('/api/admin/pending')
      setPendingExperienceCount(Array.isArray(d.pending) ? d.pending.length : 0)
    } catch (e) {
      setPendingExperiencesErr(e instanceof Error ? e.message : 'Could not load pending experiences.')
    }
  }, [])

  useEffect(() => {
    if (user?.isLeader) {
      void loadPendingAccountCount()
      void loadPendingExperienceCount()
    } else {
      setPendingAccountCount(0)
      setPendingExperienceCount(0)
    }
  }, [user?.isLeader, loadPendingAccountCount, loadPendingExperienceCount])

  /** Refetch queue counts when returning from another tab or app (common on phones). */
  useEffect(() => {
    if (!user?.isLeader) return
    const onVis = () => {
      if (document.visibilityState === 'visible') {
        void loadPendingAccountCount()
        void loadPendingExperienceCount()
      }
    }
    document.addEventListener('visibilitychange', onVis)
    return () => document.removeEventListener('visibilitychange', onVis)
  }, [user?.isLeader, loadPendingAccountCount, loadPendingExperienceCount])

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
    <>
      <PageHero
        titleId="admin-title"
        eyebrow="Leader"
        title="Leader tools"
        lede="Moderate mission experiences and update ward goals"
        compact
      />
      <main id="main" className="wrap">
        <Toast toast={toast} onDismiss={() => setToast(null)} durationMs={7000} />
        {msg ? <p className="flash">{msg}</p> : null}
        {user?.email && !user.isLeader ?
          <p className="panel-sites panel-sites--warn">That account is not authorized for leader tools.</p>
        : null}
        {user?.isLeader ?
          <>
            <p className="lead">
              Signed in as <strong>{displayName || user.email}</strong>
              {displayName ?
                <span className="panel-sites__note" style={{ marginLeft: '0.5rem' }}>
                  ({user.email})
                </span>
              : null}
            </p>
            <ul className="admin-links admin-links--sites" role="list">
              <li>
                <Link to="/admin/moderate">
                  <span className="admin-links__text">Review mission experiences</span>
                  {pendingExperienceCount > 0 ?
                    <span
                      className="admin-link-badge"
                      aria-label={`${pendingExperienceCount} mission ${pendingExperienceCount === 1 ? 'experience' : 'experiences'} awaiting review`}
                    >
                      {pendingExperienceCount}
                    </span>
                  : null}
                </Link>
              </li>
              <li>
                <Link to="/admin/approvals">
                  <span className="admin-links__text">Leader accounts</span>
                  {pendingAccountCount > 0 ?
                    <span className="admin-link-badge" aria-label={`${pendingAccountCount} pending`}>
                      {pendingAccountCount}
                    </span>
                  : null}
                </Link>
              </li>
              <li>
                <Link to="/admin/ward-plan">
                  <span className="admin-links__text">Edit ward plan goals</span>
                </Link>
              </li>
            </ul>
            {pendingAccountsErr ?
              <p className="panel-sites panel-sites--warn" style={{ marginTop: '1rem' }}>
                {pendingAccountsErr}
              </p>
            : null}
            {pendingExperiencesErr ?
              <p
                className="panel-sites panel-sites--warn"
                style={{ marginTop: pendingAccountsErr ? '0.5rem' : '1rem' }}
              >
                {pendingExperiencesErr}
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
                    maxLength={72}
                    value={password}
                    onChange={setPassword}
                  />
                </label>
                <label>
                  Confirm password
                  <PasswordInput
                    autoComplete="new-password"
                    required
                    maxLength={72}
                    value={password2}
                    onChange={setPassword2}
                  />
                </label>
                <p className="panel-sites__note" style={{ margin: 0 }}>
                  Pick any password you like (no length or character rules).
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
      </main>
    </>
  )
}
