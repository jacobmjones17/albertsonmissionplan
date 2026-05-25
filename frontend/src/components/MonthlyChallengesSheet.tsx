import { CHALLENGE_CARD_HEADING, MONTHLY_GOALS, MONTHS_SHORT, type Goal } from '../monthlyData'
import { renderChallengeBullet } from '../lib/challengeBullets'
import { monthlyChallengeImageSrc } from '../monthlyChallengeImages'

/** Shared one-page monthly challenges grid (print + download). */
export function MonthlyChallengesSheet({
  year,
  className = '',
}: {
  year: number
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
        {MONTHLY_GOALS.map((goal, index) => (
          <article className="challenges-sheet__cell" key={goal.id}>
            <div className="challenges-sheet__card">
              <div className="challenges-sheet__art">
                <img
                  src={monthlyChallengeImageSrc(goal.id)}
                  alt={CHALLENGE_CARD_HEADING[goal.id as Goal['id']]}
                  className="challenges-sheet__img"
                />
                <span className="challenges-sheet__month">{MONTHS_SHORT[index]}</span>
              </div>

              <div className="challenges-sheet__body">
                <p className="challenges-sheet__goal">
                  <strong>Goal:</strong> {goal.title}
                </p>
                <ul className="challenges-sheet__list">
                  {goal.bullets.map((b, i) => renderChallengeBullet(goal, b.who, b.text, i))}
                </ul>
              </div>
            </div>
          </article>
        ))}
      </div>
    </div>
  )
}
