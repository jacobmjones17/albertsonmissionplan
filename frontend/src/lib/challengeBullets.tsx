import type { Goal } from '../monthlyData'
import { escapeHtml } from './downloadDocument'

export function challengeBulletHtml(goal: Goal, who: string, text: string): string {
  if (goal.id === 'g3' && who === 'Individuals') {
    return `<li><strong>Individuals:</strong> Spend time each week on <a href="https://www.familysearch.org/en/" rel="noopener noreferrer">FamilySearch.org</a>.</li>`
  }
  return `<li><strong>${escapeHtml(who)}:</strong> ${escapeHtml(text)}</li>`
}

export function renderChallengeBullet(goal: Goal, who: string, text: string, key: number) {
  if (goal.id === 'g3' && who === 'Individuals') {
    return (
      <li key={key}>
        <strong>Individuals:</strong> Spend time each week on{' '}
        <a href="https://www.familysearch.org/en/" rel="noopener noreferrer">
          FamilySearch.org
        </a>
        .
      </li>
    )
  }
  return (
    <li key={key}>
      <strong>{who}:</strong> {text}
    </li>
  )
}
