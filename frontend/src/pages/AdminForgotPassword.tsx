import { type FormEvent, useState } from 'react'
import { Link } from 'react-router-dom'
import { apiJson } from '../api'
import { PageHero } from '../components/PageHero'

export function AdminForgotPassword() {
  const [email, setEmail] = useState('')
  const [busy, setBusy] = useState(false)
  const [doneMsg, setDoneMsg] = useState<string | null>(null)
  const [err, setErr] = useState<string | null>(null)

  async function onSubmit(ev: FormEvent) {
    ev.preventDefault()
    setErr(null)
    setDoneMsg(null)
    setBusy(true)
    try {
      const res = await apiJson<{ ok?: boolean; message?: string }>('/auth/forgot-password', {
        method: 'POST',
        json: { email: email.trim() },
      })
      setDoneMsg(res.message ?? 'If that email is registered, check your inbox.')
      setEmail('')
    } catch (e) {
      setErr(e instanceof Error ? e.message : 'Request failed.')
    } finally {
      setBusy(false)
    }
  }

  return (
    <>
      <PageHero
        titleId="forgot-password-title"
        eyebrow="Leader"
        title="Forgot password"
        lede="We’ll email a reset link to approved leader accounts."
        compact
      />
      <main id="main" className="wrap">
        <p className="lead">
          <Link to="/admin">← Back to sign in</Link>
        </p>
        {doneMsg ? <p className="flash">{doneMsg}</p> : null}
        {err ? <p className="flash">{err}</p> : null}
        <form className="form-stack" style={{ marginTop: '1rem', maxWidth: '28rem' }} onSubmit={onSubmit}>
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
          <p>
            <button type="submit" className="btn btn-primary" disabled={busy}>
              {busy ? 'Sending…' : 'Send reset link'}
            </button>
          </p>
        </form>
      </main>
    </>
  )
}
