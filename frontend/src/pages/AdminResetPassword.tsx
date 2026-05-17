import { type FormEvent, useEffect, useState } from 'react'
import { Link, useNavigate, useSearchParams } from 'react-router-dom'
import { apiJson } from '../api'
import { PageHero } from '../components/PageHero'

export function AdminResetPassword() {
  const [params] = useSearchParams()
  const navigate = useNavigate()
  const qToken = (params.get('t') ?? '').trim()

  const [token, setToken] = useState(qToken)
  const [password, setPassword] = useState('')
  const [password2, setPassword2] = useState('')
  const [busy, setBusy] = useState(false)
  const [err, setErr] = useState<string | null>(null)

  useEffect(() => {
    setToken(qToken)
  }, [qToken])

  async function onSubmit(ev: FormEvent) {
    ev.preventDefault()
    setErr(null)
    if (password !== password2) {
      setErr('Passwords do not match.')
      return
    }
    const t = token.trim()
    if (!t) {
      setErr('Open the link from your email, or paste the token from that link.')
      return
    }
    setBusy(true)
    try {
      await apiJson<{ ok: boolean }>('/auth/reset-password', {
        method: 'POST',
        json: { token: t, password },
      })
      navigate('/admin?msg=' + encodeURIComponent('Password updated — you can sign in.'), {
        replace: true,
      })
    } catch (e) {
      setErr(e instanceof Error ? e.message : 'Could not reset password.')
    } finally {
      setBusy(false)
    }
  }

  return (
    <>
      <PageHero
        titleId="reset-password-title"
        eyebrow="Leader"
        title="Choose a new password"
        lede="Paste the token from your email if it wasn’t filled in automatically."
        compact
      />
      <main id="main" className="wrap">
        <p className="lead">
          <Link to="/admin">← Back to sign in</Link>
        </p>
        {err ? <p className="flash">{err}</p> : null}
        <form className="form-stack" style={{ marginTop: '1rem', maxWidth: '28rem' }} onSubmit={onSubmit}>
          {!qToken ?
            <label>
              Reset token
              <input
                type="text"
                autoComplete="off"
                spellCheck={false}
                value={token}
                onChange={(e) => setToken(e.target.value)}
                placeholder="From the reset link"
              />
            </label>
          : null}
          <label>
            New password
            <input
              type="password"
              autoComplete="new-password"
              required
              maxLength={72}
              value={password}
              onChange={(e) => setPassword(e.target.value)}
            />
          </label>
          <label>
            Confirm password
            <input
              type="password"
              autoComplete="new-password"
              required
              maxLength={72}
              value={password2}
              onChange={(e) => setPassword2(e.target.value)}
            />
          </label>
          <p className="panel-sites__note" style={{ margin: 0 }}>
            Passwords may be up to 72 characters (bcrypt limit).
          </p>
          <p>
            <button type="submit" className="btn btn-primary" disabled={busy}>
              {busy ? 'Saving…' : 'Save password'}
            </button>
          </p>
        </form>
      </main>
    </>
  )
}
