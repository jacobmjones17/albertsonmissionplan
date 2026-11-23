import { useMemo, useState } from 'react'
import { MONTHS_LONG, MONTHS_SHORT } from '../monthlyData'
import { PageHero } from '../components/PageHero'
import { ChallengeCard } from '../components/ChallengeCard'
import { DocumentDownloadButton } from '../components/DocumentDownloadButton'
import { buildOfflineMonthlyChallengesHtml } from '../lib/downloadMonthlyChallenges'
import {
  preloadChallengeImage,
  shortTheme,
  useMonthlyChallenges,
  type MonthlyChallenge,
} from '../lib/monthlyChallenges'
import { MONTHLY_CHALLENGE_IMAGE_SIZES } from '../monthlyChallengeImages'

function ChallengeExplorer({ challenges, year, currentMonth }: { challenges: MonthlyChallenge[]; year: number; currentMonth: number }) {
  // Start on this month; if leaders hid it, start on the first published month.
  const [picked, setPicked] = useState<number | null>(null)
  const selectedMonth =
    picked ?? (challenges.some((c) => c.month === currentMonth) ? currentMonth : (challenges[0]?.month ?? currentMonth))
  const challenge = challenges.find((c) => c.month === selectedMonth)

  /*
   * Only the selected month's art is rendered. Other months load on demand: when their tile is
   * hovered/focused (warm-up) or picked. No idle prefetch of all twelve images.
   */

  if (!challenge) {
    return (
      <p className="challenges-intro" role="status">
        No challenges are published right now. Check back soon.
      </p>
    )
  }

  return (
    <div className="challenges-split challenges-split--explorer">
      <div className="challenges-explorer-nav">
        <p className="challenges-explorer-eyebrow">Pick a month</p>
        <div className="challenge-month-rail" role="group" aria-label={`Choose month for ${year}`}>
          {challenges.map((c) => (
            <button
              key={c.month}
              type="button"
              className={`challenge-month-tile${selectedMonth === c.month ? ' is-selected' : ''}${currentMonth === c.month ? ' is-current' : ''}`}
              aria-pressed={selectedMonth === c.month}
              aria-label={`${MONTHS_LONG[c.month - 1]}, ${c.heading}`}
              onClick={() => setPicked(c.month)}
              onMouseEnter={() => preloadChallengeImage(c, MONTHLY_CHALLENGE_IMAGE_SIZES)}
              onFocus={() => preloadChallengeImage(c, MONTHLY_CHALLENGE_IMAGE_SIZES)}
            >
              <span className="challenge-month-tile__abbr">{MONTHS_SHORT[c.month - 1]}</span>
              <span className="challenge-month-tile__theme">{shortTheme(c.heading)}</span>
            </button>
          ))}
        </div>
        <div className="challenges-explorer-grow" aria-hidden="true" />
        <h2 id="challenge-month-heading" className="challenges-explorer-selection">
          {MONTHS_LONG[selectedMonth - 1]}{' '}
          <span className="challenges-explorer-selection__year">{year}</span>
        </h2>
      </div>

      <div className="challenges-featured-stack">
        <ChallengeCard
          challenge={challenge}
          isCurrent={selectedMonth === currentMonth}
          labelledBy="challenge-month-heading"
        />

        <DocumentDownloadButton
          label={`Download Printable ${year} Challenges`}
          filename={`Albertson-Ward-Monthly-Challenges-${year}.html`}
          buildDocument={() => buildOfflineMonthlyChallengesHtml(year, challenges)}
        />
      </div>
    </div>
  )
}

export function MonthlyChallenges() {
  const today = useMemo(() => new Date(), [])
  const year = today.getFullYear()
  const currentMonth = today.getMonth() + 1
  const state = useMonthlyChallenges()

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

        {state.status === 'loading' ? (
          <p className="loading-note challenges-loading" role="status">
            <span className="loading-note__spinner" aria-hidden="true" />
            Loading this year&apos;s challenges…
          </p>
        ) : (
          <ChallengeExplorer challenges={state.challenges} year={year} currentMonth={currentMonth} />
        )}

        <blockquote className="quote quote--sites challenges-quote">
          Daily habits shape eternal destinies.
          <cite>President Ezra Taft Benson</cite>
        </blockquote>
      </div>
    </main>
  )
}
