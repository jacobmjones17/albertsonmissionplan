/** CSRF token from /api/bootstrap; sent on mutating requests. */
let csrfToken: string | null = null

export function setCsrfToken(t: string) {
  csrfToken = t
}

export type BootstrapUser = { email: string; isLeader: boolean }

export async function fetchBootstrap(): Promise<{
  csrfToken: string
  user: BootstrapUser | null
  googleSignInOk: boolean
}> {
  const res = await fetch('/api/bootstrap', { credentials: 'include' })
  if (!res.ok) throw new Error('bootstrap failed')
  const data = (await res.json()) as {
    csrfToken: string
    user: BootstrapUser | null
    googleSignInOk?: boolean
  }
  setCsrfToken(data.csrfToken)
  return {
    csrfToken: data.csrfToken,
    user: data.user,
    googleSignInOk: data.googleSignInOk !== false,
  }
}

export async function apiJson<T>(
  path: string,
  init: RequestInit & { json?: unknown } = {},
): Promise<T> {
  const headers = new Headers(init.headers)
  headers.set('Accept', 'application/json')
  if (csrfToken) headers.set('X-CSRF-Token', csrfToken)
  let body = init.body
  if (init.json !== undefined) {
    headers.set('Content-Type', 'application/json')
    body = JSON.stringify(init.json)
  }
  const res = await fetch(path, { ...init, headers, body, credentials: 'include' })
  const text = await res.text()
  let parsed: unknown = null
  if (text) {
    try {
      parsed = JSON.parse(text) as unknown
    } catch {
      parsed = { raw: text }
    }
  }
  if (!res.ok) {
    const err = parsed as { error?: string }
    throw new Error(err?.error || res.statusText || `HTTP ${res.status}`)
  }
  return parsed as T
}
