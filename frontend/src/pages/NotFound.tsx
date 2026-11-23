import { Link } from 'react-router-dom'
import { STATIC_SITE_PHOTOS } from '../staticSitePhotos'

/** Friendly "page not found" view for mistyped or outdated links (catch-all route). */
export function NotFound() {
  return (
    <main id="main" className="not-found-page">
      <section className="nf" aria-labelledby="page-title">
        <div className="nf__inner">
          <figure className="nf__art">
            <img
              src={STATIC_SITE_PHOTOS.notFound}
              alt="Painting of Jesus Christ holding a young child close in His arms"
              width={1066}
              height={1599}
              decoding="async"
              fetchPriority="high"
            />
          </figure>
          <div className="nf__body">
            <p className="nf__eyebrow">
              <span>404</span>
              <span aria-hidden="true" className="nf__dot">
                ·
              </span>
              <span>Page not found</span>
            </p>
            <h1 id="page-title" className="nf__title">
              Even the lost are found.
            </h1>
            <span className="nf__rule" aria-hidden="true" />
            <blockquote className="nf__quote">
              <p>&ldquo;Rejoice with me; for I have found my sheep which was lost.&rdquo;</p>
              <cite>Luke 15:6</cite>
            </blockquote>
            <p className="nf__lead">
              The page you were looking for isn&apos;t here. The link may be mistyped, or the
              page may have moved. Let&apos;s get you back on the path.
            </p>
            <div className="nf__actions">
              <Link className="btn btn-primary nf__btn" to="/">
                Back to Home
              </Link>
              <Link className="nf__link" to="/events">
                See ward events <span aria-hidden="true">→</span>
              </Link>
            </div>
          </div>
        </div>
      </section>
    </main>
  )
}
