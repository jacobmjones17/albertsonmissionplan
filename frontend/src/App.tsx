import { useEffect } from 'react'
import { Navigate, Outlet, Route, Routes, useLocation } from 'react-router-dom'
import { useBootstrap } from './BootstrapContext'
import { AdminLayout, Layout } from './Layout'
import { AdminApprovals } from './pages/AdminApprovals'
import { AdminForgotPassword } from './pages/AdminForgotPassword'
import { AdminHome } from './pages/AdminHome'
import { AdminModerate } from './pages/AdminModerate'
import { AdminResetPassword } from './pages/AdminResetPassword'
import { AdminWardPlan } from './pages/AdminWardPlan'
import { Experiences } from './pages/Experiences'
import { Home } from './pages/Home'
import { LoveShareInvite } from './pages/LoveShareInvite'
import { MonthlyChallenges } from './pages/MonthlyChallenges'
import { NotFound } from './pages/NotFound'
import { PrintMonthlyChallenges } from './pages/PrintMonthlyChallenges'
import { PrintWardPlan } from './pages/PrintWardPlan'
import { WardPlan } from './pages/WardPlan'
import { Gallery } from './pages/Gallery'
import { AdminGallery } from './pages/AdminGallery'
import { Events } from './pages/Events'
import { AdminEvents } from './pages/AdminEvents'
import { AdminChallenges } from './pages/AdminChallenges'

const SITE_TITLE = 'Albertson Ward Plan'

/**
 * Browser-tab title per route ("Monthly Challenges | Albertson Ward Plan"). Keep in sync with
 * the routes below and with `clientRoutes` in backend/internal/spa/spa.go (unknown paths get a
 * 404 status from the server and the NotFound page here).
 */
const PAGE_TITLES: Record<string, string> = {
  '/': '',
  '/ward-plan': 'Ward Plan',
  '/ward-plan/print': 'Printable Ward Goals',
  '/monthly-challenges': 'Monthly Challenges',
  '/monthly-challenges/print': 'Printable Monthly Challenges',
  '/experiences': 'Mission Experiences',
  '/love-share-invite': 'Love, Share, Invite',
  '/gallery': 'Ward Gallery',
  '/events': 'Ward Events',
  '/admin': 'Leader Tools',
  '/admin/forgot-password': 'Forgot Password',
  '/admin/reset-password': 'Reset Password',
  '/admin/moderate': 'Moderate Experiences',
  '/admin/approvals': 'Leader Approvals',
  '/admin/ward-plan': 'Edit Ward Plan',
  '/admin/gallery': 'Manage Ward Gallery',
  '/admin/events': 'Manage Ward Events',
  '/admin/challenges': 'Edit Monthly Challenges',
}

function pageTitleFor(pathname: string): string {
  const key = pathname.toLowerCase().replace(/\/+$/, '') || '/'
  const page = key in PAGE_TITLES ? PAGE_TITLES[key] : 'Page Not Found'
  return page ? `${page} | ${SITE_TITLE}` : SITE_TITLE
}

function SiteTitle() {
  const loc = useLocation()
  useEffect(() => {
    document.title = pageTitleFor(loc.pathname)
  }, [loc.pathname])
  return null
}

function RequireLeader() {
  const { ready, user } = useBootstrap()
  if (!ready) {
    return (
      <main id="main" className="wrap">
        <p className="lead">Loading…</p>
      </main>
    )
  }
  if (!user?.isLeader) {
    return <Navigate to="/admin" replace />
  }
  return <Outlet />
}

export function App() {
  return (
    <>
      <SiteTitle />
      <Routes>
        <Route path="/monthly-challenges/print" element={<PrintMonthlyChallenges />} />
        <Route path="/ward-plan/print" element={<PrintWardPlan />} />
      <Route element={<Layout />}>
        <Route path="/" element={<Home />} />
        <Route path="/ward-plan" element={<WardPlan />} />
        <Route path="/monthly-challenges" element={<MonthlyChallenges />} />
        <Route path="/experiences" element={<Experiences />} />
        <Route path="/love-share-invite" element={<LoveShareInvite />} />
        <Route path="/gallery" element={<Gallery />} />
        <Route path="/events" element={<Events />} />
        <Route path="*" element={<NotFound />} />
      </Route>
      <Route path="/admin" element={<AdminLayout />}>
        <Route index element={<AdminHome />} />
        <Route path="forgot-password" element={<AdminForgotPassword />} />
        <Route path="reset-password" element={<AdminResetPassword />} />
        <Route element={<RequireLeader />}>
          <Route path="moderate" element={<AdminModerate />} />
          <Route path="approvals" element={<AdminApprovals />} />
          <Route path="ward-plan" element={<AdminWardPlan />} />
          <Route path="gallery" element={<AdminGallery />} />
          <Route path="events" element={<AdminEvents />} />
          <Route path="challenges" element={<AdminChallenges />} />
        </Route>
      </Route>
    </Routes>
    </>
  )
}
