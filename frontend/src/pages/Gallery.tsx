import { useEffect, useRef, useState } from 'react'
import { apiJson } from '../api'
import { GalleryLightbox } from '../components/GalleryLightbox'
import { PageHero } from '../components/PageHero'
import { galleryCredit, type GalleryItem } from '../lib/gallery'

export function Gallery() {
  const [items, setItems] = useState<GalleryItem[] | null>(null)
  const [err, setErr] = useState<string | null>(null)
  const [openIndex, setOpenIndex] = useState<number | null>(null)
  const lastOpened = useRef<number | null>(null)

  useEffect(() => {
    apiJson<{ items: GalleryItem[] | null }>('/api/gallery')
      .then((d) => setItems(Array.isArray(d.items) ? d.items : []))
      .catch((e) => {
        setErr(e instanceof Error ? e.message : 'Could not load the gallery.')
        setItems([])
      })
  }, [])

  function openAt(i: number) {
    lastOpened.current = i
    setOpenIndex(i)
  }

  function onIndexChange(i: number | null) {
    if (i !== null) lastOpened.current = i
    setOpenIndex(i)
    if (i === null && lastOpened.current !== null) {
      // Return focus to the drawing that was last shown (it may differ from the one clicked).
      const id = items?.[lastOpened.current]?.id
      window.requestAnimationFrame(() =>
        document.querySelector<HTMLButtonElement>(`[data-gallery-id="${id}"]`)?.focus(),
      )
    }
  }

  return (
    <main id="main">
      <PageHero
        titleId="gallery-title"
        eyebrow="Made by our ward"
        title="Ward Gallery"
        lede="Drawings and art from our Primary children and ward members"
      />
      <div className="home-mission-strip home-mission-strip--tight">
        <p>
          Our children and members helped make this site. Every picture here was drawn by someone
          in the Albertson Ward. Thank you for sharing your talents and your testimony!
        </p>
      </div>
      <div className="wrap-wide gallery-page">
        {err ? <p className="panel-sites panel-sites--warn">{err}</p> : null}
        {items === null ? (
          <p className="loading-note" role="status">
            <span className="loading-note__spinner" aria-hidden="true" />
            Loading drawings…
          </p>
        ) : items.length === 0 ? (
          !err && (
            <section className="gallery-empty" aria-labelledby="gallery-empty-title">
              <div className="gallery-empty__art" aria-hidden="true">
                <svg viewBox="0 0 64 64" width="64" height="64">
                  <rect x="6" y="10" width="52" height="40" rx="6" fill="none" stroke="currentColor" strokeWidth="3" />
                  <circle cx="22" cy="24" r="5" fill="currentColor" opacity="0.55" />
                  <path d="M10 46l14-14 10 10 8-7 12 11" fill="none" stroke="currentColor" strokeWidth="3" strokeLinejoin="round" />
                  <path d="M46 4l3 6 6 1-4.5 4 1 6-5.5-3-5.5 3 1-6L37 11l6-1z" fill="#e0b868" />
                </svg>
              </div>
              <h2 id="gallery-empty-title">Drawings are coming soon</h2>
              <p>
                Primary children and ward members are making pictures for this page. If you or your
                child drew something you&apos;d like to share, give it to a Primary teacher or a ward
                leader and we&apos;ll scan it in.
              </p>
            </section>
          )
        ) : (
          <>
            <p className="gallery-count">
              {items.length} {items.length === 1 ? 'drawing' : 'drawings'} · tap any picture to see it larger
            </p>
            <ul className="gallery-grid" role="list">
              {items.map((it, i) => {
                const credit = galleryCredit(it)
                return (
                  <li key={it.id} className={`gallery-card${it.featured ? ' gallery-card--featured' : ''}`}>
                    <button
                      type="button"
                      className="gallery-card__button"
                      data-gallery-id={it.id}
                      onClick={() => openAt(i)}
                      aria-label={`View larger: ${it.alt}`}
                    >
                      <img
                        src={it.thumbUrl}
                        alt=""
                        width={it.thumbWidth}
                        height={it.thumbHeight}
                        loading={i < 6 ? 'eager' : 'lazy'}
                        decoding="async"
                      />
                      {it.featured ? <span className="gallery-card__badge">★ Featured</span> : null}
                    </button>
                    {it.title || credit || it.caption ? (
                      <div className="gallery-card__caption">
                        {it.title ? <p className="gallery-card__title">{it.title}</p> : null}
                        {credit ? <p className="gallery-card__credit">{credit}</p> : null}
                        {it.caption ? <p className="gallery-card__text">{it.caption}</p> : null}
                      </div>
                    ) : null}
                  </li>
                )
              })}
            </ul>
          </>
        )}
      </div>
      <GalleryLightbox items={items ?? []} index={openIndex} onIndexChange={onIndexChange} />
    </main>
  )
}
