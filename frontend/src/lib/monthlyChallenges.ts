import { useEffect, useState } from 'react'
import { CHALLENGE_CARD_HEADING, MONTHLY_GOALS } from '../monthlyData'
import { monthlyChallengeImageSrc, monthlyChallengeImageSrcSet } from '../monthlyChallengeImages'

export type ChallengeBullet = { who: string; text: string }

/** One month's challenge as shown on the public page and the printable sheet. */
export type MonthlyChallenge = {
  /** 1 = January … 12 = December. */
  month: number
  /** "Scripture Study Challenge" (month tiles drop the trailing "Challenge"). */
  heading: string
  goal: string
  bullets: ChallengeBullet[]
  /** 480px art: phones, print sheet, offline download, and the `src` fallback. */
  imageSrc: string
  /** Responsive `srcSet` (480w + 960w). */
  imageSrcSet: string
}

/** Shape returned by GET /api/challenges (and, with extra fields, /api/admin/challenges). */
export type ChallengeApiItem = {
  month: number
  heading: string
  goal: string
  bullets: ChallengeBullet[] | null
  imageUrl: string
  thumbUrl: string
}

export function fromApiItem(c: ChallengeApiItem): MonthlyChallenge {
  return {
    month: c.month,
    heading: c.heading,
    goal: c.goal,
    bullets: c.bullets ?? [],
    imageSrc: c.thumbUrl,
    imageSrcSet: `${c.thumbUrl} 480w, ${c.imageUrl} 960w`,
  }
}

/** The original static challenges, used only when the API can't be reached. */
export const FALLBACK_CHALLENGES: MonthlyChallenge[] = MONTHLY_GOALS.map((g, i) => ({
  month: i + 1,
  heading: CHALLENGE_CARD_HEADING[g.id],
  goal: g.title,
  bullets: g.bullets,
  imageSrc: monthlyChallengeImageSrc(g.id),
  imageSrcSet: monthlyChallengeImageSrcSet(g.id),
}))

/** "Scripture Study Challenge" → "Scripture Study" for the month tiles. */
export function shortTheme(heading: string): string {
  return heading.replace(/\s+Challenge$/i, '') || heading
}

export type ChallengesState =
  | { status: 'loading' }
  | { status: 'ready'; challenges: MonthlyChallenge[] }
  /** The API failed; `challenges` is the built-in original content. */
  | { status: 'fallback'; challenges: MonthlyChallenge[] }

let cache: Promise<MonthlyChallenge[]> | null = null

/** Fetch the published challenges once per page load (shared by the page, print view, download). */
export function loadMonthlyChallenges(): Promise<MonthlyChallenge[]> {
  if (!cache) {
    cache = fetch('/api/challenges', { credentials: 'include', headers: { Accept: 'application/json' } })
      .then(async (res) => {
        if (!res.ok) throw new Error(`HTTP ${res.status}`)
        const data = (await res.json()) as { items: ChallengeApiItem[] | null }
        return (data.items ?? []).map(fromApiItem)
      })
      .catch((err: unknown) => {
        cache = null // try again next time
        throw err
      })
  }
  return cache
}

/** Published challenges with a loading state and the original content as an offline fallback. */
export function useMonthlyChallenges(): ChallengesState {
  const [state, setState] = useState<ChallengesState>({ status: 'loading' })
  useEffect(() => {
    let alive = true
    loadMonthlyChallenges()
      .then((challenges) => alive && setState({ status: 'ready', challenges }))
      .catch(() => alive && setState({ status: 'fallback', challenges: FALLBACK_CHALLENGES }))
    return () => {
      alive = false
    }
  }, [])
  return state
}

const preloaded = new Set<string>()

/**
 * Warm the browser cache for one month's art (on hover/focus of its tile) so switching feels
 * instant without downloading all twelve images up front.
 */
export function preloadChallengeImage(c: MonthlyChallenge, sizes: string): void {
  if (preloaded.has(c.imageSrc)) return
  preloaded.add(c.imageSrc)
  const img = new Image()
  img.sizes = sizes
  img.srcset = c.imageSrcSet
  img.src = c.imageSrc
}

/** Forget the cached list (after a leader saves) so the public pages show the new content. */
export function invalidateMonthlyChallenges(): void {
  cache = null
}
