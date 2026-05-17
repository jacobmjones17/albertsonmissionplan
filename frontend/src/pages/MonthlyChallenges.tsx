import { useCallback, useMemo, useState } from 'react'
import { MONTHLY_GOALS, MONTHS_LONG, MONTHS_SHORT, type Goal } from '../monthlyData'
import { PageHero } from '../components/PageHero'
import { monthlyChallengeImageSrc } from '../monthlyChallengeImages'

/** Display titles aligned with the printed challenge cards. */
const CHALLENGE_CARD_HEADING: Record<Goal['id'], string> = {
  g1: 'Scripture Study Challenge',
  g2: 'Act of Kindness Challenge',
  g3: 'Temple & Family History Challenge',
  g4: 'Missionary Work Challenge',
  g5: 'Health & Wellness Challenge',
  g6: 'Faith & Testimony Challenge',
  g7: 'Gratitude Challenge',
  g8: 'Prayer Challenge',
  g9: 'Sabbath Day Challenge',
  g10: 'General Conference Challenge',
  g11: 'Service Challenge',
  g12: 'Christmas Challenge',
}

function renderBullet(goal: Goal, who: string, text: string, key: number) {
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

export function MonthlyChallenges() {
  const today = useMemo(() => new Date(), [])
  const year = today.getFullYear()
  const currentMonth = today.getMonth()

  const [selectedMonth, setSelectedMonth] = useState(currentMonth)
  const goal = MONTHLY_GOALS[selectedMonth]!

  const pickMonth = useCallback((index: number) => {
    setSelectedMonth(index)
  }, [])

  return (
    <>
      <PageHero
        titleId="page-title"
        eyebrow="Year-round focus"
        title="Monthly Challenges"
        lede="One themed challenge each month for individuals, couples, and families"
      />
      <main id="main" className="challenges-main challenges-main--cards">
        <div className="wrap-wide challenges-wrap">
          <p className="challenges-intro">
            Choose a month below—the challenge card updates, and stays beside the picker on wider
            screens so you rarely need to scroll.
          </p>

          <div className="challenges-split challenges-split--explorer">
            <div className="challenges-explorer-nav">
              <p className="challenges-explorer-eyebrow">Pick a month</p>
              <div className="challenge-month-rail" role="group" aria-label={`Choose month for ${year}`}>
                {MONTHLY_GOALS.map((g, i) => (
                  <button
                    key={g.id}
                    type="button"
                    className={`challenge-month-tile${selectedMonth === i ? ' is-selected' : ''}${currentMonth === i ? ' is-current' : ''}`}
                    aria-pressed={selectedMonth === i}
                    aria-label={`${MONTHS_LONG[i]}, ${CHALLENGE_CARD_HEADING[g.id]}`}
                    onClick={() => pickMonth(i)}
                  >
                    <span className="challenge-month-tile__abbr">{MONTHS_SHORT[i]}</span>
                    <span className="challenge-month-tile__theme">{g.title}</span>
                  </button>
                ))}
              </div>
              <div className="challenges-explorer-grow" aria-hidden="true" />
              <h2 id="challenge-month-heading" className="challenges-explorer-selection">
                {MONTHS_LONG[selectedMonth]}{' '}
                <span className="challenges-explorer-selection__year">{year}</span>
              </h2>
            </div>

            <article
              className={`challenge-card challenge-card--featured${selectedMonth === currentMonth ? ' challenge-card--current-accent' : ''}`}
              aria-labelledby="challenge-card-featured-title challenge-month-heading"
            >
              <div className="challenge-card__art">
                <img
                  src={monthlyChallengeImageSrc(goal.id)}
                  alt=""
                  className="challenge-card__icon"
                  loading="lazy"
                  decoding="async"
                />
                <div className="challenge-card__art-front">
                  <p className="challenge-card__month">{MONTHS_LONG[selectedMonth]}</p>
                </div>
                <span className="challenge-card__art-corner challenge-card__art-corner--tl" aria-hidden="true" />
                <span className="challenge-card__art-corner challenge-card__art-corner--br" aria-hidden="true" />
              </div>

              <h3 id="challenge-card-featured-title" className="sr-only">
                {CHALLENGE_CARD_HEADING[goal.id]}
              </h3>

              <div className="challenge-card__body">
                <p className="challenge-card__goal">
                  <strong>Goal:</strong> {goal.title}
                </p>
                <ul className="challenge-card__list clean">
                  {goal.bullets.map((b, i) => renderBullet(goal, b.who, b.text, i))}
                </ul>
              </div>
            </article>
          </div>

          <blockquote className="quote quote--sites challenges-quote">
            Daily habits shape eternal destinies.
            <cite>President Ezra Taft Benson</cite>
          </blockquote>
        </div>
      </main>
    </>
  )
}
