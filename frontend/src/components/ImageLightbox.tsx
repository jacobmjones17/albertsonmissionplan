import { useEffect, useRef, useState } from 'react'

type Props = {
  open: boolean
  src: string
  thumbSrc?: string
  alt: string
  width?: number
  height?: number
  title?: string
  onClose: () => void
}

/** Full-screen viewer for a single image (event flyers). Esc and the backdrop close it. */
export function ImageLightbox({ open, src, thumbSrc, alt, width, height, title, onClose }: Props) {
  const dialogRef = useRef<HTMLDialogElement>(null)
  const [loaded, setLoaded] = useState(false)

  useEffect(() => {
    const dlg = dialogRef.current
    if (!dlg) return
    if (open && !dlg.open) dlg.showModal()
    if (!open && dlg.open) dlg.close()
  }, [open])

  useEffect(() => setLoaded(false), [src])

  return (
    <dialog
      ref={dialogRef}
      className="gallery-lightbox"
      aria-label={title || alt || 'Flyer'}
      onClose={onClose}
      onClick={(e) => {
        if (e.target === e.currentTarget) onClose()
      }}
    >
      {open ? (
        <div className="gallery-lightbox__frame">
          <div className="gallery-lightbox__top">
            <p className="gallery-lightbox__count">{title || 'Flyer'}</p>
            <button
              type="button"
              className="gallery-lightbox__close"
              onClick={onClose}
              aria-label="Close"
              autoFocus
            >
              <span aria-hidden="true">×</span>
            </button>
          </div>
          <div className="gallery-lightbox__stage">
            <img
              key={src}
              src={src}
              alt={alt}
              width={width}
              height={height}
              className={`gallery-lightbox__img${loaded ? ' is-loaded' : ''}`}
              onLoad={() => setLoaded(true)}
              draggable={false}
              style={thumbSrc ? { backgroundImage: `url("${thumbSrc}")` } : undefined}
            />
          </div>
        </div>
      ) : null}
    </dialog>
  )
}
