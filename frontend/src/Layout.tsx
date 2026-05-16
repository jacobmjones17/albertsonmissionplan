import { useEffect, useState } from 'react'
import { Link, NavLink, Outlet, useLocation } from 'react-router-dom'
import { useBootstrap } from './BootstrapContext'
import { apiJson } from './api'

const nav = [
  { to: '/', label: 'Home', end: true },
  { to: '/ward-plan', label: 'Ward Plan' },
  { to: '/monthly-challenges', label: 'Monthly Challenges' },
  { to: '/experiences', label: 'Mission Experiences' },
  { to: '/love-share-invite', label: 'Love, Share, Invite' },
]

function BrandMark() {
  return (
    <span className="brand-mark" aria-hidden="true">
      <svg viewBox="0 0 24 24" fill="none" xmlns="http://www.w3.org/2000/svg">
        <path
          d="M12 2.5c-2 3-5 4-5 8a5 5 0 0 0 10 0c0-4-3-5-5-8Z"
          fill="currentColor"
          opacity="0.95"
        />
        <path
          d="M12 6.5c-1 2-2.5 2.5-2.5 4.5a2.5 2.5 0 1 0 5 0c0-2-1.5-2.5-2.5-4.5Z"
          fill="#fff"
          opacity="0.85"
        />
      </svg>
    </span>
  )
}

export function Layout() {
  const [open, setOpen] = useState(false)
  const loc = useLocation()

  useEffect(() => {
    setOpen(false)
  }, [loc.pathname])

  return (
    <>
      <a className="skip" href="#main">
        Skip to main content
      </a>
      <header className="site-header">
        <div className="header-inner">
          <NavLink className="brand" to="/" end>
            <BrandMark />
            <span className="brand-name">
              <span className="brand-eyebrow">Mission Plan</span>
              Albertson Ward
            </span>
          </NavLink>
          <button
            type="button"
            className="nav-toggle"
            aria-expanded={open}
            aria-controls="site-nav"
            aria-label={open ? 'Close menu' : 'Open menu'}
            onClick={() => setOpen((o) => !o)}
          >
            <span aria-hidden="true" />
          </button>
          <nav id="site-nav" className={`site-nav${open ? ' is-open' : ''}`} aria-label="Primary">
            <ul>
              {nav.map((item) => (
                <li key={item.to}>
                  <NavLink to={item.to} end={item.end}>
                    {item.label}
                  </NavLink>
                </li>
              ))}
            </ul>
          </nav>
        </div>
      </header>
      <Outlet />
      <footer className="site-footer">
        <div className="site-footer__inner">
          <p className="site-footer__meta">
            Albertson Ward &middot; Ward Mission &middot;{' '}
            <Link to="/admin">Leader sign-in</Link>
          </p>
          <div className="site-footer__contact">
            <p className="site-footer__contact-label">Contact</p>
            <p className="site-footer__contact-body">
              To request changes or share concerns about this site, contact the ward mission leaders:{' '}
              <strong>Brother Jacob Jones</strong> or <strong>Brother Michael Hallenberger</strong>.
            </p>
          </div>
          <p className="attribution">
            Built for ward members. Update content as goals change each year.
          </p>
        </div>
      </footer>
    </>
  )
}

export function AdminLayout() {
  const { refresh, user } = useBootstrap()

  async function signOut() {
    await apiJson<{ ok: boolean }>('/auth/logout', {
      method: 'POST',
      headers: { Accept: 'application/json' },
    })
    await refresh()
    window.location.href = '/'
  }

  return (
    <>
      <a className="skip" href="#main">
        Skip to main content
      </a>
      <header className="site-header">
        <div className="header-inner">
          <NavLink className="brand" to="/admin">
            <BrandMark />
            <span className="brand-name">
              <span className="brand-eyebrow">Leader Tools</span>
              Albertson Ward
            </span>
          </NavLink>
          {user?.isLeader ? (
            <button type="button" className="btn btn-secondary" onClick={() => void signOut()}>
              Sign out
            </button>
          ) : null}
        </div>
      </header>
      <Outlet />
      <footer className="site-footer">
        <p>
          Albertson Ward &middot; Ward Mission &middot; <a href="/">Public site</a>
        </p>
      </footer>
    </>
  )
}
