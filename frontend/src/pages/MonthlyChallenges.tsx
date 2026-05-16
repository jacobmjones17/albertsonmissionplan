import { useCallback, useMemo, useState } from 'react'
import {
  MONTHLY_GOALS,
  MONTHS_LONG,
  MONTHS_SHORT,
  type Goal,
  WEEKDAYS_SHORT,
} from '../monthlyData'
import { PageHero } from '../components/PageHero'

function daysInMonth(year: number, month: number) {
  return new Date(year, month + 1, 0).getDate()
}

function firstWeekday(year: number, month: number) {
  return new Date(year, month, 1).getDay()
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
  const dim = daysInMonth(year, selectedMonth)
  const startPad = firstWeekday(year, selectedMonth)
  const totalCells = Math.ceil((startPad + dim) / 7) * 7

  const pickMonth = useCallback((m: number) => {
    setSelectedMonth(m)
  }, [])

  const goPrevMonth = useCallback(() => {
    setSelectedMonth((m) => (m <= 0 ? 11 : m - 1))
  }, [])

  const goNextMonth = useCallback(() => {
    setSelectedMonth((m) => (m >= 11 ? 0 : m + 1))
  }, [])

  return (
    <>
      <PageHero
        titleId="page-title"
        eyebrow="Twelve goals"
        title="Monthly Challenges"
        lede="Ideas for individuals, families, and the ward—one focus each month of the year"
      />
      <div className="home-mission-strip home-mission-strip--tight">
        <p>
          Monthly challenges are activities individuals and families may take to improve gospel living.
          Choose a month below to see this year’s calendar and the goal for that month.
        </p>
      </div>
      <main id="main" className="challenges-main">
        <div className="wrap-wide challenges-wrap">
          <div className="challenges-toolbar">
            <p className="page-inline-eyebrow challenges-toolbar-label">Year at a glance</p>
            <p className="challenges-year" aria-hidden="true">
              {year}
            </p>
          </div>

          <div className="challenges-split">
            <div className="challenges-cal-column">
              <div className="year-month-grid" role="group" aria-label="Select month">
                {MONTHLY_GOALS.map((g, i) => (
                  <button
                    key={g.id}
                    type="button"
                    className={`year-month-cell${selectedMonth === i ? ' is-selected' : ''}${currentMonth === i ? ' is-current' : ''}`}
                    aria-pressed={selectedMonth === i}
                    aria-label={`${MONTHS_LONG[i]}, ${g.title}`}
                    onClick={() => pickMonth(i)}
                  >
                    <span className="year-month-cell__abbr">{MONTHS_SHORT[i]}</span>
                    <span className="year-month-cell__theme">{g.title}</span>
                  </button>
                ))}
              </div>

              <div className="mini-month-cal" aria-label={`${MONTHS_LONG[selectedMonth]} ${year}`}>
                <div className="mini-month-cal__header">
                  <button
                    type="button"
                    className="mini-month-cal__nav"
                    aria-label="Previous month"
                    onClick={goPrevMonth}
                  >
                    ‹
                  </button>
                  <h2 className="mini-month-cal__title" id="challenge-month-heading">
                    {MONTHS_LONG[selectedMonth]} {year}
                  </h2>
                  <button
                    type="button"
                    className="mini-month-cal__nav"
                    aria-label="Next month"
                    onClick={goNextMonth}
                  >
                    ›
                  </button>
                </div>
                <div className="mini-month-cal__weekdays">
                  {WEEKDAYS_SHORT.map((d) => (
                    <span key={d} className="mini-month-cal__wd">
                      {d}
                    </span>
                  ))}
                </div>
                <div className="mini-month-cal__days">
                  {Array.from({ length: totalCells }, (_, idx) => {
                    const dayNum = idx - startPad + 1
                    if (dayNum < 1 || dayNum > dim) {
                      return <span key={idx} className="mini-month-cal__day mini-month-cal__day--pad" />
                    }
                    const isToday =
                      year === today.getFullYear() &&
                      selectedMonth === today.getMonth() &&
                      dayNum === today.getDate()
                    return (
                      <span
                        key={idx}
                        className={`mini-month-cal__day${isToday ? ' is-today' : ''}`}
                      >
                        {dayNum}
                      </span>
                    )
                  })}
                </div>
              </div>
            </div>

            <article
              className="challenge-detail"
              aria-labelledby="challenge-month-heading challenge-goal-title"
            >
              <h3 id="challenge-goal-title" className="challenge-detail__heading">
                {goal.title}
              </h3>
              <p className="challenge-detail__lede">
                For <strong>{MONTHS_LONG[selectedMonth]}</strong>—pick one or two actions that fit your
                circumstances.
              </p>
              <ul className="challenge-detail__roles clean">
                {goal.bullets.map((b, i) => renderBullet(goal, b.who, b.text, i))}
              </ul>
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
