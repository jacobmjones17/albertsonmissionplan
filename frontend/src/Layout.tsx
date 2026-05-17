import { useEffect, useState, type ReactNode } from 'react'
import { Link, NavLink, Outlet, useLocation } from 'react-router-dom'
import { apiJson } from './api'
import { useBootstrap } from './BootstrapContext'
import { STATIC_SITE_PHOTOS } from './staticSitePhotos'

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

/** Public navbar overlays the merged hero backdrop while scrolled near top (`site-header--overlap-home`). */
function SiteHeaderShell({
  mergeHeroBackdrop = false,
  children,
}: {
  mergeHeroBackdrop?: boolean
  children: ReactNode
}) {
  return (
    <header className={`site-header${mergeHeroBackdrop ? ' site-header--overlap-home' : ''}`}>
      {children}
    </header>
  )
}

export function Layout() {
  const { user } = useBootstrap()
  const [open, setOpen] = useState(false)
  const loc = useLocation()
  const headerBackdrop = String(STATIC_SITE_PHOTOS.header).trim()

  const [pastHeroBackdrop, setPastHeroBackdrop] = useState(false)
  const mergeHeroBackdrop = Boolean(headerBackdrop) && !pastHeroBackdrop

  useEffect(() => {
    setPastHeroBackdrop(false)
  }, [loc.pathname])

  useEffect(() => {
    if (!headerBackdrop) {
      return
    }
    function thresholdY() {
      const vh = typeof window !== 'undefined' ? window.innerHeight : 800
      return Math.min(vh * 0.42, 380)
    }
    function onScroll() {
      setPastHeroBackdrop(window.scrollY > thresholdY())
    }
    onScroll()
    window.addEventListener('scroll', onScroll, { passive: true })
    return () => window.removeEventListener('scroll', onScroll)
  }, [loc.pathname, headerBackdrop])

  useEffect(() => {
    setOpen(false)
  }, [loc.pathname])

  return (
    <>
      <a className="skip" href="#main">
        Skip to main content
      </a>
      <SiteHeaderShell mergeHeroBackdrop={mergeHeroBackdrop}>
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
      </SiteHeaderShell>
      <Outlet />
      <footer className="site-footer">
        <div className="site-footer__inner">
          <p className="site-footer__meta">
            Albertson Ward &middot; Ward Mission &middot;{' '}
            <Link to="/admin" title={user?.isLeader ? 'Open leader dashboard' : undefined}>
              {user?.isLeader ? 'Leader tools' : 'Leader sign-in'}
            </Link>
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
    try {
      await apiJson<{ ok: boolean }>('/auth/logout', {
        method: 'POST',
        headers: { Accept: 'application/json' },
      })
    } catch {
      try {
        await fetch('/auth/logout', { method: 'GET', credentials: 'include' })
      } catch {
        // ignore — we still navigate away below so the user is always signed out client-side.
      }
    }
    try {
      await refresh()
    } catch {
      // ignore — bootstrap may briefly fail mid-logout; the redirect below resets state.
    }
    window.location.href = '/'
  }

  return (
    <>
      <a className="skip" href="#main">
        Skip to main content
      </a>
      <SiteHeaderShell>
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
      </SiteHeaderShell>
      <Outlet />
      <footer className="site-footer">
        <p>
          Albertson Ward &middot; Ward Mission &middot; <a href="/">Public site</a>
        </p>
      </footer>
    </>
  )
}
