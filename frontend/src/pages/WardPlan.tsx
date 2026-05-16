import { useEffect, useState } from 'react'
import { apiJson } from '../api'
import { PageHero } from '../components/PageHero'
import { WardGoalLine } from '../components/richtext/WardGoalLine'

type Org = { slug: string; title: string; bullets: string[] }

export function WardPlan() {
  const [wardGoals, setWardGoals] = useState<string[]>([])
  const [orgs, setOrgs] = useState<Org[]>([])
  const [err, setErr] = useState<string | null>(null)

  useEffect(() => {
    apiJson<{ wardGoals: string[]; orgs: Org[] }>('/api/ward-plan')
      .then((d) => {
        setWardGoals(d.wardGoals)
        setOrgs(d.orgs)
      })
      .catch(() => setErr('Could not load ward plan.'))
  }, [])

  return (
    <>
      <PageHero
        titleId="page-title"
        eyebrow="Albertson Ward"
        title="Ward Mission Plan"
        lede="For sharing the gospel and strengthening new and returning members"
      />
      <div className="home-mission-strip">
        <p>
          The ward mission plan is a simple document that outlines goals to help ward and auxiliary
          leaders participate in missionary work, retention, and activation.
        </p>
      </div>
      <main id="main">
        <div className="wrap-wide ward-plan-page">
          {err ? <p className="flash">{err}</p> : null}
          <section className="ward-goals-spotlight" aria-labelledby="ward-goals">
            <header className="ward-goals-spotlight__header">
              <p className="ward-goals-spotlight__eyebrow">Together</p>
              <h2 id="ward-goals" className="ward-goals-spotlight__title">
                Ward goals
              </h2>
              <p className="ward-goals-spotlight__lead">
                <strong>We will strengthen our ward by:</strong>
              </p>
            </header>
            <ul className="ward-goals-spotlight__list">
              {wardGoals.map((g, i) => (
                <li key={i}>
                  <WardGoalLine content={g} />
                </li>
              ))}
            </ul>
          </section>

          <section className="ward-plan-org-section" aria-labelledby="org-goals">
            <header className="ward-plan-org-section__header">
              <p className="page-inline-eyebrow">By organization</p>
              <h2 id="org-goals">Auxiliary and quorum goals</h2>
              <p className="ward-plan-org-section__intro">
                Each organization has specific ways to help accomplish the ward mission plan. Expand a
                card to read its goals.
              </p>
            </header>
            <div className="org-goals-grid">
              {orgs.map((o) => (
                <details className="org org-card" key={o.slug}>
                  <summary>{o.title}</summary>
                  <div className="org-body">
                    <ul className="clean">
                      {o.bullets.map((b, j) => (
                        <li key={o.slug + j}>
                          <WardGoalLine content={b} />
                        </li>
                      ))}
                    </ul>
                  </div>
                </details>
              ))}
            </div>
          </section>

          <h2 className="scripture-pull scripture-pull--ward" id="scripture">
            Matthew 28:19 — &ldquo;Go ye therefore, and teach all nations, baptizing them in the name of
            the Father, and of the Son, and of the Holy Ghost.&rdquo;
          </h2>
        </div>
      </main>
    </>
  )
}
