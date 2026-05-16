import {
  createContext,
  useCallback,
  useContext,
  useEffect,
  useMemo,
  useState,
  type ReactNode,
} from 'react'
import { fetchBootstrap, setCsrfToken, type BootstrapUser } from './api'

type Ctx = {
  ready: boolean
  user: BootstrapUser | null
  csrfToken: string
  /** False when server runs with DEV_SKIP_OAUTH (Google sign-in disabled). */
  googleSignInOk: boolean
  refresh: () => Promise<void>
}

const BootstrapContext = createContext<Ctx | null>(null)

export function BootstrapProvider({ children }: { children: ReactNode }) {
  const [ready, setReady] = useState(false)
  const [user, setUser] = useState<BootstrapUser | null>(null)
  const [csrfToken, setTok] = useState('')

  const [googleSignInOk, setGoogleSignInOk] = useState(true)

  const refresh = useCallback(async () => {
    const d = await fetchBootstrap()
    setTok(d.csrfToken)
    setCsrfToken(d.csrfToken)
    setUser(d.user)
    setGoogleSignInOk(d.googleSignInOk)
    setReady(true)
  }, [])

  useEffect(() => {
    refresh().catch(() => setReady(true))
  }, [refresh])

  const value = useMemo(
    () => ({ ready, user, csrfToken, googleSignInOk, refresh }),
    [ready, user, csrfToken, googleSignInOk, refresh],
  )

  return <BootstrapContext.Provider value={value}>{children}</BootstrapContext.Provider>
}

export function useBootstrap() {
  const c = useContext(BootstrapContext)
  if (!c) throw new Error('useBootstrap outside provider')
  return c
}
