import {
  CHALLENGE_CARD_HEADING,
  MONTHLY_GOALS,
  MONTHS_SHORT,
  type Goal,
} from '../monthlyData'
import { MONTHLY_CHALLENGE_IMAGE_BY_GOAL_ID } from '../monthlyChallengeImages'
import { CHALLENGES_PRINT_HINT, CHALLENGES_SHEET_STYLES } from './challengesSheetStyles'
import { challengeBulletHtml } from './challengeBullets'
import { escapeHtml, wrapPrintHtml } from './downloadDocument'
import { wardGoalLineHtml } from './wardGoalHtml'
import { WARD_PLAN_PRINT_HINT, WARD_PLAN_STYLES } from './wardPlanDocumentStyles'

export { CHALLENGES_SHEET_STYLES }

export function buildMonthlyChallengesHtml(
  year: number,
  imageSrcByGoalId: Partial<Record<Goal['id'], string>> = {},
): string {
  const cells = MONTHLY_GOALS.map((goal, index) => {
    const bullets = goal.bullets.map((b) => challengeBulletHtml(goal, b.who, b.text)).join('\n')
    const id = goal.id as Goal['id']
    const imgSrc = imageSrcByGoalId[id] ?? MONTHLY_CHALLENGE_IMAGE_BY_GOAL_ID[id]
    const img = escapeHtml(imgSrc)
    const month = escapeHtml(MONTHS_SHORT[index]!)
    const name = escapeHtml(CHALLENGE_CARD_HEADING[goal.id as Goal['id']])
    return `<article class="challenges-sheet__cell">
  <div class="challenges-sheet__card">
    <div class="challenges-sheet__art">
      <img class="challenges-sheet__img" src="${img}" alt="${name}" />
      <span class="challenges-sheet__month">${month}</span>
    </div>
    <div class="challenges-sheet__body">
      <p class="challenges-sheet__goal"><strong>Goal:</strong> ${escapeHtml(goal.title)}</p>
      <ul class="challenges-sheet__list">${bullets}</ul>
    </div>
  </div>
</article>`
  }).join('\n')

  const body = `<p class="sheet-print-hint">${CHALLENGES_PRINT_HINT}</p>
<div class="challenges-sheet">
  <header class="challenges-sheet__titlebar">
    <h1 class="challenges-sheet__title">Albertson Ward · Monthly Challenges · ${year}</h1>
  </header>
  <div class="challenges-sheet__grid">
${cells}
  </div>
</div>`

  return wrapPrintHtml(
    `Albertson Ward Monthly Challenges ${year}`,
    body,
    CHALLENGES_SHEET_STYLES,
    true,
  )
}

export function buildWardPlanHtml(wardGoals: string[]): string {
  const goalsList = wardGoals
    .map((g) => `<li><div class="ward-rich-line">${wardGoalLineHtml(g)}</div></li>`)
    .join('\n')

  const body = `<p class="sheet-print-hint">${WARD_PLAN_PRINT_HINT}</p>
<div class="ward-goals-sheet">
<header class="ward-goals-sheet__header">
  <p class="ward-goals-sheet__eyebrow">Albertson Ward</p>
  <h1 class="ward-goals-sheet__title">Ward Goals</h1>
  <p class="ward-goals-sheet__lead"><strong>We will strengthen our ward by:</strong></p>
</header>
<ul class="ward-goals-sheet__list">${goalsList}</ul>
<blockquote class="ward-goals-sheet__quote">Matthew 28:19 — &ldquo;Go ye therefore, and teach all nations, baptizing them in the name of the Father, and of the Son, and of the Holy Ghost.&rdquo;</blockquote>
</div>`

  return wrapPrintHtml('Albertson Ward Goals', body, WARD_PLAN_STYLES, true)
}
