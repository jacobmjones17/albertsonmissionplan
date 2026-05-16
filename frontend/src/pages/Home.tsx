import { useEffect, useState } from 'react'
import { Link } from 'react-router-dom'
import { apiJson } from '../api'

type Experience = { body: string; author: string; when?: string }

export function Home() {
  const [experiences, setExperiences] = useState<Experience[]>([])
  const [err, setErr] = useState<string | null>(null)
  const [heroPhoto, setHeroPhoto] = useState(true)
  const [missionaryPhoto, setMissionaryPhoto] = useState(true)

  useEffect(() => {
    apiJson<{ experiences: Experience[] }>('/api/home/experiences')
      .then((d: { experiences: Experience[] }) => setExperiences(d.experiences))
      .catch(() => setErr('Could not load experiences.'))
  }, [])

  return (
    <main id="main" className="home-page">
      <section className="hero-page hero-page--home" aria-labelledby="hero-title">
        {heroPhoto ? (
          <div className="hero-page__bg" aria-hidden="true">
            <img
              src="/ward-hero.jpg"
              alt=""
              className="hero-page__bg-img"
              decoding="async"
              onError={() => setHeroPhoto(false)}
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
              Ward plan
            </Link>
            <Link className="btn btn-secondary" to="/monthly-challenges">
              Monthly challenges
            </Link>
          </nav>
        </div>
      </section>

      <section className="home-about" aria-labelledby="home-about-title">
        <div className="home-about__inner">
          <h2 id="home-about-title">We are all missionaries</h2>
          <p>
            This site exists to encourage you and your family in the Lord&apos;s work.
          </p>

          <blockquote className="home-about__quote">
            <p>
              &ldquo;All missionaries, younger and older, serve with the sole hope of making life better
              for other people.&rdquo;
            </p>
            <cite>President Russell M. Nelson</cite>
          </blockquote>

          <p>
            Ministering is Christlike caring for others — motivated by our desire to love our neighbor and
            serve both spiritual and temporal needs.
          </p>

          {missionaryPhoto ? (
            <figure className="home-about__figure">
              <img
                src="/ward-missionaries.jpg"
                alt=""
                className="home-about__figure-img"
                decoding="async"
                onError={() => setMissionaryPhoto(false)}
              />
            </figure>
          ) : null}
        </div>
      </section>

      <div className="wrap wrap--home">
        <section className="content-block content-block--experiences" aria-labelledby="sec-experiences">
          <div className="section-heading">
            <h2 id="sec-experiences">Mission experiences</h2>
            <p className="section-sub">Stories from members, reviewed by ward leaders before publishing.</p>
          </div>
          {err ? <p className="lead">{err}</p> : null}
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
              Read more &amp; share an experience
            </Link>
          </p>
        </section>
      </div>
    </main>
  )
}
