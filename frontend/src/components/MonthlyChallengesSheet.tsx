import { MONTHS_SHORT } from '../monthlyData'
import { renderChallengeBullet } from '../lib/challengeBullets'
import type { MonthlyChallenge } from '../lib/monthlyChallenges'

/** Shared one-page monthly challenges grid (print + download). */
export function MonthlyChallengesSheet({
  year,
  challenges,
  className = '',
}: {
  year: number
  challenges: MonthlyChallenge[]
  className?: string
}) {
  return (
    <div className={`challenges-sheet ${className}`.trim()}>
      <header className="challenges-sheet__titlebar">
        <h1 className="challenges-sheet__title">
          Albertson Ward · Monthly Challenges · {year}
        </h1>
      </header>

      <div className="challenges-sheet__grid">
        {challenges.map((c) => (
          <article className="challenges-sheet__cell" key={c.month}>
            <div className="challenges-sheet__card">
              <div className="challenges-sheet__art">
                <img src={c.imageSrc} alt={c.heading} className="challenges-sheet__img" />
                <span className="challenges-sheet__month">{MONTHS_SHORT[c.month - 1]}</span>
              </div>

              <div className="challenges-sheet__body">
                <p className="challenges-sheet__goal">
                  <strong>Goal:</strong> {c.goal}
                </p>
                <ul className="challenges-sheet__list">
                  {c.bullets.map((b, i) => renderChallengeBullet(b.who, b.text, i))}
                </ul>
              </div>
            </div>
          </article>
        ))}
      </div>
    </div>
  )
}
