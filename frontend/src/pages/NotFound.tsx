import { Link } from 'react-router-dom'
import { PageHero } from '../components/PageHero'
import { STATIC_SITE_PHOTOS } from '../staticSitePhotos'

/** Friendly "page not found" view for mistyped or outdated links (catch-all route). */
export function NotFound() {
  return (
    <main id="main" className="not-found-page">
      <PageHero
        titleId="page-title"
        eyebrow="Page Not Found"
        title="We Couldn’t Find That Page"
        lede="The link may be mistyped or the page may have moved."
      />
      <div className="wrap-wide">
        <section className="not-found" aria-labelledby="not-found-message">
          <figure className="not-found__art">
            <img
              src={STATIC_SITE_PHOTOS.notFound}
              alt="Painting of Jesus Christ holding a young child close in His arms"
              width={1066}
              height={1599}
              decoding="async"
            />
          </figure>
          <div className="not-found__body">
            <h2 id="not-found-message" className="not-found__title">
              Even when we lose our way, the Savior knows where to find us.
            </h2>
            <blockquote className="not-found__quote">
              <p>&ldquo;For the Son of man is come to seek and to save that which was lost.&rdquo;</p>
              <cite>Luke 19:10</cite>
            </blockquote>
            <p className="not-found__lead">
              Let&apos;s get you back on the path. Head home, or jump to one of these pages:
            </p>
            <p className="not-found__actions">
              <Link className="btn btn-primary" to="/">
                Go to Home
              </Link>
              <Link className="btn btn-secondary" to="/ward-plan">
                Ward Plan
              </Link>
              <Link className="btn btn-secondary" to="/monthly-challenges">
                Monthly Challenges
              </Link>
            </p>
          </div>
        </section>
      </div>
    </main>
  )
}
