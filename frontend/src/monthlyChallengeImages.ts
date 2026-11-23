import type { Goal } from './monthlyData'

/**
 * Art for each monthly goal (`MONTHLY_GOALS` Jan→Dec). Files live in `public/site/monthly/`
 * as compressed WebP in two widths: `<slug>-480.webp` (phones, print sheet, downloads) and
 * `<slug>-960.webp` (2x / wide screens).
 */
const MONTHLY_CHALLENGE_IMAGE_SLUG_BY_GOAL_ID: Record<Goal['id'], string> = {
  g1: 'scripture-study',
  g2: 'act-of-kindness',
  g3: 'temple-family-history',
  g4: 'missionary-work',
  g5: 'health-wellness',
  g6: 'faith-testimony',
  g7: 'gratitude',
  g8: 'prayer',
  g9: 'sabbath-day',
  g10: 'conference',
  g11: 'service',
  g12: 'christmas',
}

/** Intrinsic size of the WebP art (3:2) so the browser reserves space before load. */
export const MONTHLY_CHALLENGE_IMAGE_WIDTH = 960
export const MONTHLY_CHALLENGE_IMAGE_HEIGHT = 640

function monthlyImagePath(goalId: Goal['id'], width: 480 | 960): string {
  return `/site/monthly/${MONTHLY_CHALLENGE_IMAGE_SLUG_BY_GOAL_ID[goalId]}-${width}.webp`
}

/** Default (small) image: print sheet, offline download, and `src` fallback. */
export function monthlyChallengeImageSrc(goalId: Goal['id']): string {
  return monthlyImagePath(goalId, 480)
}

/** Responsive `srcSet` for on-screen cards. */
export function monthlyChallengeImageSrcSet(goalId: Goal['id']): string {
  return `${monthlyImagePath(goalId, 480)} 480w, ${monthlyImagePath(goalId, 960)} 960w`
}

/** Matches the featured challenge card width (full width on phones, ~22rem beside the picker). */
export const MONTHLY_CHALLENGE_IMAGE_SIZES = '(min-width: 900px) 352px, calc(100vw - 2rem)'

const preloaded = new Set<string>()

/**
 * Warm the browser cache for one month's art (on hover/focus of its tile) so switching feels
 * instant without downloading all twelve images up front.
 */
export function preloadMonthlyChallengeImage(goalId: Goal['id']): void {
  const key = goalId
  if (preloaded.has(key)) {
    return
  }
  preloaded.add(key)
  const img = new Image()
  img.sizes = MONTHLY_CHALLENGE_IMAGE_SIZES
  img.srcset = monthlyChallengeImageSrcSet(goalId)
  img.src = monthlyChallengeImageSrc(goalId)
}
