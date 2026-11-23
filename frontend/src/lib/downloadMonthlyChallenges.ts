import { fetchImageAsDataUrl } from './fetchImageAsDataUrl'
import { loadMonthlyChallenges, FALLBACK_CHALLENGES, type MonthlyChallenge } from './monthlyChallenges'
import { buildMonthlyChallengesHtml } from './printDocuments'

/** Build a self-contained HTML document with embedded challenge artwork. */
export async function buildOfflineMonthlyChallengesHtml(
  year: number,
  challenges?: MonthlyChallenge[],
): Promise<string> {
  const list = challenges ?? (await loadMonthlyChallenges().catch(() => FALLBACK_CHALLENGES))
  const imageSrcByMonth: Record<number, string> = {}
  await Promise.all(
    list.map(async (c) => {
      imageSrcByMonth[c.month] = await fetchImageAsDataUrl(c.imageSrc)
    }),
  )
  return buildMonthlyChallengesHtml(year, list, imageSrcByMonth)
}
