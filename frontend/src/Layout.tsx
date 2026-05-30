import { useEffect, useRef, useState, type ReactNode } from 'react'
import { Link, NavLink, Outlet, useLocation, type Location } from 'react-router-dom'
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

/** Plain footer notice — not an official Church site (no boxed panel; contact keeps that treatment). */
function SiteDisclaimer() {
  return (
    <p className="site-footer__disclaimer" role="note">
      This website is not an official website of The Church of Jesus Christ of Latter-day Saints.
      It is maintained by Albertson Ward members for local ward mission planning and does not
      represent official Church doctrine, policy, or communication.
    </p>
  )
}

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
  const navRef = useRef<HTMLElement>(null)
  const navToggleRef = useRef<HTMLButtonElement>(null)
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

  useEffect(() => {
    if (!open) return

    function closeIfOutside(event: PointerEvent) {
      const target = event.target
      if (!(target instanceof Node)) return
      if (navRef.current?.contains(target) || navToggleRef.current?.contains(target)) return
      setOpen(false)
    }

    function closeOnEscape(event: KeyboardEvent) {
      if (event.key === 'Escape') setOpen(false)
    }

    document.addEventListener('pointerdown', closeIfOutside)
    document.addEventListener('keydown', closeOnEscape)
    return () => {
      document.removeEventListener('pointerdown', closeIfOutside)
      document.removeEventListener('keydown', closeOnEscape)
    }
  }, [open])

  /* Mobile: scrolling the page closes the open menu so content is not locked behind it. */
  useEffect(() => {
    if (!open) return
    const mobileNav = window.matchMedia('(max-width: 768px)')
    if (!mobileNav.matches) return

    function closeOnPageScroll() {
      setOpen(false)
    }

    function closeOnPageTouchScroll(event: TouchEvent) {
      const target = event.target
      if (!(target instanceof Node)) return
      if (navRef.current?.contains(target) || navToggleRef.current?.contains(target)) return
      setOpen(false)
    }

    window.addEventListener('scroll', closeOnPageScroll, { passive: true, capture: true })
    document.addEventListener('touchmove', closeOnPageTouchScroll, { passive: true })
    return () => {
      window.removeEventListener('scroll', closeOnPageScroll, { capture: true })
      document.removeEventListener('touchmove', closeOnPageTouchScroll)
    }
  }, [open])

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
            ref={navToggleRef}
            type="button"
            className="nav-toggle"
            aria-expanded={open}
            aria-controls="site-nav"
            aria-label={open ? 'Close menu' : 'Open menu'}
            onClick={() => setOpen((o) => !o)}
          >
            <span aria-hidden="true" />
          </button>
          <nav
            ref={navRef}
            id="site-nav"
            className={`site-nav${open ? ' is-open' : ''}`}
            aria-label="Primary"
          >
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
          <SiteDisclaimer />
        </div>
      </footer>
    </>
  )
}

function AdminBackArrowIcon() {
  return (
    <svg
      className="admin-back-link__icon"
      viewBox="0 0 24 24"
      fill="currentColor"
      aria-hidden="true"
      focusable="false"
    >
      <path d="M20 11H7.83l5.59-5.59L12 4l-8 8 8 8 1.41-1.41L7.83 13H20v-2z" />
    </svg>
  )
}

function adminShowsBackToTools({ pathname }: Location): boolean {
  return (
    pathname === '/admin/moderate' ||
    pathname === '/admin/approvals' ||
    pathname === '/admin/ward-plan'
  )
}

export function AdminLayout() {
  const { refresh, user } = useBootstrap()
  const loc = useLocation()
  const showBackToTools = adminShowsBackToTools(loc)

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
          <div className="admin-header-actions">
            <Link to="/" className="btn btn-secondary">
              Public site
            </Link>
            {user?.isLeader ? (
              <button type="button" className="btn btn-secondary" onClick={() => void signOut()}>
                Sign out
              </button>
            ) : null}
          </div>
        </div>
      </SiteHeaderShell>
      <div className="admin-shell admin-body">
        {showBackToTools ? (
          <div className="wrap admin-crumb-wrap">
            <p className="admin-crumb">
              <Link to="/admin" className="admin-back-link">
                <span className="admin-back-link__arrow">
                  <AdminBackArrowIcon />
                </span>
                Leader tools
              </Link>
            </p>
          </div>
        ) : null}
        <Outlet />
      </div>
      <footer className="site-footer">
        <div className="site-footer__inner">
          <p className="site-footer__meta">
            Albertson Ward &middot; Ward Mission &middot; <Link to="/">Public site</Link>
          </p>
          <SiteDisclaimer />
        </div>
      </footer>
    </>
  )
}
