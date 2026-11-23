import { useEffect, useRef, useState, type ReactNode } from 'react'
import { Link, NavLink, Outlet, useLocation } from 'react-router-dom'
import { apiJson } from './api'
import { LeaderCountsProvider, useLeaderCounts } from './lib/leaderCounts'
import { useBootstrap } from './BootstrapContext'
import { WARD_FACEBOOK_URL, WARD_INSTAGRAM_URL } from './churchLinks'
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

function LeaderKeyIcon() {
  return (
    <svg className="nav-leader-btn__icon" viewBox="0 0 24 24" aria-hidden="true" focusable="false">
      <path
        fill="currentColor"
        d="M12 12a4.5 4.5 0 1 0 0-9 4.5 4.5 0 0 0 0 9Zm0 2c-3.9 0-8 1.95-8 4.75V21h16v-2.25C20 15.95 15.9 14 12 14Z"
      />
    </svg>
  )
}

function FooterFacebookIcon() {
  return (
    <svg className="site-footer__social-icon" viewBox="0 0 24 24" aria-hidden="true">
      <path
        fill="currentColor"
        d="M24 12.073C24 5.405 18.627 0 12 0S0 5.405 0 12.073c0 6.027 4.388 11.023 10.125 11.926v-8.43H7.078v-3.496h3.047V9.43c0-3.007 1.792-4.669 4.533-4.669 1.312 0 2.686.234 2.686.234v2.953H15.83c-1.491 0-1.956.925-1.956 1.874v2.25h3.328l-.532 3.496h-2.796v8.43C19.612 23.096 24 18.1 24 12.073z"
      />
    </svg>
  )
}

function FooterInstagramIcon() {
  return (
    <svg className="site-footer__social-icon" viewBox="0 0 24 24" aria-hidden="true">
      <path
        fill="currentColor"
        d="M12 2.163c3.204 0 3.584.012 4.85.07 1.17.054 1.97.24 2.427.403a4.92 4.92 0 0 1 1.77 1.153 4.92 4.92 0 0 1 1.153 1.77c.163.457.349 1.257.403 2.427.058 1.266.07 1.646.07 4.85s-.012 3.584-.07 4.85c-.054 1.17-.24 1.97-.403 2.427a4.92 4.92 0 0 1-1.153 1.77 4.92 4.92 0 0 1-1.77 1.153c-.457.163-1.257.349-2.427.403-1.266.058-1.646.07-4.85.07s-3.584-.012-4.85-.07c-1.17-.054-1.97-.24-2.427-.403a4.92 4.92 0 0 1-1.77-1.153 4.92 4.92 0 0 1-1.153-1.77c-.163-.457-.349-1.257-.403-2.427C2.175 15.747 2.163 15.367 2.163 12s.012-3.584.07-4.85c.054-1.17.24-1.97.403-2.427a4.92 4.92 0 0 1 1.153-1.77 4.92 4.92 0 0 1 1.77-1.153c.457-.163 1.257-.349 2.427-.403C8.416 2.175 8.796 2.163 12 2.163zm0-2.163C8.741 0 8.332.014 7.052.072 5.775.13 4.902.333 4.14.63a6.96 6.96 0 0 0-2.51 1.64A6.96 6.96 0 0 0 .63 4.14C.333 4.902.13 5.775.072 7.052.014 8.332 0 8.741 0 12s.014 3.668.072 4.948c.058 1.277.261 2.15.558 2.912a6.96 6.96 0 0 0 1.64 2.51 6.96 6.96 0 0 0 2.51 1.64c.762.297 1.635.5 2.912.558C8.332 23.986 8.741 24 12 24s3.668-.014 4.948-.072c1.277-.058 2.15-.261 2.912-.558a6.96 6.96 0 0 0 2.51-1.64 6.96 6.96 0 0 0 1.64-2.51c.297-.762.5-1.635.558-2.912.058-1.28.072-1.689.072-4.948s-.014-3.668-.072-4.948c-.058-1.277-.261-2.15-.558-2.912a6.96 6.96 0 0 0-1.64-2.51 6.96 6.96 0 0 0-2.51-1.64c-.762-.297-1.635-.5-2.912-.558C15.668.014 15.259 0 12 0zm0 5.838a6.162 6.162 0 1 0 0 12.324 6.162 6.162 0 0 0 0-12.324zM12 16a4 4 0 1 1 0-8 4 4 0 0 1 0 8zm6.406-11.845a1.44 1.44 0 1 0 0 2.881 1.44 1.44 0 0 0 0-2.881z"
      />
    </svg>
  )
}

function SiteFooterSocial() {
  return (
    <div className="site-footer__social">
      <p className="site-footer__social-label">Christ First Social Media</p>
      <div className="site-footer__social-actions">
        <a
          className="site-footer__social-btn site-footer__social-btn--facebook"
          href={WARD_FACEBOOK_URL}
          target="_blank"
          rel="noopener noreferrer"
          aria-label="Albertson Ward on Facebook"
        >
          <FooterFacebookIcon />
        </a>
        <a
          className="site-footer__social-btn site-footer__social-btn--instagram"
          href={WARD_INSTAGRAM_URL}
          target="_blank"
          rel="noopener noreferrer"
          aria-label="Albertson Ward on Instagram"
        >
          <FooterInstagramIcon />
        </a>
      </div>
    </div>
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
              <li className="site-nav__leader">
                <NavLink
                  to="/admin"
                  className="nav-leader-btn"
                  title={user?.isLeader ? 'Open leader tools' : 'Sign in for ward leaders'}
                >
                  <LeaderKeyIcon />
                  {user?.isLeader ? 'Leader Tools' : 'Leader Sign-In'}
                </NavLink>
              </li>
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
              {user?.isLeader ? 'Leader Tools' : 'Leader Sign-In'}
            </Link>
          </p>
          <div className="site-footer__contact">
            <p className="site-footer__contact-label">Contact</p>
            <p className="site-footer__contact-body">
              To request changes or share concerns about this site, contact the ward mission leaders:{' '}
              <strong>Brother Jacob Jones</strong> or <strong>Brother Michael Hallenberger</strong>.
            </p>
          </div>
          <SiteFooterSocial />
          <p className="attribution">
            Built for ward members. Update content as goals change each year.
          </p>
          <SiteDisclaimer />
        </div>
      </footer>
    </>
  )
}

function LeaderNavBadge({ count, label }: { count: number | null; label: string }) {
  if (!count) return null
  return (
    <span className="leader-nav__badge" aria-label={`${count} ${label}`}>
      {count}
    </span>
  )
}

/** Persistent leader navigation (all /admin pages once signed in). Scrolls sideways on phones. */
function LeaderNav({ onSignOut }: { onSignOut: () => void }) {
  const { pendingExperiences, pendingAccounts } = useLeaderCounts()
  return (
    <nav className="leader-nav" aria-label="Leader tools">
      <div className="leader-nav__inner">
        <ul className="leader-nav__list">
          <li>
            <NavLink to="/admin" end>
              Home
            </NavLink>
          </li>
          <li>
            <NavLink to="/admin/moderate">
              Experiences
              <LeaderNavBadge count={pendingExperiences} label="waiting for review" />
            </NavLink>
          </li>
          <li>
            <NavLink to="/admin/approvals">
              Leader Accounts
              <LeaderNavBadge count={pendingAccounts} label="waiting for approval" />
            </NavLink>
          </li>
          <li>
            <NavLink to="/admin/ward-plan">Ward Plan</NavLink>
          </li>
        </ul>
        <ul className="leader-nav__list leader-nav__list--end">
          <li>
            <Link to="/">Public Site</Link>
          </li>
          <li>
            <button type="button" className="leader-nav__signout" onClick={onSignOut}>
              Sign Out
            </button>
          </li>
        </ul>
      </div>
    </nav>
  )
}

export function AdminLayout() {
  const { refresh, user } = useBootstrap()
  const isLeader = Boolean(user?.isLeader)

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
          {isLeader ? null : (
            <div className="admin-header-actions">
              <Link to="/" className="btn btn-secondary">
                Public Site
              </Link>
            </div>
          )}
        </div>
      </SiteHeaderShell>
      <LeaderCountsProvider enabled={isLeader}>
        {isLeader ? <LeaderNav onSignOut={() => void signOut()} /> : null}
        <div className="admin-shell admin-body">
          <Outlet />
        </div>
      </LeaderCountsProvider>
      <footer className="site-footer">
        <div className="site-footer__inner">
          <p className="site-footer__meta">
            Albertson Ward &middot; Ward Mission &middot; <Link to="/">Public Site</Link>
          </p>
          <SiteFooterSocial />
          <SiteDisclaimer />
        </div>
      </footer>
    </>
  )
}
