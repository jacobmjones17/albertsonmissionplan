import { useEffect, useState } from 'react'
import { STATIC_SITE_PHOTOS } from '../staticSitePhotos'

type PageHeroProps = {
  titleId: string
  title: string
  lede?: string
  eyebrow?: string
  /** Gray/dark strip directly under the hero (e.g. Love, Share, Invite tagline). */
  banner?: string
  /** Shorter hero for admin tools (no photo). */
  compact?: boolean
  /** Full-width banner photo like the home page (default true when not compact). */
  photo?: boolean
  /**
   * Optional override (`/site/...`). Empty uses `STATIC_SITE_PHOTOS.header` so Ward Plan /
   * Monthly Challenges / etc. match the navbar hero treatment site-wide.
   */
  photoSrc?: string
}

/**
 * Shared top banner: full photo hero (default) or compact slate bar for admin.
 */
export function PageHero({
  titleId,
  title,
  lede,
  eyebrow,
  banner,
  compact = false,
  photo = true,
  photoSrc = '',
}: PageHeroProps) {
  const fallbackHeader = String(STATIC_SITE_PHOTOS.header).trim()
  const trimmed = String(photoSrc).trim() || fallbackHeader

  /** Same global merge as Layout + Home whenever the site header image is configured */
  const mergeUnderNavbar = Boolean(fallbackHeader) && photo && !compact

  const [imgFailed, setImgFailed] = useState(false)

  useEffect(() => {
    setImgFailed(false)
  }, [trimmed])

  const usePhoto = photo && !compact && Boolean(trimmed) && !imgFailed

  if (photo && !compact) {
    return (
      <>
        <section
          className={`hero-page hero-page--home${mergeUnderNavbar ? ' hero-page--underlap-header' : ''}`}
          aria-labelledby={titleId}
        >
          {usePhoto ? (
            <div className="hero-page__bg" aria-hidden="true">
              <img
                src={trimmed}
                alt=""
                className="hero-page__bg-img"
                decoding="async"
                onError={() => setImgFailed(true)}
              />
            </div>
          ) : null}
          <div className="hero-page__dim" aria-hidden="true" />
          <div className="hero-page__inner">
            {eyebrow ? <p className="page-hero__eyebrow page-hero__eyebrow--on-photo">{eyebrow}</p> : null}
            <h1 id={titleId}>{title}</h1>
            {lede ? <p className="page-hero__lede page-hero__lede--on-photo">{lede}</p> : null}
          </div>
        </section>
        {banner ? (
          <div className="page-subbanner">
            <p>{banner}</p>
          </div>
        ) : null}
      </>
    )
  }

  return (
    <>
      <section className={`page-hero${compact ? ' page-hero--compact' : ''}`} aria-labelledby={titleId}>
        <div className="page-hero__inner">
          {eyebrow ? <p className="page-hero__eyebrow">{eyebrow}</p> : null}
          <h1 id={titleId}>{title}</h1>
          {lede ? <p className="page-hero__lede">{lede}</p> : null}
        </div>
      </section>
      {banner ? (
        <div className="page-subbanner">
          <p>{banner}</p>
        </div>
      ) : null}
    </>
  )
}
