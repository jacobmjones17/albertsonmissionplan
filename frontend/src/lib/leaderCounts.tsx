import {
  createContext,
  useCallback,
  useContext,
  useEffect,
  useMemo,
  useState,
  type ReactNode,
} from 'react'
import { useLocation } from 'react-router-dom'
import { apiJson } from '../api'

export type LeaderCounts = {
  /** Mission experiences waiting for review (GET /api/admin/pending). */
  pendingExperiences: number | null
  /** Leader account requests waiting for approval (GET /api/admin/pending-accounts). */
  pendingAccounts: number | null
  error: string | null
  refresh: () => Promise<void>
}

const noop = async () => {}

const LeaderCountsContext = createContext<LeaderCounts>({
  pendingExperiences: null,
  pendingAccounts: null,
  error: null,
  refresh: noop,
})

type PendingList = { pending: unknown[] | null }

/**
 * One source for the "waiting for you" counts shown on the leader nav and the /admin cards.
 * Refetches on route change and when the tab becomes visible again (phones), and pages call
 * `refresh()` after approving / rejecting so the badges update right away.
 */
export function LeaderCountsProvider({ enabled, children }: { enabled: boolean; children: ReactNode }) {
  const [pendingExperiences, setPendingExperiences] = useState<number | null>(null)
  const [pendingAccounts, setPendingAccounts] = useState<number | null>(null)
  const [error, setError] = useState<string | null>(null)
  const loc = useLocation()

  const refresh = useCallback(async () => {
    if (!enabled) return
    const [exp, acct] = await Promise.allSettled([
      apiJson<PendingList>('/api/admin/pending'),
      apiJson<PendingList>('/api/admin/pending-accounts'),
    ])
    const errs: string[] = []
    if (exp.status === 'fulfilled') {
      setPendingExperiences(Array.isArray(exp.value.pending) ? exp.value.pending.length : 0)
    } else {
      errs.push('Could not load how many experiences are waiting.')
    }
    if (acct.status === 'fulfilled') {
      setPendingAccounts(Array.isArray(acct.value.pending) ? acct.value.pending.length : 0)
    } else {
      errs.push('Could not load how many leader requests are waiting.')
    }
    setError(errs.length ? errs.join(' ') : null)
  }, [enabled])

  useEffect(() => {
    if (!enabled) {
      setPendingExperiences(null)
      setPendingAccounts(null)
      setError(null)
      return
    }
    void refresh()
  }, [enabled, refresh, loc.pathname])

  useEffect(() => {
    if (!enabled) return
    const onVis = () => {
      if (document.visibilityState === 'visible') void refresh()
    }
    document.addEventListener('visibilitychange', onVis)
    return () => document.removeEventListener('visibilitychange', onVis)
  }, [enabled, refresh])

  const value = useMemo(
    () => ({ pendingExperiences, pendingAccounts, error, refresh }),
    [pendingExperiences, pendingAccounts, error, refresh],
  )
  return <LeaderCountsContext.Provider value={value}>{children}</LeaderCountsContext.Provider>
}

export function useLeaderCounts(): LeaderCounts {
  return useContext(LeaderCountsContext)
}
