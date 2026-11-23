import { useEffect, useState } from 'react'
import { Link } from 'react-router-dom'
import { apiJson } from '../api'
import { FeaturePhoto } from '../components/FeaturePhoto'
import { STATIC_SITE_PHOTOS } from '../staticSitePhotos'

type Experience = { body: string; author: string; when?: string }

export function Home() {
  const [experiences, setExperiences] = useState<Experience[]>([])
  const [err, setErr] = useState<string | null>(null)
  const [loading, setLoading] = useState(true)
  const [heroBroken, setHeroBroken] = useState(false)

  const headerBackdrop = String(STATIC_SITE_PHOTOS.header).trim()
  const homeHeroUrl = STATIC_SITE_PHOTOS.pageHome.trim() || headerBackdrop
  const unifiedHomeBackdrop = Boolean(headerBackdrop)

  useEffect(() => {
    apiJson<{ experiences: Experience[] }>('/api/home/experiences')
      .then((d: { experiences: Experience[] }) => setExperiences(d.experiences))
      .catch(() => setErr('Could not load experiences right now.'))
      .finally(() => setLoading(false))
  }, [])

  useEffect(() => {
    setHeroBroken(false)
  }, [homeHeroUrl])

  const showHeroImage = Boolean(homeHeroUrl && !heroBroken)

  return (
    <main id="main" className="home-page">
      <section
        className={`hero-page hero-page--home${unifiedHomeBackdrop ? ' hero-page--underlap-header' : ''}`}
        aria-labelledby="hero-title"
      >
        {showHeroImage ? (
          <div className="hero-page__bg" aria-hidden="true">
            <img
              src={homeHeroUrl}
              alt=""
              className="hero-page__bg-img"
              decoding="async"
              onError={() => setHeroBroken(true)}
            />
          </div>
        ) : null}
        <div className="hero-page__dim" aria-hidden="true" />
        <div className="hero-page__inner">
          <h1 id="hero-title">Albertson Ward</h1>
          <p className="hero-page__tagline">
            Ward mission website — goals and ideas for missionary work, retention, and activation.
          </p>
          <nav className="home-hero__ctas" aria-label="Popular pages">
            <Link className="btn btn-primary" to="/ward-plan">
              Ward Plan
            </Link>
            <Link className="btn btn-secondary" to="/monthly-challenges">
              Monthly Challenges
            </Link>
          </nav>
        </div>
      </section>

      <section className="home-about" aria-labelledby="home-about-title">
        <div className="home-about__inner">
          <h2 id="home-about-title">We Are All Missionaries</h2>
          <p>This site exists to encourage you and your family in the Lord&apos;s work.</p>

          <blockquote className="home-about__quote">
            <p>
              &ldquo;All missionaries, younger and older, serve with the sole hope of making life better
              for other people.&rdquo;
            </p>
            <cite>President Russell M. Nelson</cite>
          </blockquote>

          <p>
            Ministering is Christlike caring for others — motivated by our desire to love our neighbor and serve
            both spiritual and temporal needs.
          </p>
        </div>
      </section>

      <div className="wrap wrap--home">
        <section className="content-block content-block--experiences" aria-labelledby="sec-experiences">
          <div className="section-heading">
            <h2 id="sec-experiences">Mission Experiences</h2>
            <p className="section-sub">Stories from members, reviewed by ward leaders before publishing.</p>
          </div>
          {loading ? (
            <p className="loading-note" role="status">
              <span className="loading-note__spinner" aria-hidden="true" />
              Loading experiences…
            </p>
          ) : null}
          {err ? <p className="lead">{err}</p> : null}
          {!loading && !err && experiences.length === 0 ? (
            <p className="home-empty-note">
              No stories have been published yet. Be the first to share how the Lord has blessed you.
            </p>
          ) : null}
          {experiences.length > 0 ? (
            <div className="testimonial-stack">
              {experiences.map((ex, i) => (
                <article className="testimonial-card testimonial-card--home" key={i}>
                  <p className="testimonial-body">{ex.body}</p>
                  <p className="testimonial-meta">
                    {ex.author}
                    {ex.when ? ` · ${ex.when}` : ''}
                  </p>
                </article>
              ))}
            </div>
          ) : null}
          <p className="cta-row">
            <Link className="btn btn-primary" to="/experiences">
              Read and Share Experiences
            </Link>
          </p>
        </section>
        <FeaturePhoto
          src={STATIC_SITE_PHOTOS.about}
          size="narrow"
          variant="flat"
          alt="“If you’re not a full-time missionary with a missionary badge pinned on your coat, now is the time to paint one on your heart—painted, as Paul said, ‘not with ink, but with the Spirit of the living God.’” — Elder Neil L. Andersen"
        />
      </div>
    </main>
  )
}
