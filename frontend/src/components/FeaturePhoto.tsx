import { useEffect, useState, type CSSProperties } from 'react'
import { useSlotDrawing } from '../lib/galleryDrawings'

type FeaturePhotoProps = {
  /** Absolute path under public/, e.g. `/site/jesus-smiling.jpg`. Empty = render nothing. */
  src: string
  /**
   * Text alternative for screen readers. Describe what the image shows (or quote any words in
   * it). Leave empty only for purely decorative images.
   */
  alt?: string
  /** Optional short caption shown beneath the photo. */
  caption?: string
  /** `lazy` (default) for below-the-fold photos; `eager` when the photo is the page's focus. */
  loading?: 'lazy' | 'eager'
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
  /**
   * Rotating slot: a unique key per slot on the page. When set, the slot shows a random Ward
   * Gallery drawing (a different one for each slot on the page, fresh on every page view) and
   * falls back to `src` when there are no drawings left or the drawing fails to load.
   */
  rotateSlot?: string
  /** Natural width / height of `src`. Rotating slots keep this box so nothing jumps on swap. */
  ratio?: number
  /** Caption to use instead of `caption` while a drawing is shown (e.g. a quote that was in the photo). */
  drawingCaption?: string
}

const DRAWING_ALT = 'Drawing by a ward member, from the Ward Gallery'

/**
 * In-page photo wrapper. Falls back to nothing if the image fails to load so a
 * missing asset never breaks the layout.
 */
export function FeaturePhoto({
  src,
  alt = '',
  caption,
  loading = 'lazy',
  size = 'wide',
  variant = 'painting',
  rotateSlot,
  ratio,
  drawingCaption,
}: FeaturePhotoProps) {
  const url = String(src).trim()
  const [broken, setBroken] = useState(false)
  const drawing = useSlotDrawing(rotateSlot)
  const [drawingBroken, setDrawingBroken] = useState(false)

  useEffect(() => {
    setBroken(false)
  }, [url])
  useEffect(() => {
    setDrawingBroken(false)
  }, [drawing?.id])

  const showDrawing = Boolean(drawing && !drawingBroken)
  const loadingList = drawing === undefined
  if (!showDrawing && !loadingList && (!url || broken)) return null

  const framed = Boolean(rotateSlot && ratio)
  const frameStyle: CSSProperties | undefined = framed ? { aspectRatio: String(ratio) } : undefined
  const shownCaption = showDrawing ? (drawingCaption ?? caption) : caption

  return (
    <figure
      className={`page-feature-photo page-feature-photo--${size} page-feature-photo--${variant}${
        framed ? ' page-feature-photo--framed' : ''
      }${showDrawing ? ' page-feature-photo--drawing' : ''}`}
    >
      {framed ? (
        <div className="page-feature-photo__frame" style={frameStyle}>
          {showDrawing && drawing ? (
            <img
              key={drawing.id}
              src={drawing.imageUrl}
              srcSet={`${drawing.thumbUrl} ${drawing.thumbWidth}w, ${drawing.imageUrl} ${drawing.width}w`}
              sizes={size === 'narrow' ? '(max-width: 520px) 92vw, 480px' : '(max-width: 760px) 92vw, 720px'}
              alt={DRAWING_ALT}
              className="page-feature-photo__img page-feature-photo__img--drawing"
              decoding="async"
              loading={loading}
              onError={() => setDrawingBroken(true)}
            />
          ) : loadingList ? null : (
            <img
              src={url}
              alt={alt}
              className="page-feature-photo__img"
              decoding="async"
              loading={loading}
              onError={() => setBroken(true)}
            />
          )}
        </div>
      ) : (
        <img
          src={url}
          alt={alt}
          className="page-feature-photo__img"
          decoding="async"
          loading={loading}
          onError={() => setBroken(true)}
        />
      )}
      {shownCaption ? <figcaption className="page-feature-photo__caption">{shownCaption}</figcaption> : null}
    </figure>
  )
}
