import { useEffect, useRef, useState, type PointerEvent as ReactPointerEvent } from 'react'
import { type GalleryItem } from '../lib/gallery'

type Props = {
  items: GalleryItem[]
  /** Index of the open drawing, or null when closed. */
  index: number | null
  onIndexChange: (index: number | null) => void
}

const SWIPE_PX = 50

/**
 * Full-screen viewer on a native <dialog> (focus trap + Esc for free). Arrow keys and the
 * on-screen buttons move between drawings; on phones a horizontal swipe does the same.
 */
export function GalleryLightbox({ items, index, onIndexChange }: Props) {
  const dialogRef = useRef<HTMLDialogElement>(null)
  const swipeStart = useRef<{ x: number; y: number } | null>(null)
  const [loaded, setLoaded] = useState(false)
  const open = index !== null && index >= 0 && index < items.length
  const item = open ? items[index] : null
  const count = items.length

  useEffect(() => {
    const dlg = dialogRef.current
    if (!dlg) return
    if (open && !dlg.open) dlg.showModal()
    if (!open && dlg.open) dlg.close()
  }, [open])

  useEffect(() => setLoaded(false), [item?.id])

  // Preload the neighbors so next/previous feel instant.
  useEffect(() => {
    if (!open || count < 2) return
    for (const d of [-1, 1]) {
      const n = items[(index + d + count) % count]
      const img = new Image()
      img.src = n.imageUrl
    }
  }, [open, index, items, count])

  function go(delta: number) {
    if (!open || count < 2) return
    onIndexChange((index + delta + count) % count)
  }

  function onKeyDown(e: React.KeyboardEvent) {
    if (e.key === 'ArrowRight') {
      e.preventDefault()
      go(1)
    } else if (e.key === 'ArrowLeft') {
      e.preventDefault()
      go(-1)
    }
  }

  function onPointerDown(e: ReactPointerEvent) {
    if (e.pointerType === 'mouse') return
    swipeStart.current = { x: e.clientX, y: e.clientY }
  }

  function onPointerUp(e: ReactPointerEvent) {
    const start = swipeStart.current
    swipeStart.current = null
    if (!start) return
    const dx = e.clientX - start.x
    const dy = e.clientY - start.y
    if (Math.abs(dx) > SWIPE_PX && Math.abs(dx) > Math.abs(dy) * 1.5) go(dx < 0 ? 1 : -1)
  }

  return (
    <dialog
      ref={dialogRef}
      className="gallery-lightbox"
      aria-label="Drawing viewer"
      onClose={() => onIndexChange(null)}
      onKeyDown={onKeyDown}
      onClick={(e) => {
        // A click on the dark backdrop (the dialog itself, not its content) closes it.
        if (e.target === e.currentTarget) onIndexChange(null)
      }}
    >
      {item ? (
        <div className="gallery-lightbox__frame">
          <div className="gallery-lightbox__top">
            <p className="gallery-lightbox__count" aria-live="polite">
              {index + 1} of {count}
            </p>
            <button
              type="button"
              className="gallery-lightbox__close"
              onClick={() => onIndexChange(null)}
              aria-label="Close"
              autoFocus
            >
              <span aria-hidden="true">×</span>
            </button>
          </div>
          <div
            className="gallery-lightbox__stage"
            onPointerDown={onPointerDown}
            onPointerUp={onPointerUp}
            onPointerCancel={() => (swipeStart.current = null)}
          >
            {count > 1 ? (
              <button
                type="button"
                className="gallery-lightbox__nav gallery-lightbox__nav--prev"
                onClick={() => go(-1)}
                aria-label="Previous drawing"
              >
                <span aria-hidden="true">‹</span>
              </button>
            ) : null}
            <img
              key={item.id}
              src={item.imageUrl}
              alt={`Drawing ${index + 1} of ${count}, by a ward member`}
              width={item.width}
              height={item.height}
              className={`gallery-lightbox__img${loaded ? ' is-loaded' : ''}`}
              onLoad={() => setLoaded(true)}
              draggable={false}
              style={{ backgroundImage: `url("${item.thumbUrl}")` }}
            />
            {count > 1 ? (
              <button
                type="button"
                className="gallery-lightbox__nav gallery-lightbox__nav--next"
                onClick={() => go(1)}
                aria-label="Next drawing"
              >
                <span aria-hidden="true">›</span>
              </button>
            ) : null}
          </div>
          {count > 1 ? (
            <p className="gallery-lightbox__hint">Swipe or use the arrow keys to see more</p>
          ) : null}
        </div>
      ) : null}
    </dialog>
  )
}
