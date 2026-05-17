/** CSRF token from /api/bootstrap; sent on mutating requests. */
let csrfToken: string | null = null

export function setCsrfToken(t: string) {
  csrfToken = t
}

export type BootstrapUser = {
  email: string
  /** Friendly "First Last" name from registration, falls back to the email local part. */
  displayName?: string
  isLeader: boolean
  /** Full-access leader (moderation + ward-wide goals + every org tab). Always true today for any signed-in leader. */
  isAdmin?: boolean
  canModerate?: boolean
  canEditWardGoals?: boolean
}

/** True when signed in and may open the ward plan editor. Every leader is a full admin in this build. */
export function hasWardPlanEditorAccess(user: BootstrapUser | null | undefined): boolean {
  return Boolean(user?.isLeader)
}

/** True when the signed-in leader may approve/reject mission experiences. */
export function canModerateMissionExperiences(user: BootstrapUser | null | undefined): boolean {
  return Boolean(user?.isLeader)
}

export async function fetchBootstrap(): Promise<{
  csrfToken: string
  user: BootstrapUser | null
}> {
  const res = await fetch('/api/bootstrap', { credentials: 'include' })
  if (!res.ok) throw new Error('bootstrap failed')
  const data = (await res.json()) as {
    csrfToken: string
    user: BootstrapUser | null
  }
  setCsrfToken(data.csrfToken)
  return {
    csrfToken: data.csrfToken,
    user: data.user,
  }
}

/** multipart POST — do not set Content-Type (boundary is set automatically). */
export async function apiForm<T>(path: string, form: FormData): Promise<T> {
  const headers = new Headers()
  headers.set('Accept', 'application/json')
  if (csrfToken) headers.set('X-CSRF-Token', csrfToken)
  const res = await fetch(path, {
    method: 'POST',
    credentials: 'include',
    body: form,
    headers,
  })
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
