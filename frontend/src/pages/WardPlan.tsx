import { useEffect, useState } from 'react'
import { Link } from 'react-router-dom'
import { apiJson } from '../api'
import { FeaturePhoto } from '../components/FeaturePhoto'
import { PageHero } from '../components/PageHero'
import { WardGoalLine } from '../components/richtext/WardGoalLine'
import { STATIC_SITE_PHOTOS } from '../staticSitePhotos'

type Org = { slug: string; title: string; bullets: string[] }

function PrintIcon() {
  return (
    <svg viewBox="0 0 24 24" width="18" height="18" fill="none" aria-hidden="true" focusable="false">
      <path
        d="M7 8V4h10v4M7 17H5a2 2 0 0 1-2-2v-5a2 2 0 0 1 2-2h14a2 2 0 0 1 2 2v5a2 2 0 0 1-2 2h-2"
        stroke="currentColor"
        strokeWidth="1.75"
        strokeLinecap="round"
        strokeLinejoin="round"
      />
      <path d="M7 13h10v7H7v-7Z" stroke="currentColor" strokeWidth="1.75" strokeLinejoin="round" />
    </svg>
  )
}

export function WardPlan() {
  const [wardGoals, setWardGoals] = useState<string[]>([])
  const [orgs, setOrgs] = useState<Org[]>([])
  const [err, setErr] = useState<string | null>(null)
  const [loading, setLoading] = useState(true)

  useEffect(() => {
    apiJson<{ wardGoals: string[]; orgs: Org[] }>('/api/ward-plan')
      .then((d) => {
        setWardGoals(d.wardGoals)
        setOrgs(d.orgs)
      })
      .catch(() => setErr('Could not load the ward plan. Please refresh the page to try again.'))
      .finally(() => setLoading(false))
  }, [])

  return (
    <main id="main">
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
      <div className="wrap-wide ward-plan-page">
        {err ? (
          <p className="flash" role="alert">
            {err}
          </p>
        ) : null}
        <section
          className="ward-goals-spotlight"
          aria-labelledby="ward-goals"
          aria-busy={loading || undefined}
        >
          <header className="ward-goals-spotlight__header">
            <p className="ward-goals-spotlight__eyebrow">Together</p>
            <h2 id="ward-goals" className="ward-goals-spotlight__title">
              Ward Goals
            </h2>
            <p className="ward-goals-spotlight__lead">
              <strong>We will strengthen our ward by:</strong>
            </p>
          </header>
          {loading ? (
            <p className="loading-note" role="status">
              <span className="loading-note__spinner" aria-hidden="true" />
              Loading ward goals…
            </p>
          ) : (
            <ul className="ward-goals-spotlight__list">
              {wardGoals.map((g, i) => (
                <li key={i}>
                  <WardGoalLine content={g} />
                </li>
              ))}
            </ul>
          )}
          <div className="ward-plan-actions">
            <Link to="/ward-plan/print" className="btn btn-secondary document-download__btn">
              <PrintIcon />
              <span>Print / Save as PDF</span>
            </Link>
          </div>
        </section>

        <section className="ward-plan-org-section" aria-labelledby="org-goals">
          <header className="ward-plan-org-section__header">
            <p className="page-inline-eyebrow">By organization</p>
            <h2 id="org-goals">Auxiliary and Quorum Goals</h2>
            <p className="ward-plan-org-section__intro">
              Each organization has specific ways to help accomplish the ward mission plan. Expand a
              card to read its goals.
            </p>
          </header>
          {loading ? (
            <p className="loading-note" role="status">
              <span className="loading-note__spinner" aria-hidden="true" />
              Loading organization goals…
            </p>
          ) : (
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
          )}
        </section>

        <blockquote className="scripture-pull scripture-pull--ward" id="scripture">
          <p>
            &ldquo;Go ye therefore, and teach all nations, baptizing them in the name of the Father,
            and of the Son, and of the Holy Ghost.&rdquo;
          </p>
          <cite>Matthew 28:19</cite>
        </blockquote>

        <FeaturePhoto
          src={STATIC_SITE_PHOTOS.wardPlanFeature}
          rotateSlot="ward-plan-feature"
          ratio={1800 / 1195}
          alt="Statue of the resurrected Jesus Christ with outstretched arms, set against the earth and a starry sky"
          caption={
            '“For God so loved the world, that he gave his only begotten Son, that whosoever believeth in him should not perish, but have everlasting life.” — John 3:16'
          }
        />
      </div>
    </main>
  )
}
