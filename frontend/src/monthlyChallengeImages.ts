import type { Goal } from './monthlyData'

/** Art for each monthly goal (`MONTHLY_GOALS` Jan→Dec); files live in `public/site/monthly/`. */
export const MONTHLY_CHALLENGE_IMAGE_BY_GOAL_ID: Record<Goal['id'], string> = {
  g1: '/site/monthly/scripture-study.png',
  g2: '/site/monthly/act-of-kindness.png',
  g3: '/site/monthly/temple-family-history.png',
  g4: '/site/monthly/missionary-work.png',
  g5: '/site/monthly/health-wellness.png',
  g6: '/site/monthly/faith-testimony.png',
  g7: '/site/monthly/gratitude.png',
  g8: '/site/monthly/prayer.png',
  g9: '/site/monthly/sabbath-day.png',
  g10: '/site/monthly/conference.png',
  g11: '/site/monthly/service.png',
  g12: '/site/monthly/christmas.png',
}

export function monthlyChallengeImageSrc(goalId: Goal['id']): string {
  return MONTHLY_CHALLENGE_IMAGE_BY_GOAL_ID[goalId]
}
