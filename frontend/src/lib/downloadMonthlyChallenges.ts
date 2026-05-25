import type { Goal } from '../monthlyData'
import { MONTHLY_GOALS } from '../monthlyData'
import { MONTHLY_CHALLENGE_IMAGE_BY_GOAL_ID } from '../monthlyChallengeImages'
import { fetchImageAsDataUrl } from './fetchImageAsDataUrl'
import { buildMonthlyChallengesHtml } from './printDocuments'

/** Build a self-contained HTML document with embedded challenge artwork. */
export async function buildOfflineMonthlyChallengesHtml(year: number): Promise<string> {
  const imageSrcByGoalId: Partial<Record<Goal['id'], string>> = {}

  await Promise.all(
    MONTHLY_GOALS.map(async (goal) => {
      const id = goal.id as Goal['id']
      const path = MONTHLY_CHALLENGE_IMAGE_BY_GOAL_ID[id]
      imageSrcByGoalId[id] = await fetchImageAsDataUrl(path)
    }),
  )

  return buildMonthlyChallengesHtml(year, imageSrcByGoalId)
}
