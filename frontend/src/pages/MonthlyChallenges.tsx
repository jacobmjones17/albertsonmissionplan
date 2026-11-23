import { useCallback, useMemo, useState } from 'react'
import {
  CHALLENGE_CARD_HEADING,
  MONTHLY_GOALS,
  MONTHS_LONG,
  MONTHS_SHORT,
  type Goal,
} from '../monthlyData'
import { PageHero } from '../components/PageHero'
import { DocumentDownloadButton } from '../components/DocumentDownloadButton'
import { buildOfflineMonthlyChallengesHtml } from '../lib/downloadMonthlyChallenges'
import { renderChallengeBullet } from '../lib/challengeBullets'
import {
  MONTHLY_CHALLENGE_IMAGE_HEIGHT,
  MONTHLY_CHALLENGE_IMAGE_SIZES,
  MONTHLY_CHALLENGE_IMAGE_WIDTH,
  monthlyChallengeImageSrc,
  monthlyChallengeImageSrcSet,
  preloadMonthlyChallengeImage,
} from '../monthlyChallengeImages'

/** Short tile label: "Scripture Study Challenge" → "Scripture Study". */
function shortTheme(goalId: Goal['id']): string {
  return CHALLENGE_CARD_HEADING[goalId].replace(/\s+Challenge$/, '')
}

function renderBullet(goal: Goal, who: string, text: string, key: number) {
  return renderChallengeBullet(goal, who, text, key)
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

  /*
   * Only the selected month's art is rendered. Other months load on demand: when their tile is
   * hovered/focused (warm-up) or picked. No idle prefetch of all twelve images.
   */

  return (
    <main id="main" className="challenges-main challenges-main--cards">
      <PageHero
        titleId="page-title"
        eyebrow="Year-round focus"
        title="Monthly Challenges"
        lede="One themed challenge each month for individuals, couples, and families"
      />
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
                  onMouseEnter={() => preloadMonthlyChallengeImage(g.id)}
                  onFocus={() => preloadMonthlyChallengeImage(g.id)}
                >
                  <span className="challenge-month-tile__abbr">{MONTHS_SHORT[i]}</span>
                  <span className="challenge-month-tile__theme">{shortTheme(g.id)}</span>
                </button>
              ))}
            </div>
            <div className="challenges-explorer-grow" aria-hidden="true" />
            <h2 id="challenge-month-heading" className="challenges-explorer-selection">
              {MONTHS_LONG[selectedMonth]}{' '}
              <span className="challenges-explorer-selection__year">{year}</span>
            </h2>
          </div>

          <div className="challenges-featured-stack">
            <article
              className={`challenge-card challenge-card--featured${selectedMonth === currentMonth ? ' challenge-card--current-accent' : ''}`}
              aria-labelledby="challenge-card-featured-title challenge-month-heading"
            >
              <div className="challenge-card__art">
                <img
                  key={goal.id}
                  src={monthlyChallengeImageSrc(goal.id)}
                  srcSet={monthlyChallengeImageSrcSet(goal.id)}
                  sizes={MONTHLY_CHALLENGE_IMAGE_SIZES}
                  width={MONTHLY_CHALLENGE_IMAGE_WIDTH}
                  height={MONTHLY_CHALLENGE_IMAGE_HEIGHT}
                  alt=""
                  className="challenge-card__icon"
                  loading="eager"
                  decoding="async"
                  fetchPriority="high"
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
                <ul className="challenge-card__list">
                  {goal.bullets.map((b, i) => renderBullet(goal, b.who, b.text, i))}
                </ul>
              </div>
            </article>

            <DocumentDownloadButton
              label={`Download Printable ${year} Challenges`}
              filename={`Albertson-Ward-Monthly-Challenges-${year}.html`}
              buildDocument={() => buildOfflineMonthlyChallengesHtml(year)}
            />
          </div>
        </div>

        <blockquote className="quote quote--sites challenges-quote">
          Daily habits shape eternal destinies.
          <cite>President Ezra Taft Benson</cite>
        </blockquote>
      </div>
    </main>
  )
}
