import type { Goal } from '../monthlyData'
import { escapeHtml } from './downloadDocument'

export function challengeBulletHtml(goal: Goal, who: string, text: string): string {
  if (goal.id === 'g3' && who === 'Individuals') {
    return `<li><span class="challenge-list-label">Individuals:</span><span class="challenge-list-text">Spend time each week on <a href="https://www.familysearch.org/en/" rel="noopener noreferrer">FamilySearch.org</a>.</span></li>`
  }
  return `<li><span class="challenge-list-label">${escapeHtml(who)}:</span><span class="challenge-list-text">${escapeHtml(text)}</span></li>`
}

export function renderChallengeBullet(goal: Goal, who: string, text: string, key: number) {
  if (goal.id === 'g3' && who === 'Individuals') {
    return (
      <li key={key}>
        <span className="challenge-list-label">Individuals:</span>
        <span className="challenge-list-text">
          Spend time each week on{' '}
          <a href="https://www.familysearch.org/en/" rel="noopener noreferrer">
            FamilySearch.org
          </a>
          .
        </span>
      </li>
    )
  }
  return (
    <li key={key}>
      <span className="challenge-list-label">{who}:</span>
      <span className="challenge-list-text">{text}</span>
    </li>
  )
}
