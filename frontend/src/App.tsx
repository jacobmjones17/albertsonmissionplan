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
import { PrintMonthlyChallenges } from './pages/PrintMonthlyChallenges'
import { WardPlan } from './pages/WardPlan'

const SITE_TITLE = 'Albertson Ward Plan'

function SiteTitle() {
  const loc = useLocation()
  useEffect(() => {
    document.title = SITE_TITLE
  }, [loc.pathname])
  return null
}

function RequireLeader() {
  const { ready, user } = useBootstrap()
  if (!ready) {
    return (
      <main className="wrap">
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
      <Route element={<Layout />}>
        <Route path="/" element={<Home />} />
        <Route path="/ward-plan" element={<WardPlan />} />
        <Route path="/monthly-challenges" element={<MonthlyChallenges />} />
        <Route path="/experiences" element={<Experiences />} />
        <Route path="/love-share-invite" element={<LoveShareInvite />} />
      </Route>
      <Route path="/admin" element={<AdminLayout />}>
        <Route index element={<AdminHome />} />
        <Route path="forgot-password" element={<AdminForgotPassword />} />
        <Route path="reset-password" element={<AdminResetPassword />} />
        <Route element={<RequireLeader />}>
          <Route path="moderate" element={<AdminModerate />} />
          <Route path="approvals" element={<AdminApprovals />} />
          <Route path="ward-plan" element={<AdminWardPlan />} />
        </Route>
      </Route>
    </Routes>
    </>
  )
}
