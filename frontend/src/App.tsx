import { Navigate, Outlet, Route, Routes } from 'react-router-dom'
import { useBootstrap } from './BootstrapContext'
import { AdminLayout, Layout } from './Layout'
import { AdminHome } from './pages/AdminHome'
import { AdminModerate } from './pages/AdminModerate'
import { AdminWardPlan } from './pages/AdminWardPlan'
import { Experiences } from './pages/Experiences'
import { Home } from './pages/Home'
import { LoveShareInvite } from './pages/LoveShareInvite'
import { MonthlyChallenges } from './pages/MonthlyChallenges'
import { WardPlan } from './pages/WardPlan'

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
    <Routes>
      <Route element={<Layout />}>
        <Route path="/" element={<Home />} />
        <Route path="/ward-plan" element={<WardPlan />} />
        <Route path="/monthly-challenges" element={<MonthlyChallenges />} />
        <Route path="/experiences" element={<Experiences />} />
        <Route path="/love-share-invite" element={<LoveShareInvite />} />
      </Route>
      <Route path="/admin" element={<AdminLayout />}>
        <Route index element={<AdminHome />} />
        <Route element={<RequireLeader />}>
          <Route path="moderate" element={<AdminModerate />} />
          <Route path="ward-plan" element={<AdminWardPlan />} />
        </Route>
      </Route>
    </Routes>
  )
}
