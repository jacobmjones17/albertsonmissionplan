import { useEffect, useState } from 'react'

type FeaturePhotoProps = {
  /** Absolute path under public/, e.g. `/site/jesus-smiling.jpg`. Empty = render nothing. */
  src: string
  /** Optional short caption shown beneath the photo. */
  caption?: string
  /**
   * `wide` (default) reads ~min(960px, 100% of parent .wrap) and looks great mid-page.
   * `narrow` caps at ~520px for sidebars or quote-style placements.
   */
  size?: 'wide' | 'narrow'
  /**
   * Visual style. `painting` adds a soft frame + caption styling tuned for sacred art
   * (the default). `flat` removes the border + shadow for graphics like the heart quote.
   */
  variant?: 'painting' | 'flat'
}

/**
 * Decorative in-page photo wrapper. Falls back to nothing if the image fails to load so a
 * missing asset never breaks the layout.
 */
export function FeaturePhoto({
  src,
  caption,
  size = 'wide',
  variant = 'painting',
}: FeaturePhotoProps) {
  const url = String(src).trim()
  const [broken, setBroken] = useState(false)

  useEffect(() => {
    setBroken(false)
  }, [url])

  if (!url || broken) return null

  return (
    <figure
      className={`page-feature-photo page-feature-photo--${size} page-feature-photo--${variant}`}
    >
      <img
        src={url}
        alt=""
        className="page-feature-photo__img"
        decoding="async"
        loading="lazy"
        onError={() => setBroken(true)}
      />
      {caption ? <figcaption className="page-feature-photo__caption">{caption}</figcaption> : null}
    </figure>
  )
}
