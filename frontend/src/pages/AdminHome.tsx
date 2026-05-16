import { useEffect, useMemo } from 'react'
import { Link, useSearchParams } from 'react-router-dom'
import { useBootstrap } from '../BootstrapContext'
import { PageHero } from '../components/PageHero'

export function AdminHome() {
  const { user, refresh, googleSignInOk } = useBootstrap()
  const [params] = useSearchParams()
  const msg = useMemo(() => params.get('msg'), [params])
  const signInEmail = useMemo(() => params.get('email'), [params])

  useEffect(() => {
    void refresh()
  }, [refresh])

  return (
    <>
      <PageHero
        titleId="admin-title"
        eyebrow="Leader"
        title="Leader tools"
        lede="Moderate mission experiences and update ward organization goals"
        compact
      />
      <main id="main" className="wrap">
        {msg ? <p className="flash">{msg}</p> : null}
        {signInEmail && !user?.isLeader ? (
          <div className="panel-sites" style={{ marginTop: '1rem' }}>
            <p className="panel-sites__title" style={{ marginTop: 0 }}>
              Allowlist this address
            </p>
            <p className="panel-sites__p">
              Google signed you in as <strong>{signInEmail}</strong>. That exact address must appear in{' '}
              <code>ALLOWED_LEADER_EMAILS</code> in your repo root <code>.env</code> (comma-separated if
              there are several leaders).
            </p>
            <p className="panel-sites__p" style={{ marginBottom: 0 }}>
              Example: <code>ALLOWED_LEADER_EMAILS={signInEmail}</code>
            </p>
            <p className="panel-sites__note">Restart the Go server after saving <code>.env</code>, then use Continue with Google again.</p>
          </div>
        ) : null}
        {user?.email && !user.isLeader ? (
          <p className="panel-sites panel-sites--warn">That account is not authorized for leader tools.</p>
        ) : null}
        {user?.isLeader ? (
          <>
            <p className="lead">
              Signed in as <strong>{user.email}</strong>
            </p>
            <ul className="admin-links admin-links--sites" role="list">
              <li>
                <Link to="/admin/moderate">Review mission experiences</Link>
              </li>
              <li>
                <Link to="/admin/ward-plan">Edit ward plan goals</Link>
              </li>
              <li>
                <Link to="/">View public site</Link>
              </li>
            </ul>
          </>
        ) : (
          <>
            {googleSignInOk ? (
              <>
                {!user ? (
                  <p className="lead">
                    Sign in with Google to moderate member experiences and update organization goals.
                  </p>
                ) : null}
                <p>
                  <a className="btn btn-primary" href="/auth/google">
                    Continue with Google
                  </a>
                </p>
              </>
            ) : (
              <div className="panel-sites panel-sites--warn">
                <p className="panel-sites__title" style={{ marginTop: 0 }}>
                  Google sign-in is off on this server
                </p>
                <p className="panel-sites__p">
                  OAuth is not configured in the running process (for example{' '}
                  <code>DEV_SKIP_OAUTH=true</code> is set, or <code>.env</code> was not loaded). Leader
                  tools need <code>GOOGLE_CLIENT_ID</code>, <code>GOOGLE_CLIENT_SECRET</code>, and{' '}
                  <code>OAUTH_REDIRECT_URL</code>.
                </p>
                <p className="panel-sites__p">
                  <strong>To fix:</strong> in the repo root <code>.env</code>, set{' '}
                  <code>DEV_SKIP_OAUTH=false</code> (or remove the line), fill in the Google variables,
                  add your email to <code>ALLOWED_LEADER_EMAILS</code>, then <strong>fully stop and
                  restart</strong> the Go server. If it still fails, in the terminal run{' '}
                  <code>unset DEV_SKIP_OAUTH</code> in case the shell exported it earlier.
                </p>
              </div>
            )}
          </>
        )}
      </main>
    </>
  )
}
