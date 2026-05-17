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
  refresh: () => Promise<void>
}

const BootstrapContext = createContext<Ctx | null>(null)

export function BootstrapProvider({ children }: { children: ReactNode }) {
  const [ready, setReady] = useState(false)
  const [user, setUser] = useState<BootstrapUser | null>(null)
  const [csrfToken, setTok] = useState('')

  const refresh = useCallback(async () => {
    const d = await fetchBootstrap()
    setTok(d.csrfToken)
    setCsrfToken(d.csrfToken)
    setUser(d.user)
    setReady(true)
  }, [])

  useEffect(() => {
    refresh().catch(() => setReady(true))
  }, [refresh])

  const value = useMemo(
    () => ({
      ready,
      user,
      csrfToken,
      refresh,
    }),
    [ready, user, csrfToken, refresh],
  )

  return <BootstrapContext.Provider value={value}>{children}</BootstrapContext.Provider>
}

export function useBootstrap() {
  const c = useContext(BootstrapContext)
  if (!c) throw new Error('useBootstrap outside provider')
  return c
}
