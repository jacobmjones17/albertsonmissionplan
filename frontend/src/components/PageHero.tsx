import { useState } from 'react'

const HERO_IMG = '/ward-hero.jpg'

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
}: PageHeroProps) {
  const [imgFailed, setImgFailed] = useState(false)
  const usePhoto = photo && !compact

  if (usePhoto) {
    return (
      <>
        <section
          className="hero-page hero-page--home"
          aria-labelledby={titleId}
        >
          {!imgFailed ? (
            <div className="hero-page__bg" aria-hidden="true">
              <img
                src={HERO_IMG}
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
      <section
        className={`page-hero${compact ? ' page-hero--compact' : ''}`}
        aria-labelledby={titleId}
      >
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
