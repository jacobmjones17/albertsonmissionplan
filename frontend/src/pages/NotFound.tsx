import { Link } from 'react-router-dom'
import { FeaturePhoto } from '../components/FeaturePhoto'
import { PageHero } from '../components/PageHero'
import { STATIC_SITE_PHOTOS } from '../staticSitePhotos'

/**
 * Friendly "page not found" view for mistyped or outdated links (catch-all route).
 * Uses the same structure as the other public pages: photo hero, mission strip, spotlight card,
 * feature painting.
 */
export function NotFound() {
  return (
    <main id="main" className="not-found-page">
      <PageHero
        titleId="page-title"
        eyebrow="Page not found"
        title="Even the Lost Are Found"
        lede={'“Rejoice with me; for I have found my sheep which was lost.” — Luke\u00a015:6'}
      />
      <div className="home-mission-strip">
        <p>
          The page you were looking for isn&apos;t here. The link may be mistyped, or the page may
          have moved.
        </p>
      </div>
      <div className="wrap-wide">
        <section className="ward-goals-spotlight not-found-card" aria-labelledby="not-found-heading">
          <header className="ward-goals-spotlight__header">
            <p className="ward-goals-spotlight__eyebrow">Let&apos;s get you back</p>
            <h2 id="not-found-heading" className="ward-goals-spotlight__title">
              Back on the Path
            </h2>
            <p className="ward-goals-spotlight__lead">
              Even when we lose our way, the Savior knows where to find us. Head home, or see
              what&apos;s coming up in the ward.
            </p>
          </header>
          <p className="not-found-card__actions">
            <Link className="btn btn-primary" to="/">
              Back to Home
            </Link>
            <Link className="btn btn-secondary" to="/events">
              See Ward Events
            </Link>
          </p>
        </section>

        <FeaturePhoto
          src={STATIC_SITE_PHOTOS.notFound}
          alt="Painting of Jesus Christ holding a young child close in His arms"
          caption={'“For the Son of man is come to seek and to save that which was lost.” — Luke\u00a019:10'}
          size="narrow"
          loading="eager"
        />
      </div>
    </main>
  )
}
